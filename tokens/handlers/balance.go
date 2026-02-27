// | KB @CerbeRus - Nexus Invest Team
// tokens/handlers/balance.go

package handlers

import (
	"GND/tokens/registry"
	"encoding/json"
	"net/http"

	"GND/vm"
)

type APIResponse struct {
	Success bool              `json:"success"`
	Data    map[string]string `json:"data,omitempty"`
	Error   string            `json:"error,omitempty"`
}

func sendJSON(w http.ResponseWriter, data interface{}, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

// TokenBalanceHandler возвращает обработчик балансов токенов из реестра (evm зарезервирован для расширения).
func TokenBalanceHandler(_ *vm.EVM) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		address := r.URL.Query().Get("address")
		if address == "" {
			sendJSON(w, map[string]interface{}{
				"success": false,
				"error":   "address required",
			}, http.StatusBadRequest)
			return
		}

		tokens := registry.GetAllTokens()
		resp := make(map[string]string)
		for _, token := range tokens {
			resp[token.Symbol] = token.TotalSupply // TotalSupply уже string в types.TokenInfo
		}

		sendJSON(w, map[string]interface{}{
			"success": true,
			"data":    resp,
		}, http.StatusOK)
	}
}
