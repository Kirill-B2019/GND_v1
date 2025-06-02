package core

import (
	"errors"
	"fmt"
	"sync"
)

// Mempool - потокобезопасный пул неподтверждённых транзакций
type Mempool struct {
	txsSlice []*Transaction  // Для сохранения порядка FIFO
	txsMap   map[string]bool // Для быстрой проверки существования
	mutex    sync.RWMutex
}

// NewMempool создает новый пустой мемпул
func NewMempool() *Mempool {
	return &Mempool{
		txsSlice: make([]*Transaction, 0),
		txsMap:   make(map[string]bool),
	}
}

// Add добавляет транзакцию в мемпул если её нет
func (mp *Mempool) Add(tx *Transaction) error {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	if mp.txsMap[tx.Hash] {
		return fmt.Errorf("транзакция %s уже существует", tx.Hash)
	}

	mp.txsSlice = append(mp.txsSlice, tx)
	mp.txsMap[tx.Hash] = true
	return nil
}

// Pop извлекает транзакцию в порядке FIFO
func (mp *Mempool) Pop() (*Transaction, error) {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	for len(mp.txsSlice) > 0 {
		// Извлекаем первую транзакцию
		tx := mp.txsSlice[0]
		mp.txsSlice = mp.txsSlice[1:]

		// Проверяем что транзакция не была удалена
		if mp.txsMap[tx.Hash] {
			delete(mp.txsMap, tx.Hash)
			return tx, nil
		}
	}

	return nil, errors.New("мемпул пуст")
}

// Remove удаляет транзакцию по хешу
func (mp *Mempool) Remove(hash string) {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()
	delete(mp.txsMap, hash)
}

// Get проверяет существование транзакции
func (mp *Mempool) Get(hash string) bool {
	mp.mutex.RLock()
	defer mp.mutex.RUnlock()
	return mp.txsMap[hash]
}

// All возвращает все актуальные транзакции
func (mp *Mempool) All() []*Transaction {
	mp.mutex.RLock()
	defer mp.mutex.RUnlock()

	result := make([]*Transaction, 0, len(mp.txsSlice))
	for _, tx := range mp.txsSlice {
		if mp.txsMap[tx.Hash] {
			result = append(result, tx)
		}
	}
	return result
}

// Size возвращает количество актуальных транзакций
func (mp *Mempool) Size() int {
	mp.mutex.RLock()
	defer mp.mutex.RUnlock()
	return len(mp.txsMap)
}

// Clear полностью очищает мемпул
func (mp *Mempool) Clear() {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()
	mp.txsSlice = make([]*Transaction, 0)
	mp.txsMap = make(map[string]bool)
}

func (mp *Mempool) Exists(hash string) bool {
	mp.mutex.RLock()
	defer mp.mutex.RUnlock()
	_, ok := mp.txsMap[hash]
	return ok
}
