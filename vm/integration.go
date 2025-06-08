// tokens/integration.go

package vm

import (
	"GND/tokens/deployer"
	"GND/tokens/registry"
	"GND/tokens/standards/gndst1"
	"context"
	"github.com/gogo/protobuf/test/issue312/events"
	"math/big"

	"GND/core"
)

func (e *EVM) DeployGNDst1Token(ctx context.Context, name, symbol string, decimals uint8, totalSupply *big.Int) (string, error) {
	from := e.config.State.GetEOA()
	bytecode := deployer.generateBytecode(name, symbol, decimals, totalSupply)
	meta := ContractMeta{
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
	registry.RegisterToken(addr, gndst1.NewGNDst1(totalSupply, addr))

	// Публикация события
	events.PublishEvent("contract", map[string]interface{}{
		"event": "deployed",
		"token": addr,
		"name":  name,
		"owner": from,
	})

	return addr, nil
}
