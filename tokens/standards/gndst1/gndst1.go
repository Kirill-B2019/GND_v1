// tokens/standards/gndst1/gndst1.go

package gndst1

import (
	"GND/core"
	"GND/tokens"
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"math/big"
)

var Tokens = map[string]*tokens.TokenInfo{}

type TokenMeta struct {
	Address     string
	Owner       string
	Standard    string
	Name        string
	Symbol      string
	Decimals    uint8
	Description string
}

// GNDst1 реализует стандарт GNDST1 для токенов
type GNDst1 struct {
	address     core.Address
	name        string
	symbol      string
	decimals    uint8
	totalSupply *big.Int
	balances    map[core.Address]*big.Int
	allowances  map[string]*big.Int // key format: "owner:spender"
	pool        *pgxpool.Pool
	kycPassed   map[string]bool
	bridge      string
}

func NewGNDst1(
	address core.Address,
	name string,
	symbol string,
	decimals uint8,
	totalSupply *big.Int,
	pool *pgxpool.Pool,
) *GNDst1 {
	return &GNDst1{
		address:     address,
		name:        name,
		symbol:      symbol,
		decimals:    decimals,
		totalSupply: totalSupply,
		balances:    make(map[core.Address]*big.Int),
		allowances:  make(map[string]*big.Int),
		pool:        pool,
		kycPassed:   make(map[string]bool),
	}
}

// --- Базовые методы ---
func (t *GNDst1) GetAddress() core.Address { return t.address }
func (t *GNDst1) GetName() string          { return t.name }
func (t *GNDst1) GetSymbol() string        { return t.symbol }
func (t *GNDst1) GetDecimals() uint8       { return t.decimals }
func (t *GNDst1) GetTotalSupply() *big.Int { return t.totalSupply }

// GetBalance возвращает баланс токенов для адреса
func (t *GNDst1) GetBalance(ctx context.Context, address core.Address) (*big.Int, error) {
	balance, exists := t.balances[address]
	if !exists {
		return big.NewInt(0), nil
	}
	return balance, nil
}

// --- ERC-20 совместимые методы ---
func (t *GNDst1) Transfer(ctx context.Context, from core.Address, to core.Address, amount *big.Int) error {
	if amount.Sign() <= 0 {
		return errors.New("amount must be positive")
	}

	// Проверяем баланс
	balance := t.balances[from]
	if balance == nil || balance.Cmp(amount) < 0 {
		return errors.New("insufficient balance")
	}

	// Выполняем перевод
	balance.Sub(balance, amount)
	if _, exists := t.balances[to]; !exists {
		t.balances[to] = new(big.Int)
	}
	t.balances[to].Add(t.balances[to], amount)

	// TODO: Сохранить в БД
	return nil
}

// Allowance возвращает количество токенов, которое spender может потратить от имени owner
func (t *GNDst1) Allowance(ctx context.Context, owner core.Address, spender core.Address) (*big.Int, error) {
	key := fmt.Sprintf("%s:%s", string(owner), string(spender))
	allowance, exists := t.allowances[key]
	if !exists {
		return big.NewInt(0), nil
	}
	return allowance, nil
}

// Approve устанавливает количество токенов, которое spender может потратить от имени owner
func (t *GNDst1) Approve(ctx context.Context, owner core.Address, spender core.Address, amount *big.Int) error {
	if amount.Sign() < 0 {
		return errors.New("amount cannot be negative")
	}

	key := fmt.Sprintf("%s:%s", string(owner), string(spender))
	t.allowances[key] = amount

	// TODO: Сохранить в БД
	return nil
}

// TransferFrom переводит amount токенов от from к to, используя разрешение
func (t *GNDst1) TransferFrom(ctx context.Context, from core.Address, to core.Address, amount *big.Int) error {
	if amount.Sign() <= 0 {
		return errors.New("amount must be positive")
	}

	// Проверяем разрешение
	allowance, err := t.Allowance(ctx, from, t.address)
	if err != nil {
		return err
	}
	if allowance.Cmp(amount) < 0 {
		return errors.New("insufficient allowance")
	}

	// Проверяем баланс
	balance := t.balances[from]
	if balance == nil || balance.Cmp(amount) < 0 {
		return errors.New("insufficient balance")
	}

	// Выполняем перевод
	balance.Sub(balance, amount)
	if _, exists := t.balances[to]; !exists {
		t.balances[to] = new(big.Int)
	}
	t.balances[to].Add(t.balances[to], amount)

	// Уменьшаем разрешение
	key := fmt.Sprintf("%s:%s", string(from), string(t.address))
	t.allowances[key].Sub(t.allowances[key], amount)

	// TODO: Сохранить в БД
	return nil
}

// --- Расширенные методы GNDst-1 ---
func (t *GNDst1) CrossChainTransfer(ctx context.Context, targetChain string, to string, amount *big.Int) error {
	if amount.Sign() <= 0 {
		return errors.New("amount must be positive")
	}

	// Проверяем баланс
	senderBalance, err := t.GetBalance(ctx, t.address)
	if err != nil {
		return fmt.Errorf("failed to check balance: %v", err)
	}
	if senderBalance.Cmp(amount) < 0 {
		return errors.New("insufficient balance")
	}

	// Выполняем перевод
	if err := t.Transfer(ctx, t.address, t.address, amount); err != nil {
		return fmt.Errorf("failed to transfer: %v", err)
	}

	return nil
}

func (t *GNDst1) SetKycStatus(user string, status bool) {
	t.kycPassed[user] = status
}

func (t *GNDst1) IsKycPassed(user string) bool {
	status, exists := t.kycPassed[user]
	return exists && status
}

// --- Метаданные ---
func (t *GNDst1) Meta() TokenMeta {
	return TokenMeta{
		Address:     "", // TODO: Добавить адрес токена при необходимости
		Owner:       "",
		Standard:    "gndst1",
		Name:        t.name,
		Symbol:      t.symbol,
		Decimals:    t.decimals,
		Description: "Ganimed Multi-standard Token",
	}
}

// --- CustomMethod для поддержки произвольного вызова ---
func (t *GNDst1) CustomMethod(method string, args ...interface{}) (interface{}, error) {
	switch method {
	case "crossChainTransfer":
		if len(args) < 3 {
			return nil, errors.New("недостаточно аргументов для перекрестного переноса по цепочке")
		}
		// TODO crossChainTransfer Перекрестный перенос
		return nil, fmt.Errorf("Перекрестный перенос по цепочке реализован не полностью") // <-- заглушка
	case "setKycStatus":
		if len(args) < 2 {
			return nil, errors.New("недостаточно аргументов для setKycStatus")
		}
		user, _ := args[0].(string)
		status, _ := args[1].(bool)
		t.SetKycStatus(user, status)
		return nil, nil
	default:
		return nil, fmt.Errorf("неизвестный пользовательский метод: %s", method)
	}
}

// BridgeTransfer выполняет перевод токенов через мост
func (t *GNDst1) BridgeTransfer(ctx context.Context, amount *big.Int) error {
	if amount.Sign() <= 0 {
		return errors.New("amount must be positive")
	}

	// Проверяем баланс
	senderBalance, err := t.GetBalance(ctx, t.address)
	if err != nil {
		return fmt.Errorf("failed to check balance: %v", err)
	}
	if senderBalance.Cmp(amount) < 0 {
		return errors.New("insufficient balance")
	}

	// Выполняем перевод
	if err := t.Transfer(ctx, t.address, t.address, amount); err != nil {
		return fmt.Errorf("failed to transfer: %v", err)
	}

	return nil
}
