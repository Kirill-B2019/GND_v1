//core/config.go

package core

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type EVMConfig struct {
	GasLimit uint64 `json:"gas_limit"`
}

type ServerConfig struct {
	RPCAddr       string `json:"rpc_addr"`
	RESTAddr      string `json:"rest_addr"`
	WebSocketAddr string `json:"ws_addr"`
}

type CoinConfig struct {
	Name            string `json:"name"`
	Symbol          string `json:"symbol"`
	Decimals        int    `json:"decimals"`
	Description     string `json:"description"`
	ContractAddress string `json:"contract_address"`
	CoinLogo        string `json:"coin_logo"`
	TotalSupply     string `json:"total_supply"` // Добавлено новое поле
}

// Основной конфиг блокчейна (config.json)
type Config struct {
	Port          int          `json:"port"`
	NodeName      string       `json:"node_name"`
	ConsensusType string       `json:"consensus_type"`
	GasLimit      uint64       `json:"gas_limit"`
	NetworkID     string       `json:"network_id"`
	Coins         []CoinConfig `json:"coins"` // Массив вместо единичного Coin
	EVM           EVMConfig    `json:"evm"`
	Server        ServerConfig `json:"server"`
	ConsensusConf string       `json:"consensus_config"` // путь к consensus.json
	MaxWorkers    int          `json:"max_workers"`
}

type GlobalConfig struct {
	mutex  sync.RWMutex
	config *Config
}

func NewConfigFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения файла конфига: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("ошибка парсинга конфига: %w", err)
	}
	return &cfg, nil
}

func (g *GlobalConfig) Get() *Config {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.config
}

func (g *GlobalConfig) Set(cfg *Config) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.config = cfg
}
