package core

import (
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"
)

type Address string

// Account — структура аккаунта (кошелька/контракта)
type Account struct {
	Address      Address
	Balance      *big.Int
	Nonce        uint64
	Storage      map[string]*big.Int // Состояние контракта
	Code         []byte              // Байткод контракта
	IsContract   bool                // Флаг: контракт или обычный аккаунт
	Owner        Address             // Владелец контракта
	ABI          string              // ABI контракта (JSON-строка)
	MetadataCID  string              // CID метаданных (например, IPFS CID)
	Description  string              // Описание контракта или токена
	Version      string              // Версия контракта
	Compiler     string              // Версия компилятора
	SourceCode   string              // Исходный код контракта
	License      string              // Лицензия исходного кода (например, MIT, GPL)
	Tags         []string            // Теги или ключевые слова (например, ["DeFi", "NFT"])
	CreationTime time.Time           // Дата и время создания аккаунта/контракта
	// Можно добавить любые другие поля по необходимости
}

// State — хранит все аккаунты и их состояние
type State struct {
	accounts map[Address]*Account
	mutex    sync.RWMutex
}

// NewState создает новое пустое состояние
func NewState() *State {
	return &State{
		accounts: make(map[Address]*Account),
	}
}

// Получить баланс
func (s *State) GetBalance(addr Address) *big.Int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if acc, ok := s.accounts[addr]; ok {
		return acc.Balance
	}
	return big.NewInt(0)
}

// Списать баланс
func (s *State) SubBalance(addr Address, amount *big.Int) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if acc, ok := s.accounts[addr]; ok {
		if acc.Balance.Cmp(amount) >= 0 {
			acc.Balance.Sub(acc.Balance, amount)
			return true
		}
	}
	return false
}

// Начислить баланс
func (s *State) Credit(addr Address, amount *big.Int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	acc, ok := s.accounts[addr]
	if !ok {
		acc = &Account{
			Address: addr,
			Balance: big.NewInt(0),
			Storage: make(map[string]*big.Int),
		}
		s.accounts[addr] = acc
	}
	acc.Balance.Add(acc.Balance, amount)
}

// Получить nonce
func (s *State) GetNonce(addr Address) uint64 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if acc, ok := s.accounts[addr]; ok {
		return acc.Nonce
	}
	return 0
}

// Инкрементировать nonce
func (s *State) IncNonce(addr Address) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if acc, ok := s.accounts[addr]; ok {
		acc.Nonce++
	}
}

func (s *State) CallStatic(from, to Address, data []byte, gasLimit, gasPrice, value uint64) ([]byte, error) {
	// Здесь должна быть логика вызова контракта "на чтение".
	// Пока можно вернуть ошибку или пустой результат.
	return nil, errors.New("CallStatic не реализован")
}

// ApplyTransaction применяет транзакцию к состоянию
func (s *State) ApplyTransaction(tx *Transaction) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	from := Address(tx.From)
	to := Address(tx.To)

	if !ValidateAddress(tx.From) || !ValidateAddress(tx.To) {
		fmt.Printf("Некорректные адреса: from=%s, to=%s\n", tx.From, tx.To)
		return false
	}

	accFrom, ok := s.accounts[from]
	if !ok {
		fmt.Printf("Отправитель не найден: %s\n", tx.From)
		return false
	}
	if tx.Nonce != accFrom.Nonce {
		fmt.Printf("Неверный nonce: ожидается %d, получен %d\n", accFrom.Nonce, tx.Nonce)
		return false
	}
	totalCost := new(big.Int).Add(big.NewInt(int64(tx.Value)), new(big.Int).Mul(big.NewInt(int64(tx.GasPrice)), big.NewInt(int64(tx.GasLimit))))
	if accFrom.Balance.Cmp(totalCost) < 0 {
		fmt.Printf("Недостаточно средств: требуется %s, доступно %s\n", totalCost.String(), accFrom.Balance.String())
		return false
	}
	accFrom.Balance.Sub(accFrom.Balance, totalCost)
	accFrom.Nonce++

	accTo, ok := s.accounts[to]
	if !ok {
		accTo = &Account{
			Address: to,
			Balance: big.NewInt(0),
			Storage: make(map[string]*big.Int),
		}
		s.accounts[to] = accTo
	}
	accTo.Balance.Add(accTo.Balance, big.NewInt(int64(tx.Value)))

	return true
}
