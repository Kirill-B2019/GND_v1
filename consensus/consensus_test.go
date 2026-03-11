// | KB @CerberRus00 - Nexus Invest Team
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

func TestSelectConsensus_ByTxType(t *testing.T) {
	// Без правил в конфиге — встроенная логика по типу и адресу
	dir := t.TempDir()
	emptyPath := filepath.Join(dir, "empty.json")
	if err := os.WriteFile(emptyPath, []byte(`{"selection_rules":[]}`), 0600); err != nil {
		t.Fatal(err)
	}
	LoadSelectionRules(emptyPath)

	tests := []struct {
		txType string
		addr   string
		expect ConsensusType
	}{
		{"contract_call", "GNDct123", ConsensusPoA},
		{"contract_call", "GNDabc", ConsensusPoA},
		{"deploy", "GNDxyz", ConsensusPoA},
		{"transfer", "GNDct123", ConsensusPoS},
		{"transfer", "GNDwallet", ConsensusPoS},
		{"", "GNDct123", ConsensusPoA},
		{"", "GNDwallet", ConsensusPoS},
		{"stake", "GNDwallet", ConsensusPoS},
	}
	for _, tt := range tests {
		got := SelectConsensus(tt.txType, tt.addr)
		if got != tt.expect {
			t.Errorf("SelectConsensus(%q, %q) = %v, ожидали %v", tt.txType, tt.addr, got, tt.expect)
		}
	}
}

func TestSelectConsensus_WithTxTypeRules(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "consensus.json")
	// Правила: contract_call/deploy → poa, transfer → pos, по адресу GNDct → poa, default → pos
	cfg := `{"selection_rules":[
		{"tx_type": "contract_call", "consensus": "poa"},
		{"tx_type": "deploy", "consensus": "poa"},
		{"tx_type": "transfer", "consensus": "pos"},
		{"address_prefix": "GNDct", "consensus": "poa"},
		{"default": true, "consensus": "pos"}
	]}`
	if err := os.WriteFile(path, []byte(cfg), 0600); err != nil {
		t.Fatal(err)
	}
	LoadSelectionRules(path)

	if got := SelectConsensus("contract_call", "GNDwallet"); got != ConsensusPoA {
		t.Errorf("contract_call → poa: got %v", got)
	}
	if got := SelectConsensus("deploy", "GNDwallet"); got != ConsensusPoA {
		t.Errorf("deploy → poa: got %v", got)
	}
	if got := SelectConsensus("transfer", "GNDct123"); got != ConsensusPoS {
		t.Errorf("transfer даже на GNDct → pos по типу: got %v", got)
	}
	if got := SelectConsensus("", "GNDct123"); got != ConsensusPoA {
		t.Errorf("пустой тип, GNDct → poa по адресу: got %v", got)
	}
	if got := SelectConsensus("unknown", "GNDxyz"); got != ConsensusPoS {
		t.Errorf("unknown type → default pos: got %v", got)
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
