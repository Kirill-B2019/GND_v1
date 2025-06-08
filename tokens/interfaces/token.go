// tokens/interfaces/token.go

package interfaces

import (
	"GND/core"
)

type Contract interface {
	Execute(method string, args []interface{}) (interface{}, error)
	Address() core.Address
	Bytecode() []byte
}
