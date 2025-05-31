package api

import (
	"GND/core"
	_ "GND/tokens"
	"GND/vm"
	"encoding/json"
	"fmt"
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
}

// --- RPC Handlers ---

func DeployContractHandler(w http.ResponseWriter, r *http.Request) {
	var params DeployContractParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	// Проверка адресов с префиксом
	if !core.ValidateAddress(params.From) || !core.ValidateAddress(params.Owner) {
		http.Error(w, "invalid address", http.StatusBadRequest)
		return
	}

	meta := vm.ContractMeta{
		Name:        params.Name,
		Standard:    vm.ContractStandard(params.Standard),
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
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := map[string]interface{}{
		"address": core.AddPrefix(addr),
	}
	json.NewEncoder(w).Encode(resp)
}

// Вспомогательные функции
func toStringMap(m map[string]interface{}) map[string]string {
	res := make(map[string]string)
	for k, v := range m {
		res[k] = fmt.Sprintf("%v", v)
	}
	return res
}
