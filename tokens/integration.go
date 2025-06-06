// tokens/integration.go

package tokens

import (
	"context"
	"github.com/gogo/protobuf/test/issue312/events"
	"math/big"

	"GND/core"
	"GND/vm"
)

func (e *EVM) DeployGNDst1Token(ctx context.Context, name, symbol string, decimals uint8, totalSupply *big.Int) (string, error) {
	from := e.config.State.GetEOA()
	bytecode := generateBytecode(name, symbol, decimals, totalSupply)
	meta := vm.ContractMeta{
		Name:     name,
		Symbol:   symbol,
		Standard: "gndst1",
		Owner:    core.Address(from),
	}
	addr, err := e.evm.DeployContract(
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

	// Регистрация токена
	RegisterToken(addr, NewGNDst1(totalSupply, addr))

	// Публикация события
	events.PublishEvent("contract", map[string]interface{}{
		"event": "deployed",
		"token": addr,
		"name":  name,
		"owner": from,
	})

	return addr, nil
}
