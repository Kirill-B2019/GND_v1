package consensus

import (
	"encoding/json"
	"os"
	"strings"
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

// Автоматический выбор консенсуса по адресу назначения
func SelectConsensusForTx(toAddress string) ConsensusType {
	if strings.HasPrefix(toAddress, "GNDct") {
		return ConsensusPoA
	}
	return ConsensusPoS
}
