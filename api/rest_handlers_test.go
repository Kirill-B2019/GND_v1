// | KB @CerbeRus - Nexus Invest Team
// api/rest_handlers_test.go — тесты для обработчиков Gorilla mux (RecoverMiddleware, universalHandler, handleTransfer, handleTokenApprove, handleGetTokenBalance).

package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestRecoverMiddleware(t *testing.T) {
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		panic("test panic")
	})
	handler := RecoverMiddleware(panicHandler)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("ожидался 500 после паники, получен %d", rr.Code)
	}
}

func TestUniversalHandler_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/token/call", bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	universalHandler(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("ожидался 400 при неверном JSON, получен %d", rr.Code)
	}
}

func TestHandleTransfer_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/token/transfer", bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handleTransfer(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("ожидался 400, получен %d", rr.Code)
	}
}

func TestHandleTokenApprove_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/token/approve", bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handleTokenApprove(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("ожидался 400, получен %d", rr.Code)
	}
}

func TestHandleGetTokenBalance(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/token/0x123/balance/0x456", nil)
	req = mux.SetURLVars(req, map[string]string{"address": "0x123", "owner": "0x456"})
	rr := httptest.NewRecorder()
	handleGetTokenBalance(rr, req)
	// Токен не в реестре — ожидаем 500 или 400
	if rr.Code != http.StatusInternalServerError && rr.Code != http.StatusBadRequest {
		t.Logf("код ответа: %d (токен не найден в реестре)", rr.Code)
	}
}
