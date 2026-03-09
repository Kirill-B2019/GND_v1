// | KB @CerbeRus - Nexus Invest Team
package consensus

import (
	"encoding/json"
	"os"
	"strings"
	"sync"

	"GND/types"
)

// Типы консенсуса
type ConsensusType string

const (
	ConsensusPoS ConsensusType = "pos"
	ConsensusPoA ConsensusType = "poa"
)

// Структуры для consensus.json
type PoSConfig struct {
	Type              string `json:"type"`
	AverageBlockDelay string `json:"average_block_delay"`
	InitialBaseTarget int64  `json:"initial_base_target"`
	InitialBalance    string `json:"initial_balance"`
}

type PoAConfig struct {
	Type              string `json:"type"`
	RoundDuration     string `json:"round_duration"`
	SyncDuration      string `json:"sync_duration"`
	BanDurationBlocks int    `json:"ban_duration_blocks"`
	WarningsForBan    int    `json:"warnings_for_ban"`
	MaxBansPercentage int    `json:"max_bans_percentage"`
}

type ConsensusSettings struct {
	PoS PoSConfig `json:"pos"`
	PoA PoAConfig `json:"poa"`
}

// SelectionRule — правило выбора консенсуса: по типу транзакции и/или по адресу получателя.
// PoA — действия с контрактами (contract_call, deploy), PoS — переводы и т.д.
type SelectionRule struct {
	TxType        string `json:"tx_type"`        // тип транзакции (например "contract_call", "deploy", "transfer")
	AddressPrefix string `json:"address_prefix"` // префикс адреса получателя (например "GNDct")
	Default       bool   `json:"default"`        // если true — правило по умолчанию
	Consensus     string `json:"consensus"`      // "poa" или "pos"
}

type consensusFile struct {
	Consensus      []map[string]interface{} `json:"consensus"`
	SelectionRules []SelectionRule          `json:"selection_rules"`
}

var (
	selectionRules   []SelectionRule
	selectionRulesMu sync.RWMutex
)

// LoadSelectionRules загружает правила выбора консенсуса из consensus.json (поле selection_rules).
// Вызывать при старте ноды (например из main). Если файл отсутствует или правил нет — используется встроенная логика.
func LoadSelectionRules(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var cf consensusFile
	if err := json.Unmarshal(data, &cf); err != nil {
		return
	}
	selectionRulesMu.Lock()
	if len(cf.SelectionRules) == 0 {
		selectionRules = nil
	} else {
		selectionRules = cf.SelectionRules
	}
	selectionRulesMu.Unlock()
}

// Загрузка consensus.json
func LoadConsensusSettings(path string) (*ConsensusSettings, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cs ConsensusSettings
	if err := json.Unmarshal(data, &cs); err != nil {
		return nil, err
	}
	return &cs, nil
}

// SelectConsensus выбирает консенсус по типу транзакции и адресу получателя.
// Порядок: правила из конфига (сначала по tx_type, затем по address_prefix), иначе встроенная логика.
// PoA — для действий с контрактами (contract_call, deploy), PoS — для переводов и прочего.
func SelectConsensus(txType, toAddress string) ConsensusType {
	selectionRulesMu.RLock()
	rules := selectionRules
	selectionRulesMu.RUnlock()

	if len(rules) > 0 {
		var defaultConsensus ConsensusType
		for _, r := range rules {
			if r.Default {
				defaultConsensus = ConsensusType(strings.ToLower(r.Consensus))
				continue
			}
			if r.TxType != "" && strings.EqualFold(strings.TrimSpace(txType), r.TxType) {
				return ConsensusType(strings.ToLower(r.Consensus))
			}
			if r.AddressPrefix != "" && strings.HasPrefix(toAddress, r.AddressPrefix) {
				return ConsensusType(strings.ToLower(r.Consensus))
			}
		}
		if defaultConsensus != "" {
			return defaultConsensus
		}
	}

	// Встроенная логика: по типу — contract_call/deploy → PoA; transfer → PoS; иначе по адресу — GNDct → PoA
	t := strings.TrimSpace(txType)
	if t == "contract_call" || t == "deploy" {
		return ConsensusPoA
	}
	if t == "transfer" {
		return ConsensusPoS
	}
	if strings.HasPrefix(toAddress, types.ContractAddressPrefix) {
		return ConsensusPoA
	}
	return ConsensusPoS
}

// SelectConsensusForTx выбирает консенсус только по адресу (обратная совместимость).
func SelectConsensusForTx(toAddress string) ConsensusType {
	return SelectConsensus("", toAddress)
}
