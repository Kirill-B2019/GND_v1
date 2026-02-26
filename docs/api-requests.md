# Памятка по запросам API (GND)

**Подключение к ноде:** **main-node.gnd-net.com** (документация по API: **api.gnd-net.com**)

Базовый URL REST API: `https://main-node.gnd-net.com/api/v1` или напрямую `http://31.128.41.155:8182/api/v1`

Формат ответа: `{ "success": true|false, "data": ..., "error": "текст", "code": число }`

---

## Здоровье и метрики

```bash
# Проверка работы API
curl -s "https://main-node.gnd-net.com/api/v1/health"

# Метрики
curl -s "https://main-node.gnd-net.com/api/v1/metrics"
curl -s "https://main-node.gnd-net.com/api/v1/metrics/transactions"
curl -s "https://main-node.gnd-net.com/api/v1/metrics/fees"

# Алерты
curl -s "https://main-node.gnd-net.com/api/v1/alerts"
```

---

## Кошельки

```bash
# Создать кошелёк (тело не требуется)
curl -s -X POST "https://main-node.gnd-net.com/api/v1/wallet" \
  -H "Content-Type: application/json"

# Ответ: { "success": true, "data": { "address": "GND...", "publicKey": "0x...", "privateKey": "0x..." } }

# Баланс по адресу (GND)
curl -s "https://main-node.gnd-net.com/api/v1/wallet/GND9jbK6Vca5VcZxATt3zb9yz5KQeMwjHFrz/balance"

# Ответ: { "success": true, "data": { "address": "GND...", "balance": "1000000" } }
```

---

## Транзакции и мемпул

```bash
# Отправить транзакцию GND
curl -s -X POST "https://main-node.gnd-net.com/api/v1/transaction" \
  -H "Content-Type: application/json" \
  -d '{
    "from": "GND9jbK6Vca5VcZxATt3zb9yz5KQeMwjHFrz",
    "to": "GND9LZxRfX64SfxTFDis68wvoNkrYa3ZVtg4",
    "value": "1000",
    "fee": "0",
    "nonce": 0,
    "type": "transfer",
    "data": "",
    "signature": ""
  }'

# Ответ: { "success": true, "data": "хеш_транзакции" }

# Получить транзакцию по хешу
curl -s "https://main-node.gnd-net.com/api/v1/transaction/abc123..."

# Мемпул (ожидающие транзакции)
curl -s "https://main-node.gnd-net.com/api/v1/mempool"

# Ответ: { "success": true, "data": { "size": 0, "pending_hashes": [] } }
```

---

## Блоки

```bash
# Последний блок
curl -s "https://main-node.gnd-net.com/api/v1/block/latest"

# Блок по номеру (0 — genesis)
curl -s "https://main-node.gnd-net.com/api/v1/block/0"
curl -s "https://main-node.gnd-net.com/api/v1/block/1"
```

---

## Контракты

```bash
# Деплой контракта
curl -s -X POST "https://main-node.gnd-net.com/api/v1/contract" \
  -H "Content-Type: application/json" \
  -d '{
    "from": "GND...",
    "bytecode": "0x60806040...",
    "name": "MyToken",
    "standard": "GND-st1",
    "owner": "GND...",
    "compiler": "solc 0.8",
    "version": "1.0",
    "params": {},
    "description": "",
    "metadata_cid": "",
    "source_code": "",
    "gas_limit": 3000000,
    "gas_price": "0",
    "nonce": 0,
    "signature": "",
    "total_supply": "0"
  }'

# Информация о контракте по адресу
curl -s "https://main-node.gnd-net.com/api/v1/contract/GND..."
```

---

## Создание токена (требуется X-API-Key)

Запрос из внешней системы с ключом API. Подробная логика: [api-token-deploy.md](api-token-deploy.md).

```bash
# Деплой токена (обязателен заголовок X-API-Key)
curl -s -X POST "https://main-node.gnd-net.com/api/v1/token/deploy" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: YOUR_API_KEY" \
  -d '{
    "name": "Test Token",
    "symbol": "TST",
    "decimals": 18,
    "total_supply": "1000000000000000000000000",
    "owner": "GND9jbK6Vca5VcZxATt3zb9yz5KQeMwjHFrz",
    "standard": "GND-st1"
  }'

# Ответ при успехе: { "success": true, "data": { "address": "...", "name", "symbol", "decimals", "total_supply", "standard" } }
# 401 — неверный или отсутствующий X-API-Key; 503 — сервис деплоя недоступен
```

---

## Токены (GND-st1)

```bash
# Перевод токена
curl -s -X POST "https://main-node.gnd-net.com/api/v1/token/transfer" \
  -H "Content-Type: application/json" \
  -d '{
    "token_address": "GND_контракт_токена",
    "from": "GND...",
    "to": "GND...",
    "amount": "1000000"
  }'

# Approve (разрешение списания)
curl -s -X POST "https://main-node.gnd-net.com/api/v1/token/approve" \
  -H "Content-Type: application/json" \
  -d '{
    "token_address": "GND_контракт_токена",
    "owner": "GND...",
    "spender": "GND...",
    "amount": "500000"
  }'

# Баланс токена у владельца
curl -s "https://main-node.gnd-net.com/api/v1/token/GND_контракт_токена/balance/GND_адрес_владельца"
```

---

## Порты (по config)

| Сервис   | Порт | Путь / назначение      |
|----------|------|-------------------------|
| REST API | 8182 | `/api/v1/*`             |
| RPC      | 8181 | `/block/latest`, `/tx/send` и др. |
| WebSocket| 8183 | `/ws`                   |

Через Nginx домен **main-node.gnd-net.com** проксируется на порты приложения (REST 8182, RPC 8181, WS 8183). Пример: `https://main-node.gnd-net.com/api/` → `http://127.0.0.1:8182/`.
