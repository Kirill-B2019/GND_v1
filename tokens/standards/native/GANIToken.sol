// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/// @title GANIToken — governance-токен GANI (Ganymede Governance) на контракте
/// @notice GNDst-1/ERC-20. Фиксированная эмиссия 100 млн GANI (6 decimals). Минтинг отключён навсегда.
/// @dev Вся эмиссия в конструкторе на один адрес (treasury); дальнейшее распределение — вручную/через контракты.

contract GANIToken {
    string public constant name = "Ganymede Governance";
    string public constant symbol = "GANI";
    uint8 public constant decimals = 6;

    /// @notice Фиксированное предложение: 100M * 10^6
    uint256 public constant TOTAL_SUPPLY = 100_000_000 * 10**6;

    uint256 private _totalSupply;
    mapping(address => uint256) private _balances;
    mapping(address => mapping(address => uint256)) private _allowances;

    address public immutable treasury;

    event Transfer(address indexed from, address indexed to, uint256 value);
    event Approval(address indexed owner, address indexed spender, uint256 value);

    error MintingDisabled();

    /// @param treasuryAddress Адрес, на который выпускается вся эмиссия (100M GANI)
    constructor(address treasuryAddress) {
        require(treasuryAddress != address(0), "zero treasury");
        treasury = treasuryAddress;
        _totalSupply = TOTAL_SUPPLY;
        _balances[treasuryAddress] = TOTAL_SUPPLY;
        emit Transfer(address(0), treasuryAddress, TOTAL_SUPPLY);
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

    /// @notice Минтинг отключён навсегда.
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
