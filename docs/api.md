# Документация API ГАНИМЕД

Базовый URL: `http://127.0.0.1:8545`

## RPC API

### Доступные эндпоинты

#### Блоки
- `/block/latest` - Получить последний блок
- `/block/by-number` - Получить блок по номеру

#### Контракты
- `/contract/deploy` - Деплой контракта
- `/contract/call` - Вызов метода контракта
- `/contract/send` - Отправка транзакции в контракт

#### Аккаунты
- `/account/balance` - Получить баланс аккаунта

#### Транзакции
- `/tx/send` - Отправить транзакцию
- `/tx/status` - Получить статус транзакции

#### Токены
- `/token/universal-call` - Универсальный вызов токена

### Детальное описание эндпоинтов

#### Получение последнего блока
```http
GET /block/latest
```

**Ответ:**
```json
{
    "success": true,
    "data": {
        "hash": "0x...",
        "number": 12345,
        "timestamp": "2024-03-09T12:00:00Z",
        "miner": "0x...",
        "transactions": []
    }
}
```

#### Получение блока по номеру
```http
GET /block/by-number?number=12345
```

**Параметры:**
- `number` - номер блока

**Ответ:** аналогичен `/block/latest`

#### Деплой контракта
```http
POST /contract/deploy
```

**Тело запроса:**
```json
{
    "from": "0x...",
    "code": "0x...",
    "args": []
}
```

**Ответ:**
```json
{
    "success": true,
    "data": {
        "contractAddress": "0x...",
        "transactionHash": "0x..."
    }
}
```

#### Вызов метода контракта
```http
POST /contract/call
```

**Тело запроса:**
```json
{
    "to": "0x...",
    "data": "0x..."
}
```

**Ответ:**
```json
{
    "success": true,
    "data": "0x..."
}
```

#### Отправка транзакции в контракт
```http
POST /contract/send
```

**Тело запроса:**
```json
{
    "from": "0x...",
    "to": "0x...",
    "value": "0x...",
    "data": "0x..."
}
```

**Ответ:**
```json
{
    "success": true,
    "data": {
        "transactionHash": "0x..."
    }
}
```

#### Получение баланса аккаунта
```http
GET /account/balance?address=0x...
```

**Параметры:**
- `address` - адрес аккаунта

**Ответ:**
```json
{
    "success": true,
    "data": {
        "address": "0x...",
        "balance": "1000000000000000000"
    }
}
```

#### Отправка транзакции
```http
POST /tx/send
```

**Тело запроса:**
```json
{
    "from": "0x...",
    "to": "0x...",
    "value": "0x...",
    "gas": "0x...",
    "gasPrice": "0x..."
}
```

**Ответ:**
```json
{
    "success": true,
    "data": {
        "transactionHash": "0x..."
    }
}
```

#### Получение статуса транзакции
```http
GET /tx/status?hash=0x...
```

**Параметры:**
- `hash` - хеш транзакции

**Ответ:**
```json
{
    "success": true,
    "data": {
        "hash": "0x...",
        "status": "confirmed",
        "blockNumber": 12345
    }
}
```

#### Универсальный вызов токена
```http
POST /token/universal-call
```

**Тело запроса:**
```json
{
    "tokenAddr": "0x...",
    "method": "transfer",
    "args": ["0x...", "0x...", "1000000000000000000"]
}
```

**Поддерживаемые методы:**
- `transfer` - перевод токенов
- `approve` - разрешение на расход токенов
- `balanceOf` - получение баланса
- `allowance` - проверка разрешения на расход
- `transferFrom` - перевод от имени другого адреса
- `increaseAllowance` - увеличение разрешения
- `decreaseAllowance` - уменьшение разрешения
- `mint` - создание новых токенов (только для владельца)
- `burn` - уничтожение токенов
- `pause` - приостановка операций (только для владельца)
- `unpause` - возобновление операций (только для владельца)
- `transferOwnership` - передача прав владельца
- `renounceOwnership` - отказ от прав владельца

**Ответ:**
```json
{
    "success": true,
    "data": {
        "result": "0x..."
    }
}
```

## Коды ошибок

- `400` - Неверный запрос
- `401` - Не авторизован
- `403` - Доступ запрещен
- `404` - Не найдено
- `500` - Внутренняя ошибка сервера

## Примеры использования

### Получение последнего блока (curl)
```bash
curl http://127.0.0.1:8545/block/latest
```

### Деплой контракта (curl)
```bash
curl -X POST http://127.0.0.1:8545/contract/deploy \
  -H "Content-Type: application/json" \
  -d '{
    "from": "0x...",
    "code": "0x...",
    "args": []
  }'
```

### Вызов метода токена (curl)
```bash
curl -X POST http://127.0.0.1:8545/token/universal-call \
  -H "Content-Type: application/json" \
  -d '{
    "tokenAddr": "0x...",
    "method": "transfer",
    "args": ["0x...", "0x...", "1000000000000000000"]
  }'
```

### JavaScript примеры

#### Получение баланса
```javascript
async function getBalance(address) {
    const response = await fetch('http://127.0.0.1:8545/account/balance', {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json',
        },
        params: {
            address: address
        }
    });
    return await response.json();
}
```

#### Отправка транзакции
```javascript
async function sendTransaction(tx) {
    const response = await fetch('http://127.0.0.1:8545/tx/send', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(tx)
    });
    return await response.json();
}
```

#### Вызов метода контракта
```javascript
async function callContract(to, data) {
    const response = await fetch('http://127.0.0.1:8545/contract/call', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            to: to,
            data: data
        })
    });
    return await response.json();
}
``` 