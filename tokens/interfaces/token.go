// tokens/interfaces/token.go

package interfaces

import (
	"context"
	"math/big"
)

// TokenInterface определяет базовый интерфейс для всех токенов
type TokenInterface interface {
	// Базовые методы
	GetAddress() string
	GetName() string
	GetSymbol() string
	GetDecimals() uint8
	GetTotalSupply() *big.Int
	GetStandard() string

	// Методы для работы с балансами
	GetBalance(ctx context.Context, address string) (*big.Int, error)
	Transfer(ctx context.Context, from, to string, amount *big.Int) error
	Approve(ctx context.Context, owner, spender string, amount *big.Int) error
	Allowance(ctx context.Context, owner, spender string) (*big.Int, error)

	// Методы для работы с событиями
	EmitTransfer(ctx context.Context, from, to string, amount *big.Int) error
	EmitApproval(ctx context.Context, owner, spender string, amount *big.Int) error
}
