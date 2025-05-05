// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

/// @title SimpleGNDToken - Пример кастомного токена для блокчейна ГАНИМЕД
/// @notice Контракт реализует стандартные функции токена (ERC-20/TRC-20) и расширяемый интерфейс для кастомных операций

contract SimpleGNDToken {
    // Название токена
    string public name = "Simple GND Token";
    // Символ токена
    string public symbol = "GND";
    // Количество знаков после запятой (десятичные)
    uint8 public decimals = 18;
    // Общее количество токенов
    uint256 public totalSupply;

    // Балансы пользователей
    mapping(address => uint256) public balanceOf;
    // Разрешения (allowance) на перевод токенов от имени пользователя
    mapping(address => mapping(address => uint256)) public allowance;

    // Событие перевода токенов
    event Transfer(address indexed from, address indexed to, uint256 value);
    // Событие утверждения разрешения
    event Approval(address indexed owner, address indexed spender, uint256 value);

    /// @notice Конструктор, создает токены и назначает их владельцу
    /// @param initialSupply Начальное количество токенов (в минимальных единицах, с учетом decimals)
    constructor(uint256 initialSupply) {
        totalSupply = initialSupply;
        balanceOf[msg.sender] = initialSupply;
        emit Transfer(address(0), msg.sender, initialSupply);
    }

    /// @notice Перевод токенов другому адресу
    /// @param to Адрес получателя
    /// @param value Количество токенов для перевода
    /// @return success Успешность операции
    function transfer(address to, uint256 value) public returns (bool success) {
        require(balanceOf[msg.sender] >= value, "Недостаточно токенов");
        require(to != address(0), "Некорректный адрес получателя");

        balanceOf[msg.sender] -= value;
        balanceOf[to] += value;

        emit Transfer(msg.sender, to, value);
        return true;
    }

    /// @notice Утверждение разрешения на перевод токенов другим адресом
    /// @param spender Адрес, которому разрешено тратить токены
    /// @param value Количество токенов для разрешения
    /// @return success Успешность операции
    function approve(address spender, uint256 value) public returns (bool success) {
        require(spender != address(0), "Некорректный адрес");

        allowance[msg.sender][spender] = value;
        emit Approval(msg.sender, spender, value);
        return true;
    }

    /// @notice Перевод токенов от имени другого пользователя (требует разрешения)
    /// @param from Адрес отправителя
    /// @param to Адрес получателя
    /// @param value Количество токенов для перевода
    /// @return success Успешность операции
    function transferFrom(address from, address to, uint256 value) public returns (bool success) {
        require(balanceOf[from] >= value, "Недостаточно токенов у отправителя");
        require(allowance[from][msg.sender] >= value, "Недостаточно разрешения");
        require(to != address(0), "Некорректный адрес получателя");

        balanceOf[from] -= value;
        balanceOf[to] += value;
        allowance[from][msg.sender] -= value;

        emit Transfer(from, to, value);
        return true;
    }

    /// @notice Получить разрешение (allowance) для spender от owner
    /// @param owner Владелец токенов
    /// @param spender Уполномоченный адрес
    /// @return Оставшееся количество разрешённых токенов
    function allowanceOf(address owner, address spender) public view returns (uint256) {
        return allowance[owner][spender];
    }

    // === Дополнительные функции для кастомного стандарта ===

    /// @notice Минтинг (создание новых токенов) - только для владельца
    /// @param to Кому начислить новые токены
    /// @param value Сколько токенов создать
    function mint(address to, uint256 value) public /* onlyOwner */ {
        // В реальной реализации добавить модификатор onlyOwner!
        require(to != address(0), "Некорректный адрес");
        totalSupply += value;
        balanceOf[to] += value;
        emit Transfer(address(0), to, value);
    }

    /// @notice Сжигание токенов (burn) - уменьшение totalSupply
    /// @param value Сколько токенов сжечь у отправителя
    function burn(uint256 value) public {
        require(balanceOf[msg.sender] >= value, "Недостаточно токенов для сжигания");
        balanceOf[msg.sender] -= value;
        totalSupply -= value;
        emit Transfer(msg.sender, address(0), value);
    }

    /// @notice Сжигание токенов с чужого адреса (если есть разрешение)
    /// @param from С какого адреса сжечь
    /// @param value Сколько токенов сжечь
    function burnFrom(address from, uint256 value) public {
        require(balanceOf[from] >= value, "Недостаточно токенов для сжигания");
        require(allowance[from][msg.sender] >= value, "Недостаточно разрешения");
        balanceOf[from] -= value;
        allowance[from][msg.sender] -= value;
        totalSupply -= value;
        emit Transfer(from, address(0), value);
    }

    /// @notice Получить метаданные токена (название, символ, decimals, totalSupply)
    function getMetadata() public view returns (string memory, string memory, uint8, uint256) {
        return (name, symbol, decimals, totalSupply);
    }

    // === Вспомогательные функции, которые можно добавить для кастомных стандартов ===

    /// @notice Массовая раздача (airdrop) токенов
    /// @param recipients Массив адресов получателей
    /// @param amount Количество токенов каждому
    function airdrop(address[] memory recipients, uint256 amount) public /* onlyOwner */ {
        // В реальной реализации добавить модификатор onlyOwner!
        for (uint256 i = 0; i < recipients.length; i++) {
            mint(recipients[i], amount);
        }
    }

    /// @notice Пример интеграции с оракулом (запись внешних данных)
    /// @param data Внешние данные (например, цена, курс)
    event OracleData(address indexed from, string data);

    function submitOracleData(string memory data) public {
        // В реальной реализации добавить проверки и права доступа
        emit OracleData(msg.sender, data);
    }
}
