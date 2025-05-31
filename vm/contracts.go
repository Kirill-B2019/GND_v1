package vm

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
)

// ContractStandard - поддерживаемые стандарты токенов
type ContractStandard string

const (
	StandardERC20  ContractStandard = "erc20"
	StandardTRC20  ContractStandard = "trc20"
	StandardCustom ContractStandard = "custom"
)

// ContractMeta - метаданные смарт-контракта
type ContractMeta struct {
	Name        string            `json:"name"`
	Standard    ContractStandard  `json:"standard"`
	Owner       string            `json:"owner"`
	Params      map[string]string `json:"params"`
	Description string            `json:"description"`
	MetadataCID string            `json:"metadata_cid"`
	SourceCode  string            `json:"source_code"`
	Version     string            `json:"version"`
	Compiler    string            `json:"compiler"`
}

// Contract - структура смарт-контракта
type Contract struct {
	Meta     ContractMeta
	Bytecode []byte
	Storage  map[string][]byte // Простое хранилище (ключ-значение)
	// Можно добавить: счетчик газа, owner, состояние, ...
	mutex sync.RWMutex
}

// ContractRegistry - глобальный реестр контрактов
type ContractRegistry struct {
	contracts map[string]*Contract // адрес → контракт
	mutex     sync.RWMutex
}

// Новый реестр контрактов
func NewContractRegistry() *ContractRegistry {
	return &ContractRegistry{
		contracts: make(map[string]*Contract),
	}
}

// RegisterContract - регистрация нового контракта в реестре
func (cr *ContractRegistry) RegisterContract(meta ContractMeta) (*Contract, error) {
	cr.mutex.Lock()
	defer cr.mutex.Unlock()
	if _, exists := cr.contracts[meta.Address]; exists {
		return nil, errors.New("contract already exists")
	}
	bytecode, err := hex.DecodeString(meta.Bytecode)
	if err != nil {
		return nil, fmt.Errorf("invalid bytecode: %v", err)
	}
	contract := &Contract{
		Meta:     meta,
		Bytecode: bytecode,
		Storage:  make(map[string][]byte),
	}
	cr.contracts[meta.Address] = contract
	return contract, nil
}

// GetContract - получить контракт по адресу
func (cr *ContractRegistry) GetContract(address string) (*Contract, error) {
	cr.mutex.RLock()
	defer cr.mutex.RUnlock()
	contract, ok := cr.contracts[address]
	if !ok {
		return nil, errors.New("contract not found")
	}
	return contract, nil
}

// ListContracts - список всех контрактов
func (cr *ContractRegistry) ListContracts() []*Contract {
	cr.mutex.RLock()
	defer cr.mutex.RUnlock()
	var result []*Contract
	for _, c := range cr.contracts {
		result = append(result, c)
	}
	return result
}

// CallContract - универсальный вызов функции контракта (эмуляция EVM)
func (cr *ContractRegistry) CallContract(
	from, to, method string,
	args []interface{},
	gasLimit, gasPrice uint64,
	nonce uint64,
	signature string,
) (interface{}, error) {
	contract, err := cr.GetContract(to)
	if err != nil {
		return nil, err
	}
	// Списание комиссии (gas) в GND должно быть реализовано на уровне ядра!
	// Здесь только эмуляция вызова
	switch contract.Meta.Standard {
	case StandardERC20:
		return callERC20(contract, method, args)
	case StandardTRC20:
		return callTRC20(contract, method, args)
	case StandardCustom:
		return callCustom(contract, method, args)
	default:
		return nil, errors.New("unsupported contract standard")
	}
}

// ===== Примеры реализаций стандартных методов =====

// callERC20 - обработка методов ERC-20
func callERC20(contract *Contract, method string, args []interface{}) (interface{}, error) {
	contract.mutex.Lock()
	defer contract.mutex.Unlock()
	switch strings.ToLower(method) {
	case "name":
		return contract.Meta.Name, nil
	case "symbol":
		return contract.Meta.Symbol, nil
	case "decimals":
		return contract.Meta.Decimals, nil
	case "totalSupply":
		return contract.Storage["totalSupply"], nil
	case "balanceOf":
		if len(args) < 1 {
			return nil, errors.New("missing address for balanceOf")
		}
		addr, ok := args[0].(string)
		if !ok {
			return nil, errors.New("invalid address type")
		}
		return contract.Storage["balance_"+addr], nil
	case "transfer":
		if len(args) < 2 {
			return nil, errors.New("missing arguments for transfer")
		}
		from, ok1 := args[0].(string)
		to, ok2 := args[1].(string)
		amount, ok3 := args[2].(uint64)
		if !ok1 || !ok2 || !ok3 {
			return nil, errors.New("invalid arguments for transfer")
		}
		// Простейшая логика (без переполнений и т.д.)
		fromKey := "balance_" + from
		toKey := "balance_" + to
		fromBal := bytesToUint64(contract.Storage[fromKey])
		toBal := bytesToUint64(contract.Storage[toKey])
		if fromBal < amount {
			return nil, errors.New("insufficient balance")
		}
		contract.Storage[fromKey] = uint64ToBytes(fromBal - amount)
		contract.Storage[toKey] = uint64ToBytes(toBal + amount)
		return true, nil
	default:
		return nil, errors.New("unknown ERC-20 method")
	}
}

// callTRC20 - обработка методов TRC-20 (аналогично ERC-20)
func callTRC20(contract *Contract, method string, args []interface{}) (interface{}, error) {
	// Можно реализовать отдельную логику, если отличается
	return callERC20(contract, method, args)
}

// callCustom - обработка кастомных методов
func callCustom(contract *Contract, method string, args []interface{}) (interface{}, error) {
	// Для MVP - просто сохраняем/читаем значения по ключу
	contract.mutex.Lock()
	defer contract.mutex.Unlock()
	switch strings.ToLower(method) {
	case "get":
		if len(args) < 1 {
			return nil, errors.New("missing key for get")
		}
		key, ok := args[0].(string)
		if !ok {
			return nil, errors.New("invalid key type")
		}
		return contract.Storage[key], nil
	case "set":
		if len(args) < 2 {
			return nil, errors.New("missing arguments for set")
		}
		key, ok := args[0].(string)
		value, ok2 := args[1].([]byte)
		if !ok || !ok2 {
			return nil, errors.New("invalid arguments for set")
		}
		contract.Storage[key] = value
		return true, nil
	default:
		return nil, errors.New("unknown custom method")
	}
}

// ===== Вспомогательные функции =====

func bytesToUint64(b []byte) uint64 {
	if len(b) != 8 {
		return 0
	}
	return uint64(b[0])<<56 | uint64(b[1])<<48 | uint64(b[2])<<40 | uint64(b[3])<<32 |
		uint64(b[4])<<24 | uint64(b[5])<<16 | uint64(b[6])<<8 | uint64(b[7])
}

func uint64ToBytes(u uint64) []byte {
	return []byte{
		byte(u >> 56), byte(u >> 48), byte(u >> 40), byte(u >> 32),
		byte(u >> 24), byte(u >> 16), byte(u >> 8), byte(u),
	}
}
