// | KB @CerberRus00 - Nexus Invest Team
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
	Owner       string // логический владелец токена (может быть пустым для системных токенов)
	Standard    string
	LogoURL     string // Ссылка на логотип (URL или путь после загрузки)
	// Deployer задаёт адрес кошелька, который выполняет деплой и оплачивает газ.
	// Если пустой — используется Owner (а при его отсутствии нода создаёт новый кошелёк владельца).
	Deployer string
	// SkipDeployFee — не взимать комиссию за деплой (например, при owner = gndself_address).
	SkipDeployFee bool
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
