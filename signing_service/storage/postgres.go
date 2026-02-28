// KB @CerbeRus - Nexus Invest Team
package storage

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Postgres реализует WalletRepo через таблицу signer_wallets.
type Postgres struct {
	pool *pgxpool.Pool
}

// NewPostgres создаёт репозиторий, работающий с переданным пулом (таблица signer_wallets).
func NewPostgres(pool *pgxpool.Pool) *Postgres {
	return &Postgres{pool: pool}
}

// GetWallet возвращает кошелёк по id из signer_wallets.
func (p *Postgres) GetWallet(ctx context.Context, id uuid.UUID) (*Wallet, error) {
	const q = `
		SELECT id, account_id, public_key, encrypted_priv, disabled, created_at, updated_at
		FROM signer_wallets WHERE id = $1
	`
	var w Wallet
	err := p.pool.QueryRow(ctx, q, id).Scan(
		&w.ID, &w.AccountID, &w.PublicKey, &w.EncryptedPriv, &w.Disabled, &w.CreatedAt, &w.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get wallet: %w", err)
	}
	return &w, nil
}

// GetWalletByAccountID возвращает кошелёк по account_id (для подписи по адресу отправителя).
func (p *Postgres) GetWalletByAccountID(ctx context.Context, accountID int) (*Wallet, error) {
	const q = `
		SELECT id, account_id, public_key, encrypted_priv, disabled, created_at, updated_at
		FROM signer_wallets WHERE account_id = $1
	`
	var w Wallet
	err := p.pool.QueryRow(ctx, q, accountID).Scan(
		&w.ID, &w.AccountID, &w.PublicKey, &w.EncryptedPriv, &w.Disabled, &w.CreatedAt, &w.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get wallet by account_id: %w", err)
	}
	return &w, nil
}

// CreateWallet вставляет новый кошелёк в signer_wallets.
func (p *Postgres) CreateWallet(ctx context.Context, w *Wallet) error {
	const q = `
		INSERT INTO signer_wallets (id, account_id, public_key, encrypted_priv, disabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := p.pool.Exec(ctx, q,
		w.ID, w.AccountID, w.PublicKey, w.EncryptedPriv, w.Disabled, w.CreatedAt, w.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create wallet: %w", err)
	}
	return nil
}
