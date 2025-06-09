//core/config.go

package core

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// Основной конфиг блокчейна (config.json)
type Config struct {
	Port          int                      `json:"port"`
	NodeName      string                   `json:"node_name"`
	ConsensusType string                   `json:"consensus_type"`
	GasLimit      uint64                   `json:"gas_limit"`
	NetworkID     string                   `json:"network_id"`
	MaxWorkers    int                      `json:"MaxWorkers"`
	Coins         []CoinConfig             `json:"coins"`
	Consensus     []map[string]interface{} `json:"consensus"`
	EVM           EVMConfig                `json:"evm"`
	Server        ServerConfig             `json:"server"`
	DB            DBConfig                 `json:"database"`
}

type GlobalConfig struct {
	mutex  sync.RWMutex
	config *Config
}
type DBConfig struct {
	Host            string        `json:"host"`
	Port            int           `json:"port"`
	User            string        `json:"user"`
	Password        string        `json:"password"`
	DBName          string        `json:"dbname"`
	SSLMode         string        `json:"sslmode"`
	MaxConns        int           `json:"max_conns"`
	MinConns        int           `json:"min_conns"`
	MaxConnLifetime time.Duration `json:"max_conn_lifetime"`
	MaxConnIdleTime time.Duration `json:"max_conn_idle_time"`
}

type CoinConfig struct {
	Name            string `json:"name"`
	Symbol          string `json:"symbol"`
	Decimals        int    `json:"decimals"`
	Description     string `json:"description"`
	ContractAddress string `json:"contract_address"`
	CoinLogo        string `json:"coin_logo"`
	TotalSupply     string `json:"total_supply"`
	Standard        string `json:"standard"`
}

type ConsensusPosConfig struct {
	Type              string `json:"type"`
	AverageBlockDelay string `json:"average_block_delay"`
	InitialBaseTarget int    `json:"initial_base_target"`
	InitialBalance    string `json:"initial_balance"`
}

type ConsensusPoaConfig struct {
	Type              string `json:"type"`
	RoundDuration     string `json:"round_duration"`
	SyncDuration      string `json:"sync_duration"`
	BanDurationBlocks int    `json:"ban_duration_blocks"`
	WarningsForBan    int    `json:"warnings_for_ban"`
	MaxBansPercentage int    `json:"max_bans_percentage"`
}

type ConsensusConfig struct {
	Consensus []map[string]interface{} `json:"consensus"`
}

type EVMConfig struct {
	GasLimit uint64 `json:"gas_limit"`
}

type ServerRPCConfig struct {
	RPCAddr string `json:"rpc_addr"`
	Name    string `json:"name"`
}

type ServerRESTConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

type ServerWSConfig struct {
	WSAddr string `json:"ws_addr"`
	Name   string `json:"name"`
}

type ServerConfig struct {
	RPC  ServerRPCConfig  `json:"rpc"`
	REST ServerRESTConfig `json:"rest"`
	WS   ServerWSConfig   `json:"ws"`
}

// Загрузка основного конфига и конфига базы данных, слияние в одну структуру
func InitGlobalConfigDefault() (*GlobalConfig, error) {
	const (
		mainPath      = "config/config.json"
		dbPath        = "config/db.json"
		coinsPath     = "config/coins.json"
		consensusPath = "config/consensus.json"
		evmPath       = "config/evm.json"
		serversPath   = "config/servers.json"
	)

	var cfg Config

	// Основной конфиг
	if data, err := os.ReadFile(mainPath); err == nil {
		_ = json.Unmarshal(data, &cfg)
	}

	// DB конфиг
	type dbWrapper struct {
		DB DBConfig `json:"database"`
	}
	if data, err := os.ReadFile(dbPath); err == nil {
		var dbCfg dbWrapper
		if err := json.Unmarshal(data, &dbCfg); err == nil {
			cfg.DB = dbCfg.DB
		}
	}

	// Coins
	type coinsWrapper struct {
		Coins []CoinConfig `json:"coins"`
	}
	if data, err := os.ReadFile(coinsPath); err == nil {
		var coinsCfg coinsWrapper
		if err := json.Unmarshal(data, &coinsCfg); err == nil {
			cfg.Coins = coinsCfg.Coins
		}
	}

	// Consensus
	type consensusWrapper struct {
		Consensus []map[string]interface{} `json:"consensus"`
	}
	if data, err := os.ReadFile(consensusPath); err == nil {
		var consensusCfg consensusWrapper
		if err := json.Unmarshal(data, &consensusCfg); err == nil {
			cfg.Consensus = consensusCfg.Consensus
		}
	}

	// EVM
	type evmWrapper struct {
		EVM EVMConfig `json:"evm"`
	}
	if data, err := os.ReadFile(evmPath); err == nil {
		var evmCfg evmWrapper
		if err := json.Unmarshal(data, &evmCfg); err == nil {
			cfg.EVM = evmCfg.EVM
		}
	}

	// Servers
	type serversWrapper struct {
		Server ServerConfig `json:"server"`
	}
	if data, err := os.ReadFile(serversPath); err == nil {
		var serversCfg serversWrapper
		if err := json.Unmarshal(data, &serversCfg); err == nil {
			cfg.Server = serversCfg.Server
		}
	}

	gc := &GlobalConfig{}
	gc.Set(&cfg)
	return gc, nil
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
