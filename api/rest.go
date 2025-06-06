// api/rest.go

package api

import (
	"GND/core"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"net/http"
	"strings"
)

// APIResponse — унифицированная структура ответа API
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func sendJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Ошибка кодирования JSON: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func sendError(w http.ResponseWriter, message string, statusCode int) {
	sendJSON(w, APIResponse{Success: false, Error: message}, statusCode)
}

// RecoverMiddleware защищает от паник в хендлерах
func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("Recovered from panic: %v", rec)
				sendError(w, "Internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// AuthMiddleware — заглушка для проверки API-ключа (в будущем реализовать полноценную проверку)

func StartRESTServer(
	bc *core.Blockchain,
	mempool *core.Mempool,
	config *core.Config,
	pool *pgxpool.Pool,
) {
	mux := http.NewServeMux()

	// Приветственное сообщение
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		nodeAddr := config.Server.REST.RESTAddr
		coinsInfo := make([]map[string]interface{}, len(config.Coins))
		for i, coin := range config.Coins {
			coinsInfo[i] = map[string]interface{}{
				"name":        coin.Name,
				"symbol":      coin.Symbol,
				"decimals":    coin.Decimals,
				"description": coin.Description,
			}
		}

		resp := map[string]interface{}{
			"message":   "Hello from Ganymede",
			"version":   "1.0",
			"node_addr": nodeAddr,
			"coins":     coinsInfo,
		}
		sendJSON(w, APIResponse{Success: true, Data: resp}, http.StatusOK)
	})

	// Получить последний блок
	mux.HandleFunc("/block/latest", func(w http.ResponseWriter, r *http.Request) {
		block := bc.LatestBlock()
		sendJSON(w, APIResponse{Success: true, Data: block}, http.StatusOK)
	})

	// Отправить транзакцию
	mux.HandleFunc("/tx/send", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var tx core.Transaction
		if r.Header.Get("Content-Type") != "application/json" {
			sendError(w, "Unsupported content type", http.StatusUnsupportedMediaType)
			return
		}

		if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
			sendError(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
			return
		}

		if !core.ValidateAddress(tx.From) || !core.ValidateAddress(tx.To) {
			sendError(w, "Invalid sender or recipient address", http.StatusBadRequest)
			return
		}

		if err := mempool.Add(&tx); err != nil {
			sendError(w, fmt.Sprintf("Transaction rejected: %v", err), http.StatusBadRequest)
			return
		}

		sendJSON(w, APIResponse{
			Success: true,
			Data: map[string]string{
				"txHash": tx.Hash,
				"status": "pending",
			},
		}, http.StatusAccepted)
	})

	// Получить баланс кошелька
	mux.HandleFunc("/api/wallet/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		path := r.URL.Path[len("/api/wallet/"):]
		parts := strings.Split(path, "/balance")
		if len(parts) != 1 || parts[0] == "" {
			sendError(w, "Not found", http.StatusNotFound)
			return
		}
		address := parts[0]

		if !core.ValidateAddress(address) {
			sendError(w, "Invalid address", http.StatusBadRequest)
			return
		}

		addr := core.Address(address)
		balances := make([]map[string]interface{}, 0, len(config.Coins))

		for _, coin := range config.Coins {
			balance := bc.State.GetBalance(addr, coin.Symbol)
			balances = append(balances, map[string]interface{}{
				"symbol":   coin.Symbol,
				"name":     coin.Name,
				"decimals": coin.Decimals,
				"balance":  balance.String(),
			})
		}

		sendJSON(w, APIResponse{
			Success: true,
			Data: map[string]interface{}{
				"address":  address,
				"balances": balances,
			},
		}, http.StatusOK)
	})

	// Создать новый кошелёк
	mux.Handle("/api/wallet/create", AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		wallet, err := core.NewWallet(pool)
		if err != nil {
			sendError(w, fmt.Sprintf("Failed to generate wallet: %v", err), http.StatusInternalServerError)
			return
		}

		resp := map[string]interface{}{
			"address":   wallet.Address,
			"publicKey": wallet.PublicKeyHex(),
			//"privateKey": wallet.PrivateKeyHex(), // никогда не отправляйте приватный ключ!
		}
		sendJSON(w, APIResponse{Success: true, Data: resp}, http.StatusOK)
	})))

	// Деплой смарт-контракта
	mux.HandleFunc("/contract/deploy", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		// Заглушка — здесь должен быть вызов EVM
		sendError(w, "Not implemented", http.StatusNotImplemented)
	})

	// Middleware: восстановление после паник + логирование
	handler := RecoverMiddleware(mux)

	addr := config.Server.REST.RESTAddr
	log.Printf("REST сервер запущен на %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Ошибка запуска REST сервера: %v", err)
	}
}
