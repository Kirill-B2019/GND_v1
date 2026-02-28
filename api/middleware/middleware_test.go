// | KB @CerbeRus - Nexus Invest Team
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORS_SetsHeaders(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := CORS(next)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("ожидали Access-Control-Allow-Origin: *")
	}
	if rec.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("ожидали Access-Control-Allow-Methods")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("ожидали 200, получили %d", rec.Code)
	}
}

func TestCORS_OptionsReturns200(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next не должен вызываться для OPTIONS")
	})
	handler := CORS(next)
	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("ожидали 200 для OPTIONS, получили %d", rec.Code)
	}
}
