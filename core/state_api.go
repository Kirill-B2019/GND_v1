// state_api.go — чтение и запись состояний аккаунтов и слотов storage контрактов через API (для GND_admin и клиентов).

package core

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/jackc/pgx/v5/pgxpool"
)

// AccountStateAtBlock — снимок состояния аккаунта на блок (из account_states).
type AccountStateAtBlock struct {
	BlockID     int64  `json:"block_id"`
	Address     string `json:"address"`
	Nonce       uint64 `json:"nonce"`
	BalanceGND  string `json:"balance_gnd"`
	StorageRoot string `json:"storage_root,omitempty"` // hex
}

// ContractStorageSlot — слот storage контракта (ключ и значение по 32 байта).
type ContractStorageSlot struct {
	SlotKey   string `json:"slot_key"`   // hex
	SlotValue string `json:"slot_value"` // hex
}

// GetAccountStateAtBlock возвращает снимок состояния аккаунта на блок из account_states.
func GetAccountStateAtBlock(ctx context.Context, pool *pgxpool.Pool, address string, blockID int64) (*AccountStateAtBlock, error) {
	if pool == nil {
		return nil, fmt.Errorf("pool is nil")
	}
	var nonce int64
	var balanceGnd string
	var storageRoot []byte
	err := pool.QueryRow(ctx, `
		SELECT nonce, balance_gnd, storage_root
		FROM account_states
		WHERE address = $1 AND block_id = $2`,
		address, blockID,
	).Scan(&nonce, &balanceGnd, &storageRoot)
	if err != nil {
		return nil, err
	}
	out := &AccountStateAtBlock{
		BlockID:    blockID,
		Address:    address,
		Nonce:      uint64(nonce),
		BalanceGND: balanceGnd,
	}
	if len(storageRoot) > 0 {
		out.StorageRoot = "0x" + hex.EncodeToString(storageRoot)
	}
	return out, nil
}

// GetContractStorageAtBlock возвращает все слоты storage контракта на блок из contract_storage.
func GetContractStorageAtBlock(ctx context.Context, pool *pgxpool.Pool, address string, blockID int64) ([]ContractStorageSlot, error) {
	if pool == nil {
		return nil, fmt.Errorf("pool is nil")
	}
	rows, err := pool.Query(ctx, `
		SELECT slot_key, slot_value
		FROM contract_storage
		WHERE address = $1 AND block_id = $2`,
		address, blockID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var slots []ContractStorageSlot
	for rows.Next() {
		var key, value []byte
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		slots = append(slots, ContractStorageSlot{
			SlotKey:   "0x" + hex.EncodeToString(key),
			SlotValue: "0x" + hex.EncodeToString(value),
		})
	}
	return slots, rows.Err()
}

// GetContractStorageLatest возвращает актуальное состояние storage контракта на момент последнего блока цепи.
// Для каждого slot_key берётся значение из строки с максимальным block_id (не больше текущей высоты).
func GetContractStorageLatest(ctx context.Context, pool *pgxpool.Pool, address string) ([]ContractStorageSlot, error) {
	if pool == nil {
		return nil, fmt.Errorf("pool is nil")
	}
	rows, err := pool.Query(ctx, `
		SELECT DISTINCT ON (slot_key) slot_key, slot_value
		FROM contract_storage
		WHERE address = $1 AND block_id <= (SELECT COALESCE(MAX(id), 0) FROM blocks)
		ORDER BY slot_key, block_id DESC`,
		address,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var slots []ContractStorageSlot
	for rows.Next() {
		var key, value []byte
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		slots = append(slots, ContractStorageSlot{
			SlotKey:   "0x" + hex.EncodeToString(key),
			SlotValue: "0x" + hex.EncodeToString(value),
		})
	}
	return slots, rows.Err()
}

// GetCurrentAccountState возвращает текущее состояние аккаунта из таблицы accounts.
func GetCurrentAccountState(ctx context.Context, pool *pgxpool.Pool, address string) (*AccountStateAtBlock, error) {
	if pool == nil {
		return nil, fmt.Errorf("pool is nil")
	}
	var nonce int64
	var balanceGnd string
	err := pool.QueryRow(ctx, `
		SELECT nonce, COALESCE(balance_gnd::text, '0')
		FROM accounts
		WHERE address = $1`,
		address,
	).Scan(&nonce, &balanceGnd)
	if err != nil {
		return nil, err
	}
	return &AccountStateAtBlock{
		BlockID:    0,
		Address:    address,
		Nonce:      uint64(nonce),
		BalanceGND: balanceGnd,
	}, nil
}

// WriteContractStorageSlot записывает один слот storage контракта для указанного блока (админ/импорт).
// slotKeyHex и slotValueHex — hex-строки (с 0x или без), по 32 байта после декодирования.
func WriteContractStorageSlot(ctx context.Context, pool *pgxpool.Pool, blockID int64, address, slotKeyHex, slotValueHex string) error {
	if pool == nil {
		return fmt.Errorf("pool is nil")
	}
	key := slotKeyHex
	if len(key) >= 2 && key[0:2] == "0x" {
		key = key[2:]
	}
	value := slotValueHex
	if len(value) >= 2 && value[0:2] == "0x" {
		value = value[2:]
	}
	keyBytes, err := hex.DecodeString(key)
	if err != nil {
		return fmt.Errorf("slot_key hex: %w", err)
	}
	valueBytes, err := hex.DecodeString(value)
	if err != nil {
		return fmt.Errorf("slot_value hex: %w", err)
	}
	if len(keyBytes) != 32 {
		return fmt.Errorf("slot_key must be 32 bytes, got %d", len(keyBytes))
	}
	if len(valueBytes) != 32 {
		return fmt.Errorf("slot_value must be 32 bytes, got %d", len(valueBytes))
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO contract_storage (block_id, address, slot_key, slot_value)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (block_id, address, slot_key) DO UPDATE SET slot_value = $4`,
		blockID, address, keyBytes, valueBytes,
	)
	return err
}

// WriteInitialStorageForDeployedContract записывает начальные слоты storage для только что задеплоенного контракта (GNDToken/GANIToken-подобный).
// Параметр initialSupply из params пишется в слот 0 (_totalSupply), чтобы totalSupply() при чтении через CallStatic возвращал верное значение.
func WriteInitialStorageForDeployedContract(ctx context.Context, pool *pgxpool.Pool, blockID int64, contractAddress string, params map[string]interface{}) error {
	if pool == nil || blockID <= 0 || contractAddress == "" {
		return nil
	}
	var initialSupply *big.Int
	if v, ok := params["initialSupply"]; ok && v != nil {
		switch val := v.(type) {
		case string:
			initialSupply = new(big.Int)
			if _, ok := initialSupply.SetString(val, 10); !ok {
				return nil
			}
		case float64:
			initialSupply = big.NewInt(int64(val))
		default:
			return nil
		}
	}
	if initialSupply == nil || initialSupply.Sign() <= 0 {
		return nil
	}
	// Слот 0 = _totalSupply (32 байта, big-endian ABI)
	slot0Key := "0x" + hex.EncodeToString(make([]byte, 32))
	slot0Value := "0x" + hex.EncodeToString(abiUint256(initialSupply))
	return WriteContractStorageSlot(ctx, pool, blockID, contractAddress, slot0Key, slot0Value)
}

// abiUint256 кодирует *big.Int в 32 байта big-endian (ABI uint256).
func abiUint256(n *big.Int) []byte {
	b := n.Bytes()
	if len(b) > 32 {
		return b[len(b)-32:]
	}
	out := make([]byte, 32)
	copy(out[32-len(b):], b)
	return out
}
