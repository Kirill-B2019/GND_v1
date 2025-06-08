// tokens/handlers/info.go

package handlers

import (
	"encoding/json"
	"net/http"

	"GND/tokens/registry"
)

func TokenInfoHandler(w http.ResponseWriter, r *http.Request) {
	addr := r.URL.Query().Get("address")
	token, err := registry.GetToken(addr)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(token)
}
