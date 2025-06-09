package main

import (
	"GND/api"
	"GND/consensus"
	"GND/core"
	"GND/vm"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
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

	// 4. Проверка существующих данных
	var existingAccount bool
	err = pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM accounts LIMIT 1)").Scan(&existingAccount)
	if err != nil {
		log.Fatalf("Ошибка проверки существующих аккаунтов: %v", err)
	}

	// 5. Генерация или загрузка кошелька валидатора
	var minerWallet *core.Wallet
	if existingAccount {
		// Загружаем существующий кошелек
		minerWallet, err = core.LoadWallet(pool)
		if err != nil {
			log.Fatalf("Ошибка загрузки существующего кошелька: %v", err)
		}
	} else {
		// Создаем новый кошелек
		minerWallet, err = core.NewWallet(pool)
		if err != nil {
			log.Fatalf("Ошибка генерации кошелька: %v", err)
		}
	}
	fmt.Printf("Адрес валидатора: %s\n", minerWallet.Address)

	// 6. Инициализация блокчейна
	var blockchain *core.Blockchain
	if !existingAccount {
		// Создаем генезис-блок
		genesis := &core.Block{
			Index:     0,
			Timestamp: time.Now(),
			Miner:     string(minerWallet.Address),
			GasUsed:   0,
			GasLimit:  10_000_000,
			Consensus: "poa",
			Nonce:     "0",
			Status:    "finalized",
		}
		genesis.Hash = genesis.CalculateHash()

		// Первый запуск - инициализируем блокчейн
		blockchain = core.NewBlockchain(genesis, pool)
		if err := blockchain.FirstLaunch(ctx, pool, minerWallet, cfg); err != nil {
			log.Fatalf("Ошибка инициализации блокчейна: %v", err)
		}
		fmt.Println("Блокчейн успешно инициализирован при первом запуске")
	} else {
		// Загружаем существующий блокчейн
		blockchain, err = core.LoadBlockchainFromDB(pool)
		if err != nil {
			log.Fatalf("Ошибка загрузки блокчейна из БД: %v", err)
		}
		fmt.Println("Блокчейн успешно восстановлен из БД")
	}

	// 11. EVM
	gasLimit := cfg.EVM.GasLimit
	if gasLimit == 0 {
		gasLimit = 10_000_000
	}
	evmInstance := vm.NewEVM(vm.EVMConfig{GasLimit: gasLimit})

	// 8. Начисление баланса для монет (только если это новый аккаунт)
	if !existingAccount {
		// Создаем генезис-блок
		genesis := &core.Block{
			Index:     0,
			Timestamp: time.Now(),
			Miner:     string(minerWallet.Address),
			GasUsed:   0,
			GasLimit:  10_000_000,
			Consensus: "poa",
			Nonce:     "0",
			Status:    "finalized",
		}
		genesis.Hash = genesis.CalculateHash()

		blockchain = core.NewBlockchain(genesis, pool)

		for i := range cfg.Coins {
			coin := &cfg.Coins[i]
			contractAddress := coin.ContractAddress
			if contractAddress == "" {
				contractAddress = fmt.Sprintf("GNDct%s", core.GenerateContractAddress())
				coin.ContractAddress = contractAddress
			}

			// Проверяем, существует ли токен с таким символом
			var existingTokenID int
			err = pool.QueryRow(ctx, "SELECT id FROM tokens WHERE symbol = $1", coin.Symbol).Scan(&existingTokenID)
			if err == nil {
				fmt.Printf("[DEBUG] Токен с символом %s уже существует, пропускаем создание\n", coin.Symbol)
				continue
			} else if err != sql.ErrNoRows {
				log.Fatalf("ошибка проверки существования токена %s: %v", coin.Symbol, err)
			}

			token := core.NewToken(
				contractAddress,
				coin.Symbol,
				coin.Name,
				coin.Decimals,
				coin.TotalSupply,
				string(minerWallet.Address),
				"gndst1",
				coin.Standard,
				int(genesis.Index),
				0,
			)

			if err := token.SaveToDB(ctx, pool); err != nil {
				log.Fatalf("ошибка создания токена %s: %v", coin.Symbol, err)
			}

			amount := new(big.Int)
			if coin.TotalSupply != "" {
				amount.SetString(coin.TotalSupply, 10)
			} else {
				amount.SetInt64(1_000_000)
				multiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(coin.Decimals)), nil)
				amount.Mul(amount, multiplier)
			}

			_, err := pool.Exec(ctx, `
				INSERT INTO token_balances (token_id, address, balance)
				VALUES (
					(SELECT t.id FROM tokens t JOIN contracts c ON t.contract_id = c.id WHERE c.address = $1),
					$2,
					$3
				)
				ON CONFLICT (token_id, address) DO UPDATE
				SET balance = EXCLUDED.balance`,
				contractAddress, string(minerWallet.Address), amount.String(),
			)
			if err != nil {
				log.Fatalf("ошибка создания баланса токена: %v", err)
			}

			tx := &core.Transaction{
				Sender:    "SYSTEM",
				Recipient: string(minerWallet.Address),
				Value:     amount,
				Fee:       big.NewInt(0),
				Nonce:     0,
				Type:      "token_mint",
				Status:    "confirmed",
				Timestamp: time.Now(),
				Symbol:    coin.Symbol,
			}
			tx.Hash = tx.CalculateHash()
			err = tx.Save(ctx, pool)
			if err != nil {
				log.Fatalf("Ошибка сохранения транзакции начисления токена: %v", err)
			}
			if blockchain != nil && blockchain.LatestBlock() != nil && blockchain.LatestBlock().Index == 0 {
				blockchain.LatestBlock().Transactions = append(blockchain.LatestBlock().Transactions, tx)
				// Обновляем состояние блокчейна после добавления транзакции
				if err := blockchain.State.ApplyTransaction(tx); err != nil {
					log.Printf("Ошибка применения транзакции в генезис-блоке: %v", err)
				}
			}
		}

		// После цикла, если были изменения, обновим coins.json
		coinsPath := "config/coins.json"
		coinsFile, err := os.OpenFile(coinsPath, os.O_WRONLY|os.O_TRUNC, 0644)
		if err == nil {
			defer coinsFile.Close()
			json.NewEncoder(coinsFile).Encode(map[string]interface{}{"coins": cfg.Coins})
		}
	}

	// 9. Вывод текущего баланса
	fmt.Printf("Баланс адреса %s:\n", minerWallet.Address)
	for _, coin := range cfg.Coins {
		balance := blockchain.State.GetBalance(minerWallet.Address, coin.Symbol)
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
	go api.StartRESTServer(blockchain, mempool, cfg, pool)
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

			consType := consensus.SelectConsensusForTx(tx.Recipient)
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
