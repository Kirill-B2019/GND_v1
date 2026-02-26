package core

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"time"

	"GND/tokens/registry"
	"GND/tokens/standards/gndst1"
	"GND/types"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Token представляет токен в блокчейне ГАНИМЕД
type Token struct {
	ID          int       // ID токена
	Address     string    // Адрес контракта токена
	Symbol      string    // Символ токена
	Name        string    // Название токена
	Decimals    int       // Количество десятичных знаков
	TotalSupply string    // Общее предложение
	Owner       string    // Владелец токена
	Type        string    // Тип токена (gndst-1)
	Standard    string    // Стандарт токена
	Status      string    // Статус токена
	BlockID     int       // ID блока создания
	TxID        int       // ID транзакции создания
	GasLimit    int64     // Лимит газа
	GasUsed     int64     // Использованный газ
	Value       string    // Значение при создании
	Data        []byte    // Данные инициализации
	CreatedAt   time.Time // Время создания
	UpdatedAt   time.Time // Время последнего обновления
	IsVerified  bool      // Проверен ли токен
	SourceCode  string    // Исходный код контракта
	Compiler    string    // Версия компилятора
	Optimized   bool      // Оптимизирован ли код
	Runs        int       // Количество запусков оптимизации
	License     string    // Лицензия контракта
	Metadata    []byte    // Метаданные токена
}

// ToTokenInfo преобразует Token в TokenInfo
func (t *Token) ToTokenInfo() *types.TokenInfo {
	return &types.TokenInfo{
		Address:     t.Address,
		Name:        t.Name,
		Symbol:      t.Symbol,
		Decimals:    uint8(t.Decimals),
		TotalSupply: t.TotalSupply,
		Standard:    t.Standard,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

// NewToken создает новый токен
func NewToken(address, symbol, name string, decimals int, totalSupply string, owner, tokenType, standard string, blockID, txID int) *Token {
	now := time.Now()

	// Если адрес не указан, генерируем новый
	if address == "" {
		address = fmt.Sprintf("GNDct%s", GenerateContractAddress())
		fmt.Printf("[DEBUG] NewToken: сгенерирован новый адрес контракта: %s\n", address)
	}

	return &Token{
		Address:     address,
		Symbol:      symbol,
		Name:        name,
		Decimals:    decimals,
		TotalSupply: totalSupply,
		Owner:       owner,
		Type:        tokenType,
		Standard:    standard,
		Status:      "pending",
		BlockID:     blockID,
		TxID:        txID,
		CreatedAt:   now,
		UpdatedAt:   now,
		IsVerified:  false,
	}
}

// SaveToDB сохраняет токен в БД
func (t *Token) SaveToDB(ctx context.Context, pool *pgxpool.Pool) error {
	// Сначала получаем contract_id из таблицы contracts
	fmt.Printf("[DEBUG] SaveToDB: ищем contract_id по адресу: %s\n", t.Address)
	var contractID int
	err := pool.QueryRow(ctx, `
		SELECT id FROM contracts WHERE address = $1`,
		t.Address,
	).Scan(&contractID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Контракт не существует - создаем его
			fmt.Printf("[DEBUG] SaveToDB: создаем новый контракт для адреса: %s\n", t.Address)
			err = pool.QueryRow(ctx, `
				INSERT INTO contracts (address, owner, type, created_at)
				VALUES ($1, $2, 'token', NOW())
				RETURNING id`,
				t.Address, t.Owner,
			).Scan(&contractID)
			if err != nil {
				fmt.Printf("[ERROR] не удалось создать контракт для address=%s: %v\n", t.Address, err)
				return fmt.Errorf("ошибка создания контракта: %w", err)
			}
			fmt.Printf("[DEBUG] SaveToDB: контракт создан с id=%d\n", contractID)
		} else {
			fmt.Printf("[ERROR] contract_id не найден для address=%s: %v\n", t.Address, err)
			return fmt.Errorf("ошибка получения contract_id: %w", err)
		}
	}

	// Затем сохраняем токен (is_verified — для нативных монет из config)
	err = pool.QueryRow(ctx, `
		INSERT INTO tokens (
			contract_id, symbol, name, decimals, total_supply,
			standard, is_verified
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`,
		contractID, t.Symbol, t.Name, t.Decimals, t.TotalSupply,
		t.Standard, t.IsVerified,
	).Scan(&t.ID)

	if err != nil {
		return fmt.Errorf("ошибка сохранения токена: %w", err)
	}

	return nil
}

// UpdateStatus обновляет статус токена
func (t *Token) UpdateStatus(ctx context.Context, pool *pgxpool.Pool, status string) error {
	t.Status = status
	t.UpdatedAt = time.Now()

	_, err := pool.Exec(ctx, `
		UPDATE tokens 
		SET status = $1, updated_at = $2
		WHERE id = $3`,
		t.Status, t.UpdatedAt, t.ID,
	)

	if err != nil {
		return fmt.Errorf("ошибка обновления статуса токена: %w", err)
	}

	return nil
}

// LoadToken загружает токен из БД по адресу
func LoadToken(ctx context.Context, pool *pgxpool.Pool, address string) (*Token, error) {
	fmt.Printf("[DEBUG] LoadToken: ищем contract_id по адресу: %s\n", address)
	var id, contractID int
	var symbol, name, standard string
	var decimals int
	var totalSupply string

	// Получаем contract_id
	err := pool.QueryRow(ctx, `
		SELECT id FROM contracts WHERE address = $1`,
		address,
	).Scan(&contractID)
	if err != nil {
		fmt.Printf("[ERROR] contract_id не найден для address=%s: %v\n", address, err)
		return nil, fmt.Errorf("ошибка получения contract_id: %w", err)
	}

	// Получаем данные токена
	err = pool.QueryRow(ctx, `
		SELECT id, symbol, name, decimals, total_supply, standard
		FROM tokens
		WHERE contract_id = $1`,
		contractID,
	).Scan(&id, &symbol, &name, &decimals, &totalSupply, &standard)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("токен не найден: %s", address)
	}
	if err != nil {
		return nil, fmt.Errorf("ошибка загрузки токена: %w", err)
	}

	return &Token{
		ID:          id,
		Address:     address,
		Symbol:      symbol,
		Name:        name,
		Decimals:    decimals,
		TotalSupply: totalSupply,
		Standard:    standard,
	}, nil
}

// GetTokenByAddress возвращает токен по адресу
func GetTokenByAddress(ctx context.Context, pool *pgxpool.Pool, address string) (*Token, error) {
	return LoadToken(ctx, pool, address)
}

// GetTokenBySymbol возвращает токен по символу (для монет из config)
func GetTokenBySymbol(ctx context.Context, pool *pgxpool.Pool, symbol string) (*Token, error) {
	var id, contractID int
	var name, standard string
	var decimals int
	var totalSupply string
	err := pool.QueryRow(ctx, `
		SELECT t.id, t.contract_id, t.name, t.decimals, t.total_supply, t.standard
		FROM tokens t
		WHERE t.symbol = $1`,
		symbol,
	).Scan(&id, &contractID, &name, &decimals, &totalSupply, &standard)
	if err != nil {
		return nil, err
	}
	var address string
	if err := pool.QueryRow(ctx, `SELECT address FROM contracts WHERE id = $1`, contractID).Scan(&address); err != nil {
		return nil, err
	}
	return &Token{
		ID:          id,
		Address:     address,
		Symbol:      symbol,
		Name:        name,
		Decimals:    decimals,
		TotalSupply: totalSupply,
		Standard:    standard,
	}, nil
}

// GetTokensByOwner возвращает все токены, созданные указанным адресом
func GetTokensByOwner(ctx context.Context, pool *pgxpool.Pool, owner string) ([]*Token, error) {
	rows, err := pool.Query(ctx, `
		SELECT t.id, c.address, t.symbol, t.name, t.decimals, t.total_supply, t.standard
		FROM tokens t
		JOIN contracts c ON t.contract_id = c.id
		WHERE c.owner = $1
		ORDER BY t.id DESC`,
		owner,
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения токенов: %w", err)
	}
	defer rows.Close()

	var tokens []*Token
	for rows.Next() {
		var id int
		var address, symbol, name, standard string
		var decimals int
		var totalSupply string

		err := rows.Scan(&id, &address, &symbol, &name, &decimals, &totalSupply, &standard)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования токена: %w", err)
		}

		tokens = append(tokens, &Token{
			ID:          id,
			Address:     address,
			Symbol:      symbol,
			Name:        name,
			Decimals:    decimals,
			TotalSupply: totalSupply,
			Standard:    standard,
		})
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации токенов: %w", err)
	}

	return tokens, nil
}

// GetTokenBalance возвращает баланс токена для указанного адреса
func GetTokenBalance(ctx context.Context, pool *pgxpool.Pool, tokenAddress, accountAddress string) (*big.Int, error) {
	var balanceStr string
	err := pool.QueryRow(ctx, `
		SELECT balance
		FROM token_balances
		WHERE token_address = $1 AND account_address = $2`,
		tokenAddress, accountAddress,
	).Scan(&balanceStr)

	if err == sql.ErrNoRows {
		return big.NewInt(0), nil
	}
	if err != nil {
		return nil, fmt.Errorf("ошибка получения баланса токена: %w", err)
	}

	balance, ok := new(big.Int).SetString(balanceStr, 10)
	if !ok {
		return nil, fmt.Errorf("ошибка парсинга баланса: %s", balanceStr)
	}

	return balance, nil
}

// Прокси-методы для токенов стандарта gndst1
func (t *Token) IsGNDst1() bool {
	return t.Standard == "gndst1"
}

func (t *Token) GNDst1Instance() *gndst1.GNDst1 {
	if t.Standard != "gndst1" {
		return nil
	}
	inst, err := registry.GetToken(t.Address)
	if err != nil {
		return nil
	}
	return inst.(*gndst1.GNDst1)
}

// UniversalCall универсальный вызов метода токена
func (t *Token) UniversalCall(ctx context.Context, method string, args ...interface{}) (interface{}, error) {
	if t.IsGNDst1() {
		inst := t.GNDst1Instance()
		if inst == nil {
			return nil, errors.New("GNDst1 instance not found")
		}
		switch method {
		case "transfer":
			if len(args) != 3 {
				return nil, errors.New("transfer requires from, to, amount")
			}
			from, to := args[0].(string), args[1].(string)
			amount := args[2].(*big.Int)
			return nil, inst.Transfer(ctx, from, to, amount)
		case "approve":
			if len(args) != 3 {
				return nil, errors.New("approve requires owner, spender, amount")
			}
			owner, spender := args[0].(string), args[1].(string)
			amount := args[2].(*big.Int)
			return nil, inst.Approve(ctx, owner, spender, amount)
		case "balanceOf":
			if len(args) != 1 {
				return nil, errors.New("balanceOf requires address")
			}
			addr := args[0].(string)
			return inst.GetBalance(ctx, addr)
		// ... другие методы ...
		default:
			return nil, errors.New("unsupported method for GNDst1")
		}
	}
	// ... универсальный вызов для других стандартов ...
	return nil, errors.New("UniversalCall not implemented for this standard")
}
