// tokens/events.go

package tokens

import (
	"fmt"
	"time"
)

type Event struct {
	Type string
	Data map[string]interface{}
	Time int64
}

var eventBus chan Event

const MaxEventBufferSize = 1000

func init() {
	eventBus = make(chan Event, MaxEventBufferSize)
}

func PublishEvent(eventType string, data map[string]interface{}) {
	event := Event{
		Type: eventType,
		Data: data,
		Time: time.Now().Unix(),
	}
	select {
	case eventBus <- event:
	default:
		fmt.Println("Буфер событий переполнен. Событие пропущено.")
	}
}

func SubscribeEvents() <-chan Event {
	return eventBus
}
