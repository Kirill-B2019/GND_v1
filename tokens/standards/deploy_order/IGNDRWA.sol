// SPDX-License-Identifier: MIT
pragma solidity ^0.8.16;

import "./IGNDst1.sol";

/// @title IGNDRWA — RWA поверх стандарта GND-st1
/// @notice Токен реальных активов: полный интерфейс GND-st1 плюс пауза, заморозка, mint/burn, maxSupply.
/// @dev Наследует IGNDst1; добавляет эмиссию/сжигание, паузы, заморозку адресов.
interface IGNDRWA is IGNDst1 {
    function mint(address to, uint256 amount) external;
    function burn(address from, uint256 amount) external;

    function setTransfersPaused(bool paused) external;
    function setMintPaused(bool paused) external;
    function setBurnPaused(bool paused) external;
    function transfersPaused() external view returns (bool);
    function mintPaused() external view returns (bool);
    function burnPaused() external view returns (bool);

    function setFrozen(address account, bool frozen) external;
    function isFrozen(address account) external view returns (bool);

    function maxSupply() external view returns (uint256);

    event Mint(address indexed to, uint256 value);
    event Burn(address indexed from, uint256 value);
    event TransfersPausedChanged(bool paused);
    event MintPausedChanged(bool paused);
    event BurnPausedChanged(bool paused);
    event FrozenChanged(address indexed account, bool frozen);
}
