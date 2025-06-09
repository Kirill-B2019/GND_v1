# API ГАНИМЕД

## Обзор

Блокчейн ГАНИМЕД предоставляет три типа API:
- REST API
- WebSocket API
- JSON-RPC API

## Базовые URL

- REST API: `http://45.12.72.15:8182/api/`
- WebSocket API: `ws://45.12.72.15:8181/ws`
- JSON-RPC API: `http://45.12.72.15:8545`

## Аутентификация

Все запросы к API должны включать заголовок `X-API-Key` с действительным API ключом. При неверном ключе возвращается статус 401 Unauthorized.

## REST API

### Кошельки

#### Создание кошелька
```http
POST /wallet/create
Content-Type: application/json

Response:
{
    "address": "GND...",
    "privateKey": "0x..."
}
```

#### Получение баланса
```http
GET /wallet/balance/{address}

Response:
{
    "address": "GND...",
    "balance": "1000000000000000000"
}
```

#### Получение транзакций
```http
GET /wallet/transactions/{address}
Query Parameters:
- page: номер страницы (default: 1)
- limit: количество транзакций (default: 10)

Response:
{
    "transactions": [
        {
            "hash": "0x...",
            "from": "GND...",
            "to": "GND...",
            "value": "1000000000000000000",
            "status": "success",
            "timestamp": 1234567890
        }
    ],
    "total": 100,
    "page": 1,
    "limit": 10
}
```

### Токены

#### Получение баланса токена
```http
GET /token/balance/{address}
Query Parameters:
- tokenAddress: адрес токена

Response:
{
    "address": "GND...",
    "tokenAddress": "GND...",
    "balance": "1000000000000000000"
}
```

#### Перевод токенов
```http
POST /token/transfer
Content-Type: application/json

{
    "from": "GND...",
    "to": "GND...",
    "tokenAddress": "GND...",
    "amount": "1000000000000000000",
    "privateKey": "0x..."
}

Response:
{
    "hash": "0x...",
    "status": "success"
}
```

#### Одобрение токенов
```http
POST /token/approve
Content-Type: application/json

{
    "owner": "GND...",
    "spender": "GND...",
    "tokenAddress": "GND...",
    "amount": "1000000000000000000",
    "privateKey": "0x..."
}

Response:
{
    "hash": "0x...",
    "status": "success"
}
```

### Транзакции

#### Отправка транзакции
```http
POST /tx/send
Content-Type: application/json

{
    "from": "GND...",
    "to": "GND...",
    "value": "1000000000000000000",
    "data": "0x...",
    "privateKey": "0x..."
}

Response:
{
    "hash": "0x...",
    "status": "success"
}
```

#### Получение информации о транзакции
```http
GET /tx/{hash}

Response:
{
    "hash": "0x...",
    "from": "GND...",
    "to": "GND...",
    "value": "1000000000000000000",
    "data": "0x...",
    "status": "success",
    "blockNumber": 123,
    "timestamp": 1234567890
}
```

### Блоки

#### Получение последнего блока
```http
GET /block/latest

Response:
{
    "number": 123,
    "hash": "0x...",
    "parentHash": "0x...",
    "timestamp": 1234567890,
    "transactions": [
        {
            "hash": "0x...",
            "from": "GND...",
            "to": "GND...",
            "value": "1000000000000000000"
        }
    ],
    "validator": "GND..."
}
```

#### Получение блока по номеру
```http
GET /block/{number}

Response:
{
    "number": 123,
    "hash": "0x...",
    "parentHash": "0x...",
    "timestamp": 1234567890,
    "transactions": [
        {
            "hash": "0x...",
            "from": "GND...",
            "to": "GND...",
            "value": "1000000000000000000"
        }
    ],
    "validator": "GND..."
}
```

### Контракты

#### Деплой контракта
```http
POST /contract/deploy
Content-Type: application/json

{
    "from": "GND...",
    "code": "0x...",
    "args": ["arg1", "arg2"],
    "privateKey": "0x..."
}

Response:
{
    "address": "GND...",
    "hash": "0x...",
    "status": "success"
}
```

#### Вызов метода контракта
```http
POST /contract/call
Content-Type: application/json

{
    "from": "GND...",
    "to": "GND...",
    "method": "transfer",
    "args": ["GND...", "1000000000000000000"],
    "privateKey": "0x..."
}

Response:
{
    "hash": "0x...",
    "status": "success",
    "result": "0x..."
}
```

#### Получение информации о контракте
```http
GET /contract/{address}

Response:
{
    "address": "GND...",
    "creator": "GND...",
    "creationTx": "0x...",
    "creationBlock": 123,
    "code": "0x...",
    "abi": [...]
}
```

## WebSocket API

### Подключение
```javascript
const ws = new WebSocket('ws://45.12.72.15:8181/ws');

ws.onopen = () => {
    ws.send(JSON.stringify({
        type: 'auth',
        apiKey: 'your-api-key'
    }));
};
```

### Подписки

#### Новые блоки
```javascript
ws.send(JSON.stringify({
    type: 'subscribe',
    channel: 'blocks'
}));

ws.onmessage = (event) => {
    const data = JSON.parse(event.data);
    if (data.type === 'block') {
        console.log('New block:', data.block);
    }
};
```

#### Новые транзакции
```javascript
ws.send(JSON.stringify({
    type: 'subscribe',
    channel: 'transactions'
}));

ws.onmessage = (event) => {
    const data = JSON.parse(event.data);
    if (data.type === 'transaction') {
        console.log('New transaction:', data.transaction);
    }
};
```

#### Ожидающие транзакции
```javascript
ws.send(JSON.stringify({
    type: 'subscribe',
    channel: 'pending'
}));

ws.onmessage = (event) => {
    const data = JSON.parse(event.data);
    if (data.type === 'pending') {
        console.log('Pending transaction:', data.transaction);
    }
};
```

#### События контракта
```javascript
ws.send(JSON.stringify({
    type: 'subscribe',
    channel: 'events',
    contract: 'GND...',
    event: 'Transfer'
}));

ws.onmessage = (event) => {
    const data = JSON.parse(event.data);
    if (data.type === 'event') {
        console.log('Contract event:', data.event);
    }
};
```

### Отписка
```javascript
ws.send(JSON.stringify({
    type: 'unsubscribe',
    channel: 'blocks'
}));
```

## JSON-RPC API

### Методы

#### eth_blockNumber
```json
{
    "jsonrpc": "2.0",
    "method": "eth_blockNumber",
    "params": [],
    "id": 1
}
```

#### eth_getBalance
```json
{
    "jsonrpc": "2.0",
    "method": "eth_getBalance",
    "params": ["GND...", "latest"],
    "id": 1
}
```

#### eth_sendTransaction
```json
{
    "jsonrpc": "2.0",
    "method": "eth_sendTransaction",
    "params": [{
        "from": "GND...",
        "to": "GND...",
        "value": "0xde0b6b3a7640000",
        "data": "0x..."
    }],
    "id": 1
}
```

#### eth_getTransactionReceipt
```json
{
    "jsonrpc": "2.0",
    "method": "eth_getTransactionReceipt",
    "params": ["0x..."],
    "id": 1
}
```

## Коды ошибок

### HTTP статусы
- 200 OK
- 400 Bad Request
- 401 Unauthorized
- 403 Forbidden
- 404 Not Found
- 429 Too Many Requests
- 500 Internal Server Error

### Коды ошибок
- 1000: Неверный API ключ
- 1001: Неверный формат запроса
- 1002: Неверные параметры
- 1003: Недостаточно средств
- 1004: Неверная подпись
- 1005: Транзакция отклонена
- 1006: Контракт не найден
- 1007: Метод не найден
- 1008: Событие не найдено
- 1009: Превышен лимит запросов

## Лимиты

### REST API
- 100 запросов в минуту
- 1000 запросов в час
- 10000 запросов в день

### WebSocket API
- 10 подключений
- 100 сообщений в минуту
- 1000 сообщений в час

### JSON-RPC API
- 100 запросов в минуту
- 1000 запросов в час
- 10000 запросов в день

## Примеры

### JavaScript
```javascript
const api = new GND.API({
    rest: 'http://45.12.72.15:8182/api/',
    ws: 'ws://45.12.72.15:8181/ws',
    rpc: 'http://45.12.72.15:8545',
    apiKey: 'your-api-key'
});

// Создание кошелька
const wallet = await api.createWallet();

// Получение баланса
const balance = await api.getBalance(wallet.address);

// Отправка транзакции
const tx = await api.sendTransaction({
    from: wallet.address,
    to: 'GND...',
    value: '1000000000000000000',
    privateKey: wallet.privateKey
});

// Подписка на события
api.subscribe('blocks', (block) => {
    console.log('New block:', block);
});
```

### Python
```python
from gnd import API

api = API(
    rest='http://45.12.72.15:8182/api/',
    ws='ws://45.12.72.15:8181/ws',
    rpc='http://45.12.72.15:8545',
    api_key='your-api-key'
)

# Создание кошелька
wallet = api.create_wallet()

# Получение баланса
balance = api.get_balance(wallet.address)

# Отправка транзакции
tx = api.send_transaction(
    from_=wallet.address,
    to='GND...',
    value='1000000000000000000',
    private_key=wallet.private_key
)

# Подписка на события
api.subscribe('blocks', lambda block: print('New block:', block))
```

### Go
```go
import "github.com/gnd/api"

config := api.Config{
    REST:   "http://45.12.72.15:8182/api/",
    WS:     "ws://45.12.72.15:8181/ws",
    RPC:    "http://45.12.72.15:8545",
    APIKey: "your-api-key",
}

a := api.New(config)

// Создание кошелька
wallet, err := a.CreateWallet()

// Получение баланса
balance, err := a.GetBalance(wallet.Address)

// Отправка транзакции
tx, err := a.SendTransaction(&api.Transaction{
    From:      wallet.Address,
    To:        "GND...",
    Value:     big.NewInt(1000000000000000000),
    PrivateKey: wallet.PrivateKey,
})

// Подписка на события
a.Subscribe("blocks", func(block *api.Block) {
    fmt.Println("New block:", block)
})
```

## Интеграция

### SDK
- JavaScript/TypeScript
- Python
- Go
- Java

### Инструменты
- CLI
- GUI
- Мониторинг
- Аналитика

## Обновления

### Версионирование
- Семантическое версионирование
- Обратная совместимость
- Миграции
- Обновления

### Миграции
- Планирование
- Тестирование
- Резервное копирование
- Откат

## Мониторинг

### Метрики
- Количество запросов
- Время ответа
- Ошибки
- Использование ресурсов

### Алерты
- Превышение лимитов
- Ошибки
- Замедление
- Аномалии

## Безопасность

### Аудит
- Код
- Конфигурация
- Доступ
- Данные

### Мониторинг
- Активность
- Аномалии
- Угрозы
- Инциденты

### Реагирование
- Обнаружение
- Анализ
- Устранение
- Профилактика 