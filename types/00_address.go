// | KB @CerbeRus - Nexus Invest Team
package types

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

// ContractAddressPrefix — префикс адреса смарт-контракта в сети ГАНИМЕД.
// Используется для однозначного определения, что адрес относится к контракту (а не к переводу на кошелёк и т.д.).
const ContractAddressPrefix = "GNDct"

// ContractAddressSuffixLen — длина hex-суффикса адреса контракта (32 hex = 16 байт).
const ContractAddressSuffixLen = 32

// IsContractAddress возвращает true, если строка является адресом контракта:
// префикс ContractAddressPrefix + ровно ContractAddressSuffixLen hex-символов (итоговая длина 5+32=37).
func IsContractAddress(s string) bool {
	if s == "" || len(s) != len(ContractAddressPrefix)+ContractAddressSuffixLen {
		return false
	}
	if !strings.HasPrefix(s, ContractAddressPrefix) {
		return false
	}
	rest := s[len(ContractAddressPrefix):]
	if _, err := hex.DecodeString(rest); err != nil {
		return false
	}
	return len(rest) == ContractAddressSuffixLen
}

// Address represents a blockchain address
type Address string

// NewAddress creates a new address from a public key
func NewAddress(publicKey []byte) Address {
	hash := sha256.Sum256(publicKey)
	return Address(hex.EncodeToString(hash[:]))
}

// IsValid checks if the address is valid (64 hex, GN_/GND wallet-style, or GNDct+hex contract)
func (a Address) IsValid() bool {
	s := string(a)
	if s == "" {
		return false
	}
	// Классический формат: 64 hex-символа (32 байта)
	if _, err := hex.DecodeString(s); err == nil && len(s) == 64 {
		return true
	}
	// Кошельки: GN_ или GND (не GNDct) + достаточно символов
	if strings.HasPrefix(s, "GN_") && len(s) > 10 {
		return true
	}
	if strings.HasPrefix(s, "GND") && !strings.HasPrefix(s, "GNDct") && len(s) > 10 {
		return true
	}
	// Контракты: префикс GNDct + 32 hex
	if IsContractAddress(s) {
		return true
	}
	return false
}

// String returns the string representation of the address
func (a Address) String() string {
	return string(a)
}

// Bytes returns the byte representation of the address
func (a Address) Bytes() ([]byte, error) {
	if !a.IsValid() {
		return nil, errors.New("invalid address")
	}
	return hex.DecodeString(string(a))
}

// ParseAddress parses an address from a string
func ParseAddress(s string) (Address, error) {
	addr := Address(s)
	if !addr.IsValid() {
		return "", fmt.Errorf("invalid address format: %s", s)
	}
	return addr, nil
}
