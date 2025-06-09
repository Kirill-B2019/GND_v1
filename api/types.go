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
