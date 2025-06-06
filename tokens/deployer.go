// tokens/deployer.go

package tokens

import (
	"GND/core"
	"GND/vm"
	"math/big"
)

func DeployGNDst1Token(evm *vm.EVM, from, name, symbol string, decimals uint8, totalSupply *big.Int) (string, error) {
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
		3_000_000,      // gas limit
		20_000_000_000, // gas price
		0,              // nonce
		"",             // signature
	)
	if err != nil {
		return "", err
	}
	RegisterToken(addr, NewGNDst1(totalSupply, addr))
	return addr, nil
}

func generateBytecode(name, symbol string, decimals uint8, totalSupply *big.Int) []byte {
	// Здесь должна быть логика генерации байткода или вызов компилятора
	return []byte{0x60, 0x60, 0x60, 0x70} // заглушка
}
