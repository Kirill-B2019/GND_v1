-- KB @CerbeRus - Nexus Invest Team
-- Колонка circulating_supply в tokens (заполняется из config coins.json).

ALTER TABLE public.tokens ADD COLUMN IF NOT EXISTS circulating_supply NUMERIC;
COMMENT ON COLUMN public.tokens.circulating_supply IS 'Обращающееся предложение; заполняется из config coins.circulating_supply';
