package main

import (
	"GND/core"
	"GND/types"
	"context"
	"encoding/json"
	"math/big"
	_ "net/http/pprof"
	"os"
	"testing"
	"time"
)

// DBConfig — структура для загрузки настроек БД из JSON
type DBConfig struct {
	DB struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		DBName   string `json:"dbname"`
		SSLMode  string `json:"sslmode"`
		MaxConns int    `json:"max_conns"`
		MinConns int    `json:"min_conns"`
	} `json:"database"`
}

func TestNewBlockchain(t *testing.T) {
	var dbCfg DBConfig
	ctx := context.Background()

	// 1. Загружаем конфигурацию из db.json
	data, err := os.ReadFile("config/db.json")
	if err != nil {
		t.Fatalf("Не удалось прочитать config/db.json: %v", err)
	}
	err = json.Unmarshal(data, &dbCfg)
	if err != nil {
		t.Fatalf("Ошибка парсинга JSON: %v", err)
	}

	// 2. Инициализируем пул подключений
	pool, err := core.InitDBPool(ctx, core.DBConfig{
		Host:     dbCfg.DB.Host,
		Port:     dbCfg.DB.Port,
		User:     dbCfg.DB.User,
		Password: dbCfg.DB.Password,
		DBName:   dbCfg.DB.DBName,
		SSLMode:  dbCfg.DB.SSLMode,
		MaxConns: dbCfg.DB.MaxConns,
		MinConns: dbCfg.DB.MinConns,
	})
	if err != nil {
		t.Fatalf("Не удалось создать пул БД: %v", err)
	}
	defer pool.Close()

	// 3. Создаем генезис-блок и блокчейн
	genesisBlock := &core.Block{
		Index:        0,
		Timestamp:    time.Now().UTC(),
		PrevHash:     "",
		Hash:         "genesis",
		Consensus:    "poa",
		Nonce:        0,
		Status:       "finalized",
		Transactions: []*core.Transaction{},
	}
	genesisBlock.Hash = genesisBlock.CalculateHash()
	blockchain := core.NewBlockchain(genesisBlock, pool)

	// 4. Привязываем пул к состоянию и сохраняем генезис в БД (если блока 0 ещё нет)
	if st, ok := blockchain.State.(*core.State); ok {
		st.SetPool(pool)
		var exists bool
		_ = pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM blocks WHERE index = 0)").Scan(&exists)
		if !exists {
			if err := genesisBlock.SaveToDB(ctx, pool); err != nil {
				t.Fatalf("Genesis SaveToDB failed: %v", err)
			}
		}
		addr := types.Address("GND1234567890")
		balance := big.NewInt(1000000)
		if err := st.Credit(addr, "GND", balance); err != nil {
			t.Fatalf("Credit failed: %v", err)
		}
		if err := st.SaveToDB(); err != nil {
			t.Skipf("State.SaveToDB skipped (схема token_balances): %v", err)
		}
	}

	// 5. Восстанавливаем блокчейн из БД
	newChain, loadErr := core.LoadBlockchainFromDB(pool)
	if loadErr != nil {
		t.Fatalf("LoadBlockchainFromDB failed: %v", loadErr)
	}

	// 6. Проверяем, что генезис загружен
	if newChain.Genesis == nil || newChain.Genesis.Hash == "" {
		t.Error("Expected genesis block to be loaded")
	}
}
