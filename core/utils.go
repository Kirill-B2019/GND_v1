package core

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
)

// GenerateRandomBytes возвращает случайный срез байт заданной длины
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// GenerateRandomAddress генерирует случайный адрес в hex формате (20 байт)
func GenerateRandomAddress() (string, error) {
	b, err := GenerateRandomBytes(20)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// Uint64ToString конвертирует uint64 в строку
func Uint64ToString(num uint64) string {
	return strconv.FormatUint(num, 10)
}

// StringToUint64 конвертирует строку в uint64, возвращает ошибку если не удалось
func StringToUint64(s string) (uint64, error) {
	return strconv.ParseUint(s, 10, 64)
}

// HashSHA256 вычисляет SHA256 хеш от входных данных и возвращает hex строку
func HashSHA256(data []byte) string {
	h := sha256Sum(data)
	return hex.EncodeToString(h)
}

// sha256Sum - внутренний хешер (возвращает байты)
func sha256Sum(data []byte) []byte {
	// Импорт crypto/sha256

	hash := sha256.Sum256(data)
	return hash[:]
}

// CheckError печатает ошибку, если она не nil
func CheckError(err error) {
	if err != nil {
		fmt.Println("Error:", err)
	}
}
