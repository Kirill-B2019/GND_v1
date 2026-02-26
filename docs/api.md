# API ГАНИМЕД

## Обзор

Подключение к ноде блокчейна ГАНИМЕД выполняется по домену **main-node.gnd-net.com**. Описание и документация API публикуются на **api.gnd-net.com**. Предоставляются три типа API:
- REST API
- WebSocket API
- RPC API

## Базовые URL (подключение к ноде main-node.gnd-net.com)

- REST API: `https://main-node.gnd-net.com/api/v1` (при прокси без порта; с портом: `https://main-node.gnd-net.com:8182/api/v1`)
- RPC API: `https://main-node.gnd-net.com:8181`
- WebSocket API: `wss://main-node.gnd-net.com:8183/ws`

## Аутентификация

Для операций от внешних систем используется заголовок **X-API-Key**. Обязательная проверка ключа реализована для эндпоинта **POST /api/v1/token/deploy** (создание токена); при отсутствии или неверном ключе возвращается 401 Unauthorized. Ключ проверяется по константе (тестовый ключ) или по таблице `public.api_keys` (поле `key`, учёт `expires_at`). Остальные эндпоинты REST в текущей реализации могут вызываться без ключа.

## REST API

### Кошельки

#### Создание кошелька
```http
POST /wallet/create
Content-Type: application/json

Response:
{
    "address": "GND...",
    "publicKey": "0x...",
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

### Токены

#### Создание токена (требуется X-API-Key)

Внешняя система создаёт и регистрирует токен запросом с заголовком **X-API-Key**. Подробно: **[api-token-deploy.md](api-token-deploy.md)**.

```http
POST /api/v1/token/deploy
Content-Type: application/json
X-API-Key: <ваш_ключ>

{ "name": "Test Token", "symbol": "TST", "decimals": 18, "total_supply": "1000000000000000000000000", "owner": "GND...", "standard": "GND-st1" }
```
Ответ 200: `{ "success": true, "data": { "address", "name", "symbol", "decimals", "total_supply", "standard" } }`. 401 — неверный ключ; 503 — сервис деплоя недоступен.

#### Универсальный вызов токена
```http
POST /token/call
Content-Type: application/json

{
    "tokenAddr": "GND...",
    "method": "transfer|approve|balanceOf",
    "args": ["from", "to", "amount"] // для transfer
    "args": ["owner", "spender", "amount"] // для approve
    "args": ["address"] // для balanceOf
}

Response:
{
    "success": true
}
```

#### Получение баланса токена
```http
GET /token/balance/{address}
Content-Type: application/json

{
    "tokenAddr": "GND..."
}

Response:
{
    "address": "GND...",
    "balance": "1000000000000000000"
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
    "nonce": 0,
    "gasLimit": 1000000,
    "gasPrice": "1",
    "privateKey": "0x..."
}

Response:
{
    "hash": "0x...",
    "status": "pending",
    "timestamp": "2024-03-21T12:00:00Z"
}
```

#### Получение статуса транзакции
```http
GET /tx/{hash}

Response:
{
    "hash": "0x...",
    "status": "success|pending|failed"
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

## RPC API

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
    "transactions": [...],
    "validator": "GND..."
}
```

#### Получение блока по номеру
```http
GET /block/by-number?number=123

Response:
{
    "number": 123,
    "hash": "0x...",
    "parentHash": "0x...",
    "timestamp": 1234567890,
    "transactions": [...],
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
    "bytecode": "0x...",
    "name": "TestContract",
    "standard": "ERC20",
    "owner": "GND...",
    "compiler": "solc",
    "version": "1.0.0",
    "params": {},
    "description": "Test contract",
    "metadata_cid": "",
    "source_code": "",
    "gas_limit": 1000000,
    "gas_price": 1,
    "nonce": 0,
    "signature": "0x...",
    "total_supply": "1000000000000000000"
}

Response:
{
    "address": "GND..."
}
```

#### Вызов метода контракта
```http
POST /contract/call
Content-Type: application/json

{
    "from": "GND...",
    "to": "GND...",
    "data": "0x...",
    "gas_limit": 1000000,
    "gas_price": 1,
    "value": 0,
    "signature": "0x..."
}

Response:
{
    "result": "0x..."
}
```

#### Отправка транзакции в контракт
```http
POST /contract/send
Content-Type: application/json

{
    "from": "GND...",
    "to": "GND...",
    "data": "0x...",
    "gas_limit": 1000000,
    "gas_price": 1,
    "value": 0,
    "nonce": 0,
    "signature": "0x..."
}

Response:
{
    "result": "0x..."
}
```

### Аккаунты

#### Получение баланса
```http
GET /account/balance?address=GND...

Response:
{
    "address": "GND...",
    "balance": "1000000000000000000"
}
```

### Транзакции

#### Отправка транзакции
```http
POST /tx/send
Content-Type: application/json

{
    "raw_tx": "0x..."
}

Response:
{
    "txHash": "0x..."
}
```

#### Получение статуса транзакции
```http
GET /tx/status?hash=0x...

Response:
{
    "status": "success|pending|failed"
}
```

### Токены

#### Универсальный вызов токена
```http
POST /token/universal-call
Content-Type: application/json

{
    "token_address": "GND...",
    "method": "transfer|approve|balanceOf",
    "args": ["from", "to", "amount"] // для transfer
    "args": ["owner", "spender", "amount"] // для approve
    "args": ["address"] // для balanceOf
}

Response:
{
    "result": "0x..."
}
```

## WebSocket API

### Подключение
```javascript
const ws = new WebSocket('ws://localhost:8183/ws');

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
    jsonrpc: "2.0",
    method: "gnd_subscribe",
    params: ["blocks"],
    id: 1
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
    jsonrpc: "2.0",
    method: "gnd_subscribe",
    params: ["transactions"],
    id: 1
}));

ws.onmessage = (event) => {
    const data = JSON.parse(event.data);
    if (data.type === 'transaction') {
        console.log('New transaction:', data.transaction);
    }
};
```

### Отписка
```javascript
ws.send(JSON.stringify({
    jsonrpc: "2.0",
    method: "gnd_unsubscribe",
    params: ["blocks"],
    id: 1
}));
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

### RPC API
- 100 запросов в минуту
- 1000 запросов в час
- 10000 запросов в день

## Примеры

### JavaScript
```javascript
const api = new GND.API({
    rest: 'https://main-node.gnd-net.com/api/v1',
    ws: 'wss://main-node.gnd-net.com:8183/ws',
    rpc: 'https://main-node.gnd-net.com:8181',
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
    rest='https://main-node.gnd-net.com/api/v1',
    ws='wss://main-node.gnd-net.com:8183/ws',
    rpc='https://main-node.gnd-net.com:8181',
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
<!-- Пример на Go (блок text, чтобы IDE не анализировала гипотетический SDK) -->
```text
package main
import "github.com/gnd/api"

config := api.Config{
    REST:   "https://main-node.gnd-net.com/api/v1",
    WS:     "wss://main-node.gnd-net.com:8183/ws",
    RPC:    "https://main-node.gnd-net.com:8181",
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

### JavaScript/TypeScript

```javascript
class GNDAPI {
    constructor(config) {
        this.restUrl = config.rest;
        this.rpcUrl = config.rpc;
        this.wsUrl = config.ws;
        this.apiKey = config.apiKey;
    }

    async createWallet() {
        const response = await fetch(`${this.restUrl}/wallet/create`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-API-Key': this.apiKey
            }
        });
        return response.json();
    }

    async getBalance(address) {
        const response = await fetch(`${this.restUrl}/wallet/balance/${address}`, {
            headers: {
                'X-API-Key': this.apiKey
            }
        });
        return response.json();
    }

    async sendTransaction(tx) {
        const response = await fetch(`${this.rpcUrl}/tx/send`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-API-Key': this.apiKey
            },
            body: JSON.stringify(tx)
        });
        return response.json();
    }

    async deployContract(params) {
        const response = await fetch(`${this.rpcUrl}/contract/deploy`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-API-Key': this.apiKey
            },
            body: JSON.stringify(params)
        });
        return response.json();
    }

    async callContract(params) {
        const response = await fetch(`${this.rpcUrl}/contract/call`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-API-Key': this.apiKey
            },
            body: JSON.stringify(params)
        });
        return response.json();
    }

    async getLatestBlock() {
        const response = await fetch(`${this.rpcUrl}/block/latest`, {
            headers: {
                'X-API-Key': this.apiKey
            }
        });
        return response.json();
    }

    subscribe(channel, callback) {
        const ws = new WebSocket(this.wsUrl);
        
        ws.onopen = () => {
            ws.send(JSON.stringify({
                type: 'auth',
                apiKey: this.apiKey
            }));
            
            ws.send(JSON.stringify({
                type: 'subscribe',
                channel: channel
            }));
        };

        ws.onmessage = (event) => {
            const data = JSON.parse(event.data);
            callback(data);
        };

        return {
            unsubscribe: () => {
                ws.send(JSON.stringify({
                    type: 'unsubscribe',
                    channel: channel
                }));
                ws.close();
            }
        };
    }
}

// Пример использования
const api = new GNDAPI({
    rest: 'https://main-node.gnd-net.com/api/v1',
    rpc: 'https://main-node.gnd-net.com:8181',
    ws: 'wss://main-node.gnd-net.com:8183/ws',
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
    data: '0x...',
    gas_limit: 1000000,
    gas_price: 1,
    nonce: 0,
    signature: '0x...'
});

// Деплой контракта
const contract = await api.deployContract({
    from: wallet.address,
    bytecode: '0x...',
    name: 'TestContract',
    standard: 'ERC20',
    owner: wallet.address,
    compiler: 'solc',
    version: '1.0.0',
    params: {},
    description: 'Test contract',
    metadata_cid: '',
    source_code: '',
    gas_limit: 1000000,
    gas_price: 1,
    nonce: 0,
    signature: '0x...',
    total_supply: '1000000000000000000'
});

// Вызов контракта
const result = await api.callContract({
    from: wallet.address,
    to: contract.address,
    data: '0x...',
    gas_limit: 1000000,
    gas_price: 1,
    value: 0,
    signature: '0x...'
});

// Подписка на события
const subscription = api.subscribe('blocks', (block) => {
    console.log('New block:', block);
});

// Отписка
subscription.unsubscribe();
```

### Python

```python
import aiohttp
import asyncio
import json
import websockets

class GNDAPI:
    def __init__(self, config):
        self.rest_url = config['rest']
        self.rpc_url = config['rpc']
        self.ws_url = config['ws']
        self.api_key = config['api_key']

    async def create_wallet(self):
        async with aiohttp.ClientSession() as session:
            async with session.post(
                f"{self.rest_url}/wallet/create",
                headers={'X-API-Key': self.api_key}
            ) as response:
                return await response.json()

    async def get_balance(self, address):
        async with aiohttp.ClientSession() as session:
            async with session.get(
                f"{self.rest_url}/wallet/balance/{address}",
                headers={'X-API-Key': self.api_key}
            ) as response:
                return await response.json()

    async def send_transaction(self, tx):
        async with aiohttp.ClientSession() as session:
            async with session.post(
                f"{self.rpc_url}/tx/send",
                headers={
                    'Content-Type': 'application/json',
                    'X-API-Key': self.api_key
                },
                json=tx
            ) as response:
                return await response.json()

    async def deploy_contract(self, params):
        async with aiohttp.ClientSession() as session:
            async with session.post(
                f"{self.rpc_url}/contract/deploy",
                headers={
                    'Content-Type': 'application/json',
                    'X-API-Key': self.api_key
                },
                json=params
            ) as response:
                return await response.json()

    async def call_contract(self, params):
        async with aiohttp.ClientSession() as session:
            async with session.post(
                f"{self.rpc_url}/contract/call",
                headers={
                    'Content-Type': 'application/json',
                    'X-API-Key': self.api_key
                },
                json=params
            ) as response:
                return await response.json()

    async def get_latest_block(self):
        async with aiohttp.ClientSession() as session:
            async with session.get(
                f"{self.rpc_url}/block/latest",
                headers={'X-API-Key': self.api_key}
            ) as response:
                return await response.json()

    async def subscribe(self, channel, callback):
        async with websockets.connect(self.ws_url) as websocket:
            # Аутентификация
            await websocket.send(json.dumps({
                'type': 'auth',
                'apiKey': self.api_key
            }))

            # Подписка
            await websocket.send(json.dumps({
                'type': 'subscribe',
                'channel': channel
            }))

            while True:
                data = await websocket.recv()
                callback(json.loads(data))

# Пример использования
async def main():
    api = GNDAPI({
        'rest': 'https://main-node.gnd-net.com/api/v1',
        'rpc': 'https://main-node.gnd-net.com:8181',
        'ws': 'wss://main-node.gnd-net.com:8183/ws',
        'api_key': 'your-api-key'
    })

    # Создание кошелька
    wallet = await api.create_wallet()

    # Получение баланса
    balance = await api.get_balance(wallet['address'])

    # Отправка транзакции
    tx = await api.send_transaction({
        'from': wallet['address'],
        'to': 'GND...',
        'value': '1000000000000000000',
        'data': '0x...',
        'gas_limit': 1000000,
        'gas_price': 1,
        'nonce': 0,
        'signature': '0x...'
    })

    # Деплой контракта
    contract = await api.deploy_contract({
        'from': wallet['address'],
        'bytecode': '0x...',
        'name': 'TestContract',
        'standard': 'ERC20',
        'owner': wallet['address'],
        'compiler': 'solc',
        'version': '1.0.0',
        'params': {},
        'description': 'Test contract',
        'metadata_cid': '',
        'source_code': '',
        'gas_limit': 1000000,
        'gas_price': 1,
        'nonce': 0,
        'signature': '0x...',
        'total_supply': '1000000000000000000'
    })

    # Вызов контракта
    result = await api.call_contract({
        'from': wallet['address'],
        'to': contract['address'],
        'data': '0x...',
        'gas_limit': 1000000,
        'gas_price': 1,
        'value': 0,
        'signature': '0x...'
    })

    # Подписка на события
    async def on_block(block):
        print('New block:', block)

    await api.subscribe('blocks', on_block)

if __name__ == '__main__':
    asyncio.run(main())
```

### Go

<!-- Пример клиента API на Go (блок text, чтобы IDE не проверяла зависимости) -->
```text
package main
import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
    "github.com/gorilla/websocket"
)

type GNDAPI struct {
    restURL  string
    rpcURL   string
    wsURL    string
    apiKey   string
    client   *http.Client
}

func NewGNDAPI(config map[string]string) *GNDAPI {
    return &GNDAPI{
        restURL: config["rest"],
        rpcURL:  config["rpc"],
        wsURL:   config["ws"],
        apiKey:  config["apiKey"],
        client:  &http.Client{Timeout: 10 * time.Second},
    }
}

func (api *GNDAPI) CreateWallet() (map[string]interface{}, error) {
    req, err := http.NewRequest("POST", api.restURL+"/wallet/create", nil)
    if err != nil {
        return nil, err
    }
    req.Header.Set("X-API-Key", api.apiKey)

    resp, err := api.client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    return result, nil
}

func (api *GNDAPI) GetBalance(address string) (map[string]interface{}, error) {
    req, err := http.NewRequest("GET", fmt.Sprintf("%s/wallet/balance/%s", api.restURL, address), nil)
    if err != nil {
        return nil, err
    }
    req.Header.Set("X-API-Key", api.apiKey)

    resp, err := api.client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    return result, nil
}

func (api *GNDAPI) SendTransaction(tx map[string]interface{}) (map[string]interface{}, error) {
    body, err := json.Marshal(tx)
    if err != nil {
        return nil, err
    }

    req, err := http.NewRequest("POST", api.rpcURL+"/tx/send", bytes.NewBuffer(body))
    if err != nil {
        return nil, err
    }
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-API-Key", api.apiKey)

    resp, err := api.client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    return result, nil
}

func (api *GNDAPI) DeployContract(params map[string]interface{}) (map[string]interface{}, error) {
    body, err := json.Marshal(params)
    if err != nil {
        return nil, err
    }

    req, err := http.NewRequest("POST", api.rpcURL+"/contract/deploy", bytes.NewBuffer(body))
    if err != nil {
        return nil, err
    }
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-API-Key", api.apiKey)

    resp, err := api.client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    return result, nil
}

func (api *GNDAPI) CallContract(params map[string]interface{}) (map[string]interface{}, error) {
    body, err := json.Marshal(params)
    if err != nil {
        return nil, err
    }

    req, err := http.NewRequest("POST", api.rpcURL+"/contract/call", bytes.NewBuffer(body))
    if err != nil {
        return nil, err
    }
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-API-Key", api.apiKey)

    resp, err := api.client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    return result, nil
}

func (api *GNDAPI) GetLatestBlock() (map[string]interface{}, error) {
    req, err := http.NewRequest("GET", api.rpcURL+"/block/latest", nil)
    if err != nil {
        return nil, err
    }
    req.Header.Set("X-API-Key", api.apiKey)

    resp, err := api.client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    return result, nil
}

func (api *GNDAPI) Subscribe(channel string, callback func(interface{})) error {
    c, _, err := websocket.DefaultDialer.Dial(api.wsURL, nil)
    if err != nil {
        return err
    }
    defer c.Close()

    // Аутентификация
    auth := map[string]string{
        "type":   "auth",
        "apiKey": api.apiKey,
    }
    if err := c.WriteJSON(auth); err != nil {
        return err
    }

    // Подписка
    subscribe := map[string]string{
        "type":    "subscribe",
        "channel": channel,
    }
    if err := c.WriteJSON(subscribe); err != nil {
        return err
    }

    for {
        _, message, err := c.ReadMessage()
        if err != nil {
            return err
        }

        var data interface{}
        if err := json.Unmarshal(message, &data); err != nil {
            return err
        }

        callback(data)
    }
}

func main() {
    api := NewGNDAPI(map[string]string{
        "rest":   "https://main-node.gnd-net.com/api/v1",
        "rpc":    "https://main-node.gnd-net.com:8181",
        "ws":     "wss://main-node.gnd-net.com:8183/ws",
        "apiKey": "your-api-key",
    })

    // Создание кошелька
    wallet, err := api.CreateWallet()
    if err != nil {
        panic(err)
    }

    // Получение баланса
    _, err = api.GetBalance(wallet["address"].(string))
    if err != nil {
        panic(err)
    }

    // Отправка транзакции
    _, err = api.SendTransaction(map[string]interface{}{
        "from":      wallet["address"],
        "to":        "GND...",
        "value":     "1000000000000000000",
        "data":      "0x...",
        "gas_limit": 1000000,
        "gas_price": 1,
        "nonce":     0,
        "signature": "0x...",
    })
    if err != nil {
        panic(err)
    }

    // Деплой контракта
    contract, err := api.DeployContract(map[string]interface{}{
        "from":         wallet["address"],
        "bytecode":     "0x...",
        "name":         "TestContract",
        "standard":     "ERC20",
        "owner":        wallet["address"],
        "compiler":     "solc",
        "version":      "1.0.0",
        "params":       map[string]interface{}{},
        "description":  "Test contract",
        "metadata_cid": "",
        "source_code":  "",
        "gas_limit":    1000000,
        "gas_price":    1,
        "nonce":        0,
        "signature":    "0x...",
        "total_supply": "1000000000000000000",
    })
    if err != nil {
        panic(err)
    }

    // Вызов контракта
    _, err = api.CallContract(map[string]interface{}{
        "from":      wallet["address"],
        "to":        contract["address"],
        "data":      "0x...",
        "gas_limit": 1000000,
        "gas_price": 1,
        "value":     0,
        "signature": "0x...",
    })
    if err != nil {
        panic(err)
    }

    // Подписка на события
    go func() {
        err := api.Subscribe("blocks", func(data interface{}) {
            fmt.Printf("New block: %v\n", data)
        })
        if err != nil {
            panic(err)
        }
    }()

    // Ожидание завершения
    select {}
}
```

## Инструменты

### CLI

```bash
# Создание кошелька
gnd-cli wallet create

# Получение баланса
gnd-cli wallet balance GND...

# Отправка транзакции
gnd-cli tx send --from GND... --to GND... --value 1000000000000000000

# Деплой контракта
gnd-cli contract deploy --from GND... --bytecode 0x... --name TestContract

# Вызов контракта
gnd-cli contract call --from GND... --to GND... --data 0x...

# Получение последнего блока
gnd-cli block latest
```

### GUI

Графический интерфейс предоставляет следующие возможности:
- Создание и управление кошельками
- Просмотр балансов и транзакций
- Отправка транзакций
- Деплой и взаимодействие с контрактами
- Мониторинг блоков и событий
- Управление токенами

### Мониторинг

```bash
# Запуск мониторинга
gnd-monitor start

# Просмотр метрик
gnd-monitor metrics

# Настройка алертов
gnd-monitor alerts set --metric requests --threshold 1000 --period 1m
```

### Аналитика

```bash
# Запуск аналитики
gnd-analytics start

# Экспорт данных
gnd-analytics export --format csv --period 24h

# Генерация отчета
gnd-analytics report --type daily
```

## Обновления

### Версионирование
- Семантическое версионирование (MAJOR.MINOR.PATCH)
- Обратная совместимость в пределах MAJOR версии
- Автоматические миграции для MINOR версий
- Ручные миграции для MAJOR версий

### Миграции
1. Планирование
   - Анализ изменений
   - Оценка рисков
   - Создание плана миграции

2. Тестирование
   - Тестирование на staging
   - Проверка обратной совместимости
   - Тестирование производительности

3. Резервное копирование
   - Бэкап данных
   - Бэкап конфигурации
   - Бэкап состояния

4. Откат
   - План отката
   - Триггеры отката
   - Процедура отката

## Мониторинг

### Метрики
- Количество запросов
- Время ответа
- Ошибки
- Использование ресурсов
- Размер блокчейна
- Количество транзакций
- Газ

### Алерты
- Превышение лимитов
- Ошибки
- Замедление
- Аномалии
- Недоступность
- Атаки

## Безопасность

### Аудит
- Код
- Конфигурация
- Доступ
- Данные
- Сеть
- Инфраструктура

### Мониторинг
- Активность
- Аномалии
- Угрозы
- Инциденты
- Доступ
- Изменения

### Реагирование
1. Обнаружение
   - Мониторинг
   - Алерты
   - Логи

2. Анализ
   - Сбор данных
   - Определение причины
   - Оценка ущерба

3. Устранение
   - Блокировка угрозы
   - Восстановление
   - Патч

4. Профилактика
   - Обновление
   - Усиление защиты
   - Документирование 