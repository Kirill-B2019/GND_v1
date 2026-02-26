package audit

import (
	"GND/core"
	"GND/types"
	"GND/vm"
	"context"
	"encoding/json"
	"math/big"
	"os"
	"testing"
	"time"
)

// dbConfig для загрузки config/db.json
type dbConfig struct {
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

// Интеграционный тест: создание блокчейна, кошелька, деплой контракта, проверка баланса.
// Требует доступный PostgreSQL (config/db.json). При недоступности БД тест пропускается.
func TestBlockchainIntegration(t *testing.T) {
	data, err := os.ReadFile("config/db.json")
	if err != nil {
		t.Skipf("config/db.json не найден, пропуск интеграционного теста: %v", err)
		return
	}
	var cfg dbConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Skipf("ошибка парсинга config/db.json: %v", err)
		return
	}
	pool, err := core.InitDBPool(context.Background(), core.DBConfig{
		Host:     cfg.DB.Host,
		Port:     cfg.DB.Port,
		User:     cfg.DB.User,
		Password: cfg.DB.Password,
		DBName:   cfg.DB.DBName,
		SSLMode:  cfg.DB.SSLMode,
		MaxConns: cfg.DB.MaxConns,
		MinConns: cfg.DB.MinConns,
	})
	if err != nil {
		t.Skipf("подключение к БД недоступно, пропуск: %v", err)
		return
	}
	defer pool.Close()

	// 1. Генезис и блокчейн
	genesis := &core.Block{
		Index:        0,
		Timestamp:    time.Now().UTC(),
		PrevHash:     "",
		Hash:         "genesis",
		Consensus:    "poa",
		Nonce:        0,
		Status:       "finalized",
		Transactions: []*core.Transaction{},
	}
	genesis.Hash = genesis.CalculateHash()
	blockchain := core.NewBlockchain(genesis, pool)

	// 2. Кошелёк (требует БД)
	wallet, err := core.NewWallet(pool)
	if err != nil {
		t.Fatalf("ошибка создания кошелька: %v", err)
	}

	// 3. Состояние и баланс (in-memory)
	st, ok := blockchain.State.(*core.State)
	if !ok {
		t.Fatal("State не *core.State")
	}
	st.SetPool(pool)
	addr := types.Address(wallet.Address)
	if err := st.Credit(addr, "GND", big.NewInt(100000)); err != nil {
		t.Fatalf("Credit: %v", err)
	}

	// 4. Проверка баланса
	balance := blockchain.State.GetBalance(types.Address(wallet.Address), "GND")
	if balance.Cmp(big.NewInt(100000)) != 0 {
		t.Errorf("ожидался баланс 100000, получено %s", balance.String())
	}

	// 5. EVM и деплой контракта (нужен totalSupply != nil и Coins)
	evm := vm.NewEVM(vm.EVMConfig{
		Blockchain: blockchain,
		State:      blockchain.State,
		GasLimit:   1000000,
		Coins:      []vm.CoinConfig{{Symbol: "GND", ContractAddress: "", Decimals: 18}},
	})
	contractAddr, err := evm.DeployContract(
		types.Address(wallet.Address),
		[]byte{0x60, 0x60, 0x60, 0x40},
		types.ContractMeta{Name: "TestContract"},
		21000, big.NewInt(1), 0, nil,
		big.NewInt(0),
	)
	if err != nil {
		t.Fatalf("ошибка деплоя контракта: %v", err)
	}
	if contractAddr == "" {
		t.Error("адрес контракта пустой")
	}
}
