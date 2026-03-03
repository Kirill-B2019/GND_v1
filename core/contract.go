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
	ID          int       // ID контракта
	Address     string    // Адрес контракта
	Creator     string    // Адрес создателя
	Name        string    // Название (для чтения в GND_admin / GetContractState)
	Symbol      string    // Символ (стандарт или тикер)
	Owner       string    // Владелец (для GetContractState и отображения)
	Code        []byte    // Код контракта (BYTEA, при деплое записывается bytecode)
	Bytecode    []byte    // Байткод контракта
	ABI         []byte    // ABI контракта
	Type        string    // Тип контракта
	Standard    string    // Стандарт (GND-st1 и т.д.)
	Description string    // Описание контракта
	Version     string    // Версия контракта
	Status      string    // Статус контракта
	BlockID     int       // ID блока создания
	TxID        int       // ID транзакции создания
	GasLimit    int64     // Лимит газа
	GasUsed     int64     // Использованный газ
	Value       string    // Значение при создании
	Data        []byte    // Данные инициализации
	CreatedAt   time.Time // Время создания
	UpdatedAt   time.Time // Время последнего обновления
	IsVerified  bool      // Проверен ли контракт
	SourceCode  string    // Исходный код контракта
	Compiler    string    // Версия компилятора
	Optimized   bool      // Оптимизирован ли код
	Runs        int       // Количество запусков оптимизации
	License     string    // Лицензия контракта
	Metadata    []byte    // Метаданные контракта (JSONB в БД)
	Params      []byte    // Параметры деплоя (JSONB в БД)
	MetadataCID string    // CID метаданных (IPFS и т.п.)
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

// SaveToDB сохраняет контракт в БД (в т.ч. code, name, symbol, owner, standard, description, params, metadata_cid).
func (c *Contract) SaveToDB(ctx context.Context, pool *pgxpool.Pool) error {
	codeToSave := c.Code
	if len(codeToSave) == 0 && len(c.Bytecode) > 0 {
		codeToSave = c.Bytecode
	}
	var metadataArg, paramsArg interface{}
	if len(c.Metadata) > 0 {
		if json.Valid(c.Metadata) {
			metadataArg = json.RawMessage(c.Metadata)
		} else {
			metadataArg = map[string]string{"v": string(c.Metadata)}
		}
	}
	if len(c.Params) > 0 {
		if json.Valid(c.Params) {
			paramsArg = json.RawMessage(c.Params)
		} else {
			paramsArg = map[string]string{"v": string(c.Params)}
		}
	}
	err := pool.QueryRow(ctx, `
		INSERT INTO contracts (
			address, creator, name, symbol, owner, code, bytecode, abi, type,
			standard, description, version, status, block_id, tx_id, gas_limit,
			gas_used, value, data, created_at, updated_at,
			is_verified, source_code, compiler, optimized,
			runs, license, metadata, params, metadata_cid
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16,
			$17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30)
		RETURNING id`,
		c.Address, c.Creator, c.Name, c.Symbol, c.Owner,
		codeToSave, c.Bytecode, c.ABI, c.Type,
		c.Standard, c.Description, c.Version, c.Status, c.BlockID, c.TxID, c.GasLimit,
		c.GasUsed, c.Value, c.Data, c.CreatedAt, c.UpdatedAt,
		c.IsVerified, c.SourceCode, c.Compiler, c.Optimized,
		c.Runs, c.License, metadataArg, paramsArg, nullIfEmpty(c.MetadataCID),
	).Scan(&c.ID)

	if err != nil {
		return fmt.Errorf("ошибка сохранения контракта: %w", err)
	}

	return nil
}

func nullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
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

// LoadContract загружает контракт из БД по адресу (включая name, symbol, owner, code, standard, description, params, metadata_cid).
func LoadContract(ctx context.Context, pool *pgxpool.Pool, address string) (*Contract, error) {
	var id, blockID, txID int
	var creator, name, symbol, owner, contractType, standard, description, version, status, value, sourceCode, compiler, license, metadataCID string
	var code, bytecode, abi, data, metadata, params []byte
	var gasLimit, gasUsed int64
	var createdAt, updatedAt time.Time
	var isVerified, optimized bool
	var runs int

	query := `
		SELECT id, creator, name, symbol, owner, code, bytecode, abi, type,
			standard, description, version, status, block_id, tx_id, gas_limit,
			gas_used, value, data, created_at, updated_at,
			is_verified, source_code, compiler, optimized,
			runs, license, metadata, params, metadata_cid
		FROM contracts
		WHERE address = $1`
	err := pool.QueryRow(ctx, query, address).Scan(
		&id, &creator, &name, &symbol, &owner, &code, &bytecode, &abi, &contractType,
		&standard, &description, &version, &status, &blockID, &txID, &gasLimit,
		&gasUsed, &value, &data, &createdAt, &updatedAt,
		&isVerified, &sourceCode, &compiler, &optimized,
		&runs, &license, &metadata, &params, &metadataCID,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("контракт не найден: %s", address)
	}
	if err != nil {
		return nil, fmt.Errorf("ошибка загрузки контракта: %w", err)
	}

	return &Contract{
		ID:          id,
		Address:     address,
		Creator:     creator,
		Name:        name,
		Symbol:      symbol,
		Owner:       owner,
		Code:        code,
		Bytecode:    bytecode,
		ABI:         abi,
		Type:        contractType,
		Standard:    standard,
		Description: description,
		Version:     version,
		Status:      status,
		BlockID:     blockID,
		TxID:        txID,
		GasLimit:    gasLimit,
		GasUsed:     gasUsed,
		Value:       value,
		Data:        data,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		IsVerified:  isVerified,
		SourceCode:  sourceCode,
		Compiler:    compiler,
		Optimized:   optimized,
		Runs:        runs,
		License:     license,
		Metadata:    metadata,
		Params:      params,
		MetadataCID: metadataCID,
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
