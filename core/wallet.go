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
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/ripemd160"
)

// Допустимые префиксы в байтовом представлении
var validPrefixes = [][]byte{
	[]byte("GND"),
	[]byte("GN_"),
}

var accountID int

type Wallet struct {
	PrivateKey *secp256k1.PrivateKey
	Address    Address
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

	// Проверяем, не существует ли уже такой адрес
	var exists bool
	err = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM accounts WHERE address = $1)", address).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("ошибка проверки существующего адреса: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("адрес %s уже существует", address)
	}

	// Создаем аккаунт с адресом сразу
	err = tx.QueryRow(ctx,
		`INSERT INTO accounts (address) VALUES ($1) RETURNING id`,
		address,
	).Scan(&accountID)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания аккаунта: %w", err)
	}

	// Подготовка данных для БД
	privateKeyBytes := privKey.Serialize()
	publicKeyHex := hex.EncodeToString(pubKey)
	privateKeyHex := hex.EncodeToString(privateKeyBytes)
	now := time.Now()

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

	// Фиксируем транзакцию
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("ошибка фиксации транзакции: %w", err)
	}

	return &Wallet{
		PrivateKey: privKey,
		Address:    Address(address),
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
	return hex.EncodeToString(w.PrivateKey.Serialize())
}

func (w *Wallet) PublicKeyHex() string {
	return hex.EncodeToString(w.PrivateKey.PubKey().SerializeCompressed())
}

// LoadWallet загружает существующий кошелек из базы данных
func LoadWallet(pool *pgxpool.Pool) (*Wallet, error) {
	var (
		address       string
		privateKeyHex string
	)

	// Получаем последний активный кошелек
	err := pool.QueryRow(context.Background(), `
		SELECT w.address, w.private_key
		FROM wallets w
		JOIN accounts a ON w.account_id = a.id
		WHERE w.status = 'active'
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
