package core

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"

	"github.com/btcsuite/btcutil/base58"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"golang.org/x/crypto/ripemd160"
)

// Wallet хранит приватный ключ и адрес
type Wallet struct {
	PrivateKey *secp256k1.PrivateKey
	Address    string
}

// NewWallet генерирует новый кошелек с адресом, начинающимся с GND или GN (только верхний регистр)
func NewWallet() (*Wallet, error) {
	privKey, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		return nil, err
	}

	pubKey := privKey.PubKey().SerializeUncompressed()

	// Хешируем публичный ключ: SHA256 + RIPEMD160
	shaHash := sha256.Sum256(pubKey[1:]) // пропускаем первый байт 0x04
	ripemdHasher := ripemd160.New()
	_, err = ripemdHasher.Write(shaHash[:])
	if err != nil {
		return nil, err
	}
	pubKeyHash := ripemdHasher.Sum(nil)

	// Префикс "GND" или "GN" (строго верхний регистр)
	prefix := randomPrefix()

	// Адрес = prefix + pubKeyHash + checksum
	payload := append([]byte(prefix), pubKeyHash...)
	checksum := checksum(payload)
	fullPayload := append(payload, checksum...)

	address := base58.Encode(fullPayload)

	return &Wallet{
		PrivateKey: privKey,
		Address:    address,
	}, nil
}

// randomPrefix возвращает "GND" или "GN" (только верхний регистр)
func randomPrefix() []byte {
	prefixes := [][]byte{[]byte("GND"), []byte("GN")}
	b := make([]byte, 1)
	_, err := rand.Read(b)
	if err != nil {
		return []byte("GND")
	}
	return prefixes[int(b[0])%len(prefixes)]
}

// checksum вычисляет первые 4 байта двойного SHA256
func checksum(payload []byte) []byte {
	first := sha256.Sum256(payload)
	second := sha256.Sum256(first[:])
	return second[:4]
}

// ValidateAddress проверяет корректность адреса (строго GND или GN в верхнем регистре)
func ValidateAddress(address string) bool {
	decoded := base58.Decode(address)
	if len(decoded) == 27 && string(decoded[:3]) == "GND" {
		// ok
	} else if len(decoded) == 26 && string(decoded[:2]) == "GN" {
		// ok
	} else {
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
	return hex.EncodeToString(w.PrivateKey.Serialize())
}

// PublicKeyHex возвращает публичный ключ в hex формате
func (w *Wallet) PublicKeyHex() string {
	return hex.EncodeToString(w.PrivateKey.PubKey().SerializeUncompressed())
}
