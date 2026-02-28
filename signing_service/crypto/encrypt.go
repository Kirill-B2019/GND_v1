// KB @CerbeRus - Nexus Invest Team
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
)

const aesKeyLen = 32

// LoadMasterKey декодирует MASTER_KEY из hex (должно быть 32 байта для AES-256).
func LoadMasterKey(raw string) ([]byte, error) {
	b, err := hex.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("decode master key: %w", err)
	}
	if len(b) != aesKeyLen {
		return nil, fmt.Errorf("master key must be %d bytes, got %d", aesKeyLen, len(b))
	}
	return b, nil
}

// EncryptPrivKey шифрует открытый приватный ключ AES-GCM. Формат: nonce || ciphertext || authTag.
func EncryptPrivKey(plain, masterKey []byte) ([]byte, error) {
	if len(masterKey) != aesKeyLen {
		return nil, fmt.Errorf("master key must be %d bytes", aesKeyLen)
	}
	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	ct := gcm.Seal(nil, nonce, plain, nil)
	return append(nonce, ct...), nil
}

// DecryptPrivKey расшифровывает (nonce || ciphertext || authTag) и возвращает открытый приватный ключ.
// Вызывающий обязан обнулить возвращённый срез после использования.
func DecryptPrivKey(ciphertext, masterKey []byte) ([]byte, error) {
	if len(masterKey) != aesKeyLen {
		return nil, fmt.Errorf("master key must be %d bytes", aesKeyLen)
	}
	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	ns := gcm.NonceSize()
	if len(ciphertext) < ns {
		return nil, fmt.Errorf("ciphertext too short")
	}
	nonce := ciphertext[:ns]
	data := ciphertext[ns:]
	return gcm.Open(nil, nonce, data, nil)
}

// ZeroBytes затирает срез нулями, чтобы не оставлять секреты в памяти.
func ZeroBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
