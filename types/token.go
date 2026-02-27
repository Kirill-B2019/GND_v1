// | KB @CerbeRus - Nexus Invest Team
package types

import (
	"time"
)

// TokenInfo содержит базовую информацию о токене
type TokenInfo struct {
	Address     string
	Name        string
	Symbol      string
	Decimals    uint8
	TotalSupply string
	Standard    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// TokenMetadata содержит метаданные токена
type TokenMetadata struct {
	Name        string
	Symbol      string
	Description string
	Standard    string
	Owner       string
	Params      map[string]string
}
