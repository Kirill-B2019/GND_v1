// | KB @CerbeRus - Nexus Invest Team
package crypto

import (
	"testing"
)

func TestGenerateKeyPair(t *testing.T) {
	priv, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}
	if priv == nil {
		t.Fatal("ожидали не nil PrivateKey")
	}
	if priv.Curve == nil {
		t.Error("Curve не должна быть nil")
	}
}

func TestPrivateKeyToHex_HexToPrivateKey_Roundtrip(t *testing.T) {
	priv, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}
	hexStr := PrivateKeyToHex(priv)
	if hexStr[:2] != "0x" {
		t.Error("ожидали префикс 0x")
	}
	restored, err := HexToPrivateKey(hexStr)
	if err != nil {
		t.Fatalf("HexToPrivateKey: %v", err)
	}
	if restored.D.Cmp(priv.D) != 0 {
		t.Error("D не совпадает после roundtrip")
	}
}

func TestSign_Verify(t *testing.T) {
	priv, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}
	data := []byte("test message")
	sig, err := Sign(data, priv)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	if len(sig) == 0 {
		t.Error("ожидали непустую подпись")
	}
	if !Verify(data, sig, &priv.PublicKey) {
		t.Error("Verify должен вернуть true для корректной подписи")
	}
	if Verify([]byte("other"), sig, &priv.PublicKey) {
		t.Error("Verify должен вернуть false для других данных")
	}
	// Испорченная подпись не должна проходить проверку
	badSig := make([]byte, len(sig))
	copy(badSig, sig)
	if len(badSig) > 0 {
		badSig[0] ^= 0xff
	}
	if Verify(data, badSig, &priv.PublicKey) {
		t.Error("Verify должен вернуть false для испорченной подписи")
	}
}

func TestHexToPrivateKey_Invalid(t *testing.T) {
	_, err := HexToPrivateKey("no-prefix")
	if err == nil {
		t.Error("ожидали ошибку для ключа без 0x")
	}
	_, err = HexToPrivateKey("0xZZ")
	if err == nil {
		t.Error("ожидали ошибку для невалидного hex")
	}
}
