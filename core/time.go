// | KB @CerberRus00 - Nexus Invest Team
// core/time.go — время блокчейна (Москва).

package core

import (
	"time"
)

// Moscow — часовой пояс блокчейна (Москва).
var Moscow *time.Location

// genesisTimestamp — фиксированное время генезиса для блока (Москва). Заполняется в init() после Moscow.
var genesisTimestamp time.Time

func init() {
	var err error
	Moscow, err = time.LoadLocation("Europe/Moscow")
	if err != nil {
		Moscow = time.FixedZone("MSK", 3*3600)
	}
	genesisTimestamp = time.Date(2025, 6, 1, 0, 0, 0, 0, Moscow)
}

// BlockchainNow возвращает текущее время в часовом поясе блокчейна (Москва).
func BlockchainNow() time.Time {
	return time.Now().In(Moscow)
}
