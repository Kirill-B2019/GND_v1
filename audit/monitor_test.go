package audit

import (
	"math/big"
	"testing"
)

// TestNewMonitor проверяет создание монитора (использует NewMonitor для линтера).
func TestNewMonitor(t *testing.T) {
	m := NewMonitor(big.NewInt(1_000_000))
	if m == nil {
		t.Fatal("NewMonitor не должен возвращать nil")
	}
	if m.Threshold.Cmp(big.NewInt(1_000_000)) != 0 {
		t.Errorf("ожидался порог 1000000, получено %s", m.Threshold.String())
	}
	if len(m.GetSuspicious()) != 0 {
		t.Error("новый монитор должен иметь пустой список подозрительных")
	}
}
