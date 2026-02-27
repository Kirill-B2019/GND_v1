// | KB @CerbeRus - Nexus Invest Team
// tokens/registry/registry.go

package registry

import (
	"GND/tokens/interfaces"
	"GND/tokens/standards/gndst1"
	"GND/types"
	"context"
	"errors"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Tokens = map[string]*gndst1.GNDst1{} // Хранение конкретной реализации GNDst1
var mutex sync.RWMutex

// TokenRegistry управляет регистрацией токенов
type TokenRegistry struct {
	pool *pgxpool.Pool
}

// NewTokenRegistry создает новый реестр токенов
func NewTokenRegistry(pool *pgxpool.Pool) *TokenRegistry {
	return &TokenRegistry{
		pool: pool,
	}
}

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
func GetAllTokens() []*types.TokenInfo {
	mutex.RLock()
	defer mutex.RUnlock()

	var list []*types.TokenInfo
	for addr, token := range Tokens {
		list = append(list, &types.TokenInfo{
			Name:        token.GetName(),
			Symbol:      token.GetSymbol(),
			Decimals:    token.GetDecimals(),
			TotalSupply: token.GetTotalSupply().String(),
			Address:     addr,
			Standard:    token.GetStandard(),
		})
	}
	return list
}

// RegisterToken регистрирует новый токен
func (r *TokenRegistry) RegisterToken(ctx context.Context, token interfaces.TokenInterface) error {
	addr := token.GetAddress()
	if addr == "" {
		return errors.New("некорректный адрес токена")
	}

	// Сохраняем токен в реестре
	if err := RegisterToken(string(addr), token.(*gndst1.GNDst1)); err != nil {
		return err
	}

	return nil
}

// GetToken возвращает информацию о токене по адресу
func (r *TokenRegistry) GetToken(ctx context.Context, address string) (interfaces.TokenInterface, error) {
	return GetToken(address)
}

// ListTokens возвращает список всех токенов
func (r *TokenRegistry) ListTokens(ctx context.Context) ([]interfaces.TokenInterface, error) {
	tokens := GetAllTokens()
	result := make([]interfaces.TokenInterface, len(tokens))
	for i, token := range tokens {
		t, err := GetToken(token.Address)
		if err != nil {
			return nil, err
		}
		result[i] = t
	}
	return result, nil
}
