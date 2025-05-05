package vm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"
)

// ContractMetadata содержит метаданные смарт-контракта
type ContractMetadata struct {
	Name        string                 `json:"name"`
	Standard    string                 `json:"standard"` // "erc20", "trc20", "custom"
	Owner       string                 `json:"owner"`
	Compiler    string                 `json:"compiler"`
	Version     string                 `json:"version"`
	Params      map[string]interface{} `json:"params"`
	Description string                 `json:"description"`
}

// CompileResult результат компиляции
type CompileResult struct {
	Bytecode string           // hex-код байткода
	ABI      string           // JSON ABI
	Metadata ContractMetadata // метаданные
	Warnings []string
	Errors   []string
}

// SolidityCompiler интерфейс для компиляции Solidity-контрактов
type SolidityCompiler interface {
	Compile(source []byte, metadata ContractMetadata) (*CompileResult, error)
}

// DefaultSolidityCompiler реализует SolidityCompiler через внешний solc
type DefaultSolidityCompiler struct {
	SolcPath string // путь к solc (например, "solc" если в $PATH)
}

// Compile компилирует исходник Solidity и возвращает байткод, ABI и метаданные
func (c *DefaultSolidityCompiler) Compile(source []byte, metadata ContractMetadata) (*CompileResult, error) {
	tmpDir, err := ioutil.TempDir("", "ganymede-solc")
	if err != nil {
		return nil, err
	}
	defer func() { _ = removeDir(tmpDir) }()

	srcFile := filepath.Join(tmpDir, "contract.sol")
	if err := ioutil.WriteFile(srcFile, source, 0644); err != nil {
		return nil, err
	}

	// Запускаем solc для получения байткода и ABI
	cmd := exec.Command(c.SolcPath,
		"--optimize",
		"--combined-json", "abi,bin",
		srcFile,
	)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("solc error: %v, %s", err, stderr.String())
	}

	var result map[string]interface{}
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		return nil, fmt.Errorf("solc output parse error: %v", err)
	}

	contracts, ok := result["contracts"].(map[string]interface{})
	if !ok || len(contracts) == 0 {
		return nil, errors.New("no contracts found in solc output")
	}

	// Берем первый контракт (или ищем по имени)
	var contractData map[string]interface{}
	for _, v := range contracts {
		contractData, _ = v.(map[string]interface{})
		break
	}
	if contractData == nil {
		return nil, errors.New("failed to parse contract data")
	}

	abi, _ := contractData["abi"].(string)
	bin, _ := contractData["bin"].(string)

	if bin == "" {
		return nil, errors.New("empty bytecode")
	}

	compileResult := &CompileResult{
		Bytecode: bin,
		ABI:      abi,
		Metadata: metadata,
	}

	return compileResult, nil
}

// removeDir безопасно удаляет временную директорию
func removeDir(dir string) error {
	return exec.Command("rm", "-rf", dir).Run()
}

// ValidateContract проверяет соответствие байткода и метаданных стандарту
func ValidateContract(result *CompileResult) error {
	standard := strings.ToLower(result.Metadata.Standard)
	switch standard {
	case "erc20":
		// Простейшая проверка: наличие методов transfer, balanceOf и т.д.
		if !strings.Contains(result.ABI, "transfer") || !strings.Contains(result.ABI, "balanceOf") {
			return errors.New("contract does not implement ERC-20 interface")
		}
	case "trc20":
		if !strings.Contains(result.ABI, "transfer") || !strings.Contains(result.ABI, "balanceOf") {
			return errors.New("contract does not implement TRC-20 interface")
		}
	case "custom":
		// Кастомные стандарты: можно добавить свою валидацию
		return nil
	default:
		return errors.New("unknown contract standard")
	}
	return nil
}

// Пример использования:
//
// func main() {
//     compiler := &DefaultSolidityCompiler{SolcPath: "solc"}
//     source, _ := ioutil.ReadFile("MyToken.sol")
//     metadata := ContractMetadata{
//         Name: "MyToken",
//         Standard: "erc20",
//         Owner: "GND1...",
//         Compiler: "solc",
//         Version: "0.8.24",
//         Params: map[string]interface{}{"initialSupply": 1000},
//         Description: "Demo ERC-20 token",
//     }
//     res, err := compiler.Compile(source, metadata)
//     if err != nil {
//         panic(err)
//     }
//     if err := ValidateContract(res); err != nil {
//         panic(err)
//     }
//     fmt.Println("Bytecode:", res.Bytecode)
//     fmt.Println("ABI:", res.ABI)
// }
