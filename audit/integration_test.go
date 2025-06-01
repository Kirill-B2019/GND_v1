package audit

import (
	"GND/core"
	"GND/vm"
	"math/big"
	"testing"
	"time"
)

// Интеграционный тест: создание блокчейна, кошелька, деплой контракта, проверка баланса
func TestBlockchainIntegration(t *testing.T) {
	// 1. Создание генезис-блока и блокчейна
	genesis := &core.Block{
		Index:        0,
		Timestamp:    time.Now().Unix(),
		PrevHash:     "",
		Hash:         "genesis",
		Transactions: []*core.Transaction{},
	}
	blockchain := core.NewBlockchain(genesis)

	// 2. Создание кошелька
	wallet, err := core.NewWallet()
	if err != nil {
		t.Fatalf("ошибка создания кошелька: %v", err)
	}

	// 3. Кредитуем кошелёк начальными средствами
	blockchain.State.Credit(wallet.Address, big.NewInt(100000))

	// 4. Проверяем баланс
	balance := blockchain.State.GetBalance(wallet.Address)
	if balance.Cmp(big.NewInt(100000)) != 0 {
		t.Errorf("ожидался баланс 1000, получено %s", balance.String())
	}

	// 5. Деплой контракта (эмуляция)
	evm := vm.NewEVM(vm.EVMConfig{
		Blockchain: blockchain,
		State:      blockchain.State,
		GasLimit:   1000000,
	})
	contractAddr, err := evm.DeployContract(
		string(wallet.Address),
		[]byte{0x60, 0x60, 0x60, 0x40}, // пример байткода
		vm.ContractMeta{Name: "TestContract"},
		21000, 1, 0, "",
	)
	if err != nil {
		t.Fatalf("ошибка деплоя контракта: %v", err)
	}

	// 6. Проверяем, что контракт зарегистрирован
	if contractAddr == "" {
		t.Error("адрес контракта пустой")
	}
}
