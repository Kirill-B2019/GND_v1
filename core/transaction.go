//core/transaction.go

package core

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"crypto/elliptic"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TxType определяет тип транзакции
type TxType string

const (
	TxTypeTransfer     TxType = "transfer"
	TxTypeContract     TxType = "contract"
	TxTypeDeploy       TxType = "deploy"
	TxTypeStake        TxType = "stake"
	TxTypeUnstake      TxType = "unstake"
	TxTypeValidator    TxType = "validator"
	TxTypeToken        TxType = "token"
	TxTypeTokenMint    TxType = "token_mint"
	TxTypeTokenBurn    TxType = "token_burn"
	TxTypeTokenPause   TxType = "token_pause"
	TxTypeTokenUnpause TxType = "token_unpause"
)

// Transaction - структура транзакции в блокчейне ГАНИМЕД
type Transaction struct {
	ID         int       // ID транзакции
	BlockID    *int      // ID блока (nil если транзакция в мемпуле)
	Hash       string    // Хеш транзакции
	From       string    // Адрес отправителя
	To         string    // Адрес получателя
	Symbol     string    // Символ токена
	Value      *big.Int  // Количество токенов
	GasPrice   *big.Int  // Цена за единицу газа
	GasLimit   uint64    // Лимит газа
	Nonce      uint64    // Номер транзакции отправителя
	Type       string    // Тип транзакции (transfer, contract_deploy, contract_call)
	ContractID *int      // ID контракта (для вызовов контракта)
	Data       []byte    // Дополнительные данные
	Status     string    // Статус транзакции (pending, confirmed, failed)
	Timestamp  time.Time // Время создания транзакции
	Signature  []byte    // Подпись транзакции
}

// NewTransaction создает новую транзакцию
func NewTransaction(from, to, symbol string, value *big.Int, gasPrice *big.Int, gasLimit uint64, nonce uint64, txType string, data []byte) *Transaction {
	tx := &Transaction{
		From:      from,
		To:        to,
		Symbol:    symbol,
		Value:     value,
		GasPrice:  gasPrice,
		GasLimit:  gasLimit,
		Nonce:     nonce,
		Type:      txType,
		Data:      data,
		Status:    "pending",
		Timestamp: time.Now(),
	}
	tx.Hash = tx.CalculateHash()
	return tx
}

// CalculateHash вычисляет хеш транзакции
func (tx *Transaction) CalculateHash() string {
	var sb strings.Builder
	sb.WriteString(tx.From)
	sb.WriteString(tx.To)
	sb.WriteString(tx.Symbol)
	sb.WriteString(tx.Value.String())
	sb.WriteString(tx.GasPrice.String())
	sb.WriteString(strconv.FormatUint(tx.GasLimit, 10))
	sb.WriteString(strconv.FormatUint(tx.Nonce, 10))
	sb.WriteString(tx.Type)
	if tx.ContractID != nil {
		sb.WriteString(strconv.Itoa(*tx.ContractID))
	}
	sb.Write(tx.Data)
	sb.WriteString(tx.Timestamp.Format(time.RFC3339))
	hash := sha256.Sum256([]byte(sb.String()))
	return hex.EncodeToString(hash[:])
}

// Sign подписывает транзакцию
func (tx *Transaction) Sign(privateKey *ecdsa.PrivateKey) error {
	hash := tx.CalculateHash()
	hashBytes, err := hex.DecodeString(hash)
	if err != nil {
		return fmt.Errorf("ошибка декодирования хеша: %w", err)
	}

	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hashBytes)
	if err != nil {
		return fmt.Errorf("ошибка подписи: %w", err)
	}

	signature := append(r.Bytes(), s.Bytes()...)
	tx.Signature = signature
	return nil
}

// Verify проверяет подпись транзакции
func (tx *Transaction) Verify() (bool, error) {
	if len(tx.Signature) != 64 {
		return false, fmt.Errorf("неверная длина подписи")
	}

	r := new(big.Int).SetBytes(tx.Signature[:32])
	s := new(big.Int).SetBytes(tx.Signature[32:])

	hash := tx.CalculateHash()
	hashBytes, err := hex.DecodeString(hash)
	if err != nil {
		return false, fmt.Errorf("ошибка декодирования хеша: %w", err)
	}

	publicKey, err := RecoverPublicKey(tx.From)
	if err != nil {
		return false, fmt.Errorf("ошибка восстановления публичного ключа: %w", err)
	}

	return ecdsa.Verify(publicKey, hashBytes, r, s), nil
}

// SaveToDB сохраняет транзакцию в БД
func (tx *Transaction) SaveToDB(ctx context.Context, pool *pgxpool.Pool) error {
	err := pool.QueryRow(ctx, `
		INSERT INTO transactions (
			block_id, hash, from_address, to_address, symbol, value,
			gas_price, gas_limit, nonce, type, contract_id, data,
			status, timestamp, signature
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id`,
		tx.BlockID, tx.Hash, tx.From, tx.To, tx.Symbol, tx.Value.String(),
		tx.GasPrice.String(), tx.GasLimit, tx.Nonce, tx.Type, tx.ContractID, tx.Data,
		tx.Status, tx.Timestamp, tx.Signature,
	).Scan(&tx.ID)

	if err != nil {
		return fmt.Errorf("ошибка сохранения транзакции: %w", err)
	}

	return nil
}

// CalculateTxFee вычисляет комиссию за транзакцию
func (tx *Transaction) CalculateTxFee(gasUsed uint64) *big.Int {
	fee := new(big.Int).Mul(tx.GasPrice, big.NewInt(int64(gasUsed)))
	return fee
}

func DecodeRawTransaction(data []byte) (*Transaction, error) {
	// Реализуйте декодирование транзакции из байтов (например, через gob, json, hex)
	return nil, errors.New("DecodeRawTransaction не реализован")
}

// RecoverPublicKey восстанавливает публичный ключ из адреса
func RecoverPublicKey(address string) (*ecdsa.PublicKey, error) {
	// Удаляем префикс адреса (GND_, GN_ и т.д.)
	addr := strings.TrimPrefix(address, "GND_")
	addr = strings.TrimPrefix(addr, "GN_")
	addr = strings.TrimPrefix(addr, "GND")
	addr = strings.TrimPrefix(addr, "GN")

	// Декодируем адрес из hex
	addrBytes, err := hex.DecodeString(addr)
	if err != nil {
		return nil, fmt.Errorf("ошибка декодирования адреса: %w", err)
	}

	// Создаем публичный ключ
	publicKey := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     new(big.Int).SetBytes(addrBytes[:32]),
		Y:     new(big.Int).SetBytes(addrBytes[32:]),
	}

	return publicKey, nil
}
