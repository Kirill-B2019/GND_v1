-- dump.sql
-- Дамп БД с сохранением схемы, данных и связей (FK, индексы, последовательности).
--
-- Запуск из корня проекта (GND_v1):
--   psql -h HOST -p PORT -U gnduser -d gnd_db -f db/dump.sql
-- Пароль: переменная PGPASSWORD или ~/.pgpass.
-- Результат: файл db/gnd_dump.sql в папке db.

\echo 'Creating dump -> db/gnd_dump.sql ...'

\! pg_dump -h "${PGHOST:-localhost}" -p "${PGPORT:-5432}" -U "${PGUSER:-gnduser}" -d "${PGDATABASE:-gnd_db}" --no-owner --no-privileges -F p -f db/gnd_dump.sql

\echo 'Done: db/gnd_dump.sql'
