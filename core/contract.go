// | KB @CerbeRus - Nexus Invest Team
package core

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Contract представляет смарт-контракт в блокчейне ГАНИМЕД
type Contract struct {
	ID         int       // ID контракта
	Address    string    // Адрес контракта
	Creator    string    // Адрес создателя
	Bytecode   []byte    // Байткод контракта
	ABI        []byte    // ABI контракта
	Type       string    // Тип контракта
	Version    string    // Версия контракта
	Status     string    // Статус контракта
	BlockID    int       // ID блока создания
	TxID       int       // ID транзакции создания
	GasLimit   int64     // Лимит газа
	GasUsed    int64     // Использованный газ
	Value      string    // Значение при создании
	Data       []byte    // Данные инициализации
	CreatedAt  time.Time // Время создания
	UpdatedAt  time.Time // Время последнего обновления
	IsVerified bool      // Проверен ли контракт
	SourceCode string    // Исходный код контракта
	Compiler   string    // Версия компилятора
	Optimized  bool      // Оптимизирован ли код
	Runs       int       // Количество запусков оптимизации
	License    string    // Лицензия контракта
	Metadata   []byte    // Метаданные контракта
}

// ContractParams represents parameters for contract deployment
type ContractParams struct {
	From        string                 `json:"from"`
	Bytecode    string                 `json:"bytecode"`
	Name        string                 `json:"name"`
	Standard    string                 `json:"standard"`
	Owner       string                 `json:"owner"`
	Compiler    string                 `json:"compiler"`
	Version     string                 `json:"version"`
	Params      map[string]interface{} `json:"params"`
	Description string                 `json:"description"`
	MetadataCID string                 `json:"metadata_cid"`
	Metadata    json.RawMessage        `json:"metadata"` // Метаданные контракта (JSON), хранятся на нашей стороне
	SourceCode  string                 `json:"source_code"`
	GasLimit    uint64                 `json:"gas_limit"`
	GasPrice    *big.Int               `json:"gas_price"`
	Nonce       uint64                 `json:"nonce"`
	Signature   string                 `json:"signature"`
	TotalSupply *big.Int               `json:"total_supply"`
}

// NewContract создает новый контракт
func NewContract(address, creator string, bytecode, abi []byte, contractType, version string, blockID, txID int) *Contract {
	now := time.Now()
	return &Contract{
		Address:    address,
		Creator:    creator,
		Bytecode:   bytecode,
		ABI:        abi,
		Type:       contractType,
		Version:    version,
		Status:     "pending",
		BlockID:    blockID,
		TxID:       txID,
		CreatedAt:  now,
		UpdatedAt:  now,
		IsVerified: false,
	}
}

// SaveToDB сохраняет контракт в БД
func (c *Contract) SaveToDB(ctx context.Context, pool *pgxpool.Pool) error {
	err := pool.QueryRow(ctx, `
		INSERT INTO contracts (
			address, creator, bytecode, abi, type,
			version, status, block_id, tx_id, gas_limit,
			gas_used, value, data, created_at, updated_at,
			is_verified, source_code, compiler, optimized,
			runs, license, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22)
		RETURNING id`,
		c.Address, c.Creator, c.Bytecode, c.ABI, c.Type,
		c.Version, c.Status, c.BlockID, c.TxID, c.GasLimit,
		c.GasUsed, c.Value, c.Data, c.CreatedAt, c.UpdatedAt,
		c.IsVerified, c.SourceCode, c.Compiler, c.Optimized,
		c.Runs, c.License, c.Metadata,
	).Scan(&c.ID)

	if err != nil {
		return fmt.Errorf("ошибка сохранения контракта: %w", err)
	}

	return nil
}

// UpdateStatus обновляет статус контракта
func (c *Contract) UpdateStatus(ctx context.Context, pool *pgxpool.Pool, status string) error {
	c.Status = status
	c.UpdatedAt = time.Now()

	_, err := pool.Exec(ctx, `
		UPDATE contracts 
		SET status = $1, updated_at = $2
		WHERE id = $3`,
		c.Status, c.UpdatedAt, c.ID,
	)

	if err != nil {
		return fmt.Errorf("ошибка обновления статуса контракта: %w", err)
	}

	return nil
}

// LoadContract загружает контракт из БД по адресу
func LoadContract(ctx context.Context, pool *pgxpool.Pool, address string) (*Contract, error) {
	var id, blockID, txID int
	var creator, contractType, version, status, value, sourceCode, compiler, license string
	var bytecode, abi, data, metadata []byte
	var gasLimit, gasUsed int64
	var createdAt, updatedAt time.Time
	var isVerified, optimized bool
	var runs int

	err := pool.QueryRow(ctx, `
		SELECT id, creator, bytecode, abi, type,
			version, status, block_id, tx_id, gas_limit,
			gas_used, value, data, created_at, updated_at,
			is_verified, source_code, compiler, optimized,
			runs, license, metadata
		FROM contracts
		WHERE address = $1`,
		address,
	).Scan(&id, &creator, &bytecode, &abi, &contractType,
		&version, &status, &blockID, &txID, &gasLimit,
		&gasUsed, &value, &data, &createdAt, &updatedAt,
		&isVerified, &sourceCode, &compiler, &optimized,
		&runs, &license, &metadata)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("контракт не найден: %s", address)
	}
	if err != nil {
		return nil, fmt.Errorf("ошибка загрузки контракта: %w", err)
	}

	return &Contract{
		ID:         id,
		Address:    address,
		Creator:    creator,
		Bytecode:   bytecode,
		ABI:        abi,
		Type:       contractType,
		Version:    version,
		Status:     status,
		BlockID:    blockID,
		TxID:       txID,
		GasLimit:   gasLimit,
		GasUsed:    gasUsed,
		Value:      value,
		Data:       data,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
		IsVerified: isVerified,
		SourceCode: sourceCode,
		Compiler:   compiler,
		Optimized:  optimized,
		Runs:       runs,
		License:    license,
		Metadata:   metadata,
	}, nil
}

// GetContractByAddress возвращает контракт по адресу
func GetContractByAddress(ctx context.Context, pool *pgxpool.Pool, address string) (*Contract, error) {
	return LoadContract(ctx, pool, address)
}

// GetContractsByCreator возвращает все контракты, созданные указанным адресом
func GetContractsByCreator(ctx context.Context, pool *pgxpool.Pool, creator string) ([]*Contract, error) {
	rows, err := pool.Query(ctx, `
		SELECT address
		FROM contracts
		WHERE creator = $1
		ORDER BY created_at DESC`,
		creator,
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения контрактов создателя: %w", err)
	}
	defer rows.Close()

	var contracts []*Contract
	for rows.Next() {
		var address string
		if err := rows.Scan(&address); err != nil {
			return nil, fmt.Errorf("ошибка сканирования адреса контракта: %w", err)
		}

		contract, err := LoadContract(ctx, pool, address)
		if err != nil {
			return nil, fmt.Errorf("ошибка загрузки контракта %s: %w", address, err)
		}

		contracts = append(contracts, contract)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при чтении контрактов: %w", err)
	}

	return contracts, nil
}
