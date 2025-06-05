package core

import (
	"context"
	"crypto/rand" // Импорт для rand.Int
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcutil/base58"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/ripemd160"
	"math/big" // Добавляем импорт math/big
	"strings"
	"time"
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

	//выбор ID для кошелька
	err := pool.QueryRow(
		context.Background(),
		`INSERT INTO accounts DEFAULT VALUES RETURNING id`,
	).Scan(&accountID)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания аккаунта: %w", err)
	}

	// Генерация ключей
	privKey, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		return nil, fmt.Errorf("ошибка генерации ключа: %w", err)
	}

	// Формирование адреса (как в оригинале)
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
	checksum := checksum(pubKeyHash)
	fullPayload := append(pubKeyHash, checksum...)
	encoded := base58.Encode(fullPayload)
	address := prefix + encoded

	// Подготовка данных для БД
	privateKeyBytes := privKey.Serialize()
	publicKeyHex := hex.EncodeToString(pubKey)
	privateKeyHex := hex.EncodeToString(privateKeyBytes) // Шифруйте на практике!
	now := time.Now()

	// SQL-запрос
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
	err = pool.QueryRow(context.Background(), query,
		accountID,
		address, // <-- добавлен адрес кошелька
		publicKeyHex,
		privateKeyHex,
		now,
		now,
		"active",
	).Scan(&walletID)

	if err != nil {
		return nil, fmt.Errorf("ошибка сохранения в БД: %w", err)
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

func checksum(payload []byte) []byte {
	first := sha256.Sum256(payload)
	second := sha256.Sum256(first[:])
	return second[:4]
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
	return bytesEqual(checksum(payload), checksumBytes)
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
