package consensus

import (
	"GND/core"
)

// Интерфейс для всех алгоритмов консенсуса
type Consensus interface {
	Start(bc *core.Blockchain, mempool *core.Mempool)
	Stop()
	Type() string
}
