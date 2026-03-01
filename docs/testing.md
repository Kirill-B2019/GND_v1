# Тестирование

Подход к тестам в проекте ГАНИМЕД.

## Запуск тестов

Из корня проекта:

```bash
go test ./... -count=1
```

Краткий сценарий (без длительных проверок БД):

```bash
go test ./... -count=1 -short
```

Скрипты: `scripts/run_tests.bat`, `scripts/run_tests.ps1`.

## Структура

- Модульные тесты в пакетах (`*_test.go`).
- Интеграционные проверки с БД — в `main_test.go` (требуют `config/db.json`).

## Тесты контрактов

- **api/rest_contract_test.go** — REST API контрактов (Gin): `POST /api/v1/contract/compile` (пустой/неверный JSON), `POST /api/v1/contract/analyze` (пустой source, валидный source с возвратом замечаний, неверный JSON).
- **audit/integration_test.go** — интеграционный тест: блокчейн, кошелёк, деплой контракта через EVM (требует `config/db.json` и БД).
- **api/api_test.go** — тесты RPC-хендлеров контрактов (DeployContract, CallContract, SendContractTx).

Запуск только тестов, связанных с контрактами:

```bash
go test ./api/... -run "Contract|Compile|Analyze" -v -count=1
go test ./audit/... -v -count=1
```

## Перезапуск ноды без обнуления БД

Тест **TestNodeRestartWithoutDBReset** (main_test.go) проверяет сценарий:

1. Подключение к БД, сохранение генезиса (если блока 0 ещё нет).
2. Закрытие пула (имитация остановки ноды).
3. Новое подключение к той же БД (имитация запуска ноды).
4. Загрузка блокчейна через `LoadBlockchainFromDB` и проверка, что генезис на месте.

Запуск:

```bash
go test . -run TestNodeRestartWithoutDBReset -v -count=1
```

Требуется доступный PostgreSQL и `config/db.json` в корне проекта.

## Тесты по ТЗ (модульная платформа, подсети)

- **api/rest_contract_test.go** — `TestHealthCheck_ReturnsChainIdAndNetworkId`: проверка, что GET `/api/v1/health` возвращает `network_id`, `chain_id`, `subnet_id` при заданном конфиге.
- **consensus/consensus_test.go** — `TestSelectConsensusForTx_WithRules`: загрузка правил из конфига (`selection_rules`) и выбор консенсуса по префиксу адреса и правилу по умолчанию.

Запуск:

```bash
go test ./consensus/... ./api/... -run "TestSelectConsensus|TestHealthCheck_ReturnsChainId|TestCompileContract|TestAnalyzeContract" -v -count=1
```

См. также [spec-modular-platform.md](spec-modular-platform.md).

## Эмуляция остановки и перезапуска ноды

Скрипт **scripts/emulate_restart.ps1** (Windows PowerShell) проводит полную эмуляцию:

1. Сборка ноды.
2. Запуск ноды в фоне, ожидание ответа REST `/api/v1/health`.
3. Остановка процесса (эмуляция выключения).
4. Пересборка бинарника (эмуляция «обновлённой» версии).
5. Повторный запуск ноды (БД не обнуляется).
6. Проверка health, остановка процесса.

Запуск из корня GND_v1:

```powershell
.\scripts\emulate_restart.ps1
```

Требования: Go в PATH, PostgreSQL (config/db.json), свободные порты 8181, 8182, 8183. После эмуляции тестовый бинарник `gnd_node_emulate.exe` удаляется.

## Связанные разделы

- [Развертывание](deployment-server.md) — раздел «Перед продом: запуск тестов»
- [services.md](services.md) — описание сервисов
