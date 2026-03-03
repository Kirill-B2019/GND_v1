// SPDX-License-Identifier: MIT
pragma solidity ^0.8.16;

import "./IGNDst1.sol";

/// @title GNDst-1: Мультистандартный токен для блокчейна ГАНИМЕД
/// @notice Совместим с ERC-20, TRC-20 и расширен новыми функциями.
/// @dev Управление только через контракт-контроллер; прямое управление с EOA заблокировано.

contract GNDst1Token is IGNDst1 {
    /// @notice Максимальное/целевое предложение: 1e27 (1 млрд GND с 18 decimals)
    uint256 public constant TOTAL_SUPPLY = 1000000000000000000000000000;

    string public name = "Ganimed";
    string public symbol = "GND";
    uint8 public decimals = 18;
    uint256 private _totalSupply;
    address public immutable controller;
    address public bridge;

    mapping(address => uint256) private _balances;
    mapping(address => mapping(address => uint256)) private _allowances;
    mapping(address => bool) private _kycPassed;

    uint256 public currentSnapshotId;
    mapping(uint256 => mapping(address => uint256)) private _snapshotBalances;
    mapping(uint256 => uint256) public dividendsPerShare;

    struct ModuleInfo {
        address moduleAddress;
        string name;
    }
    mapping(bytes32 => ModuleInfo) public registeredModules;

    event Transfer(address indexed from, address indexed to, uint256 value);
    event Approval(address indexed owner, address indexed spender, uint256 value);
    event CrossChainTransfer(address indexed from, string targetChain, address indexed to, uint256 value);
    event KycStatusChanged(address indexed user, bool status);
    event ModuleCall(bytes32 indexed moduleId, address indexed caller);
    event SnapshotCreated(uint256 indexed snapshotId, uint256 timestamp);
    event DividendClaimed(address indexed user, uint256 amount, uint256 snapshotId);
    event ModuleRegistered(bytes32 indexed moduleId, address indexed moduleAddress, string name);

    error MintingDisabled();

    modifier onlyController() {
        require(msg.sender == controller, "Only controller");
        require(_isContract(msg.sender), "Controller must be a contract");
        _;
    }

    modifier onlyKyc() {
        require(_kycPassed[msg.sender], "KYC required");
        _;
    }

    constructor(
        uint256 initialSupply,
        address bridgeAddress,
        address controllerContract
    ) {
        require(controllerContract != address(0), "Zero controller");
        require(_isContract(controllerContract), "Controller must be a contract");
        require(initialSupply <= TOTAL_SUPPLY && initialSupply > 0, "Invalid initial supply");
        controller = controllerContract;
        bridge = bridgeAddress;
        _mint(controllerContract, initialSupply);
    }

    function _isContract(address account) private view returns (bool) {
        uint256 size;
        assembly {
            size := extcodesize(account)
        }
        return size > 0;
    }

    function totalSupply() public view override returns (uint256) {
        return _totalSupply;
    }

    function balanceOf(address account) public view override returns (uint256) {
        return _balances[account];
    }

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

    function crossChainTransfer(string calldata targetChain, address to, uint256 amount) external override onlyKyc returns (bool) {
        _transfer(msg.sender, bridge, amount);
        emit CrossChainTransfer(msg.sender, targetChain, to, amount);
        return true;
    }

    function setKycStatus(address user, bool status) external override onlyController {
        _kycPassed[user] = status;
        emit KycStatusChanged(user, status);
    }

    function isKycPassed(address user) external view override returns (bool) {
        return _kycPassed[user];
    }

    function snapshot() external override onlyController returns (uint256) {
        currentSnapshotId += 1;
        uint256 id = currentSnapshotId;
        dividendsPerShare[id] = 0;
        emit SnapshotCreated(id, block.timestamp);
        return id;
    }

    function getSnapshotBalance(address user, uint256 snapshotId) external view override returns (uint256) {
        return _snapshotBalances[snapshotId][user];
    }

    /// @notice Фиксирует баланс пользователя в снимке (вызывается контроллером после snapshot).
    function setSnapshotBalance(uint256 snapshotId, address user, uint256 amount) external onlyController {
        _snapshotBalances[snapshotId][user] = amount;
    }

    /// @notice Устанавливает дивиденды на снимок (вызывается контроллером).
    function setDividendsPerShare(uint256 snapshotId, uint256 amount) external onlyController {
        dividendsPerShare[snapshotId] = amount;
    }

    function claimDividends(uint256 snapshotId) external override {
        uint256 balance = _snapshotBalances[snapshotId][msg.sender];
        uint256 dividendAmount = balance * dividendsPerShare[snapshotId];
        require(dividendAmount > 0, "No dividends");
        _transfer(controller, msg.sender, dividendAmount);
        emit DividendClaimed(msg.sender, dividendAmount, snapshotId);
    }

    function moduleCall(bytes32 moduleId, bytes calldata data) external override returns (bytes memory) {
        require(registeredModules[moduleId].moduleAddress != address(0), "Module not registered");
        emit ModuleCall(moduleId, msg.sender);
        return bytes("module call placeholder");
    }

    function registerModule(bytes32 moduleId, address moduleAddress, string calldata name) external override onlyController {
        require(moduleAddress != address(0), "Invalid address");
        require(registeredModules[moduleId].moduleAddress == address(0), "Module already exists");
        registeredModules[moduleId] = ModuleInfo({
            moduleAddress: moduleAddress,
            name: name
        });
        emit ModuleRegistered(moduleId, moduleAddress, name);
    }

    function _transfer(address from, address to, uint256 amount) internal {
        require(_balances[from] >= amount, "Insufficient balance");
        _balances[from] -= amount;
        _balances[to] += amount;
        emit Transfer(from, to, amount);
    }

    /// @notice Минтинг отключён: эмиссия только в конструкторе.
    function mint(address /* to */, uint256 /* amount */) external pure {
        revert MintingDisabled();
    }

    function _mint(address to, uint256 amount) internal {
        _totalSupply += amount;
        _balances[to] += amount;
        emit Transfer(address(0), to, amount);
    }
}
