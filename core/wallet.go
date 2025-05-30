package core

import (
	"crypto/rand" // Импорт для rand.Int
	"crypto/sha256"
	"encoding/hex"
	"github.com/btcsuite/btcutil/base58"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"golang.org/x/crypto/ripemd160"
	"math/big" // Добавляем импорт math/big
)

// Допустимые префиксы в байтовом представлении
var validPrefixes = [][]byte{
	[]byte("GND"),
	[]byte("GND_"),
	[]byte("GN"),
	[]byte("GN_"),
}

type Address string

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
	payload := append(prefix, pubKeyHash...)
	checksum := checksum(payload)
	fullPayload := append(payload, checksum...)
	address := base58.Encode(fullPayload)

	return &Wallet{
		PrivateKey: privKey,
		Address:    Address(address),
	}, nil
}

func randomPrefix() ([]byte, error) {
	// Используем math/big для работы с большими числами
	max := big.NewInt(int64(len(validPrefixes)))
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return nil, err
	}
	return validPrefixes[n.Int64()], nil
}

func checksum(payload []byte) []byte {
	first := sha256.Sum256(payload)
	second := sha256.Sum256(first[:])
	return second[:4]
}

func ValidateAddress(address string) bool {
	decoded := base58.Decode(address)

	// Минимальная длина: префикс(3-4) + хеш(20) + checksum(4)
	if len(decoded) < 8 || len(decoded) > 28 {
		return false
	}

	// Поиск совпадения префикса
	var prefix []byte
	for _, p := range validPrefixes {
		if len(decoded) >= len(p) && bytesEqual(decoded[:len(p)], p) {
			prefix = p
			break
		}
	}
	if prefix == nil {
		return false
	}

	// Проверка контрольной суммы
	payload := decoded[:len(decoded)-4]
	checksumBytes := decoded[len(decoded)-4:]
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
