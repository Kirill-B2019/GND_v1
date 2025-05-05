// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

/// @title ISimpleGNDToken
/// @notice Универсальный интерфейс для токенов стандарта SimpleGNDToken (расширяемый, совместим с ERC-20/TRC-20 и кастомными функциями)

interface ISimpleGNDToken {
    // === Базовые методы (ERC-20/TRC-20 совместимость) ===

    /// @notice Получить общее количество токенов
    function totalSupply() external view returns (uint256);

    /// @notice Получить баланс токенов у пользователя
    function balanceOf(address account) external view returns (uint256);

    /// @notice Перевести токены на другой адрес
    function transfer(address to, uint256 amount) external returns (bool);

    /// @notice Получить разрешение (allowance) на перевод токенов от owner к spender
    function allowance(address owner, address spender) external view returns (uint256);

    /// @notice Утвердить разрешение на перевод токенов
    function approve(address spender, uint256 amount) external returns (bool);

    /// @notice Перевести токены с чужого адреса (по разрешению)
    function transferFrom(address from, address to, uint256 amount) external returns (bool);

    // === События (Events) ===

    event Transfer(address indexed from, address indexed to, uint256 value);
    event Approval(address indexed owner, address indexed spender, uint256 value);

    // === Расширенные методы для кастомного стандарта ===

    /// @notice Минтинг (выпуск новых токенов)
    /// @dev Обычно доступен только владельцу или через роль MINTER
    function mint(address to, uint256 amount) external;

    /// @notice Сжигание токенов (burn) с собственного баланса
    function burn(uint256 amount) external;

    /// @notice Сжигание токенов с чужого баланса (по разрешению)
    function burnFrom(address from, uint256 amount) external;

    /// @notice Получить метаданные токена (название, символ, decimals, totalSupply)
    function getMetadata() external view returns (
        string memory name,
        string memory symbol,
        uint8 decimals,
        uint256 totalSupply
    );

    // === Дополнительные методы для кастомных сценариев ===

    /// @notice Массовая раздача токенов (airdrop)
    function airdrop(address[] calldata recipients, uint256 amount) external;

    /// @notice Пример интеграции с оракулом (запись внешних данных)
    function submitOracleData(string calldata data) external;

    event OracleData(address indexed from, string data);
}
