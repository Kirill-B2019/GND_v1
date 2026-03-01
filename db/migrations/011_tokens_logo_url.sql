-- Колонка logo_url в tokens: ссылка на логотип токена (URL или путь на сервере).
-- Валидация и загрузка: через API (250x250 px, тип картинки). Хранить на сервере в uploads/token_logos/.

ALTER TABLE public.tokens ADD COLUMN IF NOT EXISTS logo_url VARCHAR(512);
COMMENT ON COLUMN public.tokens.logo_url IS 'URL или путь к логотипу токена (до 250x250 px, изображение)';
