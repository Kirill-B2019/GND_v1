// | KB @CerbeRus - Nexus Invest Team
package utils

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

// Event представляет событие в системе
type Event struct {
	ID          int64
	Type        string
	Contract    string
	FromAddress string
	ToAddress   string
	Amount      string
	Timestamp   time.Time
	TxHash      string
	Error       string
	Metadata    map[string]interface{}
}

// EventManager управляет событиями в системе
type EventManager struct {
	pool *pgxpool.Pool
}

// NewEventManager создает новый менеджер событий
func NewEventManager(pool *pgxpool.Pool) *EventManager {
	return &EventManager{
		pool: pool,
	}
}
