package tokens

import (
	"errors"
	"sync"
)

// TRC20TokenInterface определяет стандартные методы TRC-20
type TRC20TokenInterface interface {
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

// TRC20Token реализует стандарт TRC-20 (аналогично ERC-20)
type TRC20Token struct {
	name        string
	symbol      string
	decimals    uint8
	totalSupply uint64
	balances    map[string]uint64
	allowances  map[string]map[string]uint64 // owner → spender → amount
	mutex       sync.RWMutex
}

// NewTRC20Token создает новый TRC-20 токен
func NewTRC20Token(name, symbol string, decimals uint8, initialSupply uint64, owner string) *TRC20Token {
	token := &TRC20Token{
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
func (t *TRC20Token) Name() string {
	return t.name
}

// Symbol возвращает символ токена
func (t *TRC20Token) Symbol() string {
	return t.symbol
}

// Decimals возвращает количество знаков после запятой
func (t *TRC20Token) Decimals() uint8 {
	return t.decimals
}

// TotalSupply возвращает общий объем выпуска
func (t *TRC20Token) TotalSupply() uint64 {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.totalSupply
}

// BalanceOf возвращает баланс по адресу
func (t *TRC20Token) BalanceOf(address string) uint64 {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.balances[address]
}

// Transfer переводит amount токенов от from к to
func (t *TRC20Token) Transfer(from, to string, amount uint64) error {
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
func (t *TRC20Token) Approve(owner, spender string, amount uint64) error {
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
func (t *TRC20Token) Allowance(owner, spender string) uint64 {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	if t.allowances[owner] == nil {
		return 0
	}
	return t.allowances[owner][spender]
}

// TransferFrom позволяет spender перевести amount токенов от from к to, если есть разрешение
func (t *TRC20Token) TransferFrom(spender, from, to string, amount uint64) error {
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
//     token := NewTRC20Token("DemoTRC20", "DTRC", 6, 1000000, "GNDT1...")
//     token.Transfer("GNDT1...", "GNDT2...", 100)
//     token.Approve("GNDT2...", "GNDT3...", 50)
//     token.TransferFrom("GNDT3...", "GNDT2...", "GNDT4...", 30)
//     fmt.Println("GNDT4 balance:", token.BalanceOf("GNDT4..."))
// }
