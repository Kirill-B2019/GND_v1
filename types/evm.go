package types

import (
	"math/big"
)

// EVMInterface определяет интерфейс для виртуальной машины
type EVMInterface interface {
	// DeployContract деплоит новый контракт
	DeployContract(
		from string,
		bytecode []byte,
		meta ContractMeta,
		gasLimit uint64,
		gasPrice uint64,
		nonce uint64,
		signature string,
		totalSupply *big.Int,
	) (string, error)

	// CallContract выполняет вызов контракта
	CallContract(
		from string,
		to string,
		data []byte,
		gasLimit uint64,
		gasPrice uint64,
		value uint64,
		signature string,
	) ([]byte, error)

	// GetBalance возвращает баланс для адреса
	GetBalance(address string) (*big.Int, error)
}

// ContractMeta содержит метаданные контракта
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

// ContractInterface определяет интерфейс для контракта
type ContractInterface interface {
	// Execute выполняет метод контракта
	Execute(method string, args []interface{}) (interface{}, error)

	// Address возвращает адрес контракта
	Address() string

	// Bytecode возвращает байткод контракта
	Bytecode() []byte
}
