package storage

import (
	"math/big"
	"sync"
)

// ContractStorage реализует простое key-value хранилище для одного контракта.
// Ключи и значения — строки, значения могут быть любым типом (например, *big.Int или []byte).
type ContractStorage struct {
	data  map[string]*big.Int
	mutex sync.RWMutex
}

// NewContractStorage создает новое хранилище для контракта.
func NewContractStorage() *ContractStorage {
	return &ContractStorage{
		data: make(map[string]*big.Int),
	}
}

// Set сохраняет значение по ключу.
func (s *ContractStorage) Set(key string, value *big.Int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.data[key] = new(big.Int).Set(value) // копируем значение
}

// Get возвращает значение по ключу (или nil, если такого ключа нет).
func (s *ContractStorage) Get(key string) *big.Int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	val, ok := s.data[key]
	if !ok {
		return nil
	}
	return new(big.Int).Set(val) // возвращаем копию
}

// Delete удаляет значение по ключу.
func (s *ContractStorage) Delete(key string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.data, key)
}

// Keys возвращает список всех ключей.
func (s *ContractStorage) Keys() []string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	keys := make([]string, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}
	return keys
}
