// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "./IGANI.sol";

/// @title GANICoinBase — базовая обёртка нативной монеты GANI (Ganymede Governance)
/// @notice GANI — нативная L1-монета с фиксированной эмиссией; только governance. Балансы и переводы обрабатываются протоколом (native_balances).
/// @dev Развёртывается по фиксированному адресу; вызовы должны обрабатываться precompile ноды. Распределения (DAO, гранты, vesting, DEX) реализуются отдельными контрактами, которые работают с GANI через этот интерфейс.

contract GANICoinBase is IGANI {
    string public constant override name = "Ganymede Governance";
    string public constant override symbol = "GANI";
    uint8 public constant override decimals = 6;

    error NativeCoinViewUsePrecompile();
    error NativeCoinTransferUseL1();

    function totalSupply() external pure override returns (uint256) {
        revert NativeCoinViewUsePrecompile();
    }

    function balanceOf(address) external pure override returns (uint256) {
        revert NativeCoinViewUsePrecompile();
    }

    function transfer(address, uint256) external pure override returns (bool) {
        revert NativeCoinTransferUseL1();
    }

    function allowance(address, address) external pure override returns (uint256) {
        revert NativeCoinViewUsePrecompile();
    }

    function approve(address, uint256) external pure override returns (bool) {
        revert NativeCoinTransferUseL1();
    }

    function transferFrom(address, address, uint256) external pure override returns (bool) {
        revert NativeCoinTransferUseL1();
    }
}
