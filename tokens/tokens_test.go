// | KB @CerbeRus - Nexus Invest Team
package tokens

import (
	"math/big"
	"testing"
)

func TestTokenInfo_Validate_EmptyName(t *testing.T) {
	info := &TokenInfo{Name: "", Symbol: "TKN", Decimals: 18, TotalSupply: big.NewInt(1000)}
	if err := info.Validate(); err == nil {
		t.Error("ожидали ошибку при пустом имени")
	}
}

func TestTokenInfo_Validate_Valid(t *testing.T) {
	info := &TokenInfo{
		Name:        "Test",
		Symbol:      "TKN",
		Decimals:    18,
		TotalSupply: big.NewInt(1000),
		Address:     "GNDct12345678901234567890123456789012",
	}
	if err := info.Validate(); err != nil {
		t.Errorf("валидный TokenInfo не должен возвращать ошибку: %v", err)
	}
}
