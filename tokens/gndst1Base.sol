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

contract GNDst1Token is IGNDst1 {
    string public name = "Ganimed Token";
    string public symbol = "GND";
    uint8 public decimals = 18;
    uint256 private _totalSupply;
    address public owner;
    address public bridge;

    // Основные структуры данных
    mapping(address => uint256) private _balances;
    mapping(address => mapping(address => uint256)) private _allowances;
    mapping(address => bool) private _kycPassed;

    // Снимки балансов и дивидендов
    uint256 public currentSnapshotId;
    mapping(uint256 => mapping(address => uint256)) private _snapshotBalances;
    mapping(uint256 => uint256) public dividendsPerShare;

    // Модульная система
    struct ModuleInfo {
        address moduleAddress;
        string name;
    }
    mapping(bytes32 => ModuleInfo) public registeredModules;

    // События
    event Transfer(address indexed from, address indexed to, uint256 value);
    event Approval(address indexed owner, address indexed spender, uint256 value);
    event CrossChainTransfer(address indexed from, string targetChain, address indexed to, uint256 value);
    event KycStatusChanged(address indexed user, bool status);
    event ModuleCall(bytes32 indexed moduleId, address indexed caller);
    event SnapshotCreated(uint256 indexed snapshotId, uint256 timestamp);
    event DividendClaimed(address indexed user, uint256 amount, uint256 snapshotId);
    event ModuleRegistered(bytes32 indexed moduleId, address indexed moduleAddress, string name);

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

    // --- Основные методы ---

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

    // --- Расширенные методы ---

    function crossChainTransfer(string calldata targetChain, address to, uint256 amount) external override onlyKyc returns (bool) {
        _transfer(msg.sender, bridge, amount);
        emit CrossChainTransfer(msg.sender, targetChain, to, amount);
        return true;
    }

    function setKycStatus(address user, bool status) external override onlyOwner {
        _kycPassed[user] = status;
        emit KycStatusChanged(user, status);
    }

    function isKycPassed(address user) external view override returns (bool) {
        return _kycPassed[user];
    }

    function snapshot() external override onlyOwner returns (uint256) {
        currentSnapshotId += 1;
        emit SnapshotCreated(currentSnapshotId, block.timestamp);
        return currentSnapshotId;
    }

    function getSnapshotBalance(address user, uint256 snapshotId) external view override returns (uint256) {
        return _snapshotBalances[snapshotId][user];
    }

    function claimDividends(uint256 snapshotId) external override {
        uint256 balance = _snapshotBalances[snapshotId][msg.sender];
        uint256 dividendAmount = balance * dividendsPerShare[snapshotId];

        require(dividendAmount > 0, "No dividends");
        _transfer(owner, msg.sender, dividendAmount);

        emit DividendClaimed(msg.sender, dividendAmount, snapshotId);
    }

    function moduleCall(bytes32 moduleId, bytes calldata data) external override returns (bytes memory) {
        require(registeredModules[moduleId].moduleAddress != address(0), "Module not registered");
        emit ModuleCall(moduleId, msg.sender);
        return bytes("module call placeholder");
    }

    function registerModule(bytes32 moduleId, address moduleAddress, string calldata name) external override onlyOwner {
        require(moduleAddress != address(0), "Invalid address");
        require(registeredModules[moduleId].moduleAddress == address(0), "Module already exists");

        registeredModules[moduleId] = ModuleInfo({
            moduleAddress: moduleAddress,
            name: name
        });

        emit ModuleRegistered(moduleId, moduleAddress, name);
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
