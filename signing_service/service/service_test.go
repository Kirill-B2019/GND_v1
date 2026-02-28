// | KB @CerbeRus - Nexus Invest Team
package service

import (
	"errors"
	"testing"
)

func TestErrors_Defined(t *testing.T) {
	if ErrWalletDisabled == nil {
		t.Error("ErrWalletDisabled должен быть задан")
	}
	if ErrWalletNotFound == nil {
		t.Error("ErrWalletNotFound должен быть задан")
	}
	if !errors.Is(ErrWalletDisabled, ErrWalletDisabled) {
		t.Error("ErrWalletDisabled не совпадает с собой")
	}
	if !errors.Is(ErrWalletNotFound, ErrWalletNotFound) {
		t.Error("ErrWalletNotFound не совпадает с собой")
	}
}
