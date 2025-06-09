//core/transaction.go

package core

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"GND/core/crypto"

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

// Transaction представляет транзакцию в блокчейне
type Transaction struct {
	ID         int64           `json:"id"`
	BlockID    sql.NullInt64   `json:"block_id"`
	Hash       string          `json:"hash"`
	Sender     string          `json:"sender"`
	Recipient  string          `json:"recipient"`
	Value      *big.Int        `json:"value"`
	Fee        *big.Int        `json:"fee"`
	Nonce      int64           `json:"nonce"`
	Type       string          `json:"type"`
	ContractID sql.NullInt64   `json:"contract_id"`
	Payload    json.RawMessage `json:"payload"`
	Status     string          `json:"status"`
	Timestamp  time.Time       `json:"timestamp"`
	Signature  []byte          `json:"signature"`
	// Для EVM/совместимости
	GasPrice *big.Int `json:"gas_price,omitempty"`
	GasLimit uint64   `json:"gas_limit,omitempty"`
	Data     []byte   `json:"data,omitempty"`
	Symbol   string   `json:"symbol,omitempty"`
}

// Validate проверяет валидность транзакции
func (tx *Transaction) Validate() error {
	if tx.Sender == "" {
		return fmt.Errorf("empty sender address")
	}
	if tx.Recipient == "" {
		return fmt.Errorf("empty recipient address")
	}
	if tx.Value == nil || tx.Value.Sign() < 0 {
		return fmt.Errorf("invalid value")
	}
	if tx.Fee == nil || tx.Fee.Sign() < 0 {
		return fmt.Errorf("invalid fee")
	}
	if tx.Nonce < 0 {
		return fmt.Errorf("invalid nonce")
	}
	return nil
}

// NewTransaction создает новую транзакцию
func NewTransaction(from, to string, value *big.Int, data []byte, nonce uint64, gasLimit uint64, gasPrice *big.Int) (*Transaction, error) {
	fromAddr := Address(from)
	toAddr := Address(to)

	if !fromAddr.IsValid() {
		return nil, fmt.Errorf("invalid sender address")
	}
	if !toAddr.IsValid() {
		return nil, fmt.Errorf("invalid recipient address")
	}

	tx := &Transaction{
		Sender:    fromAddr.String(),
		Recipient: toAddr.String(),
		Value:     value,
		Fee:       new(big.Int).Mul(gasPrice, big.NewInt(int64(gasLimit))),
		Nonce:     int64(nonce),
		Type:      string(TxTypeTransfer),
		Symbol:    "GND",
		Timestamp: time.Now(),
	}

	// Вычисление хеша транзакции
	tx.Hash = tx.CalculateHash()
	return tx, nil
}

// CalculateHash вычисляет хеш транзакции
func (tx *Transaction) CalculateHash() string {
	var sb string
	sb += tx.Sender
	sb += tx.Recipient
	if tx.Value != nil {
		sb += tx.Value.String()
	}
	if tx.Fee != nil {
		sb += tx.Fee.String()
	}
	sb += fmt.Sprintf("%d", tx.Nonce)
	sb += tx.Type
	if tx.ContractID.Valid {
		sb += fmt.Sprintf("%d", tx.ContractID.Int64)
	}
	sb += string(tx.Payload)
	sb += tx.Status
	sb += tx.Timestamp.Format(time.RFC3339Nano)
	if tx.GasPrice != nil {
		sb += tx.GasPrice.String()
	}
	sb += fmt.Sprintf("%d", tx.GasLimit)
	sb += string(tx.Data)
	sb += tx.Symbol

	hashArr := sha256.Sum256([]byte(sb))
	return hex.EncodeToString(hashArr[:])
}

// Sign подписывает транзакцию
func (tx *Transaction) Sign(privateKey string) error {
	// Преобразование приватного ключа
	key, err := crypto.HexToPrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("invalid private key: %v", err)
	}

	// Подпись хеша транзакции
	signature, err := crypto.Sign([]byte(tx.Hash), key)
	if err != nil {
		return fmt.Errorf("failed to sign transaction: %v", err)
	}

	tx.Signature = signature
	return nil
}

// Verify проверяет подпись транзакции
func (tx *Transaction) Verify() bool {
	if tx.Signature == nil {
		return false
	}

	r := new(big.Int).SetBytes(tx.Signature[:len(tx.Signature)/2])
	s := new(big.Int).SetBytes(tx.Signature[len(tx.Signature)/2:])

	// Проверяем подпись
	return crypto.Verify([]byte(tx.Hash), tx.Signature, &ecdsa.PublicKey{
		Curve: crypto.GetCurve(),
		X:     r,
		Y:     s,
	})
}

// SaveToDB сохраняет транзакцию в БД
func (tx *Transaction) SaveToDB(ctx context.Context, pool *pgxpool.Pool) error {
	err := pool.QueryRow(ctx, `
		INSERT INTO transactions (
			block_id, hash, sender, recipient, symbol, value, fee, nonce,
			type, contract_id, payload, status, timestamp, signature
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id`,
		tx.BlockID, tx.Hash, tx.Sender, tx.Recipient, tx.Symbol, tx.Value.String(),
		tx.Fee.String(), tx.Nonce, tx.Type, tx.ContractID, tx.Payload,
		tx.Status, tx.Timestamp, tx.Signature,
	).Scan(&tx.ID)

	if err != nil {
		return fmt.Errorf("ошибка сохранения транзакции: %w", err)
	}

	return nil
}

// CalculateTxFee вычисляет комиссию за транзакцию
func (tx *Transaction) CalculateTxFee(gasUsed uint64) *big.Int {
	fee := new(big.Int).Mul(tx.Fee, big.NewInt(int64(gasUsed)))
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

// Save сохраняет транзакцию в базу данных
func (tx *Transaction) Save(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO transactions (
			id, block_id, hash, sender, recipient, value, fee, nonce,
			type, symbol, contract_id, payload, status, timestamp
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8,
			$9, $10, $11, $12, $13, $14
		)`,
		tx.ID, tx.BlockID, tx.Hash, tx.Sender, tx.Recipient,
		tx.Value.String(), tx.Fee.String(), tx.Nonce,
		tx.Type, tx.Symbol, tx.ContractID, tx.Payload,
		tx.Status, tx.Timestamp,
	)
	return err
}

// UpdateStatus обновляет статус транзакции
func (tx *Transaction) UpdateStatus(ctx context.Context, pool *pgxpool.Pool, status string) error {
	_, err := pool.Exec(ctx, `
		UPDATE transactions 
		SET status = $1 
		WHERE hash = $2`,
		status,
		tx.Hash,
	)
	return err
}

// GetSenderAddress возвращает адрес отправителя как Address
func (tx *Transaction) GetSenderAddress() Address {
	return Address(tx.Sender)
}

// GetRecipientAddress возвращает адрес получателя как Address
func (tx *Transaction) GetRecipientAddress() Address {
	return Address(tx.Recipient)
}
