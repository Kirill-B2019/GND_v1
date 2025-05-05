package tokens

import (
	"GND/utils"
	"errors"
	"fmt"
	"sync"
)

// Структура реестра токенов
type TokenRegistry struct {
	tokens map[string]TokenInterface // адрес → токен
	mutex  sync.RWMutex
}

// NewTokenRegistry создает новый реестр токенов
func NewTokenRegistry() *TokenRegistry {
	return &TokenRegistry{
		tokens: make(map[string]TokenInterface),
	}
}

// RegisterToken регистрирует новый токен в реестре
func (tr *TokenRegistry) RegisterToken(token TokenInterface) error {
	tr.mutex.Lock()
	defer tr.mutex.Unlock()
	addr := token.Meta().Address
	if _, exists := tr.tokens[addr]; exists {
		return errors.New("token already registered")
	}
	tr.tokens[addr] = token
	return nil
}

// GetToken возвращает токен по адресу
func (tr *TokenRegistry) GetToken(address string) (TokenInterface, error) {
	address = utils.RemovePrefix(address)
	tr.mutex.RLock()
	defer tr.mutex.RUnlock()
	token, ok := tr.tokens[address]
	if !ok {
		return nil, errors.New("token not found")
	}
	return token, nil
}

// ListTokens возвращает список всех токенов
func (tr *TokenRegistry) ListTokens() []TokenInterface {
	tr.mutex.RLock()
	defer tr.mutex.RUnlock()
	result := make([]TokenInterface, 0, len(tr.tokens))
	for _, t := range tr.tokens {
		result = append(result, t)
	}
	return result
}

// CallTokenMethod универсально вызывает метод токена по адресу и имени метода
func (tr *TokenRegistry) CallTokenMethod(address, method string, args ...interface{}) (interface{}, error) {
	token, err := tr.GetToken(address)
	if err != nil {
		return nil, err
	}
	switch token.Meta().Standard {
	case StandardERC20, StandardTRC20:
		switch method {
		case "name":
			return token.Name(), nil
		case "symbol":
			return token.Symbol(), nil
		case "decimals":
			return token.Decimals(), nil
		case "totalSupply":
			return token.TotalSupply(), nil
		case "balanceOf":
			if len(args) < 1 {
				return nil, errors.New("missing address for balanceOf")
			}
			addr, ok := args[0].(string)
			if !ok {
				return nil, errors.New("invalid address type")
			}
			return token.BalanceOf(addr), nil
		case "transfer":
			if len(args) < 3 {
				return nil, errors.New("missing arguments for transfer")
			}
			from, _ := args[0].(string)
			to, _ := args[1].(string)
			amount, _ := args[2].(uint64)
			return nil, token.Transfer(from, to, amount)
		case "approve":
			if len(args) < 3 {
				return nil, errors.New("missing arguments for approve")
			}
			owner, _ := args[0].(string)
			spender, _ := args[1].(string)
			amount, _ := args[2].(uint64)
			return nil, token.Approve(owner, spender, amount)
		case "allowance":
			if len(args) < 2 {
				return nil, errors.New("missing arguments for allowance")
			}
			owner, _ := args[0].(string)
			spender, _ := args[1].(string)
			return token.Allowance(owner, spender), nil
		case "transferFrom":
			if len(args) < 4 {
				return nil, errors.New("missing arguments for transferFrom")
			}
			spender, _ := args[0].(string)
			from, _ := args[1].(string)
			to, _ := args[2].(string)
			amount, _ := args[3].(uint64)
			return nil, token.TransferFrom(spender, from, to, amount)
		default:
			return nil, fmt.Errorf("unknown method for ERC20/TRC20: %s", method)
		}
	case StandardCustom:
		return token.CustomMethod(method, args...)
	default:
		return nil, errors.New("unsupported token standard")
	}
}
