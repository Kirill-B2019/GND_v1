package core

import (
	"errors"
	"sync"
	"time"
)

type Mempool struct {
	txChan chan *Transaction
	mu     sync.RWMutex
	txMap  map[string]*Transaction
}

func NewMempool() *Mempool {
	return &Mempool{
		txChan: make(chan *Transaction, 10000), // буфер на 10 000 транзакций
	}
}

func (m *Mempool) Add(tx *Transaction) error {
	select {
	case m.txChan <- tx:
		return nil
	default:
		return errors.New("mempool is full")
	}
}

func (m *Mempool) Pop() (*Transaction, error) {
	select {
	case tx := <-m.txChan:
		return tx, nil
	case <-time.After(100 * time.Millisecond):
		return nil, errors.New("timeout")
	}
}

func (m *Mempool) Size() int {
	return len(m.txChan)
}

func (m *Mempool) Exists(txID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, exists := m.txMap[txID]
	return exists
}
