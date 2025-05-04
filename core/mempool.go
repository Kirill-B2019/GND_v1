package core

import (
	"sync"
)

// Mempool - потокобезопасный пул неподтверждённых транзакций
type Mempool struct {
	txs   map[string]*Transaction // ключ - хеш транзакции
	mutex sync.RWMutex
}

// NewMempool создает новый пустой мемпул
func NewMempool() *Mempool {
	return &Mempool{
		txs: make(map[string]*Transaction),
	}
}

// Add добавляет транзакцию в мемпул
func (mp *Mempool) Add(tx *Transaction) {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()
	mp.txs[tx.Hash] = tx
}

// Remove удаляет транзакцию по хешу
func (mp *Mempool) Remove(hash string) {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()
	delete(mp.txs, hash)
}

// Get возвращает транзакцию по хешу, если она есть
func (mp *Mempool) Get(hash string) (*Transaction, bool) {
	mp.mutex.RLock()
	defer mp.mutex.RUnlock()
	tx, ok := mp.txs[hash]
	return tx, ok
}

// All возвращает срез всех транзакций в мемпуле (копия)
func (mp *Mempool) All() []*Transaction {
	mp.mutex.RLock()
	defer mp.mutex.RUnlock()
	txs := make([]*Transaction, 0, len(mp.txs))
	for _, tx := range mp.txs {
		txs = append(txs, tx)
	}
	return txs
}

// Size возвращает количество транзакций в мемпуле
func (mp *Mempool) Size() int {
	mp.mutex.RLock()
	defer mp.mutex.RUnlock()
	return len(mp.txs)
}

// Exists проверяет, есть ли транзакция с таким хешем
func (mp *Mempool) Exists(hash string) bool {
	mp.mutex.RLock()
	defer mp.mutex.RUnlock()
	_, ok := mp.txs[hash]
	return ok
}

// Clear полностью очищает мемпул
func (mp *Mempool) Clear() {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()
	mp.txs = make(map[string]*Transaction)
}
