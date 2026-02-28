-- KB @CerbeRus - Nexus Invest Team
-- Таблица signer_wallets для кастодиальных ключей (встроенный signing_service).
-- Применять после db.sql (таблица accounts уже должна существовать).

CREATE TABLE IF NOT EXISTS public.signer_wallets (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id      INTEGER NOT NULL REFERENCES public.accounts(id) ON DELETE CASCADE,
    public_key      BYTEA NOT NULL,
    encrypted_priv  BYTEA NOT NULL,
    disabled        BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(account_id)
);

ALTER TABLE public.signer_wallets OWNER TO gnduser;
CREATE INDEX IF NOT EXISTS idx_signer_wallets_account_id ON public.signer_wallets (account_id);
