package core

import (
	"sync"
	"time"
)

// Metrics содержит все метрики блокчейна
type Metrics struct {
	mu sync.RWMutex

	// Метрики блоков
	BlockMetrics struct {
		TotalBlocks      uint64
		BlocksPerMinute  float64
		AverageBlockTime time.Duration
		LastBlockTime    time.Time
	}

	// Метрики транзакций
	TransactionMetrics struct {
		TotalTransactions     uint64
		TransactionsPerMinute float64
		PendingTransactions   uint64
		FailedTransactions    uint64
		AverageGasPrice       float64
		AverageGasUsed        float64
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

	metrics = &Metrics{}
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

	if status == "failed" {
		m.TransactionMetrics.FailedTransactions++
	}

	m.TransactionMetrics.AverageGasPrice = (m.TransactionMetrics.AverageGasPrice + float64(tx.GasPrice)) / 2
	m.TransactionMetrics.AverageGasUsed = (m.TransactionMetrics.AverageGasUsed + float64(tx.GasUsed)) / 2
}

// UpdateNetworkMetrics обновляет метрики сети
func (m *Metrics) UpdateNetworkMetrics(peers uint64, latency time.Duration, bytesReceived, bytesSent uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.NetworkMetrics.ActivePeers = peers
	m.NetworkMetrics.NetworkLatency = latency
	m.NetworkMetrics.BytesReceived = bytesReceived
	m.NetworkMetrics.BytesSent = bytesSent
}

// UpdatePerformanceMetrics обновляет метрики производительности
func (m *Metrics) UpdatePerformanceMetrics(cpu, memory, disk uint64, dbLatency, apiLatency time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.PerformanceMetrics.CPUUsage = float64(cpu)
	m.PerformanceMetrics.MemoryUsage = memory
	m.PerformanceMetrics.DiskUsage = disk
	m.PerformanceMetrics.DatabaseLatency = dbLatency
	m.PerformanceMetrics.APIResponseTime = apiLatency
}

// UpdateConsensusMetrics обновляет метрики консенсуса
func (m *Metrics) UpdateConsensusMetrics(validators, activeValidators uint64, latency time.Duration, missedBlocks, forks uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ConsensusMetrics.ValidatorsCount = validators
	m.ConsensusMetrics.ActiveValidators = activeValidators
	m.ConsensusMetrics.ConsensusLatency = latency
	m.ConsensusMetrics.MissedBlocks = missedBlocks
	m.ConsensusMetrics.ForkCount = forks
}

// GetMetrics возвращает текущие метрики
func GetMetrics() *Metrics {
	metricsMu.RLock()
	defer metricsMu.RUnlock()
	return metrics
}
