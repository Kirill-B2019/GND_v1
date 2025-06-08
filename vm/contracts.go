// vm/contracts.go

package vm

import (
	"GND/core" // импортируем core для использования Address
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

// Определяем тип Bytecode как псевдоним для []byte
type Bytecode = []byte

func RegisterContract(addr core.Address, contract Contract) {
	ContractRegistry[addr] = contract
}

func CallContract(addr core.Address) (Contract, bool) {
	c, ok := ContractRegistry[addr]
	return c, ok
}

// Contract представляет базовый интерфейс смарт-контракта
type Contract interface {
	Execute(method string, args []interface{}) (interface{}, error)
	Address() core.Address
	Bytecode() Bytecode
}

// TokenContract реализация ERC20-подобного токена
type TokenContract struct {
	address  core.Address
	bytecode Bytecode
	owner    core.Address
	name     string
	symbol   string
	decimals uint8
	balances map[core.Address]uint64
}

type ContractMeta struct {
	Name        string
	Symbol      string
	Standard    string
	Owner       core.Address
	Description string
	Version     string
	Compiler    string
	Params      map[string]string
	MetadataCID string
	SourceCode  string
	Address     string
	Bytecode    string
}

func NewTokenContract(address core.Address, bytecode Bytecode, owner core.Address, name, symbol string, decimals uint8) *TokenContract {
	return &TokenContract{
		address:  address,
		bytecode: bytecode,
		owner:    owner,
		name:     name,
		symbol:   symbol,
		decimals: decimals,
		balances: make(map[core.Address]uint64),
	}
}

func (c *TokenContract) Address() core.Address {
	return c.address
}

func (c *TokenContract) Bytecode() Bytecode {
	return c.bytecode
}

func (c *TokenContract) Execute(method string, args []interface{}) (interface{}, error) {
	switch method {
	case "transfer":
		return c.handleTransfer(args)
	case "balanceOf":
		return c.handleBalanceOf(args)
	default:
		return nil, fmt.Errorf("неизвестный метод: %s", method)
	}
}

func (c *TokenContract) handleTransfer(args []interface{}) (interface{}, error) {
	// TODO: Реализуйте обработку аргументов
	return true, nil
}

func (c *TokenContract) handleBalanceOf(args []interface{}) (interface{}, error) {
	// TODO: Реализуйте обработку аргументов
	return uint64(100), nil
}

// DeployContract создает новый экземпляр контракта (упрощенная версия)
func DeployContract(
	from core.Address,
	bytecode Bytecode,
	meta ContractMeta,
	gasLimit uint64,
	gasPrice uint64,
	nonce uint64,
	signature string,
) (core.Address, error) {

	if len(bytecode) < 20 {
		return core.Address(""), errors.New("неверный байткод")
	}

	address := core.Address(hex.EncodeToString(bytecode[:20]))
	contract := NewTokenContract(
		address,
		bytecode,
		from,
		meta.Name,
		meta.Symbol,
		18,
	)

	RegisterContract(address, contract)

	return address, nil
}

// Реестр контрактов (пример реализации)
var contractRegistry = make(map[core.Address]Contract)

func registerContract(addr core.Address, c Contract) {
	contractRegistry[addr] = c
}

func (c *TokenContract) SaveToDB(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO contracts (
			address, owner, code, abi, created_at, type
		) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (address) DO NOTHING`,
		c.address, c.owner, c.bytecode, nil, time.Now(), "erc20")

	return err
}

func (c *TokenContract) SaveTokenInfoToDB(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO tokens (
			contract_id, standard, symbol, name, decimals, total_supply
		) VALUES ($1, $2, $3, $4, $5, $6)`,
		c.address, "erc20", c.symbol, c.name, c.decimals, "1000000")

	return err
}

func (c *TokenContract) UpdateTokenBalanceInDB(ctx context.Context, pool *pgxpool.Pool, address core.Address, amount uint64) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO token_balances (
			token_id, address, balance
		) VALUES ((SELECT id FROM tokens WHERE symbol = $1), $2, $3)
		ON CONFLICT (token_id, address) DO UPDATE SET balance = $3`,
		c.symbol, address, amount)

	return err
}
