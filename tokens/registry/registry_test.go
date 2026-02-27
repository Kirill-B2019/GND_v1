// | KB @CerbeRus - Nexus Invest Team
// tokens/registry/registry_test.go

package registry

import (
	"math/big"
	"testing"

	"GND/tokens/standards/gndst1"
)

func TestRegisterToken(t *testing.T) {
	token := gndst1.NewGNDst1("GNDct1_token1", "Test Token", "T", 18, big.NewInt(1e18), nil)
	err := RegisterToken("GNDct1_token1", token)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Token registered: %v", token)
}
