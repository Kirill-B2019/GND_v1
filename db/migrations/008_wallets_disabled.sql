-- KB @CerbeRus - Nexus Invest Team
-- Мягкое удаление кошелька (скрытие из списка, запись остаётся).

ALTER TABLE public.wallets ADD COLUMN IF NOT EXISTS disabled BOOLEAN NOT NULL DEFAULT false;
COMMENT ON COLUMN public.wallets.disabled IS 'Мягкое удаление: кошелёк скрыт из списка админки';

CREATE INDEX IF NOT EXISTS idx_wallets_disabled ON public.wallets (disabled) WHERE disabled = false;
