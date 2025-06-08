// tokens/interfaces/token.go

package interfaces

import (
	"math/big"
)

type TokenInterface interface {
	Name() string
	Symbol() string
	Decimals() uint8
	TotalSupply() *big.Int
	BalanceOf(address string) *big.Int
	Transfer(from, to string, amount *big.Int) bool
	Allowance(owner, spender string) *big.Int
	Approve(spender string, amount *big.Int) bool
	TransferFrom(from, to string, amount *big.Int) bool
	CrossChainTransfer(targetChain string, to string, amount *big.Int) bool
	SetKycStatus(user string, status bool)
	IsKycPassed(user string) bool
	Meta() TokenMeta
}

type TokenMeta struct {
	Address     string
	Owner       string
	Standard    string
	Name        string
	Symbol      string
	Decimals    uint8
	Description string
}
