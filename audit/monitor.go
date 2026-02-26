package audit

import (
	"GND/core"
	"log"
	"math/big"
	"time"
)

// SuspiciousTx описывает подозрительную транзакцию
type SuspiciousTx struct {
	Tx        *core.Transaction
	Reason    string
	Timestamp time.Time
}

// Monitor — структура для мониторинга событий и транзакций
type Monitor struct {
	Suspicious       []SuspiciousTx
	Threshold        *big.Int             // Порог для крупных переводов
	lastTxTimestamps map[string]time.Time // ключ — адрес отправителя
	knownAddresses   map[string]bool
}

// NewMonitor создает новый монитор с заданным порогом
func NewMonitor(threshold *big.Int) *Monitor {
	return &Monitor{
		Suspicious:       make([]SuspiciousTx, 0),
		Threshold:        threshold,
		lastTxTimestamps: make(map[string]time.Time),
		knownAddresses:   make(map[string]bool),
	}
}

// CheckTransaction анализирует транзакцию на предмет подозрительности
func (m *Monitor) CheckTransaction(tx *core.Transaction) {
	if tx == nil {
		return
	}
	// Проверка на крупную сумму (с защитой от nil)
	if m.Threshold != nil && tx.Value != nil && tx.Value.Cmp(m.Threshold) >= 0 {
		m.AddSuspicious(tx, "Крупная сумма перевода")
	}

	// Перевод самому себе
	if tx.Sender.String() == tx.Recipient.String() {
		m.AddSuspicious(tx, "Перевод самому себе")
	}

	// Частые переводы
	now := time.Now()
	senderStr := tx.Sender.String()
	recipientStr := tx.Recipient.String()
	lastTime, exists := m.lastTxTimestamps[senderStr]
	if exists && now.Sub(lastTime) < 10*time.Second {
		m.AddSuspicious(tx, "Частые переводы с одного адреса (возможно, бот-активность)")
	}
	m.lastTxTimestamps[senderStr] = now

	// Перевод на новый/неизвестный адрес
	if !m.knownAddresses[recipientStr] {
		m.AddSuspicious(tx, "Перевод на новый/неизвестный адрес")
	}
	m.knownAddresses[recipientStr] = true
	m.knownAddresses[senderStr] = true
}

// AddSuspicious добавляет подозрительную транзакцию в журнал
func (m *Monitor) AddSuspicious(tx *core.Transaction, reason string) {
	entry := SuspiciousTx{
		Tx:        tx,
		Reason:    reason,
		Timestamp: time.Now(),
	}
	m.Suspicious = append(m.Suspicious, entry)
	log.Printf("[MONITOR] Обнаружена подозрительная транзакция: %s — %s", tx.Hash, reason)
}

// GetSuspicious возвращает список всех подозрительных транзакций
func (m *Monitor) GetSuspicious() []SuspiciousTx {
	return m.Suspicious
}

// Clear очищает журнал подозрительных транзакций
func (m *Monitor) Clear() {
	m.Suspicious = make([]SuspiciousTx, 0)
}
