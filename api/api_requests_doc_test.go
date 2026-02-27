// | KB @CerbeRus - Nexus Invest Team
// api_requests_doc_test.go — проверка ответов всех URL из docs/api-requests.md

package api

import (
	"GND/core"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func setupServerForDocTest(t *testing.T) *Server {
	genesis := &core.Block{
		Index:     0,
		Timestamp: time.Now(),
		Miner:     "test",
		GasUsed:   0,
		GasLimit:  10_000_000,
		Consensus: "poa",
		Nonce:     0,
		Status:    "finalized",
	}
	genesis.Hash = genesis.CalculateHash()
	bc := core.NewBlockchain(genesis, nil)
	return NewServer(nil, bc, core.NewMempool(), nil)
}

func TestDocURLs_HealthAndMetrics(t *testing.T) {
	s := setupServerForDocTest(t)

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
		checkData  bool
	}{
		{"GET /api/v1/health", "GET", "/api/v1/health", http.StatusOK, true},
		{"GET /api/v1/metrics", "GET", "/api/v1/metrics", http.StatusOK, true},
		{"GET /api/v1/metrics/transactions", "GET", "/api/v1/metrics/transactions", http.StatusOK, true},
		{"GET /api/v1/metrics/fees", "GET", "/api/v1/metrics/fees", http.StatusOK, true},
		{"GET /api/v1/fees", "GET", "/api/v1/fees", http.StatusOK, true},
		{"GET /api/v1/alerts", "GET", "/api/v1/alerts", http.StatusOK, false}, // data может быть пустым
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			s.router.ServeHTTP(w, req)
			if w.Code != tt.wantStatus {
				t.Errorf("получен статус %d, ожидался %d", w.Code, tt.wantStatus)
			}
			var resp APIResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if tt.wantStatus == http.StatusOK && !resp.Success {
				t.Error("ожидался success: true")
			}
			if tt.checkData && tt.wantStatus == http.StatusOK && resp.Data == nil {
				t.Error("ожидалось поле data")
			}
		})
	}
}

func TestDocURLs_WalletBalance(t *testing.T) {
	s := setupServerForDocTest(t)
	req := httptest.NewRequest("GET", "/api/v1/wallet/GND9jbK6Vca5VcZxATt3zb9yz5KQeMwjHFrz/balance", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("GET /api/v1/wallet/:address/balance: статус %d", w.Code)
	}
	var resp APIResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !resp.Success {
		t.Error("ожидался success: true")
	}
	if resp.Data == nil {
		t.Error("ожидались data (address, balances)")
	}
}

func TestDocURLs_TransactionHelpAndList(t *testing.T) {
	s := setupServerForDocTest(t)

	// GET /api/v1/transaction без хеша — подсказка 400
	req := httptest.NewRequest("GET", "/api/v1/transaction", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("GET /api/v1/transaction без хеша: ожидался 400, получен %d", w.Code)
	}

	// GET /api/v1/transaction/ с слэшем — то же 400
	req = httptest.NewRequest("GET", "/api/v1/transaction/", nil)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("GET /api/v1/transaction/: ожидался 400, получен %d", w.Code)
	}

	// GET /api/v1/transactions и /api/v1/mempool — 200, data: size, pending_hashes
	for _, path := range []string{"/api/v1/transactions", "/api/v1/mempool"} {
		req = httptest.NewRequest("GET", path, nil)
		w = httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("GET %s: статус %d", path, w.Code)
		}
		var resp APIResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if !resp.Success || resp.Data == nil {
			t.Errorf("%s: ожидались success и data", path)
		}
	}
}

func TestDocURLs_Blocks(t *testing.T) {
	s := setupServerForDocTest(t)

	// block/latest, block/0, block/1 — без БД (nil pool) возможен 500; с БД — 200 с data
	allowedStatuses := []int{http.StatusOK, http.StatusNotFound, http.StatusInternalServerError}
	for _, path := range []string{"/api/v1/block/latest", "/api/v1/block/0", "/api/v1/block/1"} {
		req := httptest.NewRequest("GET", path, nil)
		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)
		ok := false
		for _, code := range allowedStatuses {
			if w.Code == code {
				ok = true
				break
			}
		}
		if !ok {
			t.Errorf("GET %s: неожиданный статус %d", path, w.Code)
		}
		if w.Code == http.StatusOK {
			var resp APIResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Errorf("%s: decode: %v", path, err)
			} else if !resp.Success || resp.Data == nil {
				t.Errorf("%s: ожидались success и data", path)
			}
		}
	}
}

func TestDocURLs_ContractGet(t *testing.T) {
	s := setupServerForDocTest(t)
	req := httptest.NewRequest("GET", "/api/v1/contract/GNDctTestAddress123", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	// 200 — контракт найден; 404 — не найден; 500 — БД недоступна (тест без pool)
	if w.Code != http.StatusOK && w.Code != http.StatusNotFound && w.Code != http.StatusInternalServerError {
		t.Errorf("GET /api/v1/contract/:address: статус %d", w.Code)
	}
}

func TestDocURLs_TokenBalance(t *testing.T) {
	s := setupServerForDocTest(t)
	req := httptest.NewRequest("GET", "/api/v1/token/GNDctToken/balance/GND9jbK6Vca5VcZxATt3zb9yz5KQeMwjHFrz", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	// 200 с data или 500/404 при отсутствии токена — оба допустимы
	if w.Code != http.StatusOK && w.Code != http.StatusNotFound && w.Code != http.StatusInternalServerError {
		t.Errorf("GET /api/v1/token/:address/balance/:owner: статус %d", w.Code)
	}
}
