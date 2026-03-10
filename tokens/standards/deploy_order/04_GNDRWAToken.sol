// SPDX-License-Identifier: MIT
pragma solidity ^0.8.16;

import "./IGNDRWA.sol";

// Копия tokens/standards/gndrwa/GND-RWA.sol для автономной компиляции deploy_order (стандарт GND-st1 + RWA).
// Деплой: после 01–03; параметры: controller, bridgeAddress, name, symbol, decimals, maxSupply.

/// @title GND-RWA: токен реальных активов (стандарт GND-st1 + RWA)
/// @notice Управляется контроллером; пауза, заморозка, cap; KYC, snapshot/дивиденды, модули, crossChain.
/// @dev Invariants: _totalSupply <= _maxSupply при _maxSupply > 0; паузы/заморозка в переводах; только onlyController для админ-функций. См. INVARIANTS.md.
contract GNDRWAToken is IGNDRWA {
    string public name;
    string public symbol;
    uint8 public decimals;

    uint256 private _totalSupply;
    uint256 private _maxSupply;
    address public immutable controller;
    address public bridge;

    mapping(address => uint256) private _balances;
    mapping(address => mapping(address => uint256)) private _allowances;
    mapping(address => bool) private _frozen;
    mapping(address => bool) private _kycPassed;

    bool private _transfersPaused;
    bool private _mintPaused;
    bool private _burnPaused;

    uint256 public currentSnapshotId;
    mapping(uint256 => mapping(address => uint256)) private _snapshotBalances;
    mapping(uint256 => uint256) public dividendsPerShare;

    struct ModuleInfo {
        address moduleAddress;
        string name;
    }
    mapping(bytes32 => ModuleInfo) public registeredModules;

    // События Transfer, Approval, Mint, Burn, TransfersPausedChanged, MintPausedChanged, BurnPausedChanged,
    // FrozenChanged, KycStatusChanged, SnapshotCreated, DividendClaimed, ModuleRegistered — в IGNDst1/IGNDRWA
    event ModuleCall(bytes32 indexed moduleId, address indexed caller);

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
        address controllerContract,
        address bridgeAddress,
        string memory name_,
        string memory symbol_,
        uint8 decimals_,
        uint256 maxSupply_
    ) {
        require(controllerContract != address(0), "Zero controller");
        require(_isContract(controllerContract), "Controller must be a contract");
        controller = controllerContract;
        bridge = bridgeAddress;
        name = name_;
        symbol = symbol_;
        decimals = decimals_;
        _maxSupply = maxSupply_;
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

    function maxSupply() external view override returns (uint256) {
        return _maxSupply;
    }

    function transfersPaused() external view override returns (bool) {
        return _transfersPaused;
    }

    function mintPaused() external view override returns (bool) {
        return _mintPaused;
    }

    function burnPaused() external view override returns (bool) {
        return _burnPaused;
    }

    function isFrozen(address account) external view override returns (bool) {
        return _frozen[account];
    }

    function isKycPassed(address user) external view override returns (bool) {
        return _kycPassed[user];
    }

    function setTransfersPaused(bool paused) external override onlyController {
        if (_transfersPaused != paused) {
            _transfersPaused = paused;
            emit TransfersPausedChanged(paused);
        }
    }

    function setMintPaused(bool paused) external override onlyController {
        if (_mintPaused != paused) {
            _mintPaused = paused;
            emit MintPausedChanged(paused);
        }
    }

    function setBurnPaused(bool paused) external override onlyController {
        if (_burnPaused != paused) {
            _burnPaused = paused;
            emit BurnPausedChanged(paused);
        }
    }

    function setFrozen(address account, bool frozen) external override onlyController {
        require(account != address(0), "Zero address");
        if (_frozen[account] != frozen) {
            _frozen[account] = frozen;
            emit FrozenChanged(account, frozen);
        }
    }

    function setKycStatus(address user, bool status) external override onlyController {
        _kycPassed[user] = status;
        emit KycStatusChanged(user, status);
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

    function registerModule(bytes32 moduleId, address moduleAddress, string calldata moduleName) external override onlyController {
        require(moduleAddress != address(0), "Invalid address");
        require(registeredModules[moduleId].moduleAddress == address(0), "Module already exists");
        registeredModules[moduleId] = ModuleInfo({
            moduleAddress: moduleAddress,
            name: moduleName
        });
        emit ModuleRegistered(moduleId, moduleAddress, moduleName);
    }

    function moduleCall(bytes32 moduleId, bytes calldata data) external override returns (bytes memory) {
        require(registeredModules[moduleId].moduleAddress != address(0), "Module not registered");
        emit ModuleCall(moduleId, msg.sender);
        return data;
    }

    function transfer(address to, uint256 amount) public override onlyKyc returns (bool) {
        _requireTransferAllowed(msg.sender, to);
        _transfer(msg.sender, to, amount);
        return true;
    }

    function allowance(address owner, address spender) public view override returns (uint256) {
        return _allowances[owner][spender];
    }

    function approve(address spender, uint256 amount) public override returns (bool) {
        require(spender != address(0), "Approve to zero");
        _allowances[msg.sender][spender] = amount;
        emit Approval(msg.sender, spender, amount);
        return true;
    }

    function transferFrom(address from, address to, uint256 amount) public override onlyKyc returns (bool) {
        _requireTransferAllowed(from, to);
        require(_allowances[from][msg.sender] >= amount, "Allowance exceeded");
        _allowances[from][msg.sender] -= amount;
        _transfer(from, to, amount);
        return true;
    }

    function crossChainTransfer(string calldata targetChain, address to, uint256 amount) external override onlyKyc returns (bool) {
        require(bridge != address(0), "Bridge not set");
        _requireTransferAllowed(msg.sender, bridge);
        _transfer(msg.sender, bridge, amount);
        emit CrossChainTransfer(msg.sender, targetChain, to, amount);
        return true;
    }

    function _requireTransferAllowed(address from, address to) private view {
        require(!_transfersPaused, "Transfers paused");
        require(!_frozen[from], "From frozen");
        require(!_frozen[to], "To frozen");
    }

    function mint(address to, uint256 amount) external override onlyController {
        require(!_mintPaused, "Mint paused");
        require(to != address(0), "Mint to zero");
        if (_maxSupply > 0) {
            require(_totalSupply + amount <= _maxSupply, "Exceeds max supply");
        }
        _totalSupply += amount;
        _balances[to] += amount;
        emit Transfer(address(0), to, amount);
        emit Mint(to, amount);
    }

    function burn(address from, uint256 amount) external override onlyController {
        require(!_burnPaused, "Burn paused");
        require(_balances[from] >= amount, "Insufficient balance");
        _balances[from] -= amount;
        _totalSupply -= amount;
        emit Transfer(from, address(0), amount);
        emit Burn(from, amount);
    }

    function _transfer(address from, address to, uint256 amount) internal {
        require(from != address(0), "Transfer from zero");
        require(to != address(0), "Transfer to zero");
        require(_balances[from] >= amount, "Insufficient balance");
        _balances[from] -= amount;
        _balances[to] += amount;
        emit Transfer(from, to, amount);
    }
}
