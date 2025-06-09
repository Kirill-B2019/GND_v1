// vm/integration.go

package vm

import (
	"GND/tokens/registry"
	"GND/tokens/standards/gndst1"
	"GND/types"
	"context"
	"fmt"
	"math/big"
	"time"
)

// DeployGNDst1Token деплоит контракт GNDst1
func (e *EVM) DeployGNDst1Token(ctx context.Context, name, symbol string, decimals uint8, totalSupply *big.Int, from string) (string, error) {
	if from == "" {
		return "", fmt.Errorf("не удалось получить адрес отправителя")
	}

	// Генерируем байткод
	bytecode, err := generateBytecode(name, symbol, decimals, totalSupply)
	if err != nil {
		return "", fmt.Errorf("ошибка генерации байткода: %v", err)
	}

	// Создаем метаданные контракта
	meta := types.ContractMeta{
		Name:        name,
		Symbol:      symbol,
		Description: fmt.Sprintf("%s Token Contract", name),
		Standard:    "gndst1",
		Owner:       from,
		Params: map[string]string{
			"totalSupply": totalSupply.String(),
			"decimals":    fmt.Sprintf("%d", decimals),
		},
	}

	// Деплоим контракт
	addr, err := e.DeployContract(
		from,
		bytecode,
		meta,
		3_000_000,      // gasLimit
		20_000_000_000, // gasPrice
		0,              // nonce
		"",             // signature
		totalSupply,
	)
	if err != nil {
		return "", fmt.Errorf("ошибка деплоя контракта: %v", err)
	}

	// Регистрация токена
	token := gndst1.NewGNDst1(
		addr,
		name,
		symbol,
		decimals,
		totalSupply,
		nil, // TODO: добавить pool
	)
	if err := registry.RegisterToken(addr, token); err != nil {
		return "", fmt.Errorf("ошибка регистрации токена: %v", err)
	}

	// Эмитим событие
	if e.eventManager != nil {
		event := &types.Event{
			Type:        "TokenDeployed",
			Contract:    addr,
			FromAddress: from,
			Timestamp:   time.Now(),
			Metadata: map[string]interface{}{
				"address":     addr,
				"name":        name,
				"symbol":      symbol,
				"decimals":    decimals,
				"totalSupply": totalSupply.String(),
				"standard":    "gndst1",
				"owner":       from,
			},
		}
		if err := e.eventManager.SaveEvent(ctx, event); err != nil {
			return "", fmt.Errorf("ошибка сохранения события: %v", err)
		}
	}

	return addr, nil
}
