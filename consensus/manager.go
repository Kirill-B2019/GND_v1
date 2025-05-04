package consensus

import (
	"encoding/json"
	"os"
)

type ConsensusType string

type Config struct {
	Type ConsensusType `json:"type"`
}

func LoadConsensusConfig(path string) (ConsensusType, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	var cfg Config
	if err := json.Unmarshal(file, &cfg); err != nil {
		return "", err
	}
	return cfg.Type, nil
}
