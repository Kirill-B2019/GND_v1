// tokens/deployer/deployer.go

package deployer

import (
	"GND/core"
	"GND/tokens/registry"
	"GND/vm"
	"math/big"
)

func DeployGNDst1Token(evm *vm.EVM, from string, name, symbol string, decimals uint8, totalSupply *big.Int) (string, error) {
	bytecode := generateBytecode(name, symbol, decimals, totalSupply)
	meta := vm.ContractMeta{
		Name:     name,
		Symbol:   symbol,
		Standard: "gndst1",
		Owner:    core.Address(from),
	}

	addr, err := evm.DeployContract(
		from,
		bytecode,
		meta,
		3_000_000,
		20_000_000_000,
		0,
		"",
	)
	if err != nil {
		return "", err
	}

	err = registry.RegisterToken(addr, nil) // можно передать интерфейс
	if err != nil {
		return "", err
	}

	return addr, nil
}

func generateBytecode(name, symbol string, decimals uint8, totalSupply *big.Int) []byte {
	// Здесь будет генерация из Solidity
	return []byte{0x60, 0x60, 0x60, 0x70} // заглушка
}
