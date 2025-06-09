//handlers_test.go

package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHelloHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/hello", nil)
	rr := httptest.NewRecorder()

	HelloHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Ожидался статус %v, получено %v", http.StatusOK, rr.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Ошибка декодирования ответа: %v", err)
	}
	if resp["message"] != "hello" {
		t.Errorf("Ожидалось 'hello', получено '%s'", resp["message"])
	}
}
