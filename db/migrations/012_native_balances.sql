-- Нативные монеты GND и GANI: отдельная таблица для L1-балансов.
-- Изменяются только нодой (транзакции, газ, консенсус). Защита от кражи и сохранность при перезагрузке.
-- | KB @CerberRus00 - Nexus Invest Team 2026

CREATE TABLE IF NOT EXISTS public.native_balances (
    address   VARCHAR(128) NOT NULL,
    symbol    VARCHAR(10)  NOT NULL,
    balance   NUMERIC(78, 0) NOT NULL DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (address, symbol),
    CONSTRAINT chk_native_symbol CHECK (symbol IN ('GND', 'GANI'))
);

CREATE INDEX IF NOT EXISTS idx_native_balances_address ON public.native_balances (address);
CREATE INDEX IF NOT EXISTS idx_native_balances_symbol ON public.native_balances (symbol);

COMMENT ON TABLE public.native_balances IS 'Балансы нативных монет L1 (GND, GANI). Источник истины; изменяются только нодой.';
