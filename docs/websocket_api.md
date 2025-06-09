# WebSocket API

WebSocket API предоставляет возможность подписки на события блокчейна в реальном времени.

## Подключение

```javascript
const ws = new WebSocket('ws://localhost:8080/ws');
```

## Формат сообщений

Все сообщения используют формат JSON-RPC 2.0:

```json
{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "method_name",
    "params": []
}
```

## Методы

### gnd_subscribe

Подписка на события блокчейна.

**Параметры:**
- `subscription` (string) - тип подписки:
  - `newBlock` - новые блоки
  - `newTransaction` - новые транзакции
  - `pendingTransaction` - транзакции в мемпуле
  - `contractEvent` - события контрактов

**Пример запроса:**
```json
{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "gnd_subscribe",
    "params": ["newBlock"]
}
```

**Пример ответа:**
```json
{
    "jsonrpc": "2.0",
    "result": "newBlock",
    "id": 1
}
```

### gnd_unsubscribe

Отмена подписки на события.

**Параметры:**
- `subscription` (string) - тип подписки для отмены

**Пример запроса:**
```json
{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "gnd_unsubscribe",
    "params": ["newBlock"]
}
```

**Пример ответа:**
```json
{
    "jsonrpc": "2.0",
    "result": true,
    "id": 2
}
```

## События

### Новый блок

```json
{
    "jsonrpc": "2.0",
    "method": "newBlock",
    "params": {
        "hash": "0x...",
        "number": 123,
        "timestamp": 1234567890,
        "miner": "0x...",
        "transactions": []
    }
}
```

### Новая транзакция

```json
{
    "jsonrpc": "2.0",
    "method": "newTransaction",
    "params": {
        "hash": "0x...",
        "from": "0x...",
        "to": "0x...",
        "value": "1000000000000000000",
        "nonce": 1,
        "gasPrice": "20000000000",
        "gasLimit": 21000
    }
}
```

### Событие контракта

```json
{
    "jsonrpc": "2.0",
    "method": "contractEvent",
    "params": {
        "address": "0x...",
        "event": "Transfer",
        "args": {
            "from": "0x...",
            "to": "0x...",
            "value": "1000000000000000000"
        },
        "transactionHash": "0x..."
    }
}
```

## Обработка ошибок

В случае ошибки сервер отправляет сообщение в формате:

```json
{
    "jsonrpc": "2.0",
    "error": {
        "code": -32000,
        "message": "Описание ошибки"
    },
    "id": null
}
```

## Коды ошибок

- `-32700` - Ошибка разбора JSON
- `-32600` - Неверный запрос
- `-32601` - Метод не найден
- `-32602` - Неверные параметры
- `-32000` - Ошибка сервера

## Ограничения

- Максимальный размер сообщения: 512KB
- Таймаут записи: 10 секунд
- Таймаут чтения: 60 секунд
- Период отправки ping: 54 секунды

## Пример использования

```javascript
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onopen = () => {
    // Подписка на новые блоки
    ws.send(JSON.stringify({
        jsonrpc: "2.0",
        id: 1,
        method: "gnd_subscribe",
        params: ["newBlock"]
    }));
};

ws.onmessage = (event) => {
    const response = JSON.parse(event.data);
    console.log('Получено сообщение:', response);
};

ws.onerror = (error) => {
    console.error('WebSocket ошибка:', error);
};

ws.onclose = () => {
    console.log('Соединение закрыто');
};
``` 