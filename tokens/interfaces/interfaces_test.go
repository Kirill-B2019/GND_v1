// | KB @CerbeRus - Nexus Invest Team
package interfaces

import (
	"context"
	"math/big"
	"testing"
)

// TokenInterface определён в token.go; тест проверяет, что пакет компилируется и интерфейс доступен.
func TestTokenInterface_Defined(t *testing.T) {
	var _ TokenInterface = (*tokenInterfaceStub)(nil)
}

// Заглушка для проверки соответствия интерфейсу (без реализации всех методов в тесте не скомпилируется).
type tokenInterfaceStub struct{}

func (tokenInterfaceStub) GetAddress() string       { return "" }
func (tokenInterfaceStub) GetName() string          { return "" }
func (tokenInterfaceStub) GetSymbol() string        { return "" }
func (tokenInterfaceStub) GetDecimals() uint8       { return 0 }
func (tokenInterfaceStub) GetTotalSupply() *big.Int { return nil }
func (tokenInterfaceStub) GetStandard() string      { return "" }
func (tokenInterfaceStub) GetBalance(ctx context.Context, address string) (*big.Int, error) {
	return nil, nil
}
func (tokenInterfaceStub) Transfer(ctx context.Context, from, to string, amount *big.Int) error {
	return nil
}
func (tokenInterfaceStub) Approve(ctx context.Context, owner, spender string, amount *big.Int) error {
	return nil
}
func (tokenInterfaceStub) Allowance(ctx context.Context, owner, spender string) (*big.Int, error) {
	return nil, nil
}
func (tokenInterfaceStub) EmitTransfer(ctx context.Context, from, to string, amount *big.Int) error {
	return nil
}
func (tokenInterfaceStub) EmitApproval(ctx context.Context, owner, spender string, amount *big.Int) error {
	return nil
}
