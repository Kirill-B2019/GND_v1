// | KB @CerbeRus - Nexus Invest Team
package api

import (
	"encoding/json"
	"time"
)

// TransactionRequest представляет запрос на создание транзакции
type TransactionRequest struct {
	From       string          `json:"from"`
	To         string          `json:"to"`
	Value      string          `json:"value"`
	Data       json.RawMessage `json:"data,omitempty"`
	Nonce      uint64          `json:"nonce,omitempty"`
	GasLimit   uint64          `json:"gasLimit,omitempty"`
	GasPrice   string          `json:"gasPrice,omitempty"`
	PrivateKey string          `json:"privateKey"`
}

// TransactionResponse представляет ответ на создание транзакции
type TransactionResponse struct {
	Hash      string    `json:"hash"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// TestResult представляет результат теста
type TestResult struct {
	Method     string `json:"method"`
	Status     int    `json:"status"`
	Error      string `json:"error,omitempty"`
	Response   string `json:"response,omitempty"`
	Success    bool   `json:"success"`
	DurationMs int64  `json:"duration_ms"`
}

// countSuccess подсчитывает количество успешных тестов
func countSuccess(results []TestResult) int {
	count := 0
	for _, r := range results {
		if r.Success {
			count++
		}
	}
	return count
}
