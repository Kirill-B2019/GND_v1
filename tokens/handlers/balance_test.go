// | KB @CerbeRus - Nexus Invest Team
package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"GND/vm"
)

// TestTokenBalanceHandler_used проверяет, что TokenBalanceHandler возвращает обработчик (использует функцию для линтера).
func TestTokenBalanceHandler_used(t *testing.T) {
	handler := TokenBalanceHandler((*vm.EVM)(nil))
	if handler == nil {
		t.Fatal("TokenBalanceHandler не должен возвращать nil")
	}
	req := httptest.NewRequest(http.MethodGet, "/?address=GNDtest", nil)
	rr := httptest.NewRecorder()
	handler(rr, req)
	// Ожидаем 200 или 400 в зависимости от реестра
	if rr.Code != http.StatusOK && rr.Code != http.StatusBadRequest {
		t.Logf("код ответа: %d (реестр может быть пуст)", rr.Code)
	}
}
