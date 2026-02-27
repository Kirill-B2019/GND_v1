package core

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"GND/types"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Blockchain represents the blockchain structure
type Blockchain struct {
	Genesis *Block
	State   StateIface
	Pool    *pgxpool.Pool
	Blocks  []*Block
	Mempool *Mempool
	mutex   sync.Mutex
}

// NewBlockchain creates a new blockchain
func NewBlockchain(genesis *Block, pool *pgxpool.Pool) *Blockchain {
	return &Blockchain{
		Genesis: genesis,
		State:   NewState(),
		Pool:    pool,
		Blocks:  []*Block{genesis},
		Mempool: NewMempool(),
	}
}

// LoadBlockchainFromDB loads blockchain from database
func LoadBlockchainFromDB(pool *pgxpool.Pool) (*Blockchain, error) {
	ctx := context.Background()
	genesis, err := GetBlockByNumber(pool, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to load genesis block: %v", err)
	}

	state := NewState()
	state.SetPool(pool)
	if err := state.LoadFromDB(ctx); err != nil {
		return nil, fmt.Errorf("failed to load state from DB: %w", err)
	}

	return &Blockchain{
		Genesis: genesis,
		State:   state,
		Pool:    pool,
		Blocks:  []*Block{genesis},
		Mempool: NewMempool(),
	}, nil
}

// GetBlockByNumber returns a block by its number
func (bc *Blockchain) GetBlockByNumber(number uint64) (*Block, error) {
	return GetBlockByNumber(bc.Pool, number)
}

// GetBlockByHash returns a block by its hash
func (bc *Blockchain) GetBlockByHash(hash string) (*Block, error) {
	return GetBlockByHash(bc.Pool, hash)
}

// GetLatestBlock returns the latest block
func (bc *Blockchain) GetLatestBlock() (*Block, error) {
	return GetLatestBlock(bc.Pool)
}

// AddTransaction adds a transaction to the blockchain
func (bc *Blockchain) AddTransaction(tx *Transaction) error {
	// Проверяем транзакцию
	if err := tx.Validate(); err != nil {
		return fmt.Errorf("invalid transaction: %v", err)
	}

	// Проверяем баланс отправителя
	if !tx.HasSufficientBalance() {
		return errors.New("insufficient balance")
	}

	// Применяем транзакцию к состоянию
	if err := bc.State.ApplyTransaction(tx); err != nil {
		return fmt.Errorf("failed to apply transaction: %v", err)
	}

	return nil
}

// EnsureCoinsDeployed проверяет при первом запуске наличие монет из config в БД и при необходимости создаёт контракты и токены.
func EnsureCoinsDeployed(ctx context.Context, pool *pgxpool.Pool, cfg *Config, ownerAddress string) error {
	for _, coin := range cfg.Coins {
		_, err := GetTokenBySymbol(ctx, pool, coin.Symbol)
		if err == nil {
			continue
		}
		// Токен не найден — создаём контракт и токен (нативные монеты из config считаем верифицированными)
		addr := coin.ContractAddress
		if addr == "" || len(addr) < 10 {
			addr = ""
		}
		t := NewToken(addr, coin.Symbol, coin.Name, coin.Decimals, coin.TotalSupply, ownerAddress, "coin", coin.Standard, 0, 0)
		t.IsVerified = true
		if err := t.SaveToDB(ctx, pool); err != nil {
			return fmt.Errorf("деплой монеты %s: %w", coin.Symbol, err)
		}
		// Пометить контракт нативной монеты как верифицированный
		_, _ = pool.Exec(ctx, `UPDATE contracts SET is_verified = true WHERE address = $1`, t.Address)
	}
	return nil
}

// genesisTimestamp — фиксированное время генезиса, чтобы системные транзакции попадали в существующую партицию transactions (например 2025_06).
var genesisTimestamp = time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)

// FirstLaunch initializes blockchain on first launch
func (bc *Blockchain) FirstLaunch(ctx context.Context, pool *pgxpool.Pool, wallet *Wallet, cfg *Config) error {
	// 1. Проверка и деплой монет из config (контракты + токены в БД)
	if err := EnsureCoinsDeployed(ctx, pool, cfg, string(wallet.Address)); err != nil {
		return fmt.Errorf("проверка деплоя монет: %w", err)
	}

	// 2. Генезис: фиксированное время и метаданные для корректной записи в БД и в партицию transactions
	bc.Genesis.Timestamp = genesisTimestamp
	bc.Genesis.CreatedAt = genesisTimestamp
	bc.Genesis.UpdatedAt = time.Now().UTC()
	bc.Genesis.Hash = bc.Genesis.CalculateHash()

	if err := bc.Genesis.SaveToDB(ctx, pool); err != nil {
		return fmt.Errorf("failed to save genesis block: %v", err)
	}

	// 2.1. Транзакция о создании генезис-блока (данные в таблице transactions)
	sysTxGenesis := newSystemTransaction(
		int(bc.Genesis.ID),
		bc.Genesis.Timestamp,
		"genesis",
		types.Address("GND_GENESIS"),
		types.Address(wallet.Address),
		big.NewInt(0),
		"",
	)
	if err := saveSystemTransaction(ctx, pool, 1, sysTxGenesis); err != nil {
		return fmt.Errorf("запись системной транзакции genesis: %w", err)
	}

	// 3. Инициализируем состояние и привязываем пул БД
	state := NewState()
	state.SetPool(pool)
	bc.State = state

	// 4. Начисляем начальный баланс в памяти
	for _, coin := range cfg.Coins {
		amount := new(big.Int)
		amount.SetString(coin.TotalSupply, 10)
		if err := state.Credit(types.Address(wallet.Address), coin.Symbol, amount); err != nil {
			return fmt.Errorf("failed to add initial balance for %s: %v", coin.Symbol, err)
		}
	}

	// 5. Сохраняем балансы в БД: только token_balances (accounts.balance при первом запуске не начисляем)
	for _, coin := range cfg.Coins {
		amount := new(big.Int)
		amount.SetString(coin.TotalSupply, 10)
		tok, err := GetTokenBySymbol(ctx, pool, coin.Symbol)
		if err != nil {
			return fmt.Errorf("токен %s не найден после деплоя: %w", coin.Symbol, err)
		}
		_, err = pool.Exec(ctx, `
			INSERT INTO token_balances (token_id, address, balance)
			VALUES ($1, $2, $3)
			ON CONFLICT (token_id, address) DO UPDATE SET balance = $3`,
			tok.ID, string(wallet.Address), amount.String())
		if err != nil {
			return fmt.Errorf("запись баланса %s: %w", coin.Symbol, err)
		}
	}

	// 6. Транзакции о начислении на кошелёк (initial_mint по каждой монете)
	for i, coin := range cfg.Coins {
		amount := new(big.Int)
		amount.SetString(coin.TotalSupply, 10)
		sysTxMint := newSystemTransaction(
			int(bc.Genesis.ID),
			bc.Genesis.Timestamp,
			"initial_mint",
			types.Address("GND_GENESIS"),
			types.Address(wallet.Address),
			amount,
			coin.Symbol,
		)
		if err := saveSystemTransaction(ctx, pool, 2+i, sysTxMint); err != nil {
			return fmt.Errorf("запись системной транзакции initial_mint %s: %w", coin.Symbol, err)
		}
	}

	// 7. Обновляем tx_count генезис-блока в БД (1 genesis + N initial_mint)
	txCount := 1 + len(cfg.Coins)
	if _, err := pool.Exec(ctx, `UPDATE blocks SET tx_count = $1, updated_at = $2 WHERE id = $3`,
		txCount, time.Now().UTC(), bc.Genesis.ID); err != nil {
		return fmt.Errorf("обновление tx_count генезис-блока: %w", err)
	}
	bc.Genesis.TxCount = uint32(txCount)

	return nil
}

// newSystemTransaction создаёт системную транзакцию (генезис, начисление).
func newSystemTransaction(blockID int, ts time.Time, txType string, sender, recipient types.Address, value *big.Int, symbol string) *Transaction {
	if value == nil {
		value = big.NewInt(0)
	}
	tx := &Transaction{
		BlockID:    blockID,
		Sender:     sender,
		Recipient:  recipient,
		Value:      value,
		Fee:        big.NewInt(0),
		Nonce:      0,
		Type:       txType,
		Status:     "confirmed",
		Timestamp:  ts,
		Symbol:     symbol,
		IsVerified: true,
	}
	tx.Hash = tx.CalculateHash()
	return tx
}

// saveSystemTransaction сохраняет системную транзакцию в БД с заданным id (для PK id+timestamp). contract_id — NULL.
func saveSystemTransaction(ctx context.Context, pool *pgxpool.Pool, id int, tx *Transaction) error {
	feeStr := "0"
	if tx.Fee != nil {
		feeStr = tx.Fee.String()
	}
	_, err := pool.Exec(ctx, `
		INSERT INTO transactions (
			id, block_id, hash, sender, recipient, value, fee, nonce,
			type, contract_id, payload, status, timestamp, signature, is_verified
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		ON CONFLICT (id, timestamp) DO NOTHING`,
		id, tx.BlockID, tx.Hash, tx.Sender.String(), tx.Recipient.String(),
		tx.Value.String(), feeStr, tx.Nonce,
		tx.Type, (*int64)(nil), tx.Payload, // contract_id = NULL для системных транзакций
		tx.Status, tx.Timestamp, tx.Signature, tx.IsVerified,
	)
	return err
}

// storeBlock сохраняет блок в таблице blocks: записывает tx_count, created_at, updated_at и возвращает block.ID для привязки транзакций.
func (bc *Blockchain) storeBlock(block *Block) error {
	if bc.Pool == nil {
		return nil
	}
	ctx := context.Background()

	// Количество транзакций и время создания/окончания блока
	block.TxCount = uint32(len(block.Transactions))
	if block.CreatedAt.IsZero() {
		block.CreatedAt = block.Timestamp
	}
	block.UpdatedAt = time.Now()

	err := bc.Pool.QueryRow(ctx, `
		INSERT INTO blocks (index, hash, prev_hash, timestamp, miner, gas_used, gas_limit, consensus, nonce, tx_count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (index) DO UPDATE SET tx_count = EXCLUDED.tx_count, updated_at = EXCLUDED.updated_at
		RETURNING id`,
		block.Index, block.Hash, block.PrevHash, block.Timestamp,
		block.Miner, block.GasUsed, block.GasLimit, block.Consensus, block.Nonce,
		block.TxCount, block.CreatedAt, block.UpdatedAt,
	).Scan(&block.ID)
	if err != nil {
		return err
	}
	return nil
}

// validateBlock проверяет целостность блока
func (bc *Blockchain) validateBlock(block *Block) bool {
	if block.Hash != block.CalculateHash() {
		fmt.Println("Хеш блока не совпадает")
		return false
	}
	// TODO: добавить проверку подписи, времени, консенсуса и уникальности транзакций
	return true
}

// applyBlock применяет все транзакции из блока к состоянию
func (bc *Blockchain) applyBlock(block *Block) {
	for _, tx := range block.Transactions {
		if err := bc.State.ApplyTransaction(tx); err != nil {
			fmt.Printf("Транзакция %s не прошла, пропущена: %v\n", tx.Hash, err)
		}
	}
}

// Height возвращает текущую высоту цепочки
func (bc *Blockchain) Height() uint64 {
	return uint64(len(bc.Blocks) - 1)
}

// AllBlocks возвращает копию всех блоков (для API)
func (bc *Blockchain) AllBlocks() []*Block {
	blocksCopy := make([]*Block, len(bc.Blocks))
	copy(blocksCopy, bc.Blocks)
	return blocksCopy
}

// AddTx добавляет транзакцию в мемпул
func (bc *Blockchain) AddTx(tx *Transaction) error {
	bc.Mempool.Add(tx)
	return nil
}

// GetTxStatus возвращает статус транзакции: confirmed / pending / not found
func (bc *Blockchain) GetTxStatus(hash string) (string, error) {
	for _, block := range bc.Blocks {
		for _, tx := range block.Transactions {
			if tx.Hash == hash {
				return "confirmed", nil
			}
		}
	}

	if bc.Mempool.Exists(hash) {
		return "pending", nil
	}

	return "not found", errors.New("транзакция не найдена")
}

// LatestBlock возвращает последний блок
func (bc *Blockchain) LatestBlock() (*Block, error) {
	if len(bc.Blocks) == 0 {
		return nil, errors.New("no blocks in blockchain")
	}
	return bc.Blocks[len(bc.Blocks)-1], nil
}

// LoadTransactionsForBlock загружает транзакции блока по block_id (blocks.id). Для ответа API (block/latest, block/:number).
func LoadTransactionsForBlock(ctx context.Context, pool *pgxpool.Pool, blockID int64) ([]*Transaction, error) {
	if pool == nil {
		return nil, nil
	}
	rows, err := pool.Query(ctx, `
		SELECT 
			hash, 
			sender, 
			recipient, 
			value, 
			fee, 
			nonce, 
			type, 
			payload, 
			status, 
			timestamp 
		FROM transactions 
		WHERE block_id = $1`, blockID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txs []*Transaction
	for rows.Next() {
		var tx Transaction
		var valueStr string
		var payload []byte

		err := rows.Scan(
			&tx.Hash,
			&tx.Sender,
			&tx.Recipient,
			&valueStr,
			&tx.Fee,
			&tx.Nonce,
			&tx.Type,
			&payload,
			&tx.Status,
			&tx.Timestamp,
		)
		if err != nil {
			return nil, err
		}

		tx.Data = payload
		tx.Value, _ = new(big.Int).SetString(valueStr, 10)
		tx.BlockID = int(blockID)
		txs = append(txs, &tx)
	}
	return txs, nil
}

// loadTransactionsForBlock загружает транзакции для конкретного блока (по block_id = blocks.id)
func loadTransactionsForBlock(ctx context.Context, pool *pgxpool.Pool, blockID int64) ([]*Transaction, error) {
	rows, err := pool.Query(ctx, `
		SELECT 
			hash, 
			sender, 
			recipient, 
			value, 
			fee, 
			nonce, 
			type, 
			payload, 
			status, 
			timestamp 
		FROM transactions 
		WHERE block_id = $1`, blockID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txs []*Transaction
	for rows.Next() {
		var tx Transaction
		var valueStr string
		var payload []byte

		err := rows.Scan(
			&tx.Hash,
			&tx.Sender,
			&tx.Recipient,
			&valueStr,
			&tx.Fee,
			&tx.Nonce,
			&tx.Type,
			&payload,
			&tx.Status,
			&tx.Timestamp,
		)
		if err != nil {
			return nil, err
		}

		tx.Data = payload
		tx.Value, _ = new(big.Int).SetString(valueStr, 10)
		txs = append(txs, &tx)
	}
	return txs, nil
}

// AddBlock добавляет новый блок в цепочку
func (bc *Blockchain) AddBlock(block *Block) error {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	if !bc.validateBlock(block) {
		return errors.New("invalid block")
	}

	// Сохраняем блок в БД (tx_count, created_at, updated_at; block.ID заполняется)
	if err := bc.storeBlock(block); err != nil {
		return fmt.Errorf("failed to store block: %v", err)
	}

	// Привязка транзакций к блоку: сохраняем в БД с block_id = block.ID
	if bc.Pool != nil && block.ID != 0 {
		ctx := context.Background()
		for _, tx := range block.Transactions {
			tx.BlockID = int(block.ID)
			if err := tx.SaveToDB(ctx, bc.Pool); err != nil {
				// Дубликат или ошибка — логируем, не прерываем добавление блока
				fmt.Printf("предупреждение: не удалось сохранить транзакцию %s в блок %d: %v\n", tx.Hash, block.ID, err)
			}
		}
	}

	// Применяем транзакции
	bc.applyBlock(block)

	// Добавляем блок в цепочку
	bc.Blocks = append(bc.Blocks, block)

	return nil
}

// ProcessTransaction обрабатывает транзакцию
func (bc *Blockchain) ProcessTransaction(tx *Transaction) error {
	if err := bc.ValidateTransaction(tx); err != nil {
		return err
	}

	if tx.IsContractCall() {
		return bc.processContract(tx)
	}

	return bc.processTransfer(tx)
}

// processTransfer обрабатывает обычный перевод
func (bc *Blockchain) processTransfer(tx *Transaction) error {
	// Проверяем баланс отправителя
	if !tx.HasSufficientBalance() {
		return errors.New("insufficient balance")
	}

	// Выполняем перевод
	if err := bc.State.SubBalance(types.Address(tx.Sender), "GND", tx.Value); err != nil {
		return err
	}

	if err := bc.State.AddBalance(types.Address(tx.Recipient), "GND", tx.Value); err != nil {
		// Откатываем списание если не удалось начислить
		bc.State.AddBalance(types.Address(tx.Sender), "GND", tx.Value)
		return err
	}

	return nil
}

// processContract обрабатывает вызов контракта
func (bc *Blockchain) processContract(tx *Transaction) error {
	// TODO: Реализовать обработку вызова контракта
	return errors.New("contract calls not implemented")
}

// ValidateTransaction проверяет транзакцию
func (bc *Blockchain) ValidateTransaction(tx *Transaction) error {
	if err := tx.Validate(); err != nil {
		return err
	}

	// Проверяем баланс отправителя
	if !tx.HasSufficientBalance() {
		return errors.New("insufficient balance")
	}

	// Проверяем nonce
	expectedNonce := bc.State.GetNonce(types.Address(tx.Sender))
	if tx.Nonce != expectedNonce {
		return fmt.Errorf("invalid nonce: expected %d, got %d", expectedNonce, tx.Nonce)
	}

	return nil
}

// CreateWallet создает новый кошелек
func (bc *Blockchain) CreateWallet() (*Wallet, error) {
	return NewWallet(bc.Pool)
}

// GetBalance возвращает баланс адреса
func (bc *Blockchain) GetBalance(address string, symbol string) *big.Int {
	return bc.State.GetBalance(types.Address(address), symbol)
}

// SendTransaction отправляет транзакцию
func (bc *Blockchain) SendTransaction(tx *Transaction) (string, error) {
	if err := bc.ProcessTransaction(tx); err != nil {
		return "", err
	}
	return tx.Hash, nil
}

// GetTransaction возвращает транзакцию по хешу
func (bc *Blockchain) GetTransaction(hash string) (*Transaction, error) {
	// Ищем в блоках
	for _, block := range bc.Blocks {
		for _, tx := range block.Transactions {
			if tx.Hash == hash {
				return tx, nil
			}
		}
	}

	// Ищем в мемпуле
	tx, err := bc.Mempool.GetTransaction(hash)
	if err != nil {
		return nil, err
	}
	if tx != nil {
		return tx, nil
	}

	return nil, errors.New("transaction not found")
}

// DeployContract deploys a new contract
func (bc *Blockchain) DeployContract(params *ContractParams) (string, error) {
	// Convert bytecode from hex string to bytes
	bytecode, err := hex.DecodeString(params.Bytecode)
	if err != nil {
		return "", fmt.Errorf("invalid bytecode: %v", err)
	}

	// Create new contract
	contract := NewContract(
		params.From,
		params.Owner,
		bytecode,
		nil, // ABI will be added later
		params.Standard,
		params.Version,
		0, // blockID will be set when block is created
		0, // txID will be set when transaction is created
	)

	// Save contract to database
	if err := contract.SaveToDB(context.Background(), bc.Pool); err != nil {
		return "", fmt.Errorf("failed to save contract: %v", err)
	}

	return contract.Address, nil
}

// GetContract returns contract information by address
func (bc *Blockchain) GetContract(address string) (*Contract, error) {
	return GetContractByAddress(context.Background(), bc.Pool, address)
}

// toStringMap converts map[string]interface{} to map[string]string
func toStringMap(m map[string]interface{}) map[string]string {
	result := make(map[string]string)
	for k, v := range m {
		result[k] = fmt.Sprintf("%v", v)
	}
	return result
}
