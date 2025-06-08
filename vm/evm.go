// vm/evm.go
// vm/evm.go

package vm

import (
	"GND/core"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"sync"
)

// Определение EVMInterface для инъекции зависимости
type EVMInterface interface {
	CallStatic(from core.Address, to core.Address, data []byte, gasLimit uint64, gasPrice uint64, value uint64) ([]byte, error)
}

// ContractRegistryType — тип реестра контрактов
type ContractRegistryType map[core.Address]Contract

var ContractRegistry ContractRegistryType = make(ContractRegistryType)

func GetContract(addr core.Address) (Contract, bool) {
	c, ok := ContractRegistry[addr]
	return c, ok
}

// EVMConfig определяет параметры виртуальной машины
type EVMConfig struct {
	Blockchain *core.Blockchain
	State      *core.State
	GasLimit   uint64            // лимит газа на выполнение одной транзакции/контракта
	Coins      []core.CoinConfig // доступ к конфигурации монет
}

// EVM реализует изолированную виртуальную машину для исполнения байткода контрактов
type EVM struct {
	config EVMConfig
	mutex  sync.RWMutex
	evm    EVMInterface // Теперь интерфейс определен корректно
}

// NewEVM создает новый экземпляр EVM
func NewEVM(config EVMConfig) *EVM {
	return &EVM{
		config: config,
	}
}

// DeployContract деплоит байткод контракта, регистрирует его и списывает комиссию в GND
func (e *EVM) DeployContract(
	from string,
	bytecode []byte,
	meta ContractMeta,
	gasLimit uint64,
	gasPrice uint64,
	nonce uint64,
	signature string,
) (string, error) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	fromAddr := core.Address(from)
	requiredFee := new(big.Int).Mul(
		new(big.Int).SetUint64(gasLimit),
		new(big.Int).SetUint64(gasPrice),
	)

	if len(e.config.Coins) == 0 {
		return "", errors.New("требуется настройка монеты")
	}
	primarySymbol := e.config.Coins[0].Symbol

	if e.config.State.GetBalance(fromAddr, primarySymbol).Cmp(requiredFee) < 0 {
		return "", errors.New("недостаточно " + primarySymbol + " для комиссии деплоя")
	}

	addr := fmt.Sprintf("GNDct%x", hashBytes(append(bytecode, byte(nonce))))
	meta.Address = addr
	meta.Bytecode = hex.EncodeToString(bytecode)

	contract := NewTokenContract(
		core.Address(addr),
		bytecode,
		core.Address(from),
		meta.Name,
		meta.Symbol,
		18,
	)
	RegisterContract(contract.address, contract)

	e.config.State.SubBalance(fromAddr, primarySymbol, requiredFee)

	return addr, nil
}

// CallContract выполняет функцию контракта
func (e *EVM) CallContract(from, to string, data []byte, gasLimit, gasPrice, value uint64, signature string) ([]byte, error) {
	fromAddr := core.Address(from)
	toAddr := core.Address(to)
	return e.config.State.CallStatic(fromAddr, toAddr, data, gasLimit, gasPrice, value)
}

// SendContractTx отправляет транзакцию контракта
func (e *EVM) SendContractTx(from, to string, data []byte, gasLimit, gasPrice, value, nonce uint64, signature string) (string, error) {
	tx, _ := core.NewTransaction(
		from,
		to,
		"GND.c",
		new(big.Int).SetUint64(value),
		gasLimit,
		gasPrice,
		nonce,
		data,
		core.TxContractCall,
		signature,
	)
	return tx.Hash, nil
}

// GetBalance возвращает баланс GND для адреса
func (e *EVM) GetBalance(address string) (*big.Int, error) {
	primarySymbol := e.config.Coins[0].Symbol
	addr := core.Address(address)
	return e.config.State.GetBalance(addr, primarySymbol), nil
}

// LatestBlock возвращает последний блок
func (e *EVM) LatestBlock() *core.Block {
	return e.config.Blockchain.LatestBlock()
}

// BlockByNumber возвращает блок по номеру
func (e *EVM) BlockByNumber(number uint64) *core.Block {
	return e.config.Blockchain.GetBlockByNumber(number)
}

// SendRawTransaction отправляет сырую транзакцию
func (e *EVM) SendRawTransaction(rawTx []byte) (string, error) {
	tx, err := core.DecodeRawTransaction(rawTx)
	if err != nil {
		return "", err
	}
	if err := e.config.Blockchain.AddTx(tx); err != nil {
		return "", err
	}
	return tx.Hash, nil
}

// GetTxStatus возвращает статус транзакции
func (e *EVM) GetTxStatus(hash string) (string, error) {
	return e.config.Blockchain.GetTxStatus(hash)
}

// hashBytes генерирует хеш для адресации
func hashBytes(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}
