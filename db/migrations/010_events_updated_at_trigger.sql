-- Триггер автообновления updated_at для таблицы events.
-- Синхронизация с console_21.sql и 002_schema_additions.sql.
-- Применять после 001_create_events_table.sql (таблица events и при необходимости функция уже есть).

-- Функция триггера (если ещё не создана 002_schema_additions или console_21)
CREATE OR REPLACE FUNCTION public.update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

ALTER FUNCTION public.update_updated_at_column() OWNER TO gnduser;

-- Триггер на events
DROP TRIGGER IF EXISTS update_events_updated_at ON public.events;
CREATE TRIGGER update_events_updated_at
    BEFORE UPDATE ON public.events
    FOR EACH ROW
    EXECUTE PROCEDURE public.update_updated_at_column();
