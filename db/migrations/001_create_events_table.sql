-- Создание таблицы events
CREATE TABLE IF NOT EXISTS events (
    id BIGSERIAL PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    contract VARCHAR(42) NOT NULL,
    from_address VARCHAR(42),
    to_address VARCHAR(42),
    amount NUMERIC(78,0), -- Для хранения больших чисел (до 78 цифр)
    "timestamp" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    tx_hash VARCHAR(66),
    error TEXT,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Создание индексов
CREATE INDEX IF NOT EXISTS idx_events_contract ON events(contract);
CREATE INDEX IF NOT EXISTS idx_events_type ON events(type);
CREATE INDEX IF NOT EXISTS idx_events_timestamp ON events("timestamp");
CREATE INDEX IF NOT EXISTS idx_events_tx_hash ON events(tx_hash);
CREATE INDEX IF NOT EXISTS idx_events_from_address ON events(from_address);
CREATE INDEX IF NOT EXISTS idx_events_to_address ON events(to_address);
CREATE INDEX IF NOT EXISTS idx_events_contract_type ON events(contract, type);
CREATE INDEX IF NOT EXISTS idx_events_contract_timestamp ON events(contract, "timestamp");

-- Создание функции для автоматического обновления updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Создание триггера для автоматического обновления updated_at
CREATE TRIGGER update_events_updated_at
    BEFORE UPDATE ON events
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Создание представления для последних событий
CREATE OR REPLACE VIEW latest_events AS
SELECT 
    e.*,
    ROW_NUMBER() OVER (PARTITION BY e.contract, e.type ORDER BY e."timestamp" DESC) as event_rank
FROM events e;

-- Создание функции для получения последних событий
CREATE OR REPLACE FUNCTION get_latest_events(
    p_contract VARCHAR,
    p_type VARCHAR,
    p_limit INTEGER
)
RETURNS TABLE (
    id BIGINT,
    type VARCHAR,
    contract VARCHAR,
    from_address VARCHAR,
    to_address VARCHAR,
    amount NUMERIC,
    "timestamp" TIMESTAMP WITH TIME ZONE,
    tx_hash VARCHAR,
    error TEXT,
    metadata JSONB
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        e.id,
        e.type,
        e.contract,
        e.from_address,
        e.to_address,
        e.amount,
        e."timestamp",
        e.tx_hash,
        e.error,
        e.metadata
    FROM events e
    WHERE e.contract = p_contract
        AND e.type = p_type
    ORDER BY e."timestamp" DESC
    LIMIT p_limit;
END;
$$ LANGUAGE plpgsql;

-- Создание функции для агрегации событий
CREATE OR REPLACE FUNCTION get_event_stats(
    p_contract VARCHAR,
    p_start_time TIMESTAMP WITH TIME ZONE,
    p_end_time TIMESTAMP WITH TIME ZONE
)
RETURNS TABLE (
    event_type VARCHAR,
    event_count BIGINT,
    total_amount NUMERIC,
    first_event TIMESTAMP WITH TIME ZONE,
    last_event TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        e.type as event_type,
        COUNT(*) as event_count,
        COALESCE(SUM(e.amount), 0) as total_amount,
        MIN(e."timestamp") as first_event,
        MAX(e."timestamp") as last_event
    FROM events e
    WHERE e.contract = p_contract
        AND e."timestamp" BETWEEN p_start_time AND p_end_time
    GROUP BY e.type;
END;
$$ LANGUAGE plpgsql;

-- Комментарии к таблице и колонкам
COMMENT ON TABLE events IS 'Таблица для хранения событий блокчейна';
COMMENT ON COLUMN events.type IS 'Тип события (Transfer, Approval, ContractDeployment, Error)';
COMMENT ON COLUMN events.contract IS 'Адрес контракта';
COMMENT ON COLUMN events.from_address IS 'Адрес отправителя';
COMMENT ON COLUMN events.to_address IS 'Адрес получателя';
COMMENT ON COLUMN events.amount IS 'Количество токенов';
COMMENT ON COLUMN events."timestamp" IS 'Время создания события';
COMMENT ON COLUMN events.tx_hash IS 'Хеш транзакции';
COMMENT ON COLUMN events.error IS 'Текст ошибки';
COMMENT ON COLUMN events.metadata IS 'Дополнительные метаданные в формате JSON';
COMMENT ON COLUMN events.created_at IS 'Время создания записи';
COMMENT ON COLUMN events.updated_at IS 'Время последнего обновления записи'; 