# Токены в блокчейне ГАНИМЕД

## Обзор

Блокчейн ГАНИМЕД поддерживает стандарт токенов GNDst-1, который расширяет функциональность ERC-20 и TRC-20 стандартов.

## GNDst-1 стандарт

### Базовые функции

#### Баланс и переводы
```solidity
function balanceOf(address account) external view returns (uint256);
function transfer(address to, uint256 amount) external returns (bool);
function transferFrom(address from, address to, uint256 amount) external returns (bool);
```

#### Разрешения
```solidity
function approve(address spender, uint256 amount) external returns (bool);
function allowance(address owner, address spender) external view returns (uint256);
function increaseAllowance(address spender, uint256 addedValue) external returns (bool);
function decreaseAllowance(address spender, uint256 subtractedValue) external returns (bool);
```

### Расширенные функции

#### Снимки (Snapshots)
```solidity
function snapshot() external returns (uint256);
function getSnapshotBalance(uint256 snapshotId, address account) external view returns (uint256);
```

#### Дивиденды
```solidity
function claimDividends() external returns (bool);
function getDividends(address account) external view returns (uint256);
```

#### Модули
```solidity
function registerModule(address module, string memory name) external returns (bool);
function moduleCall(address module, bytes memory data) external returns (bool);
```

#### KYC и безопасность
```solidity
function setKycStatus(address account, bool status) external returns (bool);
function isKycPassed(address account) external view returns (bool);
function pause() external returns (bool);
function unpause() external returns (bool);
```

## Создание токена

### Через API
```http
POST /token/create
Content-Type: application/json

{
    "name": "My Token",
    "symbol": "MTK",
    "decimals": 18,
    "initialSupply": "1000000000000000000000000",
    "owner": "GND..."
}
```

### Через смарт-контракт
```solidity
contract MyToken is GNDst1 {
    constructor(
        string memory name,
        string memory symbol,
        uint8 decimals,
        uint256 initialSupply,
        address owner
    ) GNDst1(name, symbol, decimals, initialSupply, owner) {}
}
```

## Операции с токенами

### Получение баланса
```http
GET /token/balance/{address}
```

**Ответ:**
```json
{
    "status": "success",
    "data": {
        "balance": "1000000000000000000"
    }
}
```

### Перевод токенов
```http
POST /token/transfer
Content-Type: application/json

{
    "tokenAddress": "GND...",
    "from": "GND...",
    "to": "GND...",
    "amount": "1000000000000000000",
    "privateKey": "..."
}
```

### Подтверждение токенов
```http
POST /token/approve
Content-Type: application/json

{
    "tokenAddress": "GND...",
    "owner": "GND...",
    "spender": "GND...",
    "amount": "1000000000000000000",
    "privateKey": "..."
}
```

## Снимки и дивиденды

### Создание снимка
```http
POST /token/snapshot
Content-Type: application/json

{
    "tokenAddress": "GND...",
    "privateKey": "..."
}
```

### Получение баланса в снимке
```http
GET /token/snapshot/{snapshotId}/balance/{address}
```

### Получение дивидендов
```http
POST /token/claim-dividends
Content-Type: application/json

{
    "tokenAddress": "GND...",
    "account": "GND...",
    "privateKey": "..."
}
```

## Модули

### Регистрация модуля
```http
POST /token/register-module
Content-Type: application/json

{
    "tokenAddress": "GND...",
    "module": "GND...",
    "name": "MyModule",
    "privateKey": "..."
}
```

### Вызов модуля
```http
POST /token/module-call
Content-Type: application/json

{
    "tokenAddress": "GND...",
    "module": "GND...",
    "data": "0x...",
    "privateKey": "..."
}
```

## KYC и безопасность

### Установка KYC статуса
```http
POST /token/set-kyc
Content-Type: application/json

{
    "tokenAddress": "GND...",
    "account": "GND...",
    "status": true,
    "privateKey": "..."
}
```

### Проверка KYC статуса
```http
GET /token/kyc/{address}
```

### Пауза/возобновление
```http
POST /token/pause
Content-Type: application/json

{
    "tokenAddress": "GND...",
    "privateKey": "..."
}
```

## События

### Transfer
```solidity
event Transfer(address indexed from, address indexed to, uint256 value);
```

### Approval
```solidity
event Approval(address indexed owner, address indexed spender, uint256 value);
```

### Snapshot
```solidity
event Snapshot(uint256 indexed id);
```

### DividendsClaimed
```solidity
event DividendsClaimed(address indexed account, uint256 amount);
```

### ModuleRegistered
```solidity
event ModuleRegistered(address indexed module, string name);
```

### KycStatusChanged
```solidity
event KycStatusChanged(address indexed account, bool status);
```

### Paused/Unpaused
```solidity
event Paused(address account);
event Unpaused(address account);
```

## Безопасность

### Рекомендации
- Использовать KYC для всех операций
- Проверять модули перед регистрацией
- Осторожно работать со снимками
- Контролировать дивиденды

### Ограничения
- Только владелец может паузить токен
- Только владелец может регистрировать модули
- Только владелец может менять KYC статусы
- Требуется KYC для всех операций

## Примеры использования

### JavaScript
```javascript
const token = new GNDst1("0x...");

// Получение баланса
const balance = await token.balanceOf("0x...");

// Перевод токенов
await token.transfer("0x...", "1000000000000000000");

// Подтверждение токенов
await token.approve("0x...", "1000000000000000000");

// Создание снимка
const snapshotId = await token.snapshot();

// Получение дивидендов
await token.claimDividends();

// Регистрация модуля
await token.registerModule("0x...", "MyModule");

// Вызов модуля
await token.moduleCall("0x...", "0x...");
```

### Python
```python
from gndst1 import GNDst1

token = GNDst1("0x...")

# Получение баланса
balance = token.balance_of("0x...")

# Перевод токенов
token.transfer("0x...", "1000000000000000000")

# Подтверждение токенов
token.approve("0x...", "1000000000000000000")

# Создание снимка
snapshot_id = token.snapshot()

# Получение дивидендов
token.claim_dividends()

# Регистрация модуля
token.register_module("0x...", "MyModule")

# Вызов модуля
token.module_call("0x...", "0x...")
```

---

## 1. Поддерживаемые стандарты токенов

### 1.1. ERC-20

- Полная совместимость с экосистемой Ethereum.
- Поддержка стандартных методов: `totalSupply`, `balanceOf`, `transfer`, `approve`, `transferFrom`, `allowance`.
- Интеграция с кошельками и биржами, использующими стандарт ERC-20.

### 1.2. TRC-20

- Аналогичный стандарт для интеграции с Tron и другими сетями.
- Методы и структура идентичны ERC-20.

### 1.3. Кастомные стандарты

- Возможность создания собственных токенов с уникальной бизнес-логикой и интерфейсом.
- Метаданные контракта содержат описание стандарта (`"standard": "custom"`) и необходимые параметры.

---

## 2. Универсальный интерфейс токенов

- Все токены реализуют единый интерфейс, позволяющий добавлять новые стандарты без изменений основного кода блокчейна.
- В метаданных контракта указывается стандарт: `"erc20"`, `"trc20"` или `"custom"`.
- Контракты могут одновременно поддерживать несколько стандартов.

**Пример метаданных токена:**
{
"standard": "erc20",
"name": "MyToken",
"symbol": "MTK",
"decimals": 18
}

text

---

## 3. Работа с токенами через API

### 3.1. Деплой токена

**REST:**
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

### 3.2. Вызов метода токена

**REST:**
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

### 3.3. Получить информацию о токене

**JSON-RPC:**
{
"jsonrpc": "2.0",
"method": "token_getInfo",
"params": { "address": "GNDct1..." },
"id": 1
}

text
**Ответ:**
{
"jsonrpc": "2.0",
"result": {
"standard": "erc20",
"name": "MyToken",
"symbol": "MTK",
"decimals": 18,
"totalSupply": "1000000"
},
"id": 1
}

text

---

## 4. Примеры использования

### 4.1. Деплой ERC-20 токена

1. Скомпилируйте контракт на Solidity.
2. Отправьте байткод и метаданные через `/contract/deploy`.
3. Получите адрес токена и используйте его для операций transfer, approve и др.

### 4.2. Деплой кастомного токена

1. Определите собственные методы и события в контракте.
2. Укажите `"standard": "custom"` в метаданных.
3. После деплоя используйте методы согласно вашей бизнес-логике.

---

## 5. Комиссии и привязка токенов

- Все операции (деплой, трансфер, взаимодействие с контрактом) оплачиваются в GND.
- Гибкая настройка комиссий (gas fees) в зависимости от сложности операции.
- Возможна привязка токенов к GND для обмена или управления средствами (например, через мосты или специальные контракты).

---

## 6. Безопасность и валидация

- Перед деплоем обязательна цифровая подпись владельца.
- Контракт проходит валидацию на соответствие стандарту.
- Все вызовы функций токенов проходят через VM с контролем газа и sandbox-изоляцией.

---

## 7. Расширение стандартов

- Для добавления нового стандарта реализуйте интерфейс токена в модуле tokens/.
- Зарегистрируйте стандарт через метаданные.
- Контракты могут поддерживать сразу несколько стандартов (например, ERC-20 + custom).

---

## 8. Тестирование и аудит

- Для всех токенов реализуются юнит- и интеграционные тесты.
- Проводится аудит безопасности, особенно для кастомных стандартов и валидации байткода.

---

## 9. Ссылки

- [api.md](api.md) - описание API для работы с токенами
- [contracts.md](contracts.md) - описание работы со смарт-контрактами
- [architecture.md](architecture.md) - архитектура блокчейна
- [consensus.md](consensus.md) - алгоритмы консенсуса
- [contracts.md](contracts.md) - работа со смарт-контрактами
- [integration.md](integration.md) - интеграция с GND
---