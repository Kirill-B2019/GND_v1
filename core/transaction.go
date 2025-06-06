//core/transaction.go

package core

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"math/big"
	"time"
)

// TxType определяет тип транзакции
type TxType string

const (
	TxNormal         TxType = "normal"
	TxTransfer       TxType = "transfer"        // Обычный перевод GND
	TxContractDeploy TxType = "contract_deploy" // Деплой смарт-контракта
	TxContractCall   TxType = "contract_call"   // Вызов функции смарт-контракта
)

// Transaction - структура транзакции в блокчейне ГАНИМЕД
type Transaction struct {
	ID        string
	Hash      string // Хеш транзакции (вычисляется при создании)
	From      string // Адрес отправителя (Base58, с префиксом GND, GND_, GN, GN_)
	To        string // Адрес получателя (или адрес контракта, с префиксом)
	Symbol    string // Добавляем поле для указания валюты
	Value     *big.Int
	GasPrice  uint64 // Цена газа (GND за единицу газа)
	GasLimit  uint64 // Лимит газа для транзакции
	Nonce     uint64 // Порядковый номер транзакции отправителя
	Data      []byte // Поле для данных (например, байткод контракта или input для вызова)
	Type      TxType // Тип транзакции
	Signature string // Подпись отправителя (hex)
	// Можно добавить поле ChainID для мультицепей
}

// NewTransaction - конструктор транзакции с обязательной валидацией адресов
func NewTransaction(
	from, to, symbol string,
	value *big.Int,
	gasPrice, gasLimit, nonce uint64,
	data []byte,
	txType TxType,
	signature string,
) (*Transaction, error) {
	if !ValidateAddress(from) {
		return nil, errors.New("invalid sender address (from)")
	}
	if !ValidateAddress(to) {
		return nil, errors.New("invalid recipient address (to)")
	}
	tx := &Transaction{
		From:      from,
		To:        to,
		Symbol:    symbol,
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
		tx.Symbol +
		tx.Value.String() +
		uint64ToString(tx.GasPrice) +
		uint64ToString(tx.GasLimit) +
		uint64ToString(tx.Nonce) +
		string(tx.Data) +
		string(tx.Type)
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}

// Вспомогательная функция для uint64 -> string
func uint64ToString(u uint64) string {
	return new(big.Int).SetUint64(u).String()
}

// HashString возвращает хеш транзакции (для совместимости)
func (tx *Transaction) HashString() string {
	return tx.Hash
}

func DecodeRawTransaction(data []byte) (*Transaction, error) {
	// Реализуйте декодирование транзакции из байтов (например, через gob, json, hex)
	return nil, errors.New("DecodeRawTransaction не реализован")
}

// Сохраняет транзакцию в БД (секционируемая таблица)
func (tx *Transaction) SaveToDB(ctx context.Context, pool *pgxpool.Pool, timestamp time.Time) error {
	_, err := pool.Exec(ctx,
		`INSERT INTO transactions (
			id, hash, sender, recipient, value, fee, nonce, type, payload, status, timestamp
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		tx.ID, tx.Hash, tx.From, tx.To, tx.Value.String(), CalculateTxFee(tx), tx.Nonce, tx.Type, nil, "pending", timestamp)
	return err
}
