// | KB @CerbeRus - Nexus Invest Team
package integration

import (
	"testing"
)

func TestValidateAddress(t *testing.T) {
	if ValidateAddress("") {
		t.Error("ожидали false для пустого адреса")
	}
	if ValidateAddress("invalid") {
		t.Error("ожидали false для адреса без префикса GND/GNDct")
	}
}

func TestNewAddress_EmptyKey(t *testing.T) {
	_, err := NewAddress(nil)
	if err == nil {
		t.Error("ожидали ошибку для nil public key")
	}
	_, err = NewAddress([]byte{})
	if err == nil {
		t.Error("ожидали ошибку для пустого public key")
	}
}
