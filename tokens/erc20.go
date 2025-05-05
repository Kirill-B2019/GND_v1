package tokens

import (
	"errors"
	"sync"
)

// ERC20TokenInterface определяет стандартные методы ERC-20
type ERC20TokenInterface interface {
	Name() string
	Symbol() string
	Decimals() uint8
	TotalSupply() uint64
	BalanceOf(address string) uint64
	Transfer(from, to string, amount uint64) error
	Approve(owner, spender string, amount uint64) error
	Allowance(owner, spender string) uint64
	TransferFrom(spender, from, to string, amount uint64) error
}

// ERC20Token реализует стандарт ERC-20
type ERC20Token struct {
	name        string
	symbol      string
	decimals    uint8
	totalSupply uint64
	balances    map[string]uint64
	allowances  map[string]map[string]uint64 // owner → spender → amount
	mutex       sync.RWMutex
}

// NewERC20Token создает новый ERC-20 токен
func NewERC20Token(name, symbol string, decimals uint8, initialSupply uint64, owner string) *ERC20Token {
	token := &ERC20Token{
		name:        name,
		symbol:      symbol,
		decimals:    decimals,
		totalSupply: initialSupply,
		balances:    make(map[string]uint64),
		allowances:  make(map[string]map[string]uint64),
	}
	token.balances[owner] = initialSupply
	return token
}

// Name возвращает имя токена
func (t *ERC20Token) Name() string {
	return t.name
}

// Symbol возвращает символ токена
func (t *ERC20Token) Symbol() string {
	return t.symbol
}

// Decimals возвращает количество знаков после запятой
func (t *ERC20Token) Decimals() uint8 {
	return t.decimals
}

// TotalSupply возвращает общий объем выпуска
func (t *ERC20Token) TotalSupply() uint64 {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.totalSupply
}

// BalanceOf возвращает баланс по адресу
func (t *ERC20Token) BalanceOf(address string) uint64 {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.balances[address]
}

// Transfer переводит amount токенов от from к to
func (t *ERC20Token) Transfer(from, to string, amount uint64) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if t.balances[from] < amount {
		return errors.New("insufficient balance")
	}
	t.balances[from] -= amount
	t.balances[to] += amount
	return nil
}

// Approve разрешает spender тратить amount токенов от имени owner
func (t *ERC20Token) Approve(owner, spender string, amount uint64) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if t.balances[owner] < amount {
		return errors.New("insufficient balance for approval")
	}
	if t.allowances[owner] == nil {
		t.allowances[owner] = make(map[string]uint64)
	}
	t.allowances[owner][spender] = amount
	return nil
}

// Allowance возвращает, сколько spender может потратить от owner
func (t *ERC20Token) Allowance(owner, spender string) uint64 {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	if t.allowances[owner] == nil {
		return 0
	}
	return t.allowances[owner][spender]
}

// TransferFrom позволяет spender перевести amount токенов от from к to, если есть разрешение
func (t *ERC20Token) TransferFrom(spender, from, to string, amount uint64) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	allowance := t.allowances[from][spender]
	if allowance < amount {
		return errors.New("allowance exceeded")
	}
	if t.balances[from] < amount {
		return errors.New("insufficient balance")
	}
	t.allowances[from][spender] -= amount
	t.balances[from] -= amount
	t.balances[to] += amount
	return nil
}

// Пример использования:
//
// func main() {
//     token := NewERC20Token("DemoToken", "DMT", 18, 1000000, "GND1...")
//     token.Transfer("GND1...", "GND2...", 100)
//     token.Approve("GND2...", "GND3...", 50)
//     token.TransferFrom("GND3...", "GND2...", "GND4...", 30)
//     fmt.Println("GND4 balance:", token.BalanceOf("GND4..."))
// }
