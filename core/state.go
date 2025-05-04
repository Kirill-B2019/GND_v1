package core

import (
	"fmt"
	"sync"
)

// State хранит состояние блокчейна: балансы, nonce и другую информацию
type State struct {
	balances map[string]uint64 // Балансы адресов (GND и др. токены можно расширить)
	nonces   map[string]uint64 // Нонсы для предотвращения повторных транзакций
	mutex    sync.RWMutex
}

// NewState создает новое пустое состояние
func NewState() *State {
	return &State{
		balances: make(map[string]uint64),
		nonces:   make(map[string]uint64),
	}
}

// BalanceOf возвращает баланс по адресу
func (s *State) BalanceOf(addr string) uint64 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.balances[addr]
}

// NonceOf возвращает текущий nonce по адресу
func (s *State) NonceOf(addr string) uint64 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.nonces[addr]
}

// Credit увеличивает баланс адреса на amount
func (s *State) Credit(addr string, amount uint64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.balances[addr] += amount
}

// Debit уменьшает баланс адреса на amount, возвращает false если недостаточно средств
func (s *State) Debit(addr string, amount uint64) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.balances[addr] < amount {
		return false
	}
	s.balances[addr] -= amount
	return true
}

// ApplyTransaction применяет транзакцию к состоянию, возвращает true если успешно
func (s *State) ApplyTransaction(tx *Transaction) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Проверка nonce
	expectedNonce := s.nonces[tx.From]
	if tx.Nonce != expectedNonce {
		fmt.Printf("Неверный nonce для %s: ожидается %d, получен %d\n", tx.From, expectedNonce, tx.Nonce)
		return false
	}

	// Вычисление полной стоимости (value + комиссия)
	totalCost := tx.Value + tx.GasPrice*tx.GasLimit
	if s.balances[tx.From] < totalCost {
		fmt.Printf("Недостаточно средств у %s: требуется %d, доступно %d\n", tx.From, totalCost, s.balances[tx.From])
		return false
	}

	// Списание средств с отправителя
	s.balances[tx.From] -= totalCost

	// Начисление получателю (только value, комиссия пойдет майнеру/валидатору отдельно)
	s.balances[tx.To] += tx.Value

	// Увеличение nonce отправителя
	s.nonces[tx.From]++

	// TODO: обработка вызова смарт-контракта через VM, начисление комиссии валидатору

	return true
}
