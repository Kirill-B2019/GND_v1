package api

import (
	"GND/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"GND/tokens"
	"GND/vm"
)

// Глобальный реестр токенов (можно сделать синглтоном в tokens/registry.go)
var tokenReg = tokens.NewTokenRegistry()

// Глобальный экземпляр виртуальной машины (пример, настройте под свою архитектуру)
var evm = vm.NewEVM(vm.EVMConfig{GasLimit: 10000000}, nil) // передайте нужный ContractRegistry

// --- Примеры структур для запросов ---

type DeployContractParams struct {
	From        string                 `json:"from"`
	Bytecode    []byte                 `json:"bytecode"`
	Name        string                 `json:"name"`
	Standard    string                 `json:"standard"`
	Owner       string                 `json:"owner"`
	Compiler    string                 `json:"compiler"`
	Version     string                 `json:"version"`
	Params      map[string]interface{} `json:"params"`
	Description string                 `json:"description"`
	GasLimit    uint64                 `json:"gas_limit"`
	GasPrice    uint64                 `json:"gas_price"`
	Nonce       uint64                 `json:"nonce"`
	Signature   string                 `json:"signature"`
}

// --- RPC Handlers ---

// Handler для деплоя смарт-контракта
func DeployContractHandler(w http.ResponseWriter, r *http.Request) {
	var params DeployContractParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	meta := vm.ContractMeta{
		Name:        params.Name,
		Standard:    vm.ContractStandard(params.Standard),
		Owner:       params.Owner,
		Params:      toStringMap(params.Params),
		Description: params.Description,
	}

	addr, err := evm.DeployContract(
		params.From,
		params.Bytecode,
		meta,
		params.GasLimit,
		params.GasPrice,
		params.Nonce,
		params.Signature,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp := map[string]interface{}{
		"address": utils.AddPrefix(addr), // добавляем префикс GN
	}
	json.NewEncoder(w).Encode(resp)
}

// Handler для получения информации о токене
func TokenInfoHandler(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	address = utils.RemovePrefix(address) // убираем префикс, если пришёл
	token, err := tokenReg.GetToken(address)
	if err != nil {
		http.Error(w, "token not found", http.StatusNotFound)
		return
	}
	meta := token.Meta()
	meta.Address = utils.AddPrefix(meta.Address) // добавляем префикс для отображения
	json.NewEncoder(w).Encode(meta)
}

// Handler для баланса токена
func TokenBalanceHandler(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("token")
	user := r.URL.Query().Get("user")
	token, err := tokenReg.GetToken(address)
	if err != nil {
		http.Error(w, "token not found", http.StatusNotFound)
		return
	}
	balance := token.BalanceOf(user)
	resp := map[string]interface{}{
		"user":    user,
		"token":   address,
		"balance": balance,
	}
	json.NewEncoder(w).Encode(resp)
}

// Handler для трансфера токена
func TokenTransferHandler(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("token")
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	amountStr := r.URL.Query().Get("amount")
	amount, err := strconv.ParseUint(amountStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid amount", http.StatusBadRequest)
		return
	}
	token, err := tokenReg.GetToken(address)
	if err != nil {
		http.Error(w, "token not found", http.StatusNotFound)
		return
	}
	if err := token.Transfer(from, to, amount); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp := map[string]interface{}{
		"from":   from,
		"to":     to,
		"amount": amount,
		"token":  address,
		"status": "ok",
	}
	json.NewEncoder(w).Encode(resp)
}

// Универсальный вызов метода токена
func TokenCallHandler(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("token")
	method := r.URL.Query().Get("method")
	args := r.URL.Query()["arg"] // можно передавать несколько arg=...
	_, err := tokenReg.GetToken(address)
	if err != nil {
		http.Error(w, "token not found", http.StatusNotFound)
		return
	}
	result, err := tokenReg.CallTokenMethod(address, method, toInterfaceSlice(args)...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp := map[string]interface{}{
		"result": result,
	}
	json.NewEncoder(w).Encode(resp)
}

// Вспомогательная функция для преобразования []string в []interface{}
func toInterfaceSlice(args []string) []interface{} {
	out := make([]interface{}, len(args))
	for i, v := range args {
		out[i] = v
	}
	return out
}

func toStringMap(m map[string]interface{}) map[string]string {
	res := make(map[string]string)
	for k, v := range m {
		res[k] = fmt.Sprintf("%v", v)
	}
	return res
}

// Пример регистрации маршрутов (в main.go или router.go)
/*
import "net/http"

func RegisterRoutes() {
	http.HandleFunc("/api/contract/deploy", DeployContractHandler)
	http.HandleFunc("/api/token/info", TokenInfoHandler)
	http.HandleFunc("/api/token/balance", TokenBalanceHandler)
	http.HandleFunc("/api/token/transfer", TokenTransferHandler)
	http.HandleFunc("/api/token/call", TokenCallHandler)
}
*/
