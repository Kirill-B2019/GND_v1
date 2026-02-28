-- KB @CerbeRus - Nexus Invest Team
-- Имена и роли для кошельков (в т.ч. системных: validator, treasury).

ALTER TABLE public.wallets ADD COLUMN IF NOT EXISTS name VARCHAR(255);
ALTER TABLE public.wallets ADD COLUMN IF NOT EXISTS role VARCHAR(64);

COMMENT ON COLUMN public.wallets.name IS 'Человекочитаемое имя (например: Validator, Treasury)';
COMMENT ON COLUMN public.wallets.role IS 'Системная роль: validator, treasury, fee_collector или NULL';

CREATE INDEX IF NOT EXISTS idx_wallets_role ON public.wallets (role) WHERE role IS NOT NULL;
