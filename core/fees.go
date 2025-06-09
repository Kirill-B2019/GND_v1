package core

import "math/big"

// Fees related code
// CalculateTxFee вычисляет комиссию за транзакцию
func CalculateTxFee(tx *Transaction) *big.Int {
	fee := new(big.Int).Mul(tx.GasPrice, big.NewInt(int64(tx.GasLimit)))
	return fee
}
