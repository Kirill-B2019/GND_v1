// SPDX-License-Identifier: MIT
pragma solidity ^0.8.16;

/// @title IGNDRWA — интерфейс стандарта GND-RWA
/// @notice Токен реальных активов (RWA), управляемый контрактом-контроллером.
/// @dev Расширения: пауза, заморозка, cap; KYC, snapshot/дивиденды, модули (KycStatusChanged, SnapshotCreated, DividendClaimed, ModuleRegistered).

interface IGNDRWA {
    // --- ERC-20 ---
    function totalSupply() external view returns (uint256);
    function balanceOf(address account) external view returns (uint256);
    function transfer(address to, uint256 amount) external returns (bool);
    function approve(address spender, uint256 amount) external returns (bool);
    function allowance(address owner, address spender) external view returns (uint256);
    function transferFrom(address from, address to, uint256 amount) external returns (bool);

    // --- Эмиссия (только контроллер) ---
    function mint(address to, uint256 amount) external;
    function burn(address from, uint256 amount) external;

    // --- Пауза (только контроллер) ---
    function setTransfersPaused(bool paused) external;
    function setMintPaused(bool paused) external;
    function setBurnPaused(bool paused) external;
    function transfersPaused() external view returns (bool);
    function mintPaused() external view returns (bool);
    function burnPaused() external view returns (bool);

    // --- Заморозка адресов (только контроллер) ---
    function setFrozen(address account, bool frozen) external;
    function isFrozen(address account) external view returns (bool);

    // --- Лимит эмиссии (0 = без лимита) ---
    function maxSupply() external view returns (uint256);

    // --- KYC (только контроллер) ---
    function setKycStatus(address user, bool status) external;
    function isKycPassed(address user) external view returns (bool);

    // --- Снимки и дивиденды ---
    function snapshot() external returns (uint256);
    function getSnapshotBalance(address user, uint256 snapshotId) external view returns (uint256);
    function claimDividends(uint256 snapshotId) external;

    // --- Модули (только контроллер — регистрация) ---
    function registerModule(bytes32 moduleId, address moduleAddress, string calldata name) external;
    function moduleCall(bytes32 moduleId, bytes calldata data) external returns (bytes memory);

    event Transfer(address indexed from, address indexed to, uint256 value);
    event Approval(address indexed owner, address indexed spender, uint256 value);
    event Mint(address indexed to, uint256 value);
    event Burn(address indexed from, uint256 value);
    event TransfersPausedChanged(bool paused);
    event MintPausedChanged(bool paused);
    event BurnPausedChanged(bool paused);
    event FrozenChanged(address indexed account, bool frozen);
    event KycStatusChanged(address indexed user, bool status);
    event SnapshotCreated(uint256 indexed snapshotId, uint256 timestamp);
    event DividendClaimed(address indexed user, uint256 amount, uint256 snapshotId);
    event ModuleRegistered(bytes32 indexed moduleId, address indexed moduleAddress, string name);
    event ModuleCall(bytes32 indexed moduleId, address indexed caller);
}
