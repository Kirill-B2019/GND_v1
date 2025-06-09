package main

import (
	"GND/api"
	"GND/consensus"
	"GND/core"
	"GND/vm"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sync"
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
		// Первый запуск - инициализируем блокчейн
		blockchain = core.NewBlockchain(nil, pool)
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

	// 7. Кэш множителей для big.Int
	decimalsCache := sync.Map{}
	getDecimalsMultiplier := func(decimals int64) *big.Int {
		if val, ok := decimalsCache.Load(decimals); ok {
			return val.(*big.Int)
		}
		multiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(decimals), nil)
		decimalsCache.Store(decimals, multiplier)
		return multiplier
	}

	// 11. EVM
	gasLimit := cfg.EVM.GasLimit
	if gasLimit == 0 {
		gasLimit = 10_000_000
	}
	evmInstance := vm.NewEVM(vm.EVMConfig{GasLimit: gasLimit})

	// 8. Начисление баланса для монет (только если это новый аккаунт)
	if !existingAccount {
		for _, coin := range cfg.Coins {
			amount := new(big.Int)
			if coin.TotalSupply != "" {
				amount.SetString(coin.TotalSupply, 10)
			} else {
				amount.SetInt64(1_000_000)
				amount = amount.Mul(amount, getDecimalsMultiplier(int64(coin.Decimals)))
			}

			// Проверяем стандарт токена
			if coin.Standard == "gndst1" {
				addr, err := evmInstance.DeployGNDst1Token(ctx, coin.Name, coin.Symbol, uint8(coin.Decimals), amount)
				if err != nil {
					log.Fatalf("Ошибка деплоя токена GNDst1: %v", err)
				}
				fmt.Printf("GNDst1 токен %s (%s) задеплоен по адресу %s\n", coin.Name, coin.Symbol, addr)
				continue
			}

			// Проверяем существование контракта
			var contractExists bool
			err = pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM contracts WHERE address = $1)", coin.ContractAddress).Scan(&contractExists)
			if err != nil {
				log.Fatalf("Ошибка проверки контракта: %v", err)
			}

			if !contractExists {
				_, err = pool.Exec(ctx, `
					INSERT INTO contracts (address, owner, type, created_at)
					VALUES ($1, $2, 'token', NOW())
					ON CONFLICT (address) DO NOTHING`,
					coin.ContractAddress, minerWallet.Address)
				if err != nil {
					log.Fatalf("Ошибка создания контракта: %v", err)
				}
			}

			// Проверяем существование токена
			var tokenExists bool
			err = pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM tokens WHERE symbol = $1)", coin.Symbol).Scan(&tokenExists)
			if err != nil {
				log.Fatalf("Ошибка проверки токена: %v", err)
			}

			if !tokenExists {
				_, err = pool.Exec(ctx, `
					INSERT INTO tokens (contract_id, standard, symbol, name, decimals, total_supply)
					SELECT id, $1, $2, $3, $4, $5
					FROM contracts WHERE address = $6`,
					coin.Standard, coin.Symbol, coin.Name, coin.Decimals, amount.String(), coin.ContractAddress)
				if err != nil {
					log.Fatalf("Ошибка создания токена: %v", err)
				}
			}

			_, err = pool.Exec(ctx, `
				INSERT INTO token_balances (token_id, address, balance)
				SELECT t.id, $1, $2
				FROM tokens t
				WHERE t.symbol = $3
				ON CONFLICT (token_id, address) DO UPDATE
				SET balance = $2`,
				minerWallet.Address, amount.String(), coin.Symbol)
			if err != nil {
				log.Fatalf("Ошибка создания баланса токена: %v", err)
			}

			blockchain.State.Credit(minerWallet.Address, coin.Symbol, amount)
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
