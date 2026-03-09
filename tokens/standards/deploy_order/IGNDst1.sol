// SPDX-License-Identifier: MIT
pragma solidity ^0.8.16;

// Копия tokens/standards/gndst1/IGNDst1.sol для автономной компиляции deploy_order.
// Синхронизировать при изменении канонического интерфейса.

/// @title IGNDst1 — интерфейс стандарта GNDst-1
/// @notice Совместим с ERC-20, TRC-20. Расширения: KYC, snapshot, дивиденды, crossChain, модули.

interface IGNDst1 {
    function totalSupply() external view returns (uint256);
    function balanceOf(address account) external view returns (uint256);
    function transfer(address to, uint256 amount) external returns (bool);
    function approve(address spender, uint256 amount) external returns (bool);
    function allowance(address owner, address spender) external view returns (uint256);
    function transferFrom(address from, address to, uint256 amount) external returns (bool);

    function crossChainTransfer(
        string calldata targetChain,
        address to,
        uint256 amount
    ) external returns (bool);

    function setKycStatus(address user, bool status) external;
    function isKycPassed(address user) external view returns (bool);

    function moduleCall(
        bytes32 moduleId,
        bytes calldata data
    ) external returns (bytes memory);

    function snapshot() external returns (uint256);
    function getSnapshotBalance(address user, uint256 snapshotId)
    external
    view
    returns (uint256);

    function claimDividends(uint256 snapshotId) external;

    function registerModule(
        bytes32 moduleId,
        address moduleAddress,
        string calldata name
    ) external;

    event Transfer(address indexed from, address indexed to, uint256 value);
    event Approval(address indexed owner, address indexed spender, uint256 value);
    event CrossChainTransfer(address indexed from, string targetChain, address indexed to, uint256 value);
    event KycStatusChanged(address indexed user, bool status);
    event SnapshotCreated(uint256 indexed snapshotId, uint256 timestamp);
    event DividendClaimed(address indexed user, uint256 amount, uint256 snapshotId);
    event ModuleRegistered(bytes32 indexed moduleId, address indexed moduleAddress, string name);
}
