package main

import (
	"GND/utils"
	"GND/vm"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"GND/api"
	"GND/consensus"
	"GND/core"
)

func main() {
	// 1. Загрузка глобальной конфигурации
	cfg, err := core.NewConfigFromFile("config/config.json")
	if err != nil {
		log.Fatalf("Ошибка загрузки конфига: %v", err)
	}
	fmt.Printf("Конфиг: consensus=%s, gas_limit=%d, network_id=%s\n", cfg.ConsensusType, cfg.GasLimit, cfg.NetworkID)

	// 2. Генерация или загрузка кошелька валидатора/майнера
	minerWallet, err := core.NewWallet()
	if err != nil {
		log.Fatalf("Ошибка генерации кошелька: %v", err)
	}
	fmt.Printf("Адрес валидатора/майнера: %s\n", utils.AddPrefix(string(minerWallet.Address)))

	// 3. Создание генезис-блока
	genesisBlock := core.NewBlock(
		0,                           // index
		"",                          // prevHash
		string(minerWallet.Address), // miner
		[]*core.Transaction{},       // пустой список транзакций
		cfg.GasLimit,                // gasLimit
		cfg.ConsensusType,           // consensus
	)

	// 4. Инициализация блокчейна
	blockchain := core.NewBlockchain(genesisBlock)
	blockchain.State.Credit(string(minerWallet.Address), 1_000_000_000) // начальный баланс GND

	// 5. Инициализация мемпула транзакций
	mempool := core.NewMempool()

	// 6. Запуск выбранного алгоритма консенсуса
	var consensusEngine consensus.Consensus
	switch cfg.ConsensusType {
	case "pos":
		consensusEngine = consensus.NewPoS()
	case "poa":
		consensusEngine = consensus.NewPoA()
	default:
		log.Fatalf("Неизвестный тип консенсуса: %s", cfg.ConsensusType)
	}
	consensusEngine.Start(blockchain, mempool)
	fmt.Printf("Консенсус %s запущен\n", consensusEngine.Type())

	// 7. Запуск REST API (асинхронно)
	evmInstance := vm.NewEVM(vm.EVMConfig{GasLimit: 10000000}) //
	api.StartRPCServer(evmInstance, ":8081")
	go api.StartRESTServer(blockchain, mempool, cfg)
	// Запуск WebSocket-сервера
	go api.StartWebSocketServer(blockchain, cfg)

	// 8. Грейсфул-шатдаун (Ctrl+C)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	fmt.Println("Нода GND_0x0001 запущен. ")
	fmt.Println("Блокчейн ГАНИМЕД запущен. Для остановки нажмите Ctrl+C.")
	<-sigs

	// 9. Остановка консенсуса и сервисов
	consensusEngine.Stop()
	fmt.Println("Нода ГАНИМЕД остановлена.")

	select {}
}
