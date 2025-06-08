// tokens/registry/registry.go

package registry

import (
	"GND/tokens"
	"errors"
	"sync"

	"GND/tokens/interfaces"
	"GND/tokens/standards/gndst1"
)

var Tokens = map[string]*gndst1.GNDst1{} // Хранение конкретной реализации GNDst1
var mutex sync.RWMutex

// RegisterToken регистрирует новый токен в реестре
func RegisterToken(addr string, token *gndst1.GNDst1) error {
	mutex.Lock()
	defer mutex.Unlock()

	if _, exists := Tokens[addr]; exists {
		return errors.New("токен уже зарегистрирован")
	}

	Tokens[addr] = token
	return nil
}

// GetToken возвращает токен по адресу
func GetToken(addr string) (interfaces.TokenInterface, error) {
	mutex.RLock()
	defer mutex.RUnlock()

	token, ok := Tokens[addr]
	if !ok {
		return nil, errors.New("токен не найден")
	}
	return token, nil
}

// GetAllTokens возвращает список информации обо всех зарегистрированных токенах
func GetAllTokens() []*tokens.TokenInfo {
	mutex.RLock()
	defer mutex.RUnlock()

	var list []*tokens.TokenInfo
	for addr, token := range Tokens {
		list = append(list, &tokens.TokenInfo{
			Name:        token.Name(),
			Symbol:      token.Symbol(),
			Decimals:    token.Decimals(),
			TotalSupply: token.TotalSupply(),
			Address:     addr,
		})
	}
	return list
}
