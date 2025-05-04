package core

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/btcsuite/btcutil/base58"
	"golang.org/x/crypto/ripemd160"
)

// Wallet хранит приватный ключ и адрес
type Wallet struct {
	PrivateKey *ecdsa.PrivateKey
	Address    string
}

// NewWallet генерирует новый кошелек с адресом, начинающимся с GND или GN
func NewWallet() (*Wallet, error) {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	pubKey := append(privKey.PublicKey.X.Bytes(), privKey.PublicKey.Y.Bytes()...)
	// Хешируем публичный ключ: SHA256 + RIPEMD160 (как в Bitcoin/Ethereum)
	shaHash := sha256.Sum256(pubKey)
	ripemdHasher := ripemd160.New()
	_, err = ripemdHasher.Write(shaHash[:])
	if err != nil {
		return nil, err
	}
	pubKeyHash := ripemdHasher.Sum(nil)

	// Создаем адрес с префиксом GND или GN
	// Можно случайно выбирать префикс или по условию
	prefix := randomPrefix()

	// Адрес = prefix + base58(pubKeyHash + checksum)
	payload := append([]byte(prefix), pubKeyHash...)

	// Добавляем контрольную сумму (первые 4 байта SHA256(SHA256(payload)))
	checksum := checksum(payload)
	fullPayload := append(payload, checksum...)

	address := base58.Encode(fullPayload)

	return &Wallet{
		PrivateKey: privKey,
		Address:    address,
	}, nil
}

// randomPrefix возвращает "GND" или "GN" с разным регистром букв
func randomPrefix() []byte {
	// Для примера: чередуем варианты
	// Можно улучшить генерацию для большего разнообразия
	prefixes := []string{
		"GND",
		"gnd",
		"GnD",
		"gNd",
		"GN",
		"gn",
		"Gn",
		"gN",
	}
	// Выбор случайного префикса
	idx, err := randInt(len(prefixes))
	if err != nil {
		return []byte("GND")
	}
	return []byte(prefixes[idx])
}

// randInt возвращает случайное число в [0, max)
func randInt(max int) (int, error) {
	b := make([]byte, 1)
	_, err := rand.Read(b)
	if err != nil {
		return 0, err
	}
	return int(b[0]) % max, nil
}

// checksum вычисляет первые 4 байта двойного SHA256
func checksum(payload []byte) []byte {
	first := sha256.Sum256(payload)
	second := sha256.Sum256(first[:])
	return second[:4]
}

// ValidateAddress проверяет корректность адреса (префикс + base58 + checksum)
func ValidateAddress(address string) bool {
	decoded := base58.Decode(address)
	if len(decoded) < 7 { // минимум: префикс(2-3) + hash(20) + checksum(4)
		return false
	}
	// Проверяем префикс
	prefix := decoded[:3]
	prefixStr := string(prefix)
	validPrefixes := []string{"GND", "gnd", "GnD", "gNd", "GN", "gn", "Gn", "gN"}
	found := false
	for _, p := range validPrefixes {
		if strings.EqualFold(prefixStr, p) {
			found = true
			break
		}
	}
	if !found {
		return false
	}
	// Проверяем контрольную сумму
	payload := decoded[:len(decoded)-4]
	checksumBytes := decoded[len(decoded)-4:]
	calcChecksum := checksum(payload)
	for i := 0; i < 4; i++ {
		if checksumBytes[i] != calcChecksum[i] {
			return false
		}
	}
	return true
}

// PrivateKeyHex возвращает приватный ключ в hex формате
func (w *Wallet) PrivateKeyHex() string {
	return hex.EncodeToString(w.PrivateKey.D.Bytes())
}
