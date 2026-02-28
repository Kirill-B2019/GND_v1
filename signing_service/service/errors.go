// KB @CerbeRus - Nexus Invest Team
package service

import "errors"

var (
	// ErrWalletDisabled возвращается при запросе подписи для отключённого кошелька.
	ErrWalletDisabled = errors.New("wallet is disabled")
	// ErrWalletNotFound возвращается, когда кошелёк не найден.
	ErrWalletNotFound = errors.New("wallet not found")
)
