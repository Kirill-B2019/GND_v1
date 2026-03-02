// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/// @title GNDToken — утилитарная монета GND (Ganymede Coin) на контракте
/// @notice GNDst-1/ERC-20. Начальная эмиссия в конструкторе на treasury; минтинг после деплоя отключён.
/// @dev Макс. эмиссия 1 млрд GND (18 decimals). Начальная циркуляция по ТЗ — 100M на treasury для распределения.

contract GNDToken {
    string public constant name = "Ganymede Coin";
    string public constant symbol = "GND";
    uint8 public constant decimals = 18;

    /// @notice Максимальное предложение: 1e9 * 10^18
    uint256 public constant MAX_SUPPLY = 1_000_000_000 * 10**18;

    uint256 private _totalSupply;
    mapping(address => uint256) private _balances;
    mapping(address => mapping(address => uint256)) private _allowances;

    address public immutable treasury;

    event Transfer(address indexed from, address indexed to, uint256 value);
    event Approval(address indexed owner, address indexed spender, uint256 value);

    error ExceedsMaxSupply();
    error MintingDisabled();

    /// @param initialSupply Начальная эмиссия (например 100_000_000 * 10**18 для 100M GND)
    /// @param treasuryAddress Адрес казначейства/мигратора, на который минтятся токены
    constructor(uint256 initialSupply, address treasuryAddress) {
        require(treasuryAddress != address(0), "zero treasury");
        require(initialSupply <= MAX_SUPPLY && initialSupply > 0, "invalid initial supply");
        treasury = treasuryAddress;
        _totalSupply = initialSupply;
        _balances[treasuryAddress] = initialSupply;
        emit Transfer(address(0), treasuryAddress, initialSupply);
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

    /// @notice Минтинг отключён навсегда после деплоя (только начальный mint в конструкторе).
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
