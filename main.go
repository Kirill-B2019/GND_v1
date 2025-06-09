package main

import (
	"GND/api"
	"GND/consensus"
	"GND/core"
	"GND/vm"
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"math/big"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Загрузка глобальной конфигурации
	globalConfig, err := core.InitGlobalConfigDefault()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфига: %v", err)
	}
	cfg := globalConfig.Get()
	log.Println("Node name:", cfg.NodeName)

	// 2. Загрузка настроек консенсуса PoS/PoA
	if len(cfg.Consensus) == 0 {
		cfg.Consensus = []map[string]interface{}{
			{
				"type":                "pos",
				"average_block_delay": "60s",
				"initial_base_target": 153722867,
				"initial_balance":     "10000000",
			},
			{
				"type":                "poa",
				"round_duration":      "17s",
				"sync_duration":       "3s",
				"ban_duration_blocks": 100,
				"warnings_for_ban":    3,
				"max_bans_percentage": 40,
			},
		}
	}

	// Извлечение настройки PoS
	var posConfig core.ConsensusPosConfig
	for _, c := range cfg.Consensus {
		if c["type"] == "pos" {
			data, _ := json.Marshal(c)
			json.Unmarshal(data, &posConfig)
			break
		}
	}
	if posConfig.Type == "" {
		log.Fatal("Конфигурация PoS не найдена")
	}
	consensus.InitPosConsensus(&posConfig)

	// 3. Инициализация пула соединений
	pool, err := core.InitDBPool(ctx, cfg.DB)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer pool.Close()
	go monitorPoolStats(ctx, pool)

	// 4. Генерация кошелька валидатора
	minerWallet, err := core.NewWallet(pool)
	if err != nil {
		log.Fatalf("Ошибка генерации кошелька: %v", err)
	}
	fmt.Printf("Адрес genesis-валидатора: %s\n", minerWallet.Address)

	// 5. Инициализация блокчейна
	var blockchain *core.Blockchain
	genesisBlock := core.NewBlock(0, "", string(minerWallet.Address), []*core.Transaction{}, cfg.GasLimit, "pos")

	blockchain, err = core.LoadBlockchainFromDB(pool)
	if err != nil {
		log.Printf("Не удалось загрузить блокчейн из БД: %v", err)
		blockchain = core.NewBlockchain(genesisBlock, pool)
	} else {
		fmt.Printf("Блокчейн успешно восстановлен из БД\n")
	}
	fmt.Printf("Генезис-блок #%d создан\n", genesisBlock.Index)

	// 6. Кэш множителей для big.Int
	decimalsCache := sync.Map{}
	getDecimalsMultiplier := func(decimals int64) *big.Int {
		if val, ok := decimalsCache.Load(decimals); ok {
			return val.(*big.Int)
		}
		multiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(decimals), nil)
		decimalsCache.Store(decimals, multiplier)
		return multiplier
	}

	// 7. Начисление баланса для монет
	for _, coin := range cfg.Coins {
		amount := new(big.Int)
		if coin.TotalSupply != "" {
			amount.SetString(coin.TotalSupply, 10)
		} else {
			amount.SetInt64(1_000_000)
			amount = amount.Mul(amount, getDecimalsMultiplier(int64(coin.Decimals)))
		}

		_, err := pool.Exec(context.Background(), `
INSERT INTO accounts (address) VALUES ($1)
ON CONFLICT (address) DO NOTHING`,
			string(minerWallet.Address))

		if err != nil {
			log.Fatalf("Failed to insert account: %v", err)
		}
		blockchain.State.Credit(minerWallet.Address, coin.Symbol, amount)
	}

	// 8. Вывод текущего баланса
	fmt.Printf("Баланс адреса %s:\n", minerWallet.Address)
	for _, coin := range cfg.Coins {
		balance := blockchain.State.GetBalance(minerWallet.Address, coin.Symbol)
		fmt.Printf("%s: %s  %s. Знаков: %d\n", coin.Name, balance.String(), coin.Symbol, coin.Decimals)
	}

	// 9. Мемпул
	mempool := core.NewMempool()

	// 10. EVM
	gasLimit := cfg.EVM.GasLimit
	if gasLimit == 0 {
		gasLimit = 10_000_000
	}
	evmInstance := vm.NewEVM(vm.EVMConfig{GasLimit: gasLimit})

	// 11. Серверы
	go func() {
		err := api.StartRPCServer(evmInstance, cfg.Server.RPC.RPCAddr)
		if err != nil {
			fmt.Printf("Ошибка запуска RPCServer %s:\n", err)
		}
	}()
	go api.StartRESTServer(blockchain, mempool, cfg, pool)
	go api.StartWebSocketServer(blockchain, cfg.Server.WS.WSAddr)

	// 12. Обработка транзакций
	go processTransactions(mempool, cfg.MaxWorkers)

	// 13. Грейсфул-шатдаун
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	fmt.Printf("Нода %s ГАНИМЕД запущена.\nДля остановки нажмите Ctrl+C.\n", cfg.NodeName)
	<-sigs
	fmt.Println("Нода ГАНИМЕД остановлена.")
}

// Горутины обработки транзакций
func processTransactions(mempool *core.Mempool, maxWorkers int) {
	sem := make(chan struct{}, maxWorkers)
	logger := log.New(os.Stdout, "[TX Processor] ", log.LstdFlags)

	for {
		sem <- struct{}{}
		go func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Printf("Panic recovered in transaction processor: %v", r)
				}
				<-sem
			}()

			tx, err := mempool.Pop()
			if err != nil {
				logger.Printf("Error popping transaction from mempool: %v", err)
				return
			}

			consType := consensus.SelectConsensusForTx(tx.To)
			logger.Printf("Processing transaction %s through %s consensus", tx.ID, consType)

			switch consType {
			case consensus.ConsensusPoS:
				logger.Printf("Transaction %s: processing through PoS", tx.ID)
			case consensus.ConsensusPoA:
				logger.Printf("Transaction %s: processing through PoA", tx.ID)
			default:
				logger.Printf("Transaction %s: unknown consensus type", tx.ID)
			}
		}()
	}
}

// Мониторинг пула подключений к БД
func monitorPoolStats(ctx context.Context, pool *pgxpool.Pool) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	logger := log.New(os.Stdout, "[DB Pool Monitor] ", log.LstdFlags)

	for {
		select {
		case <-ticker.C:
			stats := pool.Stat()
			logger.Printf(
				"Pool statistics:\n"+
					"  Total connections: %d\n"+
					"  Idle connections: %d\n"+
					"  Acquired connections: %d",
				stats.TotalConns(),
				stats.IdleConns(),
				stats.AcquiredConns(),
			)

			if float64(stats.AcquiredConns())/float64(stats.TotalConns()) > 0.8 {
				logger.Printf("WARNING: Database pool is near capacity (%.1f%% used)",
					float64(stats.AcquiredConns())/float64(stats.TotalConns())*100)
			}
		case <-ctx.Done():
			logger.Println("Database pool monitoring stopped")
			return
		}
	}
}
