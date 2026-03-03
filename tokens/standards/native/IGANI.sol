// SPDX-License-Identifier: MIT
pragma solidity ^0.8.16;

import "./INativeCoin.sol";

/// @title IGANI — интерфейс нативной монеты управления GANI (Ganimed Governance)
/// @notice Фиксированное предложение 100000000000000 (100M при 6 decimals). Минт только отдельным контрактом. Управляется только внешним контрактом. Только governance.
/// @dev L1-монета; распределения (DAO, гранты, vesting) регулируются отдельными контрактами, использующими этот интерфейс.

interface IGANI is INativeCoin {
    function name() external view returns (string memory);
    function symbol() external view returns (string memory);
    function decimals() external view returns (uint8);
}
