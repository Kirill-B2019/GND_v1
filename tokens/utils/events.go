// tokens/utils/events.go

package utils

import (
	"time"
)

type Event struct {
	Type string
	Data map[string]interface{}
	Time int64
}

var eventBus = make(chan Event, 1000)

func PublishEvent(eventType string, data map[string]interface{}) {
	eventBus <- Event{
		Type: eventType,
		Data: data,
		Time: time.Now().Unix(),
	}
}

func SubscribeEvents() <-chan Event {
	return eventBus
}
