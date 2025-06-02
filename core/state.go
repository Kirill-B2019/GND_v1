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
	Balances     map[string]*big.Int // ключ — символ монеты
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
func (s *State) GetBalance(addr Address, symbol string) *big.Int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if acc, ok := s.accounts[addr]; ok {
		if bal, ok := acc.Balances[symbol]; ok {
			return new(big.Int).Set(bal)
		}
	}
	return big.NewInt(0)
}

// Списать баланс
func (s *State) SubBalance(addr Address, symbol string, amount *big.Int) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if acc, ok := s.accounts[addr]; ok {
		if bal, ok := acc.Balances[symbol]; ok && bal.Cmp(amount) >= 0 {
			acc.Balances[symbol].Sub(bal, amount)
			return true
		}
	}
	return false
}

// Начислить баланс
func (s *State) Credit(addr Address, symbol string, amount *big.Int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	acc, ok := s.accounts[addr]
	if !ok {
		acc = &Account{
			Address:  addr,
			Balances: make(map[string]*big.Int),
		}
		s.accounts[addr] = acc
	}
	if _, ok := acc.Balances[symbol]; !ok {
		acc.Balances[symbol] = big.NewInt(0)
	}
	acc.Balances[symbol].Add(acc.Balances[symbol], amount)
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
	symbol := tx.Symbol // Используем символ из транзакции

	// Валидация основных параметров
	if !ValidateAddress(tx.From) || !ValidateAddress(tx.To) || symbol == "" {
		fmt.Println("Invalid transaction parameters")
		return false
	}

	// Получаем аккаунты
	accFrom, ok := s.accounts[from]
	if !ok {
		fmt.Printf("Sender not found: %s\n", tx.From)
		return false
	}

	// Проверка nonce
	if tx.Nonce != accFrom.Nonce {
		fmt.Printf("Invalid nonce: expected %d, got %d\n", accFrom.Nonce, tx.Nonce)
		return false
	}

	// Расчет комиссии
	gasCost := new(big.Int).Mul(
		big.NewInt(int64(tx.GasPrice)),
		big.NewInt(int64(tx.GasLimit)),
	)
	totalCost := new(big.Int).Add(tx.Value, gasCost)

	// Проверка баланса
	balance := accFrom.Balances[symbol]
	if balance == nil || balance.Cmp(totalCost) < 0 {
		fmt.Printf("Insufficient %s balance: need %s, have %s\n",
			symbol,
			totalCost.String(),
			balance.String())
		return false
	}

	// Списание средств
	accFrom.Balances[symbol].Sub(balance, totalCost)
	accFrom.Nonce++

	// Зачисление получателю
	accTo, ok := s.accounts[to]
	if !ok {
		accTo = &Account{
			Address:  to,
			Balances: make(map[string]*big.Int),
		}
		s.accounts[to] = accTo
	}

	if accTo.Balances[symbol] == nil {
		accTo.Balances[symbol] = big.NewInt(0)
	}
	accTo.Balances[symbol].Add(accTo.Balances[symbol], tx.Value)

	return true
}
