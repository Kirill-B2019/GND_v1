// tokens/abi.go

package tokens

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"strings"
)

// ABI вручную определенный для IGNDst1
var GNDst1ABI abi.ABI

func init() {
	GNDst1ABI, _ = abi.JSON(strings.NewReader(`[
		{"constant":true,"inputs":[],"name":"totalSupply","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},
		{"constant":true,"inputs":[{"name":"account","type":"address"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},
		{"constant":false,"inputs":[{"name":"to","type":"address"},{"name":"amount","type":"uint256"}],"name":"transfer","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},
		{"constant":true,"inputs":[{"name":"owner","type":"address"},{"name":"spender","type":"address"}],"name":"allowance","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},
		{"constant":false,"inputs":[{"name":"spender","type":"address"},{"name":"amount","type":"uint256"}],"name":"approve","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},
		{"constant":false,"inputs":[{"name":"from","type":"address"},{"name":"to","type":"address"},{"name":"amount","type":"uint256"}],"name":"transferFrom","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},
		{"constant":false,"inputs":[{"name":"targetChain","type":"string"},{"name":"to","type":"address"},{"name":"amount","type":"uint256"}],"name":"crossChainTransfer","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},
		{"constant":false,"inputs":[{"name":"user","type":"address"},{"name":"status","type":"bool"}],"name":"setKycStatus","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},
		{"constant":true,"inputs":[{"name":"user","type":"address"}],"name":"isKycPassed","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"},
		{"constant":false,"inputs":[{"name":"moduleId","type":"bytes32"},{"name":"data","type":"bytes"}],"name":"moduleCall","outputs":[{"name":"","type":"bytes"}],"payable":false,"stateMutability":"nonpayable","type":"function"},
		{"constant":false,"inputs":[],"name":"snapshot","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"nonpayable","type":"function"},
		{"constant":true,"inputs":[{"name":"user","type":"address"},{"name":"snapshotId","type":"uint256"}],"name":"getSnapshotBalance","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"}
	]`))
}
