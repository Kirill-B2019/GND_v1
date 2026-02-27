// | KB @CerbeRus - Nexus Invest Team
package core

import (
	"crypto/sha256"
	"encoding/hex"
	"math/rand"
)

// Checksum вычисляет контрольную сумму для payload
func Checksum(payload []byte) []byte {
	first := sha256.Sum256(payload)
	second := sha256.Sum256(first[:])
	return second[:4]
}

// BytesEqual безопасно сравнивает два байтовых среза
func BytesEqual(a, b []byte) bool {
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

// GenerateContractAddress генерирует уникальный адрес контракта
func GenerateContractAddress() string {
	// Генерируем случайные 16 байт
	randomBytes := make([]byte, 16)
	rand.Read(randomBytes)

	// Кодируем в hex
	return hex.EncodeToString(randomBytes)
}
