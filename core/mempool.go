package core

import (
	"errors"
	"log"
	"sync"
	"time"
)

type Mempool struct {
	txChan chan *Transaction
	mu     sync.RWMutex
	txMap  map[string]*Transaction
	logger *log.Logger
}

func NewMempool() *Mempool {
	return &Mempool{
		txChan: make(chan *Transaction, 10000), // буфер на 10 000 транзакций
		txMap:  make(map[string]*Transaction),
		logger: log.New(log.Writer(), "[Mempool] ", log.LstdFlags),
	}
}

func (m *Mempool) Add(tx *Transaction) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Проверяем, не существует ли уже такая транзакция
	if _, exists := m.txMap[tx.ID]; exists {
		return errors.New("transaction already exists")
	}

	select {
	case m.txChan <- tx:
		m.txMap[tx.ID] = tx
		m.logger.Printf("Transaction %s added to mempool", tx.ID)
		return nil
	default:
		return errors.New("mempool is full")
	}
}

func (m *Mempool) Pop() (*Transaction, error) {
	select {
	case tx := <-m.txChan:
		m.mu.Lock()
		delete(m.txMap, tx.ID)
		m.mu.Unlock()
		m.logger.Printf("Transaction %s popped from mempool", tx.ID)
		return tx, nil
	case <-time.After(100 * time.Millisecond):
		return nil, errors.New("timeout")
	}
}

func (m *Mempool) Size() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.txMap)
}

func (m *Mempool) Exists(txID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, exists := m.txMap[txID]
	return exists
}

// GetTransaction возвращает транзакцию по ID
func (m *Mempool) GetTransaction(txID string) (*Transaction, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if tx, exists := m.txMap[txID]; exists {
		return tx, nil
	}
	return nil, errors.New("transaction not found")
}

// Clear очищает мемпул
func (m *Mempool) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Очищаем канал
	for len(m.txChan) > 0 {
		<-m.txChan
	}

	// Очищаем карту
	m.txMap = make(map[string]*Transaction)
	m.logger.Println("Mempool cleared")
}

// GetPendingTransactions возвращает список всех ожидающих транзакций
func (m *Mempool) GetPendingTransactions() []*Transaction {
	m.mu.RLock()
	defer m.mu.RUnlock()

	txs := make([]*Transaction, 0, len(m.txMap))
	for _, tx := range m.txMap {
		txs = append(txs, tx)
	}
	return txs
}
