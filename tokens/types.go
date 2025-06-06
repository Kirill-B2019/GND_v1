// tokens/types.go

package tokens

import "math/big"

type TokenInfo struct {
	Name        string
	Symbol      string
	Decimals    uint8
	TotalSupply *big.Int
	Address     string
}

type TransferEvent struct {
	From   string
	To     string
	Value  *big.Int
	Token  string
	TxHash string
	Block  uint64
}
