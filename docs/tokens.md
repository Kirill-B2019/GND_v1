# Токены в ГАНИМЕД

## Обзор

ГАНИМЕД поддерживает **нативные монеты L1** (GND и GANI) и собственный стандарт токенов GNDst-1, а также стандартные токены ERC20, ERC721 и ERC1155. GNDst-1 представляет собой расширенный стандарт, совместимый с ERC-20 и TRC-20, с дополнительными возможностями для управления токенами и интеграции с платформой.

## Нативные монеты (GND, GANI)

GND и GANI — нативные активы протокола L1, а не контрактные токены. Их балансы хранятся в таблице **native_balances**, загружаются и сохраняются нодой при старте и после применения блоков.

- **Источник истины:** состояние ноды; изменение балансов только через L1 (транзакции перевода, списание газа в GND, начисление при первом запуске).
- **Газ:** оплачивается только в GND.
- **Переводы:** транзакция перевода указывает `symbol` (GND или GANI); проверка баланса и списание/начисление выполняются по этому символу.
- **Циркулирующее предложение (жёсткий лимит):** в таблице `tokens` хранится `circulating_supply` — максимально допустимая сумма балансов по данной нативной монете. При любом начислении (AddBalance/Credit), в т.ч. при первом запуске и будущей эмиссии, нода проверяет: сумма всех балансов по символу + сумма начисления не должна превышать лимит; иначе операция отклоняется. Начальная эмиссия при FirstLaunch использует объём из `config/coins.json` → `circulating_supply` (не total_supply).
- **Защита и сохранность:** см. [roles-and-operations.md](roles-and-operations.md) и [implementation-plan-native-coins.md](implementation-plan-native-coins.md).

### Интерфейсы и Base-контракты (Solidity)

Нативные монеты GND и GANI представлены в Solidity интерфейсами и базовыми контрактами для совместимости с контрактами распределения (vesting, пулы, казначейство):

| Файл | Назначение |
| :-- | :-- |
| **tokens/standards/native/INativeCoin.sol** | Базовый интерфейс нативной монеты (ERC-20 ядро: totalSupply, balanceOf, transfer, allowance, approve, transferFrom). |
| **tokens/standards/native/IGND.sol** | Интерфейс GND (Ganymede Coin): INativeCoin + name, symbol, decimals. |
| **tokens/standards/native/IGANI.sol** | Интерфейс GANI (Ganymede Governance): INativeCoin + name, symbol, decimals. |
| **tokens/standards/native/GNDCoinBase.sol** | Базовая обёртка GND; вызовы рассчитаны на обработку precompile ноды (источник истины — L1 state). |
| **tokens/standards/native/GANICoinBase.sol** | Базовая обёртка GANI; вызовы рассчитаны на обработку precompile ноды. |

Все прочие распределения (пулы валидаторов, экосистемный фонд, vesting, DAO, DEX) должны регулироваться отдельными смарт-контрактами, использующими эти интерфейсы для работы с GND и GANI.

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

<div style="text-align: center">| KB @CerbeRus - Nexus Invest Team 2026</div>