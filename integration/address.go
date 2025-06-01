package integration

import (
	"GND/core"
	"crypto/sha256"
	"errors"
	"github.com/btcsuite/btcutil/base58"
	"golang.org/x/crypto/ripemd160"
)

// Просто проксируем функцию из core
func ValidateAddress(addr string) bool {
	return core.ValidateAddress(addr)
}

// Генерация адреса (пример — всегда с префиксом GND_)
func NewAddress(pubKey []byte) (string, error) {
	if len(pubKey) == 0 {
		return "", errors.New("empty public key")
	}
	shaHash := sha256.Sum256(pubKey)
	ripemdHasher := ripemd160.New()
	ripemdHasher.Write(shaHash[:])
	ripemdHash := ripemdHasher.Sum(nil)
	prefix := []byte("")
	payload := append(prefix, ripemdHash...)
	checksum := core.Checksum(payload) // используем функцию из core
	fullPayload := append(payload, checksum...)
	addr := base58.Encode(fullPayload)
	return addr, nil
}
