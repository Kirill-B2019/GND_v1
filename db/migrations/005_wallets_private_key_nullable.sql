-- KB @CerbeRus - Nexus Invest Team
-- Разрешить NULL в wallets.private_key для кошельков, созданных через signing_service.
-- Добавить ссылку на signer_wallets для подписи сервером.

ALTER TABLE public.wallets ALTER COLUMN private_key DROP NOT NULL;

-- Ссылка на signer_wallets (при создании кошелька через signing_service заполняется)
ALTER TABLE public.wallets
  ADD COLUMN IF NOT EXISTS signer_wallet_id UUID REFERENCES public.signer_wallets(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_wallets_signer_wallet_id ON public.wallets (signer_wallet_id) WHERE signer_wallet_id IS NOT NULL;
