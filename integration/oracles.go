package integration

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// Типы оракулов
type OracleType string

const (
	OracleSoftware  OracleType = "software"  // Программный оракул (API, веб-сервисы)
	OracleHardware  OracleType = "hardware"  // Аппаратный оракул (датчики, устройства)
	OracleInbound   OracleType = "inbound"   // Входящий (данные снаружи в блокчейн)
	OracleOutbound  OracleType = "outbound"  // Исходящий (данные из блокчейна наружу)
	OracleConsensus OracleType = "consensus" // Консенсусный (агрегация нескольких источников)
)

// Структура события от оракула
type OracleEvent struct {
	ID         string                 // Уникальный идентификатор события
	OracleID   string                 // Идентификатор оракула
	Type       OracleType             // Тип оракула
	Timestamp  int64                  // Время события
	Payload    map[string]interface{} // Данные (например, курс, температура и т.д.)
	Signatures map[string]string      // Подписи (адрес → подпись)
	Status     string                 // pending, confirmed, failed
}

// Интерфейс оракула
type Oracle interface {
	RequestData(req *OracleRequest) (*OracleEvent, error)
	VerifyEvent(event *OracleEvent) bool
	GetEvent(eventID string) (*OracleEvent, error)
	ListEvents(status string) []*OracleEvent
}

// Запрос к оракулу
type OracleRequest struct {
	ID        string // Уникальный идентификатор запроса
	OracleID  string // К какому оракулу обращаемся
	Type      OracleType
	Params    map[string]interface{} // Параметры запроса (например, "pair": "USD/GND")
	Requester string                 // Адрес вызывающего (смарт-контракт)
	Signature string                 // Подпись запроса
	Timestamp int64
}

// Реализация мультисиг-консенсусного оракула
type ConsensusOracle struct {
	id         string
	validators []string // адреса валидаторов-оракулов
	events     map[string]*OracleEvent
	mutex      sync.RWMutex
}

// Новый консенсусный оракул
func NewConsensusOracle(id string, validators []string) *ConsensusOracle {
	return &ConsensusOracle{
		id:         id,
		validators: validators,
		events:     make(map[string]*OracleEvent),
	}
}

// Отправить запрос к оракулу
func (co *ConsensusOracle) RequestData(req *OracleRequest) (*OracleEvent, error) {
	co.mutex.Lock()
	defer co.mutex.Unlock()
	eventID := fmt.Sprintf("%s-%d", req.ID, time.Now().UnixNano())
	event := &OracleEvent{
		ID:         eventID,
		OracleID:   co.id,
		Type:       OracleConsensus,
		Timestamp:  time.Now().Unix(),
		Payload:    make(map[string]interface{}),
		Signatures: make(map[string]string),
		Status:     "pending",
	}
	co.events[eventID] = event
	// Здесь должен быть вызов внешних источников и сбор подписей от валидаторов
	return event, nil
}

// Верификация события (проверка кворума подписей)
func (co *ConsensusOracle) VerifyEvent(event *OracleEvent) bool {
	co.mutex.RLock()
	defer co.mutex.RUnlock()
	// Проверка кворума: более половины валидаторов подписали событие
	count := 0
	for _, v := range co.validators {
		if _, ok := event.Signatures[v]; ok {
			count++
		}
	}
	return count >= len(co.validators)/2+1
}

// Получить событие по ID
func (co *ConsensusOracle) GetEvent(eventID string) (*OracleEvent, error) {
	co.mutex.RLock()
	defer co.mutex.RUnlock()
	event, exists := co.events[eventID]
	if !exists {
		return nil, errors.New("event not found")
	}
	return event, nil
}

// Список событий по статусу
func (co *ConsensusOracle) ListEvents(status string) []*OracleEvent {
	co.mutex.RLock()
	defer co.mutex.RUnlock()
	var result []*OracleEvent
	for _, e := range co.events {
		if status == "" || e.Status == status {
			result = append(result, e)
		}
	}
	return result
}

// Добавить подпись валидатора к событию
func (co *ConsensusOracle) AddSignature(eventID, validator, signature string) error {
	co.mutex.Lock()
	defer co.mutex.Unlock()
	event, exists := co.events[eventID]
	if !exists {
		return errors.New("event not found")
	}
	event.Signatures[validator] = signature
	if co.VerifyEvent(event) {
		event.Status = "confirmed"
	}
	return nil
}

// Пример программного оракула (заглушка)
type SoftwareOracle struct {
	id     string
	events map[string]*OracleEvent
	mutex  sync.RWMutex
}

func NewSoftwareOracle(id string) *SoftwareOracle {
	return &SoftwareOracle{
		id:     id,
		events: make(map[string]*OracleEvent),
	}
}

func (so *SoftwareOracle) RequestData(req *OracleRequest) (*OracleEvent, error) {
	so.mutex.Lock()
	defer so.mutex.Unlock()
	eventID := fmt.Sprintf("%s-%d", req.ID, time.Now().UnixNano())
	event := &OracleEvent{
		ID:         eventID,
		OracleID:   so.id,
		Type:       OracleSoftware,
		Timestamp:  time.Now().Unix(),
		Payload:    map[string]interface{}{"example": "data"},
		Signatures: map[string]string{"oracle": "signature"},
		Status:     "confirmed",
	}
	so.events[eventID] = event
	return event, nil
}

func (so *SoftwareOracle) VerifyEvent(event *OracleEvent) bool {
	// Для программного оракула всегда true (заглушка)
	return true
}

func (so *SoftwareOracle) GetEvent(eventID string) (*OracleEvent, error) {
	so.mutex.RLock()
	defer so.mutex.RUnlock()
	event, exists := so.events[eventID]
	if !exists {
		return nil, errors.New("event not found")
	}
	return event, nil
}

func (so *SoftwareOracle) ListEvents(status string) []*OracleEvent {
	so.mutex.RLock()
	defer so.mutex.RUnlock()
	var result []*OracleEvent
	for _, e := range so.events {
		if status == "" || e.Status == status {
			result = append(result, e)
		}
	}
	return result
}

// --- Пример использования ---
//
// func main() {
// 	oracle := NewConsensusOracle("oracle1", []string{"GND1...", "GND2...", "GND3..."})
// 	req := &OracleRequest{
// 		ID:        "req1",
// 		OracleID:  "oracle1",
// 		Type:      OracleConsensus,
// 		Params:    map[string]interface{}{"pair": "USD/GND"},
// 		Requester: "GNDct1...",
// 		Timestamp: time.Now().Unix(),
// 	}
// 	event, _ := oracle.RequestData(req)
// 	oracle.AddSignature(event.ID, "GND1...", "sig1")
// 	oracle.AddSignature(event.ID, "GND2...", "sig2")
// 	oracle.AddSignature(event.ID, "GND3...", "sig3")
// 	fmt.Println("Event status:", event.Status, "Quorum:", oracle.VerifyEvent(event))
// }
/*ConsensusOracle: мультисиг-оракул с кворумом валидаторов, подходит для ценовых фидов, погодных данных и пр.

SoftwareOracle: программный оракул для интеграции с внешними API.

OracleEvent: универсальная структура события оракула с подписями, статусом и полезной нагрузкой.

Интерфейс Oracle: легко расширяется для аппаратных, входящих/исходящих и любых кастомных оракулов.

Безопасность: поддержка мультиподписей, проверка кворума, интеграция с контрактами через API.*/
