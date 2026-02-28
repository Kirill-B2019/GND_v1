-- 003_reset_database.sql
-- Удаление всех записей из таблиц и сброс счётчиков (sequences).
-- ВНИМАНИЕ: полная очистка данных. Выполнять от имени владельца БД (gnduser или superuser).
-- Порядок: дочерние таблицы и партиции первыми (CASCADE очистит зависимые партиции).
-- После выполнения при следующем запуске ноды будет создан новый генезис и кошелёк.

BEGIN;

-- Партиционированные и зависимые таблицы
TRUNCATE TABLE public.logs CASCADE;
TRUNCATE TABLE public.transactions CASCADE;

-- Таблицы, зависящие от blocks, contracts, accounts, tokens, validators
TRUNCATE TABLE public.token_balances CASCADE;
TRUNCATE TABLE public.states CASCADE;
TRUNCATE TABLE public.events RESTART IDENTITY CASCADE;
TRUNCATE TABLE public.blocks RESTART IDENTITY CASCADE;
TRUNCATE TABLE public.wallets CASCADE;
TRUNCATE TABLE public.signer_wallets CASCADE;
TRUNCATE TABLE public.poa_validators CASCADE;
TRUNCATE TABLE public.pos_validators CASCADE;
TRUNCATE TABLE public.tokens RESTART IDENTITY CASCADE;
TRUNCATE TABLE public.contracts RESTART IDENTITY CASCADE;
TRUNCATE TABLE public.validators RESTART IDENTITY CASCADE;
TRUNCATE TABLE public.api_keys RESTART IDENTITY CASCADE;
TRUNCATE TABLE public.oracles RESTART IDENTITY CASCADE;
TRUNCATE TABLE public.metrics RESTART IDENTITY CASCADE;
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
