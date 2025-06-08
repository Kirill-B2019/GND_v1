// tokens/registry/registry_test.go

package registry

import (
	"math/big"
	"testing"

	"GND/tokens/standards/gndst1"
)

func TestRegisterToken(t *testing.T) {
	token := gndst1.NewGNDst1(big.NewInt(1e18), "bridge1")
	err := RegisterToken("GNDct1_token1", token)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Token registered: %v", token)
}
