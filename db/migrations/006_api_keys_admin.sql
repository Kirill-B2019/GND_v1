-- KB @CerbeRus - Nexus Invest Team
-- Расширение api_keys для админ-выдачи ключей: имя, префикс, хеш ключа, отключение.

-- Поле key оставляем для обратной совместимости (старые ключи в открытом виде).
-- Новые ключи: храним только key_hash (SHA-256 hex) и key_prefix (первые 8 символов для отображения).
-- key для старых записей может оставаться; для новых не заполняем.
ALTER TABLE public.api_keys ADD COLUMN IF NOT EXISTS name VARCHAR(255);
ALTER TABLE public.api_keys ADD COLUMN IF NOT EXISTS key_prefix VARCHAR(16);
ALTER TABLE public.api_keys ADD COLUMN IF NOT EXISTS key_hash VARCHAR(64) UNIQUE;
ALTER TABLE public.api_keys ADD COLUMN IF NOT EXISTS disabled BOOLEAN NOT NULL DEFAULT false;
-- Разрешить NULL в key (новые ключи хранятся только в виде key_hash)
DO $$ BEGIN ALTER TABLE public.api_keys ALTER COLUMN key DROP NOT NULL; EXCEPTION WHEN OTHERS THEN NULL; END $$;

COMMENT ON COLUMN public.api_keys.name IS 'Человекочитаемое имя ключа (например: Laravel Backend)';
COMMENT ON COLUMN public.api_keys.key_prefix IS 'Префикс ключа для отображения в списке (например gnd_ab12)';
COMMENT ON COLUMN public.api_keys.key_hash IS 'SHA-256 хеш ключа в hex; для проверки без хранения открытого ключа';
COMMENT ON COLUMN public.api_keys.disabled IS 'Отозванный ключ не принимается';

CREATE INDEX IF NOT EXISTS idx_api_keys_key_hash ON public.api_keys (key_hash) WHERE key_hash IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_api_keys_disabled ON public.api_keys (disabled) WHERE disabled = false;
