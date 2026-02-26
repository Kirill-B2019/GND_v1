# Запуск ноды ГАНИМЕД на сервере

## 1. Запуск без падения по окончании сессии

Чтобы процесс не завершался при закрытии SSH, используйте один из способов.

### Вариант A: nohup

```bash
cd /path/to/GND_v1
nohup ./gnd-node > logs/node.log 2>&1 &
echo $! > gnd-node.pid
```

Остановка: `kill $(cat gnd-node.pid)`

### Вариант B: screen

```bash
screen -S gnd
cd /path/to/GND_v1
./gnd-node
# Отсоединиться: Ctrl+A, затем D
# Подключиться снова: screen -r gnd
```

### Вариант C: tmux

```bash
tmux new -s gnd
cd /path/to/GND_v1
./gnd-node
# Отсоединиться: Ctrl+B, затем D
# Подключиться: tmux attach -t gnd
```

### Вариант D: systemd (рекомендуется для продакшена)

См. раздел «Автозагрузка при перезагрузке сервера» ниже.

---

## 2. Проверка API

После запуска ноды проверьте, что API отвечает.

Подставьте свой хост (например `31.128.41.155`) и при необходимости порты из `config/servers.json` (REST: 8182, RPC: 8181, WS: 8183). Документация по API (описание эндпоинтов): **api.gnd-net.com**. Подключение клиентов — к ноде **main-node.gnd-net.com**.

### REST API

```bash
# Health
curl -s http://31.128.41.155:8182/api/v1/health

# Баланс кошелька (подставьте адрес валидатора)
curl -s "http://31.128.41.155:8182/api/v1/wallet/WALLET_ADDRESS/balance"

# Мемпул (размер и список хешей ожидающих транзакций)
curl -s http://31.128.41.155:8182/api/v1/mempool

# Последний блок
curl -s http://31.128.41.155:8182/api/v1/block/latest
```

### RPC (если включён)

```bash
curl -s -X POST http://31.128.41.155:8181 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}'
```

Успешный ответ по health обычно содержит `"success": true` и данные о состоянии сервиса.

---

## 3. Автозагрузка при перезагрузке сервера (systemd)

Создайте unit-файл и включите автозапуск.

### Файл сервиса

Создайте `/etc/systemd/system/gnd-node.service` (через `sudo`):

```ini
[Unit]
Description=GND blockchain node
After=network.target postgresql.service

[Service]
Type=simple
User=root
WorkingDirectory=/root/GND_v1
ExecStart=/root/GND_v1/gnd-node
Restart=on-failure
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

Измените `WorkingDirectory` и `ExecStart` на фактический путь к проекту и бинарнику (например, если бинарник в `GND_v1` и называется `gnd-node` или `GND_v1` — укажите его).

### Включение и управление

```bash
# Перезагрузить конфигурацию systemd
sudo systemctl daemon-reload

# Включить автозапуск при загрузке ОС
sudo systemctl enable gnd-node

# Запустить сейчас
sudo systemctl start gnd-node

# Статус
sudo systemctl status gnd-node

# Логи
journalctl -u gnd-node -f
```

После перезагрузки сервера нода поднимется автоматически.

---

## 4. Баланс аккаунта и несколько кошельков

- Балансы по монетам хранятся в `token_balances` (по `token_id` и адресу). Поле `accounts.balance` дублирует баланс нативной монеты (GND) для совместимости.
- При загрузке состояния из БД (`LoadFromDB`) значение `accounts.balance` синхронизируется с балансом GND из `token_balances`, поэтому после перезапуска баланс в таблице `accounts` не остаётся нулевым.
- У одного аккаунта может быть несколько кошельков (разные ключи/адреса); валидатор использует один загруженный кошелёк (`LoadWallet`). Остальные адреса также могут иметь записи в `accounts` и `token_balances`.

---

## 5. Монеты при первом запуске и перезапуске

- При **первом запуске** (когда в БД ещё нет генезис-блока): создаются 2 монеты и 2 контракта по данным из `config/coins.json`, затем начисляются балансы валидатору.
- При **перезапуске**: проверяется наличие генезис-блока. Если он есть, нода только загружает блокчейн и состояние из БД; **генерация монет не запускается** — в `EnsureCoinsDeployed` для каждого символа из config проверяется наличие токена в БД (`GetTokenBySymbol`), и при наличии создание пропускается.

---

## 6. Проверка работы mempool

- В коде: транзакции попадают в мемпул через `blockchain.AddTx`; воркеры в `processTransactions` забирают их через `mempool.Pop()` и обрабатывают (PoA/PoS).
- Через API: запрос `GET /api/v1/mempool` возвращает `size` (число транзакций в очереди) и `pending_hashes` (список хешей). Отправка транзакции через `POST /api/v1/transaction` добавляет её в мемпул; после этого при повторном запросе к `/api/v1/mempool` размер должен увеличиться, затем уменьшиться после обработки.

Пример:

```bash
# До отправки транзакции
curl -s http://31.128.41.155:8182/api/v1/mempool

# После отправки (если есть эндпоинт отправки) — снова проверить mempool
curl -s http://31.128.41.155:8182/api/v1/mempool
```

---

## Перед продом: запуск тестов

Перед выкладкой в прод рекомендуется прогнать все тесты из корня проекта:

```bash
cd /path/to/GND_v1

# Короткие тесты (без длительных/интеграционных)
go test ./... -count=1 -short

# Полный прогон (включая интеграционные при доступной БД)
go test ./... -count=1
```

**Windows (из корня репозитория):**

```bat
scripts\run_tests.bat
```

или PowerShell:

```powershell
.\scripts\run_tests.ps1
```

Тесты, требующие БД (например `main_test.TestNewBlockchain`, `api_wallet_test`, `audit.TestBlockchainIntegration`), при недоступном PostgreSQL пропускаются (`t.Skip`).

---

## Go не найден в PATH

Если при запуске тестов или сборке появляется ошибка **«go не найден»** / **«go is not recognized»**, добавьте Go в переменную окружения PATH.

### Установка Go (если ещё не установлен)

- **Windows:** скачайте установщик с [go.dev/dl](https://go.dev/dl/), запустите и отметьте опцию «Add to PATH» (или добавьте вручную, см. ниже).
- **Linux (Ubuntu/Debian):**  
  `sudo apt update && sudo apt install -y golang-go`  
  или установите вручную с [go.dev/dl](https://go.dev/dl/) и добавьте в PATH.

### Добавление Go в PATH вручную

**Windows (постоянно для пользователя):**

1. Узнайте каталог установки Go (часто `C:\Program Files\Go\bin` или `C:\Go\bin`).
2. Панель управления → Система → Дополнительные параметры системы → Переменные среды.
3. В «Переменные среды пользователя» выберите `Path` → Изменить → Создать → укажите путь к папке `bin` (например `C:\Program Files\Go\bin`).
4. OK во всех окнах. **Перезапустите терминал** (или Cursor/IDE).

**Windows (одна сессия, PowerShell):**

```powershell
$env:Path += ";C:\Program Files\Go\bin"
go version
```

**Linux / macOS (в профиле оболочки):**

Если Go установлен, например, в `/usr/local/go`:

```bash
export PATH=$PATH:/usr/local/go/bin
```

Чтобы сделать это постоянным, добавьте эту строку в `~/.bashrc`, `~/.profile` или `~/.zshrc` и выполните `source ~/.bashrc` (или перезайдите в терминал).

**Проверка:** в новом терминале выполните `go version` — должна вывестись версия Go.
