package types

import (
	"math/big"
)

// EVMInterface defines the interface for virtual machine
type EVMInterface interface {
	// DeployContract deploys a new contract
	DeployContract(
		from Address,
		bytecode []byte,
		meta ContractMeta,
		gasLimit uint64,
		gasPrice *big.Int,
		nonce uint64,
		signature []byte,
		totalSupply *big.Int,
	) (string, error)

	// CallContract executes a contract call
	CallContract(
		from Address,
		to Address,
		data []byte,
		gasLimit uint64,
		gasPrice *big.Int,
		value *big.Int,
		signature []byte,
	) ([]byte, error)

	// GetBalance returns the balance for an address
	GetBalance(address Address) (*big.Int, error)
}

// ContractMeta contains contract metadata
type ContractMeta struct {
	Name        string
	Symbol      string
	Standard    string
	Owner       string
	Description string
	Version     string
	Compiler    string
	Params      map[string]string
	MetadataCID string
	SourceCode  string
	Address     string
	Bytecode    string
}

// ContractInterface defines the interface for a contract
type ContractInterface interface {
	// Execute executes a contract method
	Execute(method string, args []interface{}) (interface{}, error)

	// Address returns the contract address
	Address() string

	// Bytecode returns the contract bytecode
	Bytecode() []byte
}

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
type ChangeType uint8

const (
	ChangeTypeBalance ChangeType = iota
	ChangeTypeStorage
)

// NewStateChange creates a new state change
func NewStateChange(changeType ChangeType, address Address, symbol string, amount *big.Int) *StateChange {
	return &StateChange{
		Type:    changeType,
		Address: address.String(),
		Symbol:  symbol,
		Amount:  amount,
	}
}

// NewStorageChange creates a new storage change
func NewStorageChange(address Address, key, value []byte) *StateChange {
	return &StateChange{
		Type:    ChangeTypeStorage,
		Address: address.String(),
		Key:     key,
		Value:   value,
	}
}
