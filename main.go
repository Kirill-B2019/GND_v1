package main

import (
	"GND/api"
	"GND/consensus"
	"GND/core"
	"GND/vm"
	"fmt"
	"log"
	"math/big"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	// 1. Загрузка глобальной конфигурации
	cfg, err := core.NewConfigFromFile("config/config.json")
	if err != nil {
		log.Fatalf("Ошибка загрузки конфига: %v", err)
	}
	fmt.Printf("Конфиг: gas_limit=%d, network_id=%s\n", cfg.GasLimit, cfg.NetworkID)

	// 2. Загрузка настроек консенсуса PoS/PoA
	if _, err := consensus.LoadConsensusSettings(cfg.ConsensusConf); err != nil {
		log.Fatalf("Ошибка загрузки consensus.json: %v", err)
	}

	// 3. Генерация кошелька валидатора
	minerWallet, err := core.NewWallet()
	if err != nil {
		log.Fatalf("Ошибка генерации кошелька: %v", err)
	}
	fmt.Printf("Адрес genesis-валидатора: %s\n", minerWallet.Address)

	// 4. Создание генезис-блока
	genesisBlock := core.NewBlock(
		0,
		"",
		string(minerWallet.Address),
		[]*core.Transaction{},
		cfg.GasLimit,
		"pos",
	)

	// 5. Инициализация блокчейна и мемпула
	blockchain := core.NewBlockchain(genesisBlock)
	mempool := core.NewMempool()
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
		blockchain.State.Credit(minerWallet.Address, coin.Symbol, amount)
	}
	fmt.Printf("Баланс адреса %s:\n", minerWallet.Address)
	for _, coin := range cfg.Coins {
		balance := blockchain.State.GetBalance(minerWallet.Address, coin.Symbol)
		fmt.Printf("%s: %s  %s. Знаков: %d\n", coin.Name, balance.String(), coin.Symbol, coin.Decimals)
	}

	// 6. Запуск серверов
	evmInstance := vm.NewEVM(vm.EVMConfig{GasLimit: cfg.EVM.GasLimit})
	go func() {
		err := api.StartRPCServer(evmInstance, cfg.Server.RPCAddr)
		if err != nil {
			fmt.Printf("Ошибка запуска RPCServer %s:\n", err)
		}
	}()
	go api.StartRESTServer(blockchain, mempool, cfg)
	go api.StartWebSocketServer(blockchain, cfg.Server.WebSocketAddr)

	// 7. Обработка транзакций через worker pool
	go processTransactions(mempool, cfg.MaxWorkers)

	// 8. Мониторинг числа горутин
	/*	go func() {
		for {
			log.Printf("Goroutines: %d", runtime.NumGoroutine())
			time.Sleep(5 * time.Second)
		}
	}()*/

	// 9. Запуск pprof для профилирования
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	// 10. Грейсфул-шатдаун
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	fmt.Printf("Нода %s ГАНИМЕД запущена.\nДля остановки нажмите Ctrl+C.\n", cfg.NodeName)
	<-sigs
	fmt.Println("Нода ГАНИМЕД остановлена.\n")
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
