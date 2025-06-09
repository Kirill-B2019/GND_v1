// vm/evm.go

package vm

import (
	"GND/core"
	"GND/types"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"sync"
)

// EVMConfig определяет параметры виртуальной машины
type EVMConfig struct {
	Blockchain *core.Blockchain
	State      core.StateIface
	GasLimit   uint64            // лимит газа на выполнение одной транзакции/контракта
	Coins      []core.CoinConfig // доступ к конфигурации монет
}

// EVM реализует изолированную виртуальную машину для исполнения байткода контрактов
type EVM struct {
	config       EVMConfig
	mutex        sync.RWMutex
	eventManager types.EventManager
}

// NewEVM создает новый экземпляр EVM
func NewEVM(config EVMConfig) *EVM {
	return &EVM{
		config: config,
	}
}

// SetEventManager устанавливает менеджер событий
func (e *EVM) SetEventManager(em types.EventManager) {
	e.eventManager = em
}

// GetSender возвращает адрес отправителя
func (e *EVM) GetSender() string {
	if len(e.config.Coins) == 0 {
		return ""
	}
	return e.config.Coins[0].ContractAddress
}

// DeployContract деплоит байткод контракта, регистрирует его и списывает комиссию в GND
func (e *EVM) DeployContract(
	from string,
	bytecode []byte,
	meta types.ContractMeta,
	gasLimit uint64,
	gasPrice uint64,
	nonce uint64,
	signature string,
	totalSupply *big.Int,
) (string, error) {
	// Валидация входных параметров
	if len(bytecode) == 0 {
		return "", errors.New("пустой байткод контракта")
	}
	if gasLimit == 0 || gasPrice == 0 {
		return "", errors.New("некорректные параметры газа")
	}
	if totalSupply == nil {
		return "", errors.New("totalSupply не может быть nil")
	}

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

	// Проверка баланса
	balance := e.config.State.GetBalance(fromAddr, primarySymbol)
	if balance.Cmp(requiredFee) < 0 {
		return "", fmt.Errorf("недостаточно %s для комиссии деплоя (требуется: %s, доступно: %s)",
			primarySymbol, requiredFee.String(), balance.String())
	}

	// Генерация адреса контракта с использованием keccak256
	addr := fmt.Sprintf("GNDct%x", hashBytes(append(bytecode, byte(nonce))))

	// Проверка существования контракта
	if _, exists := ContractRegistry[core.Address(addr)]; exists {
		return "", errors.New("контракт с таким адресом уже существует")
	}

	meta.Address = addr
	meta.Bytecode = hex.EncodeToString(bytecode)

	contract, err := NewTokenContract(
		core.Address(addr),
		bytecode,
		core.Address(from),
		meta.Name,
		meta.Symbol,
		18, // TODO: сделать настраиваемым
		totalSupply,
		nil, // TODO: добавить pool
	)
	if err != nil {
		return "", fmt.Errorf("ошибка создания токен-контракта: %v", err)
	}

	// Регистрация контракта
	ContractRegistry[contract.address] = contract

	// Списание комиссии
	if err := e.config.State.SubBalance(fromAddr, primarySymbol, requiredFee); err != nil {
		return "", fmt.Errorf("ошибка списания комиссии: %v", err)
	}

	return addr, nil
}

// CallContract выполняет функцию контракта
func (e *EVM) CallContract(
	from string,
	to string,
	data []byte,
	gasLimit uint64,
	gasPrice uint64,
	value uint64,
	signature string,
) ([]byte, error) {
	fromAddr := core.Address(from)
	toAddr := core.Address(to)
	return e.config.State.CallStatic(fromAddr, toAddr, data, gasLimit, gasPrice, value)
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
