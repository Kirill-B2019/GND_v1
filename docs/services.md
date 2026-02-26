# Реализованные сервисы ГАНИМЕД

Документ описывает все реализованные компоненты проекта: ядро, API, мониторинг, аудит, консенсус, VM и интеграции.

---

## 1. Ядро (core)

**Назначение:** управление блокчейном, состоянием, транзакциями и кошельками.

| Компонент | Описание |
|-----------|----------|
| **Blockchain** | Цепочка блоков, генезис, загрузка/сохранение из БД, FirstLaunch (деплой монет, начисление балансов), системные транзакции. |
| **Block** | Структура блока (Hash, PrevHash, Timestamp, Miner, Consensus, Index, Transactions), сохранение/загрузка из PostgreSQL. |
| **State** | Балансы по адресам и токенам (GND, GANI и др.), nonce, token_balances, синхронизация с БД (LoadFromDB, SaveToDB), ApplyTransaction, ApplyExecutionResult. |
| **Transaction** | Транзакция (Sender, Recipient, Value, Fee, Nonce, Hash, Type, Status), валидация, подпись, сохранение в БД (в т.ч. партиционированная таблица). |
| **Mempool** | Очередь ожидающих транзакций (Add, Pop, GetPendingTransactions, Exists, GetTransaction). |
| **Wallet** | Создание кошелька (NewWallet), загрузка из БД (LoadWallet), адрес и ключи. |
| **Token** | Токены в БД (GetTokenBySymbol, SaveToDB), прокси для стандарта GND-st1 (IsGNDst1, GNDst1Instance, UniversalCall). |
| **Contract** | Контракты (SaveToDB, загрузка по адресу), ContractParams для деплоя. |
| **Config** | Глобальная конфигурация (InitGlobalConfigDefault), NodeName, DB, Coins, Consensus, EVM, Server, MaxWorkers. |
| **Metrics** | Метрики блоков, транзакций, комиссий, алерты (GetMetrics, UpdateBlockMetrics, UpdateTransactionMetrics, SetAlertThresholds). |
| **Pool / InitDBPool** | Пул подключений PostgreSQL (pgxpool). |
| **Crypto** | Ключи и подпись (HexToPrivateKey, Sign). |

**Логика запуска (main.go):** загрузка конфига → инициализация БД → проверка генезис-блока и аккаунтов → создание/загрузка кошелька валидатора → создание или загрузка блокчейна из БД → установка глобального State → EVM → при первом запуске FirstLaunch (монеты, балансы, системные транзакции) → мемпул → запуск REST, RPC, WebSocket → воркеры обработки транзакций из мемпула (PoA/PoS) → мониторинг пула БД.

---

## 2. API (api)

**Назначение:** REST, RPC и WebSocket для клиентов.

| Сервис | Порт | Описание |
|--------|------|----------|
| **REST API** | 8182 | Gin: `/api/v1/health`, `/api/v1/metrics`, `/api/v1/metrics/transactions`, `/api/v1/metrics/fees`, `/api/v1/alerts`, `/api/v1/wallet` (POST), `/api/v1/wallet/:address/balance`, `/api/v1/transaction` (POST), `/api/v1/transaction/:hash`, `/api/v1/mempool`, `/api/v1/block/latest`, `/api/v1/block/:number`, `/api/v1/contract` (POST/GET), **`/api/v1/token/deploy`** (POST, **обязателен X-API-Key** — создание и регистрация токена для внешних систем), `/api/v1/token/transfer`, `/api/v1/token/approve`, `/api/v1/token/:address/balance/:owner`. Ответы в формате `{ success, data, error, code }`. |
| **RPC API** | 8181 | HTTP: `/block/latest`, `/contract/deploy`, `/contract/call`, `/contract/send`, `/account/balance`, `/block/by-number`, `/tx/send`, `/tx/status`, `/token/universal-call`. CORS и заголовки безопасности. |
| **WebSocket** | 8183 | Подписки на события (блоки, транзакции), аутентификация по API ключу. |

**Дополнительно:** **auth.go** — `ValidateAPIKey` (константа или таблица `api_keys`), **evm_adapter.go** — приведение EVM к интерфейсу для deployer, **eventmanager_stub.go** — заглушка EventManager для деплоера; middleware (CORS, X-API-Key); константы (RestURL, RpcURL, WsURL, NodeHost, ApiDocHost, TokenStandardGNDst1). Подробная логика создания токена через API: **docs/api-token-deploy.md**.

---

## 3. Консенсус (consensus)

**Назначение:** выбор и параметры консенсуса для транзакций.

| Компонент | Описание |
|-----------|----------|
| **PoA** | InitPoaConsensus, RoundDuration, SyncDuration, BanDurationBlocks, WarningsForBan, MaxBansPercentage. |
| **PoS** | Параметры (AverageBlockDelay, InitialBaseTarget, InitialBalance). |
| **SelectConsensusForTx** | Выбор консенсуса по получателю транзакции (ConsensusPoA / ConsensusPoS). |

Обработка транзакций в main: processPoATransaction / processPoSTransaction (валидация, баланс, контракт или перевод).

---

## 4. Виртуальная машина (vm)

**Назначение:** исполнение смарт-контрактов и деплой токенов.

| Компонент | Описание |
|-----------|----------|
| **EVM** | NewEVM (Blockchain, State, GasLimit, Coins), DeployContract, CallContract, конфиг монет. |
| **Контракты** | TokenContract (балансы, стандарт GND-st1), компиляция (SolidityCompiler), ValidateContract (GND-st1, erc20, trc20). |
| **Integration** | DeployGNDst1Token (генерация байткода, деплой, регистрация токена в registry, событие TokenDeployed). |

---

## 5. Токены (tokens)

**Назначение:** стандарты токенов и реестр.

| Компонент | Описание |
|-----------|----------|
| **GND-st1** | Пакет `tokens/standards/gndst1`: тип GNDst1, Transfer, Approve, GetBalance, Allowance, TransferFrom, CrossChainTransfer, Snapshot, Dividends, ModuleCall, GetStandard() = "GND-st1". |
| **Registry** | RegisterToken, GetToken по адресу, хранение *gndst1.GNDst1. |
| **Handlers** | balance, info для токенов. |
| **Deployer** | **DeployToken** (генерация байткода, evm.DeployContract, событие Deploy), **registerToken** (GNDst1 + SetInitialBalance, registry.RegisterToken, запись в БД: contracts, tokens). Вызывается из REST **POST /api/v1/token/deploy** при валидном X-API-Key. |

---

## 6. Мониторинг (monitoring)

**Назначение:** метрики, алерты и события.

| Компонент | Описание |
|-----------|----------|
| **MetricsRegistry** | Gauge, Counter, Histogram; запись в файл (NewMetricsRegistry), IncCounter, SetGauge. |
| **AlertManager** | NewAlertManager(logFile), SendAlert(level, component, message, details), ListAlerts, кольцевой буфер алертов, фоновый processAlerts. |
| **EventLogger** | NewEventLogger(logFilePath, minLevel), логирование событий по уровням (DEBUG, INFO, WARN, ERROR, FATAL), Component, EventType, TxID, BlockID. |

В core используются глобальные метрики (GetMetrics, UpdateBlockMetrics, UpdateTransactionMetrics) и алерты (SetAlertThresholds, AlertHistory).

---

## 7. Аудит (audit)

**Назначение:** проверка транзакций и контрактов на подозрительную активность и уязвимости.

| Компонент | Описание |
|-----------|----------|
| **Monitor** | NewMonitor(threshold), CheckTransaction (крупная сумма, перевод самому себе, частые переводы, новый адрес), AddSuspicious, GetSuspicious, Clear. Защита от nil (Value, Sender/Recipient). |
| **AuditReport** | AuditFinding (Severity, Description, Location, Recommendation), AuditMetadata, NewAuditReport. |
| **Rules** | Rule (Blacklist, Whitelist, Limit, Pattern), DefaultRules, Blacklist/Whitelist списки. |
| **Security_checks** | CheckOverflow, CheckReentrancy, CheckPublicFunctions, CheckOwnerCheck, CheckDeprecatedUsage, CheckEvents, RunSecurityChecks(contract). |

---

## 8. Интеграция (integration)

**Назначение:** мосты, оракулы, адреса, IPFS.

| Компонент | Описание |
|-----------|----------|
| **Bridges** | Интеграция с внешними сетями. |
| **Oracles** | Оракулы для внешних данных. |
| **Address** | Конвертация/валидация адресов. |
| **IPFS** | Работа с метаданными (MetadataCID). |

---

## 9. База данных

**Хост/порт:** из config/db.json (по умолчанию 31.128.41.155:5432). Таблицы: accounts, wallets, blocks, transactions (партиционирована по времени), token_balances, tokens, contracts, states, events, api_keys, oracles, metrics, validators (poa_validators, pos_validators), logs. Миграции: db/002_schema_additions.sql, db/003_reset_database.sql, db/dump.sql.

---

## 10. Конфигурация

| Файл | Назначение |
|------|------------|
| config/config.json | Глобальный конфиг (если используется). |
| config/db.json | PostgreSQL: host, port, user, password, dbname, sslmode, max_conns, min_conns. |
| config/coins.json | Монеты (GND, GANI): name, symbol, decimals, total_supply, standard (GND-st1). |
| config/consensus.json | Параметры консенсуса. |
| config/servers.json | rpc_addr, rest.host/port, ws_addr (31.128.41.155, порты 8181, 8182, 8183). |

---

## Сводка по запуску

1. PostgreSQL доступен, применены миграции.
2. Запуск: `./gnd-node` или через systemd (см. docs/deployment-server.md).
3. REST: `http://31.128.41.155:8182/api/v1` или https://main-node.gnd-net.com/api/v1 (документация: api.gnd-net.com).
4. Документация по запросам: docs/api-requests.md, полное описание API: docs/api.md.
