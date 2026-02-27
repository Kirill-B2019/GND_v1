// | KB @CerbeRus - Nexus Invest Team
package core

import (
	"GND/types"
	"math/big"
)

// BlockchainIface определяет интерфейс для блокчейна
type BlockchainIface interface {
	LatestBlock() (*Block, error)
	GetBlockByNumber(number uint64) (*Block, error)
	AddTx(tx *Transaction) error
	GetTxStatus(hash string) (string, error)
}

// StateIface определяет интерфейс для состояния
type StateIface interface {
	GetBalance(address types.Address, symbol string) *big.Int
	AddBalance(address types.Address, symbol string, amount *big.Int) error
	SubBalance(address types.Address, symbol string, amount *big.Int) error
	CallStatic(tx *Transaction) (*types.ExecutionResult, error)
	ApplyTransaction(tx *Transaction) error
	GetNonce(address types.Address) int64
}
