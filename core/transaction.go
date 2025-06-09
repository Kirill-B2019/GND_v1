//core/transaction.go

package core

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strconv"
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
	From      Address   `json:"from"`
	To        Address   `json:"to"`
	Value     *big.Int  `json:"value"`
	Data      []byte    `json:"data,omitempty"`
	Nonce     uint64    `json:"nonce"`
	GasLimit  uint64    `json:"gasLimit"`
	GasPrice  *big.Int  `json:"gasPrice"`
	Hash      string    `json:"hash"`
	Signature []byte    `json:"signature"`
	Timestamp time.Time `json:"timestamp"`
}

// Validate проверяет валидность транзакции
func (tx *Transaction) Validate() error {
	if tx.From == "" {
		return fmt.Errorf("empty sender address")
	}
	if tx.To == "" {
		return fmt.Errorf("empty recipient address")
	}
	if tx.Value == nil || tx.Value.Sign() < 0 {
		return fmt.Errorf("invalid value")
	}
	if tx.GasLimit == 0 {
		return fmt.Errorf("invalid gas limit")
	}
	if tx.GasPrice == nil || tx.GasPrice.Sign() <= 0 {
		return fmt.Errorf("invalid gas price")
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
		From:      fromAddr,
		To:        toAddr,
		Value:     value,
		Data:      data,
		Nonce:     nonce,
		GasLimit:  gasLimit,
		GasPrice:  gasPrice,
		Timestamp: time.Now(),
	}

	// Вычисление хеша транзакции
	tx.Hash = tx.CalculateHash()
	return tx, nil
}

// CalculateHash вычисляет хеш транзакции
func (tx *Transaction) CalculateHash() string {
	var sb strings.Builder
	sb.WriteString(tx.From.String())
	sb.WriteString(tx.To.String())
	sb.WriteString(tx.Value.String())
	sb.WriteString(tx.GasPrice.String())
	sb.WriteString(strconv.FormatUint(tx.GasLimit, 10))
	sb.WriteString(strconv.FormatUint(tx.Nonce, 10))
	sb.Write(tx.Data)
	sb.WriteString(tx.Timestamp.Format(time.RFC3339))

	hash := sha256.Sum256([]byte(sb.String()))
	return "0x" + hex.EncodeToString(hash[:])
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

	// Восстанавливаем публичный ключ из подписи
	hash := sha256.Sum256([]byte(tx.Hash))
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

// ProcessTransaction обрабатывает транзакцию немедленно (0 подтверждений)
func (b *Blockchain) ProcessTransaction(tx *Transaction) error {
	// Проверка транзакции
	if err := tx.Validate(); err != nil {
		return fmt.Errorf("invalid transaction: %v", err)
	}

	// Применение транзакции к состоянию
	if err := b.State.ApplyTransaction(tx); err != nil {
		return fmt.Errorf("failed to apply transaction: %v", err)
	}

	// Обновление балансов
	senderBalance := b.State.GetBalance(tx.From, "GND")

	if senderBalance.Cmp(tx.Value) < 0 {
		return fmt.Errorf("insufficient balance")
	}

	// Обновление баланса отправителя
	b.State.Credit(tx.From, "GND", new(big.Int).Neg(tx.Value))

	// Обновление баланса получателя
	b.State.Credit(tx.To, "GND", tx.Value)

	return nil
}

// Save сохраняет транзакцию в базу данных
func (tx *Transaction) Save(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO transactions (
			id, hash, sender, recipient, value, nonce, gas_limit, gas_price,
			type, data, status, timestamp
		) VALUES (
			nextval('transactions_id_seq'), $1, $2, $3, $4, $5, $6, $7,
			'transfer', $8, 'pending', $9
		)`,
		tx.Hash,
		tx.From.String(),
		tx.To.String(),
		tx.Value.String(),
		tx.Nonce,
		tx.GasLimit,
		tx.GasPrice.String(),
		tx.Data,
		tx.Timestamp,
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
