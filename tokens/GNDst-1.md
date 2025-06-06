

## GNDst-1: Стандарт мультицепочечных токенов для блокчейна «ГАНИМЕД»

### 1. Введение

**GNDst-1** — открытый стандарт токенов для блокчейна «ГАНИМЕД», обеспечивающий полную совместимость с ERC-20 (Ethereum) и TRC-20 (Tron), а также расширяемость для поддержки новых функций, кроссчейн-операций, модульности и встроенного KYC. Стандарт предназначен для DeFi, DAO, NFT и корпоративных решений.

---

### 2. Цели стандарта

- Совместимость с ERC-20 и TRC-20 для легкой интеграции с существующими инструментами и биржами.
- Поддержка кроссчейн-переводов и мостов.
- Встроенная система KYC/AML.
- Модульная архитектура для расширения функционала без изменения ядра токена.
- Поддержка снимков балансов (snapshot) для голосований и дивидендов.

---

### 3. Интерфейс токена

#### 3.1. Базовые методы (ERC-20/TRC-20)

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


---

### 4. События

- `Transfer(address indexed from, address indexed to, uint256 value)`
- `Approval(address indexed owner, address indexed spender, uint256 value)`
- `CrossChainTransfer(address indexed from, string targetChain, address indexed to, uint256 value)`
- `KycStatusChanged(address indexed user, bool status)`
- `ModuleCall(bytes32 indexed moduleId, address indexed caller)`

---

### 5. Требования совместимости

- Все методы и события ERC-20/TRC-20 должны быть реализованы с идентичными сигнатурами.
- Контракт должен поддерживать работу с адресами обоих форматов (Ethereum/Tron).
- Кроссчейн-функции реализуются через мостовые контракты, указанные в параметрах.

---

### 6. Безопасность и KYC

- Все методы перевода и списания требуют прохождения KYC (`onlyKyc`).
- Управление статусом KYC осуществляется только владельцем токена (owner).
- Возможна интеграция с внешними KYC-провайдерами через модульную систему.

---

### 7. Модульность

- Вызовы `moduleCall` позволяют подключать новые функции без обновления основного контракта.
- Модули могут быть реализованы как отдельные контракты и регистрироваться в ядре блокчейна.

---

### 8. Снимки (Snapshot)

- Система снимков балансов позволяет реализовать голосования, дивиденды и другие DAO-функции.
- Снимки создаются только владельцем токена.

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

    // GNDst-1
    function crossChainTransfer(string calldata targetChain, address to, uint256 amount) external returns (bool);
    function setKycStatus(address user, bool status) external;
    function isKycPassed(address user) external view returns (bool);
    function moduleCall(bytes32 moduleId, bytes calldata data) external returns (bytes memory);
    function snapshot() external returns (uint256);
    function getSnapshotBalance(address user, uint256 snapshotId) external view returns (uint256);
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

**GNDst-1 — стандарт для будущего мультицепочечных, модульных и безопасных токенов!**

