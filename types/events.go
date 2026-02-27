// | KB @CerbeRus - Nexus Invest Team
package types

import (
	"context"
	"time"
)

// EventType определяет тип события
type EventType string

const (
	EventTransfer EventType = "Transfer"
	EventApproval EventType = "Approval"
	EventDeploy   EventType = "Deploy"
	EventError    EventType = "Error"
	EventCustom   EventType = "Custom"
)

// Event представляет событие в системе
type Event struct {
	ID          int64
	Type        EventType
	Contract    string
	FromAddress string
	ToAddress   string
	Amount      string
	Timestamp   time.Time
	TxHash      string
	Error       string
	Metadata    map[string]interface{}
}

// EventStats представляет статистику по событиям
type EventStats struct {
	EventType   string
	EventCount  int64
	TotalAmount string
	FirstEvent  time.Time
	LastEvent   time.Time
}

// EventManager определяет интерфейс для управления событиями
type EventManager interface {
	// Emit отправляет событие в систему
	Emit(event *Event) error

	// SaveEvent сохраняет событие в базу данных
	SaveEvent(ctx context.Context, event *Event) error

	// GetEventsByContractAndType получает события по контракту и типу
	GetEventsByContractAndType(ctx context.Context, contract string, eventType EventType, limit int) ([]*Event, error)

	// GetEventStats получает статистику по событиям
	GetEventStats(ctx context.Context, startTime, endTime time.Time) (*EventStats, error)

	// GetLatestEvents получает последние события
	GetLatestEvents(ctx context.Context, contract string, eventType EventType, limit int) ([]*Event, error)

	// DeleteOldEvents удаляет старые события
	DeleteOldEvents(ctx context.Context, before time.Time) error

	// GetEventByID получает событие по ID
	GetEventByID(ctx context.Context, id int64) (*Event, error)

	// UpdateEventMetadata обновляет метаданные события
	UpdateEventMetadata(ctx context.Context, id int64, metadata map[string]interface{}) error
}
