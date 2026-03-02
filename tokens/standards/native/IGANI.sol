// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "./INativeCoin.sol";

/// @title IGANI — интерфейс нативной монеты управления GANI (Ganymede Governance)
/// @notice Фиксированная эмиссия 100 млн GANI, decimals 18. Только governance.
/// @dev L1-монета; распределения (DAO, гранты, vesting) регулируются отдельными контрактами, использующими этот интерфейс.

interface IGANI is INativeCoin {
    function name() external view returns (string memory);
    function symbol() external view returns (string memory);
    function decimals() external view returns (uint8);
}
