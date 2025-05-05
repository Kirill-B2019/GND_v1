package integration

import (
	"crypto/sha256"
	"errors"
	"strings"

	"github.com/btcsuite/btcutil/base58"
	"golang.org/x/crypto/ripemd160"
)

// Address стандартный тип для адресов в блокчейне ГАНИМЕД
type Address string

// ValidateAddress проверяет корректность адреса с учетом стандартов Base58 и префиксов GND/GN
func ValidateAddress(addr string) bool {
	addr = strings.TrimSpace(addr)
	if len(addr) < 26 || len(addr) > 35 {
		return false
	}

	// Проверяем префикс (GND/GN с разным регистром)
	prefixes := []string{"GND", "gnd", "GnD", "gNd", "GN", "gn", "Gn", "gN"}
	validPrefix := false
	for _, p := range prefixes {
		if strings.HasPrefix(addr, p) {
			validPrefix = true
			break
		}
	}
	if !validPrefix {
		return false
	}

	// Проверка Base58 и контрольной суммы
	decoded := base58.Decode(addr)
	if len(decoded) < 8 {
		return false
	}
	payload := decoded[:len(decoded)-4]
	checksum := decoded[len(decoded)-4:]
	calculated := checksumBytes(payload)
	for i := 0; i < 4; i++ {
		if checksum[i] != calculated[i] {
			return false
		}
	}
	return true
}

// checksumBytes вычисляет контрольную сумму для payload
func checksumBytes(payload []byte) []byte {
	first := sha256.Sum256(payload)
	second := sha256.Sum256(first[:])
	return second[:4]
}

// NewAddress создает новый адрес из публичного ключа (хеширование, добавление префикса и контрольной суммы)
func NewAddress(pubKey []byte) (Address, error) {
	if len(pubKey) == 0 {
		return "", errors.New("empty public key")
	}
	// SHA256 + RIPEMD160
	shaHash := sha256.Sum256(pubKey)
	ripemdHasher := ripemd160.New()
	ripemdHasher.Write(shaHash[:])
	ripemdHash := ripemdHasher.Sum(nil)

	// Добавляем префикс (например, "GND")
	prefix := []byte("GND")
	payload := append(prefix, ripemdHash...)

	// Контрольная сумма
	checksum := checksumBytes(payload)
	fullPayload := append(payload, checksum...)

	// Кодируем Base58
	addr := base58.Encode(fullPayload)
	return Address(addr), nil
}
