# Интеграция контрактов: GND_v1 ↔ GND_admin

Краткая проверка связки ноды и админ-панели по работе с контрактами (чтение/запись состояния, storage).

## Жизненный цикл контракта

Все контракты проходят: **проверка (компиляция) → анализ → деплой**. После деплоя чтение и запись выполняются через интерфейс GND_admin и API ноды.

- **Проверка/компиляция:** исходный код компилируется (POST /contract/compile на ноде), сохраняется bytecode.
- **Анализ:** опциональный шаг анализа безопасности (POST /contract/analyze).
- **Деплой:** создаёт контракт на ноде и запись в таблице `contracts`; **формирует транзакцию** в блокчейне (тип `contract_deploy`, запись в `transactions`).
- **Чтение:** вкладка «Прочитать контракт» и API (GET /contract/:address/state, GET /state/contract/:address/storage и т.д.).
- **Запись:** вкладка «Записать контракт» (слот storage); **формирует транзакцию** (тип `contract_storage_write` в `transactions`).

**Все действия с контрактами (деплой, запись storage) формируют транзакции в блокчейне** (таблица `transactions` на ноде).

## Эндпоинты GND_v1 (REST API)

| Метод | Путь | Назначение |
|-------|------|------------|
| GET | `/api/v1/contract/:address` | Информация о контракте (core.GetContract) |
| GET | `/api/v1/contract/:address/state?addresses=addr1,addr2` | Состояние контракта: name, symbol, owner, decimals, total_supply, balances (core.GetContractState) |
| GET | `/api/v1/state/account/:address` | Текущее состояние аккаунта из `accounts`: nonce, balance_gnd (core.GetCurrentAccountState) |
| GET | `/api/v1/state/account/:address/block/:blockId` | Снимок состояния на блок из `account_states` |
| GET | `/api/v1/state/contract/:address/storage?block_id=N` | Слоты storage контракта на блок (core.GetContractStorageAtBlock) |
| POST | `/api/v1/admin/state/contract/:address/storage` | Запись слота storage (X-Admin-Token), body: block_id, slot_key, slot_value; создаёт транзакцию `contract_storage_write` |

## Источники данных на ноде

- **Контракт (состояние «функций»):** таблицы `contracts` (name, symbol, owner), `tokens` (decimals, total_supply, standard), `token_balances` (balances по token_id + address).
- **Текущий аккаунт:** таблица `accounts` (nonce, balance_gnd). Запись появляется после обработки транзакций/блоков.
- **Снимки на блок:** `account_states` (balance_gnd), `contract_storage` (slot_key, slot_value).

## GND_admin: страница контракта

- **Прочитать контракт:** данные с ноды (contract state + account state + storage по block_id). Ошибки разделены: контракт не найден / аккаунт не найден / ошибка storage.
- **Записать контракт:** форма записи слота storage (block_id, slot_key, slot_value); подтверждение через SweetAlert2.
- Пустой `block_id` в URL приводит к редиректу на страницу без query. Форма «Прочитать storage» требует заполненный ID блока (required, min=1).

## Требования к БД ноды

- Миграции 014 (account_states, contract_storage, balance_gnd в accounts) и 015 (переименование balance_wei → balance_gnd в account_states).
- В `contracts` должна быть запись с `address` контракта (после деплоя через ноду или импорта).
- Запись в `accounts` по адресу контракта появляется при сохранении состояния (SaveToDB при обработке блоков).

## Типичная причина 404

Сообщение «Контракт не найден на ноде (404)» возвращается, если по адресу контракта нет строки в таблице `contracts` на ноде. Решение: задеплоить контракт через админку (деплой на ноду) или убедиться, что адрес в GND_admin совпадает с адресом в БД ноды.
