// SPDX-License-Identifier: MIT
pragma solidity ^0.8.16;

import "./IGNDst1.sol";

/// @title GANIToken — governance-токен GANI (Ganimed Governance) по стандарту GND-st1
/// @notice Деплой: шаг 3. Конструктор: controllerContract (адрес из шага 1). Минт только с контроллера (mintGANI).
/// @dev Реализует IGNDst1. Всего объём 100M; первая циркулирующая эмиссия 20M; следующие эмиссии — через контроллер (логика дорабатывается).
/// @dev Invariants: _totalSupply <= TOTAL_SUPPLY; mint только onlyController; onlyKyc для переводов. См. INVARIANTS.md.
contract GANIToken is IGNDst1 {
    /// @notice Максимальное предложение: 100M при 6 decimals
    uint256 public constant TOTAL_SUPPLY = 100_000_000 * 10**6;
    /// @notice Первая циркулирующая эмиссия: 20M при 6 decimals (выпускается в конструкторе на контроллер)
    uint256 public constant FIRST_EMISSION = 20_000_000 * 10**6;

    string public name = "GANI (Ganimed Governance)";
    string public symbol = "GANI";
    uint8 public decimals = 6;

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

    event ModuleCall(bytes32 indexed moduleId, address indexed caller);

    error OnlyController();
    error ExceedsTotalSupply();
    error ZeroAddress();

    modifier onlyController() {
        require(msg.sender == controller, "Only controller");
        require(_isContract(msg.sender), "Controller must be a contract");
        _;
    }

    modifier onlyKyc() {
        require(_kycPassed[msg.sender], "KYC required");
        _;
    }

    /// @param controllerContract Адрес контракта из шага 1 (NativeTokensController)
    /// @dev В первую эмиссию выпускается FIRST_EMISSION (20M) на контроллер. Доп. эмиссии — через mint() с контроллера (заглушка, логика дорабатывается).
    constructor(address controllerContract) {
        require(controllerContract != address(0), "Zero controller");
        require(_isContract(controllerContract), "Controller must be a contract");
        controller = controllerContract;
        bridge = address(0);
        _totalSupply = FIRST_EMISSION;
        _balances[controllerContract] = FIRST_EMISSION;
        emit Transfer(address(0), controllerContract, FIRST_EMISSION);
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
        require(to != address(0), "Transfer to zero");
        require(_allowances[from][msg.sender] >= amount, "Allowance exceeded");
        _allowances[from][msg.sender] -= amount;
        _transfer(from, to, amount);
        return true;
    }

    function crossChainTransfer(string calldata targetChain, address to, uint256 amount) external override onlyKyc returns (bool) {
        require(bridge != address(0), "Bridge not set");
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

    function setSnapshotBalance(uint256 snapshotId, address user, uint256 amount) external onlyController {
        _snapshotBalances[snapshotId][user] = amount;
    }

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

    function registerModule(bytes32 moduleId, address moduleAddress, string calldata name_) external override onlyController {
        require(moduleAddress != address(0), "Invalid address");
        require(registeredModules[moduleId].moduleAddress == address(0), "Module already exists");
        registeredModules[moduleId] = ModuleInfo({
            moduleAddress: moduleAddress,
            name: name_
        });
        emit ModuleRegistered(moduleId, moduleAddress, name_);
    }

    /// @notice Минт только с адреса контроллера (вызов из NativeTokensController.mintGANI). Следующие эмиссии после первой — управляются контрактом (логика дорабатывается).
    function mint(address to, uint256 amount) external onlyController {
        if (to == address(0)) revert ZeroAddress();
        if (_totalSupply + amount > TOTAL_SUPPLY) revert ExceedsTotalSupply();
        _totalSupply += amount;
        _balances[to] += amount;
        emit Transfer(address(0), to, amount);
    }

    function _transfer(address from, address to, uint256 amount) internal {
        require(_balances[from] >= amount, "Insufficient balance");
        _balances[from] -= amount;
        _balances[to] += amount;
        emit Transfer(from, to, amount);
    }
}
