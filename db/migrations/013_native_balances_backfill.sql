-- Перенос балансов GND и GANI из token_balances в native_balances без полного сброса блокчейна.
-- Выполнять после 012_native_balances.sql. Идемпотентно: можно запускать повторно.
-- | KB @CerbeRus - Nexus Invest Team 2026

INSERT INTO public.native_balances (address, symbol, balance, updated_at)
SELECT
    tb.address::VARCHAR(128),
    t.symbol,
    COALESCE(tb.balance, 0)::NUMERIC(78, 0),
    CURRENT_TIMESTAMP
FROM public.token_balances tb
JOIN public.tokens t ON t.id = tb.token_id
WHERE t.symbol IN ('GND', 'GANI')
  AND tb.address IS NOT NULL
  AND t.symbol IS NOT NULL
ON CONFLICT (address, symbol) DO UPDATE SET
    balance   = EXCLUDED.balance,
    updated_at = EXCLUDED.updated_at;
