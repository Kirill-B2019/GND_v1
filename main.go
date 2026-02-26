package main

import (
	"GND/api"
	"GND/consensus"
	"GND/core"
	"GND/types"
	"GND/vm"
	"context"
	"encoding/json"
	"fmt"
	"log"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
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
				"type":                "poa",
				"round_duration":      "17s",
				"sync_duration":       "3s",
				"ban_duration_blocks": 100,
				"warnings_for_ban":    3,
				"max_bans_percentage": 40,
			},
			{
				"type":                "pos",
				"average_block_delay": "60s",
				"initial_base_target": 153722867,
				"initial_balance":     "10000000",
			},
		}
	}

	// Извлечение настройки PoA
	var poaConfig core.ConsensusPoaConfig
	for _, c := range cfg.Consensus {
		if c["type"] == "poa" {
			data, _ := json.Marshal(c)
			json.Unmarshal(data, &poaConfig)
			break
		}
	}
	if poaConfig.Type == "" {
		log.Fatal("Конфигурация PoA не найдена")
	}
	consensus.InitPoaConsensus(&poaConfig)

	// 3. Инициализация пула соединений
	pool, err := core.InitDBPool(ctx, cfg.DB)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer pool.Close()
	go monitorPoolStats(ctx, pool)

	// 4. Проверка существующих данных: аккаунты и наличие генезис-блока
	var existingAccount bool
	err = pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM accounts LIMIT 1)").Scan(&existingAccount)
	if err != nil {
		log.Fatalf("Ошибка проверки существующих аккаунтов: %v", err)
	}
	var genesisExists bool
	err = pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM blocks WHERE index = 0)").Scan(&genesisExists)
	if err != nil {
		log.Fatalf("Ошибка проверки генезис-блока: %v", err)
	}

	// 5. Генерация или загрузка кошелька валидатора (у аккаунта может быть несколько кошельков — загружаем первый/валидатора)
	var minerWallet *core.Wallet
	if existingAccount {
		minerWallet, err = core.LoadWallet(pool)
		if err != nil {
			log.Fatalf("Ошибка загрузки существующего кошелька: %v", err)
		}
	} else {
		minerWallet, err = core.NewWallet(pool)
		if err != nil {
			log.Fatalf("Ошибка генерации кошелька: %v", err)
		}
	}
	fmt.Printf("Адрес валидатора: %s\n", minerWallet.Address)

	// 6. Инициализация блокчейна (первый запуск = нет генезис-блока; монеты не пересоздаются, если уже есть в БД)
	var blockchain *core.Blockchain
	if !genesisExists {
		// Первый запуск: создаём генезис-блок
		genesis := &core.Block{
			Index:     0,
			Timestamp: time.Now().UTC(),
			Miner:     string(minerWallet.Address),
			GasUsed:   0,
			GasLimit:  10_000_000,
			Consensus: "poa",
			Nonce:     0,
			Status:    "finalized",
		}
		genesis.Hash = genesis.CalculateHash()
		blockchain = core.NewBlockchain(genesis, pool)
	} else {
		blockchain, err = core.LoadBlockchainFromDB(pool)
		if err != nil {
			log.Fatalf("Ошибка загрузки блокчейна: %v", err)
		}
	}
	// Глобальное состояние для processTransactions и HasSufficientBalance
	if st, ok := blockchain.State.(*core.State); ok {
		core.SetState(st)
	}

	// 11. EVM
	gasLimit := cfg.EVM.GasLimit
	if gasLimit == 0 {
		gasLimit = 10_000_000
	}
	evmInstance := vm.NewEVM(vm.EVMConfig{
		Blockchain: blockchain,
		State:      blockchain.State,
		GasLimit:   gasLimit,
		Coins:      convertCoinsToInterface(cfg.Coins),
	})

	// 8. Первый запуск: деплой монет из config (если ещё нет в БД), генезис, начисление балансов
	if !genesisExists {
		if err := blockchain.FirstLaunch(ctx, pool, minerWallet, cfg); err != nil {
			log.Fatalf("Ошибка инициализации блокчейна: %v", err)
		}
		fmt.Println("Блокчейн успешно инициализирован при первом запуске")
	}

	// 9. Вывод текущего баланса
	fmt.Printf("Баланс адреса %s:\n", minerWallet.Address)
	for _, coin := range cfg.Coins {
		balance := blockchain.State.GetBalance(types.Address(minerWallet.Address), coin.Symbol)
		fmt.Printf("%s: %s  %s. Знаков: %d\n", coin.Name, balance.String(), coin.Symbol, coin.Decimals)
	}

	// 10. Мемпул
	mempool := core.NewMempool()

	// 12. Серверы
	go func() {
		err := api.StartRPCServer(evmInstance, cfg.Server.RPC.RPCAddr)
		if err != nil {
			fmt.Printf("Ошибка запуска RPCServer %s:\n", err)
		}
	}()
	go api.StartRESTServer(blockchain, mempool, cfg, pool, evmInstance)
	go api.StartWebSocketServer(blockchain, mempool, cfg)

	// 13. Обработка транзакций
	go processTransactions(mempool, cfg.MaxWorkers)

	// 14. Грейсфул-шатдаун
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
				if err.Error() == "timeout" {
					// Это нормальная ситуация, когда нет транзакций
					return
				}
				logger.Printf("Error popping transaction from mempool: %v", err)
				return
			}

			consType := consensus.SelectConsensusForTx(string(tx.Recipient))
			logger.Printf("Processing transaction %s through %s consensus", tx.ID, consType)

			// Process transaction based on consensus type
			switch consType {
			case consensus.ConsensusPoS:
				if err := processPoSTransaction(tx); err != nil {
					logger.Printf("Error processing PoS transaction %s: %v", tx.ID, err)
					return
				}
				logger.Printf("Successfully processed PoS transaction %s", tx.ID)
			case consensus.ConsensusPoA:
				if err := processPoATransaction(tx); err != nil {
					logger.Printf("Error processing PoA transaction %s: %v", tx.ID, err)
					return
				}
				logger.Printf("Successfully processed PoA transaction %s", tx.ID)
			default:
				logger.Printf("Transaction %s: unknown consensus type", tx.ID)
			}
		}()
	}
}

func processPoSTransaction(tx *core.Transaction) error {
	// Validate transaction
	if err := tx.Validate(); err != nil {
		return fmt.Errorf("transaction validation failed: %v", err)
	}

	// Check sender's balance
	if !tx.HasSufficientBalance() {
		return fmt.Errorf("insufficient balance for transaction")
	}

	// Process transaction through EVM if it's a contract call
	if tx.IsContractCall() {
		return processContractTransaction(tx)
	}

	// Process regular transfer
	return processTransferTransaction(tx)
}

func processPoATransaction(tx *core.Transaction) error {
	// Validate transaction
	if err := tx.Validate(); err != nil {
		return fmt.Errorf("transaction validation failed: %v", err)
	}

	// Check sender's balance
	if !tx.HasSufficientBalance() {
		return fmt.Errorf("insufficient balance for transaction")
	}

	// Process transaction through EVM if it's a contract call
	if tx.IsContractCall() {
		return processContractTransaction(tx)
	}

	// Process regular transfer
	return processTransferTransaction(tx)
}

func processContractTransaction(tx *core.Transaction) error {
	// Get state
	state := core.GetState()
	if state == nil {
		return fmt.Errorf("state not available")
	}

	// Create EVM instance
	evm := vm.NewEVM(vm.EVMConfig{
		State:    state,
		GasLimit: 1000000, // Default gas limit
	})

	// Execute transaction in EVM
	result, err := evm.CallContract(
		string(tx.Sender),
		string(tx.Recipient),
		tx.Data,
		tx.GasLimit,
		tx.GasPrice.Uint64(),
		tx.Value.Uint64(),
	)
	if err != nil {
		return fmt.Errorf("EVM execution failed: %v", err)
	}

	// Update state with result
	if err := updateStateWithResult(tx, result); err != nil {
		return fmt.Errorf("failed to update state: %v", err)
	}

	return nil
}

func processTransferTransaction(tx *core.Transaction) error {
	// Update balances
	if err := updateBalances(tx); err != nil {
		return fmt.Errorf("failed to update balances: %v", err)
	}

	return nil
}

func updateBalances(tx *core.Transaction) error {
	// Get state
	state := core.GetState()
	if state == nil {
		return fmt.Errorf("state not available")
	}

	// Update sender's balance
	if err := state.SubBalance(tx.Sender, "GND", tx.Value); err != nil {
		return err
	}

	// Update recipient's balance
	if err := state.AddBalance(tx.Recipient, "GND", tx.Value); err != nil {
		// Revert sender's balance if recipient update fails
		state.AddBalance(tx.Sender, "GND", tx.Value)
		return err
	}

	return nil
}

func updateStateWithResult(tx *core.Transaction, result *types.ExecutionResult) error {
	// Get state
	state := core.GetState()
	if state == nil {
		return fmt.Errorf("state not available")
	}

	// Update state with execution result
	if err := state.ApplyExecutionResult(tx, result); err != nil {
		return fmt.Errorf("failed to apply execution result: %v", err)
	}

	return nil
}

// Функция для конвертации []core.CoinConfig в []vm.CoinConfig
func convertCoinsToInterface(coins []core.CoinConfig) []vm.CoinConfig {
	result := make([]vm.CoinConfig, len(coins))
	for i, coin := range coins {
		result[i] = vm.CoinConfig{
			Symbol:          coin.Symbol,
			ContractAddress: coin.ContractAddress,
			Decimals:        uint8(coin.Decimals),
		}
	}
	return result
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
