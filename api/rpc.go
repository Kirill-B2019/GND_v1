package api

import (
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
	mux.HandleFunc("/block/latest", LatestBlockHandler(evm))
	mux.HandleFunc("/contract/deploy", DeployContractHandler(evm))
	mux.HandleFunc("/contract/call", CallContractHandler(evm))
	mux.HandleFunc("/contract/send", SendContractTxHandler(evm))
	mux.HandleFunc("/account/balance", AccountBalanceHandler(evm))
	mux.HandleFunc("/block/by-number", BlockByNumberHandler(evm))
	mux.HandleFunc("/tx/send", SendTxHandler(evm))
	mux.HandleFunc("/tx/status", TxStatusHandler(evm))
	log.Printf("RPC Server сервер запущен на %s", addr)
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
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
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

// Вспомогательные функции
func toStringMap(m map[string]interface{}) map[string]string {
	res := make(map[string]string)
	for k, v := range m {
		res[k] = fmt.Sprintf("%v", v)
	}
	return res
}
