# Детальный план реализации нативных монет GND и GANI (GND_v1)

| KB @CerberRus00 - Nexus Invest Team 2026

**Статус:** реализовано (миграция 012, core/native.go, state, blockchain, API, GND_admin, документация). Лимит циркулирующего предложения (жёсткая проверка в AddBalance) — реализован. Поддержан режим «всё на контрактах»: при задании адресов в `config/native_contracts.json` GND/GANI учитываются через `token_balances`; см. [deployment-contracts-variant-c.md](deployment-contracts-variant-c.md).

## 1. Цель

Сделать **GND** и **GANI** нативными активами L1: источник истины — состояние ноды; изменение балансов только через L1 (транзакции, газ, консенсус). Описание монет — из `config/coins.json`. Обеспечить защиту от кражи и сохранность при перезагрузке ноды.

---

## 2. Исходное состояние

- **config/coins.json** — массив `coins` с полями для GND и GANI (name, symbol, decimals, total_supply, circulating_supply и т.д.).
- **core/config.go** — тип `CoinConfig`, загрузка `cfg.Coins`.
- **core/state.go** — `State.balances[Address][symbol]`; загрузка из `token_balances` (JOIN tokens); сохранение в `token_balances (address, symbol, balance)` (схема частично расходится с token_id).
- **core/transaction.go** — тип `TxTypeTransfer`; газ и перевод только по символу `"GND"`; у `Transaction` есть поле `Symbol`.
- **core/blockchain.go** — при первом запуске начисление по `cfg.Coins` в state и запись в `token_balances (token_id, address, balance)` через `GetTokenBySymbol`.

---

## 3. План по шагам

### 3.1. База данных

| Шаг | Действие |
|-----|----------|
| 3.1.1 | Создать миграцию `db/migrations/012_native_balances.sql`: таблица `native_balances (address VARCHAR NOT NULL, symbol VARCHAR(10) NOT NULL, balance NUMERIC NOT NULL, updated_at TIMESTAMP, PRIMARY KEY (address, symbol))`, CHECK (symbol IN ('GND','GANI')). |
| 3.1.2 | Индексы: по address, по symbol. |

### 3.2. Конфигурация и список нативных монет

| Шаг | Действие |
|-----|----------|
| 3.2.1 | В `core/config.go` или новом `core/native.go`: константа или функция `NativeSymbols() []string` — возвращать символы из `cfg.Coins` (сейчас GND, GANI). Либо константа `NativeSymbols = []string{"GND","GANI"}` и проверка, что символ из этого списка. |
| 3.2.2 | Функция `IsNativeSymbol(symbol string) bool` для использования в state и tx. |

### 3.3. Слой State (core/state.go)

| Шаг | Действие |
|-----|----------|
| 3.3.1 | Добавить функции нативного слоя: `GetNativeBalance(address, symbol string) *big.Int`, `AddNativeBalance`, `SubNativeBalance` — внутри работа с `s.balances` только для разрешённых символов (GND, GANI). |
| 3.3.2 | В `LoadFromDB`: сначала загрузить нативные балансы из таблицы `native_balances` в `s.balances` для символов GND, GANI; затем загрузить контрактные из `token_balances` JOIN tokens (как сейчас). |
| 3.3.3 | В `SaveToDB`: для символов GND и GANI записывать в `native_balances` (INSERT/UPDATE по (address, symbol)); для остальных символов — не менять запись в native_balances (контрактные токены остаются в token_balances через другой путь). Либо разделить явно: при сохранении state сохранять нативные в native_balances, контрактные — по текущей логике token_balances. |
| 3.3.4 | ApplyTransaction: использовать `tx.Symbol` (если пусто — считать GND). Проверять `IsNativeSymbol(tx.Symbol)`; списание/начисление по `tx.Symbol`. Газ по-прежнему только в GND — проверять при tx.Symbol == "GND" достаточность balance GND для value + gas. |
| 3.3.5 | ApplyExecutionResult: газ списывать только в GND из нативного слоя. |

### 3.4. Транзакции (core/transaction.go)

| Шаг | Действие |
|-----|----------|
| 3.4.1 | HasSufficientBalance: для перевода использовать баланс по `tx.Symbol` (GetBalance(sender, tx.Symbol)); для газа — баланс GND. Итого: required = value + (если газ в GND) gasCost; проверять баланс по tx.Symbol для value и баланс GND для gas. |
| 3.4.2 | Валидация: при типе transfer проверять, что tx.Symbol входит в список нативных. |

### 3.5. Блокчейн и первый запуск (core/blockchain.go)

| Шаг | Действие |
|-----|----------|
| 3.5.1 | InitFirstRun: при начислении начального баланса по cfg.Coins использовать **circulating_supply** (не total_supply); записывать GND и GANI в таблицу `native_balances` и в state (Credit). Лимит циркулирующего предложения проверяется в AddBalance. |
| 3.5.2 | При применении блока: после ApplyTransaction вызывать сохранение нативных балансов в БД (в рамках той же транзакции БД, что и блок) — либо в SaveToDB по итогам блока, либо отдельный вызов PersistNativeBalances. |

### 3.6. Защита от кражи и сохранность

| Шаг | Действие |
|-----|----------|
| 3.6.1 | Изменение нативных балансов только через AddNativeBalance/SubNativeBalance из ApplyTransaction, ApplyExecutionResult (газ), InitFirstRun и консенсуса (награды). Никакой внешний API не вызывает эти функции напрямую. |
| 3.6.2 | При применении блока использовать транзакцию БД: запись блока + транзакций + обновление native_balances — один COMMIT. При ошибке — откат. |
| 3.6.3 | При старте ноды: LoadFromDB загружает native_balances первым; источник истины после перезапуска — БД. |

### 3.7. API (api/rest.go)

| Шаг | Действие |
|-----|----------|
| 3.7.1 | Эндпоинт баланса кошелька (GET /wallet/:address/balance): возвращать нативные GND и GANI из state (GetNativeBalance или GetBalance(addr, "GND"), GetBalance(addr, "GANI")) и контрактные токены как сейчас. |
| 3.7.2 | Перевод нативной монеты: если есть эндпоинт перевода с указанием символа — убедиться, что создаётся транзакция с полем Symbol (GND или GANI). |

### 3.8. Документация

| Шаг | Действие |
|-----|----------|
| 3.8.1 | Обновить docs/database.md — описание таблицы native_balances и роли нативных монет. |
| 3.8.2 | Обновить docs/architecture.md или docs/tokens.md — раздел про нативные GND/GANI и источник истины. |
| 3.8.3 | Кратко описать в docs/security.md защиту нативных балансов и персистентность. |

### 3.9. GND_admin (SigningService)

| Шаг | Действие |
|-----|----------|
| 3.9.1 | Убедиться, что запрос баланса кошелька к ноде возвращает нативные GND/GANI (эндпоинт уже отдаёт данные из state/token_balances — после изменений ноды будут отдаваться и из native_balances через state). При необходимости обновить отображение в админке (коин GND/GANI из списка coins). |
| 3.9.2 | Перевод нативной монеты с админки: если поддерживается выбор «монета GND/GANI», запрос к ноде должен передавать symbol и использовать корректный API. |

### 3.10. Документ «Работа и роли»

| Шаг | Действие |
|-----|----------|
| 3.10.1 | Создать docs/roles-and-operations.md (или аналог): описание ролей (нода, админ-панель, пользователь), операций (перевод нативной монеты, газ, награды), и где хранятся/как защищены нативные балансы. |

---

## 4. Критерии приёмки

- [ ] Таблица `native_balances` создана и используется только для GND и GANI.
- [ ] Список нативных символов задаётся из конфига/константы; в коде есть проверка IsNativeSymbol.
- [ ] LoadFromDB загружает нативные балансы из native_balances; SaveToDB (или отдельный persist) записывает нативные балансы в native_balances.
- [ ] ApplyTransaction использует tx.Symbol (GND или GANI); газ списывается только в GND.
- [ ] HasSufficientBalance учитывает баланс по tx.Symbol и баланс GND для газа.
- [ ] InitFirstRun записывает начальные GND/GANI в native_balances и state.
- [ ] Применение блока атомарно с записью нативных балансов в БД.
- [ ] Документация и документ о ролях/операциях обновлены.

---

## 5. Порядок реализации

1. Миграция 012_native_balances.sql  
2. core: NativeSymbols / IsNativeSymbol  
3. core/state.go: Get/Add/SubNativeBalance (обёртки над балансами с проверкой символа), LoadFromDB — загрузка native_balances, SaveToDB — сохранение нативных в native_balances  
4. core/transaction.go: HasSufficientBalance по tx.Symbol и GND для газа  
5. core/state.go: ApplyTransaction по tx.Symbol  
6. core/blockchain.go: InitFirstRun — запись в native_balances; при применении блока — сохранение нативных балансов  
7. API: проверить/добавить возврат нативных балансов в GET balance  
8. Документация GND_v1  
9. Изменения GND_admin  
10. Документ roles-and-operations  
