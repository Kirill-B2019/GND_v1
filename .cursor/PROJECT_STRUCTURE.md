# Единое описание структуры проекта GND (для IDE)

**Назначение:** единственный актуальный источник структуры проекта. При добавлении/удалении пакетов или значимых файлов — обновлять этот файл и при необходимости `docs/FileStructure.md`.

**Обновлено:** по состоянию репозитория (актуальная структура каталогов и файлов).

---

## Дерево каталогов и файлов

```
GND_v1/
├── main.go
├── main_test.go
├── go.mod
├── go.sum
├── README.md
├── package.json
├── package-lock.json
├── .gitignore
│
├── .cursor/
│   ├── PROJECT_STRUCTURE.md    ← единое описание структуры (этот файл)
│   └── rules/
│       ├── project-conventions.mdc
│       ├── go-standards.mdc
│       ├── database.mdc
│       ├── server-deployment.mdc
│       └── unified-structure.mdc
│
├── .vscode/
│   └── settings.json
│
├── config/
│   ├── config.json      # основной конфиг ноды
│   ├── db.json          # параметры PostgreSQL
│   ├── consensus.json
│   ├── evm.json
│   ├── coins.json
│   ├── servers.json
│   └── req.conf
│
├── core/
│   ├── block.go
│   ├── blockchain.go
│   ├── config.go
│   ├── pool.go          # InitDBPool, pgxpool
│   ├── state.go
│   ├── transaction.go
│   ├── mempool.go
│   ├── wallet.go
│   ├── wallet_test.go
│   ├── account.go
│   ├── contract.go
│   ├── token.go
│   ├── event.go
│   ├── events.go
│   ├── address.go
│   ├── fees.go
│   ├── interfaces.go    # BlockchainIface, StateIface
│   ├── logger.go
│   ├── utils.go
│   ├── metrics.go
│   └── crypto/
│       └── keys.go
│
├── types/
│   ├── 00_address.go
│   ├── state.go
│   ├── token.go
│   ├── evm.go
│   └── events.go
│
├── consensus/
│   ├── consensus.go
│   ├── manager.go
│   ├── poa.go
│   └── pos.go
│
├── api/
│   ├── rest.go
│   ├── rpc.go
│   ├── websocket.go
│   ├── middleware.go
│   ├── types.go
│   ├── constants.go
│   ├── api_test.go
│   ├── api_token_test.go
│   ├── api_wallet_test.go
│   └── middleware/
│       ├── gin.go
│       └── middleware.go
│
├── tokens/
│   ├── types.go
│   ├── metadata.go
│   ├── interfaces/
│   │   └── token.go
│   ├── types/
│   │   └── token.go
│   ├── registry/
│   │   ├── registry.go
│   │   └── registry_test.go
│   ├── deployer/
│   │   ├── deployer.go
│   │   └── compiler.go
│   ├── handlers/
│   │   ├── balance.go
│   │   └── info.go
│   ├── standards/
│   │   └── gndst1/
│   │       ├── gndst1.go
│   │       ├── gndst1_test.go
│   │       ├── gndst1.abi.json
│   │       └── gndst1Base.sol
│   └── utils/
│       ├── helpers.go
│       └── events.go
│
├── vm/
│   ├── evm.go
│   ├── contracts.go
│   ├── sandbox.go
│   ├── cache.go
│   ├── events.go
│   ├── integration.go
│   └── compiler/
│       └── compiler.go
│
├── integration/
│   ├── address.go
│   ├── bridges.go
│   ├── ipfs.go
│   └── oracles.go
│
├── monitoring/
│   ├── metrics.go
│   ├── events.go
│   └── alerts.go
│
├── audit/
│   ├── audit_report.go
│   ├── monitor.go
│   ├── rules.go
│   ├── security_checks.go
│   ├── integration_test.go
│   ├── README.md
│   ├── external_tools.md
│   └── examples/
│       └── sample_report.md
│
├── utils/
│   ├── handlers.go
│   └── handlers_test.go
│
├── db/
│   ├── db.sql
│   ├── console_21.sql
│   ├── dump0906.sql
│   └── migrations/
│       └── 001_create_events_table.sql
│
└── docs/
    ├── README.md
    ├── FileStructure.md   # расширенное описание + взаимодействие модулей
    ├── architecture.md
    ├── api.md
    ├── consensus.md
    ├── contracts.md
    ├── database.md
    ├── events.md
    ├── integration.md
    ├── tokens.md
    ├── websocket_api.md
    ├── GNDst-1.md
    ├── wallwt.md
    ├── arhitech_gnd_step1.md
    ├── diagramDB.drawio
    └── diagram.png
```

---

## Назначение модулей (кратко)

| Модуль | Назначение |
|--------|------------|
| **main.go** | Точка входа: конфиг, пул БД, кошелёк, блокчейн, консенсус, API, graceful shutdown. |
| **config/** | Конфигурация ноды, БД, консенсуса, EVM, монет, серверов. |
| **core/** | Блоки, цепь, состояние, транзакции, mempool, кошелёк, аккаунты, контракты, события, комиссии, пул БД, интерфейсы. |
| **types/** | Общие типы: адреса, состояние, токены, EVM, события. |
| **consensus/** | PoA, PoS, менеджер консенсуса. |
| **api/** | REST, RPC, WebSocket, middleware (Gin), типы запросов/ответов. |
| **tokens/** | Реестр, деплой, стандарт GNDst-1, handlers баланса/инфо, метаданные, утилиты. |
| **vm/** | EVM, контракты, sandbox, кэш, компилятор, события, интеграция с core. |
| **integration/** | Адреса, мосты, IPFS, оракулы. |
| **monitoring/** | Метрики, события, алерты. |
| **audit/** | Отчёты, мониторинг, правила, проверки безопасности, примеры. |
| **utils/** | Хендлеры и тесты. |
| **db/** | SQL-скрипты и миграции. |
| **docs/** | Документация по архитектуре, API, БД, токенам и т.д. |

---

## Зависимости между пакетами (направление импортов)

- `main` → api, consensus, core, types, vm
- `api` → core, types (не наоборот)
- `core` → types
- `consensus` → core, types
- `tokens` → core, types
- `vm` → core, types
- `integration` → types (и при необходимости core)
- `monitoring`, `audit` — используют core/types по необходимости

Новые пакеты размещать в соответствии с этой схемой и добавлять в этот файл и в `docs/FileStructure.md`.
