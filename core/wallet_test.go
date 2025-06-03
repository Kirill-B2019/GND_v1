//core/wallet_test.go

package core

import (
	"testing"
)

// Тест на успешную генерацию кошелька
func TestNewWallet(t *testing.T) {
	wallet, err := NewWallet()
	if err != nil {
		t.Fatalf("ошибка генерации кошелька: %v", err)
	}

	// Проверяем, что адрес не пустой
	if wallet.Address == "" {
		t.Error("адрес кошелька пустой")
	}

	// Проверяем, что приватный ключ не nil
	if wallet.PrivateKey == nil {
		t.Error("приватный ключ не сгенерирован")
	}

	// Проверяем, что адрес проходит валидацию
	if !ValidateAddress(string(wallet.Address)) {
		t.Errorf("адрес %s не проходит валидацию", wallet.Address)
	}

	// Проверяем, что публичный ключ корректный (длина 33 байта для secp256k1 compressed)
	pubKey := wallet.PrivateKey.PubKey().SerializeCompressed()
	if len(pubKey) != 33 {
		t.Errorf("некорректная длина публичного ключа: %d", len(pubKey))
	}
}
