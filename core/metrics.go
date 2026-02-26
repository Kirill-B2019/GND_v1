package core

import (
	"math/big"
	"sync"
	"time"
)

// TransactionTypeMetrics содержит метрики для конкретного типа транзакций
type TransactionTypeMetrics struct {
	Count           uint64
	SuccessCount    uint64
	FailedCount     uint64
	TotalFee        *big.Int
	AverageFee      float64
	MinFee          *big.Int
	MaxFee          *big.Int
	LastMinuteCount uint64
	LastHourCount   uint64
}

// Metrics содержит все метрики блокчейна
type Metrics struct {
	mu sync.RWMutex

	// Метрики блоков
	BlockMetrics struct {
		TotalBlocks      uint64
		BlocksPerMinute  float64
		AverageBlockTime time.Duration
		LastBlockTime    time.Time
		BlockSize        uint64
		GasUsed          uint64
		GasLimit         uint64
	}

	// Метрики транзакций
	TransactionMetrics struct {
		TotalTransactions     uint64
		TransactionsPerMinute float64
		PendingTransactions   uint64
		FailedTransactions    uint64
		AverageFee            float64
		MinFee                *big.Int
		MaxFee                *big.Int
		TotalFee              *big.Int
		LastMinuteCount       uint64
		LastHourCount         uint64
		TypeMetrics           map[string]*TransactionTypeMetrics // Метрики по типам транзакций
		StatusMetrics         map[string]uint64                  // Метрики по статусам
		FeeDistribution       map[string]uint64                  // Распределение комиссий
	}

	// Метрики сети
	NetworkMetrics struct {
		ActivePeers       uint64
		NetworkLatency    time.Duration
		BytesReceived     uint64
		BytesSent         uint64
		RequestsPerMinute float64
	}

	// Метрики производительности
	PerformanceMetrics struct {
		CPUUsage        float64
		MemoryUsage     uint64
		DiskUsage       uint64
		DatabaseLatency time.Duration
		APIResponseTime time.Duration
	}

	// Метрики консенсуса
	ConsensusMetrics struct {
		ValidatorsCount  uint64
		ActiveValidators uint64
		ConsensusLatency time.Duration
		MissedBlocks     uint64
		ForkCount        uint64
	}

	// Алерты
	Alerts struct {
		HighFeeThreshold     *big.Int
		LowFeeThreshold      *big.Int
		HighLatencyThreshold time.Duration
		HighCPUThreshold     float64
		HighMemoryThreshold  uint64
		AlertHistory         []Alert
	}
}

// Alert представляет собой алерт
type Alert struct {
	Type      string
	Message   string
	Value     interface{}
	Threshold interface{}
	Timestamp time.Time
}

var (
	metrics   = &Metrics{}
	metricsMu sync.RWMutex
	startTime = time.Now()
)

// ResetMetrics сбрасывает все метрики
func ResetMetrics() {
	metricsMu.Lock()
	defer metricsMu.Unlock()

	metrics = &Metrics{
		TransactionMetrics: struct {
			TotalTransactions     uint64
			TransactionsPerMinute float64
			PendingTransactions   uint64
			FailedTransactions    uint64
			AverageFee            float64
			MinFee                *big.Int
			MaxFee                *big.Int
			TotalFee              *big.Int
			LastMinuteCount       uint64
			LastHourCount         uint64
			TypeMetrics           map[string]*TransactionTypeMetrics
			StatusMetrics         map[string]uint64
			FeeDistribution       map[string]uint64
		}{
			TypeMetrics:     make(map[string]*TransactionTypeMetrics),
			StatusMetrics:   make(map[string]uint64),
			FeeDistribution: make(map[string]uint64),
		},
	}
	startTime = time.Now()
}

// UpdateBlockMetrics обновляет метрики блоков
func (m *Metrics) UpdateBlockMetrics(block *Block) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	m.BlockMetrics.TotalBlocks++

	if !m.BlockMetrics.LastBlockTime.IsZero() {
		blockTime := now.Sub(m.BlockMetrics.LastBlockTime)
		m.BlockMetrics.AverageBlockTime = (m.BlockMetrics.AverageBlockTime + blockTime) / 2
	}

	m.BlockMetrics.LastBlockTime = now
	m.BlockMetrics.BlocksPerMinute = float64(m.BlockMetrics.TotalBlocks) / time.Since(startTime).Minutes()
}

// UpdateTransactionMetrics обновляет метрики транзакций
func (m *Metrics) UpdateTransactionMetrics(tx *Transaction, status string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TransactionMetrics.TotalTransactions++
	m.TransactionMetrics.TransactionsPerMinute = float64(m.TransactionMetrics.TotalTransactions) / time.Since(startTime).Minutes()

	// Обновляем метрики по статусу
	m.TransactionMetrics.StatusMetrics[status]++
	if status == "failed" {
		m.TransactionMetrics.FailedTransactions++
	}

	// Обновляем метрики по типу транзакции
	if tx != nil {
		// Инициализируем метрики для типа транзакции, если их еще нет
		if _, exists := m.TransactionMetrics.TypeMetrics[tx.Type]; !exists {
			m.TransactionMetrics.TypeMetrics[tx.Type] = &TransactionTypeMetrics{
				MinFee: new(big.Int).Set(tx.Fee),
				MaxFee: new(big.Int).Set(tx.Fee),
			}
		}

		typeMetrics := m.TransactionMetrics.TypeMetrics[tx.Type]
		typeMetrics.Count++
		if status == "success" {
			typeMetrics.SuccessCount++
		} else {
			typeMetrics.FailedCount++
		}

		// Обновляем метрики комиссий
		if tx.Fee != nil {
			// Общие метрики комиссий
			if m.TransactionMetrics.MinFee == nil || tx.Fee.Cmp(m.TransactionMetrics.MinFee) < 0 {
				m.TransactionMetrics.MinFee = new(big.Int).Set(tx.Fee)
			}
			if m.TransactionMetrics.MaxFee == nil || tx.Fee.Cmp(m.TransactionMetrics.MaxFee) > 0 {
				m.TransactionMetrics.MaxFee = new(big.Int).Set(tx.Fee)
			}
			if m.TransactionMetrics.TotalFee == nil {
				m.TransactionMetrics.TotalFee = new(big.Int)
			}
			m.TransactionMetrics.TotalFee.Add(m.TransactionMetrics.TotalFee, tx.Fee)
			m.TransactionMetrics.AverageFee = float64(m.TransactionMetrics.TotalFee.Uint64()) / float64(m.TransactionMetrics.TotalTransactions)

			// Метрики комиссий по типу транзакции
			if typeMetrics.MinFee == nil || tx.Fee.Cmp(typeMetrics.MinFee) < 0 {
				typeMetrics.MinFee = new(big.Int).Set(tx.Fee)
			}
			if typeMetrics.MaxFee == nil || tx.Fee.Cmp(typeMetrics.MaxFee) > 0 {
				typeMetrics.MaxFee = new(big.Int).Set(tx.Fee)
			}
			if typeMetrics.TotalFee == nil {
				typeMetrics.TotalFee = new(big.Int)
			}
			typeMetrics.TotalFee.Add(typeMetrics.TotalFee, tx.Fee)
			typeMetrics.AverageFee = float64(typeMetrics.TotalFee.Uint64()) / float64(typeMetrics.Count)

			// Распределение комиссий
			feeRange := getFeeRange(tx.Fee)
			m.TransactionMetrics.FeeDistribution[feeRange]++
		}

		// Проверяем алерты
		m.checkAlerts(tx)
	}
}

// getFeeRange возвращает диапазон комиссии
func getFeeRange(fee *big.Int) string {
	if fee == nil {
		return "unknown"
	}
	feeUint := fee.Uint64()
	switch {
	case feeUint < 1000:
		return "low"
	case feeUint < 10000:
		return "medium"
	case feeUint < 100000:
		return "high"
	default:
		return "very_high"
	}
}

// checkAlerts проверяет условия для алертов
func (m *Metrics) checkAlerts(tx *Transaction) {
	if tx == nil || tx.Fee == nil {
		return
	}

	// Проверка высокой комиссии
	if m.Alerts.HighFeeThreshold != nil && tx.Fee.Cmp(m.Alerts.HighFeeThreshold) > 0 {
		m.addAlert("high_fee", "Высокая комиссия за транзакцию", tx.Fee, m.Alerts.HighFeeThreshold)
	}

	// Проверка низкой комиссии
	if m.Alerts.LowFeeThreshold != nil && tx.Fee.Cmp(m.Alerts.LowFeeThreshold) < 0 {
		m.addAlert("low_fee", "Низкая комиссия за транзакцию", tx.Fee, m.Alerts.LowFeeThreshold)
	}
}

// addAlert добавляет новый алерт
func (m *Metrics) addAlert(alertType, message string, value, threshold interface{}) {
	alert := Alert{
		Type:      alertType,
		Message:   message,
		Value:     value,
		Threshold: threshold,
		Timestamp: time.Now(),
	}
	m.Alerts.AlertHistory = append(m.Alerts.AlertHistory, alert)
}

// GetMetrics возвращает текущие метрики
func GetMetrics() *Metrics {
	metricsMu.RLock()
	defer metricsMu.RUnlock()
	return metrics
}

// InitBlockMetricsFromBlock заполняет метрики блоков из текущего состояния цепи (при старте ноды).
// Если latest == nil, метрики не меняются.
func InitBlockMetricsFromBlock(latest *Block) {
	if latest == nil {
		return
	}
	metricsMu.Lock()
	defer metricsMu.Unlock()

	metrics.BlockMetrics.TotalBlocks = latest.Height + 1
	metrics.BlockMetrics.LastBlockTime = latest.Timestamp
	metrics.BlockMetrics.BlockSize = latest.Size
	metrics.BlockMetrics.GasUsed = latest.GasUsed
	metrics.BlockMetrics.GasLimit = latest.GasLimit
	elapsed := time.Since(startTime).Minutes()
	if elapsed > 0 {
		metrics.BlockMetrics.BlocksPerMinute = float64(metrics.BlockMetrics.TotalBlocks) / elapsed
	}
}

// SetAlertThresholds устанавливает пороговые значения для алертов
func SetAlertThresholds(highFee, lowFee *big.Int, highLatency time.Duration, highCPU float64, highMemory uint64) {
	metricsMu.Lock()
	defer metricsMu.Unlock()

	metrics.Alerts.HighFeeThreshold = highFee
	metrics.Alerts.LowFeeThreshold = lowFee
	metrics.Alerts.HighLatencyThreshold = highLatency
	metrics.Alerts.HighCPUThreshold = highCPU
	metrics.Alerts.HighMemoryThreshold = highMemory
}
