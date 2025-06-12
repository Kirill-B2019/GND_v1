package api

import (
	"GND/core"
	_ "GND/tokens"
	"GND/types"
	"GND/vm"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	_ "strconv"
)

// --- Расширенная структура параметров деплоя контракта ---
type DeployContractParams struct {
	From        string                 `json:"from"`         // Адрес отправителя (с префиксом)
	Bytecode    []byte                 `json:"bytecode"`     // Байткод контракта (raw)
	Name        string                 `json:"name"`         // Имя контракта
	Standard    string                 `json:"standard"`     // Стандарт (например, ERC20)
	Owner       string                 `json:"owner"`        // Владелец (с префиксом)
	Compiler    string                 `json:"compiler"`     // Версия компилятора
	Version     string                 `json:"version"`      // Версия контракта
	Params      map[string]interface{} `json:"params"`       // Параметры конструктора (любые)
	Description string                 `json:"description"`  // Описание
	MetadataCID string                 `json:"metadata_cid"` // CID в IPFS для метаданных (если есть)
	SourceCode  string                 `json:"source_code"`  // Исходный код (опционально)
	GasLimit    uint64                 `json:"gas_limit"`
	GasPrice    uint64                 `json:"gas_price"`
	Nonce       uint64                 `json:"nonce"`
	Signature   string                 `json:"signature"`
	TotalSupply *big.Int               `json:"total_supply"`
}

func StartRPCServer(evm *vm.EVM, addr string) error {
	mux := http.NewServeMux()

	// Регистрация всех эндпоинтов
	endpoints := map[string]http.HandlerFunc{
		"/block/latest":         LatestBlockHandler(evm),
		"/contract/deploy":      DeployContractHandler(evm),
		"/contract/call":        CallContractHandler(evm),
		"/contract/send":        SendContractTxHandler(evm),
		"/account/balance":      AccountBalanceHandler(evm),
		"/block/by-number":      BlockByNumberHandler(evm),
		"/tx/send":              SendTxHandler(evm),
		"/tx/status":            TxStatusHandler(evm),
		"/token/universal-call": UniversalTokenCallHandler(),
	}

	// Добавляем middleware для CORS и безопасности
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "https://api.gnd-net.com")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key")
		w.Header().Set("Access-Control-Max-Age", "86400")

		// Security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Обработка запроса
		path := r.URL.Path
		if handler, ok := endpoints[path]; ok {
			handler(w, r)
		} else {
			http.Error(w, "Not Found", http.StatusNotFound)
		}
	})

	// Логирование информации о запуске
	log.Printf("=== RPC Server запущен на %s ===", addr)
	log.Println("Доступные эндпоинты:")
	for path := range endpoints {
		log.Printf("  %s", path)
	}
	log.Println("===============================")

	return http.ListenAndServe(addr, mux)
}

// --- RPC Handlers ---
func DeployContractHandler(evm *vm.EVM) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var params DeployContractParams
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		// ... валидация ...

		meta := types.ContractMeta{
			Name:        params.Name,
			Standard:    params.Standard,
			Owner:       params.Owner,
			Params:      toStringMap(params.Params),
			Description: params.Description,
			MetadataCID: params.MetadataCID,
			SourceCode:  params.SourceCode,
			Version:     params.Version,
			Compiler:    params.Compiler,
		}

		addr, err := evm.DeployContract(
			params.From,
			params.Bytecode,
			meta,
			params.GasLimit,
			params.GasPrice,
			params.Nonce,
			params.Signature,
			params.TotalSupply,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resp := map[string]interface{}{"address": addr}
		json.NewEncoder(w).Encode(resp)
	}
}
func CallContractHandler(evm *vm.EVM) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var params struct {
			From      string `json:"from"`
			To        string `json:"to"`
			Data      []byte `json:"data"` // ABI-encoded вызов
			GasLimit  uint64 `json:"gas_limit"`
			GasPrice  uint64 `json:"gas_price"`
			Value     uint64 `json:"value"`
			Signature string `json:"signature"`
		}
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		result, err := evm.CallContract(params.From, params.To, params.Data, params.GasLimit, params.GasPrice, params.Value, params.Signature)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"result": result})
	}
}
func SendContractTxHandler(evm *vm.EVM) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var params struct {
			From      string `json:"from"`
			To        string `json:"to"`
			Data      []byte `json:"data"`
			GasLimit  uint64 `json:"gas_limit"`
			GasPrice  uint64 `json:"gas_price"`
			Value     uint64 `json:"value"`
			Nonce     uint64 `json:"nonce"`
			Signature string `json:"signature"`
		}
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		result, err := evm.CallContract(
			params.From,
			params.To,
			params.Data,
			params.GasLimit,
			params.GasPrice,
			params.Value,
			params.Signature,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"result": result})
	}
}
func AccountBalanceHandler(evm *vm.EVM) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		address := r.URL.Query().Get("address")
		if address == "" {
			http.Error(w, "address required", http.StatusBadRequest)
			return
		}
		balance, err := evm.GetBalance(address)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"address": address, "balance": balance})
	}
}
func LatestBlockHandler(evm *vm.EVM) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		block := evm.LatestBlock()
		if block == nil {
			http.Error(w, "block not found", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(block)
	}
}
func BlockByNumberHandler(evm *vm.EVM) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var params struct {
			Number uint64 `json:"number"`
		}
		if r.Method == "GET" {
			// Для обратной совместимости: поддержка query-параметра
			numStr := r.URL.Query().Get("number")
			if numStr != "" {
				fmt.Sscanf(numStr, "%d", &params.Number)
			}
		} else {
			if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
				http.Error(w, "invalid request", http.StatusBadRequest)
				return
			}
		}
		block := evm.BlockByNumber(params.Number)
		if block == nil {
			http.Error(w, "block not found", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(block)
	}
}
func SendTxHandler(evm *vm.EVM) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var params struct {
			RawTx []byte `json:"raw_tx"`
		}
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		txHash, err := evm.SendRawTransaction(params.RawTx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"txHash": txHash})
	}
}
func TxStatusHandler(evm *vm.EVM) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		txHash := r.URL.Query().Get("hash")
		if txHash == "" {
			http.Error(w, "hash required", http.StatusBadRequest)
			return
		}
		status, err := evm.GetTxStatus(txHash)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"status": status})
	}
}

// Пример универсального вызова токена через RPC
func UniversalTokenCallHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var params struct {
			TokenAddress string        `json:"token_address"`
			Method       string        `json:"method"`
			Args         []interface{} `json:"args"`
		}
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		token := &core.Token{Address: params.TokenAddress}
		res, err := token.UniversalCall(r.Context(), params.Method, params.Args...)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"result": res})
	}
}

// Вспомогательные функции
func toStringMap(m map[string]interface{}) map[string]string {
	res := make(map[string]string)
	for k, v := range m {
		res[k] = fmt.Sprintf("%v", v)
	}
	return res
}
