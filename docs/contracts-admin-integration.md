# Интеграция контрактов: GND_v1 ↔ GND_admin

Краткая проверка связки ноды и админ-панели по работе с контрактами (чтение/запись состояния, storage).

## Жизненный цикл контракта

Все контракты проходят: **проверка (компиляция) → анализ → деплой**. После деплоя чтение и запись выполняются через интерфейс GND_admin и API ноды.

- **Проверка/компиляция:** исходный код компилируется (POST /contract/compile на ноде), сохраняется bytecode.
- **Анализ:** опциональный шаг анализа безопасности (POST /contract/analyze).
- **Деплой:** создаёт контракт на ноде и запись в таблице `contracts`; **формирует транзакцию** в блокчейне (тип `contract_deploy`, запись в `transactions`).
- **Чтение:** вкладка «Прочитать контракт» и API (GET /contract/:address/state, GET /state/contract/:address/storage и т.д.).
- **Запись:** вкладка «Записать контракт» (слот storage); **формирует транзакцию** (тип `contract_storage_write` в `transactions`).

**Все действия с контрактами, токенами и кошельками формируют транзакции в блокчейне** (таблица `transactions` на ноде):

**Кошельки:**
- **Создание кошелька** — тип `wallet_create` (RecordAdminTransaction при POST /wallet).
- **Блокировка кошелька** — тип `wallet_disable` (при POST /api/v1/admin/wallets/:address/disable).
- **Снятие блокировки** — тип `wallet_enable` (при POST /api/v1/admin/wallets/:address/enable).
- **Удаление кошелька** — тип `wallet_delete` (при DELETE или POST /api/v1/admin/wallets/:address/delete).

**Контракты:**
- **Деплой контракта** — тип `contract_deploy` (RecordAdminTransaction при POST /contract).
- **Запись слота storage** (админ) — тип `contract_storage_write` (при POST /api/v1/admin/state/contract/:address/storage).
- **Блокировка контракта** — тип `contract_disable` (при POST /api/v1/admin/contracts/:address/disable).
- **Удаление контракта** — тип `contract_delete` (при POST /api/v1/admin/contracts/:address/delete).
- **Вызов методов контракта** (transfer, approve и т.д.) — тип `contract_call` (processContract при POST /contract/:address/send).

**Токены:**
- **Деплой токена** — тип `token_deploy` (RecordAdminTransaction при POST /token/deploy).
- **Блокировка токена** — тип `token_disable` (при POST /api/v1/admin/tokens/:id/disable).
- **Удаление токена** — тип `token_delete` (при POST /api/v1/admin/tokens/:id/delete).

## Эндпоинты GND_v1 (REST API)

| Метод | Путь | Назначение |
|-------|------|------------|
| GET | `/api/v1/contract/:address` | Информация о контракте (core.GetContract) |
| GET | `/api/v1/contract/:address/view` | Просмотр контракта: исходный код (Solidity), ABI, список методов контракта (view_functions, write_functions из ABI) |
| GET | `/api/v1/contract/:address/state?addresses=addr1,addr2` | Состояние контракта: name, symbol, owner, decimals, total_supply, balances (core.GetContractState) |
| POST | `/api/v1/contract/:address/call` | Вызов view/constant метода без транзакции. Body: `data` (hex calldata), опционально `from` |
| POST | `/api/v1/contract/:address/send` | Отправка транзакции вызова метода (transfer, approve и т.д.). Body: `from`, `data`, опционально `value`, `gas_limit` |
| GET | `/api/v1/state/account/:address` | Текущее состояние аккаунта из `accounts`: nonce, balance_gnd (core.GetCurrentAccountState) |
| GET | `/api/v1/state/account/:address/block/:blockId` | Снимок состояния на блок из `account_states` |
| GET | `/api/v1/state/contract/:address/storage?block_id=N` | Слоты storage контракта на блок (core.GetContractStorageAtBlock) |
| POST | `/api/v1/admin/state/contract/:address/storage` | Запись слота storage (X-Admin-Token), body: block_id, slot_key, slot_value; создаёт транзакцию `contract_storage_write` |

## Источники данных на ноде

- **Контракт (состояние «функций»):** таблицы `contracts` (name, symbol, owner), `tokens` (decimals, total_supply, standard), `token_balances` (balances по token_id + address).
- **Текущий аккаунт:** таблица `accounts` (nonce, balance_gnd). Запись появляется после обработки транзакций/блоков.
- **Снимки на блок:** `account_states` (balance_gnd), `contract_storage` (slot_key, slot_value).

## GND_admin: страница контракта

Просмотр, чтение и запись относятся к **методам и функциям самого контракта** (которые объявлены в его коде на Solidity). Три вкладки:

- **Просмотр контракта:** отображается **исходный код контракта на Solidity** (с ноды или из карточки контракта), адрес, список **методов контракта** из ABI (view/pure и write), ABI (JSON). Всё это — данные самого контракта, а не «базовая инфо из БД».
- **Чтение контракта:** вызов **view/pure методов контракта** (функции внутри контракта): name(), symbol(), decimals(), totalSupply(), balanceOf(address), allowance(owner, spender) и т.д. Без газа и без изменения состояния. Выбор метода → ввод аргументов → «Вызвать метод контракта» → ответ (return_data).
- **Запись в контракт:** вызов **методов контракта**, меняющих состояние: transfer(), approve(), transferFrom() и т.д. Выбор метода → аргументы → выбор кошелька (from) → «Отправить транзакцию».

Кодирование calldata (ABI) выполняется на клиенте (ethers.js); запросы к ноде идут через Laravel (POST /admin/contracts/:id/call и /admin/contracts/:id/send). GET /api/v1/contract/:address/view возвращает в т.ч. source_code (код на Solidity) и списки view_functions / write_functions из ABI контракта.

## Требования к БД ноды

- Миграции 014 (account_states, contract_storage, balance_gnd в accounts) и 015 (переименование balance_wei → balance_gnd в account_states).
- В `contracts` должна быть запись с `address` контракта (после деплоя через ноду или импорта).
- Запись в `accounts` по адресу контракта появляется при сохранении состояния (SaveToDB при обработке блоков).

### Таблица `gnd_db.public.transactions`

- Таблица **партиционирована по RANGE (timestamp)**. Для каждой записи должна существовать партиция на месяц `timestamp` (например `transactions_2026_03`). Если записей нет — выполните скрипт создания партиций: `db/create_transactions_partitions.sql`.
- Колонка `payload` имеет тип **jsonb**; код передаёт `NULL` или JSON-совместимое значение. Ошибки записи логируются в stdout ноды (`[REST] запись транзакции ... в gnd_db.transactions: ...`).
- Пул БД для записи: при наличии `s.core.Pool` используется он, иначе `s.db` (оба должны указывать на **gnd_db**).

## Типичная причина 404

Сообщение «Контракт не найден на ноде (404)» возвращается, если по адресу контракта нет строки в таблице `contracts` на ноде. Решение: задеплоить контракт через админку (деплой на ноду) или убедиться, что адрес в GND_admin совпадает с адресом в БД ноды.
