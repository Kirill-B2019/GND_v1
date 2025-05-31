package core

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strconv"
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
	From      string // Адрес отправителя (Base58, с префиксом GND, GND_, GN, GN_)
	To        string // Адрес получателя (или адрес контракта, с префиксом)
	Value     uint64 // Сумма перевода в GND (или 0 для вызова/деплоя контракта)
	GasPrice  uint64 // Цена газа (GND за единицу газа)
	GasLimit  uint64 // Лимит газа для транзакции
	Nonce     uint64 // Порядковый номер транзакции отправителя
	Data      []byte // Поле для данных (например, байткод контракта или input для вызова)
	Type      TxType // Тип транзакции
	Signature string // Подпись отправителя (hex)
	// Можно добавить поле ChainID для мультицепей
}

// NewTransaction - конструктор транзакции с обязательной валидацией адресов
func NewTransaction(from, to string, value, gasPrice, gasLimit, nonce uint64, data []byte, txType TxType, signature string) (*Transaction, error) {
	if !ValidateAddress(from) {
		return nil, errors.New("invalid sender address (from)")
	}
	if !ValidateAddress(to) {
		return nil, errors.New("invalid recipient address (to)")
	}
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
	return tx, nil
}

// CalculateHash - вычисляет хеш транзакции (без подписи)
func (tx *Transaction) CalculateHash() string {
	input := tx.From +
		tx.To +
		strconv.FormatUint(tx.Value, 10) +
		strconv.FormatUint(tx.GasPrice, 10) +
		strconv.FormatUint(tx.GasLimit, 10) +
		strconv.FormatUint(tx.Nonce, 10) +
		string(tx.Data) +
		string(tx.Type)
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}

// HashString возвращает хеш транзакции (для совместимости)
func (tx *Transaction) HashString() string {
	return tx.Hash
}
