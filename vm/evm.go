// | KB @CerbeRus - Nexus Invest Team
// vm/evm.go

package vm

import (
	"GND/core"
	"GND/types"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"sync"
)

// ExecutionResult represents the result of a contract execution
type ExecutionResult struct {
	GasUsed      uint64
	StateChanges []*StateChange
	ReturnData   []byte
	Error        error
}

// StateChange represents a change to the blockchain state
type StateChange struct {
	Type    ChangeType
	Address string
	Symbol  string
	Amount  *big.Int
	Key     []byte
	Value   []byte
}

// ChangeType represents the type of state change
type ChangeType int

const (
	ChangeTypeBalance ChangeType = iota
	ChangeTypeStorage
)

// CoinConfig represents the configuration for a coin
type CoinConfig struct {
	Symbol          string
	ContractAddress string
	Decimals        uint8
}

// EVMConfig represents the configuration for the EVM
type EVMConfig struct {
	Blockchain core.BlockchainIface
	State      core.StateIface
	GasLimit   uint64
	Coins      []CoinConfig
}

// EVM represents the Ethereum Virtual Machine
type EVM struct {
	config       EVMConfig
	mutex        sync.RWMutex
	eventManager types.EventManager
	ctx          context.Context
}

// NewEVM creates a new EVM instance
func NewEVM(config EVMConfig) *EVM {
	return &EVM{
		config: config,
		ctx:    context.Background(),
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

// DeployContract deploys a contract bytecode, registers it, and deducts the fee in GND (соответствует types.EVMInterface)
func (e *EVM) DeployContract(
	from types.Address,
	bytecode []byte,
	meta types.ContractMeta,
	gasLimit uint64,
	gasPrice *big.Int,
	nonce uint64,
	_ []byte, // signature — зарезервировано для проверки подписи при реализации
	totalSupply *big.Int,
) (string, error) {
	// Validate input parameters
	if len(bytecode) == 0 {
		return "", errors.New("empty contract bytecode")
	}
	if gasLimit == 0 || (gasPrice != nil && gasPrice.Sign() == 0) {
		return "", errors.New("invalid gas parameters")
	}
	if totalSupply == nil {
		return "", errors.New("totalSupply cannot be nil")
	}

	e.mutex.Lock()
	defer e.mutex.Unlock()

	fromAddr := from
	gp := uint64(0)
	if gasPrice != nil {
		gp = gasPrice.Uint64()
	}
	requiredFee := new(big.Int).Mul(
		new(big.Int).SetUint64(gasLimit),
		new(big.Int).SetUint64(gp),
	)

	if len(e.config.Coins) == 0 {
		return "", errors.New("coin configuration is required")
	}
	primarySymbol := e.config.Coins[0].Symbol

	// Check balance
	balance := e.config.State.GetBalance(fromAddr, primarySymbol)
	if balance.Cmp(requiredFee) < 0 {
		return "", fmt.Errorf("insufficient %s for deployment fee (required: %s, available: %s)",
			primarySymbol, requiredFee.String(), balance.String())
	}

	// Generate contract address using keccak256
	addr := fmt.Sprintf("GNDct%x", hashBytes(append(bytecode, byte(nonce))))

	// Check if contract already exists
	if _, exists := ContractRegistry[core.Address(addr)]; exists {
		return "", errors.New("contract with this address already exists")
	}

	meta.Address = addr
	meta.Bytecode = hex.EncodeToString(bytecode)

	contract, err := NewTokenContract(
		core.Address(addr),
		bytecode,
		core.Address(from.String()),
		meta.Name,
		meta.Symbol,
		18, // TODO: make configurable
		totalSupply,
		nil, // TODO: add pool
	)
	if err != nil {
		return "", fmt.Errorf("error creating token contract: %v", err)
	}

	// Register contract
	ContractRegistry[contract.address] = contract

	// Deduct fee
	if err := e.config.State.SubBalance(fromAddr, primarySymbol, requiredFee); err != nil {
		return "", fmt.Errorf("error deducting fee: %v", err)
	}

	return addr, nil
}

// CallContract выполняет функцию контракта
func (e *EVM) CallContract(from string, to string, data []byte, gasLimit uint64, gasPrice uint64, value uint64) (*types.ExecutionResult, error) {
	// Создаем транзакцию
	tx := &core.Transaction{
		Sender:    types.Address(from),
		Recipient: types.Address(to),
		Data:      data,
		GasLimit:  gasLimit,
		GasPrice:  big.NewInt(int64(gasPrice)),
		Value:     big.NewInt(int64(value)),
	}

	// Выполняем транзакцию
	return e.ExecuteTransaction(tx)
}

// GetBalance возвращает баланс GND для адреса
func (e *EVM) GetBalance(address string) (*big.Int, error) {
	balance := e.config.State.GetBalance(types.Address(address), e.config.Coins[0].Symbol)
	if balance == nil {
		return nil, errors.New("failed to get balance")
	}
	return balance, nil
}

// LatestBlock возвращает последний блок
func (e *EVM) LatestBlock() (*core.Block, error) {
	return e.config.Blockchain.LatestBlock()
}

// BlockByNumber возвращает блок по номеру
func (e *EVM) BlockByNumber(number uint64) (*core.Block, error) {
	return e.config.Blockchain.GetBlockByNumber(number)
}

// SendRawTransaction sends a raw transaction
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

// GetTxStatus returns the transaction status
func (e *EVM) GetTxStatus(hash string) (string, error) {
	return e.config.Blockchain.GetTxStatus(hash)
}

// hashBytes generates a hash for address
func hashBytes(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// ExecuteTransaction выполняет транзакцию
func (e *EVM) ExecuteTransaction(tx *core.Transaction) (*types.ExecutionResult, error) {
	// Проверяем баланс отправителя
	balance := e.config.State.GetBalance(types.Address(tx.Sender), e.config.Coins[0].Symbol)
	if balance.Cmp(tx.Value) < 0 {
		return nil, errors.New("insufficient balance")
	}

	// Вычитаем баланс отправителя
	if err := e.config.State.SubBalance(types.Address(tx.Sender), e.config.Coins[0].Symbol, tx.Value); err != nil {
		return nil, err
	}

	// Если это вызов контракта
	if tx.IsContractCall() {
		// Выполняем статический вызов
		result, err := e.config.State.CallStatic(tx)
		if err != nil {
			// Возвращаем баланс в случае ошибки
			e.config.State.AddBalance(types.Address(tx.Sender), e.config.Coins[0].Symbol, tx.Value)
			return nil, err
		}

		// Проверяем баланс получателя
		recipientBalance := e.config.State.GetBalance(types.Address(tx.Recipient), e.config.Coins[0].Symbol)
		if recipientBalance == nil {
			return nil, errors.New("failed to get recipient balance")
		}

		// Добавляем транзакцию в блокчейн
		if err := e.config.Blockchain.AddTx(tx); err != nil {
			return nil, err
		}

		return result, nil
	}

	// Для обычного перевода просто добавляем баланс получателю
	if err := e.config.State.AddBalance(types.Address(tx.Recipient), e.config.Coins[0].Symbol, tx.Value); err != nil {
		// Возвращаем баланс отправителю в случае ошибки
		e.config.State.AddBalance(types.Address(tx.Sender), e.config.Coins[0].Symbol, tx.Value)
		return nil, err
	}

	return &types.ExecutionResult{
		GasUsed: 0, // TODO: Implement gas calculation
		Error:   nil,
	}, nil
}

// GetEVMInstance returns the global EVM instance
func GetEVMInstance() *EVM {
	// TODO: Implement singleton pattern
	return nil
}
