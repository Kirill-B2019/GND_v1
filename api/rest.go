package api

import (
	"GND/core"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func StartRESTServer(bc *core.Blockchain, mempool *core.Mempool, config *core.Config) {
	http.HandleFunc("/block/latest", func(w http.ResponseWriter, r *http.Request) {
		block := bc.LatestBlock()
		json.NewEncoder(w).Encode(block)
	})

	http.HandleFunc("/tx/send", func(w http.ResponseWriter, r *http.Request) {
		var tx core.Transaction
		if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if !core.ValidateAddress(tx.From) || !core.ValidateAddress(tx.To) {
			http.Error(w, "invalid address", http.StatusBadRequest)
			return
		}
		mempool.Add(&tx)
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{"txHash": tx.Hash})
	})

	http.HandleFunc("/api/wallet/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path[len("/api/wallet/"):]
		parts := strings.Split(path, "/")
		if len(parts) != 2 || parts[1] != "balance" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		address := parts[0]
		if !core.ValidateAddress(address) {
			http.Error(w, "invalid address", http.StatusBadRequest)
			return
		}
		balance := bc.State.BalanceOf(address)

		resp := map[string]interface{}{
			"address":  address,
			"balance":  balance,
			"name":     config.Coin.Name,
			"symbol":   config.Coin.Symbol,
			"decimals": config.Coin.Decimals,
		}
		json.NewEncoder(w).Encode(resp)
	})

	addr := fmt.Sprintf(":%d", config.RestPort)
	http.ListenAndServe(addr, nil)
}
