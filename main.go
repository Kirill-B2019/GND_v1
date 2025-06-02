package main

import (
	"GND/api"
	"GND/consensus"
	"GND/core"
	"GND/vm"
	"fmt"
	"log"
	"math/big"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// 1. Загрузка глобальной конфигурации
	cfg, err := core.NewConfigFromFile("config/config.json")
	if err != nil {
		log.Fatalf("Ошибка загрузки конфига: %v", err)
	}
	fmt.Printf("Конфиг: gas_limit=%d, network_id=%s\n", cfg.GasLimit, cfg.NetworkID)

	// 2. Загрузка настроек консенсуса PoS/PoA (убрана неиспользуемая переменная)
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
	fmt.Printf(
		"Генезис-блок #%d создан\n",
		genesisBlock.Index,
	)
	// Начисление баланса для первой монеты из конфига
	for _, coin := range cfg.Coins {
		amount := new(big.Int)
		if coin.TotalSupply != "" {
			amount.SetString(coin.TotalSupply, 10)
		} else {
			amount.SetInt64(1_000_000)
			amount = amount.Mul(amount, new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(coin.Decimals)), nil))
		}
		blockchain.State.Credit(minerWallet.Address, coin.Symbol, amount)
	}
	fmt.Printf(
		"Баланс адреса %s:\n",
		minerWallet.Address,
	)
	// Получение баланса
	for _, coin := range cfg.Coins {
		balance := blockchain.State.GetBalance(minerWallet.Address, coin.Symbol)
		fmt.Printf(
			"%s: %s  %s. Знаков: %d\n",
			coin.Name, balance.String(), coin.Symbol, coin.Decimals,
		)
	}

	// 6. Запуск серверов
	evmInstance := vm.NewEVM(vm.EVMConfig{GasLimit: cfg.EVM.GasLimit})
	go api.StartRPCServer(evmInstance, cfg.Server.RPCAddr)
	go api.StartRESTServer(blockchain, mempool, cfg)
	go api.StartWebSocketServer(blockchain, cfg.Server.WebSocketAddr)

	// 7. Обработка транзакций (пример)
	go processTransactions(mempool)

	// 8. Грейсфул-шатдаун
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	fmt.Printf("Нода %s ГАНИМЕД запущена.\nДля остановки нажмите Ctrl+C.\n", cfg.NodeName)
	<-sigs
	fmt.Println("Нода ГАНИМЕД остановлена.\n")
}

func processTransactions(mempool *core.Mempool) {
	for {
		tx, err := mempool.Pop()
		if err != nil {
			continue
		}
		consType := consensus.SelectConsensusForTx(tx.To)
		switch consType {
		case consensus.ConsensusPoS:
			fmt.Printf("Tx %s: обработка через PoS\n", tx.ID)
		case consensus.ConsensusPoA:
			fmt.Printf("Tx %s: обработка через PoA\n", tx.ID)
		}
	}
	select {}
}
