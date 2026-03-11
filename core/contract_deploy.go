// | KB @CerberRus00 - Nexus Invest Team
// Пакет core: склейка bytecode с ABI-кодированными аргументами конструктора при деплое.

package core

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// AppendConstructorArgs возвращает bytecode + ABI-кодированные аргументы конструктора.
// abiJSON — полный ABI контракта (массив); params — карта имя_аргумента -> значение (строка или число).
// Адреса могут быть в формате GNDct<32hex>, 0x<40hex> или GN_/GND (тогда извлекаем 20 байт из hex-суффикса или используем хеш).
func AppendConstructorArgs(bytecodeHex string, abiJSON []byte, params map[string]interface{}) ([]byte, error) {
	if len(params) == 0 {
		bytecode, err := hex.DecodeString(strings.TrimPrefix(strings.TrimSpace(bytecodeHex), "0x"))
		if err != nil {
			return nil, fmt.Errorf("bytecode hex: %w", err)
		}
		return bytecode, nil
	}

	parsed, err := abi.JSON(strings.NewReader(string(abiJSON)))
	if err != nil {
		return nil, fmt.Errorf("parse ABI: %w", err)
	}

	if parsed.Constructor.Inputs == nil || len(parsed.Constructor.Inputs) == 0 {
		bytecode, err := hex.DecodeString(strings.TrimPrefix(strings.TrimSpace(bytecodeHex), "0x"))
		if err != nil {
			return nil, fmt.Errorf("bytecode hex: %w", err)
		}
		return bytecode, nil
	}

	// Собираем аргументы в порядке inputs
	args := make([]interface{}, 0, len(parsed.Constructor.Inputs))
	for _, inp := range parsed.Constructor.Inputs {
		v, ok := params[inp.Name]
		if !ok {
			return nil, fmt.Errorf("параметр конструктора не задан: %s", inp.Name)
		}
		converted, err := convertABIValue(inp.Type.String(), v)
		if err != nil {
			return nil, fmt.Errorf("аргумент %s: %w", inp.Name, err)
		}
		args = append(args, converted)
	}

	packed, err := parsed.Constructor.Inputs.Pack(args...)
	if err != nil {
		return nil, fmt.Errorf("pack constructor args: %w", err)
	}

	bytecode, err := hex.DecodeString(strings.TrimPrefix(strings.TrimSpace(bytecodeHex), "0x"))
	if err != nil {
		return nil, fmt.Errorf("bytecode hex: %w", err)
	}

	return append(bytecode, packed...), nil
}

// convertABIValue приводит значение из params (JSON/map) к типу для ABI (address, uint256, string и т.д.).
func convertABIValue(typ string, v interface{}) (interface{}, error) {
	if v == nil {
		return nil, fmt.Errorf("nil value")
	}
	switch {
	case typ == "address" || strings.HasPrefix(typ, "address"):
		return parseAddress(v)
	case typ == "uint8":
		return parseUint8(v)
	case strings.HasPrefix(typ, "uint") || strings.HasPrefix(typ, "int"):
		return parseInt256(v)
	case typ == "string" || typ == "bytes":
		return parseStringOrBytes(v)
	default:
		return parseInt256(v) // uint256 по умолчанию для чисел
	}
}

func parseAddress(v interface{}) (common.Address, error) {
	s, ok := v.(string)
	if !ok {
		return common.Address{}, fmt.Errorf("address: ожидается строка, получен %T", v)
	}
	s = strings.TrimSpace(s)
	// 0x + 40 hex = 20 байт
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		if len(s) != 42 {
			return common.Address{}, fmt.Errorf("address 0x: ожидается 40 hex-символов")
		}
		b, err := hex.DecodeString(s[2:])
		if err != nil {
			return common.Address{}, err
		}
		var a common.Address
		copy(a[:], b)
		return a, nil
	}
	// GNDct + 32 hex = 16 байт — дополняем слева нулями до 20
	if strings.HasPrefix(s, "GNDct") && len(s) == 37 {
		b, err := hex.DecodeString(s[5:])
		if err != nil {
			return common.Address{}, err
		}
		var a common.Address
		copy(a[20-len(b):], b)
		return a, nil
	}
	// GN_ / GND (кошелёк) — обычно длинный base58 или hex; для простоты считаем, что может быть 0x40 hex
	if len(s) >= 40 && len(s) <= 66 {
		if b, err := hex.DecodeString(strings.TrimPrefix(s, "0x")); err == nil && len(b) <= 20 {
			var a common.Address
			copy(a[20-len(b):], b)
			return a, nil
		}
	}
	// Попытка декодировать как hex (40 символов = 20 байт)
	if len(s) == 40 {
		b, err := hex.DecodeString(s)
		if err != nil {
			return common.Address{}, err
		}
		var a common.Address
		copy(a[:], b)
		return a, nil
	}
	return common.Address{}, fmt.Errorf("неизвестный формат адреса: %s", s)
}

func parseInt256(v interface{}) (*big.Int, error) {
	switch x := v.(type) {
	case string:
		s := strings.TrimSpace(x)
		if s == "" {
			return big.NewInt(0), nil
		}
		n := new(big.Int)
		if _, ok := n.SetString(s, 10); ok {
			return n, nil
		}
		if strings.HasPrefix(s, "0x") {
			if _, ok := n.SetString(s[2:], 16); ok {
				return n, nil
			}
		}
		return nil, fmt.Errorf("uint256: не удалось разобрать число %s", s)
	case float64:
		return big.NewInt(int64(x)), nil
	case json.Number:
		n := new(big.Int)
		if _, ok := n.SetString(x.String(), 10); !ok {
			return nil, fmt.Errorf("uint256: не удалось разобрать %s", x.String())
		}
		return n, nil
	default:
		return nil, fmt.Errorf("uint256: неверный тип %T", v)
	}
}

func parseUint8(v interface{}) (uint8, error) {
	switch x := v.(type) {
	case string:
		s := strings.TrimSpace(x)
		if s == "" {
			return 0, nil
		}
		var n int64
		if _, err := fmt.Sscanf(s, "%d", &n); err == nil && n >= 0 && n <= 255 {
			return uint8(n), nil
		}
		return 0, fmt.Errorf("uint8: ожидается 0–255, получено %s", s)
	case float64:
		if x >= 0 && x <= 255 {
			return uint8(x), nil
		}
		return 0, fmt.Errorf("uint8: ожидается 0–255, получено %v", x)
	case json.Number:
		n, err := x.Int64()
		if err != nil || n < 0 || n > 255 {
			return 0, fmt.Errorf("uint8: ожидается 0–255")
		}
		return uint8(n), nil
	default:
		return 0, fmt.Errorf("uint8: неверный тип %T", v)
	}
}

func parseStringOrBytes(v interface{}) (interface{}, error) {
	s, ok := v.(string)
	if !ok {
		return nil, fmt.Errorf("ожидается строка, получен %T", v)
	}
	return s, nil
}
