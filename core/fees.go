package core

import "math/big"

// Fees related code
// CalculateTxFee возвращает комиссию за транзакцию
func CalculateTxFee(tx *Transaction) *big.Int {
	return tx.Fee
}
