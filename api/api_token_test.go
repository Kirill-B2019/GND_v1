// | KB @CerbeRus - Nexus Invest Team
package api

import (
	"bytes"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTokenBalanceHandler(t *testing.T) {
	server := setupTestServer(t)

	// Test case 1: Get balance for valid address
	reqBody := map[string]interface{}{
		"tokenAddr": "GNDctTest",
	}
	jsonBody, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("GET", "/token/balance/test_address", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Test case 2: Invalid token address
	reqBody = map[string]interface{}{
		"tokenAddr": "invalid_address",
	}
	jsonBody, _ = json.Marshal(reqBody)
	req, err = http.NewRequest("GET", "/token/balance/test_address", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatal(err)
	}

	rr = httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}
}

func TestTokenTransferHandler(t *testing.T) {
	server := setupTestServer(t)

	// Test case 1: Valid transfer
	reqBody := map[string]interface{}{
		"tokenAddr": "GNDctTest",
		"method":    "transfer",
		"args": []interface{}{
			"from_address",
			"to_address",
			big.NewInt(1000),
		},
	}
	jsonBody, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", "/token/call", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Test case 2: Invalid method
	reqBody = map[string]interface{}{
		"tokenAddr": "GNDctTest",
		"method":    "invalid_method",
		"args":      []interface{}{},
	}
	jsonBody, _ = json.Marshal(reqBody)
	req, err = http.NewRequest("POST", "/token/call", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatal(err)
	}

	rr = httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	// Test case 3: Invalid arguments
	reqBody = map[string]interface{}{
		"tokenAddr": "GNDctTest",
		"method":    "transfer",
		"args":      []interface{}{"from_address"}, // Missing required arguments
	}
	jsonBody, _ = json.Marshal(reqBody)
	req, err = http.NewRequest("POST", "/token/call", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatal(err)
	}

	rr = httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestTokenApproveHandler(t *testing.T) {
	server := setupTestServer(t)

	// Test case 1: Valid approve
	reqBody := map[string]interface{}{
		"tokenAddr": "GNDctTest",
		"method":    "approve",
		"args": []interface{}{
			"owner_address",
			"spender_address",
			big.NewInt(1000),
		},
	}
	jsonBody, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", "/token/call", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Test case 2: Invalid amount (negative)
	reqBody = map[string]interface{}{
		"tokenAddr": "GNDctTest",
		"method":    "approve",
		"args": []interface{}{
			"owner_address",
			"spender_address",
			big.NewInt(-1000),
		},
	}
	jsonBody, _ = json.Marshal(reqBody)
	req, err = http.NewRequest("POST", "/token/call", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatal(err)
	}

	rr = httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}
