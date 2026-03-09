// SPDX-License-Identifier: MIT
pragma solidity ^0.8.16;

/// @title Контракт-контроллер для GND и GANI (стандарт GND-st1)
/// @notice Деплоится первым (шаг 1). Его адрес передаётся в конструкторы GNDToken и GANIToken.
/// @dev Контракты GND/GANI — GND-st1; controller должен быть контрактом. Владелец: setGaniToken, mintGANI, setKycGnd, setKycGani.
contract NativeTokensController {
    address public immutable owner;
    address public gndToken;
    address public ganiToken;

    error OnlyOwner();
    error GaniTokenNotSet();
    error GndTokenNotSet();

    constructor() {
        owner = msg.sender;
    }

    modifier onlyOwner() {
        if (msg.sender != owner) revert OnlyOwner();
        _;
    }

    /// @notice Задать адрес контракта GND (опционально, для вызова setKycGnd).
    function setGndToken(address _gndToken) external onlyOwner {
        gndToken = _gndToken;
    }

    /// @notice Задать адрес контракта GANI (вызвать один раз после деплоя GANIToken).
    function setGaniToken(address _ganiToken) external onlyOwner {
        ganiToken = _ganiToken;
    }

    /// @notice Выпустить GANI на адрес to (только владелец). Лимиты проверяются в GANIToken.
    function mintGANI(address to, uint256 amount) external onlyOwner {
        if (ganiToken == address(0)) revert GaniTokenNotSet();
        (bool ok, ) = ganiToken.call(abi.encodeWithSelector(0x40c10f19, to, amount));
        require(ok, "GANI mint failed");
    }

    /// @notice Включить/выключить KYC для адреса user на токене GND (только владелец).
    function setKycGnd(address user, bool status) external onlyOwner {
        if (gndToken == address(0)) revert GndTokenNotSet();
        (bool ok, ) = gndToken.call(abi.encodeWithSignature("setKycStatus(address,bool)", user, status));
        require(ok, "setKycStatus GND failed");
    }

    /// @notice Включить/выключить KYC для адреса user на токене GANI (только владелец).
    function setKycGani(address user, bool status) external onlyOwner {
        if (ganiToken == address(0)) revert GaniTokenNotSet();
        (bool ok, ) = ganiToken.call(abi.encodeWithSignature("setKycStatus(address,bool)", user, status));
        require(ok, "setKycStatus GANI failed");
    }
}
