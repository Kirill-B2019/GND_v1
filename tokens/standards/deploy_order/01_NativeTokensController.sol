// SPDX-License-Identifier: MIT
pragma solidity ^0.8.16;

/// @title Контракт-контроллер для GND и GANI (стандарт GND-st1)
/// @notice Деплоится первым (шаг 1). Его адрес передаётся в конструкторы GNDToken и GANIToken.
/// @dev owner задаётся при деплое из config (gndself_address в native_contracts.json). Все функции изменения состояния защищены onlyOwner.
/// @dev Invariants: owner immutable; gndToken/ganiToken задаются один раз; только onlyOwner меняет состояние; reentrancy в доверенные токены. См. INVARIANTS.md.
contract NativeTokensController {
    address public immutable owner;
    address public gndToken;
    address public ganiToken;

    error OnlyOwner();
    error GaniTokenNotSet();
    error GndTokenNotSet();
    error NotContract();
    error TokenAlreadySet();
    error InvalidTokenContract();
    error ZeroOwner();

    event GndTokenSet(address indexed token);
    event GaniTokenSet(address indexed token);
    event GANIMinted(address indexed to, uint256 amount);
    event KycGndSet(address indexed user, bool status);
    event KycGaniSet(address indexed user, bool status);

    /// @param owner_ Адрес владельца (системный кошелёк ГАНИМЕД — gndself_address из config/native_contracts.json).
    constructor(address owner_) {
        if (owner_ == address(0)) revert ZeroOwner();
        owner = owner_;
    }

    modifier onlyOwner() {
        if (msg.sender != owner) revert OnlyOwner();
        _;
    }

    /// @dev Проверяет, что адрес — контракт с кодом и поддерживает totalSupply() (ожидаемый интерфейс токена GND-st1/ERC-20).
    function _isValidTokenContract(address token) private view returns (bool) {
        if (token == address(0)) return false;
        if (token.code.length == 0) return false;
        // totalSupply() selector 0x18160ddd — есть у ERC-20 и GND-st1
        (bool ok, ) = token.staticcall(abi.encodeWithSelector(0x18160ddd));
        return ok;
    }

    /// @notice Задать адрес контракта GND один раз (опционально, для вызова setKycGnd). Повторная запись запрещена.
    /// @custom:security onlyOwner — вызов разрешён только владельцу контракта.
    function setGndToken(address _gndToken) external onlyOwner {
        if (gndToken != address(0)) revert TokenAlreadySet();
        if (!_isValidTokenContract(_gndToken)) revert InvalidTokenContract();
        gndToken = _gndToken;
        emit GndTokenSet(_gndToken);
    }

    /// @notice Задать адрес контракта GANI один раз (вызвать после деплоя GANIToken). Повторная запись запрещена.
    /// @custom:security onlyOwner — вызов разрешён только владельцу контракта.
    function setGaniToken(address _ganiToken) external onlyOwner {
        if (ganiToken != address(0)) revert TokenAlreadySet();
        if (!_isValidTokenContract(_ganiToken)) revert InvalidTokenContract();
        ganiToken = _ganiToken;
        emit GaniTokenSet(_ganiToken);
    }

    /// @notice Выпустить GANI на адрес to (только владелец). Лимиты проверяются в GANIToken.
    /// @custom:security onlyOwner — вызов разрешён только владельцу контракта.
    function mintGANI(address to, uint256 amount) external onlyOwner {
        if (ganiToken == address(0)) revert GaniTokenNotSet();
        (bool ok, ) = ganiToken.call(abi.encodeWithSelector(0x40c10f19, to, amount));
        require(ok, "GANI mint failed");
        emit GANIMinted(to, amount);
    }

    /// @notice Включить/выключить KYC для адреса user на токене GND (только владелец).
    /// @custom:security onlyOwner — вызов разрешён только владельцу контракта.
    function setKycGnd(address user, bool status) external onlyOwner {
        if (gndToken == address(0)) revert GndTokenNotSet();
        (bool ok, ) = gndToken.call(abi.encodeWithSignature("setKycStatus(address,bool)", user, status));
        require(ok, "setKycStatus GND failed");
        emit KycGndSet(user, status);
    }

    /// @notice Включить/выключить KYC для адреса user на токене GANI (только владелец).
    /// @custom:security onlyOwner — вызов разрешён только владельцу контракта.
    function setKycGani(address user, bool status) external onlyOwner {
        if (ganiToken == address(0)) revert GaniTokenNotSet();
        (bool ok, ) = ganiToken.call(abi.encodeWithSignature("setKycStatus(address,bool)", user, status));
        require(ok, "setKycStatus GANI failed");
        emit KycGaniSet(user, status);
    }
}
