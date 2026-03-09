## GNDst-1: Стандарт мультицепочечных токенов для блокчейна «ГАНИМЕД»

### 1. Введение

**GNDst-1** — открытый стандарт токенов для блокчейна «ГАНИМЕД», обеспечивающий полную совместимость с ERC-20 (Ethereum) и TRC-20 (Tron), а также расширяемость для поддержки новых функций, кроссчейн-операций, модульности и встроенного KYC. Стандарт предназначен для DeFi, DAO, NFT и корпоративных решений.

В проекте реализованы:
- **Нативная реализация (Go):** пакет `tokens/standards/gndst1/` — структура `GNDst1`, все методы стандарта, хранение балансов/allowances/snapshots/модулей/KYC в памяти; идентификатор стандарта в API и БД: **`GND-st1`**.
- **Референсный контракт (Solidity):** `tokens/standards/gndst1/gndst1Base.sol` — интерфейс и контракт для совместимости с EVM-инструментами.
- **Деплой:** через `vm.DeployGNDst1Token` и REST API (`POST /api/v1/token/deploy`); после деплоя токен регистрируется в `tokens/registry` и владельцу назначается начальный баланс через `SetInitialBalance`.

---

### 2. Цели стандарта

- Совместимость с ERC-20 и TRC-20 для легкой интеграции с существующими инструментами и биржами.
- Поддержка кроссчейн-переводов и мостов.
- Встроенная система KYC/AML.
- Модульная архитектура для расширения функционала без изменения ядра токена.
- Поддержка снимков балансов (snapshot) для голосований и дивидендов.
- Встроенная система дивидендов и событий для управления модулями.

---

### 3. Интерфейс токена

#### 3.1. Базовые методы GNDst-1

| Метод | Описание |
| :-- | :-- |
| totalSupply() | Общее предложение токенов |
| balanceOf(addr) | Баланс адреса |
| transfer(to, amt) | Перевод токенов |
| approve(spender, amt) | Разрешение на списание |
| allowance(owner, spender) | Лимит разрешения |
| transferFrom(from, to, amt) | Перевод с разрешения |

#### 3.2. Расширенные методы GNDst-1

| Метод | Описание |
| :-- | :-- |
| crossChainTransfer(targetChain, to, amt) | Кроссчейн-перевод через мост (в Solidity — только для прошедших KYC) |
| setKycStatus(user, status) | Установка KYC-статуса адреса (только owner) |
| isKycPassed(user) | Проверка KYC-статуса |
| moduleCall(moduleId, data) | Вызов внешнего модуля (расширяемость); в Solidity — moduleId bytes32, в Go API — string |
| snapshot() | Создание снимка балансов (только owner); возвращает ID снимка (uint256 в Solidity, uint64 в Go) |
| getSnapshotBalance(user, snapshotId) | Получение баланса на момент снимка |
| claimDividends(snapshotId) | Получение дивидендов по снимку вызывающим абонентом |
| registerModule(moduleId, moduleAddress, name) | Регистрация нового модуля (только owner) |


---

### 4. События

- `Transfer(address indexed from, address indexed to, uint256 value)`
- `Approval(address indexed owner, address indexed spender, uint256 value)`
- `CrossChainTransfer(address indexed from, string targetChain, address indexed to, uint256 value)`
- `KycStatusChanged(address indexed user, bool status)`
- `ModuleCall(bytes32 indexed moduleId, address indexed caller)`
- `SnapshotCreated(uint256 indexed snapshotId, uint256 timestamp)`
  *Создается при фиксации состояния балансов для snapshot-функций и дивидендов.*
- `DividendClaimed(address indexed user, uint256 amount, uint256 snapshotId)`
  *Фиксирует успешное получение дивидендов пользователем за определённый снимок.*
- `ModuleRegistered(bytes32 indexed moduleId, address indexed moduleAddress, string name)`
  *Фиксирует регистрацию нового модуля в системе.*

---

### 5. Требования совместимости

- Все методы и события ERC-20/TRC-20 должны быть реализованы с идентичными сигнатурами.
- Контракт должен поддерживать работу с адресами обоих форматов (Ethereum/Tron).
- Кроссчейн-функции реализуются через мостовые контракты, указанные в параметрах.

---

### 6. Безопасность и KYC

- По спецификации (Solidity) методы `transfer`, `transferFrom` и `crossChainTransfer` должны выполняться только для адресов, прошедших KYC (`onlyKyc`).
- Управление статусом KYC осуществляется только владельцем токена (owner) через `setKycStatus`.
- В нативной реализации (Go) проверка KYC при переводах и проверка owner для `setKycStatus`/`snapshot`/`registerModule` запланированы; текущая версия не ограничивает вызовы по ролям.
- Возможна интеграция с внешними KYC-провайдерами через модульную систему.

---

### 7. Модульность

- Вызовы `moduleCall` позволяют подключать новые функции без обновления основного контракта.
- Модули регистрируются через `registerModule(moduleId, moduleAddress, name)`; в Solidity идентификатор модуля — `bytes32`, в Go API — строка.
- Модули могут быть реализованы как отдельные контракты; событие `ModuleRegistered` фиксирует регистрацию.
- Регистрация модулей осуществляется только владельцем токена (в Solidity — `onlyOwner`). В нативной реализации (Go) вызов внешнего модуля через `moduleCall` пока возвращает «module call not implemented»; регистрация модуля сохраняет данные в памяти.

---

### 8. Снимки (Snapshot) и дивиденды

- Система снимков балансов позволяет реализовать голосования, дивиденды и другие DAO-функции.
- Снимки создаются только владельцем токена (в Solidity — `onlyOwner`); событие `SnapshotCreated` фиксирует факт создания снимка.
- Идентификатор снимка в интерфейсе Solidity — `uint256`; в нативной реализации Go — `uint64`.
- Дивиденды задаются на снимок (в Go — поле `dividends[snapshotId]`). Пользователи получают выплаты через `claimDividends(snapshotId)`; доля вычисляется по балансу вызывающего на момент снимка. Событие `DividendClaimed` фиксирует факт выплаты.

---

### 9. Пример интерфейса (Solidity)

Интерфейс и референсная реализация с модификаторами `onlyOwner` (setKycStatus, snapshot, registerModule) и `onlyKyc` (transfer, transferFrom, crossChainTransfer) — в файле `tokens/standards/gndst1/gndst1Base.sol`.

```solidity
interface IGNDst1 {
    function totalSupply() external view returns (uint256);
    function balanceOf(address account) external view returns (uint256);
    function transfer(address to, uint256 amount) external returns (bool);
    function approve(address spender, uint256 amount) external returns (bool);
    function allowance(address owner, address spender) external view returns (uint256);
    function transferFrom(address from, address to, uint256 amount) external returns (bool);

    // GNDst-1 расширения
    function crossChainTransfer(string calldata targetChain, address to, uint256 amount) external returns (bool);
    function setKycStatus(address user, bool status) external;
    function isKycPassed(address user) external view returns (bool);
    function moduleCall(bytes32 moduleId, bytes calldata data) external returns (bytes memory);
    function snapshot() external returns (uint256);
    function getSnapshotBalance(address user, uint256 snapshotId) external view returns (uint256);
    function claimDividends(uint256 snapshotId) external;
    function registerModule(bytes32 moduleId, address moduleAddress, string calldata name) external;

    // События
    event SnapshotCreated(uint256 indexed snapshotId, uint256 timestamp);
    event DividendClaimed(address indexed user, uint256 amount, uint256 snapshotId);
    event ModuleRegistered(bytes32 indexed moduleId, address indexed moduleAddress, string name);
}
```


---

### 10. Расширяемость

- Стандарт допускает добавление новых функций через модульную систему без хардфорка.
- Все расширения должны быть задокументированы и совместимы с ядром GNDst-1.

---

### 11. Лицензия и открытость

Стандарт GNDst-1 является открытым и может свободно использоваться для разработки токенов и сервисов в экосистеме «ГАНИМЕД».

---

### 12. Расширенные функции GNDst-1

#### 12.1. Система снимков (Snapshots)
Снимки позволяют фиксировать состояние балансов токенов в определенный момент времени. Это полезно для:
- Голосований в DAO
- Распределения дивидендов
- Аудита и отчетности

```solidity
function snapshot() external returns (uint256);
function getSnapshotBalance(address user, uint256 snapshotId) external view returns (uint256);
```

#### 12.2. Система дивидендов
Позволяет распределять дивиденды между держателями токенов на основе снимков:
```solidity
function claimDividends(uint256 snapshotId) external;
```

#### 12.3. Модульная система
Позволяет расширять функционал токена без изменения основного контракта:
```solidity
function moduleCall(bytes32 moduleId, bytes calldata data) external returns (bytes memory);
function registerModule(bytes32 moduleId, address moduleAddress, string calldata name) external;
```

#### 12.4. KYC и безопасность
Все операции с токенами требуют прохождения KYC:
```solidity
function setKycStatus(address user, bool status) external;
function isKycPassed(address user) external view returns (bool);
```

#### 12.5. Кроссчейн-операции
Поддержка переводов между разными блокчейнами (в референсном контракте — списание с отправителя на адрес моста и эмиссия события):
```solidity
function crossChainTransfer(string calldata targetChain, address to, uint256 amount) external returns (bool);
```
В нативной реализации (Go) списание выполняется с адреса контракта на `to`; полная интеграция с мостом и параметр `targetChain` — в разработке.

### 13. Примеры использования

#### 13.1. Создание снимка и распределение дивидендов
```solidity
// Создание снимка
uint256 snapshotId = token.snapshot();

// Распределение дивидендов
token.claimDividends(snapshotId);
```

#### 13.2. Регистрация и использование модуля
```solidity
// Регистрация модуля
token.registerModule(keccak256("voting"), votingModuleAddress, "Voting Module");

// Вызов метода модуля
bytes memory result = token.moduleCall(keccak256("voting"), abi.encode("vote", proposalId, true));
```

#### 13.3. Кроссчейн-перевод
```solidity
// Перевод токенов в другую сеть
token.crossChainTransfer("ethereum", recipientAddress, amount);
```

### 14. Безопасность и рекомендации

1. Все операции с токенами должны проходить через KYC (в референсном контракте — модификатор `onlyKyc`).
2. Модули должны быть тщательно проверены перед регистрацией.
3. Кроссчейн-операции должны использовать проверенные мосты.
4. Снимки должны создаваться только владельцем токена (`onlyOwner`).
5. Дивиденды должны распределяться с учётом балансов на снимке и корректной настройки пула дивидендов.

---

### 15. Соответствие реализации коду

| Аспект | Реализация |
|--------|------------|
| **Пакет Go** | `tokens/standards/gndst1/gndst1.go` — структура `GNDst1`, все методы из разд. 3. |
| **Регистрация** | `tokens/registry` хранит экземпляры по адресу; после деплоя вызывается `SetInitialBalance` для владельца. |
| **Вызов методов** | Через `core.Token.UniversalCall`: поддерживаются `transfer`, `approve`, `balanceOf`. Остальные методы вызываются напрямую через экземпляр GNDst1. |
| **События** | В стандарте определены Transfer, Approval и др. В Go реализация **EmitTransfer** и **EmitApproval** записывает события в таблицу БД `events` (типы `Transfer`, `Approval`; поля contract, from_address, to_address, amount, timestamp) и опционально уведомляет подписчиков WebSocket API (порт 8183) через callback `TokenEventNotifier`, устанавливаемый при создании REST-сервера. Это позволяет фронтендам и индексаторам получать историю переводов и разрешений в реальном времени без опроса REST. |
| **Доп. метод** | В Go реализован `BridgeTransfer(ctx, amount)` для перевода через мост (внутреннее использование). |
| **Кроссчейн** | `CrossChainTransfer` в Go списывает средства с адреса контракта на указанный `to`; полная интеграция с мостом — в разработке. |

---



<div style="text-align: center">| KB @CerbeRus - Nexus Invest Team 2026</div>
