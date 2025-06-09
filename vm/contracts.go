// vm/contracts.go

package vm

import (
	"GND/core" // импортируем core для использования Address
	"GND/vm/compiler"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
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
	standard string
	balances map[core.Address]*big.Int
	pool     *pgxpool.Pool
}

var ContractRegistry = make(ContractRegistryType)

type ContractRegistryType map[core.Address]Contract

// CompileResult представляет результат компиляции контракта
type CompileResult struct {
	Bytecode string
	ABI      string
	Errors   []string
}

// NewTokenContract создает новый токен с начальным балансом владельца
func NewTokenContract(
	address core.Address,
	bytecode []byte,
	owner core.Address,
	name, symbol string,
	decimals uint8,
	totalSupply *big.Int,
	pool *pgxpool.Pool,
) (*TokenContract, error) {

	token := &TokenContract{
		address:  address,
		bytecode: bytecode,
		owner:    owner,
		name:     name,
		symbol:   symbol,
		decimals: decimals,
		standard: "gndst1", // Устанавливаем стандарт по умолчанию
		balances: make(map[core.Address]*big.Int),
		pool:     pool,
	}

	// Начальный баланс владельца
	token.balances[owner] = new(big.Int).Set(totalSupply)

	// Сохраняем в БД
	if err := token.SaveToDB(context.Background()); err != nil {
		return nil, err
	}

	RegisterContract(address, token)
	return token, nil
}

func (c *TokenContract) Address() core.Address {
	return c.address
}

func (c *TokenContract) Bytecode() Bytecode {
	return c.bytecode
}

func (c *TokenContract) Execute(method string, args []interface{}) (interface{}, error) {
	if len(args) == 0 {
		return nil, errors.New("отсутствуют аргументы")
	}

	switch method {
	case "transfer":
		if len(args) != 2 {
			return nil, errors.New("метод transfer требует 2 аргумента: to и amount")
		}
		return c.handleTransfer(args)
	case "balanceOf":
		if len(args) != 1 {
			return nil, errors.New("метод balanceOf требует 1 аргумент: address")
		}
		return c.handleBalanceOf(args)
	default:
		return nil, fmt.Errorf("неизвестный метод: %s", method)
	}
}

func (c *TokenContract) handleTransfer(args []interface{}) (interface{}, error) {
	to, ok := args[0].(core.Address)
	if !ok {
		return nil, errors.New("некорректный тип адреса получателя")
	}

	amount, ok := args[1].(*big.Int)
	if !ok {
		return nil, errors.New("некорректный тип суммы перевода")
	}

	if amount.Sign() <= 0 {
		return nil, errors.New("сумма перевода должна быть положительной")
	}

	// Проверяем баланс отправителя
	senderBalance := c.balances[c.owner]
	if senderBalance.Cmp(amount) < 0 {
		return nil, errors.New("недостаточно средств для перевода")
	}

	// Выполняем перевод
	senderBalance.Sub(senderBalance, amount)

	// Инициализируем баланс получателя, если его нет
	if _, exists := c.balances[to]; !exists {
		c.balances[to] = new(big.Int)
	}
	c.balances[to].Add(c.balances[to], amount)

	// Обновляем балансы в БД
	ctx := context.Background()
	if err := c.UpdateTokenBalanceInDB(ctx, c.pool, c.owner, senderBalance.Uint64()); err != nil {
		return nil, fmt.Errorf("ошибка обновления баланса отправителя: %v", err)
	}
	if err := c.UpdateTokenBalanceInDB(ctx, c.pool, to, c.balances[to].Uint64()); err != nil {
		return nil, fmt.Errorf("ошибка обновления баланса получателя: %v", err)
	}

	return true, nil
}

func (c *TokenContract) handleBalanceOf(args []interface{}) (interface{}, error) {
	address, ok := args[0].(core.Address)
	if !ok {
		return nil, errors.New("некорректный тип адреса")
	}

	balance, exists := c.balances[address]
	if !exists {
		return new(big.Int), nil
	}

	return balance, nil
}

// SaveToDB сохраняет контракт в базу данных
func (c *TokenContract) SaveToDB(ctx context.Context) error {
	tx, err := c.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("ошибка начала транзакции: %v", err)
	}
	defer tx.Rollback(ctx)

	// Сохраняем информацию о контракте
	_, err = tx.Exec(ctx, `
		INSERT INTO contracts (
			address, owner, code, abi, created_at, type
		) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (address) DO NOTHING`,
		c.address, c.owner, c.bytecode, nil, time.Now(), c.standard)
	if err != nil {
		return fmt.Errorf("ошибка сохранения контракта: %v", err)
	}

	// Сохраняем информацию о токене
	_, err = tx.Exec(ctx, `
		INSERT INTO tokens (
			contract_id, standard, symbol, name, decimals, total_supply
		) VALUES (
			(SELECT id FROM contracts WHERE address = $1),
			$2, $3, $4, $5, $6
		) ON CONFLICT (symbol) DO NOTHING`,
		c.address, c.standard, c.symbol, c.name, c.decimals, c.balances[c.owner].String())
	if err != nil {
		return fmt.Errorf("ошибка сохранения токена: %v", err)
	}

	return tx.Commit(ctx)
}

// UpdateTokenBalanceInDB обновляет баланс токена в базе данных
func (c *TokenContract) UpdateTokenBalanceInDB(ctx context.Context, pool *pgxpool.Pool, address core.Address, amount uint64) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("ошибка начала транзакции: %v", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO token_balances (
			token_id, address, balance
		) VALUES ((SELECT id FROM tokens WHERE symbol = $1), $2, $3)
		ON CONFLICT (token_id, address) DO UPDATE SET balance = $3`,
		c.symbol, address, amount)
	if err != nil {
		return fmt.Errorf("ошибка обновления баланса: %v", err)
	}

	return tx.Commit(ctx)
}

// Реестр контрактов (пример реализации)
var contractRegistry = make(map[core.Address]Contract)

func registerContract(addr core.Address, c Contract) {
	contractRegistry[addr] = c
}

func (cr ContractRegistryType) Register(addr core.Address, contract Contract) {
	cr[addr] = contract
}

func (cr ContractRegistryType) Get(addr core.Address) (Contract, bool) {
	c, ok := cr[addr]
	return c, ok
}

func generateBytecode(name, symbol string, decimals uint8, totalSupply *big.Int) ([]byte, error) {
	source := fmt.Sprintf(`
	pragma solidity ^0.8.0;

	contract %s {
		string public name = "%s";
		string public symbol = "%s";
		uint8 public decimals = %d;
		uint256 public totalSupply = %s;
		mapping(address => uint) public balanceOf;

		constructor() {
			balanceOf[msg.sender] = totalSupply;
		}

		function transfer(address to, uint amount) external {
			require(balanceOf[msg.sender] >= amount);
			balanceOf[msg.sender] -= amount;
			balanceOf[to] += amount;
		}
	}`, symbol, name, symbol, decimals, totalSupply.String())

	// Создаем компилятор с путем к solc
	solc := compiler.DefaultSolidityCompiler{
		SolcPath: "solc", // или укажите полный путь к solc
	}

	// Создаем метаданные контракта
	metadata := compiler.ContractMetadata{
		Name:        name,
		Standard:    "gndst1",
		Version:     "1.0.0",
		Compiler:    "solc",
		Description: fmt.Sprintf("Token %s (%s)", name, symbol),
		Params: map[string]interface{}{
			"symbol": symbol,
		},
	}

	// Компилируем контракт
	result, err := solc.Compile([]byte(source), metadata)
	if err != nil {
		return nil, fmt.Errorf("ошибка компиляции: %v", err)
	}

	// Декодируем байткод
	bytecode, err := hex.DecodeString(result.Bytecode)
	if err != nil {
		return nil, fmt.Errorf("ошибка декодирования байткода: %v", err)
	}

	return bytecode, nil
}
