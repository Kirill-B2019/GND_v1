package deployer

import (
	"testing"
)

// TestNewDeployer проверяет создание Deployer (использует NewDeployer, убирает предупреждение линтера).
func TestNewDeployer(t *testing.T) {
	d := NewDeployer(nil, nil, nil)
	if d == nil {
		t.Fatal("NewDeployer не должен возвращать nil")
	}
}
