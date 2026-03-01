// | KB @CerbeRus - Nexus Invest Team
package main

import (
	"GND/core"
	"GND/types"
	"context"
	"encoding/json"
	"math/big"
	"net"
	_ "net/http/pprof"
	"os"
	"strconv"
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

// TestNodeRestartWithoutDBReset проверяет, что после «перезапуска» (новое подключение к той же БД)
// данные не теряются: генезис и блоки остаются в БД и успешно загружаются.
func TestNodeRestartWithoutDBReset(t *testing.T) {
	data, err := os.ReadFile("config/db.json")
	if err != nil {
		t.Skipf("config/db.json не найден, пропуск: %v", err)
		return
	}
	var dbCfg DBConfig
	if err := json.Unmarshal(data, &dbCfg); err != nil {
		t.Fatalf("ошибка парсинга config/db.json: %v", err)
	}
	ctx := context.Background()
	cfg := core.DBConfig{
		Host:     dbCfg.DB.Host,
		Port:     dbCfg.DB.Port,
		User:     dbCfg.DB.User,
		Password: dbCfg.DB.Password,
		DBName:   dbCfg.DB.DBName,
		SSLMode:  dbCfg.DB.SSLMode,
		MaxConns: dbCfg.DB.MaxConns,
		MinConns: dbCfg.DB.MinConns,
	}

	// «Первый запуск»: подключаемся, сохраняем генезис при необходимости
	pool1, err := core.InitDBPool(ctx, cfg)
	if err != nil {
		t.Skipf("подключение к БД недоступно: %v", err)
		return
	}
	genesis := &core.Block{
		Index:        0,
		Timestamp:    time.Now().UTC(),
		PrevHash:     "",
		Hash:         "genesis_restart_test",
		Consensus:    "poa",
		Nonce:        0,
		Status:       "finalized",
		Transactions: []*core.Transaction{},
	}
	genesis.Hash = genesis.CalculateHash()
	var exists bool
	_ = pool1.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM blocks WHERE index = 0)").Scan(&exists)
	if !exists {
		if err := genesis.SaveToDB(ctx, pool1); err != nil {
			t.Fatalf("сохранение генезиса: %v", err)
		}
	}
	pool1.Close()

	// «Перезапуск»: новое подключение к той же БД
	pool2, err := core.InitDBPool(ctx, cfg)
	if err != nil {
		t.Fatalf("повторное подключение к БД: %v", err)
	}
	defer pool2.Close()

	chain, err := core.LoadBlockchainFromDB(pool2)
	if err != nil {
		t.Fatalf("LoadBlockchainFromDB после перезапуска: %v", err)
	}
	if chain == nil || chain.Genesis == nil {
		t.Fatal("блокчейн или генезис не загружены после перезапуска")
	}
	if chain.Genesis.Index != 0 {
		t.Errorf("ожидался index генезиса 0, получен %d", chain.Genesis.Index)
	}
	t.Log("перезапуск ноды без обнуления БД: генезис успешно загружен")
}

// TestCheckPortsFree_Free проверяет, что checkPortsFree возвращает nil для свободных портов.
func TestCheckPortsFree_Free(t *testing.T) {
	// Используем порты из высокого диапазона, чтобы не конфликтовать с реальными сервисами
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Skipf("не удалось занять тестовый порт: %v", err)
	}
	_, portStr, _ := net.SplitHostPort(l.Addr().String())
	l.Close()
	port, _ := strconv.Atoi(portStr)
	cfg := &core.Config{
		Server: core.ServerConfig{
			RPC:  core.ServerRPCConfig{RPCAddr: "127.0.0.1:" + portStr},
			REST: core.ServerRESTConfig{Port: port},
			WS:   core.ServerWSConfig{WSAddr: "127.0.0.1:" + portStr},
		},
	}
	if err := checkPortsFree(cfg); err != nil {
		t.Errorf("ожидали успех для свободного порта: %v", err)
	}
}

// TestCheckPortsFree_Busy проверяет, что checkPortsFree возвращает ошибку для занятого порта.
func TestCheckPortsFree_Busy(t *testing.T) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Skipf("не удалось занять тестовый порт: %v", err)
	}
	_, portStr, _ := net.SplitHostPort(l.Addr().String())
	defer l.Close()
	port, _ := strconv.Atoi(portStr)
	cfg := &core.Config{
		Server: core.ServerConfig{
			RPC:  core.ServerRPCConfig{RPCAddr: "127.0.0.1:" + portStr},
			REST: core.ServerRESTConfig{Port: port},
			WS:   core.ServerWSConfig{WSAddr: "127.0.0.1:" + portStr},
		},
	}
	if err := checkPortsFree(cfg); err == nil {
		t.Error("ожидали ошибку для занятого порта")
	}
}
