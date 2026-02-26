# Интеграция ГАНИМЕД

## Обзор

Интеграция ГАНИМЕД включает в себя подключение к блокчейну и использование его функциональности.

## SDK

### JavaScript/TypeScript
```javascript
// Установка
npm install @ganymed/sdk

// Инициализация
const { Ganymed } = require('@ganymed/sdk');
const gnd = new Ganymed({
  rpcUrl: 'http://31.128.41.155:8181',
  wsUrl: 'ws://31.128.41.155:8183/ws',
  apiKey: 'your-api-key'
});

// Создание кошелька
const wallet = await gnd.wallet.create();

// Получение баланса
const balance = await gnd.wallet.getBalance(wallet.address);

// Отправка транзакции
const tx = await gnd.wallet.send({
  from: wallet.address,
  to: 'GND...',
  value: '1',
  unit: 'ether'
});

// Подписка на события
gnd.ws.subscribe('blocks', (block) => {
  console.log('New block:', block);
});
```

### Python
```python
# Установка
pip install ganymed-sdk

# Инициализация
from ganymed import Ganymed

gnd = Ganymed(
    rpc_url='http://31.128.41.155:8181',
    ws_url='ws://31.128.41.155:8183/ws',
    api_key='your-api-key'
)

# Создание кошелька
wallet = gnd.wallet.create()

# Получение баланса
balance = gnd.wallet.get_balance(wallet.address)

# Отправка транзакции
tx = gnd.wallet.send(
    from_=wallet.address,
    to='GND...',
    value='1',
    unit='ether'
)

# Подписка на события
@gnd.ws.on('blocks')
def on_block(block):
    print('New block:', block)
```

### Go
```go
// Установка
go get github.com/ganymed/sdk

// Инициализация
package main

import (
    "github.com/ganymed/sdk"
)

func main() {
    gnd := sdk.NewGanymed(&sdk.Config{
        RPCURL:  "http://31.128.41.155:8181",
        WSURL:   "ws://31.128.41.155:8183/ws",
        APIKey:  "your-api-key",
    })

    // Создание кошелька
    wallet, err := gnd.Wallet.Create()
    if err != nil {
        panic(err)
    }

    // Получение баланса
    balance, err := gnd.Wallet.GetBalance(wallet.Address)
    if err != nil {
        panic(err)
    }

    // Отправка транзакции
    tx, err := gnd.Wallet.Send(&sdk.Transaction{
        From:  wallet.Address,
        To:    "GND...",
        Value: "1",
        Unit:  "ether",
    })
    if err != nil {
        panic(err)
    }

    // Подписка на события
    gnd.WS.Subscribe("blocks", func(block *sdk.Block) {
        println("New block:", block)
    })
}
```

### Java
```java
// Установка
<dependency>
    <groupId>com.ganymed</groupId>
    <artifactId>sdk</artifactId>
    <version>1.0.0</version>
</dependency>

// Инициализация
import com.ganymed.sdk.Ganymed;
import com.ganymed.sdk.Config;

public class Main {
    public static void main(String[] args) {
        Ganymed gnd = new Ganymed(new Config(
            "http://31.128.41.155:8181",
            "ws://31.128.41.155:8183/ws",
            "your-api-key"
        ));

        // Создание кошелька
        Wallet wallet = gnd.wallet().create();

        // Получение баланса
        Balance balance = gnd.wallet().getBalance(wallet.getAddress());

        // Отправка транзакции
        Transaction tx = gnd.wallet().send(new Transaction(
            wallet.getAddress(),
            "GND...",
            "1",
            "ether"
        ));

        // Подписка на события
        gnd.ws().subscribe("blocks", block -> {
            System.out.println("New block: " + block);
        });
    }
}
```

## Инструменты

### CLI
```bash
# Установка
curl -L https://github.com/ganymed/cli/releases/latest/download/gnd-cli -o gnd-cli
chmod +x gnd-cli
sudo mv gnd-cli /usr/local/bin/

# Конфигурация
gnd-cli config init
gnd-cli config set api-key your-api-key
gnd-cli config set network mainnet
gnd-cli config set rpc-url http://31.128.41.155:8181
gnd-cli config set ws-url ws://31.128.41.155:8183/ws

# Создание кошелька
gnd-cli wallet create

# Получение баланса
gnd-cli wallet balance GND...

# Отправка транзакции
gnd-cli wallet send --from GND... --to GND... --value 1 --unit ether
```

### GUI
```bash
# Установка
curl -L https://github.com/ganymed/gui/releases/latest/download/gnd-gui -o gnd-gui
chmod +x gnd-gui
sudo mv gnd-gui /usr/local/bin/

# Запуск
gnd-gui
```

### Аналитика
```bash
# Установка
curl -L https://github.com/ganymed/analytics/releases/latest/download/gnd-analytics -o gnd-analytics
chmod +x gnd-analytics
sudo mv gnd-analytics /usr/local/bin/

# Запуск
gnd-analytics
```

## API

### REST API
```bash
# Базовый URL
http://31.128.41.155:8182/api/v1

# Аутентификация
curl -H "X-API-Key: your-api-key" http://31.128.41.155:8182/api/v1/wallet/balance/GND...

# Создание кошелька
curl -X POST -H "X-API-Key: your-api-key" http://31.128.41.155:8182/api/v1/wallet/create

# Получение баланса
curl -H "X-API-Key: your-api-key" http://31.128.41.155:8182/api/v1/wallet/balance/GND...

# Отправка транзакции
curl -X POST -H "X-API-Key: your-api-key" -H "Content-Type: application/json" \
  -d '{"from":"GND...","to":"GND...","value":"1","unit":"ether"}' \
  http://31.128.41.155:8182/api/v1/wallet/send
```

### WebSocket API
```javascript
// Подключение
const ws = new WebSocket('ws://31.128.41.155:8183/ws');

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

// Обработка сообщений
ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Message:', data);
};
```

### JSON-RPC API
```javascript
// Базовый URL
http://31.128.41.155:8181

// Аутентификация
curl -X POST -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
  http://31.128.41.155:8181

// Получение баланса
curl -X POST -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{"jsonrpc":"2.0","method":"eth_getBalance","params":["GND...","latest"],"id":1}' \
  http://31.128.41.155:8181

// Отправка транзакции
curl -X POST -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{"jsonrpc":"2.0","method":"eth_sendTransaction","params":[{"from":"GND...","to":"GND...","value":"0xde0b6b3a7640000"}],"id":1}' \
  http://31.128.41.155:8181
```

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

---