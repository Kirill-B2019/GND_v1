// | KB @CerbeRus - Nexus Invest Team
package consensus

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSelectConsensusForTx(t *testing.T) {
	tests := []struct {
		addr   string
		expect ConsensusType
	}{
		{"GNDct123", ConsensusPoA},
		{"GNDct", ConsensusPoA},
		{"GND123", ConsensusPoS},
		{"", ConsensusPoS},
		{"GND", ConsensusPoS},
	}
	for _, tt := range tests {
		got := SelectConsensusForTx(tt.addr)
		if got != tt.expect {
			t.Errorf("SelectConsensusForTx(%q) = %v, ожидали %v", tt.addr, got, tt.expect)
		}
	}
}

func TestLoadConsensusSettings_NoFile(t *testing.T) {
	_, err := LoadConsensusSettings(filepath.Join(os.TempDir(), "nonexistent_consensus_12345.json"))
	if err == nil {
		t.Error("ожидали ошибку при отсутствии файла")
	}
}
