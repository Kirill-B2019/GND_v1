// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/// @title GNDst-1: Мультистандартный токен для блокчейна ГАНИМЕД
/// @notice Совместим с ERC-20, TRC-20 и расширен новыми функциями

interface IGNDst1 {
    // Базовые методы ERC-20/TRC-20
    function totalSupply() external view returns (uint256);
    function balanceOf(address account) external view returns (uint256);
    function transfer(address to, uint256 amount) external returns (bool);
    function allowance(address owner, address spender) external view returns (uint256);
    function approve(address spender, uint256 amount) external returns (bool);
    function transferFrom(address from, address to, uint256 amount) external returns (bool);

    // Новые функции GNDst-1
    function crossChainTransfer(string calldata targetChain, address to, uint256 amount) external returns (bool);
    function setKycStatus(address user, bool status) external;
    function isKycPassed(address user) external view returns (bool);
    function moduleCall(bytes32 moduleId, bytes calldata data) external returns (bytes memory);
    function snapshot() external returns (uint256);
    function getSnapshotBalance(address user, uint256 snapshotId) external view returns (uint256);
}

contract GNDst1Token is IGNDst1 {
    string public name = "Ganimed Token";
    string public symbol = "GND";
    uint8 public decimals = 18;
    uint256 private _totalSupply;
    address public owner;
    address public bridge; // адрес контракта-моста

    mapping(address => uint256) private _balances;
    mapping(address => mapping(address => uint256)) private _allowances;
    mapping(address => bool) private _kycPassed;

    // Снимки балансов (snapshot)
    uint256 public currentSnapshotId;
    mapping(uint256 => mapping(address => uint256)) private _snapshotBalances;

    // События
    event Transfer(address indexed from, address indexed to, uint256 value);
    event Approval(address indexed owner, address indexed spender, uint256 value);
    event CrossChainTransfer(address indexed from, string targetChain, address indexed to, uint256 value);
    event KycStatusChanged(address indexed user, bool status);
    event ModuleCall(bytes32 indexed moduleId, address indexed caller);

    modifier onlyOwner() {
        require(msg.sender == owner, "Not owner");
        _;
    }

    modifier onlyKyc() {
        require(_kycPassed[msg.sender], "KYC required");
        _;
    }

    constructor(uint256 initialSupply, address bridgeAddress) {
        owner = msg.sender;
        bridge = bridgeAddress;
        _mint(owner, initialSupply);
    }

    // --- ERC-20/TRC-20 стандарт ---
    function totalSupply() public view override returns (uint256) { return _totalSupply; }
    function balanceOf(address account) public view override returns (uint256) { return _balances[account]; }

    function transfer(address to, uint256 amount) public override onlyKyc returns (bool) {
        _transfer(msg.sender, to, amount);
        return true;
    }

    function allowance(address owner_, address spender) public view override returns (uint256) {
        return _allowances[owner_][spender];
    }

    function approve(address spender, uint256 amount) public override returns (bool) {
        _allowances[msg.sender][spender] = amount;
        emit Approval(msg.sender, spender, amount);
        return true;
    }

    function transferFrom(address from, address to, uint256 amount) public override onlyKyc returns (bool) {
        require(_allowances[from][msg.sender] >= amount, "Allowance exceeded");
        _allowances[from][msg.sender] -= amount;
        _transfer(from, to, amount);
        return true;
    }

    // --- Новые функции GNDst-1 ---

    /// @notice Кроссчейн-перевод через мост
    function crossChainTransfer(string calldata targetChain, address to, uint256 amount) external override onlyKyc returns (bool) {
        _transfer(msg.sender, bridge, amount);
        emit CrossChainTransfer(msg.sender, targetChain, to, amount);
        // Вызов модуля моста (bridge) для дальнейшей обработки
        return true;
    }

    /// @notice Установка статуса KYC
    function setKycStatus(address user, bool status) external override onlyOwner {
        _kycPassed[user] = status;
        emit KycStatusChanged(user, status);
    }

    /// @notice Проверка KYC
    function isKycPassed(address user) external view override returns (bool) {
        return _kycPassed[user];
    }

    /// @notice Вызов внешнего модуля (расширяемость)
    function moduleCall(bytes32 moduleId, bytes calldata data) external override returns (bytes memory) {
        // Пример: вызов внешнего контракта по moduleId (реестр модулей вне этого контракта)
        emit ModuleCall(moduleId, msg.sender);
        // Здесь должна быть интеграция с модульной системой ядра блокчейна
        return bytes("module call placeholder");
    }

    /// @notice Снимок балансов (snapshot)
    function snapshot() external override onlyOwner returns (uint256) {
        currentSnapshotId += 1;
        for (uint i = 0; i < 10; i++) { // пример: только для первых 10 адресов, для реального использования нужен off-chain индексатор
            // _snapshotBalances[currentSnapshotId][address(i)] = _balances[address(i)];
        }
        return currentSnapshotId;
    }

    function getSnapshotBalance(address user, uint256 snapshotId) external view override returns (uint256) {
        return _snapshotBalances[snapshotId][user];
    }

    // --- Внутренние функции ---

    function _transfer(address from, address to, uint256 amount) internal {
        require(_balances[from] >= amount, "Insufficient balance");
        _balances[from] -= amount;
        _balances[to] += amount;
        emit Transfer(from, to, amount);
    }

    function _mint(address to, uint256 amount) internal {
        _totalSupply += amount;
        _balances[to] += amount;
        emit Transfer(address(0), to, amount);
    }
}
