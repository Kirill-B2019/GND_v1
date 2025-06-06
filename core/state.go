package core

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"math/big"
	"sync"
)

// State представляет состояние блокчейна (балансы, данные о транзакциях и т.д.)
type State struct {
	pool *pgxpool.Pool
	mu   sync.RWMutex
}

// NewState создает новое состояние, используя пул подключений к PostgreSQL
func NewState(pool *pgxpool.Pool) *State {
	return &State{
		pool: pool,
	}
}

// GetBalance возвращает баланс адреса по символу монеты
func (s *State) GetBalance(address Address, symbol string) *big.Int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var balanceStr string
	query := `
		SELECT COALESCE(balance, '0') FROM token_balances 
		WHERE address = $1 AND token_id = (SELECT id FROM tokens WHERE symbol = $2)
	`

	err := s.pool.QueryRow(context.Background(), query, string(address), symbol).Scan(&balanceStr)
	if err != nil {
		return big.NewInt(0)
	}

	balance := new(big.Int)
	if _, ok := balance.SetString(balanceStr, 10); !ok {
		return big.NewInt(0)
	}

	return balance
}

// AddBalance увеличивает баланс адреса на указанную сумму
func (s *State) AddBalance(address Address, symbol string, amount *big.Int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if amount.Sign() <= 0 {
		return fmt.Errorf("сумма пополнения должна быть положительной")
	}
	tokenID, err := s.getTokenID(symbol)
	if err != nil {
		return fmt.Errorf("не удалось получить ID токена для символа %s: %v", symbol, err)
	}

	tx, err := s.pool.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("не удалось начать транзакцию: %v", err)
	}
	defer tx.Rollback(context.Background())

	var exists bool
	err = tx.QueryRow(
		context.Background(),
		"SELECT EXISTS(SELECT 1 FROM token_balances WHERE address = $1 AND token_id = $2)",
		string(address),
		tokenID,
	).Scan(&exists)
	if err != nil {
		return fmt.Errorf("ошибка проверки существования записи: %v", err)
	}

	if exists {
		_, err = tx.Exec(
			context.Background(),
			`
			UPDATE token_balances
			SET balance = balance + $3::numeric
			WHERE address = $1 AND token_id = $2
			`,
			string(address),
			tokenID,
			amount.String(),
		)
	} else {
		_, err = tx.Exec(
			context.Background(),
			`
			INSERT INTO token_balances (token_id, address, balance)
			VALUES ($1, $2, $3::numeric)
			`,
			tokenID,
			string(address),
			amount.String(),
		)
	}

	if err != nil {
		return fmt.Errorf("не удалось обновить баланс: %v", err)
	}

	return tx.Commit(context.Background())
}

// SubBalance уменьшает баланс адреса на указанную сумму
func (s *State) SubBalance(address Address, symbol string, amount *big.Int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tokenID, err := s.getTokenID(symbol)
	if err != nil {
		return fmt.Errorf("не удалось получить ID токена для символа %s: %v", symbol, err)
	}

	tx, err := s.pool.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("не удалось начать транзакцию: %v", err)
	}
	defer tx.Rollback(context.Background())

	var currentBalanceStr string
	err = tx.QueryRow(
		context.Background(),
		"SELECT balance FROM token_balances WHERE address = $1 AND token_id = $2",
		string(address),
		tokenID,
	).Scan(&currentBalanceStr)
	if err != nil {
		return fmt.Errorf("не удалось получить текущий баланс: %v", err)
	}

	currentBalance := new(big.Int)
	if _, ok := currentBalance.SetString(currentBalanceStr, 10); !ok {
		return fmt.Errorf("недопустимый формат баланса: %s", currentBalanceStr)
	}

	if currentBalance.Cmp(amount) < 0 {
		return fmt.Errorf("недостаточно средств для списания")
	}

	newBalance := new(big.Int).Sub(currentBalance, amount)

	_, err = tx.Exec(
		context.Background(),
		`
		UPDATE token_balances
		SET balance = $3::numeric
		WHERE address = $1 AND token_id = $2
		`,
		string(address),
		tokenID,
		newBalance.String(),
	)
	if err != nil {
		return fmt.Errorf("не удалось обновить баланс: %v", err)
	}

	return tx.Commit(context.Background())
}

// Credit добавляет указанный токен на адрес
func (s *State) Credit(address Address, symbol string, amount *big.Int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tokenID, err := s.getTokenID(symbol)
	if err != nil {
		panic(fmt.Sprintf("невозможно получить ID токена: %v", err))
	}

	_, err = s.pool.Exec(
		context.Background(),
		`
		INSERT INTO token_balances (token_id, address, balance)
		VALUES ($1, $2, $3::numeric)
		ON CONFLICT (token_id, address) DO UPDATE
		SET balance = token_balances.balance + EXCLUDED.balance
		`,
		tokenID,
		string(address),
		amount.String(),
	)
	if err != nil {
		panic(fmt.Sprintf("не удалось зачислить средства: %v", err))
	}
}

// SaveToDB сохраняет текущее состояние в БД
func (s *State) SaveToDB() error {
	// В данном случае мы не сохраняем напрямую — это делают другие методы.
	// Пример реализации при необходимости:
	return nil
}

// LoadTokenBalances загружает балансы токенов для заданного адреса
func (s *State) LoadTokenBalances(address Address) map[string]*big.Int {
	rows, err := s.pool.Query(
		context.Background(),
		`
		SELECT t.symbol, b.balance 
		FROM token_balances b
		JOIN tokens t ON b.token_id = t.id
		WHERE b.address = $1
		`,
		string(address),
	)
	if err != nil {
		return nil
	}
	defer rows.Close()

	balances := make(map[string]*big.Int)
	for rows.Next() {
		var symbol, balanceStr string
		if err := rows.Scan(&symbol, &balanceStr); err != nil {
			continue
		}

		balance := new(big.Int)
		if _, ok := balance.SetString(balanceStr, 10); !ok {
			continue
		}

		balances[symbol] = balance
	}

	return balances
}

// getTokenID возвращает ID токена по его символу
func (s *State) getTokenID(symbol string) (int, error) {
	var tokenID int
	err := s.pool.QueryRow(
		context.Background(),
		"SELECT id FROM tokens WHERE symbol = $1",
		symbol,
	).Scan(&tokenID)

	if err != nil {
		return 0, fmt.Errorf("токен с символом %s не найден", symbol)
	}

	return tokenID, nil
}

// ApplyTransaction применяет транзакцию к состоянию
func (s *State) ApplyTransaction(tx *Transaction) error {
	from := Address(tx.From)
	to := Address(tx.To)
	value := tx.Value

	if value.Sign() == 0 {
		return nil // нулевая сумма перевода
	}

	if err := s.SubBalance(from, tx.Symbol, value); err != nil {
		return fmt.Errorf("не удалось списать средство: %v", err)
	}

	if err := s.AddBalance(to, tx.Symbol, value); err != nil {
		return fmt.Errorf("не удалось зачислить средство: %v", err)
	}

	return nil
}

// TransferToken передает токены от одного адреса к другому
func (s *State) TransferToken(from, to Address, symbol string, amount *big.Int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if amount.Sign() <= 0 {
		return fmt.Errorf("сумма перевода должна быть положительной")
	}

	if from == to {
		return fmt.Errorf("нельзя перевести самому себе")
	}

	if err := s.SubBalance(from, symbol, amount); err != nil {
		return err
	}

	if err := s.AddBalance(to, symbol, amount); err != nil {
		return err
	}

	return nil
}

// UpdateNonce обновляет nonce для адреса
func (s *State) UpdateNonce(address Address, nonce uint64) error {
	_, err := s.pool.Exec(
		context.Background(),
		`UPDATE accounts SET nonce = $1 WHERE address = $2`,
		nonce,
		string(address),
	)
	return err
}

// GetNonce получает текущий nonce для адреса
func (s *State) GetNonce(address Address) (uint64, error) {
	var nonce uint64
	err := s.pool.QueryRow(
		context.Background(),
		"SELECT nonce FROM accounts WHERE address = $1",
		string(address),
	).Scan(&nonce)

	if err != nil {
		return 0, fmt.Errorf("не удалось получить nonce: %v", err)
	}

	return nonce, nil
}

// ValidateAddress проверяет, существует ли адрес в системе
func (s *State) ValidateAddress(address Address) bool {
	var exists bool
	err := s.pool.QueryRow(
		context.Background(),
		"SELECT EXISTS(SELECT 1 FROM accounts WHERE address = $1)",
		string(address),
	).Scan(&exists)

	if err != nil {
		return false
	}

	return exists
}

// CallStatic выполняет статический вызов транзакции (без изменения состояния)
func (s *State) CallStatic(from, to Address, data []byte, gasLimit, gasPrice, value uint64) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Здесь можно реализовать логику выполнения контракта без изменения балансов
	// Например, эмуляция вызова метода balanceOf, transfer и т.д.

	// Пример заглушки:
	if to == "" {
		return nil, errors.New("invalid contract address")
	}

	// Получаем баланс (пример)
	balance := s.GetBalance(to, "GND") // предположим, что символ = "GND"
	return []byte(fmt.Sprintf("balance: %s", balance.String())), nil
}

// Close освобождает ресурсы состояния
func (s *State) Close() {
	// Здесь можно реализовать логику завершения, если нужно
}
