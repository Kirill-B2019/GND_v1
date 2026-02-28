// KB @CerbeRus - Nexus Invest Team
package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"GND/signing_service/crypto"
	"GND/signing_service/storage"
)

// WalletRepo — интерфейс хранения кошельков в signer_wallets.
type WalletRepo interface {
	GetWallet(ctx context.Context, id uuid.UUID) (*storage.Wallet, error)
	GetWalletByAccountID(ctx context.Context, accountID int) (*storage.Wallet, error)
	CreateWallet(ctx context.Context, w *storage.Wallet) error
}

// SignerService — сервис подписи и создания кошельков (встроенный в GND).
type SignerService struct {
	repo      WalletRepo
	masterKey []byte
}

// NewSignerService создаёт SignerService.
func NewSignerService(repo WalletRepo, masterKey []byte) *SignerService {
	return &SignerService{repo: repo, masterKey: masterKey}
}

// SignDigest загружает кошелёк, расшифровывает ключ, подписывает digest, обнуляет буфер ключа.
func (s *SignerService) SignDigest(ctx context.Context, walletID uuid.UUID, digest []byte) ([]byte, error) {
	w, err := s.repo.GetWallet(ctx, walletID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrWalletNotFound
		}
		return nil, err
	}
	if w.Disabled {
		return nil, ErrWalletDisabled
	}

	plain, err := crypto.DecryptPrivKey(w.EncryptedPriv, s.masterKey)
	if err != nil {
		return nil, err
	}
	defer crypto.ZeroBytes(plain)

	priv, err := crypto.PrivKeyFromBytes(plain)
	if err != nil {
		return nil, err
	}
	return crypto.SignDigest(priv, digest)
}

// GetPublicKey возвращает байты публичного ключа кошелька.
func (s *SignerService) GetPublicKey(ctx context.Context, walletID uuid.UUID) ([]byte, error) {
	w, err := s.repo.GetWallet(ctx, walletID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrWalletNotFound
		}
		return nil, err
	}
	return w.PublicKey, nil
}

// CreateWalletResult — результат CreateWallet.
type CreateWalletResult struct {
	WalletID  uuid.UUID
	PublicKey []byte
	CreatedAt time.Time
}

// CreateWallet генерирует новый ключ secp256k1, шифрует приватный ключ, сохраняет в signer_wallets по account_id.
func (s *SignerService) CreateWallet(ctx context.Context, accountID int) (*CreateWalletResult, error) {
	priv, err := crypto.NewSecp256k1Key()
	if err != nil {
		return nil, err
	}
	defer crypto.ZeroBytes(crypto.PrivKeyToBytes(priv))

	pub := crypto.PublicKeyBytes(priv)
	encrypted, err := crypto.EncryptPrivKey(crypto.PrivKeyToBytes(priv), s.masterKey)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	w := &storage.Wallet{
		ID:            uuid.Must(uuid.NewV7()),
		AccountID:     accountID,
		PublicKey:     pub,
		EncryptedPriv: encrypted,
		Disabled:      false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := s.repo.CreateWallet(ctx, w); err != nil {
		return nil, err
	}
	return &CreateWalletResult{
		WalletID:  w.ID,
		PublicKey: w.PublicKey,
		CreatedAt: w.CreatedAt,
	}, nil
}

// GenerateKeyForNewWallet генерирует ключевую пару и возвращает публичный ключ и зашифрованный приватный (без записи в БД).
// Используется для создания кошелька: вызывающий получает адрес из pubKey, создаёт account, затем вызывает StoreWallet.
func (s *SignerService) GenerateKeyForNewWallet() (pubKey []byte, encryptedPriv []byte, err error) {
	priv, err := crypto.NewSecp256k1Key()
	if err != nil {
		return nil, nil, err
	}
	defer crypto.ZeroBytes(crypto.PrivKeyToBytes(priv))
	pub := crypto.PublicKeyBytes(priv)
	encrypted, err := crypto.EncryptPrivKey(crypto.PrivKeyToBytes(priv), s.masterKey)
	if err != nil {
		return nil, nil, err
	}
	return pub, encrypted, nil
}

// StoreWallet сохраняет кошелёк в signer_wallets по account_id (после создания account по адресу из pubKey).
func (s *SignerService) StoreWallet(ctx context.Context, accountID int, pubKey, encryptedPriv []byte) (walletID uuid.UUID, err error) {
	now := time.Now().UTC()
	w := &storage.Wallet{
		ID:            uuid.Must(uuid.NewV7()),
		AccountID:     accountID,
		PublicKey:     pubKey,
		EncryptedPriv: encryptedPriv,
		Disabled:      false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := s.repo.CreateWallet(ctx, w); err != nil {
		return uuid.Nil, err
	}
	return w.ID, nil
}
