package vm

import (
	"GND/types"
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// EventManager реализует интерфейс types.EventManager
type EventManager struct {
	pool     *pgxpool.Pool
	events   chan *types.Event
	subs     map[types.EventType][]chan *types.Event
	mu       sync.RWMutex
	stopChan chan struct{}
}

// NewEventManager создает новый менеджер событий
func NewEventManager(pool *pgxpool.Pool) *EventManager {
	em := &EventManager{
		pool:     pool,
		events:   make(chan *types.Event, 1000),
		subs:     make(map[types.EventType][]chan *types.Event),
		stopChan: make(chan struct{}),
	}
	go em.processEvents()
	return em
}

// Subscribe подписывается на события определенного типа
func (em *EventManager) Subscribe(eventType types.EventType) <-chan *types.Event {
	em.mu.Lock()
	defer em.mu.Unlock()

	ch := make(chan *types.Event, 100)
	em.subs[eventType] = append(em.subs[eventType], ch)
	return ch
}

// Emit отправляет событие
func (em *EventManager) Emit(event *types.Event) {
	select {
	case em.events <- event:
	case <-em.stopChan:
	}
}

// processEvents обрабатывает события в фоновом режиме
func (em *EventManager) processEvents() {
	for {
		select {
		case event := <-em.events:
			em.mu.RLock()
			subs := em.subs[event.Type]
			em.mu.RUnlock()

			for _, ch := range subs {
				select {
				case ch <- event:
				default:
					// Если канал переполнен, пропускаем событие
				}
			}

			// Сохраняем событие в БД
			if err := em.SaveEvent(event); err != nil {
				// TODO: Добавить логирование ошибки
			}
		case <-em.stopChan:
			return
		}
	}
}

// Stop останавливает обработку событий
func (em *EventManager) Stop() {
	close(em.stopChan)
	em.mu.Lock()
	defer em.mu.Unlock()

	for _, subs := range em.subs {
		for _, ch := range subs {
			close(ch)
		}
	}
}

// SaveEvent сохраняет событие в базу данных
func (em *EventManager) SaveEvent(event *types.Event) error {
	query := `
		INSERT INTO events (
			type, contract, from_address, to_address, amount,
			"timestamp", tx_hash, error, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id`

	var metadata json.RawMessage
	if event.Metadata != nil {
		var err error
		metadata, err = json.Marshal(event.Metadata)
		if err != nil {
			return err
		}
	}

	var id int64
	err := em.pool.QueryRow(context.Background(), query,
		event.Type,
		event.Contract,
		event.FromAddress,
		event.ToAddress,
		event.Amount,
		event.Timestamp,
		event.TxHash,
		event.Error,
		metadata,
	).Scan(&id)

	if err != nil {
		return err
	}

	event.ID = id
	return nil
}

// GetEventsByContractAndType получает события по контракту и типу
func (em *EventManager) GetEventsByContractAndType(contract string, eventType types.EventType, limit int) ([]*types.Event, error) {
	query := `
		SELECT id, type, contract, from_address, to_address, amount,
			"timestamp", tx_hash, error, metadata
		FROM events
		WHERE contract = $1 AND type = $2
		ORDER BY "timestamp" DESC
		LIMIT $3`

	rows, err := em.pool.Query(context.Background(), query, contract, eventType, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*types.Event
	for rows.Next() {
		event := &types.Event{}
		var metadata json.RawMessage
		err := rows.Scan(
			&event.ID,
			&event.Type,
			&event.Contract,
			&event.FromAddress,
			&event.ToAddress,
			&event.Amount,
			&event.Timestamp,
			&event.TxHash,
			&event.Error,
			&metadata,
		)
		if err != nil {
			return nil, err
		}

		if len(metadata) > 0 {
			if err := json.Unmarshal(metadata, &event.Metadata); err != nil {
				return nil, err
			}
		}

		events = append(events, event)
	}

	return events, rows.Err()
}

// GetEventStats получает статистику по событиям
func (em *EventManager) GetEventStats(contract string, startTime, endTime time.Time) ([]*types.EventStats, error) {
	query := `
		SELECT event_type, event_count, total_amount, first_event, last_event
		FROM get_event_stats($1, $2, $3)`

	rows, err := em.pool.Query(context.Background(), query, contract, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []*types.EventStats
	for rows.Next() {
		stat := &types.EventStats{}
		err := rows.Scan(
			&stat.EventType,
			&stat.EventCount,
			&stat.TotalAmount,
			&stat.FirstEvent,
			&stat.LastEvent,
		)
		if err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}

	return stats, rows.Err()
}

// GetLatestEvents получает последние события
func (em *EventManager) GetLatestEvents(contract string, eventType types.EventType, limit int) ([]*types.Event, error) {
	query := `
		SELECT id, type, contract, from_address, to_address, amount,
			"timestamp", tx_hash, error, metadata
		FROM latest_events
		WHERE contract = $1 AND type = $2 AND event_rank <= $3
		ORDER BY "timestamp" DESC`

	rows, err := em.pool.Query(context.Background(), query, contract, eventType, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*types.Event
	for rows.Next() {
		event := &types.Event{}
		var metadata json.RawMessage
		err := rows.Scan(
			&event.ID,
			&event.Type,
			&event.Contract,
			&event.FromAddress,
			&event.ToAddress,
			&event.Amount,
			&event.Timestamp,
			&event.TxHash,
			&event.Error,
			&metadata,
		)
		if err != nil {
			return nil, err
		}

		if len(metadata) > 0 {
			if err := json.Unmarshal(metadata, &event.Metadata); err != nil {
				return nil, err
			}
		}

		events = append(events, event)
	}

	return events, rows.Err()
}

// DeleteOldEvents удаляет старые события
func (em *EventManager) DeleteOldEvents(olderThan time.Time) error {
	query := `DELETE FROM events WHERE "timestamp" < $1`
	_, err := em.pool.Exec(context.Background(), query, olderThan)
	return err
}

// GetEventByID получает событие по ID
func (em *EventManager) GetEventByID(id int64) (*types.Event, error) {
	query := `
		SELECT id, type, contract, from_address, to_address, amount,
			"timestamp", tx_hash, error, metadata
		FROM events
		WHERE id = $1`

	event := &types.Event{}
	var metadata json.RawMessage
	err := em.pool.QueryRow(context.Background(), query, id).Scan(
		&event.ID,
		&event.Type,
		&event.Contract,
		&event.FromAddress,
		&event.ToAddress,
		&event.Amount,
		&event.Timestamp,
		&event.TxHash,
		&event.Error,
		&metadata,
	)
	if err != nil {
		return nil, err
	}

	if len(metadata) > 0 {
		if err := json.Unmarshal(metadata, &event.Metadata); err != nil {
			return nil, err
		}
	}

	return event, nil
}

// UpdateEventMetadata обновляет метаданные события
func (em *EventManager) UpdateEventMetadata(id int64, metadata map[string]interface{}) error {
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	query := `UPDATE events SET metadata = $1 WHERE id = $2`
	_, err = em.pool.Exec(context.Background(), query, metadataJSON, id)
	return err
}
