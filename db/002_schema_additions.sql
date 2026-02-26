-- 002_schema_additions.sql
-- Дополнения к схеме console_21: недостающие таблицы, столбцы и зависимости по анализу кода.
-- Выполнять после применения db/console_21.sql (и при необходимости db/migrations/001_create_events_table.sql).
-- Владелец объектов: gnduser.

-- =============================================================================
-- 1. Функция для триггера events (если ещё не создана миграцией 001)
-- plpgsql — встроенный процедурный язык PostgreSQL (не требует CREATE LANGUAGE).
-- =============================================================================
-- noinspection SqlResolve
CREATE OR REPLACE FUNCTION public.update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- =============================================================================
-- 2. Таблица states (core/state.go: BlockchainState)
-- =============================================================================
CREATE TABLE IF NOT EXISTS public.states (
    id            SERIAL PRIMARY KEY,
    block_id      INTEGER NOT NULL,
    address       VARCHAR NOT NULL,
    balance       NUMERIC,
    nonce         BIGINT DEFAULT 0,
    storage_root  VARCHAR,
    code_hash     VARCHAR,
    created_at    TIMESTAMP,
    updated_at    TIMESTAMP,
    metadata      BYTEA
);

ALTER TABLE public.states OWNER TO gnduser;
CREATE INDEX IF NOT EXISTS idx_states_block_id ON public.states (block_id);
CREATE INDEX IF NOT EXISTS idx_states_address ON public.states (address);
CREATE UNIQUE INDEX IF NOT EXISTS idx_states_address_block ON public.states (address, block_id);

-- =============================================================================
-- 3. Блоки (blocks): столбцы для core/block.go
-- =============================================================================
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'blocks' AND column_name = 'merkle_root') THEN
        ALTER TABLE public.blocks ADD COLUMN merkle_root VARCHAR;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'blocks' AND column_name = 'height') THEN
        ALTER TABLE public.blocks ADD COLUMN height BIGINT;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'blocks' AND column_name = 'version') THEN
        ALTER TABLE public.blocks ADD COLUMN version INTEGER DEFAULT 1;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'blocks' AND column_name = 'size') THEN
        ALTER TABLE public.blocks ADD COLUMN size BIGINT;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'blocks' AND column_name = 'difficulty') THEN
        ALTER TABLE public.blocks ADD COLUMN difficulty BIGINT;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'blocks' AND column_name = 'extra_data') THEN
        ALTER TABLE public.blocks ADD COLUMN extra_data BYTEA;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'blocks' AND column_name = 'created_at') THEN
        ALTER TABLE public.blocks ADD COLUMN created_at TIMESTAMP;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'blocks' AND column_name = 'updated_at') THEN
        ALTER TABLE public.blocks ADD COLUMN updated_at TIMESTAMP;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'blocks' AND column_name = 'status') THEN
        ALTER TABLE public.blocks ADD COLUMN status VARCHAR;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'blocks' AND column_name = 'parent_id') THEN
        ALTER TABLE public.blocks ADD COLUMN parent_id BIGINT;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'blocks' AND column_name = 'is_orphaned') THEN
        ALTER TABLE public.blocks ADD COLUMN is_orphaned BOOLEAN DEFAULT FALSE;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'blocks' AND column_name = 'is_finalized') THEN
        ALTER TABLE public.blocks ADD COLUMN is_finalized BOOLEAN DEFAULT FALSE;
    END IF; END $$;

-- Синхронизация height с index при наличии данных (одноразово; index в blocks NOT NULL)
UPDATE public.blocks SET height = index WHERE height IS NULL;

-- =============================================================================
-- 4. Транзакции (transactions): signature, is_verified (core/transaction.go)
-- =============================================================================
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'transactions' AND column_name = 'signature') THEN
        ALTER TABLE public.transactions ADD COLUMN signature VARCHAR;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'transactions' AND column_name = 'is_verified') THEN
        ALTER TABLE public.transactions ADD COLUMN is_verified BOOLEAN DEFAULT FALSE;
    END IF; END $$;

-- =============================================================================
-- 5. Аккаунты (accounts): столбцы для core/account.go
-- =============================================================================
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'accounts' AND column_name = 'type') THEN
        ALTER TABLE public.accounts ADD COLUMN type VARCHAR;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'accounts' AND column_name = 'status') THEN
        ALTER TABLE public.accounts ADD COLUMN status VARCHAR;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'accounts' AND column_name = 'block_id') THEN
        ALTER TABLE public.accounts ADD COLUMN block_id INTEGER;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'accounts' AND column_name = 'tx_id') THEN
        ALTER TABLE public.accounts ADD COLUMN tx_id INTEGER;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'accounts' AND column_name = 'gas_limit') THEN
        ALTER TABLE public.accounts ADD COLUMN gas_limit BIGINT;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'accounts' AND column_name = 'gas_used') THEN
        ALTER TABLE public.accounts ADD COLUMN gas_used BIGINT;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'accounts' AND column_name = 'value') THEN
        ALTER TABLE public.accounts ADD COLUMN value VARCHAR;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'accounts' AND column_name = 'data') THEN
        ALTER TABLE public.accounts ADD COLUMN data BYTEA;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'accounts' AND column_name = 'updated_at') THEN
        ALTER TABLE public.accounts ADD COLUMN updated_at TIMESTAMP;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'accounts' AND column_name = 'is_verified') THEN
        ALTER TABLE public.accounts ADD COLUMN is_verified BOOLEAN DEFAULT FALSE;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'accounts' AND column_name = 'source_code') THEN
        ALTER TABLE public.accounts ADD COLUMN source_code TEXT;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'accounts' AND column_name = 'compiler') THEN
        ALTER TABLE public.accounts ADD COLUMN compiler VARCHAR;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'accounts' AND column_name = 'optimized') THEN
        ALTER TABLE public.accounts ADD COLUMN optimized BOOLEAN DEFAULT FALSE;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'accounts' AND column_name = 'runs') THEN
        ALTER TABLE public.accounts ADD COLUMN runs INTEGER;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'accounts' AND column_name = 'license') THEN
        ALTER TABLE public.accounts ADD COLUMN license VARCHAR;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'accounts' AND column_name = 'metadata') THEN
        ALTER TABLE public.accounts ADD COLUMN metadata JSONB;
    END IF; END $$;

-- =============================================================================
-- 6. Контракты (contracts): столбцы для core/contract.go и core/state.go
-- =============================================================================
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'contracts' AND column_name = 'creator') THEN
        ALTER TABLE public.contracts ADD COLUMN creator VARCHAR;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'contracts' AND column_name = 'bytecode') THEN
        ALTER TABLE public.contracts ADD COLUMN bytecode BYTEA;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'contracts' AND column_name = 'name') THEN
        ALTER TABLE public.contracts ADD COLUMN name VARCHAR;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'contracts' AND column_name = 'symbol') THEN
        ALTER TABLE public.contracts ADD COLUMN symbol VARCHAR;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'contracts' AND column_name = 'standard') THEN
        ALTER TABLE public.contracts ADD COLUMN standard VARCHAR;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'contracts' AND column_name = 'description') THEN
        ALTER TABLE public.contracts ADD COLUMN description TEXT;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'contracts' AND column_name = 'version') THEN
        ALTER TABLE public.contracts ADD COLUMN version VARCHAR;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'contracts' AND column_name = 'status') THEN
        ALTER TABLE public.contracts ADD COLUMN status VARCHAR;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'contracts' AND column_name = 'block_id') THEN
        ALTER TABLE public.contracts ADD COLUMN block_id INTEGER;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'contracts' AND column_name = 'tx_id') THEN
        ALTER TABLE public.contracts ADD COLUMN tx_id INTEGER;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'contracts' AND column_name = 'gas_limit') THEN
        ALTER TABLE public.contracts ADD COLUMN gas_limit BIGINT;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'contracts' AND column_name = 'gas_used') THEN
        ALTER TABLE public.contracts ADD COLUMN gas_used BIGINT;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'contracts' AND column_name = 'value') THEN
        ALTER TABLE public.contracts ADD COLUMN value VARCHAR;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'contracts' AND column_name = 'data') THEN
        ALTER TABLE public.contracts ADD COLUMN data BYTEA;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'contracts' AND column_name = 'updated_at') THEN
        ALTER TABLE public.contracts ADD COLUMN updated_at TIMESTAMP;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'contracts' AND column_name = 'is_verified') THEN
        ALTER TABLE public.contracts ADD COLUMN is_verified BOOLEAN DEFAULT FALSE;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'contracts' AND column_name = 'source_code') THEN
        ALTER TABLE public.contracts ADD COLUMN source_code TEXT;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'contracts' AND column_name = 'compiler') THEN
        ALTER TABLE public.contracts ADD COLUMN compiler VARCHAR;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'contracts' AND column_name = 'optimized') THEN
        ALTER TABLE public.contracts ADD COLUMN optimized BOOLEAN DEFAULT FALSE;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'contracts' AND column_name = 'runs') THEN
        ALTER TABLE public.contracts ADD COLUMN runs INTEGER;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'contracts' AND column_name = 'license') THEN
        ALTER TABLE public.contracts ADD COLUMN license VARCHAR;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'contracts' AND column_name = 'metadata') THEN
        ALTER TABLE public.contracts ADD COLUMN metadata JSONB;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'contracts' AND column_name = 'params') THEN
        ALTER TABLE public.contracts ADD COLUMN params JSONB;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'contracts' AND column_name = 'metadata_cid') THEN
        ALTER TABLE public.contracts ADD COLUMN metadata_cid VARCHAR;
    END IF; END $$;

-- =============================================================================
-- 7. Токены (tokens): столбцы для core/token.go
-- =============================================================================
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'tokens' AND column_name = 'status') THEN
        ALTER TABLE public.tokens ADD COLUMN status VARCHAR;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'tokens' AND column_name = 'updated_at') THEN
        ALTER TABLE public.tokens ADD COLUMN updated_at TIMESTAMP;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'tokens' AND column_name = 'is_verified') THEN
        ALTER TABLE public.tokens ADD COLUMN is_verified BOOLEAN DEFAULT FALSE;
    END IF; END $$;

-- =============================================================================
-- 8. Балансы токенов (token_balances): столбец symbol для core/state.go
-- =============================================================================
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'token_balances' AND column_name = 'symbol') THEN
        ALTER TABLE public.token_balances ADD COLUMN symbol VARCHAR;
    END IF; END $$;
-- Частичный уникальный индекс (address, symbol) для ON CONFLICT в state.go, только при непустом symbol
CREATE UNIQUE INDEX IF NOT EXISTS idx_token_balances_address_symbol
    ON public.token_balances (address, symbol) WHERE symbol IS NOT NULL;

-- =============================================================================
-- 9. События (events): столбцы для core/event.go (BlockchainEvent)
-- =============================================================================
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'events' AND column_name = 'block_id') THEN
        ALTER TABLE public.events ADD COLUMN block_id INTEGER;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'events' AND column_name = 'tx_id') THEN
        ALTER TABLE public.events ADD COLUMN tx_id INTEGER;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'events' AND column_name = 'address') THEN
        ALTER TABLE public.events ADD COLUMN address VARCHAR(42);
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'events' AND column_name = 'topics') THEN
        ALTER TABLE public.events ADD COLUMN topics JSONB;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'events' AND column_name = 'data') THEN
        ALTER TABLE public.events ADD COLUMN data JSONB;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'events' AND column_name = 'index') THEN
        ALTER TABLE public.events ADD COLUMN index INTEGER;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'events' AND column_name = 'removed') THEN
        ALTER TABLE public.events ADD COLUMN removed BOOLEAN DEFAULT FALSE;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'events' AND column_name = 'status') THEN
        ALTER TABLE public.events ADD COLUMN status VARCHAR;
    END IF; END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'events' AND column_name = 'processed_at') THEN
        ALTER TABLE public.events ADD COLUMN processed_at TIMESTAMP WITH TIME ZONE;
    END IF; END $$;

-- Триггер updated_at для events (если ещё не создан)
DROP TRIGGER IF EXISTS update_events_updated_at ON public.events;
CREATE TRIGGER update_events_updated_at
    BEFORE UPDATE ON public.events
    FOR EACH ROW
    EXECUTE PROCEDURE public.update_updated_at_column();

-- =============================================================================
-- 10. Логи (logs): последовательность для id (api/websocket.go вставляет без id)
-- =============================================================================
CREATE SEQUENCE IF NOT EXISTS public.logs_id_seq OWNED BY public.logs.id;
ALTER TABLE public.logs ALTER COLUMN id SET DEFAULT nextval('public.logs_id_seq');

-- =============================================================================
-- Комментарий
-- =============================================================================
COMMENT ON TABLE public.states IS 'Состояние по адресам и блокам (core.BlockchainState)';
