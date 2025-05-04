package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	_ "reflect"
	_ "strings"

	"GND/core"
	"GND/tokens"
	"GND/vm"
	// другие импорты по мере необходимости
)

// JSON-RPC 2.0 структура запроса
type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
	ID      interface{}     `json:"id"`
}

// JSON-RPC 2.0 структура ответа
type rpcResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *rpcError   `json:"error,omitempty"`
	ID      interface{} `json:"id"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// RPC-сервер: принимает запросы и маршрутизирует по методам
func StartRPCServer(bc *core.Blockchain, mempool *core.Mempool, vm *vm.EVM, tokenReg *tokens.Registry) {
	http.HandleFunc("/rpc", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
			return
		}
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			writeRPCError(w, nil, -32700, "Parse error: "+err.Error())
			return
		}
		var req rpcRequest
		if err := json.Unmarshal(body, &req); err != nil {
			writeRPCError(w, nil, -32700, "Parse error: "+err.Error())
			return
		}
		// Маршрутизация методов
		switch req.Method {
		case "blockchain_latestBlock":
			block := bc.LatestBlock()
			writeRPCResult(w, req.ID, block)
		case "blockchain_getBlockByHash":
			var params struct{ Hash string }
			json.Unmarshal(req.Params, &params)
			block := bc.GetBlockByHash(params.Hash)
			writeRPCResult(w, req.ID, block)
		case "blockchain_getBlockByIndex":
			var params struct{ Index uint64 }
			json.Unmarshal(req.Params, &params)
			block := bc.GetBlockByIndex(params.Index)
			writeRPCResult(w, req.ID, block)
		case "blockchain_sendTx":
			var tx core.Transaction
			json.Unmarshal(req.Params, &tx)
			mempool.Add(&tx)
			writeRPCResult(w, req.ID, map[string]string{"txHash": tx.Hash})
		case "blockchain_getTx":
			var params struct{ Hash string }
			json.Unmarshal(req.Params, &params)
			// Поиск по всем блокам (оптимизировать индексом)
			var tx *core.Transaction
			for _, block := range bc.AllBlocks() {
				for _, t := range block.Transactions {
					if t.Hash == params.Hash {
						tx = t
						break
					}
				}
			}
			writeRPCResult(w, req.ID, tx)
		case "state_getBalance":
			var params struct{ Address string }
			json.Unmarshal(req.Params, &params)
			balance := bc.State.BalanceOf(params.Address)
			writeRPCResult(w, req.ID, balance)
		case "state_getNonce":
			var params struct{ Address string }
			json.Unmarshal(req.Params, &params)
			nonce := bc.State.NonceOf(params.Address)
			writeRPCResult(w, req.ID, nonce)
		// ===== Методы для смарт-контрактов и токенов =====
		case "contract_deploy":
			var params struct {
				From      string
				Bytecode  []byte
				GasLimit  uint64
				GasPrice  uint64
				Nonce     uint64
				Metadata  map[string]interface{}
				Signature string
			}
			json.Unmarshal(req.Params, &params)
			// Пример: деплой контракта через VM
			address, err := vm.DeployContract(params.From, params.Bytecode, params.GasLimit, params.GasPrice, params.Nonce, params.Metadata, params.Signature)
			if err != nil {
				writeRPCError(w, req.ID, -32000, err.Error())
				return
			}
			writeRPCResult(w, req.ID, map[string]string{"contractAddress": address})
		case "contract_call":
			var params struct {
				From      string
				To        string
				Data      []byte
				GasLimit  uint64
				GasPrice  uint64
				Nonce     uint64
				Signature string
			}
			json.Unmarshal(req.Params, &params)
			// Вызов функции контракта через VM
			result, err := vm.CallContract(params.From, params.To, params.Data, params.GasLimit, params.GasPrice, params.Nonce, params.Signature)
			if err != nil {
				writeRPCError(w, req.ID, -32000, err.Error())
				return
			}
			writeRPCResult(w, req.ID, result)
		case "token_getInfo":
			var params struct{ Address string }
			json.Unmarshal(req.Params, &params)
			info := tokenReg.GetInfo(params.Address)
			writeRPCResult(w, req.ID, info)
		case "token_call":
			var params struct {
				TokenAddress string
				Method       string
				Args         []interface{}
			}
			json.Unmarshal(req.Params, &params)
			result, err := tokenReg.Call(params.TokenAddress, params.Method, params.Args...)
			if err != nil {
				writeRPCError(w, req.ID, -32000, err.Error())
				return
			}
			writeRPCResult(w, req.ID, result)
		// ====== Системные методы ======
		case "system_ping":
			writeRPCResult(w, req.ID, "pong")
		default:
			writeRPCError(w, req.ID, -32601, "Method not found: "+req.Method)
		}
	})

	log.Println("JSON-RPC сервер запущен на /rpc")
	http.ListenAndServe(":8545", nil) // порт можно вынести в конфиг
}

// Вспомогательные функции для ответа

func writeRPCResult(w http.ResponseWriter, id interface{}, result interface{}) {
	resp := rpcResponse{
		JSONRPC: "2.0",
		Result:  result,
		ID:      id,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func writeRPCError(w http.ResponseWriter, id interface{}, code int, message string) {
	resp := rpcResponse{
		JSONRPC: "2.0",
		Error:   &rpcError{Code: code, Message: message},
		ID:      id,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
