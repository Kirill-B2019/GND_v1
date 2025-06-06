// tokens/handler.go

package tokens

import (
	"encoding/json"
	"net/http"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func sendJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func sendError(w http.ResponseWriter, message string, code int) {
	sendJSON(w, APIResponse{Success: false, Error: message}, code)
}

func TokenListHandler(tokens []*TokenInfo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sendJSON(w, APIResponse{Success: true, Data: tokens}, http.StatusOK)
	}
}

func TokenBalanceHandler(evm *vm.EVM) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		address := r.URL.Query().Get("address")
		if address == "" {
			sendError(w, "Address required", http.StatusBadRequest)
			return
		}
		resp := make(map[string]interface{})
		for _, token := range GetAllTokens() {
			balance, _ := token.BalanceOf(vm.Address(address))
			resp[token.Symbol] = balance.String()
		}
		sendJSON(w, APIResponse{Success: true, Data: resp}, http.StatusOK)
	}
}
