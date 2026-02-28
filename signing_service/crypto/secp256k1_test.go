// KB @CerbeRus - Nexus Invest Team
package crypto

import (
	"crypto/sha256"
	"testing"
)

func TestNewSecp256k1Key(t *testing.T) {
	priv, err := NewSecp256k1Key()
	if err != nil {
		t.Fatalf("NewSecp256k1Key: %v", err)
	}
	if priv == nil {
		t.Fatal("ключ nil")
	}
	pub := PublicKeyBytes(priv)
	if len(pub) != 33 {
		t.Errorf("ожидался сжатый публичный ключ 33 байта, получено %d", len(pub))
	}
	raw := PrivKeyToBytes(priv)
	if len(raw) != 32 {
		t.Errorf("ожидался сырой приватный ключ 32 байта, получено %d", len(raw))
	}
}

func TestPrivKeyFromBytes(t *testing.T) {
	priv, _ := NewSecp256k1Key()
	raw := PrivKeyToBytes(priv)
	restored, err := PrivKeyFromBytes(raw)
	if err != nil {
		t.Fatalf("PrivKeyFromBytes: %v", err)
	}
	if !bytesEqual(PublicKeyBytes(priv), PublicKeyBytes(restored)) {
		t.Error("восстановленный ключ не совпадает по публичной части")
	}
	_, err = PrivKeyFromBytes([]byte{})
	if err == nil {
		t.Error("ожидалась ошибка для пустого среза")
	}
}

func TestSignDigest(t *testing.T) {
	priv, _ := NewSecp256k1Key()
	digest := sha256.Sum256([]byte("tx hash to sign"))
	sig, err := SignDigest(priv, digest[:])
	if err != nil {
		t.Fatalf("SignDigest: %v", err)
	}
	if len(sig) == 0 {
		t.Error("подпись пустая")
	}
	_, err = SignDigest(priv, nil)
	if err == nil {
		t.Error("ожидалась ошибка для пустого digest")
	}
}

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
