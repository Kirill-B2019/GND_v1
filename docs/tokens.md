# Токены в ГАНИМЕД

## Обзор

ГАНИМЕД поддерживает собственный стандарт токенов GNDst-1, а также стандартные токены ERC20, ERC721 и ERC1155. GNDst-1 представляет собой расширенный стандарт, совместимый с ERC-20 и TRC-20, с дополнительными возможностями для управления токенами и интеграции с платформой.

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
    standard: "GNDst-1"
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