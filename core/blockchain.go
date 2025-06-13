package core

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"sync"

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
	// Загружаем генезис-блок
	genesis, err := GetBlockByNumber(pool, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to load genesis block: %v", err)
	}

	return &Blockchain{
		Genesis: genesis,
		State:   NewState(),
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

// FirstLaunch initializes blockchain on first launch
func (bc *Blockchain) FirstLaunch(ctx context.Context, pool *pgxpool.Pool, wallet *Wallet, cfg *Config) error {
	// Сохраняем генезис-блок
	if err := bc.Genesis.SaveToDB(ctx, pool); err != nil {
		return fmt.Errorf("failed to save genesis block: %v", err)
	}

	// Инициализируем состояние
	state := NewState()
	bc.State = state

	// Начисляем начальный баланс для всех монет
	for _, coin := range cfg.Coins {
		amount := new(big.Int)
		amount.SetString(coin.TotalSupply, 10)
		if err := state.Credit(types.Address(wallet.Address), coin.Symbol, amount); err != nil {
			return fmt.Errorf("failed to add initial balance for %s: %v", coin.Symbol, err)
		}
	}

	return nil
}

// storeBlock сохраняет блок в таблице blocks
func (bc *Blockchain) storeBlock(block *Block) error {
	if bc.Pool == nil {
		// В тестах или без БД пропускаем запись
		return nil
	}
	_, err := bc.Pool.Exec(context.Background(), `
		INSERT INTO blocks (index, hash, prev_hash, timestamp, miner, gas_used, gas_limit, consensus, nonce)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (index) DO NOTHING`,
		block.Index, block.Hash, block.PrevHash, block.Timestamp,
		block.Miner, block.GasUsed, block.GasLimit, block.Consensus, block.Nonce,
	)
	return err
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

// loadTransactionsForBlock загружает транзакции для конкретного блока
func loadTransactionsForBlock(ctx context.Context, pool *pgxpool.Pool, blockIndex uint64) ([]*Transaction, error) {
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
		WHERE block_id = $1`, blockIndex)
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

	// Сохраняем блок в БД
	if err := bc.storeBlock(block); err != nil {
		return fmt.Errorf("failed to store block: %v", err)
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
