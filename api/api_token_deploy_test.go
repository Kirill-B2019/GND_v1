// | KB @CerbeRus - Nexus Invest Team
// api/api_token_deploy_test.go

package api

import (
	"GND/core"
	"GND/tokens/deployer"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestDeployToken_NoAPIKey_Returns401(t *testing.T) {
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
	// Ненулевой deployer, чтобы проверка 503 не сработала и сработала проверка API key.
	server := NewServer(nil, bc, core.NewMempool(), deployer.NewDeployer(nil, nil, nil), nil)

	body := map[string]interface{}{
		"name":         "Test",
		"symbol":       "TST",
		"decimals":     18,
		"total_supply": "1000000000000000000000000",
		"owner":        "GND9jbK6Vca5VcZxATt3zb9yz5KQeMwjHFrz",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/token/deploy", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("ожидался 401 без X-API-Key, получен %d", w.Code)
	}
	var resp APIResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Success {
		t.Error("ожидался success: false")
	}
	if resp.Error == "" {
		t.Error("ожидалось сообщение об ошибке")
	}
}

func TestDeployToken_InvalidAPIKey_Returns401(t *testing.T) {
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
	// Ненулевой deployer, чтобы проверка 503 не сработала и сработала проверка API key.
	server := NewServer(nil, bc, core.NewMempool(), deployer.NewDeployer(nil, nil, nil), nil)

	body := map[string]interface{}{
		"name":         "Test",
		"symbol":       "TST",
		"decimals":     18,
		"total_supply": "1000000000000000000000000",
		"owner":        "GND9jbK6Vca5VcZxATt3zb9yz5KQeMwjHFrz",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/token/deploy", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "invalid_key")
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("ожидался 401 при неверном ключе, получен %d", w.Code)
	}
}

func TestDeployToken_ValidAPIKey_NoDeployer_Returns503(t *testing.T) {
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
	server := NewServer(nil, bc, core.NewMempool(), nil, nil)

	body := map[string]interface{}{
		"name":         "Test",
		"symbol":       "TST",
		"decimals":     18,
		"total_supply": "1000000000000000000000000",
		"owner":        "GND9jbK6Vca5VcZxATt3zb9yz5KQeMwjHFrz",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/token/deploy", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	if k := os.Getenv("GND_API_KEY"); k != "" {
		req.Header.Set("X-API-Key", k)
	}
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("ожидался 503 при отсутствии deployer, получен %d", w.Code)
	}
	var resp APIResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Success {
		t.Error("ожидался success: false")
	}
}

func TestValidateAPIKey_EmptyKey_ReturnsFalse(t *testing.T) {
	if ValidateAPIKey(context.Background(), nil, "") {
		t.Error("пустой ключ не должен проходить")
	}
}

func TestValidateAPIKey_NoPool_NonEmptyKey_ReturnsFalse(t *testing.T) {
	// При pool == nil проверка идёт только по БД; без пула любой ключ не проходит.
	if ValidateAPIKey(context.Background(), nil, "any_key") {
		t.Error("при pool=nil ключ не должен проходить")
	}
}

func TestValidateAPIKey_UnknownKey_ReturnsFalse(t *testing.T) {
	if ValidateAPIKey(context.Background(), nil, "unknown_key") {
		t.Error("неизвестный ключ не должен проходить при pool=nil")
	}
}
