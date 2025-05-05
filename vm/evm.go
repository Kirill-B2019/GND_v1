package vm

import (
	"GND/utils"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
)

// EVMConfig определяет параметры виртуальной машины
type EVMConfig struct {
	GasLimit uint64 // лимит газа на выполнение одной транзакции/контракта
}

// EVM реализует изолированную виртуальную машину для исполнения байткода контрактов
type EVM struct {
	config    EVMConfig
	contracts *ContractRegistry // глобальный реестр контрактов
	state     map[string]uint64 // простейшее состояние для счетчиков газа и балансов
	mutex     sync.RWMutex
}

// NewEVM создает новый экземпляр EVM
func NewEVM(config EVMConfig, contracts *ContractRegistry) *EVM {
	return &EVM{
		config:    config,
		contracts: contracts,
		state:     make(map[string]uint64),
	}
}

// DeployContract деплоит байткод контракта, регистрирует его и списывает комиссию в GND
func (evm *EVM) DeployContract(from string, bytecode []byte, meta ContractMeta, gasLimit, gasPrice, nonce uint64, signature string) (string, error) {
	evm.mutex.Lock()
	defer evm.mutex.Unlock()

	// Проверка газа и баланса (упрощенно)
	requiredFee := gasLimit * gasPrice
	if evm.state[from] < requiredFee {
		return "", errors.New("insufficient GND for deploy fee")
	}
	evm.state[from] -= requiredFee

	// Генерация адреса контракта (упрощенно: hash от байткода + nonce)
	addr := fmt.Sprintf("GNDct%x", hashBytes(append(bytecode, byte(nonce))))
	meta.Address = addr
	meta.Bytecode = hex.EncodeToString(bytecode)

	_, err := evm.contracts.RegisterContract(meta)
	if err != nil {
		return "", err
	}
	return addr, nil
}

// CallContract выполняет функцию контракта с учетом лимита газа и списания комиссии
func (evm *EVM) CallContract(from, to string, data []byte, gasLimit, gasPrice, nonce uint64, signature string) (interface{}, error) {
	from = utils.RemovePrefix(from)
	to = utils.RemovePrefix(to)
	evm.mutex.Lock()
	defer evm.mutex.Unlock()

	requiredFee := gasLimit * gasPrice
	if evm.state[from] < requiredFee {
		return nil, errors.New("insufficient GND for call fee")
	}
	evm.state[from] -= requiredFee

	// Эмуляция вызова: определяем метод и аргументы из data (упрощенно)
	method, args, err := decodeCallData(data)
	if err != nil {
		return nil, err
	}

	// Вызов универсального обработчика
	result, err := evm.contracts.CallContract(from, to, method, args, gasLimit, gasPrice, nonce, signature)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// SetBalance устанавливает баланс GND для адреса (для тестов/деплоя)
func (evm *EVM) SetBalance(addr string, amount uint64) {
	evm.mutex.Lock()
	defer evm.mutex.Unlock()
	evm.state[addr] = amount
}

// GetBalance возвращает баланс GND для адреса
func (evm *EVM) GetBalance(addr string) uint64 {
	evm.mutex.RLock()
	defer evm.mutex.RUnlock()
	return evm.state[addr]
}

// hashBytes простая хеш-функция для адресации (можно заменить на sha256/ripemd160)
func hashBytes(data []byte) []byte {
	var h uint64 = 5381
	for _, b := range data {
		h = ((h << 5) + h) + uint64(b)
	}
	out := make([]byte, 8)
	for i := 0; i < 8; i++ {
		out[i] = byte((h >> (8 * uint(i))) & 0xff)
	}
	return out
}

// decodeCallData разбирает входные данные вызова контракта (упрощенно)
// В реальной EVM используется ABI-декодер
func decodeCallData(data []byte) (string, []interface{}, error) {
	if len(data) == 0 {
		return "", nil, errors.New("empty calldata")
	}
	// Пример: "transfer:GND2...,1000" → method="transfer", args=["GND2...", 1000]
	parts := string(data)
	split := []rune(parts)
	for i, c := range split {
		if c == ':' {
			method := string(split[:i])
			argsStr := string(split[i+1:])
			args := parseArgs(argsStr)
			return method, args, nil
		}
	}
	return string(data), nil, nil
}

// parseArgs разбирает строку аргументов (упрощенно)
func parseArgs(argsStr string) []interface{} {
	var args []interface{}
	for _, s := range splitAndTrim(argsStr, ",") {
		if n, err := parseUint64(s); err == nil {
			args = append(args, n)
		} else {
			args = append(args, s)
		}
	}
	return args
}

func splitAndTrim(s string, sep string) []string {
	var result []string
	for _, part := range split(s, sep) {
		result = append(result, trim(part))
	}
	return result
}

func split(s, sep string) []string {
	var res []string
	i := 0
	for {
		j := i
		for j < len(s) && string(s[j]) != sep {
			j++
		}
		res = append(res, s[i:j])
		if j == len(s) {
			break
		}
		i = j + 1
	}
	return res
}

func trim(s string) string {
	return s // для простоты, можно strings.TrimSpace(s)
}

func parseUint64(s string) (uint64, error) {
	var n uint64
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}
