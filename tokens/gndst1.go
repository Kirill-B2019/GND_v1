// tokens/gndst1.go

package tokens

import (
	"math/big"
)

type GNDst1 struct {
	name        string
	symbol      string
	decimals    uint8
	totalSupply *big.Int
	balances    map[string]*big.Int            // address -> balance
	allowances  map[string]map[string]*big.Int // owner -> spender -> amount
	kycPassed   map[string]bool                // address -> status
	bridge      string                         // адрес моста
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

// --- ERC-20/TRC-20 методы ---
func (t *GNDst1) TotalSupply() *big.Int {
	return t.totalSupply
}

func (t *GNDst1) BalanceOf(account string) *big.Int {
	return t.balances[account]
}

func (t *GNDst1) Transfer(to string, amount *big.Int) bool {
	if t.BalanceOf(to).Cmp(amount) < 0 {
		return false
	}
	t._transfer("msg.sender", to, amount)
	return true
}

func (t *GNDst1) Allowance(owner, spender string) *big.Int {
	if _, ok := t.allowances[owner]; !ok {
		return big.NewInt(0)
	}
	if _, ok := t.allowances[owner][spender]; !ok {
		return big.NewInt(0)
	}
	return t.allowances[owner][spender]
}

func (t *GNDst1) Approve(spender string, amount *big.Int) bool {
	if t.allowances["msg.sender"] == nil {
		t.allowances["msg.sender"] = make(map[string]*big.Int)
	}
	t.allowances["msg.sender"][spender] = amount
	return true
}

func (t *GNDst1) TransferFrom(from, to string, amount *big.Int) bool {
	if t.Allowance(from, "msg.sender").Cmp(amount) < 0 {
		return false
	}
	t.allowances[from]["msg.sender"].Sub(t.allowances[from]["msg.sender"], amount)
	t._transfer(from, to, amount)
	return true
}

// --- Дополнительные функции GNDst1 ---
func (t *GNDst1) CrossChainTransfer(targetChain string, to string, amount *big.Int) bool {
	t._transfer("msg.sender", t.bridge, amount)
	return true
}

func (t *GNDst1) SetKycStatus(user string, status bool) {
	t.kycPassed[user] = status
}

func (t *GNDst1) IsKycPassed(user string) bool {
	return t.kycPassed[user]
}

func (t *GNDst1) ModuleCall(moduleId [32]byte, data []byte) ([]byte, error) {
	// stub
	return []byte("module call placeholder"), nil
}

func (t *GNDst1) Snapshot() uint256 {
	panic("not implemented")
}

func (t *GNDst1) GetSnapshotBalance(user string, snapshotId uint256) *big.Int {
	panic("not implemented")
}

// --- Внутренние функции ---
func (t *GNDst1) _transfer(from, to string, amount *big.Int) {
	if t.balances[from].Cmp(amount) < 0 {
		panic("insufficient balance")
	}
	t.balances[from].Sub(t.balances[from], amount)
	t.balances[to].Add(t.balances[to], amount)
}
