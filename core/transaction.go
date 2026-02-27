// | KB @CerbeRus - Nexus Invest Team
//core/transaction.go

package core

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"GND/core/crypto"
	"GND/types"

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

// Transaction represents a blockchain transaction
type Transaction struct {
	ID         string        `json:"id"`
	Sender     types.Address `json:"sender"`
	Recipient  types.Address `json:"recipient"`
	Value      *big.Int      `json:"value"`
	Data       []byte        `json:"data"`
	Nonce      int64         `json:"nonce"`
	GasLimit   uint64        `json:"gas_limit"`
	GasPrice   *big.Int      `json:"gas_price"`
	Signature  []byte        `json:"signature"`
	Hash       string        `json:"hash"`
	Fee        *big.Int      `json:"fee"`
	Type       string        `json:"type"`
	Status     string        `json:"status"`
	Timestamp  time.Time     `json:"timestamp"`
	BlockID    int           `json:"block_id"`
	ContractID sql.NullInt64 `json:"contract_id"`
	Payload    []byte        `json:"payload"`
	Symbol     string        `json:"symbol"`
	IsVerified bool          `json:"is_verified"` // true для транзакций с нативной монетой (GND)
}

// Validate checks if the transaction is valid
func (tx *Transaction) Validate() error {
	if !tx.Sender.IsValid() {
		return errors.New("invalid sender address")
	}
	if !tx.Recipient.IsValid() {
		return errors.New("invalid recipient address")
	}
	if tx.Value == nil || tx.Value.Sign() < 0 {
		return errors.New("invalid value")
	}
	if tx.GasLimit == 0 {
		return errors.New("gas limit must be greater than 0")
	}
	if tx.GasPrice == nil || tx.GasPrice.Sign() <= 0 {
		return errors.New("invalid gas price")
	}
	return nil
}

// HasSufficientBalance checks if the sender has enough balance for the transaction
func (tx *Transaction) HasSufficientBalance() bool {
	state := GetState()
	if state == nil {
		return false
	}

	requiredBalance := new(big.Int).Mul(tx.Value, big.NewInt(1))
	gasCost := new(big.Int).Mul(tx.GasPrice, big.NewInt(int64(tx.GasLimit)))
	requiredBalance.Add(requiredBalance, gasCost)

	balance := state.GetBalance(tx.Sender, "GND")
	return balance.Cmp(requiredBalance) >= 0
}

// IsContractCall checks if the transaction is a contract call
func (tx *Transaction) IsContractCall() bool {
	return len(tx.Data) > 0
}

// GetTotalCost calculates the total cost of the transaction including gas
func (tx *Transaction) GetTotalCost() *big.Int {
	total := new(big.Int).Set(tx.Value)
	gasCost := new(big.Int).Mul(tx.GasPrice, big.NewInt(int64(tx.GasLimit)))
	return total.Add(total, gasCost)
}

// NewTransaction creates a new transaction
func NewTransaction(sender, recipient types.Address, value *big.Int, data []byte, nonce uint64, gasLimit uint64, gasPrice *big.Int) (*Transaction, error) {
	if !sender.IsValid() {
		return nil, errors.New("invalid sender address")
	}
	if !recipient.IsValid() {
		return nil, errors.New("invalid recipient address")
	}

	return &Transaction{
		ID:        generateTransactionID(),
		Sender:    sender,
		Recipient: recipient,
		Value:     value,
		Data:      data,
		Nonce:     int64(nonce),
		GasLimit:  gasLimit,
		GasPrice:  gasPrice,
	}, nil
}

// generateTransactionID generates a unique transaction ID
func generateTransactionID() string {
	// TODO: Implement proper ID generation
	return "tx_" + time.Now().Format("20060102150405")
}

// CalculateHash вычисляет хеш транзакции
func (tx *Transaction) CalculateHash() string {
	var sb string
	sb += tx.Sender.String()
	sb += tx.Recipient.String()
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

// SaveToDB сохраняет транзакцию в БД (id берётся из nextval(transactions_id_seq), contract_id — NULL если не задан).
func (tx *Transaction) SaveToDB(ctx context.Context, pool *pgxpool.Pool) error {
	var contractID *int64
	if tx.ContractID.Valid {
		contractID = &tx.ContractID.Int64
	}
	feeStr := "0"
	if tx.Fee != nil {
		feeStr = tx.Fee.String()
	}
	var dbID int64
	err := pool.QueryRow(ctx, `
		INSERT INTO transactions (
			id, block_id, hash, sender, recipient, value, fee, nonce,
			type, contract_id, payload, status, timestamp, signature, is_verified
		) VALUES (nextval('transactions_id_seq'), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id`,
		tx.BlockID, tx.Hash, tx.Sender.String(), tx.Recipient.String(), tx.Value.String(),
		feeStr, tx.Nonce, tx.Type, contractID, tx.Payload,
		tx.Status, tx.Timestamp, tx.Signature, tx.IsVerified,
	).Scan(&dbID)
	if err == nil {
		tx.ID = strconv.FormatInt(dbID, 10)
	}

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
			type, contract_id, payload, status, timestamp, signature, is_verified
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8,
			$9, $10, $11, $12, $13, $14, $15
		)`,
		tx.ID, tx.BlockID, tx.Hash, tx.Sender.String(), tx.Recipient.String(),
		tx.Value.String(), tx.Fee.String(), tx.Nonce,
		tx.Type, tx.ContractID, tx.Payload,
		tx.Status, tx.Timestamp, tx.Signature, tx.IsVerified,
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
func (tx *Transaction) GetSenderAddress() types.Address {
	return tx.Sender
}

// GetRecipientAddress возвращает адрес получателя как Address
func (tx *Transaction) GetRecipientAddress() types.Address {
	return tx.Recipient
}

// GetState возвращает глобальное состояние блокчейна
func GetState() *State {
	stateMutex.RLock()
	defer stateMutex.RUnlock()
	return globalState
}

// SetState устанавливает глобальное состояние блокчейна
func SetState(state *State) {
	stateMutex.Lock()
	defer stateMutex.Unlock()
	globalState = state
}
