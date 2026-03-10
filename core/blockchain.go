// | KB @CerbeRus - Nexus Invest Team
package core

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"

	"GND/core/crypto"
	"GND/types"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Blockchain represents the blockchain structure
type Blockchain struct {
	Genesis       *Block
	State         StateIface
	Pool          *pgxpool.Pool
	Blocks        []*Block
	Mempool       *Mempool
	mutex         sync.Mutex
	SignerCreator SignerWalletCreator // опционально: для создания кошельков через signing_service
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

// LoadBlockchainFromDB loads blockchain from database.
// Загружается вся цепочка блоков (ORDER BY index ASC), чтобы после перезапуска ноды (Ctrl+C и снова go run) последний блок считался и цепь продолжалась с следующего.
// Если передан cfg с NativeContracts (адреса контрактов GND/GANI), перед загрузкой состояния включается режим «всё на контрактах».
func LoadBlockchainFromDB(pool *pgxpool.Pool, configOptional ...*Config) (*Blockchain, error) {
	ctx := context.Background()
	genesis, err := GetBlockByNumber(pool, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to load genesis block: %v", err)
	}

	state := NewState()
	state.SetPool(pool)
	if len(configOptional) > 0 && configOptional[0] != nil && configOptional[0].NativeContracts != nil {
		nc := configOptional[0].NativeContracts
		state.SetNativeContractAddresses(nc.GndContractAddress, nc.GaniContractAddress)
		if nc.GndselfAddress != "" {
			state.SetGndselfAddress(nc.GndselfAddress)
		}
	}
	if err := state.LoadFromDB(ctx); err != nil {
		return nil, fmt.Errorf("failed to load state from DB: %w", err)
	}

	// Цепочка блоков по порядку index (генезис, 1, 2, …) — чтобы продолжить с последнего после перезапуска
	blocks, err := LoadChainBlocksOrderedByIndex(pool, 10_000_000)
	if err != nil {
		return nil, fmt.Errorf("failed to load chain blocks: %w", err)
	}
	if len(blocks) == 0 {
		blocks = []*Block{genesis}
	}

	return &Blockchain{
		Genesis: genesis,
		State:   state,
		Pool:    pool,
		Blocks:  blocks,
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
// Для уже существующих токенов обновляет circulating_supply из конфига.
// genesisBlockID — ID генезис-блока в БД для заполнения contracts.block_id и contracts.tx_id.
func EnsureCoinsDeployed(ctx context.Context, pool *pgxpool.Pool, cfg *Config, ownerAddress string, genesisBlockID int) error {
	for _, coin := range cfg.Coins {
		circulating := coin.CirculatingSupply
		if circulating == "" {
			circulating = coin.TotalSupply
		}
		_, err := GetTokenBySymbol(ctx, pool, coin.Symbol)
		if err == nil {
			// Токен уже есть — обновляем decimals, total_supply, circulating_supply и logo_url из конфига
			if coin.CoinLogo != "" {
				_, _ = pool.Exec(ctx, `UPDATE public.tokens SET decimals = $1, total_supply = $2, circulating_supply = $3, logo_url = $4, updated_at = $5 WHERE symbol = $6`,
					coin.Decimals, coin.TotalSupply, circulating, strings.TrimSpace(coin.CoinLogo), BlockchainNow(), coin.Symbol)
			} else {
				_, _ = pool.Exec(ctx, `UPDATE public.tokens SET decimals = $1, total_supply = $2, circulating_supply = $3, updated_at = $4 WHERE symbol = $5`,
					coin.Decimals, coin.TotalSupply, circulating, BlockchainNow(), coin.Symbol)
			}
			continue
		}
		addr := coin.ContractAddress
		if addr == "" || len(addr) < 10 {
			addr = ""
		}
		t := NewToken(addr, coin.Symbol, coin.Name, coin.Decimals, coin.TotalSupply, circulating, ownerAddress, "coin", coin.Standard, genesisBlockID, 0)
		t.IsVerified = true
		t.LogoURL = strings.TrimSpace(coin.CoinLogo)
		if err := t.SaveToDB(ctx, pool); err != nil {
			return fmt.Errorf("деплой монеты %s: %w", coin.Symbol, err)
		}
		// Пометить контракт нативной монеты как верифицированный
		_, _ = pool.Exec(ctx, `UPDATE contracts SET is_verified = true WHERE address = $1`, t.Address)
	}
	return nil
}

// genesisTimestamp задаётся в core/time.go (Москва). Записи в transactions используют BlockchainNow().

// EnsureGenesisTransactions при загрузке из БД дописывает системные транзакции в transactions, если для генезис-блока их ещё нет.
// Timestamp записей — BlockchainNow() (Москва), чтобы строка попала в текущую партицию transactions (RANGE по timestamp).
func EnsureGenesisTransactions(ctx context.Context, pool *pgxpool.Pool, bc *Blockchain, wallet *Wallet, cfg *Config) error {
	if pool == nil || bc == nil || bc.Genesis == nil {
		return nil
	}
	var count int
	err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM transactions WHERE block_id = $1`, bc.Genesis.ID).Scan(&count)
	if err != nil || count > 0 {
		return err
	}
	ts := BlockchainNow()
	sysTxGenesis := newSystemTransaction(
		int(bc.Genesis.ID),
		ts,
		"genesis",
		types.Address("GND_GENESIS"),
		types.Address(wallet.Address),
		big.NewInt(0),
		"",
	)
	if err := saveSystemTransaction(ctx, pool, 1, sysTxGenesis); err != nil {
		return fmt.Errorf("дозапись системной транзакции genesis: %w", err)
	}
	for i, coin := range cfg.Coins {
		amountStr := coin.CirculatingSupply
		if amountStr == "" {
			amountStr = coin.TotalSupply
		}
		amount := new(big.Int)
		amount.SetString(amountStr, 10)
		sysTxMint := newSystemTransaction(
			int(bc.Genesis.ID),
			ts,
			"initial_mint",
			types.Address("GND_GENESIS"),
			types.Address(wallet.Address),
			amount,
			coin.Symbol,
		)
		if err := saveSystemTransaction(ctx, pool, 2+i, sysTxMint); err != nil {
			return fmt.Errorf("дозапись системной транзакции initial_mint %s: %w", coin.Symbol, err)
		}
	}
	txCount := 1 + len(cfg.Coins)
	_, _ = pool.Exec(ctx, `UPDATE blocks SET tx_count = $1, updated_at = $2 WHERE id = $3`,
		txCount, BlockchainNow(), bc.Genesis.ID)
	bc.Genesis.TxCount = uint32(txCount)
	return nil
}

// FirstLaunch инициализирует блокчейн при первом запуске (пустая БД).
// Создаётся только генезис-блок, одна нулевая транзакция; блок финализируется.
// Кошелёк не создаётся, эмиссии и токены не разворачиваются.
func (bc *Blockchain) FirstLaunch(ctx context.Context, pool *pgxpool.Pool, wallet *Wallet, cfg *Config) error {
	// 1. Генезис: сохраняем блок (финализирован)
	bc.Genesis.Timestamp = genesisTimestamp
	bc.Genesis.CreatedAt = genesisTimestamp
	bc.Genesis.UpdatedAt = BlockchainNow()
	bc.Genesis.Status = "finalized"
	bc.Genesis.Hash = bc.Genesis.CalculateHash()

	if err := bc.Genesis.SaveToDB(ctx, pool); err != nil {
		return fmt.Errorf("failed to save genesis block: %v", err)
	}

	// 2. Одна нулевая транзакция генезиса (без эмиссий и токенов)
	recipient := types.Address("GND_GENESIS")
	if wallet != nil {
		recipient = types.Address(wallet.Address)
	}
	// Timestamp записи — текущее время, чтобы попасть в существующую партицию transactions
	sysTxGenesis := newSystemTransaction(
		int(bc.Genesis.ID),
		BlockchainNow(),
		"genesis",
		types.Address("GND_GENESIS"),
		recipient,
		big.NewInt(0),
		"",
	)
	if err := saveSystemTransaction(ctx, pool, 1, sysTxGenesis); err != nil {
		return fmt.Errorf("запись системной транзакции genesis: %w", err)
	}

	// 3. Обновляем tx_count генезис-блока (только одна транзакция)
	bc.Genesis.TxCount = 1
	if _, err := pool.Exec(ctx, `UPDATE blocks SET tx_count = 1, updated_at = $1 WHERE id = $2`,
		BlockchainNow(), bc.Genesis.ID); err != nil {
		return fmt.Errorf("обновление tx_count генезис-блока: %w", err)
	}

	// 4. Инициализируем состояние (без начисления балансов)
	state := NewState()
	state.SetPool(pool)
	bc.State = state

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

// RecordAdminTransaction записывает в БД транзакцию административного действия (создание кошелька, деплой контракта, верификация контракта, создание API-ключа).
// Используется блок генезиса (block_id). Если genesisBlockID <= 0, берётся id из blocks WHERE index = 0.
func RecordAdminTransaction(ctx context.Context, pool *pgxpool.Pool, genesisBlockID int64, txType, sender, recipient, payload string) error {
	if pool == nil {
		return nil
	}
	blockID := genesisBlockID
	if blockID <= 0 {
		if err := pool.QueryRow(ctx, `SELECT id FROM blocks WHERE index = 0 LIMIT 1`).Scan(&blockID); err != nil {
			return fmt.Errorf("genesis block not found: %w", err)
		}
	}
	ts := BlockchainNow()
	tx := &Transaction{
		BlockID:    int(blockID),
		Sender:     types.Address(sender),
		Recipient:  types.Address(recipient),
		Value:      big.NewInt(0),
		Fee:        big.NewInt(0),
		Nonce:      0,
		Type:       txType,
		Status:     "confirmed",
		Timestamp:  ts,
		Payload:    []byte(payload),
		IsVerified: true,
	}
	tx.Hash = tx.CalculateHash()
	return tx.SaveToDB(ctx, pool)
}

// saveSystemTransaction сохраняет системную транзакцию в БД с заданным id (для PK id+timestamp). contract_id в SQL = NULL.
// signature и payload передаём в безопасном виде (hex/JSON), чтобы избежать ошибки 22P05 (unsupported Unicode escape).
func saveSystemTransaction(ctx context.Context, pool *pgxpool.Pool, id int, tx *Transaction) error {
	feeStr := "0"
	if tx.Fee != nil {
		feeStr = tx.Fee.String()
	}
	var signatureArg interface{}
	if len(tx.Signature) == 0 {
		signatureArg = nil
	} else {
		signatureArg = hex.EncodeToString(tx.Signature)
	}
	var payloadArg interface{}
	if len(tx.Payload) == 0 {
		payloadArg = nil
	} else if json.Valid(tx.Payload) {
		payloadArg = json.RawMessage(tx.Payload)
	} else {
		payloadArg = map[string]string{"hex": hex.EncodeToString(tx.Payload)}
	}
	_, err := pool.Exec(ctx, `
		INSERT INTO transactions (
			id, block_id, hash, sender, recipient, value, fee, nonce,
			type, contract_id, payload, status, timestamp, signature, is_verified
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NULL, $10, $11, $12, $13, $14)
		ON CONFLICT (id, timestamp) DO NOTHING`,
		id, tx.BlockID, tx.Hash, tx.Sender.String(), tx.Recipient.String(),
		tx.Value.String(), feeStr, tx.Nonce,
		tx.Type, payloadArg,
		tx.Status, tx.Timestamp, signatureArg, tx.IsVerified,
	)
	return err
}

// storeBlock сохраняет блок в таблице blocks. created_at = время создания блока, updated_at = время финализации.
func (bc *Blockchain) storeBlock(block *Block) error {
	if bc.Pool == nil {
		return nil
	}
	ctx := context.Background()

	block.TxCount = uint32(len(block.Transactions))
	// created_at — когда блок создан (временная метка блока), updated_at — когда финализирован (сейчас)
	block.CreatedAt = block.Timestamp
	block.UpdatedAt = BlockchainNow()

	// nonce в БД — varchar, передаём строку. is_finalized = true для финализированных блоков.
	// height записываем для GET /api/v1/block/:number (поиск по height в цепи).
	isFinalized := block.Status == "finalized"
	nonceStr := strconv.FormatUint(block.Nonce, 10)
	err := bc.Pool.QueryRow(ctx, `
		INSERT INTO blocks (index, height, hash, prev_hash, merkle_root, timestamp, miner, gas_used, gas_limit, consensus, nonce, tx_count, created_at, updated_at, is_finalized)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		ON CONFLICT (index) DO UPDATE SET height = EXCLUDED.height, tx_count = EXCLUDED.tx_count, merkle_root = EXCLUDED.merkle_root, updated_at = EXCLUDED.updated_at, is_finalized = EXCLUDED.is_finalized
		RETURNING id`,
		block.Index, block.Height, block.Hash, block.PrevHash, block.MerkleRoot, block.Timestamp,
		block.Miner, block.GasUsed, block.GasLimit, block.Consensus, nonceStr,
		block.TxCount, block.CreatedAt, block.UpdatedAt, isFinalized,
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

// applyBlock применяет все транзакции из блока к состоянию.
// Для contract_call строит результат из calldata (storage) и вызывает ApplyExecutionResult,
// чтобы изменения контракта попали в contract_storage при SaveToDB.
func (bc *Blockchain) applyBlock(block *Block) {
	if st, ok := bc.State.(*State); ok {
		st.ClearTouched()
	}
	for _, tx := range block.Transactions {
		if tx == nil {
			continue
		}
		if tx.IsContractCall() {
			result := buildContractCallExecutionResult(tx)
			if result != nil {
				if st, ok := bc.State.(*State); ok {
					expected := bc.State.GetNonce(types.Address(tx.Sender))
					if int64(tx.Nonce) != expected {
						fmt.Printf("Транзакция %s: неверный nonce (expected %d, got %d), пропущена\n", tx.Hash, expected, tx.Nonce)
						continue
					}
					symbol := "GND"
					if tx.Symbol != "" {
						symbol = tx.Symbol
					}
					if tx.Value != nil && tx.Value.Sign() > 0 {
						if err := st.SubBalance(tx.Sender, symbol, tx.Value); err != nil {
							fmt.Printf("Транзакция %s не прошла (value): %v\n", tx.Hash, err)
							continue
						}
						if err := st.AddBalance(tx.Recipient, symbol, tx.Value); err != nil {
							st.AddBalance(tx.Sender, symbol, tx.Value)
							fmt.Printf("Транзакция %s не прошла (value): %v\n", tx.Hash, err)
							continue
						}
					}
					if err := st.ApplyExecutionResult(tx, result); err != nil {
						fmt.Printf("Транзакция %s не прошла (ApplyExecutionResult): %v\n", tx.Hash, err)
					}
					continue
				}
			}
		}
		if err := bc.State.ApplyTransaction(tx); err != nil {
			exp := bc.State.GetNonce(types.Address(tx.Sender))
			fmt.Printf("Транзакция %s не прошла, пропущена: %v (sender nonce в tx: %d, expected: %d)\n", tx.Hash, err, tx.Nonce, exp)
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

// ProduceNextBlock создаёт новый блок, включая до maxTxs транзакций из mempool, и добавляет его в цепь.
// При открытии нового блока предыдущий (текущий tip) финализируется, даже если он пустой.
// miner — адрес валидатора. Вызывается по таймеру (например из main с интервалом round_duration).
func (bc *Blockchain) ProduceNextBlock(mempool *Mempool, miner string, maxTxs int) error {
	if mempool == nil {
		fmt.Println("[BlockProducer] mempool == nil, блок не создаётся")
		return nil
	}
	last, err := bc.LatestBlock()
	if err != nil || last == nil {
		return err
	}
	// Финализировать предыдущий блок при открытии нового (даже если он пустой)
	if last.Status != "finalized" && bc.Pool != nil && last.ID != 0 {
		ctx := context.Background()
		if err := last.UpdateStatus(ctx, bc.Pool, "finalized"); err != nil {
			return fmt.Errorf("финализация предыдущего блока %d: %w", last.Index, err)
		}
	}
	last.Status = "finalized"

	height := last.Index + 1
	prevHash := last.Hash
	if prevHash == "" {
		prevHash = "0"
	}
	block := NewBlock(prevHash, height, miner)
	block.Index = height
	block.Reward = big.NewInt(0)
	block.GasLimit = 10_000_000
	block.GasUsed = 0
	block.Consensus = "poa"
	block.Status = "finalized"
	block.IsFinalized = true

	rawTxs := mempool.TakePending(maxTxs)
	// Фильтруем транзакции с неверным nonce — не включаем в блок, чтобы не спамить "invalid nonce" при каждом applyBlock
	var txs []*Transaction
	for _, tx := range rawTxs {
		expected := bc.State.GetNonce(types.Address(tx.Sender))
		if int64(tx.Nonce) != expected {
			fmt.Printf("[Mempool] Транзакция %s не включена в блок: invalid nonce (expected %d, got %d)\n", tx.Hash, expected, tx.Nonce)
			continue
		}
		txs = append(txs, tx)
	}
	block.Transactions = txs
	block.TxCount = uint32(len(txs))
	block.MerkleRoot = ComputeMerkleRoot(txs)

	block.Hash = block.CalculateHash()

	if err := bc.AddBlock(block); err != nil {
		return fmt.Errorf("ProduceNextBlock AddBlock: %w", err)
	}
	return nil
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

// LoadPendingTransactionsFromDB загружает все транзакции с block_id IS NULL (ожидающие включения в блок).
// Используется при старте ноды, чтобы после перезапуска они попали в мемпул и в следующий блок.
func LoadPendingTransactionsFromDB(ctx context.Context, pool *pgxpool.Pool) ([]*Transaction, error) {
	if pool == nil {
		return nil, nil
	}
	rows, err := pool.Query(ctx, `
		SELECT block_id, hash, sender, recipient, value, fee, nonce, type, payload, status, timestamp, contract_id
		FROM transactions
		WHERE block_id IS NULL
		ORDER BY timestamp ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*Transaction
	for rows.Next() {
		var tx Transaction
		var valueStr, feeStr string
		var payload []byte
		var blockIDNull sql.NullInt64
		var contractIDNull sql.NullInt64
		if err := rows.Scan(
			&blockIDNull,
			&tx.Hash,
			&tx.Sender,
			&tx.Recipient,
			&valueStr,
			&feeStr,
			&tx.Nonce,
			&tx.Type,
			&payload,
			&tx.Status,
			&tx.Timestamp,
			&contractIDNull,
		); err != nil {
			continue
		}
		tx.Data = payload
		if len(tx.Payload) == 0 {
			tx.Payload = payload
		}
		if blockIDNull.Valid {
			tx.BlockID = int(blockIDNull.Int64)
		}
		tx.ContractID = contractIDNull
		tx.Value, _ = new(big.Int).SetString(valueStr, 10)
		if tx.Value == nil {
			tx.Value = big.NewInt(0)
		}
		tx.Fee, _ = new(big.Int).SetString(feeStr, 10)
		if tx.Fee == nil {
			tx.Fee = big.NewInt(0)
		}
		list = append(list, &tx)
	}
	return list, rows.Err()
}

// LoadTransactionByHash загружает транзакцию из gnd_db.transactions по хешу.
// При нескольких записях (pending и confirmed) возвращаем подтверждённую (с block_id).
// Загружает signature и is_verified, чтобы API возвращал корректные значения.
func LoadTransactionByHash(ctx context.Context, pool *pgxpool.Pool, hash string) (*Transaction, error) {
	if pool == nil || hash == "" {
		return nil, errors.New("pool or hash empty")
	}
	var tx Transaction
	var valueStr, feeStr string
	var payload []byte
	var blockIDNull sql.NullInt64
	var contractIDNull sql.NullInt64
	var sigStr sql.NullString
	var isVerifiedNull sql.NullBool
	err := pool.QueryRow(ctx, `
		SELECT block_id, hash, sender, recipient, value, fee, nonce, type, payload, status, timestamp, contract_id, signature, is_verified
		FROM transactions
		WHERE hash = $1
		ORDER BY block_id DESC NULLS LAST
		LIMIT 1`, hash).Scan(
		&blockIDNull,
		&tx.Hash,
		&tx.Sender,
		&tx.Recipient,
		&valueStr,
		&feeStr,
		&tx.Nonce,
		&tx.Type,
		&payload,
		&tx.Status,
		&tx.Timestamp,
		&contractIDNull,
		&sigStr,
		&isVerifiedNull,
	)
	if err != nil {
		return nil, err
	}
	tx.Data = payload
	if blockIDNull.Valid {
		tx.BlockID = int(blockIDNull.Int64)
	}
	tx.ContractID = contractIDNull
	tx.Value, _ = new(big.Int).SetString(valueStr, 10)
	tx.Fee, _ = new(big.Int).SetString(feeStr, 10)
	tx.IsVerified = isVerifiedNull.Valid && isVerifiedNull.Bool
	if sigStr.Valid && sigStr.String != "" {
		tx.Signature, _ = hex.DecodeString(strings.TrimPrefix(strings.TrimPrefix(sigStr.String, "0x"), "0X"))
	}
	return &tx, nil
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

	// Обновляем существующую запись транзакции (pending): выставляем block_id, status и contract_id по recipient
	if bc.Pool != nil && block.ID != 0 {
		ctx := context.Background()
		for _, tx := range block.Transactions {
			if tx.Timestamp.IsZero() {
				tx.Timestamp = block.Timestamp
			}
			tx.BlockID = int(block.ID)
			if _, err := bc.Pool.Exec(ctx, `
				UPDATE transactions SET block_id = $1, status = 'confirmed',
					contract_id = (SELECT id FROM contracts WHERE address = $2 LIMIT 1)
				WHERE hash = $3 AND block_id IS NULL`,
				block.ID, tx.Recipient.String(), tx.Hash); err != nil {
				fmt.Printf("предупреждение: не удалось обновить транзакцию %s для блока %d: %v\n", tx.Hash, block.ID, err)
			}
		}
	}

	// Применяем транзакции к состоянию
	bc.applyBlock(block)

	// Сохраняем состояние (accounts, account_states, contract_storage при blockID > 0)
	if bc.Pool != nil && bc.State != nil {
		blockID := int64(0)
		if block != nil && block.ID != 0 {
			blockID = int64(block.ID)
		}
		if err := bc.State.SaveToDB(blockID); err != nil {
			fmt.Printf("предупреждение: не удалось сохранить состояние после блока %d: %v\n", block.ID, err)
		}
		if st, ok := bc.State.(*State); ok {
			// state_root — хеш состояния после применения блока
			block.StateRoot = st.RootHash()
			if block.ID != 0 {
				_, _ = bc.Pool.Exec(context.Background(), `UPDATE blocks SET state_root = $1 WHERE id = $2`, block.StateRoot, block.ID)
			}
			st.ClearTouched()
		}
	}

	// Добавляем блок в цепочку
	bc.Blocks = append(bc.Blocks, block)

	// Обновляем метрики блоков и транзакций для актуальных значений в API
	if m := GetMetrics(); m != nil {
		m.UpdateBlockMetrics(block)
		for _, tx := range block.Transactions {
			m.UpdateTransactionMetrics(tx, "confirmed")
		}
	}

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

// processTransfer обрабатывает перевод нативной монеты (GND или GANI)
func (bc *Blockchain) processTransfer(tx *Transaction) error {
	if !tx.HasSufficientBalance() {
		return errors.New("insufficient balance")
	}
	symbol := tx.Symbol
	if symbol == "" {
		symbol = GasSymbol
	}
	if !IsNativeSymbol(symbol) {
		return errors.New("symbol must be GND or GANI for native transfer")
	}
	if err := bc.State.SubBalance(types.Address(tx.Sender), symbol, tx.Value); err != nil {
		return err
	}
	if err := bc.State.AddBalance(types.Address(tx.Recipient), symbol, tx.Value); err != nil {
		bc.State.AddBalance(types.Address(tx.Sender), symbol, tx.Value)
		return err
	}
	return nil
}

// processContract обрабатывает вызов контракта: добавляет транзакцию в мемпул и записывает в БД (таблица transactions),
// чтобы все вызовы (transfer, approve и т.д.) включались в транзакции блокчейна.
func (bc *Blockchain) processContract(tx *Transaction) error {
	if tx.Type == "" {
		tx.Type = "contract_call"
	}
	if tx.Status == "" {
		tx.Status = "pending"
	}
	tx.BlockID = 0
	if len(tx.Data) > 0 && len(tx.Payload) == 0 {
		tx.Payload = tx.Data
	}
	if tx.Hash == "" {
		tx.Hash = tx.CalculateHash()
	}
	if bc.Mempool != nil {
		bc.Mempool.Add(tx)
	}
	if bc.Pool != nil {
		if err := tx.SaveToDB(context.Background(), bc.Pool); err != nil {
			return fmt.Errorf("сохранение транзакции вызова контракта: %w", err)
		}
	}
	return nil
}

// ValidateTransaction проверяет транзакцию (формат, подпись, баланс, nonce).
func (bc *Blockchain) ValidateTransaction(tx *Transaction) error {
	if err := tx.Validate(); err != nil {
		return err
	}

	// Проверка подписи для пользовательских транзакций (системные пропускаем).
	if !IsSystemTransaction(tx) {
		if tx.Hash == "" {
			tx.Hash = tx.CalculateHash()
		}
		if len(tx.Signature) == 0 {
			return errors.New("транзакция должна быть подписана (signature обязателен)")
		}
		if tx.SenderPublicKeyHex == "" {
			return errors.New("для проверки подписи укажите sender_public_key (hex публичного ключа P-256)")
		}
		pubKey, err := crypto.ParsePublicKeyHex(tx.SenderPublicKeyHex)
		if err != nil {
			return fmt.Errorf("неверный sender_public_key: %w", err)
		}
		if !VerifyTransactionSignature(tx, pubKey) {
			return errors.New("неверная подпись транзакции")
		}
		// Проверка соответствия адреса ключу (формат 64 hex)
		if tx.Sender.IsValid() {
			derivedAddr := crypto.PublicKeyToAddressP256(pubKey)
			if derivedAddr != tx.Sender.String() {
				return errors.New("адрес отправителя не соответствует публичному ключу")
			}
		}
	}

	// Проверяем баланс отправителя. Для вызова контракта с владельцем gndself газ не списывается — не требуем баланс на газ.
	if bc.State != nil && bc.State.WillSkipGasForTx(tx) {
		if tx.Value != nil && tx.Value.Sign() > 0 {
			symbol := tx.Symbol
			if symbol == "" {
				symbol = "GND"
			}
			balance := bc.State.GetBalance(types.Address(tx.Sender), symbol)
			if balance.Cmp(tx.Value) < 0 {
				return errors.New("insufficient balance")
			}
		}
	} else if !tx.HasSufficientBalance() {
		return errors.New("insufficient balance")
	}

	// Проверяем nonce
	expectedNonce := bc.State.GetNonce(types.Address(tx.Sender))
	if tx.Nonce != expectedNonce {
		return fmt.Errorf("invalid nonce: expected %d, got %d", expectedNonce, tx.Nonce)
	}

	return nil
}

// CreateWallet создает новый кошелек. Если задан SignerCreator — ключ хранится в signer_wallets (без private_key в wallets).
func (bc *Blockchain) CreateWallet(ctx context.Context) (*Wallet, error) {
	if bc.SignerCreator != nil {
		return NewWalletWithSigner(ctx, bc.Pool, bc.SignerCreator)
	}
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
	// При наличии params и ABI дополняем bytecode ABI-кодированными аргументами конструктора
	var bytecode []byte
	if len(params.Params) > 0 && len(params.ABI) > 0 {
		full, err := AppendConstructorArgs(params.Bytecode, params.ABI, params.Params)
		if err != nil {
			return "", fmt.Errorf("constructor args: %w", err)
		}
		bytecode = full
	} else {
		var err error
		bytecode, err = hex.DecodeString(params.Bytecode)
		if err != nil {
			return "", fmt.Errorf("invalid bytecode: %v", err)
		}
	}

	// Уникальный адрес контракта: hash(bytecode, from, nonce). При nonce=0 подставляем UnixNano, чтобы один кошелёк мог деплоить несколько контрактов без дубликата contracts_address_key
	nonce := params.Nonce
	if nonce == 0 {
		nonce = uint64(time.Now().UnixNano())
	}
	contractAddress := generateContractAddress(bytecode, params.From, nonce)

	// ABI для сохранения в БД (нужен для GetContractView — список методов чтения/записи)
	abiBytes := []byte(params.ABI)
	if len(abiBytes) == 0 {
		abiBytes = nil
	}

	// Create new contract
	contract := NewContract(
		contractAddress,
		params.From,
		bytecode,
		abiBytes,
		params.Standard,
		params.Version,
		0, // blockID will be set when block is created
		0, // txID will be set when transaction is created
	)
	contract.Creator = params.From
	contract.Name = params.Name
	contract.Symbol = params.Standard
	contract.Standard = params.Standard
	contract.Description = params.Description
	contract.MetadataCID = params.MetadataCID
	if params.Owner != "" {
		contract.Owner = params.Owner
	} else {
		contract.Owner = params.From
	}
	contract.SourceCode = params.SourceCode
	contract.Compiler = params.Compiler
	if len(params.Metadata) > 0 {
		contract.Metadata = params.Metadata
	}
	if len(params.Params) > 0 {
		contract.Params, _ = json.Marshal(params.Params)
	}
	// code записывается в БД: при деплое подставляем bytecode
	contract.Code = bytecode

	// Save contract to database
	if err := contract.SaveToDB(context.Background(), bc.Pool); err != nil {
		return "", fmt.Errorf("failed to save contract: %v", err)
	}

	// Записываем начальный storage (слот 0 = _totalSupply) для корректного чтения totalSupply() через CallStatic
	ctx := context.Background()
	blockID := int64(0)
	if bc.Genesis != nil {
		blockID = bc.Genesis.ID
	}
	if blockID <= 0 && bc.Pool != nil {
		_ = bc.Pool.QueryRow(ctx, `SELECT id FROM blocks WHERE index = 0 LIMIT 1`).Scan(&blockID)
	}
	if blockID > 0 && bc.Pool != nil && len(params.Params) > 0 {
		if errInit := WriteInitialStorageForDeployedContract(ctx, bc.Pool, blockID, contract.Address, params.Params); errInit != nil {
			log.Printf("[DeployContract] запись начального storage для %s: %v", contract.Address, errInit)
		}
	}

	return contract.Address, nil
}

// generateContractAddress формирует уникальный адрес контракта по bytecode, адресу создателя и nonce (избегает дубликата contracts_address_key)
func generateContractAddress(bytecode []byte, creator string, nonce uint64) string {
	h := sha256.New()
	h.Write(bytecode)
	h.Write([]byte(creator))
	var nonceBuf [8]byte
	binary.BigEndian.PutUint64(nonceBuf[:], nonce)
	h.Write(nonceBuf[:])
	sum := h.Sum(nil)
	return types.ContractAddressPrefix + hex.EncodeToString(sum[:16]) // 16 байт = 32 hex-символа
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
