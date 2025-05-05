package tokens

import (
	"errors"
	"fmt"
	"sync"
)

// CustomTokenInterface определяет универсальный интерфейс для кастомных токенов
type CustomTokenInterface interface {
	Name() string
	Symbol() string
	Decimals() uint8
	TotalSupply() uint64
	BalanceOf(address string) uint64
	Transfer(from, to string, amount uint64) error
	CustomMethod(method string, args ...interface{}) (interface{}, error)
}

// CustomToken реализует кастомный токен с расширяемой бизнес-логикой
type CustomToken struct {
	name        string
	symbol      string
	decimals    uint8
	totalSupply uint64
	balances    map[string]uint64
	owner       string
	// Дополнительные параметры и методы
	customLogic map[string]func(*CustomToken, []interface{}) (interface{}, error)
	mutex       sync.RWMutex
}

// NewCustomToken создает новый кастомный токен
func NewCustomToken(name, symbol string, decimals uint8, initialSupply uint64, owner string) *CustomToken {
	ct := &CustomToken{
		name:        name,
		symbol:      symbol,
		decimals:    decimals,
		totalSupply: initialSupply,
		balances:    make(map[string]uint64),
		owner:       owner,
		customLogic: make(map[string]func(*CustomToken, []interface{}) (interface{}, error)),
	}
	// Владелец получает весь начальный выпуск
	ct.balances[owner] = initialSupply
	return ct
}

// Name возвращает имя токена
func (ct *CustomToken) Name() string {
	return ct.name
}

// Symbol возвращает символ токена
func (ct *CustomToken) Symbol() string {
	return ct.symbol
}

// Decimals возвращает количество знаков после запятой
func (ct *CustomToken) Decimals() uint8 {
	return ct.decimals
}

// TotalSupply возвращает общий объем выпуска
func (ct *CustomToken) TotalSupply() uint64 {
	ct.mutex.RLock()
	defer ct.mutex.RUnlock()
	return ct.totalSupply
}

// BalanceOf возвращает баланс по адресу
func (ct *CustomToken) BalanceOf(address string) uint64 {
	ct.mutex.RLock()
	defer ct.mutex.RUnlock()
	return ct.balances[address]
}

// Transfer осуществляет перевод токенов
func (ct *CustomToken) Transfer(from, to string, amount uint64) error {
	ct.mutex.Lock()
	defer ct.mutex.Unlock()
	if ct.balances[from] < amount {
		return errors.New("insufficient balance")
	}
	ct.balances[from] -= amount
	ct.balances[to] += amount
	return nil
}

// RegisterCustomMethod позволяет добавить кастомный метод в токен
func (ct *CustomToken) RegisterCustomMethod(method string, handler func(*CustomToken, []interface{}) (interface{}, error)) {
	ct.mutex.Lock()
	defer ct.mutex.Unlock()
	ct.customLogic[method] = handler
}

// CustomMethod вызывает кастомный метод по имени
func (ct *CustomToken) CustomMethod(method string, args ...interface{}) (interface{}, error) {
	ct.mutex.RLock()
	handler, ok := ct.customLogic[method]
	ct.mutex.RUnlock()
	if !ok {
		return nil, fmt.Errorf("custom method '%s' not implemented", method)
	}
	return handler(ct, args)
}

// Пример кастомного метода: заморозка адреса
func FreezeAddressMethod(ct *CustomToken, args []interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, errors.New("missing address argument")
	}
	address, ok := args[0].(string)
	if !ok {
		return nil, errors.New("invalid address type")
	}
	// Для примера просто выводим сообщение (реализация логики заморозки - на усмотрение разработчика)
	fmt.Printf("Address %s is now frozen (custom logic)\n", address)
	return true, nil
}

// Пример использования:
//
// func main() {
//     token := NewCustomToken("MyCustomToken", "MCT", 8, 1000000, "GND1...")
//     token.RegisterCustomMethod("freeze", FreezeAddressMethod)
//     token.Transfer("GND1...", "GND2...", 100)
//     result, err := token.CustomMethod("freeze", "GND2...")
//     fmt.Println("Freeze result:", result, "err:", err)
// }
