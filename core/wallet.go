// | KB @CerbeRus - Nexus Invest Team
package core

import (
	"context"
	"crypto/rand" // Импорт для rand.Int
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big" // Добавляем импорт math/big
	"strings"
	"time"

	"github.com/btcsuite/btcutil/base58"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/ripemd160"
)

// Допустимые префиксы в байтовом представлении
var validPrefixes = [][]byte{
	[]byte("GND"),
	[]byte("GN_"),
}

var accountID int

// SignerWalletCreator — интерфейс для создания кошелька через signing_service (без хранения приватного ключа в БД).
type SignerWalletCreator interface {
	GenerateKeyForNewWallet() (pubKey []byte, encryptedPriv []byte, err error)
	StoreWallet(ctx context.Context, accountID int, pubKey, encryptedPriv []byte) (walletID uuid.UUID, err error)
}

type Wallet struct {
	PrivateKey     *secp256k1.PrivateKey
	Address        Address
	SignerWalletID *uuid.UUID // заполняется для кошельков, созданных через signing_service
}

func NewWallet(pool *pgxpool.Pool) (*Wallet, error) {
	ctx := context.Background()

	// Начинаем транзакцию
	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("ошибка начала транзакции: %w", err)
	}
	defer tx.Rollback(ctx)

	// Генерация ключей
	privKey, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		return nil, fmt.Errorf("ошибка генерации ключа: %w", err)
	}

	// Формирование адреса
	pubKey := privKey.PubKey().SerializeCompressed()
	shaHash := sha256.Sum256(pubKey)
	ripemdHasher := ripemd160.New()
	if _, err := ripemdHasher.Write(shaHash[:]); err != nil {
		return nil, err
	}
	pubKeyHash := ripemdHasher.Sum(nil)
	prefix, err := randomPrefix()
	if err != nil {
		return nil, err
	}
	checksum := Checksum(pubKeyHash)
	fullPayload := append(pubKeyHash, checksum...)
	encoded := base58.Encode(fullPayload)
	address := prefix + encoded

	// Валидация адреса
	if !ValidateAddress(address) {
		return nil, fmt.Errorf("сгенерированный адрес не прошел валидацию: %s", address)
	}

	// Проверяем, не существует ли уже такой адрес
	var exists bool
	err = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM accounts WHERE address = $1)", address).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("ошибка проверки существующего адреса: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("адрес %s уже существует", address)
	}

	// Проверяем существование токенов
	var tokenCount int
	err = tx.QueryRow(ctx, "SELECT COUNT(*) FROM tokens").Scan(&tokenCount)
	if err != nil {
		return nil, fmt.Errorf("ошибка проверки количества токенов: %w", err)
	}

	// Создаем аккаунт без заполнения balance (accounts.balance не начисляем)
	err = tx.QueryRow(ctx,
		`INSERT INTO accounts (address, nonce, created_at) VALUES ($1, $2, $3) RETURNING id`,
		address,
		0, // Начальный nonce
		time.Now(),
	).Scan(&accountID)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания аккаунта: %w", err)
	}

	// Инициализируем начальное состояние для всех токенов
	if tokenCount > 0 {
		_, err = tx.Exec(ctx, `
			INSERT INTO token_balances (token_id, address, balance)
			SELECT t.id, $1, '0'
			FROM tokens t
			WHERE NOT EXISTS (
				SELECT 1 FROM token_balances tb 
				WHERE tb.token_id = t.id AND tb.address = $1
			)`,
			address,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка инициализации балансов токенов: %w", err)
		}

		// Проверяем, что все токены были инициализированы
		var initializedCount int
		err = tx.QueryRow(ctx, `
			SELECT COUNT(*) FROM token_balances 
			WHERE address = $1`,
			address,
		).Scan(&initializedCount)
		if err != nil {
			return nil, fmt.Errorf("ошибка проверки инициализации балансов: %w", err)
		}
		if initializedCount != tokenCount {
			return nil, fmt.Errorf("не все токены были инициализированы: ожидалось %d, получено %d",
				tokenCount, initializedCount)
		}
	}

	// Подготовка данных для БД
	privateKeyBytes := privKey.Serialize()
	publicKeyHex := hex.EncodeToString(pubKey)
	privateKeyHex := hex.EncodeToString(privateKeyBytes)
	now := time.Now()

	// Валидация ключей
	if len(privateKeyBytes) != 32 {
		return nil, fmt.Errorf("некорректная длина приватного ключа: %d", len(privateKeyBytes))
	}
	if len(pubKey) != 33 {
		return nil, fmt.Errorf("некорректная длина публичного ключа: %d", len(pubKey))
	}

	// Создаем кошелек
	query := `
    INSERT INTO wallets (
        account_id,
        address,
        public_key,
        private_key,
        created_at,
        updated_at,
        status
    ) VALUES ($1, $2, $3, $4, $5, $6, $7)
    RETURNING id
`
	var walletID int
	err = tx.QueryRow(ctx, query,
		accountID,
		address,
		publicKeyHex,
		privateKeyHex,
		now,
		now,
		"active",
	).Scan(&walletID)

	if err != nil {
		return nil, fmt.Errorf("ошибка сохранения кошелька в БД: %w", err)
	}

	// Проверяем, что кошелек создан корректно
	var walletExists bool
	err = tx.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM wallets 
			WHERE id = $1 AND address = $2 AND status = 'active'
		)`,
		walletID, address,
	).Scan(&walletExists)
	if err != nil {
		return nil, fmt.Errorf("ошибка проверки создания кошелька: %w", err)
	}
	if !walletExists {
		return nil, fmt.Errorf("кошелек не был корректно создан")
	}

	// Фиксируем транзакцию
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("ошибка фиксации транзакции: %w", err)
	}

	return &Wallet{
		PrivateKey: privKey,
		Address:    Address(address),
	}, nil
}

// NewWalletWithSigner создаёт кошелёк через SignerWalletCreator: ключ хранится только в signer_wallets (зашифрован).
func NewWalletWithSigner(ctx context.Context, pool *pgxpool.Pool, creator SignerWalletCreator) (*Wallet, error) {
	pubKey, encryptedPriv, err := creator.GenerateKeyForNewWallet()
	if err != nil {
		return nil, fmt.Errorf("генерация ключа signer: %w", err)
	}
	address, err := AddressFromPublicKey(pubKey)
	if err != nil {
		return nil, fmt.Errorf("адрес из публичного ключа: %w", err)
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("начало транзакции: %w", err)
	}
	defer tx.Rollback(ctx)

	var exists bool
	if err := tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM accounts WHERE address = $1)", address).Scan(&exists); err != nil {
		return nil, fmt.Errorf("проверка адреса: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("адрес %s уже существует", address)
	}
	var tokenCount int
	if err := tx.QueryRow(ctx, "SELECT COUNT(*) FROM tokens").Scan(&tokenCount); err != nil {
		return nil, fmt.Errorf("проверка токенов: %w", err)
	}
	var newAccountID int
	if err := tx.QueryRow(ctx,
		`INSERT INTO accounts (address, nonce, created_at) VALUES ($1, $2, $3) RETURNING id`,
		address, 0, time.Now(),
	).Scan(&newAccountID); err != nil {
		return nil, fmt.Errorf("создание аккаунта: %w", err)
	}
	if tokenCount > 0 {
		if _, err := tx.Exec(ctx, `
			INSERT INTO token_balances (token_id, address, balance)
			SELECT t.id, $1, '0' FROM tokens t
			WHERE NOT EXISTS (SELECT 1 FROM token_balances tb WHERE tb.token_id = t.id AND tb.address = $1)`,
			address); err != nil {
			return nil, fmt.Errorf("инициализация балансов токенов: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("фиксация транзакции: %w", err)
	}
	signerWalletID, err := creator.StoreWallet(ctx, newAccountID, pubKey, encryptedPriv)
	if err != nil {
		return nil, fmt.Errorf("сохранение в signer_wallets: %w", err)
	}
	now := time.Now()
	publicKeyHex := hex.EncodeToString(pubKey)
	_, err = pool.Exec(ctx, `
		INSERT INTO wallets (account_id, address, public_key, private_key, signer_wallet_id, created_at, updated_at, status)
		VALUES ($1, $2, $3, NULL, $4, $5, $6, 'active')`,
		newAccountID, address, publicKeyHex, signerWalletID, now, now)
	if err != nil {
		return nil, fmt.Errorf("запись в wallets: %w", err)
	}
	return &Wallet{
		Address:        Address(address),
		SignerWalletID: &signerWalletID,
	}, nil
}

func randomPrefix() (string, error) {
	validPrefixes := []string{"GND", "GN_"}
	max := big.NewInt(int64(len(validPrefixes)))
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return validPrefixes[n.Int64()], nil
}

// AddressFromPublicKey строит GND-адрес из сжатого публичного ключа secp256k1 (33 байта).
// Используется при создании кошелька через signing_service.
func AddressFromPublicKey(pubKey []byte) (string, error) {
	if len(pubKey) != 33 {
		return "", fmt.Errorf("ожидается сжатый публичный ключ 33 байта, получено %d", len(pubKey))
	}
	shaHash := sha256.Sum256(pubKey)
	ripemdHasher := ripemd160.New()
	if _, err := ripemdHasher.Write(shaHash[:]); err != nil {
		return "", err
	}
	pubKeyHash := ripemdHasher.Sum(nil)
	prefix, err := randomPrefix()
	if err != nil {
		return "", err
	}
	checksum := Checksum(pubKeyHash)
	fullPayload := append(pubKeyHash, checksum...)
	encoded := base58.Encode(fullPayload)
	return prefix + encoded, nil
}

func ValidateAddress(address string) bool {
	// Проверяем префикс
	var prefixLen int
	switch {
	case strings.HasPrefix(address, "GNDct"):
		prefixLen = 5
	case strings.HasPrefix(address, "GND"):
		prefixLen = 3
	case strings.HasPrefix(address, "GN_"):
		prefixLen = 3
	default:
		return false
	}

	if len(address) <= prefixLen {
		return false
	}

	// Отделяем base58-часть
	encoded := address[prefixLen:]

	// Декодируем base58-часть
	decoded := base58.Decode(encoded)
	if len(decoded) != 24 { // 20 байт хеша + 4 байта checksum
		return false
	}

	// Проверяем контрольную сумму (Base58Check)
	payload := decoded[:20]
	checksumBytes := decoded[20:]
	return BytesEqual(Checksum(payload), checksumBytes)
}

// Безопасное сравнение байтовых срезов
func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (w *Wallet) PrivateKeyHex() string {
	if w.PrivateKey != nil {
		return hex.EncodeToString(w.PrivateKey.Serialize())
	}
	return ""
}

func (w *Wallet) PublicKeyHex() string {
	if w.PrivateKey != nil {
		return hex.EncodeToString(w.PrivateKey.PubKey().SerializeCompressed())
	}
	return ""
}

// LoadWallet загружает последний активный кошелёк с приватным ключом (для валидатора/майнера).
// Кошельки только с signer_wallet_id (private_key IS NULL) не выбираются.
func LoadWallet(pool *pgxpool.Pool) (*Wallet, error) {
	var (
		address       string
		privateKeyHex string
	)

	err := pool.QueryRow(context.Background(), `
		SELECT w.address, w.private_key
		FROM wallets w
		JOIN accounts a ON w.account_id = a.id
		WHERE w.status = 'active' AND w.private_key IS NOT NULL
		ORDER BY w.created_at DESC
		LIMIT 1
	`).Scan(&address, &privateKeyHex)

	if err != nil {
		return nil, fmt.Errorf("ошибка загрузки кошелька: %w", err)
	}

	// Декодируем приватный ключ
	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("ошибка декодирования приватного ключа: %w", err)
	}

	// Создаем объект приватного ключа
	privKey := secp256k1.PrivKeyFromBytes(privateKeyBytes)

	return &Wallet{
		PrivateKey: privKey,
		Address:    Address(address),
	}, nil
}
