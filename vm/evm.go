//vm/evm.go

package vm

import (
	"GND/core"
	_ "GND/utils"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"sync"
)

type ContractRegistryType map[core.Address]Contract

func GetContract(addr core.Address) (Contract, bool) {
	c, ok := ContractRegistry[addr]
	return c, ok
}

// EVMConfig определяет параметры виртуальной машины
type EVMConfig struct {
	Blockchain *core.Blockchain
	State      *core.State
	GasLimit   uint64 // лимит газа на выполнение одной транзакции/контракта
}

// EVM реализует изолированную виртуальную машину для исполнения байткода контрактов
type EVM struct {
	config EVMConfig
	mutex  sync.RWMutex
}

// NewEVM создает новый экземпляр EVM
func NewEVM(config EVMConfig) *EVM {
	return &EVM{
		config: config,
	}
}

// DeployContract деплоит байткод контракта, регистрирует его и списывает комиссию в GND
func (evm *EVM) DeployContract(
	from string,
	bytecode []byte,
	meta ContractMeta,
	gasLimit, gasPrice, nonce uint64,
	signature string,
) (string, error) {
	evm.mutex.Lock()
	defer evm.mutex.Unlock()

	fromAddr := core.Address(from)
	requiredFee := new(big.Int).Mul(
		new(big.Int).SetUint64(gasLimit),
		new(big.Int).SetUint64(gasPrice),
	)

	// Проверка баланса
	if evm.config.State.GetBalance(fromAddr).Cmp(requiredFee) < 0 {
		return "", errors.New("insufficient GND for deploy fee")
	}

	// Генерация адреса контракта
	addr := fmt.Sprintf("GNDct%x", hashBytes(append(bytecode, byte(nonce))))
	meta.Address = addr
	meta.Bytecode = hex.EncodeToString(bytecode)

	// Регистрация контракта
	contract := NewTokenContract(
		core.Address(addr),
		bytecode,
		core.Address(from),
		meta.Name,
		meta.Symbol,
		18,
	)
	ContractRegistry[core.Address(addr)] = contract

	// Списание комиссии
	evm.config.State.SubBalance(fromAddr, requiredFee)

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
		from, to,
		value, gasLimit, gasPrice, nonce,
		data,
		core.TxContractCall, // тип транзакции
		signature,
	)
	return tx.Hash, nil
}

// GetBalance возвращает баланс GND для адреса
func (e *EVM) GetBalance(address string) (*big.Int, error) {
	addr := core.Address(address)
	return e.config.State.GetBalance(addr), nil
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
	var h uint64 = 5381
	for _, b := range data {
		h = ((h << 5) + h) + uint64(b)
	}
	out := make([]byte, 8)
	for i := 0; i < 8; i++ {
		out[i] = byte((h >> (8 * uint(i))) & 0xff)
	}
	return out
}
