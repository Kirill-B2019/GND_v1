// | KB @CerberRus00 - Nexus Invest Team
// api/abi_helper.go — разбор ABI контракта для разделения view- и write-функций.

package api

import (
	"encoding/json"
)

// ABIFunction — элемент ABI функции (упрощённо для отображения в UI).
type ABIFunction struct {
	Type            string      `json:"type"`
	Name            string      `json:"name"`
	Inputs          []ABIInput  `json:"inputs"`
	Outputs         []ABIOutput `json:"outputs,omitempty"`
	StateMutability string      `json:"stateMutability,omitempty"`
	Constant        bool        `json:"constant,omitempty"`
}

// ABIInput — входной аргумент.
type ABIInput struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// ABIOutput — выходное значение.
type ABIOutput struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// ParseABIFunctions парсит JSON ABI и возвращает списки view- и write-функций.
// View: stateMutability view/pure или constant true. Остальные функции — write.
func ParseABIFunctions(abiJSON []byte) (viewFuncs, writeFuncs []ABIFunction, err error) {
	if len(abiJSON) == 0 {
		return nil, nil, nil
	}
	var abi []ABIFunction
	if err := json.Unmarshal(abiJSON, &abi); err != nil {
		return nil, nil, err
	}
	for _, f := range abi {
		if f.Type != "function" {
			continue
		}
		isView := f.StateMutability == "view" || f.StateMutability == "pure" || f.Constant
		if isView {
			viewFuncs = append(viewFuncs, f)
		} else {
			writeFuncs = append(writeFuncs, f)
		}
	}
	return viewFuncs, writeFuncs, nil
}
