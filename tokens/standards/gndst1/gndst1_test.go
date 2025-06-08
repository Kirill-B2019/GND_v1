// tokens/standards/gndst1/gndst1_test.go

package gndst1

import (
	"math/big"
	"testing"
)

func TestNewGNDst1(t *testing.T) {
	token := NewGNDst1(big.NewInt(1e18), "bridge1")
	if token.BalanceOf("owner").Cmp(big.NewInt(1e18)) != 0 {
		t.FailNow()
	}
}
