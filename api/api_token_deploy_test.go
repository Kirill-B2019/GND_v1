// | KB @CerbeRus - Nexus Invest Team
// api/api_token_deploy_test.go

package api

import (
	"GND/core"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
	server := NewServer(nil, bc, core.NewMempool(), nil)

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
	server := NewServer(nil, bc, core.NewMempool(), nil)

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
	server := NewServer(nil, bc, core.NewMempool(), nil)

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
	req.Header.Set("X-API-Key", ApiKey)
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

func TestValidateAPIKey_ConstantKey_ReturnsTrue(t *testing.T) {
	if !ValidateAPIKey(context.Background(), nil, ApiKey) {
		t.Error("константный ApiKey должен проходить")
	}
}

func TestValidateAPIKey_UnknownKey_ReturnsFalse(t *testing.T) {
	if ValidateAPIKey(context.Background(), nil, "unknown_key") {
		t.Error("неизвестный ключ не должен проходить при pool=nil")
	}
}
