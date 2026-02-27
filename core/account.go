package core

import (
	"context"
	"database/sql"
	"fmt"
	"math/big"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Account представляет аккаунт в блокчейне ГАНИМЕД
type Account struct {
	ID         int       // ID аккаунта
	Address    string    // Адрес аккаунта
	Balance    string    // Баланс в нативных токенах
	Nonce      uint64    // Номер последней транзакции
	Type       string    // Тип аккаунта (user, contract, system)
	Status     string    // Статус аккаунта
	BlockID    int       // ID блока создания
	TxID       int       // ID транзакции создания
	GasLimit   int64     // Лимит газа
	GasUsed    int64     // Использованный газ
	Value      string    // Значение при создании
	Data       []byte    // Данные аккаунта
	CreatedAt  time.Time // Время создания
	UpdatedAt  time.Time // Время последнего обновления
	IsVerified bool      // Проверен ли аккаунт
	SourceCode string    // Исходный код (для контрактов)
	Compiler   string    // Версия компилятора
	Optimized  bool      // Оптимизирован ли код
	Runs       int       // Количество запусков оптимизации
	License    string    // Лицензия
	Metadata   []byte    // Метаданные аккаунта
}

// NewAccount создает новый аккаунт
func NewAccount(address string, accountType string, blockID, txID int) *Account {
	now := time.Now()
	return &Account{
		Address:    address,
		Balance:    "0",
		Nonce:      0,
		Type:       accountType,
		Status:     "active",
		BlockID:    blockID,
		TxID:       txID,
		CreatedAt:  now,
		UpdatedAt:  now,
		IsVerified: false,
	}
}

// SaveToDB сохраняет аккаунт в БД
func (a *Account) SaveToDB(ctx context.Context, pool *pgxpool.Pool) error {
	err := pool.QueryRow(ctx, `
		INSERT INTO accounts (
			address, balance, nonce, type, status,
			block_id, tx_id, gas_limit, gas_used, value,
			data, created_at, updated_at, is_verified,
			source_code, compiler, optimized, runs,
			license, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, $17, $18, $19, $20)
		RETURNING id`,
		a.Address, a.Balance, a.Nonce, a.Type, a.Status,
		a.BlockID, a.TxID, a.GasLimit, a.GasUsed, a.Value,
		a.Data, a.CreatedAt, a.UpdatedAt, a.IsVerified,
		a.SourceCode, a.Compiler, a.Optimized, a.Runs,
		a.License, a.Metadata,
	).Scan(&a.ID)

	if err != nil {
		return fmt.Errorf("ошибка сохранения аккаунта: %w", err)
	}

	return nil
}

// UpdateBalance обновляет баланс аккаунта только в памяти; в БД accounts.balance не пишем (по требованию).
func (a *Account) UpdateBalance(ctx context.Context, pool *pgxpool.Pool, newBalance string) error {
	a.Balance = newBalance
	a.UpdatedAt = time.Now()
	// Не обновляем accounts.balance в БД — баланс хранится в token_balances
	return nil
}

// IncrementNonce увеличивает nonce аккаунта
func (a *Account) IncrementNonce(ctx context.Context, pool *pgxpool.Pool) error {
	a.Nonce++
	a.UpdatedAt = time.Now()

	_, err := pool.Exec(ctx, `
		UPDATE accounts 
		SET nonce = $1, updated_at = $2
		WHERE id = $3`,
		a.Nonce, a.UpdatedAt, a.ID,
	)

	if err != nil {
		return fmt.Errorf("ошибка обновления nonce: %w", err)
	}

	return nil
}

// LoadAccount загружает аккаунт из БД по адресу
func LoadAccount(ctx context.Context, pool *pgxpool.Pool, address string) (*Account, error) {
	var id, blockID, txID int
	var balance, accountType, status, value, sourceCode, compiler, license string
	var nonce uint64
	var gasLimit, gasUsed int64
	var data, metadata []byte
	var createdAt, updatedAt time.Time
	var isVerified, optimized bool
	var runs int

	err := pool.QueryRow(ctx, `
		SELECT id, balance, nonce, type, status,
			block_id, tx_id, gas_limit, gas_used, value,
			data, created_at, updated_at, is_verified,
			source_code, compiler, optimized, runs,
			license, metadata
		FROM accounts
		WHERE address = $1`,
		address,
	).Scan(&id, &balance, &nonce, &accountType, &status,
		&blockID, &txID, &gasLimit, &gasUsed, &value,
		&data, &createdAt, &updatedAt, &isVerified,
		&sourceCode, &compiler, &optimized, &runs,
		&license, &metadata)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("аккаунт не найден: %s", address)
	}
	if err != nil {
		return nil, fmt.Errorf("ошибка загрузки аккаунта: %w", err)
	}

	return &Account{
		ID:         id,
		Address:    address,
		Balance:    balance,
		Nonce:      nonce,
		Type:       accountType,
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

// GetAccountByAddress возвращает аккаунт по адресу
func GetAccountByAddress(ctx context.Context, pool *pgxpool.Pool, address string) (*Account, error) {
	return LoadAccount(ctx, pool, address)
}

// GetAccountBalance возвращает баланс аккаунта
func GetAccountBalance(ctx context.Context, pool *pgxpool.Pool, address string) (*big.Int, error) {
	var balanceStr string
	err := pool.QueryRow(ctx, `
		SELECT balance
		FROM accounts
		WHERE address = $1`,
		address,
	).Scan(&balanceStr)

	if err == sql.ErrNoRows {
		return big.NewInt(0), nil
	}
	if err != nil {
		return nil, fmt.Errorf("ошибка получения баланса: %w", err)
	}

	balance, ok := new(big.Int).SetString(balanceStr, 10)
	if !ok {
		return nil, fmt.Errorf("ошибка парсинга баланса: %s", balanceStr)
	}

	return balance, nil
}

// GetAccountNonce возвращает nonce аккаунта
func GetAccountNonce(ctx context.Context, pool *pgxpool.Pool, address string) (uint64, error) {
	var nonce uint64
	err := pool.QueryRow(ctx, `
		SELECT nonce
		FROM accounts
		WHERE address = $1`,
		address,
	).Scan(&nonce)

	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("ошибка получения nonce: %w", err)
	}

	return nonce, nil
}
