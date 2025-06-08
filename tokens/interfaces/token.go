// tokens/interfaces/token.go

package interfaces

import (
	"GND/core"
	"context"
	"math/big"
)

// TokenInterface определяет интерфейс для работы с токенами
type TokenInterface interface {
	// GetAddress возвращает адрес токена
	GetAddress() core.Address

	// GetName возвращает имя токена
	GetName() string

	// GetSymbol возвращает символ токена
	GetSymbol() string

	// GetDecimals возвращает количество десятичных знаков
	GetDecimals() uint8

	// GetTotalSupply возвращает общее предложение токенов
	GetTotalSupply() *big.Int

	// GetBalance возвращает баланс токенов для адреса
	GetBalance(ctx context.Context, address core.Address) (*big.Int, error)

	// Transfer переводит токены с одного адреса на другой
	Transfer(ctx context.Context, from, to core.Address, amount *big.Int) error

	// Approve разрешает расход токенов
	Approve(ctx context.Context, owner, spender core.Address, amount *big.Int) error

	// Allowance возвращает разрешенное количество токенов
	Allowance(ctx context.Context, owner, spender core.Address) (*big.Int, error)
}
