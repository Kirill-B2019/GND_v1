# Токены в ГАНИМЕД

## Обзор

ГАНИМЕД поддерживает **нативные монеты L1** (GND и GANI) и собственный стандарт токенов GNDst-1, а также стандартные токены ERC20, ERC721 и ERC1155. GNDst-1 представляет собой расширенный стандарт, совместимый с ERC-20 и TRC-20, с дополнительными возможностями для управления токенами и интеграции с платформой.

## Нативные монеты (GND, GANI)

Поддерживаются два режима:

1. **Нативная модель L1 (по умолчанию, если не заданы адреса контрактов):** балансы GND и GANI хранятся в таблице **native_balances**, загружаются и сохраняются нодой при старте и после применения блоков.
2. **Режим «всё на контрактах»:** при указании в `config/native_contracts.json` адресов задеплоенных контрактов GND и GANI нода использует **token_balances** (по token_id, привязанному к этим контрактам). Таблица `native_balances` для GND/GANI не используется. Деплой и переход описаны в [deployment-contracts-variant-c.md](deployment-contracts-variant-c.md).

Общие параметры:

- **Decimals:** GND — 18 знаков после запятой; GANI — 6 знаков после запятой. Значения задаются в `config/coins.json` и при наличии записи в БД обновляются при каждом запуске ноды.
- **Газ:** оплачивается только в GND (в режиме контрактов — списание по token_balances для токена GND).
- **Переводы:** транзакция перевода указывает `symbol` (GND или GANI); проверка баланса и списание/начисление выполняются по этому символу (в режиме контрактов — через token_balances).
- **Циркулирующее предложение:** в нативной модели — жёсткий лимит в `tokens.circulating_supply`; в режиме контрактов правила эмиссии заданы в Solidity (GNDToken.sol, GANIToken.sol).
- **Защита и сохранность:** см. [roles-and-operations.md](roles-and-operations.md) и [implementation-plan-native-coins.md](implementation-plan-native-coins.md).

### Контракты GND/GANI (режим «всё на контрактах»)

| Файл | Назначение |
| :-- | :-- |
| **tokens/standards/native/GNDToken.sol** | Деплоируемый ERC-20/GNDst-1 контракт GND: начальная эмиссия в конструкторе на treasury, минтинг отключён. |
| **tokens/standards/native/GANIToken.sol** | Деплоируемый ERC-20 контракт GANI: фиксированная эмиссия 100M (6 decimals) в конструкторе, минтинг отключён навсегда. |

### Интерфейсы и Base-обёртки (Solidity)

Для совместимости с контрактами распределения (vesting, пулы, казначейство) используются интерфейсы и обёртки:

| Файл | Назначение |
| :-- | :-- |
| **tokens/standards/native/INativeCoin.sol** | Базовый интерфейс (ERC-20 ядро). |
| **tokens/standards/native/IGND.sol**, **IGANI.sol** | Интерфейсы GND/GANI. |
| **tokens/standards/native/GNDCoinBase.sol**, **GANICoinBase.sol** | Обёртки для precompile/L1 (при нативной модели); при режиме контрактов источник истины — задеплоенные GNDToken/GANIToken. |

Все прочие распределения (пулы валидаторов, экосистемный фонд, vesting, DAO, DEX) регулируются отдельными смарт-контрактами.

## Создание токена через API (внешние системы)

Создание и регистрация токена выполняется запросом **POST /api/v1/token/deploy** к REST API ноды. Обязателен заголовок **X-API-Key** (проверка по константе или таблице `api_keys` в БД). В теле запроса передаются: `name`, `symbol`, `decimals`, `total_supply`, `owner`, опционально `standard` (по умолчанию GND-st1). После успешного деплоя токен регистрируется в in-memory реестре и в БД (таблицы `contracts`, `tokens`), владельцу назначается начальный баланс. Подробное описание логики, кодов ответов и примеры — **[api-token-deploy.md](api-token-deploy.md)**. Примеры curl — в [api-requests.md](api-requests.md).

## GNDst-1

### Основные характеристики
- Совместимость с ERC-20 и TRC-20
- Кросс-чейн трансферы
- KYC интеграция
- Модульная система расширений
- Снимки балансов и дивиденды
- Управление правами доступа

### Интерфейс
```solidity
interface IGNDst1 {
    // Базовые методы ERC-20/TRC-20
    function totalSupply() external view returns (uint256);
    function balanceOf(address account) external view returns (uint256);
    function transfer(address to, uint256 amount) external returns (bool);
    function allowance(address owner, address spender) external view returns (uint256);
    function approve(address spender, uint256 amount) external returns (bool);
    function transferFrom(address from, address to, uint256 amount) external returns (bool);

    // Расширенные методы GNDst-1
    function crossChainTransfer(string calldata targetChain, address to, uint256 amount) external returns (bool);
    function setKycStatus(address user, bool status) external;
    function isKycPassed(address user) external view returns (bool);
    function moduleCall(bytes32 moduleId, bytes calldata data) external returns (bytes memory);
    function snapshot() external returns (uint256);
    function getSnapshotBalance(address user, uint256 snapshotId) external view returns (uint256);
    function claimDividends(uint256 snapshotId) external;
    function registerModule(bytes32 moduleId, address moduleAddress, string calldata name) external;
}
```

### События
```solidity
event Transfer(address indexed from, address indexed to, uint256 value);
event Approval(address indexed owner, address indexed spender, uint256 value);
event CrossChainTransfer(address indexed from, string targetChain, address indexed to, uint256 value);
event KycStatusChanged(address indexed user, bool status);
event ModuleCall(bytes32 indexed moduleId, address indexed caller);
event SnapshotCreated(uint256 indexed snapshotId, uint256 timestamp);
event DividendClaimed(address indexed user, uint256 amount, uint256 snapshotId);
event ModuleRegistered(bytes32 indexed moduleId, address indexed moduleAddress, string name);
```

### Особенности

#### Кросс-чейн трансферы
- Поддержка трансферов между разными блокчейнами
- Интеграция с мостами
- Безопасная передача токенов

#### KYC интеграция
- Управление статусом KYC пользователей
- Ограничение операций для неподтвержденных адресов
- Интеграция с внешними KYC провайдерами

#### Модульная система
- Регистрация новых модулей
- Расширение функциональности токена
- Изолированное выполнение модулей

#### Снимки и дивиденды
- Создание снимков балансов
- Распределение дивидендов
- Отслеживание истории балансов

## Стандартные токены

### ERC20
- Базовый стандарт для взаимозаменяемых токенов
- Поддержка всех стандартных методов
- Совместимость с существующими инструментами

### ERC721
- Стандарт для невзаимозаменяемых токенов (NFT)
- Уникальные идентификаторы
- Метаданные и URI

### ERC1155
- Мульти-токен стандарт
- Эффективные батч-операции
- Гибридные токены

## Интеграция

### Создание токена
```solidity
// Создание GNDst-1 токена
contract MyToken is GNDst1Token {
    constructor(uint256 initialSupply, address bridgeAddress) 
        GNDst1Token(initialSupply, bridgeAddress) {
    }
}
```

### Использование API
```javascript
// Создание токена
const token = await api.createToken({
    name: "MyToken",
    symbol: "MTK",
    decimals: 18,
    totalSupply: "1000000000000000000000000",
    standard: "GND-st1"
});

// Трансфер токенов
await token.transfer(recipient, amount);

// Кросс-чейн трансфер
await token.crossChainTransfer("ethereum", recipient, amount);
```

## Безопасность

### Рекомендации
- Использование проверенных библиотек
- Аудит кода
- Тестирование
- Мониторинг газа
- Безопасные паттерны

### Аудит
- Проверка кода
- Тестирование безопасности
- Анализ уязвимостей
- Рекомендации по улучшению

## Мониторинг

### Метрики
- Количество транзакций
- Объем трансферов
- Активность токенов
- Использование газа
- События

### Алерты
- Аномальная активность
- Большие трансферы
- Ошибки контрактов
- Проблемы с газом

## Обновления

### Версионирование
- Семантическое версионирование
- Обратная совместимость
- Миграции

### Процесс
1. Планирование
2. Тестирование
3. Аудит
4. Развертывание
5. Мониторинг

## Документация

### Стандарты
- [GNDst-1](GNDst-1.md)
- [ERC20](https://eips.ethereum.org/EIPS/eip-20)
- [ERC721](https://eips.ethereum.org/EIPS/eip-721)
- [ERC1155](https://eips.ethereum.org/EIPS/eip-1155)

### Примеры
- [Примеры контрактов](../tokens/standards/gndst1/)
- [Тесты](../tokens/standards/gndst1/gndst1_test.go)
- [Интеграция](../integration/)

<div style="text-align: center">| KB @CerberRus00 - Nexus Invest Team 2026</div>