// SPDX-License-Identifier: MIT
pragma solidity ^0.8.16;

import "./INativeCoin.sol";

/// @title IGND — интерфейс нативной утилитарной монеты GND (Ganimed)
/// @notice Макс. предложение 1e27 (1 млрд GND), decimals 18. Управляется только внешним контрактом. Используется для комиссий, стейкинга, наград валидаторов.
/// @dev L1-монета; распределения (пулы, vesting) регулируются отдельными контрактами, использующими этот интерфейс.

interface IGND is INativeCoin {
    function name() external view returns (string memory);
    function symbol() external view returns (string memory);
    function decimals() external view returns (uint8);
}
