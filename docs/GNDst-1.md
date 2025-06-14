## GNDst-1: Стандарт мультицепочечных токенов для блокчейна «ГАНИМЕД»

### 1. Введение

**GNDst-1** — открытый стандарт токенов для блокчейна «ГАНИМЕД», обеспечивающий полную совместимость с ERC-20 (Ethereum) и TRC-20 (Tron), а также расширяемость для поддержки новых функций, кроссчейн-операций, модульности и встроенного KYC. Стандарт предназначен для DeFi, DAO, NFT и корпоративных решений[^1].

---

### 2. Цели стандарта

- Совместимость с ERC-20 и TRC-20 для легкой интеграции с существующими инструментами и биржами.
- Поддержка кроссчейн-переводов и мостов.
- Встроенная система KYC/AML.
- Модульная архитектура для расширения функционала без изменения ядра токена.
- Поддержка снимков балансов (snapshot) для голосований и дивидендов[^1].
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
| crossChainTransfer(chain, to, amt) | Кроссчейн-перевод через мост |
| setKycStatus(user, status) | Установка KYC-статуса адреса (только owner) |
| isKycPassed(user) | Проверка KYC-статуса |
| moduleCall(moduleId, data) | Вызов внешнего модуля (расширяемость) |
| snapshot() | Создание снимка балансов (только owner) |
| getSnapshotBalance(user, snapshotId) | Получение баланса на момент снимка |
| claimDividends(snapshotId) | Получение дивидендов по снимку |
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
- Кроссчейн-функции реализуются через мостовые контракты, указанные в параметрах[^1].

---

### 6. Безопасность и KYC

- Все методы перевода и списания требуют прохождения KYC (`onlyKyc`).
- Управление статусом KYC осуществляется только владельцем токена (owner).
- Возможна интеграция с внешними KYC-провайдерами через модульную систему[^1].

---

### 7. Модульность

- Вызовы `moduleCall` позволяют подключать новые функции без обновления основного контракта.
- Модули могут быть реализованы как отдельные контракты и регистрироваться в ядре блокчейна через событие `ModuleRegistered`.
- Регистрация модулей осуществляется только владельцем токена для обеспечения безопасности и целостности системы.

---

### 8. Снимки (Snapshot) и дивиденды

- Система снимков балансов позволяет реализовать голосования, дивиденды и другие DAO-функции.
- Снимки создаются только владельцем токена, событие `SnapshotCreated` фиксирует факт создания снимка.
- Дивиденды могут быть распределены на основе снимков, пользователи получают выплаты через функцию `claimDividends`, событие `DividendClaimed` фиксирует факт выплаты.

---

### 9. Пример интерфейса (Solidity)

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

Стандарт GNDst-1 является открытым и может свободно использоваться для разработки токенов и сервисов в экосистеме «ГАНИМЕД»[^1].

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
Поддержка переводов между разными блокчейнами:
```solidity
function crossChainTransfer(string calldata targetChain, address to, uint256 amount) external returns (bool);
```

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

1. Все операции с токенами должны проходить через KYC
2. Модули должны быть тщательно проверены перед регистрацией
3. Кроссчейн-операции должны использовать проверенные мосты
4. Снимки должны создаваться только владельцем токена
5. Дивиденды должны распределяться с учетом всех держателей токенов

---

**GNDst-1 — стандарт для будущего мультицепочечных, модульных и безопасных токенов!**
[^1]

<div style="text-align: center">⁂</div>

[^1]: GNDst-1_-Standart-multitsepochechnykh-tokenov-dlia-blo.md

[^2]: https://www-nds.iaea.org/publications/indc/indc-iae-asterisk041D.pdf

[^3]: https://github.com/njoy/GNDStk/blob/master/docs/motive.rst

[^4]: https://www.scribd.com/document/694172439/Pioneer-Vsx-lx50-91txh-9120txh-k

[^5]: https://docs.audio-technica.com/all/3000IEM_IP_Control_Protocol_Specifications_V1_EN_web_240203.pdf

[^6]: https://inldigitallibrary.inl.gov/Reports/ANL-EAD-1.pdf

[^7]: https://www.scribd.com/document/508513115/PIONEER-VSX-1016TXV

[^8]: https://offices.mtholyoke.edu/sites/default/files/registrar/docs/2017-18Bulletin-Catalog.pdf

[^9]: https://github.com/gnistdesign/gdcs

[^10]: https://inis.iaea.org/records/x557e-4s888/files/27047423.pdf?download=1

[^11]: https://github.com/gnistdesign

