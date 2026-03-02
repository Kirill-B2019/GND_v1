// | KB @CerbeRus - Nexus Invest Team
package core

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"GND/types"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	globalState *State
	stateMutex  sync.RWMutex
)

// BlockchainState - structure of blockchain state
type BlockchainState struct {
	ID          int       // ID of state
	BlockID     int       // ID of block
	Address     string    // Account address
	Balance     *big.Int  // Balance in native tokens
	Nonce       uint64    // Number of the last transaction
	StorageRoot string    // Storage root
	CodeHash    string    // Contract code hash
	CreatedAt   time.Time // Creation time
	UpdatedAt   time.Time // Last update time
	Metadata    []byte    // State metadata
}

// NewBlockchainState creates a new state
func NewBlockchainState(blockID int, address string, balance *big.Int, nonce uint64, storageRoot, codeHash string) *BlockchainState {
	now := time.Now()
	return &BlockchainState{
		BlockID:     blockID,
		Address:     address,
		Balance:     balance,
		Nonce:       nonce,
		StorageRoot: storageRoot,
		CodeHash:    codeHash,
		CreatedAt:   now,
		UpdatedAt:   now,
		Metadata:    []byte{},
	}
}

// SaveToDB saves state to DB
func (s *BlockchainState) SaveToDB(ctx context.Context, pool *pgxpool.Pool) error {
	err := pool.QueryRow(ctx, `
		INSERT INTO states (
			block_id, address, balance, nonce, storage_root,
			code_hash, created_at, updated_at, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id`,
		s.BlockID, s.Address, s.Balance.String(), s.Nonce, s.StorageRoot,
		s.CodeHash, s.CreatedAt, s.UpdatedAt, s.Metadata,
	).Scan(&s.ID)

	if err != nil {
		return fmt.Errorf("state saving error: %w", err)
	}

	return nil
}

// UpdateBalance updates state balance
func (s *BlockchainState) UpdateBalance(ctx context.Context, pool *pgxpool.Pool, newBalance *big.Int) error {
	s.Balance = newBalance
	s.UpdatedAt = time.Now()

	_, err := pool.Exec(ctx, `
		UPDATE states 
		SET balance = $1, updated_at = $2
		WHERE id = $3`,
		s.Balance.String(), s.UpdatedAt, s.ID,
	)

	if err != nil {
		return fmt.Errorf("balance update error: %w", err)
	}

	return nil
}

// IncrementNonce increases state nonce
func (s *BlockchainState) IncrementNonce(ctx context.Context, pool *pgxpool.Pool) error {
	s.Nonce++
	s.UpdatedAt = time.Now()

	_, err := pool.Exec(ctx, `
		UPDATE states 
		SET nonce = $1, updated_at = $2
		WHERE id = $3`,
		s.Nonce, s.UpdatedAt, s.ID,
	)

	if err != nil {
		return fmt.Errorf("nonce update error: %w", err)
	}

	return nil
}

// LoadBlockchainState loads state from DB by address and block ID
func LoadBlockchainState(ctx context.Context, pool *pgxpool.Pool, address string, blockID int) (*BlockchainState, error) {
	var id int
	var balanceStr string
	var nonce uint64
	var storageRoot, codeHash string
	var createdAt, updatedAt time.Time
	var metadata []byte

	err := pool.QueryRow(ctx, `
		SELECT id, balance, nonce, storage_root, code_hash,
			created_at, updated_at, metadata
		FROM states
		WHERE address = $1 AND block_id = $2`,
		address, blockID,
	).Scan(&id, &balanceStr, &nonce, &storageRoot, &codeHash,
		&createdAt, &updatedAt, &metadata)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("state not found: %s (block %d)", address, blockID)
	}
	if err != nil {
		return nil, fmt.Errorf("state loading error: %w", err)
	}

	balance, ok := new(big.Int).SetString(balanceStr, 10)
	if !ok {
		return nil, fmt.Errorf("balance parsing error: %s", balanceStr)
	}

	return &BlockchainState{
		ID:          id,
		BlockID:     blockID,
		Address:     address,
		Balance:     balance,
		Nonce:       nonce,
		StorageRoot: storageRoot,
		CodeHash:    codeHash,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		Metadata:    metadata,
	}, nil
}

// GetStateBalance returns state balance
func GetStateBalance(ctx context.Context, pool *pgxpool.Pool, address string, blockID int) (*big.Int, error) {
	var balanceStr string
	err := pool.QueryRow(ctx, `
		SELECT balance
		FROM states
		WHERE address = $1 AND block_id = $2`,
		address, blockID,
	).Scan(&balanceStr)

	if err == sql.ErrNoRows {
		return big.NewInt(0), nil
	}
	if err != nil {
		return nil, fmt.Errorf("state balance getting error: %w", err)
	}

	balance, ok := new(big.Int).SetString(balanceStr, 10)
	if !ok {
		return nil, fmt.Errorf("balance parsing error: %s", balanceStr)
	}

	return balance, nil
}

// GetStateNonce returns state nonce
func GetStateNonce(ctx context.Context, pool *pgxpool.Pool, address string, blockID int) (uint64, error) {
	var nonce uint64
	err := pool.QueryRow(ctx, `
		SELECT nonce
		FROM states
		WHERE address = $1 AND block_id = $2`,
		address, blockID,
	).Scan(&nonce)

	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("state nonce getting error: %w", err)
	}

	return nonce, nil
}

// State представляет состояние блокчейна
type State struct {
	balances         map[types.Address]map[string]*big.Int
	nonces           map[types.Address]uint64
	pool             *pgxpool.Pool
	mutex            sync.RWMutex
	gndContractAddr  string // если задан — GND берётся из token_balances по token_id
	ganiContractAddr string
}

// NewState создает новое состояние
func NewState() *State {
	return &State{
		balances: make(map[types.Address]map[string]*big.Int),
		nonces:   make(map[types.Address]uint64),
	}
}

// SetPool устанавливает пул соединений с БД
func (s *State) SetPool(pool *pgxpool.Pool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.pool = pool
}

// SetNativeContractAddresses задаёт адреса контрактов GND/GANI (режим «всё на контрактах»).
// Пустые строки — использовать native_balances по-прежнему.
func (s *State) SetNativeContractAddresses(gndAddr, ganiAddr string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.gndContractAddr = strings.TrimSpace(gndAddr)
	s.ganiContractAddr = strings.TrimSpace(ganiAddr)
}

// getTokenIDForNativeContractLocked возвращает token_id по символу и адресу контракта (tokens JOIN contracts).
// Вызывать при удержанном s.mutex (RLock или Lock).
func (s *State) getTokenIDForNativeContractLocked(symbol, contractAddr string) (int, error) {
	if s.pool == nil || contractAddr == "" {
		return -1, fmt.Errorf("no pool or contract address")
	}
	var tokenID int
	err := s.pool.QueryRow(context.Background(), `
		SELECT t.id FROM tokens t
		INNER JOIN contracts c ON t.contract_id = c.id
		WHERE t.symbol = $1 AND c.address = $2`,
		symbol, contractAddr,
	).Scan(&tokenID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return -1, fmt.Errorf("token not found for %s at %s", symbol, contractAddr)
		}
		return -1, err
	}
	return tokenID, nil
}

// GetBalance возвращает баланс адреса. Для GND/GANI в режиме контрактов — из token_balances.
func (s *State) GetBalance(address types.Address, symbol string) *big.Int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Режим «всё на контрактах»: GND/GANI из token_balances по token_id
	if symbol == GasSymbol && s.gndContractAddr != "" {
		if tokenID, err := s.getTokenIDForNativeContractLocked(GasSymbol, s.gndContractAddr); err == nil {
			var balanceStr string
			if err := s.pool.QueryRow(context.Background(),
				`SELECT COALESCE(balance::text, '0') FROM token_balances WHERE token_id = $1 AND address = $2`,
				tokenID, string(address)).Scan(&balanceStr); err == nil {
				b := new(big.Int)
				if _, ok := b.SetString(balanceStr, 10); ok {
					return b
				}
			}
			return big.NewInt(0)
		}
	}
	if symbol == "GANI" && s.ganiContractAddr != "" {
		if tokenID, err := s.getTokenIDForNativeContractLocked("GANI", s.ganiContractAddr); err == nil {
			var balanceStr string
			if err := s.pool.QueryRow(context.Background(),
				`SELECT COALESCE(balance::text, '0') FROM token_balances WHERE token_id = $1 AND address = $2`,
				tokenID, string(address)).Scan(&balanceStr); err == nil {
				b := new(big.Int)
				if _, ok := b.SetString(balanceStr, 10); ok {
					return b
				}
			}
			return big.NewInt(0)
		}
	}

	if balances, ok := s.balances[address]; ok {
		if balance, ok := balances[symbol]; ok {
			return new(big.Int).Set(balance)
		}
	}
	return big.NewInt(0)
}

// getTotalNativeBalanceLocked возвращает сумму балансов по нативному символу (вызывать при удержанном mutex).
func (s *State) getTotalNativeBalanceLocked(symbol string) *big.Int {
	total := big.NewInt(0)
	for _, balances := range s.balances {
		if b, ok := balances[symbol]; ok && b != nil {
			total.Add(total, b)
		}
	}
	return total
}

// getCirculatingSupplyCap возвращает лимит циркулирующего предложения из tokens.circulating_supply по символу.
func (s *State) getCirculatingSupplyCap(symbol string) (*big.Int, error) {
	if s.pool == nil {
		return nil, nil
	}
	var capStr string
	err := s.pool.QueryRow(context.Background(),
		`SELECT COALESCE(circulating_supply::text, total_supply::text) FROM tokens WHERE symbol = $1`,
		symbol,
	).Scan(&capStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	cap := new(big.Int)
	if _, ok := cap.SetString(capStr, 10); !ok {
		return nil, fmt.Errorf("invalid circulating_supply for %s", symbol)
	}
	return cap, nil
}

// AddBalance добавляет баланс адресу. Для нативных монет (GND, GANI) проверяет лимит циркулирующего предложения.
// В режиме контрактов (gnd/gani_contract_address заданы) обновляет token_balances.
func (s *State) AddBalance(address types.Address, symbol string, amount *big.Int) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if amount.Sign() <= 0 {
		return nil
	}

	// Режим контрактов: GND/GANI — в token_balances
	if symbol == GasSymbol && s.gndContractAddr != "" && s.pool != nil {
		tokenID, err := s.getTokenIDForNativeContractLocked(GasSymbol, s.gndContractAddr)
		if err == nil {
			_, err = s.pool.Exec(context.Background(), `
				INSERT INTO token_balances (token_id, address, balance) VALUES ($1, $2, $3)
				ON CONFLICT (token_id, address) DO UPDATE SET balance = token_balances.balance + $3`,
				tokenID, string(address), amount.String())
			return err
		}
	}
	if symbol == "GANI" && s.ganiContractAddr != "" && s.pool != nil {
		tokenID, err := s.getTokenIDForNativeContractLocked("GANI", s.ganiContractAddr)
		if err == nil {
			_, err = s.pool.Exec(context.Background(), `
				INSERT INTO token_balances (token_id, address, balance) VALUES ($1, $2, $3)
				ON CONFLICT (token_id, address) DO UPDATE SET balance = token_balances.balance + $3`,
				tokenID, string(address), amount.String())
			return err
		}
	}

	if IsNativeSymbol(symbol) && s.pool != nil {
		cap, err := s.getCirculatingSupplyCap(symbol)
		if err == nil && cap != nil && cap.Sign() >= 0 {
			current := s.getTotalNativeBalanceLocked(symbol)
			newTotal := new(big.Int).Add(current, amount)
			if newTotal.Cmp(cap) > 0 {
				return fmt.Errorf("circulating supply limit exceeded for %s: current + amount > cap", symbol)
			}
		}
	}

	if s.balances[address] == nil {
		s.balances[address] = make(map[string]*big.Int)
	}
	if s.balances[address][symbol] == nil {
		s.balances[address][symbol] = big.NewInt(0)
	}
	s.balances[address][symbol].Add(s.balances[address][symbol], amount)
	return nil
}

// SubBalance вычитает баланс у адреса. В режиме контрактов обновляет token_balances.
func (s *State) SubBalance(address types.Address, symbol string, amount *big.Int) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if amount.Sign() <= 0 {
		return nil
	}

	// Режим контрактов: GND/GANI — в token_balances
	if symbol == GasSymbol && s.gndContractAddr != "" && s.pool != nil {
		tokenID, err := s.getTokenIDForNativeContractLocked(GasSymbol, s.gndContractAddr)
		if err == nil {
			res, err := s.pool.Exec(context.Background(), `
				UPDATE token_balances SET balance = balance - $3 WHERE token_id = $1 AND address = $2 AND balance >= $3`,
				tokenID, string(address), amount.String())
			if err != nil {
				return err
			}
			if res.RowsAffected() == 0 {
				return errors.New("insufficient balance")
			}
			return nil
		}
	}
	if symbol == "GANI" && s.ganiContractAddr != "" && s.pool != nil {
		tokenID, err := s.getTokenIDForNativeContractLocked("GANI", s.ganiContractAddr)
		if err == nil {
			res, err := s.pool.Exec(context.Background(), `
				UPDATE token_balances SET balance = balance - $3 WHERE token_id = $1 AND address = $2 AND balance >= $3`,
				tokenID, string(address), amount.String())
			if err != nil {
				return err
			}
			if res.RowsAffected() == 0 {
				return errors.New("insufficient balance")
			}
			return nil
		}
	}

	if s.balances[address] == nil {
		s.balances[address] = make(map[string]*big.Int)
	}
	if s.balances[address][symbol] == nil {
		s.balances[address][symbol] = big.NewInt(0)
	}
	if s.balances[address][symbol].Cmp(amount) < 0 {
		return errors.New("insufficient balance")
	}
	s.balances[address][symbol].Sub(s.balances[address][symbol], amount)
	return nil
}

// GetNonce возвращает nonce адреса
func (s *State) GetNonce(address types.Address) int64 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return int64(s.nonces[address])
}

// IncrementNonce увеличивает nonce адреса
func (s *State) IncrementNonce(address types.Address) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.nonces[address]++
}

// ApplyTransaction применяет транзакцию к состоянию. Используется tx.Symbol (GND или GANI); при пустом — GND.
func (s *State) ApplyTransaction(tx *Transaction) error {
	// Проверяем nonce
	if int64(tx.Nonce) != s.GetNonce(types.Address(tx.Sender)) {
		return errors.New("invalid nonce")
	}

	// Проверяем баланс (сумма + газ при необходимости)
	if !tx.HasSufficientBalance() {
		return errors.New("insufficient balance")
	}

	symbol := tx.Symbol
	if symbol == "" {
		symbol = GasSymbol // GND
	}
	if !IsNativeSymbol(symbol) {
		return errors.New("symbol must be native (GND or GANI)")
	}

	// Списываем баланс отправителя по символу транзакции
	if err := s.SubBalance(types.Address(tx.Sender), symbol, tx.Value); err != nil {
		return err
	}

	// Начисляем баланс получателю
	if err := s.AddBalance(types.Address(tx.Recipient), symbol, tx.Value); err != nil {
		s.AddBalance(types.Address(tx.Sender), symbol, tx.Value)
		return err
	}

	s.IncrementNonce(types.Address(tx.Sender))
	return nil
}

// SaveToDB сохраняет состояние в БД. Нативные балансы (GND, GANI) — в native_balances; nonces — в accounts.
func (s *State) SaveToDB() error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Сохраняем нативные балансы (GND, GANI) в native_balances только если не включён режим контрактов
	for address, balances := range s.balances {
		for symbol, balance := range balances {
			if !IsNativeSymbol(symbol) {
				continue
			}
			if symbol == GasSymbol && s.gndContractAddr != "" {
				continue // GND в token_balances
			}
			if symbol == "GANI" && s.ganiContractAddr != "" {
				continue // GANI в token_balances
			}
			_, err := s.pool.Exec(context.Background(), `
				INSERT INTO native_balances (address, symbol, balance, updated_at)
				VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
				ON CONFLICT (address, symbol) DO UPDATE
				SET balance = $3, updated_at = CURRENT_TIMESTAMP`, address, symbol, balance.String())
			if err != nil {
				return err
			}
		}
	}

	// Сохраняем nonce
	for address, nonce := range s.nonces {
		_, err := s.pool.Exec(context.Background(), `
			INSERT INTO accounts (address, nonce)
			VALUES ($1, $2)
			ON CONFLICT (address) DO UPDATE
			SET nonce = $2`, address, nonce)
		if err != nil {
			return err
		}
	}

	return nil
}

// LoadFromDB загружает состояние из БД. Сначала нативные балансы (native_balances), затем контрактные токены (token_balances JOIN tokens).
func (s *State) LoadFromDB(ctx context.Context) error {
	if s.pool == nil {
		return errors.New("database pool not set")
	}

	// 1. Загружаем нативные балансы (GND, GANI) из native_balances; в режиме контрактов GND/GANI не грузим (источник — token_balances)
	rowsNative, err := s.pool.Query(ctx, `SELECT address, symbol, balance FROM native_balances`)
	if err != nil {
		return err
	}
	for rowsNative.Next() {
		var address types.Address
		var symbol string
		var balanceStr string
		if err := rowsNative.Scan(&address, &symbol, &balanceStr); err != nil {
			rowsNative.Close()
			return err
		}
		if symbol == GasSymbol && s.gndContractAddr != "" {
			continue
		}
		if symbol == "GANI" && s.ganiContractAddr != "" {
			continue
		}
		balance := new(big.Int)
		if _, ok := balance.SetString(balanceStr, 10); !ok {
			continue
		}
		if s.balances[address] == nil {
			s.balances[address] = make(map[string]*big.Int)
		}
		s.balances[address][symbol] = balance
	}
	rowsNative.Close()

	// 2. Загружаем балансы контрактных токенов (token_balances JOIN tokens); нативные не перезаписываем
	rows, err := s.pool.Query(ctx, `
		SELECT tb.address, t.symbol, tb.balance
		FROM token_balances tb
		JOIN tokens t ON t.id = tb.token_id`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var address types.Address
		var symbol string
		var balanceStr string
		if err := rows.Scan(&address, &symbol, &balanceStr); err != nil {
			return err
		}
		if IsNativeSymbol(symbol) {
			continue // уже загружены из native_balances
		}
		balance := new(big.Int)
		balance.SetString(balanceStr, 10)
		if s.balances[address] == nil {
			s.balances[address] = make(map[string]*big.Int)
		}
		s.balances[address][symbol] = balance
	}

	// 3. Загружаем nonces
	rows, err = s.pool.Query(ctx, `
		SELECT address, nonce
		FROM accounts`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var address types.Address
		var nonce uint64
		if err := rows.Scan(&address, &nonce); err != nil {
			return err
		}
		s.nonces[address] = nonce
	}

	// accounts.balance не синхронизируем из token_balances (отключено по требованию)

	return nil
}

// ApplyExecutionResult применяет результат выполнения контракта. Газ списывается в GND.
func (s *State) ApplyExecutionResult(tx *Transaction, result *types.ExecutionResult) error {
	gasUsed := new(big.Int).SetUint64(result.GasUsed)
	if err := s.SubBalance(types.Address(tx.Sender), GasSymbol, gasUsed); err != nil {
		return err
	}

	for _, change := range result.StateChanges {
		switch change.Type {
		case types.ChangeTypeBalance:
			if err := s.AddBalance(types.Address(change.Address), change.Symbol, change.Amount); err != nil {
				s.AddBalance(types.Address(tx.Sender), GasSymbol, gasUsed)
				return err
			}
		}
	}

	// Увеличиваем nonce отправителя
	s.IncrementNonce(types.Address(tx.Sender))

	return nil
}

// LoadTokenBalances loads token balances for a given address
func (s *State) LoadTokenBalances(address types.Address) map[string]*big.Int {
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

// getTokenID returns token ID by its symbol
func (s *State) getTokenID(symbol string) (int, error) {
	var tokenID int
	err := s.pool.QueryRow(
		context.Background(),
		"SELECT id FROM tokens WHERE symbol = $1",
		symbol,
	).Scan(&tokenID)

	if err != nil {
		return 0, fmt.Errorf("token with symbol %s not found", symbol)
	}

	return tokenID, nil
}

// TransferToken transfers tokens from one address to another
func (s *State) TransferToken(from, to types.Address, symbol string, amount *big.Int) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if amount.Sign() <= 0 {
		return fmt.Errorf("transfer amount must be positive")
	}

	if from == to {
		return fmt.Errorf("cannot transfer to yourself")
	}

	if err := s.SubBalance(from, symbol, amount); err != nil {
		return err
	}

	if err := s.AddBalance(to, symbol, amount); err != nil {
		return err
	}

	return nil
}

// UpdateNonce updates nonce for address
func (s *State) UpdateNonce(address types.Address, nonce uint64) error {
	_, err := s.pool.Exec(
		context.Background(),
		`UPDATE accounts SET nonce = $1 WHERE address = $2`,
		nonce,
		string(address),
	)
	return err
}

// ValidateAddress checks if address exists in system
func (s *State) ValidateAddress(address types.Address) bool {
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
func (s *State) CallStatic(tx *Transaction) (*types.ExecutionResult, error) {
	if tx == nil || tx.Recipient == "" {
		return nil, errors.New("invalid contract call")
	}
	balance := s.GetBalance(tx.Recipient, "GND")
	return &types.ExecutionResult{
		GasUsed:    0,
		ReturnData: []byte(fmt.Sprintf("balance: %s", balance.String())),
		Error:      nil,
	}, nil
}

// Close releases state resources
func (s *State) Close() {
	// Here you can implement logic to complete if needed
}

// Credit adds balance for an address and symbol
func (s *State) Credit(address types.Address, symbol string, amount *big.Int) error {
	return s.AddBalance(address, symbol, amount)
}

// SaveContract saves a contract to state
func (s *State) SaveContract(contract *types.Contract) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Save contract to database
	_, err := s.pool.Exec(context.Background(), `
		INSERT INTO contracts (
			address, bytecode, name, symbol, standard,
			owner, description, version, compiler,
			params, metadata_cid, source_code
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (address) DO UPDATE SET
			bytecode = $2,
			name = $3,
			symbol = $4,
			standard = $5,
			owner = $6,
			description = $7,
			version = $8,
			compiler = $9,
			params = $10,
			metadata_cid = $11,
			source_code = $12`,
		contract.Address,
		contract.Bytecode,
		contract.Name,
		contract.Symbol,
		contract.Standard,
		contract.Owner,
		contract.Description,
		contract.Version,
		contract.Compiler,
		contract.Params,
		contract.MetadataCID,
		contract.SourceCode,
	)

	if err != nil {
		return fmt.Errorf("contract saving error: %w", err)
	}

	return nil
}

// GetContract returns a contract by address
func (s *State) GetContract(address string) (*types.Contract, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var contract types.Contract
	var params map[string]string

	err := s.pool.QueryRow(context.Background(), `
		SELECT address, bytecode, name, symbol, standard,
			owner, description, version, compiler,
			params, metadata_cid, source_code
		FROM contracts
		WHERE address = $1`,
		address,
	).Scan(
		&contract.Address,
		&contract.Bytecode,
		&contract.Name,
		&contract.Symbol,
		&contract.Standard,
		&contract.Owner,
		&contract.Description,
		&contract.Version,
		&contract.Compiler,
		&params,
		&contract.MetadataCID,
		&contract.SourceCode,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("contract not found: %s", address)
	}
	if err != nil {
		return nil, fmt.Errorf("contract loading error: %w", err)
	}

	contract.Params = params
	return &contract, nil
}
