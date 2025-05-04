package core

import (
	"crypto/sha256"
	"encoding/hex"
)

// TxType определяет тип транзакции
type TxType string

const (
	TxTransfer       TxType = "transfer"        // Обычный перевод GND
	TxContractDeploy TxType = "contract_deploy" // Деплой смарт-контракта
	TxContractCall   TxType = "contract_call"   // Вызов функции смарт-контракта
)

// Transaction - структура транзакции в блокчейне ГАНИМЕД
type Transaction struct {
	Hash      string // Хеш транзакции (вычисляется при создании)
	From      string // Адрес отправителя (hex/base58)
	To        string // Адрес получателя (или адрес контракта)
	Value     uint64 // Сумма перевода в GND (или 0 для вызова/деплоя контракта)
	GasPrice  uint64 // Цена газа (GND за единицу газа)
	GasLimit  uint64 // Лимит газа для транзакции
	Nonce     uint64 // Порядковый номер транзакции отправителя
	Data      []byte // Поле для данных (например, байткод контракта или input для вызова)
	Type      TxType // Тип транзакции
	Signature string // Подпись отправителя (hex)
	// Можно добавить поле ChainID для мультицепей
}

// NewTransaction - конструктор транзакции
func NewTransaction(from, to string, value, gasPrice, gasLimit, nonce uint64, data []byte, txType TxType, signature string) *Transaction {
	tx := &Transaction{
		From:      from,
		To:        to,
		Value:     value,
		GasPrice:  gasPrice,
		GasLimit:  gasLimit,
		Nonce:     nonce,
		Data:      data,
		Type:      txType,
		Signature: signature,
	}
	tx.Hash = tx.CalculateHash()
	return tx
}

// CalculateHash - вычисляет хеш транзакции (без подписи)
func (tx *Transaction) CalculateHash() string {
	input := tx.From +
		tx.To +
		uintToString(tx.Value) +
		uintToString(tx.GasPrice) +
		uintToString(tx.GasLimit) +
		uintToString(tx.Nonce) +
		string(tx.Data) +
		string(tx.Type)
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}

// Hash возвращает хеш транзакции (для совместимости)
func (tx *Transaction) HashString() string {
	return tx.Hash
}

// uintToString - вспомогательная функция для конвертации uint64 в строку
func uintToString(n uint64) string {
	// Можно заменить на strconv.FormatUint(n, 10)
	b := make([]byte, 0, 20)
	for n > 0 {
		b = append([]byte{byte(n%10) + '0'}, b...)
		n /= 10
	}
	if len(b) == 0 {
		return "0"
	}
	return string(b)
}
