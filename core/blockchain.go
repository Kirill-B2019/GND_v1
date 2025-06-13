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
	State   StateIface    // Состояние (балансы, аккаунты) теперь интерфейс
	mempool *Mempool      // Пул неподтвержденных транзакций
	mutex   sync.RWMutex  // Mutex для потокобезопасного доступа к блокам
	pool    *pgxpool.Pool // Подключение к БД
}

// NewBlockchain создает новую цепочку с генезис-блоком и сохраняет его в БД
func NewBlockchain(genesis *Block, pool *pgxpool.Pool) *Blockchain {
	bc := &Blockchain{
		blocks:  []*Block{genesis},
		State:   NewState(pool), // NewState возвращает *State, который реализует StateIface
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
	if bc.pool == nil {
		// В тестах или без БД пропускаем запись
		return nil
	}
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
		State:   NewState(pool), // StateIface
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
		return errors.New("неверный блок")
	}

	bc.blocks = append(bc.blocks, block)
	bc.applyBlock(block)

	// Обновляем метрики
	metrics.UpdateBlockMetrics(block)

	// Обновляем метрики транзакций
	for _, tx := range block.Transactions {
		metrics.UpdateTransactionMetrics(tx, "success")
	}

	return bc.storeBlock(block)
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
		contractAddress := coin.ContractAddress
		if contractAddress == "" {
			contractAddress = fmt.Sprintf("GNDct%s", GenerateContractAddress())
			coin.ContractAddress = contractAddress
		}
		token := NewToken(
			contractAddress,
			coin.Symbol,
			coin.Name,
			coin.Decimals,
			coin.TotalSupply,
			string(wallet.Address),
			"gndst1",
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
			VALUES (
				(SELECT t.id FROM tokens t JOIN contracts c ON t.contract_id = c.id WHERE c.address = $1),
				$2,
				$3
			)
			ON CONFLICT (token_id, address) DO UPDATE
			SET balance = EXCLUDED.balance`,
			contractAddress, string(wallet.Address), amount.String())
		if err != nil {
			return fmt.Errorf("ошибка создания баланса токена: %w", err)
		}
	}

	return nil
}

// ProcessTransaction обрабатывает транзакцию
func (b *Blockchain) ProcessTransaction(tx *Transaction) error {
	// Проверка баланса отправителя
	senderBalance := b.State.GetBalance(tx.GetSenderAddress(), tx.Symbol)
	if senderBalance.Cmp(tx.Value) < 0 {
		return fmt.Errorf("insufficient balance")
	}

	// Проверка комиссии
	if tx.Fee == nil || tx.Fee.Sign() <= 0 {
		return fmt.Errorf("invalid fee")
	}

	// Проверка типа транзакции
	switch tx.Type {
	case "transfer":
		return b.processTransfer(tx)
	case "contract":
		return b.processContract(tx)
	default:
		return fmt.Errorf("unknown transaction type: %s", tx.Type)
	}
}

// processTransfer обрабатывает транзакцию перевода
func (b *Blockchain) processTransfer(tx *Transaction) error {
	// Проверка баланса отправителя
	senderBalance := b.State.GetBalance(tx.GetSenderAddress(), tx.Symbol)
	if senderBalance.Cmp(tx.Value) < 0 {
		return fmt.Errorf("insufficient balance")
	}

	// Обновление баланса отправителя
	b.State.Credit(tx.GetSenderAddress(), tx.Symbol, new(big.Int).Neg(tx.Value))

	// Обновление баланса получателя
	b.State.Credit(tx.GetRecipientAddress(), tx.Symbol, tx.Value)

	return nil
}

// processContract обрабатывает транзакцию контракта
func (b *Blockchain) processContract(tx *Transaction) error {
	// TODO: Реализовать обработку контрактных транзакций
	return fmt.Errorf("contract transactions not implemented yet")
}

// ValidateTransaction проверяет валидность транзакции
func (b *Blockchain) ValidateTransaction(tx *Transaction) error {
	// Проверка подписи
	if !tx.Verify() {
		return fmt.Errorf("invalid signature")
	}

	// Проверка баланса отправителя
	senderBalance := b.State.GetBalance(tx.GetSenderAddress(), tx.Symbol)
	if senderBalance.Cmp(tx.Value) < 0 {
		return fmt.Errorf("insufficient balance")
	}

	// Проверка комиссии
	if tx.Fee == nil || tx.Fee.Sign() <= 0 {
		return fmt.Errorf("invalid fee")
	}

	// Проверка nonce
	expectedNonce := b.State.GetNonce(tx.GetSenderAddress())
	if tx.Nonce != expectedNonce {
		return fmt.Errorf("invalid nonce: expected %d, got %d", expectedNonce, tx.Nonce)
	}

	return nil
}
