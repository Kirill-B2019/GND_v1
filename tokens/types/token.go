// | KB @CerbeRus - Nexus Invest Team
package types

import (
	"math/big"
)

// TokenParams содержит параметры для создания токена
type TokenParams struct {
	Name        string
	Symbol      string
	Decimals    uint8
	TotalSupply *big.Int
	Owner       string
	Standard    string
	LogoURL     string // Ссылка на логотип (URL или путь после загрузки)
}

// TokenInfo содержит информацию о токене
type TokenInfo struct {
	Address     string
	Owner       string
	Name        string
	Symbol      string
	Decimals    uint8
	TotalSupply *big.Int
	Standard    string
	CreatedAt   int64
	LogoURL     string // Ссылка на логотип (URL или путь)
}
