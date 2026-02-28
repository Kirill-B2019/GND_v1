// | KB @CerbeRus - Nexus Invest Team
package storage

import (
	"testing"

	"github.com/google/uuid"
)

func TestWallet_ZeroValue(t *testing.T) {
	var w Wallet
	if w.ID != uuid.Nil {
		t.Errorf("нулевой Wallet должен иметь uuid.Nil, получили %v", w.ID)
	}
	if w.AccountID != 0 {
		t.Errorf("нулевой Wallet должен иметь AccountID 0, получили %d", w.AccountID)
	}
	if w.Disabled != false {
		t.Error("нулевой Wallet должен иметь Disabled false")
	}
}
