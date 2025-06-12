// api/rest.go

package api

import (
	"GND/core"
	"GND/tokens/interfaces"
	"GND/tokens/registry"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"context"
	"errors"
	"math/big"

	"GND/api/middleware"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	blockchain *core.Blockchain
	mempool    *core.Mempool
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

// Обработчик создания кошелька
func handleCreateWallet(w http.ResponseWriter, r *http.Request) {
	wallet, err := core.NewWallet(nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := struct {
		Address    string `json:"address"`
		PublicKey  string `json:"publicKey"`
		PrivateKey string `json:"privateKey"`
	}{
		Address:    string(wallet.Address),
		PublicKey:  wallet.PublicKeyHex(),
		PrivateKey: wallet.PrivateKeyHex(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Обработчик получения баланса
func handleGetBalance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := core.Address(vars["address"])

	balance := blockchain.State.GetBalance(address, "GND")
	response := struct {
		Address string   `json:"address"`
		Balance *big.Int `json:"balance"`
	}{
		Address: string(address),
		Balance: balance,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Обработчик вызова токена
func handleTokenCall(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TokenAddr string        `json:"tokenAddr"`
		Method    string        `json:"method"`
		Args      []interface{} `json:"args"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	token, err := registry.GetToken(req.TokenAddr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if gnd, ok := token.(interfaces.TokenInterface); ok {
		switch req.Method {
		case "transfer":
			if len(req.Args) != 3 {
				http.Error(w, "transfer требует from, to, amount", http.StatusBadRequest)
				return
			}
			from, to := req.Args[0].(string), req.Args[1].(string)
			amount := req.Args[2].(*big.Int)
			err := gnd.Transfer(r.Context(), from, to, amount)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		case "approve":
			if len(req.Args) != 3 {
				http.Error(w, "approve требует owner, spender, amount", http.StatusBadRequest)
				return
			}
			owner, spender := req.Args[0].(string), req.Args[1].(string)
			amount := req.Args[2].(*big.Int)
			err := gnd.Approve(r.Context(), owner, spender, amount)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		default:
			http.Error(w, "неподдерживаемый метод", http.StatusBadRequest)
			return
		}
	} else {
		http.Error(w, "токен не реализует интерфейс TokenInterface", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Обработчик получения баланса токена
func handleGetTokenBalance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]

	var req struct {
		TokenAddr string `json:"tokenAddr"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	token, err := registry.GetToken(req.TokenAddr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if gnd, ok := token.(interfaces.TokenInterface); ok {
		balance, err := gnd.GetBalance(r.Context(), address)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		response := struct {
			Address string   `json:"address"`
			Balance *big.Int `json:"balance"`
		}{
			Address: address,
			Balance: balance,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	} else {
		http.Error(w, "токен не реализует интерфейс TokenInterface", http.StatusInternalServerError)
	}
}

// Обработчик отправки транзакции
func handleSendTransaction(w http.ResponseWriter, r *http.Request) {
	var tx core.Transaction
	if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := mempool.Add(&tx); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Обработчик получения транзакции
func handleGetTransaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hash := vars["hash"]

	status, err := blockchain.GetTxStatus(hash)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	response := struct {
		Hash   string `json:"hash"`
		Status string `json:"status"`
	}{
		Hash:   hash,
		Status: status,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Обработчик получения последнего блока
func handleGetLatestBlock(w http.ResponseWriter, r *http.Request) {
	block := blockchain.LatestBlock()
	if block == nil {
		http.Error(w, "блок не найден", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(block)
}

// Обработчик получения блока по номеру
func handleGetBlockByNumber(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	number := vars["number"]

	var blockNumber uint64
	if _, err := fmt.Sscanf(number, "%d", &blockNumber); err != nil {
		http.Error(w, "неверный формат номера блока", http.StatusBadRequest)
		return
	}

	block := blockchain.GetBlockByNumber(blockNumber)
	if block == nil {
		http.Error(w, "блок не найден", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(block)
}

// Обработчик деплоя контракта
func handleDeployContract(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Code []byte `json:"code"`
		ABI  []byte `json:"abi"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO: Реализовать деплой контракта
	http.Error(w, "деплой контракта пока не реализован", http.StatusNotImplemented)
}

// Обработчик получения контракта
func handleGetContract(w http.ResponseWriter, r *http.Request) {
	// TODO: Реализовать получение контракта
	http.Error(w, "получение контракта пока не реализовано", http.StatusNotImplemented)
}

// Запуск REST сервера
func StartRESTServer(bc *core.Blockchain, mp *core.Mempool, cfg *core.Config, pool *pgxpool.Pool) {
	blockchain = bc
	mempool = mp

	r := mux.NewRouter()

	// Middleware
	r.Use(RecoverMiddleware)
	r.Use(middleware.LoggerMiddleware)
	r.Use(middleware.CORSMiddleware)
	r.Use(middleware.RateLimitMiddleware)
	r.Use(middleware.AuthMiddleware)

	// API маршруты
	api := r.PathPrefix("/api").Subrouter()

	// Эндпоинты кошелька
	api.HandleFunc("/wallet/create", handleCreateWallet).Methods("POST")
	api.HandleFunc("/wallet/balance/{address}", handleGetBalance).Methods("GET")

	// Эндпоинты токенов
	api.HandleFunc("/token/call", handleTokenCall).Methods("POST")
	api.HandleFunc("/token/balance/{address}", handleGetTokenBalance).Methods("GET")

	// Эндпоинты транзакций
	api.HandleFunc("/tx/send", handleSendTransaction).Methods("POST")
	api.HandleFunc("/tx/{hash}", handleGetTransaction).Methods("GET")

	// Эндпоинты блоков
	api.HandleFunc("/block/latest", handleGetLatestBlock).Methods("GET")
	api.HandleFunc("/block/{number}", handleGetBlockByNumber).Methods("GET")

	// Эндпоинты контрактов
	api.HandleFunc("/contract/deploy", handleDeployContract).Methods("POST")
	api.HandleFunc("/contract/{address}", handleGetContract).Methods("GET")

	// Middleware для проверки API ключа
	apiKeyMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get("X-API-Key")
			if apiKey == "" {
				http.Error(w, "API key required", http.StatusUnauthorized)
				return
			}

			// Проверка API ключа в базе данных
			var exists bool
			err := pool.QueryRow(r.Context(),
				"SELECT EXISTS(SELECT 1 FROM api_keys WHERE key = $1 AND expires_at > NOW())",
				apiKey).Scan(&exists)

			if err != nil || !exists {
				http.Error(w, "Invalid or expired API key", http.StatusUnauthorized)
				return
			}
			log.Printf("=== REST Server запущен на %s ===", fmt.Sprintf("0.0.0.0:%d", cfg.Server.REST.Port))
			log.Println("Доступные эндпоинты:")
			next.ServeHTTP(w, r)
		}
	}

	// Обработчик для отправки транзакций
	api.HandleFunc("/api/v1/transactions", apiKeyMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req TransactionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Преобразование значения в big.Int
		value := new(big.Int)
		value.SetString(req.Value, 10)

		// Преобразование цены газа в big.Int
		gasPrice := new(big.Int)
		gasPrice.SetString(req.GasPrice, 10)

		// Создание и подписание транзакции
		tx, err := core.NewTransaction(
			req.From,
			req.To,
			value,
			req.Data,
			req.Nonce,
			req.GasLimit,
			gasPrice,
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error creating transaction: %v", err), http.StatusBadRequest)
			return
		}

		// Подписание транзакции
		if err := tx.Sign(req.PrivateKey); err != nil {
			http.Error(w, fmt.Sprintf("Error signing transaction: %v", err), http.StatusBadRequest)
			return
		}

		// Сохранение транзакции в БД
		if err := tx.Save(r.Context(), pool); err != nil {
			http.Error(w, fmt.Sprintf("Error saving transaction: %v", err), http.StatusInternalServerError)
			return
		}

		// Добавление транзакции в мемпул
		if err := mempool.Add(tx); err != nil {
			http.Error(w, fmt.Sprintf("Error adding transaction to mempool: %v", err), http.StatusInternalServerError)
			return
		}

		// Немедленная обработка транзакции (0 подтверждений)
		go func() {
			if err := blockchain.ProcessTransaction(tx); err != nil {
				fmt.Printf("Error processing transaction: %v\n", err)
				// Обновление статуса транзакции в случае ошибки
				if err := tx.UpdateStatus(r.Context(), pool, "failed"); err != nil {
					fmt.Printf("Error updating transaction status: %v\n", err)
				}
			} else {
				// Обновление статуса транзакции при успешной обработке
				if err := tx.UpdateStatus(r.Context(), pool, "confirmed"); err != nil {
					fmt.Printf("Error updating transaction status: %v\n", err)
				}
			}
		}()

		// Отправка ответа
		resp := TransactionResponse{
			Hash:      tx.Hash,
			Status:    "pending",
			Timestamp: time.Now(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))

	// Формируем адрес сервера
	addr := fmt.Sprintf("0.0.0.0:%d", cfg.Server.REST.Port)
	log.Printf("=== REST Server запущен на %s ===", addr)
	log.Println("Доступные эндпоинты:")
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Ошибка запуска REST сервера: %v", err)
	}
}

// Универсальный обработчик токенов
func universalHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TokenAddr string        `json:"tokenAddr"`
		Method    string        `json:"method"`
		Args      []interface{} `json:"args"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	token, err := registry.GetToken(req.TokenAddr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if gnd, ok := token.(interfaces.TokenInterface); ok {
		switch req.Method {
		case "transfer":
			if len(req.Args) != 3 {
				http.Error(w, "transfer требует from, to, amount", http.StatusBadRequest)
				return
			}
			from, to := req.Args[0].(string), req.Args[1].(string)
			amount := req.Args[2].(*big.Int)
			err := gnd.Transfer(r.Context(), from, to, amount)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		case "approve":
			if len(req.Args) != 3 {
				http.Error(w, "approve требует owner, spender, amount", http.StatusBadRequest)
				return
			}
			owner, spender := req.Args[0].(string), req.Args[1].(string)
			amount := req.Args[2].(*big.Int)
			err := gnd.Approve(r.Context(), owner, spender, amount)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		case "balanceOf":
			if len(req.Args) != 1 {
				http.Error(w, "balanceOf требует address", http.StatusBadRequest)
				return
			}
			addr := req.Args[0].(string)
			balance, err := gnd.GetBalance(r.Context(), addr)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"balance": balance.String()})
		default:
			http.Error(w, "неподдерживаемый метод", http.StatusBadRequest)
		}
	} else {
		http.Error(w, "токен не реализует интерфейс TokenInterface", http.StatusInternalServerError)
	}
}

// Пример для обработчика transfer
func handleTransfer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TokenAddress string   `json:"tokenAddress"`
		From         string   `json:"from"`
		To           string   `json:"to"`
		Amount       *big.Int `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	token, err := registry.GetToken(req.TokenAddress)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if gnd, ok := token.(interfaces.TokenInterface); ok {
		err := gnd.Transfer(r.Context(), req.From, req.To, req.Amount)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
		return
	}
	http.Error(w, "unsupported token standard", http.StatusBadRequest)
}

// UniversalTokenCall вызывает метод токена с учетом стандарта
func UniversalTokenCall(ctx context.Context, tokenAddr string, method string, args ...interface{}) (interface{}, error) {
	token, err := registry.GetToken(tokenAddr)
	if err != nil {
		return nil, err
	}
	if gnd, ok := token.(interfaces.TokenInterface); ok {
		switch method {
		case "transfer":
			if len(args) == 3 {
				return nil, gnd.Transfer(ctx, args[0].(string), args[1].(string), args[2].(*big.Int))
			}
		case "approve":
			if len(args) == 3 {
				return nil, gnd.Approve(ctx, args[0].(string), args[1].(string), args[2].(*big.Int))
			}
		case "balanceOf":
			if len(args) == 1 {
				return gnd.GetBalance(ctx, args[0].(string))
			}
			// ... другие методы ...
		}
	}
	return nil, errors.New("unsupported token standard or method")
}
