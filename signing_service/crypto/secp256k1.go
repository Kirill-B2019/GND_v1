// KB @CerbeRus - Nexus Invest Team
package crypto

import (
	"encoding/hex"
	"fmt"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
)

// NewSecp256k1Key создаёт новую пару ключей secp256k1.
func NewSecp256k1Key() (*secp256k1.PrivateKey, error) {
	return secp256k1.GeneratePrivateKey()
}

// PrivKeyFromBytes восстанавливает приватный ключ из сырых байт (например после расшифровки).
func PrivKeyFromBytes(b []byte) (*secp256k1.PrivateKey, error) {
	if len(b) == 0 {
		return nil, fmt.Errorf("empty private key bytes")
	}
	return secp256k1.PrivKeyFromBytes(b), nil
}

// PublicKeyBytes возвращает сжатый публичный ключ (33 байта).
func PublicKeyBytes(priv *secp256k1.PrivateKey) []byte {
	return priv.PubKey().SerializeCompressed()
}

// PrivKeyToBytes возвращает 32 байта сырого приватного ключа для шифрования/хранения.
func PrivKeyToBytes(priv *secp256k1.PrivateKey) []byte {
	return priv.Serialize()
}

// SignDigest подписывает digest по secp256k1 ECDSA. Возвращает сериализованную подпись (DER).
func SignDigest(priv *secp256k1.PrivateKey, digest []byte) ([]byte, error) {
	if len(digest) == 0 {
		return nil, fmt.Errorf("empty digest")
	}
	sig := ecdsa.Sign(priv, digest)
	return sig.Serialize(), nil
}

// HexToBytes декодирует hex-строку в байты.
func HexToBytes(h string) ([]byte, error) {
	return hex.DecodeString(h)
}

// BytesToHex кодирует байты в hex-строку.
func BytesToHex(b []byte) string {
	return hex.EncodeToString(b)
}
