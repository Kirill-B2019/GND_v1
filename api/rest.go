//api/rest.go

package api

import (
	"GND/core"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func StartRESTServer(bc *core.Blockchain, mempool *core.Mempool, config *core.Config) {
	mux := http.NewServeMux()
	mux.HandleFunc("/block/latest", func(w http.ResponseWriter, r *http.Request) {
		block := bc.LatestBlock()
		json.NewEncoder(w).Encode(block)
	})
	addr := config.Server.RESTAddr // ← используйте это поле
	log.Printf("REST сервер запущен на %s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Ошибка запуска REST сервера: %v", err)
	}

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

		// Возвращаем балансы по всем монетам
		balances := make([]map[string]interface{}, 0)
		for _, coin := range config.Coins {
			balance := bc.State.GetBalance(core.Address(address), coin.Symbol)
			balances = append(balances, map[string]interface{}{
				"symbol":   coin.Symbol,
				"name":     coin.Name,
				"decimals": coin.Decimals,
				"balance":  balance.String(),
			})
		}

		resp := map[string]interface{}{
			"address":  address,
			"balances": balances,
		}
		json.NewEncoder(w).Encode(resp)
	})

	fmt.Sprintf(":%d", config.Server.RESTAddr)
	http.ListenAndServe(addr, nil)
}
