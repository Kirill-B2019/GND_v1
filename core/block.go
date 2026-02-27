// | KB @CerbeRus - Nexus Invest Team
package core

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Block represents a block in the blockchain
type Block struct {
	ID           int64          // ID блока
	Hash         string         // Хеш блока
	PrevHash     string         // Хеш предыдущего блока
	MerkleRoot   string         // Корень дерева Меркла
	Timestamp    time.Time      // Временная метка создания блока
	Height       uint64         // Высота блока
	Version      uint32         // Версия блока
	Size         uint64         // Размер блока в байтах
	TxCount      uint32         // Количество транзакций в блоке
	GasUsed      uint64         // Использованный газ
	GasLimit     uint64         // Лимит газа
	Difficulty   uint64         // Сложность блока
	Nonce        uint64         // Нонс блока
	Miner        string         // Адрес майнера
	Reward       *big.Int       // Награда за блок
	ExtraData    []byte         // Дополнительные данные
	CreatedAt    time.Time      // Время создания записи
	UpdatedAt    time.Time      // Время последнего обновления
	Status       string         // Статус блока
	ParentID     *int64         // ID родительского блока
	IsOrphaned   bool           // Является ли блок орфаном
	IsFinalized  bool           // Является ли блок финализированным
	Index        uint64         // Индекс блока
	Consensus    string         // Тип консенсуса
	Header       *BlockHeader   // Заголовок блока
	Transactions []*Transaction // Транзакции в блоке
}

// BlockHeader represents a block header
type BlockHeader struct {
	Number    uint64
	Timestamp time.Time
	Hash      string
	PrevHash  string
}

// NewBlock создает новый блок
func NewBlock(prevHash string, height uint64, miner string) *Block {
	now := time.Now()
	return &Block{
		PrevHash:    prevHash,
		Height:      height,
		Version:     1,
		Timestamp:   now,
		Miner:       miner,
		CreatedAt:   now,
		UpdatedAt:   now,
		Status:      "pending",
		IsOrphaned:  false,
		IsFinalized: false,
	}
}

// SaveToDB сохраняет блок в БД. created_at = время создания блока (timestamp), updated_at = время финализации.
func (b *Block) SaveToDB(ctx context.Context, pool *pgxpool.Pool) error {
	createdAt := b.CreatedAt
	if createdAt.IsZero() {
		createdAt = b.Timestamp
	}
	updatedAt := b.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = time.Now().UTC()
	}

	rewardStr := "0"
	if b.Reward != nil {
		rewardStr = b.Reward.String()
	}
	nonceStr := strconv.FormatUint(b.Nonce, 10) // в БД blocks.nonce — varchar
	err := pool.QueryRow(ctx, `
		INSERT INTO blocks (
			hash, prev_hash, merkle_root, timestamp, height,
			version, size, tx_count, gas_used, gas_limit,
			difficulty, nonce, miner, reward, extra_data,
			created_at, updated_at, status, parent_id,
			is_orphaned, is_finalized, index, consensus
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23)
		RETURNING id`,
		b.Hash, b.PrevHash, b.MerkleRoot, b.Timestamp, b.Height,
		b.Version, b.Size, b.TxCount, b.GasUsed, b.GasLimit,
		b.Difficulty, nonceStr, b.Miner, rewardStr, b.ExtraData,
		createdAt, updatedAt, b.Status, b.ParentID,
		b.IsOrphaned, b.IsFinalized, b.Index, b.Consensus,
	).Scan(&b.ID)

	if err != nil {
		return fmt.Errorf("ошибка сохранения блока: %w", err)
	}

	return nil
}

// UpdateStatus обновляет статус блока
func (b *Block) UpdateStatus(ctx context.Context, pool *pgxpool.Pool, status string) error {
	b.Status = status
	b.UpdatedAt = time.Now()

	_, err := pool.Exec(ctx, `
		UPDATE blocks 
		SET status = $1, updated_at = $2
		WHERE id = $3`,
		b.Status, b.UpdatedAt, b.ID,
	)

	if err != nil {
		return fmt.Errorf("ошибка обновления статуса блока: %w", err)
	}

	return nil
}

// parseBlockNonce парсит nonce из varchar в uint64 (в БД blocks.nonce — varchar)
func parseBlockNonce(s string) (uint64, error) {
	if s == "" {
		return 0, nil
	}
	return strconv.ParseUint(s, 10, 64)
}

// LoadBlockByHash загружает блок из БД по хешу
func LoadBlockByHash(pool *pgxpool.Pool, hash string) (*Block, error) {
	var block Block
	var rewardStr, nonceStr string
	err := pool.QueryRow(context.Background(), `
		SELECT id, hash, prev_hash, merkle_root, timestamp, height, version, size, tx_count, gas_used, gas_limit, difficulty, nonce, miner, reward, extra_data, created_at, updated_at, status, parent_id, is_orphaned, is_finalized, index, consensus
		FROM blocks WHERE hash = $1`, hash).Scan(
		&block.ID,
		&block.Hash,
		&block.PrevHash,
		&block.MerkleRoot,
		&block.Timestamp,
		&block.Height,
		&block.Version,
		&block.Size,
		&block.TxCount,
		&block.GasUsed,
		&block.GasLimit,
		&block.Difficulty,
		&nonceStr,
		&block.Miner,
		&rewardStr,
		&block.ExtraData,
		&block.CreatedAt,
		&block.UpdatedAt,
		&block.Status,
		&block.ParentID,
		&block.IsOrphaned,
		&block.IsFinalized,
		&block.Index,
		&block.Consensus,
	)
	if err != nil {
		return nil, err
	}
	block.Nonce, _ = parseBlockNonce(nonceStr)

	block.Reward = new(big.Int)
	block.Reward.SetString(rewardStr, 10)

	return &block, nil
}

// LoadBlock загружает блок из БД по высоте
func LoadBlock(pool *pgxpool.Pool, height uint64) (*Block, error) {
	var block Block
	var rewardStr, nonceStr string
	err := pool.QueryRow(context.Background(), `
		SELECT id, hash, prev_hash, merkle_root, timestamp, height, version, size, tx_count, gas_used, gas_limit, difficulty, nonce, miner, reward, extra_data, created_at, updated_at, status, parent_id, is_orphaned, is_finalized, index, consensus
		FROM blocks WHERE index = $1`, height).Scan(
		&block.ID,
		&block.Hash,
		&block.PrevHash,
		&block.MerkleRoot,
		&block.Timestamp,
		&block.Height,
		&block.Version,
		&block.Size,
		&block.TxCount,
		&block.GasUsed,
		&block.GasLimit,
		&block.Difficulty,
		&nonceStr,
		&block.Miner,
		&rewardStr,
		&block.ExtraData,
		&block.CreatedAt,
		&block.UpdatedAt,
		&block.Status,
		&block.ParentID,
		&block.IsOrphaned,
		&block.IsFinalized,
		&block.Index,
		&block.Consensus,
	)
	if err != nil {
		return nil, err
	}
	block.Nonce, _ = parseBlockNonce(nonceStr)

	block.Reward = new(big.Int)
	block.Reward.SetString(rewardStr, 10)

	return &block, nil
}

// GetBlockByHeight returns a block by its height
func GetBlockByHeight(pool *pgxpool.Pool, height uint64) (*Block, error) {
	var block Block
	var rewardStr, nonceStr string
	err := pool.QueryRow(context.Background(),
		"SELECT id, hash, prev_hash, merkle_root, timestamp, height, version, size, tx_count, gas_used, gas_limit, difficulty, nonce, miner, reward, extra_data, created_at, updated_at, status, parent_id, is_orphaned, is_finalized, index, consensus FROM blocks WHERE height = $1",
		height,
	).Scan(
		&block.ID,
		&block.Hash,
		&block.PrevHash,
		&block.MerkleRoot,
		&block.Timestamp,
		&block.Height,
		&block.Version,
		&block.Size,
		&block.TxCount,
		&block.GasUsed,
		&block.GasLimit,
		&block.Difficulty,
		&nonceStr,
		&block.Miner,
		&rewardStr,
		&block.ExtraData,
		&block.CreatedAt,
		&block.UpdatedAt,
		&block.Status,
		&block.ParentID,
		&block.IsOrphaned,
		&block.IsFinalized,
		&block.Index,
		&block.Consensus,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get block by height: %v", err)
	}
	block.Nonce, _ = parseBlockNonce(nonceStr)

	block.Reward = new(big.Int)
	block.Reward.SetString(rewardStr, 10)

	return &block, nil
}

// GetLatestBlock returns the latest block
func GetLatestBlock(pool *pgxpool.Pool) (*Block, error) {
	var block Block
	var rewardStr, nonceStr string
	err := pool.QueryRow(context.Background(),
		"SELECT id, hash, prev_hash, merkle_root, timestamp, height, version, size, tx_count, gas_used, gas_limit, difficulty, nonce, miner, reward, extra_data, created_at, updated_at, status, parent_id, is_orphaned, is_finalized, index, consensus FROM blocks ORDER BY index DESC LIMIT 1",
	).Scan(
		&block.ID,
		&block.Hash,
		&block.PrevHash,
		&block.MerkleRoot,
		&block.Timestamp,
		&block.Height,
		&block.Version,
		&block.Size,
		&block.TxCount,
		&block.GasUsed,
		&block.GasLimit,
		&block.Difficulty,
		&nonceStr,
		&block.Miner,
		&rewardStr,
		&block.ExtraData,
		&block.CreatedAt,
		&block.UpdatedAt,
		&block.Status,
		&block.ParentID,
		&block.IsOrphaned,
		&block.IsFinalized,
		&block.Index,
		&block.Consensus,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest block: %v", err)
	}
	block.Nonce, _ = parseBlockNonce(nonceStr)

	block.Reward = new(big.Int)
	block.Reward.SetString(rewardStr, 10)

	return &block, nil
}

// CalculateHash вычисляет хеш блока
func (b *Block) CalculateHash() string {
	var sb strings.Builder
	sb.WriteString(b.PrevHash)
	sb.WriteString(b.MerkleRoot)
	sb.WriteString(b.Timestamp.Format(time.RFC3339))
	sb.WriteString(strconv.FormatInt(int64(b.Height), 10))
	sb.WriteString(strconv.Itoa(int(b.Version)))
	sb.WriteString(strconv.Itoa(int(b.TxCount)))
	sb.WriteString(strconv.FormatInt(int64(b.GasUsed), 10))
	sb.WriteString(strconv.FormatInt(int64(b.GasLimit), 10))
	sb.WriteString(strconv.FormatInt(int64(b.Difficulty), 10))
	sb.WriteString(strconv.FormatInt(int64(b.Nonce), 10))
	sb.WriteString(b.Miner)
	sb.WriteString(b.Reward.String())
	sb.Write(b.ExtraData)

	hash := sha256.Sum256([]byte(sb.String()))
	return hex.EncodeToString(hash[:])
}

// CalculateHashWithoutNonce вычисляет хеш блока без учета nonce
func (b *Block) CalculateHashWithoutNonce() string {
	var sb strings.Builder
	sb.WriteString(b.PrevHash)
	sb.WriteString(b.MerkleRoot)
	sb.WriteString(b.Timestamp.Format(time.RFC3339))
	sb.WriteString(strconv.FormatInt(int64(b.Height), 10))
	sb.WriteString(strconv.Itoa(int(b.Version)))
	sb.WriteString(strconv.Itoa(int(b.TxCount)))
	sb.WriteString(strconv.FormatInt(int64(b.GasUsed), 10))
	sb.WriteString(strconv.FormatInt(int64(b.GasLimit), 10))
	sb.WriteString(strconv.FormatInt(int64(b.Difficulty), 10))
	sb.WriteString(b.Miner)
	sb.WriteString(b.Reward.String())
	sb.Write(b.ExtraData)

	hash := sha256.Sum256([]byte(sb.String()))
	return hex.EncodeToString(hash[:])
}

// GetBlockByNumber returns a block by its number
func GetBlockByNumber(pool *pgxpool.Pool, number uint64) (*Block, error) {
	var block Block
	var rewardStr, nonceStr string
	err := pool.QueryRow(context.Background(),
		"SELECT id, hash, prev_hash, merkle_root, timestamp, height, version, size, tx_count, gas_used, gas_limit, difficulty, nonce, miner, reward, extra_data, created_at, updated_at, status, parent_id, is_orphaned, is_finalized, index, consensus FROM blocks WHERE index = $1",
		number,
	).Scan(
		&block.ID,
		&block.Hash,
		&block.PrevHash,
		&block.MerkleRoot,
		&block.Timestamp,
		&block.Height,
		&block.Version,
		&block.Size,
		&block.TxCount,
		&block.GasUsed,
		&block.GasLimit,
		&block.Difficulty,
		&nonceStr,
		&block.Miner,
		&rewardStr,
		&block.ExtraData,
		&block.CreatedAt,
		&block.UpdatedAt,
		&block.Status,
		&block.ParentID,
		&block.IsOrphaned,
		&block.IsFinalized,
		&block.Index,
		&block.Consensus,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get block by number: %v", err)
	}
	block.Nonce, _ = parseBlockNonce(nonceStr)

	block.Reward = new(big.Int)
	block.Reward.SetString(rewardStr, 10)

	return &block, nil
}

// GetBlockByHash returns a block by its hash
func GetBlockByHash(pool *pgxpool.Pool, hash string) (*Block, error) {
	var block Block
	var rewardStr, nonceStr string
	err := pool.QueryRow(context.Background(),
		"SELECT id, hash, prev_hash, merkle_root, timestamp, height, version, size, tx_count, gas_used, gas_limit, difficulty, nonce, miner, reward, extra_data, created_at, updated_at, status, parent_id, is_orphaned, is_finalized, index, consensus FROM blocks WHERE hash = $1",
		hash,
	).Scan(
		&block.ID,
		&block.Hash,
		&block.PrevHash,
		&block.MerkleRoot,
		&block.Timestamp,
		&block.Height,
		&block.Version,
		&block.Size,
		&block.TxCount,
		&block.GasUsed,
		&block.GasLimit,
		&block.Difficulty,
		&nonceStr,
		&block.Miner,
		&rewardStr,
		&block.ExtraData,
		&block.CreatedAt,
		&block.UpdatedAt,
		&block.Status,
		&block.ParentID,
		&block.IsOrphaned,
		&block.IsFinalized,
		&block.Index,
		&block.Consensus,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get block by hash: %v", err)
	}
	block.Nonce, _ = parseBlockNonce(nonceStr)

	block.Reward = new(big.Int)
	block.Reward.SetString(rewardStr, 10)

	return &block, nil
}

// GetBlocks returns a list of blocks with pagination
func GetBlocks(pool *pgxpool.Pool, limit, offset int) ([]*Block, error) {
	rows, err := pool.Query(context.Background(),
		"SELECT id, hash, prev_hash, merkle_root, timestamp, height, version, size, tx_count, gas_used, gas_limit, difficulty, nonce, miner, reward, extra_data, created_at, updated_at, status, parent_id, is_orphaned, is_finalized, index, consensus FROM blocks ORDER BY height DESC LIMIT $1 OFFSET $2",
		limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get blocks: %v", err)
	}
	defer rows.Close()

	var blocks []*Block
	for rows.Next() {
		var block Block
		var rewardStr, nonceStr string
		err := rows.Scan(
			&block.ID,
			&block.Hash,
			&block.PrevHash,
			&block.MerkleRoot,
			&block.Timestamp,
			&block.Height,
			&block.Version,
			&block.Size,
			&block.TxCount,
			&block.GasUsed,
			&block.GasLimit,
			&block.Difficulty,
			&nonceStr,
			&block.Miner,
			&rewardStr,
			&block.ExtraData,
			&block.CreatedAt,
			&block.UpdatedAt,
			&block.Status,
			&block.ParentID,
			&block.IsOrphaned,
			&block.IsFinalized,
			&block.Index,
			&block.Consensus,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan block: %v", err)
		}
		block.Nonce, _ = parseBlockNonce(nonceStr)

		block.Reward = new(big.Int)
		block.Reward.SetString(rewardStr, 10)

		blocks = append(blocks, &block)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating blocks: %v", err)
	}

	return blocks, nil
}
