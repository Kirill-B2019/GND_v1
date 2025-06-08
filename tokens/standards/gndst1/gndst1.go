// tokens/standards/gndst1/gndst1.go

package gndst1

import (
	"GND/tokens"
	"errors"
	"fmt"
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

type GNDst1 struct {
	name        string
	symbol      string
	decimals    uint8
	totalSupply *big.Int
	balances    map[string]*big.Int
	allowances  map[string]map[string]*big.Int
	kycPassed   map[string]bool
	bridge      string
}

func NewGNDst1(initialSupply *big.Int, bridgeAddress string) *GNDst1 {
	return &GNDst1{
		name:        "Ganimed Token",
		symbol:      "GND",
		decimals:    18,
		totalSupply: new(big.Int).Set(initialSupply),
		balances: map[string]*big.Int{
			"owner": new(big.Int).Set(initialSupply),
		},
		allowances: make(map[string]map[string]*big.Int),
		kycPassed:  make(map[string]bool),
		bridge:     bridgeAddress,
	}
}

// --- Базовые методы ---
func (t *GNDst1) Name() string          { return t.name }
func (t *GNDst1) Symbol() string        { return t.symbol }
func (t *GNDst1) Decimals() uint8       { return t.decimals }
func (t *GNDst1) TotalSupply() *big.Int { return t.totalSupply }

func (t *GNDst1) BalanceOf(account string) *big.Int {
	if balance, ok := t.balances[account]; ok {
		return balance
	}
	return big.NewInt(0)
}

// --- ERC-20 совместимые методы ---
func (t *GNDst1) Transfer(from string, to string, amount *big.Int) bool {
	if amount.Cmp(big.NewInt(0)) <= 0 {
		return false
	}
	if senderBalance, ok := t.balances[from]; !ok || senderBalance.Cmp(amount) < 0 {
		return false
	}

	t.balances[from] = new(big.Int).Sub(t.balances[from], amount)
	t.balances[to] = new(big.Int).Add(t.balances[to], amount)
	return true
}

func (t *GNDst1) Allowance(owner string, spender string) *big.Int {
	if _, ok := t.allowances[owner]; !ok {
		return big.NewInt(0)
	}
	if _, ok := t.allowances[owner][spender]; !ok {
		return big.NewInt(0)
	}
	return t.allowances[owner][spender]
}

func (t *GNDst1) Approve(spender string, amount *big.Int) bool {
	if amount.Cmp(big.NewInt(0)) < 0 {
		return false
	}
	if t.allowances[spender] == nil {
		t.allowances[spender] = make(map[string]*big.Int)
	}
	t.allowances[spender]["msg.sender"] = amount
	return true
}

func (t *GNDst1) TransferFrom(from string, to string, amount *big.Int) bool {
	if amount.Cmp(big.NewInt(0)) <= 0 {
		return false
	}

	allowance := t.Allowance(from, "msg.sender")
	if allowance.Cmp(amount) < 0 {
		return false
	}

	t.allowances["msg.sender"][from] = new(big.Int).Sub(allowance, amount)
	t.Transfer(from, to, amount)
	return true
}

// --- Расширенные методы GNDst-1 ---
func (t *GNDst1) CrossChainTransfer(targetChain string, to string, amount *big.Int) bool {
	if amount.Cmp(big.NewInt(0)) <= 0 {
		return false
	}
	if senderBalance := t.BalanceOf("msg.sender"); senderBalance.Cmp(amount) < 0 {
		return false
	}

	t.Transfer("msg.sender", t.bridge, amount)
	return true
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
