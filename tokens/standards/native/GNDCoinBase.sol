// SPDX-License-Identifier: MIT
pragma solidity ^0.8.16;

import "./IGND.sol";

/// @title GNDCoinBase — базовая обёртка нативной монеты GND (Ganimed)
/// @notice GND — нативная L1-монета; балансы и переводы обрабатываются протоколом (native_balances). Управляется только внешним контрактом.
/// @dev Развёртывается по фиксированному адресу; вызовы должны обрабатываться precompile ноды (balanceOf/totalSupply — чтение из state, transfer/approve/transferFrom — вызов L1). Распределения (пулы, vesting, сбор комиссий) реализуются отдельными контрактами, которые работают с GND через этот интерфейс.

contract GNDCoinBase is IGND {
    string public constant override name = "Ganimed";
    string public constant override symbol = "GND";
    uint8 public constant override decimals = 18;

    error NativeCoinViewUsePrecompile();
    error NativeCoinTransferUseL1();

    /// @notice Общее предложение — источник истины в ноде; при вызове из контракта требуется precompile.
    function totalSupply() external pure override returns (uint256) {
        revert NativeCoinViewUsePrecompile();
    }

    /// @notice Баланс адреса — источник истины в ноде (native_balances); при вызове из контракта требуется precompile.
    function balanceOf(address) external pure override returns (uint256) {
        revert NativeCoinViewUsePrecompile();
    }

    /// @notice Перевод GND выполняется через L1 (нода/precompile).
    function transfer(address, uint256) external pure override returns (bool) {
        revert NativeCoinTransferUseL1();
    }

    /// @notice Allowance для нативной монеты может поддерживаться precompile или не использоваться.
    function allowance(address, address) external pure override returns (uint256) {
        revert NativeCoinViewUsePrecompile();
    }

    /// @notice Approve выполняется через L1/precompile при поддержке делегирования.
    function approve(address, uint256) external pure override returns (bool) {
        revert NativeCoinTransferUseL1();
    }

    /// @notice TransferFrom выполняется через L1/precompile.
    function transferFrom(address, address, uint256) external pure override returns (bool) {
        revert NativeCoinTransferUseL1();
    }
}
