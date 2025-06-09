package core

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/big"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Blockchain - основная структура цепочки блоков
type Blockchain struct {
	blocks  []*Block      // Все блоки цепочки
	State   *State        // Состояние (балансы, аккаунты)
	mempool *Mempool      // Пул неподтвержденных транзакций
	mutex   sync.RWMutex  // Mutex для потокобезопасного доступа к блокам
	pool    *pgxpool.Pool // Подключение к БД
}

// NewBlockchain создает новую цепочку с генезис-блоком и сохраняет его в БД
func NewBlockchain(genesis *Block, pool *pgxpool.Pool) *Blockchain {
	bc := &Blockchain{
		blocks:  []*Block{genesis},
		State:   NewState(pool),
		mempool: NewMempool(),
		pool:    pool,
	}

	// Сохраняем генезис-блок в БД
	err := bc.storeBlock(genesis)
	if err != nil {
		log.Printf("Не удалось сохранить генезис-блок: %v", err)
	}

	// Применяем транзакции из генезис-блока
	bc.applyBlock(genesis)
	return bc
}

// storeBlock сохраняет блок в таблице blocks
func (bc *Blockchain) storeBlock(block *Block) error {
	_, err := bc.pool.Exec(context.Background(), `
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

// GetBlockByHash находит блок по хешу
func (bc *Blockchain) GetBlockByHash(hash string) *Block {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	for _, b := range bc.blocks {
		if b.Hash == hash {
			return b
		}
	}
	return nil
}

// GetBlockByNumber возвращает блок по номеру (индексу)
func (bc *Blockchain) GetBlockByNumber(number uint64) *Block {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	if number >= uint64(len(bc.blocks)) {
		return nil
	}
	return bc.blocks[number]
}

// Height возвращает текущую высоту цепочки
func (bc *Blockchain) Height() uint64 {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()
	return uint64(len(bc.blocks) - 1)
}

// AllBlocks возвращает копию всех блоков (для API)
func (bc *Blockchain) AllBlocks() []*Block {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	blocksCopy := make([]*Block, len(bc.blocks))
	copy(blocksCopy, bc.blocks)
	return blocksCopy
}

// AddTx добавляет транзакцию в мемпул
func (bc *Blockchain) AddTx(tx *Transaction) error {
	bc.mempool.Add(tx)
	return nil
}

// GetTxStatus возвращает статус транзакции: confirmed / pending / not found
func (bc *Blockchain) GetTxStatus(hash string) (string, error) {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	for _, block := range bc.blocks {
		for _, tx := range block.Transactions {
			if tx.Hash == hash {
				return "confirmed", nil
			}
		}
	}

	if bc.mempool.Exists(hash) {
		return "pending", nil
	}

	return "not found", errors.New("транзакция не найдена")
}

// LatestBlock возвращает последний блок
func (bc *Blockchain) LatestBlock() *Block {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	if len(bc.blocks) == 0 {
		return nil
	}
	return bc.blocks[len(bc.blocks)-1]
}

// LoadBlockchainFromDB загружает блокчейн из базы данных
func LoadBlockchainFromDB(pool *pgxpool.Pool) (*Blockchain, error) {
	ctx := context.Background()
	bc := &Blockchain{
		blocks:  []*Block{},
		State:   NewState(pool),
		mempool: NewMempool(),
		pool:    pool,
	}

	// Загрузка блоков
	rows, err := pool.Query(ctx, `
		SELECT index, hash, prev_hash, timestamp, miner, gas_used, gas_limit, consensus, nonce 
		FROM blocks ORDER BY index ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var b Block
		err := rows.Scan(
			&b.Index, &b.Hash, &b.PrevHash, &b.Timestamp,
			&b.Miner, &b.GasUsed, &b.GasLimit, &b.Consensus, &b.Nonce,
		)
		if err != nil {
			return nil, err
		}
		b.Transactions, err = loadTransactionsForBlock(ctx, pool, b.Index)
		if err != nil {
			return nil, err
		}
		bc.blocks = append(bc.blocks, &b)
	}

	// Применяем транзакции ко всем блокам
	for _, block := range bc.blocks {
		bc.applyBlock(block)
	}

	return bc, nil
}

// loadTransactionsForBlock загружает транзакции для конкретного блока
func loadTransactionsForBlock(ctx context.Context, pool *pgxpool.Pool, blockIndex uint64) ([]*Transaction, error) {
	rows, err := pool.Query(ctx, `
		SELECT hash, from_address, to_address, symbol, value, gas_price, gas_limit, nonce, data, type, signature 
		FROM transactions 
		WHERE block_index = $1`, blockIndex)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txs []*Transaction
	for rows.Next() {
		var tx Transaction
		var data []byte
		var valueStr string

		err := rows.Scan(
			&tx.Hash,
			&tx.From,
			&tx.To,
			&tx.Symbol,
			&valueStr,
			&tx.GasPrice,
			&tx.GasLimit,
			&tx.Nonce,
			&data,
			&tx.Type,
			&tx.Signature,
		)
		if err != nil {
			return nil, err
		}

		tx.Data = data
		tx.Value, _ = new(big.Int).SetString(valueStr, 10)
		txs = append(txs, &tx)
	}
	return txs, nil
}

func (bc *Blockchain) AddBlock(block *Block) error {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	if !bc.validateBlock(block) {
		return errors.New("блок не прошел валидацию")
	}

	bc.blocks = append(bc.blocks, block)
	bc.applyBlock(block)

	return nil
}

// FirstLaunch выполняет инициализацию блокчейна при первом запуске
func (bc *Blockchain) FirstLaunch(ctx context.Context, pool *pgxpool.Pool, wallet *Wallet, cfg *Config) error {
	// Создаем генезис-блок
	genesis := NewBlock("", 0, string(wallet.Address))
	genesis.Hash = genesis.CalculateHash()
	genesis.Consensus = "pos"
	genesis.Index = 0

	// Создаем токены из конфигурации
	for _, coin := range cfg.Coins {
		token := NewToken(
			coin.ContractAddress,
			coin.Symbol,
			coin.Name,
			coin.Decimals,
			coin.TotalSupply,
			string(wallet.Address),
			"ERC20",
			coin.Standard,
			genesis.ID,
			0,
		)

		if err := token.SaveToDB(ctx, pool); err != nil {
			return fmt.Errorf("ошибка создания токена %s: %w", coin.Symbol, err)
		}

		// Создаем начальный баланс
		amount := new(big.Int)
		if coin.TotalSupply != "" {
			amount.SetString(coin.TotalSupply, 10)
		} else {
			amount.SetInt64(1_000_000)
			multiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(coin.Decimals)), nil)
			amount = amount.Mul(amount, multiplier)
		}

		_, err := pool.Exec(ctx, `
			INSERT INTO token_balances (token_id, address, balance)
			SELECT id, $1, $2
			FROM tokens WHERE address = $3
			ON CONFLICT (token_id, address) DO UPDATE
			SET balance = $2`,
			string(wallet.Address), amount.String(), coin.ContractAddress)
		if err != nil {
			return fmt.Errorf("ошибка создания баланса токена: %w", err)
		}
	}

	return nil
}
