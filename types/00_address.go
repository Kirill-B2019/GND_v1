// | KB @CerbeRus - Nexus Invest Team
package types

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
)

// Address represents a blockchain address
type Address string

// NewAddress creates a new address from a public key
func NewAddress(publicKey []byte) Address {
	hash := sha256.Sum256(publicKey)
	return Address(hex.EncodeToString(hash[:]))
}

// IsValid checks if the address is valid
func (a Address) IsValid() bool {
	// Check if the address is a valid hex string
	_, err := hex.DecodeString(string(a))
	return err == nil && len(a) == 64 // 32 bytes = 64 hex characters
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
