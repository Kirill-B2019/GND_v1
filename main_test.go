package main

import (
	"GND/core"
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
		Timestamp:    time.Now().Unix(),
		PrevHash:     "",
		Hash:         "genesis",
		Transactions: []*core.Transaction{},
	}
	blockchain := core.NewBlockchain(genesisBlock, pool)

	// 4. Начисляем баланс
	addr := core.Address("GND1234567890")
	balance := big.NewInt(1000000)
	blockchain.State.Credit(addr, "GND.c", balance)

	// 5. Сохраняем блокчейн в БД
	err = blockchain.SaveToDB()
	if err != nil {
		t.Errorf("SaveToDB failed: %v", err)
	}

	// 6. Восстанавливаем блокчейн из БД
	newChain, loadErr := core.LoadBlockchainFromDB(pool)
	if loadErr != nil {
		t.Fatalf("LoadBlockchainFromDB failed: %v", loadErr)
	}

	// 7. Проверяем сохранённый баланс
	storedBalance := newChain.State.GetBalance(addr, "GND.c")
	if storedBalance.Cmp(balance) != 0 {
		t.Errorf("Expected balance %d, got %d", balance, storedBalance)
	}
}
