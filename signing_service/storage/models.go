// KB @CerbeRus - Nexus Invest Team
package storage

import (
	"time"

	"github.com/google/uuid"
)

// Wallet — строка таблицы signer_wallets (кастодиальные ключи).
type Wallet struct {
	ID            uuid.UUID
	AccountID     int
	PublicKey     []byte
	EncryptedPriv []byte
	Disabled      bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
