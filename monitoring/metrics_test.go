// | KB @CerbeRus - Nexus Invest Team
package monitoring

import (
	"testing"
)

func TestNewMetricsRegistry_EmptyPath(t *testing.T) {
	reg, err := NewMetricsRegistry("")
	if err != nil {
		t.Fatalf("NewMetricsRegistry: %v", err)
	}
	if reg == nil {
		t.Fatal("ожидали не nil MetricsRegistry")
	}
}
