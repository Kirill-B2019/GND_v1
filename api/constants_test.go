// | KB @CerbeRus - Nexus Invest Team
package api

import "testing"

// TestConstants использует константы API (для линтера и документации; значения заданы в constants.go).
func TestConstants(t *testing.T) {
	_ = RestURL
	_ = RpcURL
	_ = WsURL
	_ = ApiDocHost
	_ = NodeHost
	_ = TokenStandardGNDst1
	_ = ApiKey
	_ = HttpTimeout
	_ = WsTimeout
}
