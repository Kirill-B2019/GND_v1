package core

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// BlockchainEvent представляет событие в блокчейне ГАНИМЕД
type BlockchainEvent struct {
	ID          int       // ID события
	BlockID     int       // ID блока
	TxID        int       // ID транзакции
	Type        string    // Тип события
	Address     string    // Адрес контракта
	Topics      []string  // Темы события
	Data        []byte    // Данные события
	Index       int       // Индекс события в блоке
	Removed     bool      // Удалено ли событие
	Status      string    // Статус события
	ProcessedAt time.Time // Время обработки
	CreatedAt   time.Time // Время создания
	UpdatedAt   time.Time // Время последнего обновления
	Metadata    []byte    // Метаданные события
}

// NewBlockchainEvent создает новое событие
func NewBlockchainEvent(blockID, txID int, eventType, address string, topics []string, data []byte, index int) *BlockchainEvent {
	now := time.Now()
	return &BlockchainEvent{
		BlockID:   blockID,
		TxID:      txID,
		Type:      eventType,
		Address:   address,
		Topics:    topics,
		Data:      data,
		Index:     index,
		Removed:   false,
		Status:    "pending",
		CreatedAt: now,
		UpdatedAt: now,
		Metadata:  []byte{},
	}
}

// SaveToDB сохраняет событие в БД
func (e *BlockchainEvent) SaveToDB(ctx context.Context, pool *pgxpool.Pool) error {
	err := pool.QueryRow(ctx, `
		INSERT INTO events (
			block_id, tx_id, type, address, topics,
			data, index, removed, status, processed_at,
			created_at, updated_at, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13)
		RETURNING id`,
		e.BlockID, e.TxID, e.Type, e.Address, e.Topics,
		e.Data, e.Index, e.Removed, e.Status, e.ProcessedAt,
		e.CreatedAt, e.UpdatedAt, e.Metadata,
	).Scan(&e.ID)

	if err != nil {
		return fmt.Errorf("ошибка сохранения события: %w", err)
	}

	return nil
}

// UpdateStatus обновляет статус события
func (e *BlockchainEvent) UpdateStatus(ctx context.Context, pool *pgxpool.Pool, status string) error {
	e.Status = status
	e.UpdatedAt = time.Now()

	_, err := pool.Exec(ctx, `
		UPDATE events 
		SET status = $1, updated_at = $2
		WHERE id = $3`,
		e.Status, e.UpdatedAt, e.ID,
	)

	if err != nil {
		return fmt.Errorf("ошибка обновления статуса события: %w", err)
	}

	return nil
}

// MarkAsProcessed помечает событие как обработанное
func (e *BlockchainEvent) MarkAsProcessed(ctx context.Context, pool *pgxpool.Pool) error {
	e.Status = "processed"
	e.ProcessedAt = time.Now()
	e.UpdatedAt = time.Now()

	_, err := pool.Exec(ctx, `
		UPDATE events 
		SET status = $1, processed_at = $2, updated_at = $3
		WHERE id = $4`,
		e.Status, e.ProcessedAt, e.UpdatedAt, e.ID,
	)

	if err != nil {
		return fmt.Errorf("ошибка пометки события как обработанного: %w", err)
	}

	return nil
}

// LoadBlockchainEvent загружает событие из БД по ID
func LoadBlockchainEvent(ctx context.Context, pool *pgxpool.Pool, id int) (*BlockchainEvent, error) {
	var blockID, txID, index int
	var eventType, address string
	var topics []string
	var data, metadata []byte
	var removed bool
	var status string
	var processedAt, createdAt, updatedAt time.Time

	err := pool.QueryRow(ctx, `
		SELECT block_id, tx_id, type, address, topics,
			data, index, removed, status, processed_at,
			created_at, updated_at, metadata
		FROM events
		WHERE id = $1`,
		id,
	).Scan(&blockID, &txID, &eventType, &address, &topics,
		&data, &index, &removed, &status, &processedAt,
		&createdAt, &updatedAt, &metadata)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("событие не найдено: %d", id)
	}
	if err != nil {
		return nil, fmt.Errorf("ошибка загрузки события: %w", err)
	}

	return &BlockchainEvent{
		ID:          id,
		BlockID:     blockID,
		TxID:        txID,
		Type:        eventType,
		Address:     address,
		Topics:      topics,
		Data:        data,
		Index:       index,
		Removed:     removed,
		Status:      status,
		ProcessedAt: processedAt,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		Metadata:    metadata,
	}, nil
}

// GetBlockchainEventsByBlockID возвращает все события в блоке
func GetBlockchainEventsByBlockID(ctx context.Context, pool *pgxpool.Pool, blockID int) ([]*BlockchainEvent, error) {
	rows, err := pool.Query(ctx, `
		SELECT id
		FROM events
		WHERE block_id = $1
		ORDER BY index ASC`,
		blockID,
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения событий блока: %w", err)
	}
	defer rows.Close()

	var events []*BlockchainEvent
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("ошибка сканирования ID события: %w", err)
		}

		event, err := LoadBlockchainEvent(ctx, pool, id)
		if err != nil {
			return nil, fmt.Errorf("ошибка загрузки события %d: %w", id, err)
		}

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при чтении событий: %w", err)
	}

	return events, nil
}

// GetBlockchainEventsByTxID возвращает все события в транзакции
func GetBlockchainEventsByTxID(ctx context.Context, pool *pgxpool.Pool, txID int) ([]*BlockchainEvent, error) {
	rows, err := pool.Query(ctx, `
		SELECT id
		FROM events
		WHERE tx_id = $1
		ORDER BY index ASC`,
		txID,
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения событий транзакции: %w", err)
	}
	defer rows.Close()

	var events []*BlockchainEvent
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("ошибка сканирования ID события: %w", err)
		}

		event, err := LoadBlockchainEvent(ctx, pool, id)
		if err != nil {
			return nil, fmt.Errorf("ошибка загрузки события %d: %w", id, err)
		}

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при чтении событий: %w", err)
	}

	return events, nil
}
