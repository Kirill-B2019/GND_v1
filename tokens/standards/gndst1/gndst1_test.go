// tokens/standards/gndst1/gndst1_test.go

package gndst1

import (
	"context"
	"math/big"
	"testing"
)

func TestNewGNDst1(t *testing.T) {
	token := NewGNDst1(big.NewInt(1e18), "bridge1")
	if token.BalanceOf("owner").Cmp(big.NewInt(1e18)) != 0 {
		t.FailNow()
	}
}

func TestGNDst1Snapshots(t *testing.T) {
	// Создаем тестовый токен
	token := NewGNDst1(
		"0x123",
		"Test Token",
		"TEST",
		18,
		big.NewInt(1000000),
		nil,
	)

	// Добавляем балансы
	token.balances["0x1"] = big.NewInt(100)
	token.balances["0x2"] = big.NewInt(200)

	// Создаем снимок
	snapshotId, err := token.Snapshot(context.Background())
	if err != nil {
		t.Fatalf("Failed to create snapshot: %v", err)
	}

	// Проверяем балансы в снимке
	balance1, err := token.GetSnapshotBalance(context.Background(), "0x1", snapshotId)
	if err != nil {
		t.Fatalf("Failed to get snapshot balance: %v", err)
	}
	if balance1.Cmp(big.NewInt(100)) != 0 {
		t.Errorf("Expected balance 100, got %v", balance1)
	}

	balance2, err := token.GetSnapshotBalance(context.Background(), "0x2", snapshotId)
	if err != nil {
		t.Fatalf("Failed to get snapshot balance: %v", err)
	}
	if balance2.Cmp(big.NewInt(200)) != 0 {
		t.Errorf("Expected balance 200, got %v", balance2)
	}
}

func TestGNDst1Dividends(t *testing.T) {
	token := NewGNDst1(
		"0x123",
		"Test Token",
		"TEST",
		18,
		big.NewInt(1000000),
		nil,
	)

	// Создаем снимок
	snapshotId, err := token.Snapshot(context.Background())
	if err != nil {
		t.Fatalf("Failed to create snapshot: %v", err)
	}

	// Устанавливаем дивиденды
	token.dividends[snapshotId] = big.NewInt(1000)

	// Пытаемся получить дивиденды
	err = token.ClaimDividends(context.Background(), snapshotId)
	if err == nil {
		t.Error("Expected error when claiming dividends with no balance")
	}

	// Добавляем баланс и пробуем снова
	token.balances[token.address] = big.NewInt(100)
	err = token.ClaimDividends(context.Background(), snapshotId)
	if err != nil {
		t.Fatalf("Failed to claim dividends: %v", err)
	}
}

func TestGNDst1Modules(t *testing.T) {
	token := NewGNDst1(
		"0x123",
		"Test Token",
		"TEST",
		18,
		big.NewInt(1000000),
		nil,
	)

	// Регистрируем модуль
	err := token.RegisterModule(context.Background(), "test", "0x456", "Test Module")
	if err != nil {
		t.Fatalf("Failed to register module: %v", err)
	}

	// Пытаемся вызвать несуществующий модуль
	_, err = token.ModuleCall(context.Background(), "nonexistent", []byte{})
	if err == nil {
		t.Error("Expected error when calling nonexistent module")
	}

	// Вызываем существующий модуль
	_, err = token.ModuleCall(context.Background(), "test", []byte{})
	if err == nil {
		t.Error("Expected error when calling unimplemented module")
	}
}
