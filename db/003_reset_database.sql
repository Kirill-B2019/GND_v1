-- 003_reset_database.sql
-- Удаление всех записей из таблиц и сброс счётчиков (sequences) для auto increment.
-- ВНИМАНИЕ: полная очистка данных. Выполнять от имени владельца БД (gnduser или superuser).
-- После выполнения при следующем запуске ноды будет создан новый генезис и кошелёк.

BEGIN;

-- Очистка: порядок учитывает FK. CASCADE очистит зависимые партиции/таблицы.
-- Партиционированные таблицы (transactions, logs) — очищаются по родителю.
TRUNCATE TABLE public.logs CASCADE;
TRUNCATE TABLE public.transactions CASCADE;

-- states и token_balances (опционально: только если таблица есть)
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'states') THEN
    EXECUTE 'TRUNCATE TABLE public.states CASCADE';
  END IF;
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'token_balances') THEN
    EXECUTE 'TRUNCATE TABLE public.token_balances CASCADE';
  END IF;
END $$;

TRUNCATE TABLE public.blocks RESTART IDENTITY CASCADE;
TRUNCATE TABLE public.wallets CASCADE;
-- signer_wallets (кастодиальные ключи): очистить до accounts из-за FK account_id
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'signer_wallets') THEN
    EXECUTE 'TRUNCATE TABLE public.signer_wallets CASCADE';
  END IF;
END $$;
TRUNCATE TABLE public.tokens RESTART IDENTITY CASCADE;
TRUNCATE TABLE public.events RESTART IDENTITY CASCADE;
TRUNCATE TABLE public.api_keys RESTART IDENTITY CASCADE;
TRUNCATE TABLE public.oracles RESTART IDENTITY CASCADE;
TRUNCATE TABLE public.metrics RESTART IDENTITY CASCADE;
TRUNCATE TABLE public.poa_validators RESTART IDENTITY CASCADE;
TRUNCATE TABLE public.pos_validators RESTART IDENTITY CASCADE;
TRUNCATE TABLE public.validators RESTART IDENTITY CASCADE;
TRUNCATE TABLE public.contracts RESTART IDENTITY CASCADE;
TRUNCATE TABLE public.accounts RESTART IDENTITY CASCADE;

-- Сброс всех последовательностей схемы public на 1
DO $$
DECLARE
  r RECORD;
BEGIN
  FOR r IN
    SELECT seq.relname AS seq_name
    FROM pg_class seq
    JOIN pg_namespace n ON n.oid = seq.relnamespace
    WHERE seq.relkind = 'S'
      AND n.nspname = 'public'
  LOOP
    EXECUTE format('ALTER SEQUENCE public.%I RESTART WITH 1', r.seq_name);
  END LOOP;
END $$;

COMMIT;
