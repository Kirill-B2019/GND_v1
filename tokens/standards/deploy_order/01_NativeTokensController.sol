// SPDX-License-Identifier: MIT
pragma solidity ^0.8.16;

/// @title Контракт-контроллер для GND и GANI
/// @notice Деплоится первым (шаг 1). Его адрес передаётся в конструкторы GNDToken и GANIToken.
/// @dev Контракты GND/GANI требуют, чтобы controller был контрактом (extcodesize > 0). Владелец может вызывать mintGANI после установки адреса GANI-токена.
contract NativeTokensController {
    address public immutable owner;
    address public ganiToken;

    error OnlyOwner();
    error GaniTokenNotSet();

    constructor() {
        owner = msg.sender;
    }

    modifier onlyOwner() {
        if (msg.sender != owner) revert OnlyOwner();
        _;
    }

    /// @notice Задать адрес контракта GANI (вызвать один раз после деплоя GANIToken).
    function setGaniToken(address _ganiToken) external onlyOwner {
        ganiToken = _ganiToken;
    }

    /// @notice Выпустить GANI на адрес to (только владелец). Лимиты проверяются в GANIToken.
    function mintGANI(address to, uint256 amount) external onlyOwner {
        if (ganiToken == address(0)) revert GaniTokenNotSet();
        (bool ok, ) = ganiToken.call(abi.encodeWithSignature("mint(address,uint256)", to, amount));
        require(ok, "GANI mint failed");
    }
}
