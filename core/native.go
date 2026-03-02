// | KB @CerbeRus - Nexus Invest Team 2026
// core/native.go — нативные монеты L1 (GND, GANI). Список символов и проверки.

package core

// NativeSymbols — символы нативных монет, которые хранятся в native_balances и изменяются только нодой.
var NativeSymbols = []string{"GND", "GANI"}

// IsNativeSymbol возвращает true, если symbol — нативная монета (GND или GANI).
func IsNativeSymbol(symbol string) bool {
	switch symbol {
	case "GND", "GANI":
		return true
	default:
		return false
	}
}

// GasSymbol — символ монеты для оплаты газа (всегда GND).
const GasSymbol = "GND"
