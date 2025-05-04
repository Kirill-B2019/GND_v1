# API блокчейна «ГАНИМЕД»

## Общие сведения

- Все взаимодействие с сетью осуществляется через публичные API (REST, JSON-RPC, WebSocket).
- Все комиссии и операции оплачиваются в GND.
- Для доступа к приватным методам требуется аутентификация (API-ключ, JWT, OAuth2).

---

## 1. REST API

### 1.1. Получение информации о блоках

#### Получить последний блок

GET /block/latest

text

**Пример ответа:**
{
"index": 12345,
"timestamp": 1714820000,
"prevHash": "0xabc...",
"hash": "0xdef...",
"miner": "GND1...",
"transactions": [ ... ],
"gasUsed": 90000,
"gasLimit": 1000000,
"consensus": "pos"
}

text

#### Получить блок по хешу

GET /block/{hash}

text

---

### 1.2. Транзакции

#### Отправить транзакцию

POST /tx/send
Content-Type: application/json

text
**Тело запроса:**
{
"from": "GND1...",
"to": "GND2...",
"value": 100,
"gasPrice": 1,
"gasLimit": 21000,
"nonce": 5,
"type": "transfer",
"data": "",
"signature": "..."
}

text
**Ответ:**
{
"txHash": "0x123..."
}

text

#### Получить транзакцию по хешу

GET /tx/{hash}

text

---

### 1.3. Баланс и nonce

#### Получить баланс

GET /account/{address}/balance

text
**Ответ:**
{ "balance": 100000 }

text

#### Получить nonce

GET /account/{address}/nonce

text
**Ответ:**
{ "nonce": 6 }

text

---

### 1.4. Деплой и вызов смарт-контракта

#### Деплой контракта

POST /contract/deploy
Content-Type: application/json

text
**Тело запроса:**
{
"from": "GND1...",
"bytecode": "<hex>",
"gasLimit": 2000000,
"gasPrice": 1,
"nonce": 7,
"metadata": {
"standard": "erc20",
"name": "MyToken",
"symbol": "MTK",
"decimals": 18
},
"signature": "..."
}

text
**Ответ:**
{
"contractAddress": "GNDct1..."
}

text

#### Вызов функции контракта

POST /contract/call
Content-Type: application/json

text
**Тело запроса:**
{
"from": "GND1...",
"to": "GNDct1...",
"data": "<hex>",
"gasLimit": 80000,
"gasPrice": 1,
"nonce": 8,
"signature": "..."
}

text
**Ответ:**
{
"result": "...",
"gasUsed": 50000
}

text

---

## 2. JSON-RPC API

- Все методы вызываются через POST на `/rpc`
- Формат запроса соответствует стандарту JSON-RPC 2.0

### Пример запроса

{
"jsonrpc": "2.0",
"method": "blockchain_latestBlock",
"params": {},
"id": 1
}

text

### Пример ответа

{
"jsonrpc": "2.0",
"result": { ... },
"id": 1
}

text

---

### Основные методы

#### Получить последний блок

- **method:** `blockchain_latestBlock`
- **params:** `{}`

#### Получить блок по хешу

- **method:** `blockchain_getBlockByHash`
- **params:** `{ "hash": "0x..." }`

#### Получить баланс

- **method:** `state_getBalance`
- **params:** `{ "address": "GND1..." }`

#### Отправить транзакцию

- **method:** `blockchain_sendTx`
- **params:** (см. REST-пример выше)

#### Деплой контракта

- **method:** `contract_deploy`
- **params:** (см. REST-пример выше)

#### Вызов функции контракта

- **method:** `contract_call`
- **params:** (см. REST-пример выше)

#### Получить информацию о токене

- **method:** `token_getInfo`
- **params:** `{ "address": "GNDct1..." }`

#### Вызов метода токена

- **method:** `token_call`
- **params:** `{ "tokenAddress": "...", "method": "transfer", "args": ["GND2...", 100] }`

---

## 3. WebSocket API

- Подключение: `ws://<host>:8090/ws`
- Сообщения приходят в формате:
  {
  "type": "block" | "tx" | "event",
  "data": { ... }
  }

text
- Можно реализовать подписки на адреса, события, типы токенов (расширяется через будущие версии API).

---

## 4. Аутентификация и лимитирование

- Для приватных методов используйте заголовок `X-API-Key`.
- Для публичных методов лимит по IP (100 запросов в минуту по умолчанию).
- Для production рекомендуется использовать JWT или OAuth2.

---

## 5. Примеры сценариев

### 5.1. Деплой ERC-20 токена

1. Скомпилируйте контракт на Solidity (например, в Remix).
2. Отправьте байткод через `/contract/deploy` с метаданными:
    - `standard: "erc20"`
    - `name`, `symbol`, `decimals`
3. Получите адрес контракта и используйте его для работы с токеном.

### 5.2. Вызов метода transfer у токена

1. Получите ABI и адрес токена.
2. Сформируйте calldata (например, через web3.js или ganymede-cli).
3. Отправьте через `/contract/call` или `contract_call` в JSON-RPC.

---

## 6. Ошибки API

- Все ошибки возвращаются с HTTP-кодом 4xx/5xx и полем `error` или в формате JSON-RPC:
  {
  "jsonrpc": "2.0",
  "error": { "code": -32000, "message": "Insufficient funds" },
  "id": 1
  }

text

---

## 7. Расширение и кастомизация

- Для добавления новых стандартов токенов реализуйте интерфейс в модуле токенов.
- Для интеграции с внешними сервисами используйте REST/JSON-RPC/WebSocket API.
- Для мониторинга используйте WebSocket и методы `/metrics` (будет реализовано).

---

## 8. Безопасность

- Все приватные методы требуют аутентификации.
- Все транзакции должны быть подписаны приватным ключом отправителя.
- Для деплоя и вызова контрактов обязательно списание комиссии в GND.

---

## 9. Дополнительные ресурсы

- [architecture.md](architecture.md) - описание архитектуры блокчейна
- [contracts.md](contracts.md) - руководство по работе со смарт-контрактами
- [tokens.md](tokens.md) - описание стандартов токенов
- [consensus.md](consensus.md) - инструкции по переключению алгоритмов консенсуса
- [integration.md](integration.md) - интеграция с мостами, оракулами, IPFS

---

**API блокчейна «ГАНИМЕД» поддерживает все современные сценарии работы с токенами, смарт-контрактами и внешними сервисами. 