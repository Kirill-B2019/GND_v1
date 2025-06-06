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
	// После загрузки основного конфига
	if len(cfg.Consensus) == 0 {
		// Если консенсусы не заданы вообще — добавляем PoS и PoA
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

	// Извлекаем настройки PoS
	var posConfig core.ConsensusPosConfig
	for _, c := range cfg.Consensus {
		if c["type"] == "pos" {
			data, _ := json.Marshal(c)
			json.Unmarshal(data, &posConfig)
			break
		}
	}

	// Если PoS всё ещё пустой — fatal error
	if posConfig.Type == "" {
		log.Fatal("Конфигурация PoS не найдена")
	}
	// Инициализируем модуль консенсуса
	consensus.InitPosConsensus(&posConfig)

	// 2. Инициализация пула соединений
	pool, err := core.InitDBPool(ctx, cfg.DB)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer pool.Close()
	go monitorPoolStats(ctx, pool) //Total — всего соединений в пуле (и занятых, и свободных) Idle — сколько соединений сейчас свободно Acquired — сколько соединений занято (используется приложением)

	// 3. Генерация кошелька валидатора
	minerWallet, err := core.NewWallet(pool)
	if err != nil {
		log.Fatalf("Ошибка генерации кошелька: %v", err)
	}
	fmt.Printf("Адрес genesis-валидатора: %s\n", minerWallet.Address)

	// 4. Инициализация блокчейна
	var blockchain *core.Blockchain
	genesisBlock := core.NewBlock(
		0,
		"",
		string(minerWallet.Address),
		[]*core.Transaction{},
		cfg.GasLimit,
		"pos",
	)

	// Попытка загрузить существующую цепочку из БД
	blockchain, err = core.LoadBlockchainFromDB(pool)
	if err != nil {
		log.Printf("Не удалось загрузить блокчейн из БД: %v", err)
		blockchain = core.NewBlockchain(genesisBlock, pool)
	} else {
		fmt.Printf("Блокчейн успешно восстановлен из БД\n")
	}
	fmt.Printf("Генезис-блок #%d создан\n", genesisBlock.Index)

	// Кэш множителей для big.Int
	decimalsCache := sync.Map{}

	getDecimalsMultiplier := func(decimals int64) *big.Int {
		if val, ok := decimalsCache.Load(decimals); ok {
			return val.(*big.Int)
		}
		multiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(decimals), nil)
		decimalsCache.Store(decimals, multiplier)
		return multiplier
	}

	// Начисление баланса для первой монеты из конфига
	for _, coin := range cfg.Coins {
		amount := new(big.Int)
		if coin.TotalSupply != "" {
			amount.SetString(coin.TotalSupply, 10)
		} else {
			amount.SetInt64(1_000_000)
			amount = amount.Mul(amount, getDecimalsMultiplier(int64(coin.Decimals)))
		}
		// First save wallet to insert account
		_, err := pool.Exec(context.Background(), `
	INSERT INTO accounts (address) VALUES ($1)`,
			string(minerWallet.Address))
		if err != nil {
			log.Fatalf("Failed to insert account: %v", err)
		}
		blockchain.State.Credit(minerWallet.Address, coin.Symbol, amount)
	}
	fmt.Printf("Баланс адреса %s:\n", minerWallet.Address)
	for _, coin := range cfg.Coins {
		balance := blockchain.State.GetBalance(minerWallet.Address, coin.Symbol)
		fmt.Printf("%s: %s  %s. Знаков: %d\n", coin.Name, balance.String(), coin.Symbol, coin.Decimals)
	}
	// 5. мемпула
	mempool := core.NewMempool()

	// 6. Запуск серверов
	gasLimit := cfg.EVM.GasLimit
	if gasLimit == 0 {
		gasLimit = 10_000_000 // default EVM Gas Limit
	}
	evmInstance := vm.NewEVM(vm.EVMConfig{GasLimit: gasLimit})
	go func() {
		err := api.StartRPCServer(evmInstance, cfg.Server.RPC.RPCAddr)
		if err != nil {
			fmt.Printf("Ошибка запуска RPCServer %s:\n", err)
		}
	}()
	go api.StartRESTServer(blockchain, mempool, cfg, pool)
	go api.StartWebSocketServer(blockchain, cfg.Server.WS.WSAddr)

	// 7. Обработка транзакций через worker pool
	go processTransactions(mempool, cfg.MaxWorkers)

	// 8. Мониторинг числа горутин
	/*	go func() {
		for {
			log.Printf("Goroutines: %d", runtime.NumGoroutine())
			time.Sleep(5 * time.Second)
		}
	}()*/

	// 10. Грейсфул-шатдаун
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	fmt.Printf("Нода %s ГАНИМЕД запущена.\nДля остановки нажмите Ctrl+C.\n", cfg.NodeName)
	<-sigs
	fmt.Println("Нода ГАНИМЕД остановлена.")
}

// Ограничение числа воркеров для обработки транзакций
func processTransactions(mempool *core.Mempool, maxWorkers int) {
	sem := make(chan struct{}, maxWorkers)
	for {
		sem <- struct{}{}
		go func() {
			defer func() { <-sem }()
			tx, err := mempool.Pop()
			if err != nil {
				return
			}
			consType := consensus.SelectConsensusForTx(tx.To)
			switch consType {
			case consensus.ConsensusPoS:
				fmt.Printf("Tx %s: обработка через PoS\n", tx.ID)
			case consensus.ConsensusPoA:
				fmt.Printf("Tx %s: обработка через PoA\n", tx.ID)
			}
		}()
	}
}
func monitorPoolStats(ctx context.Context, pool *pgxpool.Pool) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			stats := pool.Stat()
			log.Printf(
				"Pool stats: Total=%d, Idle=%d, Acquired=%d",
				stats.TotalConns(),
				stats.IdleConns(),
				stats.AcquiredConns(),
			)
		case <-ctx.Done():
			return
		}
	}
}
