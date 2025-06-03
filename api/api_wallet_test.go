//api/api_wallet_test.go

package api

import (
	"GND/core"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Пример теста для эндпоинта /api/wallet/create
func TestWalletCreateHandler(t *testing.T) {
	// Создаём тестовый запрос
	req := httptest.NewRequest("POST", "/api/wallet/create", nil)
	req.Header.Set("X-API-Key", "ganymede-demo-key")

	// Создаём ResponseRecorder для записи ответа
	rr := httptest.NewRecorder()

	// Оборачиваем обработчик в middleware
	handler := AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wallet, err := core.NewWallet()
		if err != nil {
			http.Error(w, "failed to generate wallet", http.StatusInternalServerError)
			return
		}
		resp := map[string]interface{}{
			"address":   wallet.Address,
			"publicKey": wallet.PublicKeyHex(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))

	// Запускаем обработчик
	handler.ServeHTTP(rr, req)

	// Проверяем статус-код
	if rr.Code != http.StatusOK {
		t.Errorf("ожидался статус 200, получено %d", rr.Code)
	}

	// Проверяем, что в ответе есть адрес и публичный ключ
	body := rr.Body.String()
	if !strings.Contains(body, "address") || !strings.Contains(body, "publicKey") {
		t.Errorf("в ответе отсутствуют ключи address или publicKey: %s", body)
	}
}
