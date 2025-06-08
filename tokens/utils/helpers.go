// tokens/utils/helpers.go

package utils

import "math/big"

func FormatBalance(balance *big.Int) string {
	return balance.String()
}
