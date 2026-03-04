// SPDX-License-Identifier: MIT
pragma solidity ^0.8.16;

/// @title GANIToken — governance-токен GANI (Ganimed Governance) на контракте
/// @notice GNDst-1/ERC-20. Фиксированное предложение 100000000000000 (100M при 6 decimals). Минтинг отключён. Управляется только внешним контрактом (controller).
/// @dev Деплой: шаг 3. Параметр конструктора: controllerAddress = адрес контракта из шага 1 (тот же, что и для GND).

contract GANIToken {
    string public constant name = "Ganimed Governance";
    string public constant symbol = "GANI";
    uint8 public constant decimals = 6;

    /// @notice Фиксированное предложение: 100000000000000 (100M * 10^6)
    uint256 public constant TOTAL_SUPPLY = 100000000000000;

    uint256 private _totalSupply;
    mapping(address => uint256) private _balances;
    mapping(address => mapping(address => uint256)) private _allowances;

    address public immutable controller;

    event Transfer(address indexed from, address indexed to, uint256 value);
    event Approval(address indexed owner, address indexed spender, uint256 value);

    error MintingDisabled();
    error ZeroController();
    error ControllerMustBeContract();
    error ZeroAddress();
    error ExceedsTotalSupply();

    /// @param controllerContract Адрес контракта из шага 1 (01_NativeTokensController) — тот же, что использовали для GND
    constructor(address controllerContract) {
        if (controllerContract == address(0)) revert ZeroController();
        if (_isContract(controllerContract) == false) revert ControllerMustBeContract();
        controller = controllerContract;
        _totalSupply = TOTAL_SUPPLY;
        _balances[controllerContract] = TOTAL_SUPPLY;
        emit Transfer(address(0), controllerContract, TOTAL_SUPPLY);
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

    /// @notice Минт только с адреса контроллера. Лимит — TOTAL_SUPPLY.
    function mint(address to, uint256 amount) external {
        if (msg.sender != controller) revert MintingDisabled();
        if (to == address(0)) revert ZeroAddress();
        if (_totalSupply + amount > TOTAL_SUPPLY) revert ExceedsTotalSupply();
        unchecked {
            _totalSupply += amount;
            _balances[to] += amount;
        }
        emit Transfer(address(0), to, amount);
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
