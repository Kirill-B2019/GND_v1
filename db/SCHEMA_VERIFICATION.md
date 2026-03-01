# Сверка схемы console_21.sql с доработками

**Дата сверки:** 2026-03-01  
**Файл схемы:** `db/console_21.sql`

## Итог

Схема **console_21.sql** приведена в соответствие со всеми доработками (002_schema_additions, миграции 001, 004–009). В файл добавлены заголовочный комментарий и недостающий триггер для таблицы `events`.

---

## Соответствие по таблицам

| Таблица / объект | 002 / миграции | В console_21 | Примечание |
|------------------|----------------|--------------|------------|
| **contracts** | creator, bytecode, name, symbol, standard, description, version, status, block_id, tx_id, gas_* , value, data, updated_at, is_verified, source_code, compiler, optimized, runs, license, metadata, params, metadata_cid | ✓ Все колонки | — |
| **tokens** | status, updated_at, is_verified (002); circulating_supply (009); logo_url (011) | ✓ Все колонки + комментарии | — |
| **accounts** | type, status, block_id, tx_id, gas_*, value, data, updated_at, is_verified, source_code, compiler, optimized, runs, license, metadata | ✓ Все колонки | — |
| **api_keys** | name, key_prefix, key_hash, disabled (006); key nullable | ✓ Все колонки, key без NOT NULL | — |
| **blocks** | merkle_root, height, version, size, difficulty, extra_data, created_at, updated_at, status, parent_id, is_orphaned, is_finalized | ✓ Все колонки | — |
| **transactions** | signature, is_verified; sequence transactions_id_seq | ✓ | — |
| **token_balances** | symbol, UNIQUE (address, symbol) WHERE symbol IS NOT NULL | ✓ | — |
| **events** | block_id, tx_id, address, topics, data, index, removed, status, processed_at | ✓ Все колонки | — |
| **events** — триггер | update_events_updated_at (002/001) | ✓ Добавлен в конец файла | Раньше отсутствовал |
| **states** | 002 | ✓ Таблица и индексы | — |
| **signer_wallets** | 004 | ✓ | — |
| **wallets** | private_key nullable, signer_wallet_id (005); name, role (007); disabled (008) | ✓ Все колонки и комментарии | — |
| **logs** | logs_id_seq, id DEFAULT nextval | ✓ | — |
| **update_updated_at_column()** | 002 | ✓ В конце файла | — |

---

## Замечание по коду

- **core/state.go** (SaveToDB): выполняется `INSERT INTO token_balances (address, symbol, balance)` без `token_id`. В схеме у `token_balances` колонка `token_id` объявлена как `NOT NULL` и входит в первичный ключ. Этот фрагмент кода не соответствует текущей схеме (либо это устаревший путь, либо нужна доработка кода/схемы).

---

## Порядок применения при чистом развёртывании

1. Создать БД и роль `gnduser`.
2. Выполнить **console_21.sql** — этого достаточно для получения актуальной схемы (миграции 001, 002, 004–009 уже учтены в файле).
3. При развёртывании по миграциям: 001 → 002 → 004–009 → 010 → **011_tokens_logo_url.sql** (колонка logo_url для логотипов токенов).
