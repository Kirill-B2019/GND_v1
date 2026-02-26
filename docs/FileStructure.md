# Архитектура и структура файлов блокчейна «ГАНИМЕД»

> **Единое описание структуры для IDE:** актуальное дерево каталогов и краткое назначение модулей хранятся в [.cursor/PROJECT_STRUCTURE.md](../.cursor/PROJECT_STRUCTURE.md). При изменении структуры проекта обновляйте и этот файл, и указанный.

## Общая структура проекта

Ниже — актуальная структура каталогов и ключевых файлов (синхронизирована с `.cursor/PROJECT_STRUCTURE.md`).

```
GND_v1/
├── main.go, main_test.go
├── go.mod, go.sum
├── README.md
├── config/
│   ├── config.json
│   ├── db.json
│   ├── consensus.json
│   ├── evm.json
│   ├── coins.json
│   ├── servers.json
│   └── req.conf
├── core/
│   ├── block.go, blockchain.go, config.go, pool.go, state.go
│   ├── transaction.go, mempool.go, wallet.go, account.go
│   ├── contract.go, token.go, event.go, events.go
│   ├── address.go, fees.go, interfaces.go, logger.go, utils.go, metrics.go
│   ├── wallet_test.go
│   └── crypto/keys.go
├── types/
│   ├── 00_address.go, state.go, token.go, evm.go, events.go
├── consensus/
│   ├── consensus.go, manager.go, poa.go, pos.go
├── api/
│   ├── rest.go, rpc.go, websocket.go, middleware.go, types.go, constants.go
│   ├── api_test.go, api_token_test.go, api_wallet_test.go
│   └── middleware/gin.go, middleware.go
├── tokens/
│   ├── types.go, metadata.go
│   ├── interfaces/token.go
│   ├── types/token.go
│   ├── registry/registry.go, registry_test.go
│   ├── deployer/deployer.go, compiler.go
│   ├── handlers/balance.go, info.go
│   ├── standards/gndst1/ (gndst1.go, тесты, abi, sol)
│   └── utils/helpers.go, events.go
├── vm/
│   ├── evm.go, contracts.go, sandbox.go, cache.go, events.go, integration.go
│   └── compiler/compiler.go
├── integration/
│   ├── address.go, bridges.go, ipfs.go, oracles.go
├── monitoring/
│   ├── metrics.go, events.go, alerts.go
├── audit/
│   ├── audit_report.go, monitor.go, rules.go, security_checks.go
│   ├── integration_test.go, README.md, external_tools.md
│   └── examples/sample_report.md
├── utils/
│   ├── handlers.go, handlers_test.go
├── db/
│   ├── db.sql, console_21.sql, dump0906.sql
│   └── migrations/001_create_events_table.sql
└── docs/
    ├── FileStructure.md (этот файл), README.md, architecture.md
    ├── api.md, consensus.md, contracts.md, database.md, events.md
    ├── integration.md, tokens.md, websocket_api.md, GNDst-1.md, wallwt.md
    └── arhitech_gnd_step1.md, diagramDB.drawio, diagram.png
```

---

## Описание основных модулей и их взаимодействия

### **main.go**
- Точка входа. Инициализирует конфиг, кошелек, генезис-блок, блокчейн, консенсус, mempool, API и сервисы.
- Управляет graceful shutdown (корректное завершение работы узла).
- Запускает REST и WebSocket серверы, консенсусный механизм.  
*Источник: main.go*

---

### **core/**
- **block.go, blockchain.go** — структуры блоков, логика построения цепи, добавление и валидация блоков.
- **state.go** — текущее состояние сети (балансы, nonce, стейкинг), работа с БД.
- **pool.go** — инициализация пула PostgreSQL (InitDBPool, pgxpool).
- **wallet.go** — генерация и загрузка кошельков, работа с приватными ключами.
- **transaction.go, mempool.go** — обработка транзакций, хранение неподтверждённых транзакций.
- **account.go, contract.go, token.go, event.go, events.go** — аккаунты, контракты, токены, события (с доступом к БД).
- **address.go, interfaces.go** — адреса, интерфейсы BlockchainIface, StateIface.
- **config.go** — загрузка и парсинг конфигурации (в т.ч. DBConfig).
- **fees.go** — расчёт и применение комиссий.
- **logger.go, utils.go, metrics.go** — логирование, утилиты, метрики.
- **crypto/keys.go** — криптографические ключи.

**Взаимодействие:**  
`main.go` использует методы из `core` для создания блокчейна, управления кошельками, обработки транзакций и состояния.

---

### **consensus/**
- **consensus.go** — базовые интерфейсы консенсуса.
- **pos.go, poa.go** — реализация Proof-of-Stake и Proof-of-Authority.
- **manager.go** — управление валидаторами, переключение алгоритмов.

**Взаимодействие:**  
Использует данные из `core` для финализации блоков, выбора валидаторов, работы с транзакциями.

---

### **api/**
- **rest.go** — REST API для доступа к блокам, отправки транзакций, получения информации.
- **rpc.go** — JSON-RPC API для работы с контрактами и токенами.
- **websocket.go** — WebSocket сервер для real-time событий (новые блоки, транзакции).
- **middleware.go** — подключение middleware; **middleware/** (gin.go, middleware.go) — аутентификация, лимитирование, аудит.
- **types.go, constants.go** — типы и константы API.

**Взаимодействие:**  
API обращается к методам `core` и консенсуса, предоставляет внешний интерфейс для пользователей, кошельков, dApp.  
*Источники: rest.go, rpc.go, websocket.go, middleware.go*

---

### **tokens/**
- **types.go, metadata.go** — типы и метаданные токенов.
- **interfaces/token.go** — интерфейсы токенов.
- **registry/** — реестр токенов (registry.go, тесты).
- **deployer/** — деплой и компиляция контрактов токенов.
- **handlers/** — обработчики баланса и информации по токенам (balance.go, info.go).
- **standards/gndst1/** — стандарт GNDst-1 (gndst1.go, тесты, ABI, Solidity).
- **utils/** — хелперы и события.

**Взаимодействие:**  
Токены регистрируются и управляются через API и ядро, используются в смарт-контрактах и пользовательских операциях.

---

### **vm/**
- **evm.go, contracts.go, sandbox.go** — EVM, контракты, изолированное выполнение.
- **cache.go, events.go, integration.go** — кэш, события, интеграция с ядром.
- **compiler/compiler.go** — компиляция контрактов.

**Взаимодействие:**  
Исполнение смарт-контрактов, интеграция с core и types.

---

### **integration/**
- **bridges.go, oracles.go** — кроссчейн-мосты и оракулы для обмена с другими сетями.
- **ipfs.go** — интеграция с IPFS для хранения данных.
- **address.go** — стандартизация адресов для внешних интеграций.

**Взаимодействие:**  
Обеспечивает взаимодействие с внешними блокчейнами, хранение и верификацию данных вне сети.

---

### **monitoring/** и **audit/**
- **metrics.go, events.go, alerts.go** — сбор метрик, событий, алерты.
- **audit/** — аудит безопасности, логи, отчеты.

**Взаимодействие:**  
Модули мониторинга интегрируются с ядром и API, фиксируют события для анализа и безопасности.

---

### **db/**
- SQL-скрипты (db.sql, дампы), миграции (migrations/).

**Взаимодействие:**  
Используются при инициализации и обновлении схемы PostgreSQL (см. docs/database.md, core/pool.go).

---

### **docs/**
- Документация по архитектуре, API, контрактам, токенам, консенсусу и интеграциям.

---

## Взаимосвязь компонентов

- **main.go** — точка входа, инициализация всех ключевых модулей, запуск сервисов.
- **core ↔ consensus** — ядро предоставляет данные для консенсуса, консенсус финализирует блоки.
- **core, consensus ↔ api** — API обращается к методам ядра и консенсуса для выполнения внешних запросов.
- **core, vm, tokens ↔ contracts** — ядро и VM обеспечивают исполнение и регистрацию смарт-контрактов.
- **integration ↔ core, api** — интеграционные сервисы используют ядро и API для кроссчейн-операций.
- **monitoring, audit ↔ все модули** — собирают метрики, логи, аудит действий для анализа и безопасности.

---

## Принципы модульной архитектуры

> Модульная архитектура позволяет разделить обязанности между слоями и модулями: выполнение транзакций, консенсус, хранение данных, интеграция, мониторинг и пользовательские интерфейсы. Это обеспечивает масштабируемость, гибкость и безопасность сети.

---

## Рекомендации для новых разработчиков

- Актуальное дерево структуры — в **.cursor/PROJECT_STRUCTURE.md** (единое описание для IDE); при изменении структуры обновляйте его и этот файл.
- Начинайте с **README.md** и документации в **docs/** — там описаны архитектура, API, принципы работы.
- Изучите **main.go** для понимания порядка инициализации и запуска системы.
- Для разработки смарт-контрактов — смотрите **tokens/standards/gndst1/** и **vm/**.
- Для интеграции с внешними системами — смотрите **integration/** и соответствующие разделы документации.
- Для клиентских сценариев — подключайтесь к ноде **main-node.gnd-net.com**; описание API (документация): **api.gnd-net.com**, **docs/api.md**.

---

## Итог

Структура проекта и взаимодействие модулей в «ГАНИМЕД» организованы по лучшим практикам современной блокчейн-разработки, что обеспечивает удобство масштабирования, тестирования и поддержки.  
*Подробнее в README.md*

---

**Ссылки на ключевые документы:**
- [api.md](api.md) — описание API (документация: api.gnd-net.com; подключение к ноде: main-node.gnd-net.com)
- [contracts.md](contracts.md) — описание работы со смарт-контрактами
- [architecture.md](architecture.md) — архитектура блокчейна
- [consensus.md](consensus.md) — алгоритмы консенсуса
- [integration.md](integration.md) — интеграция с GND
