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

func TestSelectConsensusForTx_WithRules(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "consensus.json")
	emptyPath := filepath.Join(dir, "empty.json")
	if err := os.WriteFile(emptyPath, []byte(`{"selection_rules":[]}`), 0600); err != nil {
		t.Fatal(err)
	}
	// Правила: префикс GNDct → poa, по умолчанию pos
	if err := os.WriteFile(path, []byte(`{"selection_rules":[{"address_prefix":"GNDct","consensus":"poa"},{"default":true,"consensus":"pos"}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	LoadSelectionRules(path)
	defer LoadSelectionRules(emptyPath)

	if got := SelectConsensusForTx("GNDct123"); got != ConsensusPoA {
		t.Errorf("GNDct123: got %v, want poa", got)
	}
	if got := SelectConsensusForTx("GND123"); got != ConsensusPoS {
		t.Errorf("GND123: got %v, want pos", got)
	}
}
