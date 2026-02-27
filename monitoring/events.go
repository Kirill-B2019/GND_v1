// | KB @CerbeRus - Nexus Invest Team
package monitoring

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// EventLevel определяет уровень логируемого события
type EventLevel string

const (
	EventDebug EventLevel = "DEBUG"
	EventInfo  EventLevel = "INFO"
	EventWarn  EventLevel = "WARN"
	EventError EventLevel = "ERROR"
	EventFatal EventLevel = "FATAL"
)

// Event описывает структуру события в блокчейне
type Event struct {
	Timestamp time.Time  // Время события
	Level     EventLevel // Уровень события
	Component string     // Компонент системы (например: "Consensus", "Tx", "SmartContract")
	EventType string     // Тип события (например: "BlockCreated", "TxConfirmed", "ContractError")
	TxID      string     // Идентификатор транзакции (если применимо)
	BlockID   string     // Идентификатор блока (если применимо)
	Address   string     // Адрес участника (если применимо)
	Message   string     // Краткое описание события
	Details   string     // Дополнительные детали (JSON, текст)
}

// EventLogger управляет логированием и хранением событий
type EventLogger struct {
	events    []Event
	eventChan chan Event
	mutex     sync.RWMutex
	logFile   *os.File
	level     EventLevel
}

// NewEventLogger создает новый логгер событий с записью в файл
func NewEventLogger(logFilePath string, minLevel EventLevel) (*EventLogger, error) {
	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	el := &EventLogger{
		events:    make([]Event, 0, 1000),
		eventChan: make(chan Event, 1000),
		logFile:   file,
		level:     minLevel,
	}
	go el.processEvents()
	return el, nil
}

// processEvents обрабатывает события из канала, пишет в файл и хранит в памяти
func (el *EventLogger) processEvents() {
	for event := range el.eventChan {
		if !el.shouldLog(event.Level) {
			continue
		}
		el.mutex.Lock()
		if len(el.events) >= 5000 {
			el.events = el.events[1:] // кольцевой буфер
		}
		el.events = append(el.events, event)
		el.mutex.Unlock()
		fmt.Fprintf(el.logFile, "[%s] [%s] [%s] [%s] Tx:%s Block:%s Addr:%s %s | %s\n",
			event.Timestamp.Format(time.RFC3339), event.Level, event.Component, event.EventType,
			event.TxID, event.BlockID, event.Address, event.Message, event.Details)
	}
}

// shouldLog проверяет, соответствует ли уровень события текущему уровню логирования
func (el *EventLogger) shouldLog(level EventLevel) bool {
	order := map[EventLevel]int{
		EventDebug: 1,
		EventInfo:  2,
		EventWarn:  3,
		EventError: 4,
		EventFatal: 5,
	}
	return order[level] >= order[el.level]
}

// LogEvent отправляет событие в систему логирования
func (el *EventLogger) LogEvent(level EventLevel, component, eventType, txID, blockID, address, message, details string) {
	event := Event{
		Timestamp: time.Now(),
		Level:     level,
		Component: component,
		EventType: eventType,
		TxID:      txID,
		BlockID:   blockID,
		Address:   address,
		Message:   message,
		Details:   details,
	}
	el.eventChan <- event
}

// ListEvents возвращает последние N событий
func (el *EventLogger) ListEvents(n int) []Event {
	el.mutex.RLock()
	defer el.mutex.RUnlock()
	if n > len(el.events) {
		n = len(el.events)
	}
	return append([]Event(nil), el.events[len(el.events)-n:]...)
}

// SetLevel позволяет динамически изменить уровень логирования
func (el *EventLogger) SetLevel(level EventLevel) {
	el.mutex.Lock()
	defer el.mutex.Unlock()
	el.level = level
}

// Close завершает работу логгера и закрывает файл
func (el *EventLogger) Close() error {
	close(el.eventChan)
	return el.logFile.Close()
}

// Пример использования:
/*
func main() {
	logger, err := NewEventLogger("events.log", EventInfo)
	if err != nil {
		panic(err)
	}
	defer logger.Close()

	logger.LogEvent(EventInfo, "Consensus", "LeaderElected", "", "12345", "GND1...", "Новый лидер выбран", "Validator: GND1...")
	logger.LogEvent(EventError, "SmartContract", "ExecutionFailed", "0xabc", "12345", "GNDct1...", "Ошибка исполнения контракта", "out of gas")

	time.Sleep(1 * time.Second) // дать время обработать события

	for _, event := range logger.ListEvents(10) {
		fmt.Printf("[%s] %s: %s - %s\n", event.Level, event.Component, event.EventType, event.Message)
	}
}
*/
