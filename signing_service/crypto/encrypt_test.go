// KB @CerbeRus - Nexus Invest Team
package crypto

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestLoadMasterKey(t *testing.T) {
	key32 := make([]byte, 32)
	for i := range key32 {
		key32[i] = byte(i)
	}
	keyHex := hex.EncodeToString(key32)
	loaded, err := LoadMasterKey(keyHex)
	if err != nil {
		t.Fatalf("LoadMasterKey: %v", err)
	}
	if !bytes.Equal(loaded, key32) {
		t.Error("LoadMasterKey: ключ не совпадает")
	}
	_, err = LoadMasterKey(hex.EncodeToString([]byte{1, 2, 3}))
	if err == nil {
		t.Error("ожидалась ошибка для короткого ключа")
	}
	_, err = LoadMasterKey("zz")
	if err == nil {
		t.Error("ожидалась ошибка для не-hex строки")
	}
}

func TestEncryptDecrypt(t *testing.T) {
	key, _ := hex.DecodeString("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	plain := []byte("secret private key 32 bytes!!!!!!!!!")
	ct, err := EncryptPrivKey(plain, key)
	if err != nil {
		t.Fatalf("EncryptPrivKey: %v", err)
	}
	if len(ct) <= len(plain) {
		t.Error("шифртекст должен быть длиннее (nonce+tag)")
	}
	dec, err := DecryptPrivKey(ct, key)
	if err != nil {
		t.Fatalf("DecryptPrivKey: %v", err)
	}
	if !bytes.Equal(dec, plain) {
		t.Error("расшифрованный текст не совпадает")
	}
	_, err = DecryptPrivKey(ct[:10], key)
	if err == nil {
		t.Error("ожидалась ошибка для короткого шифртекста")
	}
	wrongKey, _ := hex.DecodeString("fedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210")
	_, err = DecryptPrivKey(ct, wrongKey)
	if err == nil {
		t.Error("ожидалась ошибка при неверном ключе")
	}
}

func TestZeroBytes(t *testing.T) {
	b := []byte{1, 2, 3, 4, 5}
	ZeroBytes(b)
	for i, v := range b {
		if v != 0 {
			t.Errorf("ZeroBytes: байт [%d] = %d", i, v)
		}
	}
}
