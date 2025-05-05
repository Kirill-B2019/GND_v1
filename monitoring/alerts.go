package monitoring

import (
	"log"
	"os"
	"sync"
	"time"
)

// AlertLevel определяет уровень тревоги/предупреждения
type AlertLevel string

const (
	AlertInfo     AlertLevel = "INFO"
	AlertWarning  AlertLevel = "WARNING"
	AlertError    AlertLevel = "ERROR"
	AlertCritical AlertLevel = "CRITICAL"
)

// Alert описывает структуру предупреждения/тревоги
type Alert struct {
	Timestamp time.Time  // Время возникновения
	Level     AlertLevel // Уровень тревоги
	Component string     // Компонент системы (например: "Consensus", "Node", "SmartContract")
	Message   string     // Описание события
	Details   string     // Дополнительные детали (JSON, текст)
}

// AlertManager управляет отправкой и хранением алертов
type AlertManager struct {
	alerts    []Alert
	alertChan chan Alert
	mutex     sync.RWMutex
	logger    *log.Logger
}

// NewAlertManager создает новый менеджер алертов, логирует в файл и канал
func NewAlertManager(logFile string) (*AlertManager, error) {
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	logger := log.New(file, "", log.LstdFlags)
	am := &AlertManager{
		alerts:    make([]Alert, 0, 100),
		alertChan: make(chan Alert, 100),
		logger:    logger,
	}
	// Запуск фоновой горутины для обработки алертов
	go am.processAlerts()
	return am, nil
}

// processAlerts обрабатывает входящие алерты (логирование и хранение)
func (am *AlertManager) processAlerts() {
	for alert := range am.alertChan {
		am.mutex.Lock()
		if len(am.alerts) >= 1000 {
			am.alerts = am.alerts[1:] // кольцевой буфер
		}
		am.alerts = append(am.alerts, alert)
		am.mutex.Unlock()
		// Логируем алерт
		am.logger.Printf("[%s] [%s] [%s] %s | %s\n", alert.Timestamp.Format(time.RFC3339), alert.Level, alert.Component, alert.Message, alert.Details)
	}
}

// SendAlert отправляет алерт в систему мониторинга
func (am *AlertManager) SendAlert(level AlertLevel, component, message, details string) {
	alert := Alert{
		Timestamp: time.Now(),
		Level:     level,
		Component: component,
		Message:   message,
		Details:   details,
	}
	am.alertChan <- alert
}

// ListAlerts возвращает список последних N алертов
func (am *AlertManager) ListAlerts(n int) []Alert {
	am.mutex.RLock()
	defer am.mutex.RUnlock()
	if n > len(am.alerts) {
		n = len(am.alerts)
	}
	return append([]Alert(nil), am.alerts[len(am.alerts)-n:]...)
}

// Close корректно завершает работу менеджера алертов
func (am *AlertManager) Close() {
	close(am.alertChan)
}

// Пример использования:
/*
func main() {
	am, err := NewAlertManager("alerts.log")
	if err != nil {
		panic(err)
	}
	defer am.Close()

	am.SendAlert(AlertWarning, "Consensus", "Leader changed", "New leader: GND1abc...")
	am.SendAlert(AlertError, "Node", "Node offline", "Node GND2xyz lost connection")
	am.SendAlert(AlertCritical, "SmartContract", "Execution failed", "Tx: 0xabc, error: out of gas")

	time.Sleep(1 * time.Second) // дать время обработать алерты

	for _, alert := range am.ListAlerts(10) {
		fmt.Printf("[%s] %s: %s\n", alert.Level, alert.Component, alert.Message)
	}
}
*/
