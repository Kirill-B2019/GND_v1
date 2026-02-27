// | KB @CerbeRus - Nexus Invest Team
// tokens/deployer/compiler.go

package deployer

import "os"

func CompileSolidity(path string) ([]byte, error) {
	// здесь логика вызова solc
	return os.ReadFile(path)
}
