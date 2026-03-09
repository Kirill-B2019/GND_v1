// SPDX-License-Identifier: MIT
pragma solidity ^0.8.16;

import "./gndst1Base.sol";

/// @title GNDToken — утилитарная монета GND (Ganimed) по стандарту GND-st1
/// @notice Деплой: шаг 2. Параметры конструктора: initialSupply (1e27), bridgeAddress, controllerContract (адрес из шага 1).
/// @dev Реализует IGNDst1 через наследование GNDst1Token. Минтинг отключён; эмиссия только в конструкторе. Инварианты см. INVARIANTS.md (GNDst1Token).
contract GNDToken is GNDst1Token {
    constructor(
        uint256 initialSupply,
        address bridgeAddress,
        address controllerContract
    ) GNDst1Token(initialSupply, bridgeAddress, controllerContract) {}
}
