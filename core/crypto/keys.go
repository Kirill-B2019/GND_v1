// | KB @CerbeRus - Nexus Invest Team
package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
)

// GenerateKeyPair генерирует новую пару ключей
func GenerateKeyPair() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

// PrivateKeyToHex конвертирует приватный ключ в hex строку
func PrivateKeyToHex(key *ecdsa.PrivateKey) string {
	return "0x" + hex.EncodeToString(key.D.Bytes())
}

// HexToPrivateKey конвертирует hex строку в приватный ключ
func HexToPrivateKey(hexKey string) (*ecdsa.PrivateKey, error) {
	if len(hexKey) < 2 || hexKey[:2] != "0x" {
		return nil, fmt.Errorf("invalid hex key format")
	}

	keyBytes, err := hex.DecodeString(hexKey[2:])
	if err != nil {
		return nil, fmt.Errorf("invalid hex key: %v", err)
	}

	key := new(ecdsa.PrivateKey)
	key.Curve = elliptic.P256()
	key.D = new(big.Int).SetBytes(keyBytes)

	key.PublicKey.X, key.PublicKey.Y = key.Curve.ScalarBaseMult(keyBytes)

	return key, nil
}

// p256CurveOrderSize — размер r/s в байтах для P-256 (N порядка 2^256)
const p256CurveOrderSize = 32

// Sign подписывает данные приватным ключом. Подпись: r (32 байта) || s (32 байта).
func Sign(data []byte, privateKey *ecdsa.PrivateKey) ([]byte, error) {
	hash := sha256.Sum256(data)
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash[:])
	if err != nil {
		return nil, fmt.Errorf("failed to sign: %v", err)
	}
	rb := padBigIntBytes(r.Bytes(), p256CurveOrderSize)
	sb := padBigIntBytes(s.Bytes(), p256CurveOrderSize)
	return append(rb, sb...), nil
}

// Verify проверяет подпись. Ожидает подпись длиной 64 байта (r || s по 32 байта).
func Verify(data []byte, signature []byte, publicKey *ecdsa.PublicKey) bool {
	if len(signature) != p256CurveOrderSize*2 {
		return false
	}
	hash := sha256.Sum256(data)
	r := new(big.Int).SetBytes(signature[:p256CurveOrderSize])
	s := new(big.Int).SetBytes(signature[p256CurveOrderSize:])
	return ecdsa.Verify(publicKey, hash[:], r, s)
}

// padBigIntBytes дополняет байты big.Int слева нулями до нужной длины.
func padBigIntBytes(b []byte, size int) []byte {
	if len(b) >= size {
		return b[len(b)-size:]
	}
	padded := make([]byte, size)
	copy(padded[size-len(b):], b)
	return padded
}

// PublicKeyToAddress конвертирует публичный ключ в адрес
func PublicKeyToAddress(publicKey *ecdsa.PublicKey) string {
	// Берем последние 20 байт хеша публичного ключа
	hash := sha256.Sum256(append(publicKey.X.Bytes(), publicKey.Y.Bytes()...))
	return "0x" + hex.EncodeToString(hash[len(hash)-20:])
}

// GetCurve возвращает используемую эллиптическую кривую
func GetCurve() elliptic.Curve {
	return elliptic.P256()
}

// ParsePublicKeyHex парсит P-256 публичный ключ из hex (uncompressed: 04 + X + Y = 130 hex символов).
func ParsePublicKeyHex(hexStr string) (*ecdsa.PublicKey, error) {
	s := hexStr
	if len(s) >= 2 && (s[:2] == "0x" || s[:2] == "0X") {
		s = s[2:]
	}
	if len(s) != 130 {
		return nil, fmt.Errorf("ожидается 130 hex символов (uncompressed), получено %d", len(s))
	}
	if s[:2] != "04" {
		return nil, fmt.Errorf("uncompressed ключ должен начинаться с 04")
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("hex: %w", err)
	}
	// b = 04 || X (32) || Y (32)
	if len(b) != 65 {
		return nil, fmt.Errorf("неверная длина ключа: %d", len(b))
	}
	x := new(big.Int).SetBytes(b[1:33])
	y := new(big.Int).SetBytes(b[33:65])
	return &ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}, nil
}

// PublicKeyToAddressP256 возвращает адрес в формате types (64 hex = SHA256(pubkey_bytes)).
// Используется для сравнения с tx.Sender при проверке подписи.
func PublicKeyToAddressP256(publicKey *ecdsa.PublicKey) string {
	b := PublicKeyUncompressedBytes(publicKey)
	hash := sha256.Sum256(b)
	return hex.EncodeToString(hash[:])
}

// PublicKeyUncompressedBytes сериализует P-256 публичный ключ в uncompressed (04||X||Y).
func PublicKeyUncompressedBytes(publicKey *ecdsa.PublicKey) []byte {
	const size = 32
	x := padBigIntBytes(publicKey.X.Bytes(), size)
	y := padBigIntBytes(publicKey.Y.Bytes(), size)
	return append(append([]byte{0x04}, x...), y...)
}
