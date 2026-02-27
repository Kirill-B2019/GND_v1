// | KB @CerbeRus - Nexus Invest Team
package types

import (
	"math/big"
)

// StateIface defines the interface for blockchain state
type StateIface interface {
	// GetBalance returns the balance for an address and symbol
	GetBalance(address Address, symbol string) *big.Int

	// AddBalance adds amount to the balance of address for symbol
	AddBalance(address Address, symbol string, amount *big.Int) error

	// SubBalance subtracts amount from the balance of address for symbol
	SubBalance(address Address, symbol string, amount *big.Int) error

	// SaveContract saves a contract to state
	SaveContract(contract *Contract) error

	// GetContract returns a contract by address
	GetContract(address string) (*Contract, error)
}

// Contract represents a smart contract
type Contract struct {
	Address     string
	Bytecode    []byte
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
}

// Execute executes a contract method
func (c *Contract) Execute(method string, args []interface{}) (interface{}, error) {
	// TODO: Implement contract execution
	return nil, nil
}
