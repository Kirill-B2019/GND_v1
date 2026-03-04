// | KB @CerbeRus - Nexus Invest Team
package types

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

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
	// Контракты: GNDct + 32 hex
	if strings.HasPrefix(s, "GNDct") && len(s) == 37 {
		rest := s[5:]
		if _, err := hex.DecodeString(rest); err == nil && len(rest) == 32 {
			return true
		}
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
