//core/wallet_test.go

package core

import (
	"context"
	"encoding/json"
	"os"
	"testing"
)

// dbConfig для опционального подключения к БД
type dbConfig struct {
	DB struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		DBName   string `json:"dbname"`
		SSLMode  string `json:"sslmode"`
		MaxConns int    `json:"max_conns"`
		MinConns int    `json:"min_conns"`
	} `json:"database"`
}

// Тест на успешную генерацию кошелька (требует БД; при недоступности — пропуск)
func TestNewWallet(t *testing.T) {
	data, err := os.ReadFile("config/db.json")
	if err != nil {
		t.Skipf("config/db.json не найден: %v", err)
		return
	}
	var cfg dbConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Skipf("ошибка парсинга db.json: %v", err)
		return
	}
	pool, err := InitDBPool(context.Background(), DBConfig{
		Host: cfg.DB.Host, Port: cfg.DB.Port, User: cfg.DB.User,
		Password: cfg.DB.Password, DBName: cfg.DB.DBName, SSLMode: cfg.DB.SSLMode,
		MaxConns: cfg.DB.MaxConns, MinConns: cfg.DB.MinConns,
	})
	if err != nil {
		t.Skipf("подключение к БД недоступно: %v", err)
		return
	}
	defer pool.Close()

	wallet, err := NewWallet(pool)
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
