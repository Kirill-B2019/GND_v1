package core

// Fees related code
// CalculateTxFee вычисляет комиссию за транзакцию
func CalculateTxFee(tx *Transaction) uint64 {
	return tx.GasPrice * tx.GasLimit
}
