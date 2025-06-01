// config/config.go
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
}

// Config описывает параметры блокчейна, которые могут быть заданы через файл или API
type Config struct {
	ConsensusType string       `json:"consensus_type"` // "pos" или "poa"
	GasLimit      uint64       `json:"gas_limit"`      // лимит газа на блок
	NetworkID     string       `json:"network_id"`     // идентификатор сети
	RpcPort       int          `json:"rpc_port"`       // порт для RPC API
	RestPort      int          `json:"rest_port"`      // порт для REST API
	WsPort        int          `json:"ws_port"`        // порт для WebSocket
	Coin          CoinConfig   `json:"coin"`
	EVM           EVMConfig    `json:"evm"`
	Server        ServerConfig `json:"server"`
	// Можно добавить другие параметры: комиссия, минимальный стейк, список авторитетов и т.д.
}

// GlobalConfig - потокобезопасная обертка для конфигурации (если нужно динамическое обновление)
type GlobalConfig struct {
	mutex  sync.RWMutex
	config *Config
}

// NewConfigFromFile загружает конфиг из файла
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

// Пример использования потокобезопасной обертки
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
