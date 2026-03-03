// SPDX-License-Identifier: MIT
pragma solidity ^0.8.16;

/// @title GNDToken — утилитарная монета GND (Ganimed) на контракте
/// @notice GNDst-1/ERC-20. Эмиссия только в конструкторе; минтинг отключён. Управляется только внешним контрактом (controller).
/// @dev Макс. предложение 1e27 (1 млрд GND, 18 decimals).

contract GNDToken {
    string public constant name = "Ganimed";
    string public constant symbol = "GND";
    uint8 public constant decimals = 18;

    /// @notice Максимальное предложение: 1000000000000000000000000000 (1e27)
    uint256 public constant TOTAL_SUPPLY = 1000000000000000000000000000;

    uint256 private _totalSupply;
    mapping(address => uint256) private _balances;
    mapping(address => mapping(address => uint256)) private _allowances;

    address public immutable controller;

    event Transfer(address indexed from, address indexed to, uint256 value);
    event Approval(address indexed owner, address indexed spender, uint256 value);

    error MintingDisabled();
    error ZeroController();
    error ControllerMustBeContract();
    error InvalidInitialSupply();

    /// @param initialSupply Начальная эмиссия (не более TOTAL_SUPPLY)
    /// @param controllerContract Адрес контракта-контроллера (только он управляет; должен быть контрактом)
    constructor(uint256 initialSupply, address controllerContract) {
        if (controllerContract == address(0)) revert ZeroController();
        if (_isContract(controllerContract) == false) revert ControllerMustBeContract();
        if (initialSupply == 0 || initialSupply > TOTAL_SUPPLY) revert InvalidInitialSupply();
        controller = controllerContract;
        _totalSupply = initialSupply;
        _balances[controllerContract] = initialSupply;
        emit Transfer(address(0), controllerContract, initialSupply);
    }

    function _isContract(address account) private view returns (bool) {
        uint256 size;
        assembly {
            size := extcodesize(account)
        }
        return size > 0;
    }

    function totalSupply() external view returns (uint256) {
        return _totalSupply;
    }

    function balanceOf(address account) external view returns (uint256) {
        return _balances[account];
    }

    function allowance(address owner, address spender) external view returns (uint256) {
        return _allowances[owner][spender];
    }

    function transfer(address to, uint256 amount) external returns (bool) {
        _transfer(msg.sender, to, amount);
        return true;
    }

    function approve(address spender, uint256 amount) external returns (bool) {
        _allowances[msg.sender][spender] = amount;
        emit Approval(msg.sender, spender, amount);
        return true;
    }

    function transferFrom(address from, address to, uint256 amount) external returns (bool) {
        uint256 currentAllowance = _allowances[from][msg.sender];
        if (currentAllowance != type(uint256).max) {
            require(currentAllowance >= amount, "insufficient allowance");
            unchecked { _allowances[from][msg.sender] = currentAllowance - amount; }
        }
        _transfer(from, to, amount);
        return true;
    }

    /// @notice Минтинг отключён (дополнительная эмиссия только отдельным контрактом при необходимости).
    function mint(address, uint256) external pure {
        revert MintingDisabled();
    }

    function _transfer(address from, address to, uint256 amount) internal {
        require(from != address(0), "transfer from zero");
        require(to != address(0), "transfer to zero");
        require(_balances[from] >= amount, "insufficient balance");
        unchecked {
            _balances[from] -= amount;
            _balances[to] += amount;
        }
        emit Transfer(from, to, amount);
    }
}
