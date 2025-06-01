package core

import (
	"crypto/rand" // Импорт для rand.Int
	"crypto/sha256"
	"encoding/hex"
	"github.com/btcsuite/btcutil/base58"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"golang.org/x/crypto/ripemd160"
	"math/big" // Добавляем импорт math/big
	"strings"
)

// Допустимые префиксы в байтовом представлении
var validPrefixes = [][]byte{
	[]byte("GND"),
	[]byte("GN_"),
}

type Wallet struct {
	PrivateKey *secp256k1.PrivateKey
	Address    Address
}

func NewWallet() (*Wallet, error) {
	privKey, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		return nil, err
	}

	// Генерация публичного ключа (сжатый формат)
	pubKey := privKey.PubKey().SerializeCompressed()

	// SHA-256 + RIPEMD-160
	shaHash := sha256.Sum256(pubKey)
	ripemdHasher := ripemd160.New()
	if _, err := ripemdHasher.Write(shaHash[:]); err != nil {
		return nil, err
	}
	pubKeyHash := ripemdHasher.Sum(nil)

	// Выбор префикса
	prefix, err := randomPrefix()
	if err != nil {
		return nil, err
	}

	// Формирование адреса
	checksum := checksum(pubKeyHash)
	fullPayload := append(pubKeyHash, checksum...)
	encoded := base58.Encode(fullPayload)

	address := prefix + encoded

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
