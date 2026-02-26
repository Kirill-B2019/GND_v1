// api/eventmanager_stub.go

package api

import (
	"GND/types"
	"context"
	"time"
)

// noopEventManager реализует types.EventManager заглушкой (только Emit пишет в канал/игнорируется).
// Используется для деплоера токенов, когда полноценный vm.EventManager не передаётся в API.
type noopEventManager struct{}

func (*noopEventManager) Emit(event *types.Event) error                           { return nil }
func (*noopEventManager) SaveEvent(ctx context.Context, event *types.Event) error { return nil }
func (*noopEventManager) GetEventsByContractAndType(ctx context.Context, contract string, eventType types.EventType, limit int) ([]*types.Event, error) {
	return nil, nil
}
func (*noopEventManager) GetEventStats(ctx context.Context, startTime, endTime time.Time) (*types.EventStats, error) {
	return nil, nil
}
func (*noopEventManager) GetLatestEvents(ctx context.Context, contract string, eventType types.EventType, limit int) ([]*types.Event, error) {
	return nil, nil
}
func (*noopEventManager) DeleteOldEvents(ctx context.Context, before time.Time) error { return nil }
func (*noopEventManager) GetEventByID(ctx context.Context, id int64) (*types.Event, error) {
	return nil, nil
}
func (*noopEventManager) UpdateEventMetadata(ctx context.Context, id int64, metadata map[string]interface{}) error {
	return nil
}
