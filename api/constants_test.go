package api

import (
	"testing"
	"time"
)

// TestConstants проверяет, что константы API заданы (использует для линтера и документации).
func TestConstants(t *testing.T) {
	if RestURL == "" || RpcURL == "" || WsURL == "" {
		t.Error("URL-константы должны быть заданы")
	}
	if ApiDocHost == "" || NodeHost == "" {
		t.Error("хосты документации и ноды должны быть заданы")
	}
	if TokenStandardGNDst1 != "GND-st1" {
		t.Errorf("TokenStandardGNDst1: ожидалось GND-st1, получено %q", TokenStandardGNDst1)
	}
	if ApiKey == "" {
		t.Error("ApiKey должен быть задан для тестов")
	}
	if HttpTimeout < time.Second || WsTimeout < time.Second {
		t.Error("таймауты должны быть не менее 1s")
	}
}
