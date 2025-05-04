package api

import (
	"GND/core" // Импортируйте свой пакет core
	"encoding/json"
	"net/http"
)

// Экспортируемая функция с большой буквы!
func StartRESTServer(bc *core.Blockchain, mempool *core.Mempool) {
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
		mempool.Add(&tx)
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{"txHash": tx.Hash})
	})

	// Порт можно взять из конфига, здесь для примера 8080
	http.ListenAndServe(":8080", nil)
}
