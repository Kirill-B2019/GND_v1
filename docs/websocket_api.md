# WebSocket API ГАНИМЕД

## Обзор

WebSocket API ГАНИМЕД предоставляет возможность получать данные в реальном времени через WebSocket соединение.

## Базовый URL

```
ws://31.128.41.155:8183/ws
```

## Аутентификация

При подключении необходимо отправить сообщение с API ключом:

```json
{
    "type": "auth",
    "apiKey": "your-api-key"
}
```

## Подписки

### Новые блоки
```json
{
    "type": "subscribe",
    "channel": "blocks"
}
```

Ответ:
```json
{
    "type": "block",
    "block": {
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
}
```

### Новые транзакции
```json
{
    "type": "subscribe",
    "channel": "transactions"
}
```

Ответ:
```json
{
    "type": "transaction",
    "transaction": {
        "hash": "0x...",
        "from": "GND...",
        "to": "GND...",
        "value": "1000000000000000000",
        "data": "0x...",
        "nonce": 1,
        "gasPrice": "20000000000",
        "gasLimit": 21000,
        "status": "success"
    }
}
```

### Ожидающие транзакции
```json
{
    "type": "subscribe",
    "channel": "pending"
}
```

Ответ:
```json
{
    "type": "pending",
    "transaction": {
        "hash": "0x...",
        "from": "GND...",
        "to": "GND...",
        "value": "1000000000000000000",
        "data": "0x...",
        "nonce": 1,
        "gasPrice": "20000000000",
        "gasLimit": 21000
    }
}
```

### События контракта
```json
{
    "type": "subscribe",
    "channel": "events",
    "contract": "GND...",
    "event": "Transfer"
}
```

Ответ:
```json
{
    "type": "event",
    "event": {
        "contract": "GND...",
        "name": "Transfer",
        "data": {
            "from": "GND...",
            "to": "GND...",
            "value": "1000000000000000000"
        },
        "blockNumber": 123,
        "transactionHash": "0x..."
    }
}
```

### Переводы токенов
```json
{
    "type": "subscribe",
    "channel": "transfers",
    "token": "GND..."
}
```

Ответ:
```json
{
    "type": "transfer",
    "transfer": {
        "token": "GND...",
        "from": "GND...",
        "to": "GND...",
        "value": "1000000000000000000",
        "blockNumber": 123,
        "transactionHash": "0x..."
    }
}
```

### Обновления валидаторов
```json
{
    "type": "subscribe",
    "channel": "validators"
}
```

Ответ:
```json
{
    "type": "validator",
    "validator": {
        "address": "GND...",
        "stake": "1000000000000000000",
        "status": "active",
        "blockNumber": 123
    }
}
```

## Отписка

```json
{
    "type": "unsubscribe",
    "channel": "blocks"
}
```

## Обработка ошибок

### Ошибки подключения
```json
{
    "type": "error",
    "code": 1000,
    "message": "Connection refused"
}
```

### Ошибки аутентификации
```json
{
    "type": "error",
    "code": 1001,
    "message": "Invalid API key"
}
```

### Ошибки подписки
```json
{
    "type": "error",
    "code": 1002,
    "message": "Invalid channel"
}
```

### Ошибки сообщений
```json
{
    "type": "error",
    "code": 1003,
    "message": "Invalid message format"
}
```

## Коды ошибок

- 1000: Ошибка подключения
- 1001: Неверный API ключ
- 1002: Неверный канал
- 1003: Неверный формат сообщения
- 1004: Превышен лимит подписок
- 1005: Превышен лимит сообщений
- 1006: Неверный контракт
- 1007: Неверное событие
- 1008: Неверный токен
- 1009: Неверный валидатор
- 1010: Внутренняя ошибка

## Лимиты

- Максимум 10 подключений
- Максимум 100 сообщений в минуту
- Максимум 1000 сообщений в час
- Максимум 10000 сообщений в день

## Примеры

### JavaScript
```javascript
const ws = new WebSocket('ws://31.128.41.155:8183/ws');

ws.onopen = () => {
    // Аутентификация
    ws.send(JSON.stringify({
        type: 'auth',
        apiKey: 'your-api-key'
    }));

    // Подписка на блоки
    ws.send(JSON.stringify({
        type: 'subscribe',
        channel: 'blocks'
    }));

    // Подписка на транзакции
    ws.send(JSON.stringify({
        type: 'subscribe',
        channel: 'transactions'
    }));

    // Подписка на события контракта
    ws.send(JSON.stringify({
        type: 'subscribe',
        channel: 'events',
        contract: 'GND...',
        event: 'Transfer'
    }));
};

ws.onmessage = (event) => {
    const data = JSON.parse(event.data);
    
    switch (data.type) {
        case 'block':
            console.log('New block:', data.block);
            break;
        case 'transaction':
            console.log('New transaction:', data.transaction);
            break;
        case 'event':
            console.log('Contract event:', data.event);
            break;
        case 'error':
            console.error('Error:', data.message);
            break;
    }
};

ws.onerror = (error) => {
    console.error('WebSocket error:', error);
};

ws.onclose = () => {
    console.log('WebSocket connection closed');
};
```

### Python
```python
import websocket
import json
import threading
import time

def on_message(ws, message):
    data = json.loads(message)
    
    if data['type'] == 'block':
        print('New block:', data['block'])
    elif data['type'] == 'transaction':
        print('New transaction:', data['transaction'])
    elif data['type'] == 'event':
        print('Contract event:', data['event'])
    elif data['type'] == 'error':
        print('Error:', data['message'])

def on_error(ws, error):
    print('WebSocket error:', error)

def on_close(ws):
    print('WebSocket connection closed')

def on_open(ws):
    # Аутентификация
    ws.send(json.dumps({
        'type': 'auth',
        'apiKey': 'your-api-key'
    }))

    # Подписка на блоки
    ws.send(json.dumps({
        'type': 'subscribe',
        'channel': 'blocks'
    }))

    # Подписка на транзакции
    ws.send(json.dumps({
        'type': 'subscribe',
        'channel': 'transactions'
    }))

    # Подписка на события контракта
    ws.send(json.dumps({
        'type': 'subscribe',
        'channel': 'events',
        'contract': 'GND...',
        'event': 'Transfer'
    }))

def run_websocket():
    websocket.enableTrace(True)
    ws = websocket.WebSocketApp(
        'ws://31.128.41.155:8183/ws',
        on_message=on_message,
        on_error=on_error,
        on_close=on_close,
        on_open=on_open
    )
    ws.run_forever()

if __name__ == '__main__':
    thread = threading.Thread(target=run_websocket)
    thread.start()
```

### Go
```go
package main

import (
    "fmt"
    "log"
    "net/url"
    "os"
    "os/signal"
    "time"

    "github.com/gorilla/websocket"
)

type Message struct {
    Type    string      `json:"type"`
    Channel string      `json:"channel,omitempty"`
    APIKey  string      `json:"apiKey,omitempty"`
    Contract string     `json:"contract,omitempty"`
    Event   string      `json:"event,omitempty"`
    Data    interface{} `json:"data,omitempty"`
}

func main() {
    interrupt := make(chan os.Signal, 1)
    signal.Notify(interrupt, os.Interrupt)

    u := url.URL{Scheme: "ws", Host: "31.128.41.155:8183", Path: "/ws"}
    log.Printf("connecting to %s", u.String())

    c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
    if err != nil {
        log.Fatal("dial:", err)
    }
    defer c.Close()

    done := make(chan struct{})

    go func() {
        defer close(done)
        for {
            _, message, err := c.ReadMessage()
            if err != nil {
                log.Println("read:", err)
                return
            }
            log.Printf("recv: %s", message)
        }
    }()

    // Аутентификация
    err = c.WriteJSON(Message{
        Type:   "auth",
        APIKey: "your-api-key",
    })
    if err != nil {
        log.Println("write:", err)
        return
    }

    // Подписка на блоки
    err = c.WriteJSON(Message{
        Type:    "subscribe",
        Channel: "blocks",
    })
    if err != nil {
        log.Println("write:", err)
        return
    }

    // Подписка на транзакции
    err = c.WriteJSON(Message{
        Type:    "subscribe",
        Channel: "transactions",
    })
    if err != nil {
        log.Println("write:", err)
        return
    }

    // Подписка на события контракта
    err = c.WriteJSON(Message{
        Type:     "subscribe",
        Channel:  "events",
        Contract: "GND...",
        Event:    "Transfer",
    })
    if err != nil {
        log.Println("write:", err)
        return
    }

    for {
        select {
        case <-done:
            return
        case <-interrupt:
            log.Println("interrupt")

            // Отписка от всех каналов
            err := c.WriteJSON(Message{
                Type:    "unsubscribe",
                Channel: "blocks",
            })
            if err != nil {
                log.Println("write:", err)
                return
            }

            err = c.WriteJSON(Message{
                Type:    "unsubscribe",
                Channel: "transactions",
            })
            if err != nil {
                log.Println("write:", err)
                return
            }

            err = c.WriteJSON(Message{
                Type:    "unsubscribe",
                Channel: "events",
            })
            if err != nil {
                log.Println("write:", err)
                return
            }

            // Закрытие соединения
            err = c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
            if err != nil {
                log.Println("write close:", err)
                return
            }
            select {
            case <-done:
            case <-time.After(time.Second):
            }
            return
        }
    }
}
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
- Количество подключений
- Количество сообщений
- Время ответа
- Ошибки

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