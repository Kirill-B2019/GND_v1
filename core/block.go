package core

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Block представляет блок в блокчейне ГАНИМЕД
type Block struct {
	ID           int            // ID блока
	Hash         string         // Хеш блока
	PrevHash     string         // Хеш предыдущего блока
	MerkleRoot   string         // Корень дерева Меркла
	Timestamp    time.Time      // Временная метка создания блока
	Height       int64          // Высота блока
	Version      int            // Версия блока
	Size         int            // Размер блока в байтах
	TxCount      int            // Количество транзакций в блоке
	GasUsed      int64          // Использованный газ
	GasLimit     int64          // Лимит газа
	Difficulty   int64          // Сложность блока
	Nonce        int64          // Нонс блока
	Miner        string         // Адрес майнера
	Reward       string         // Награда за блок
	ExtraData    []byte         // Дополнительные данные
	CreatedAt    time.Time      // Время создания записи
	UpdatedAt    time.Time      // Время последнего обновления
	Status       string         // Статус блока
	ParentID     *int           // ID родительского блока
	IsOrphaned   bool           // Является ли блок орфаном
	IsFinalized  bool           // Является ли блок финализированным
	Index        uint64         // Индекс блока
	Consensus    string         // Тип консенсуса
	Transactions []*Transaction // Транзакции в блоке
}

// NewBlock создает новый блок
func NewBlock(prevHash string, height int64, miner string) *Block {
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

// SaveToDB сохраняет блок в БД
func (b *Block) SaveToDB(ctx context.Context, pool *pgxpool.Pool) error {
	err := pool.QueryRow(ctx, `
		INSERT INTO blocks (
			hash, prev_hash, merkle_root, timestamp, height,
			version, size, tx_count, gas_used, gas_limit,
			difficulty, nonce, miner, reward, extra_data,
			created_at, updated_at, status, parent_id,
			is_orphaned, is_finalized
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
		RETURNING id`,
		b.Hash, b.PrevHash, b.MerkleRoot, b.Timestamp, b.Height,
		b.Version, b.Size, b.TxCount, b.GasUsed, b.GasLimit,
		b.Difficulty, b.Nonce, b.Miner, b.Reward, b.ExtraData,
		b.CreatedAt, b.UpdatedAt, b.Status, b.ParentID,
		b.IsOrphaned, b.IsFinalized,
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

// LoadBlock загружает блок из БД по хешу
func LoadBlock(ctx context.Context, pool *pgxpool.Pool, hash string) (*Block, error) {
	var id int
	var prevHash, merkleRoot, miner, reward string
	var timestamp, createdAt, updatedAt time.Time
	var height int64
	var version, size, txCount int
	var gasUsed, gasLimit, difficulty, nonce int64
	var extraData []byte
	var status string
	var parentID *int
	var isOrphaned, isFinalized bool

	err := pool.QueryRow(ctx, `
		SELECT id, prev_hash, merkle_root, timestamp, height,
			version, size, tx_count, gas_used, gas_limit,
			difficulty, nonce, miner, reward, extra_data,
			created_at, updated_at, status, parent_id,
			is_orphaned, is_finalized
		FROM blocks
		WHERE hash = $1`,
		hash,
	).Scan(&id, &prevHash, &merkleRoot, &timestamp, &height,
		&version, &size, &txCount, &gasUsed, &gasLimit,
		&difficulty, &nonce, &miner, &reward, &extraData,
		&createdAt, &updatedAt, &status, &parentID,
		&isOrphaned, &isFinalized)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("блок не найден: %s", hash)
	}
	if err != nil {
		return nil, fmt.Errorf("ошибка загрузки блока: %w", err)
	}

	return &Block{
		ID:          id,
		Hash:        hash,
		PrevHash:    prevHash,
		MerkleRoot:  merkleRoot,
		Timestamp:   timestamp,
		Height:      height,
		Version:     version,
		Size:        size,
		TxCount:     txCount,
		GasUsed:     gasUsed,
		GasLimit:    gasLimit,
		Difficulty:  difficulty,
		Nonce:       nonce,
		Miner:       miner,
		Reward:      reward,
		ExtraData:   extraData,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		Status:      status,
		ParentID:    parentID,
		IsOrphaned:  isOrphaned,
		IsFinalized: isFinalized,
	}, nil
}

// GetBlockByHeight возвращает блок по высоте
func GetBlockByHeight(ctx context.Context, pool *pgxpool.Pool, height int64) (*Block, error) {
	var hash string
	err := pool.QueryRow(ctx, `
		SELECT hash
		FROM blocks
		WHERE height = $1 AND is_orphaned = false
		ORDER BY created_at DESC
		LIMIT 1`,
		height,
	).Scan(&hash)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("блок не найден на высоте %d", height)
	}
	if err != nil {
		return nil, fmt.Errorf("ошибка получения блока по высоте: %w", err)
	}

	return LoadBlock(ctx, pool, hash)
}

// GetLatestBlock возвращает последний блок
func GetLatestBlock(ctx context.Context, pool *pgxpool.Pool) (*Block, error) {
	var hash string
	err := pool.QueryRow(ctx, `
		SELECT hash
		FROM blocks
		WHERE is_orphaned = false
		ORDER BY height DESC, created_at DESC
		LIMIT 1`,
	).Scan(&hash)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("блоки не найдены")
	}
	if err != nil {
		return nil, fmt.Errorf("ошибка получения последнего блока: %w", err)
	}

	return LoadBlock(ctx, pool, hash)
}

// CalculateHash вычисляет хеш блока
func (b *Block) CalculateHash() string {
	var sb strings.Builder
	sb.WriteString(b.PrevHash)
	sb.WriteString(b.MerkleRoot)
	sb.WriteString(b.Timestamp.Format(time.RFC3339))
	sb.WriteString(strconv.FormatInt(b.Height, 10))
	sb.WriteString(strconv.Itoa(b.Version))
	sb.WriteString(strconv.Itoa(b.TxCount))
	sb.WriteString(strconv.FormatInt(b.GasUsed, 10))
	sb.WriteString(strconv.FormatInt(b.GasLimit, 10))
	sb.WriteString(strconv.FormatInt(b.Difficulty, 10))
	sb.WriteString(strconv.FormatInt(b.Nonce, 10))
	sb.WriteString(b.Miner)
	sb.WriteString(b.Reward)
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
	sb.WriteString(strconv.FormatInt(b.Height, 10))
	sb.WriteString(strconv.Itoa(b.Version))
	sb.WriteString(strconv.Itoa(b.TxCount))
	sb.WriteString(strconv.FormatInt(b.GasUsed, 10))
	sb.WriteString(strconv.FormatInt(b.GasLimit, 10))
	sb.WriteString(strconv.FormatInt(b.Difficulty, 10))
	sb.WriteString(b.Miner)
	sb.WriteString(b.Reward)
	sb.Write(b.ExtraData)

	hash := sha256.Sum256([]byte(sb.String()))
	return hex.EncodeToString(hash[:])
}
