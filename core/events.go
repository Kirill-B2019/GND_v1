// | KB @CerbeRus - Nexus Invest Team
package core

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Типы событий блокчейна
const (
	// События первого запуска
	EventFirstLaunch    = "ПервыйЗапуск"
	EventTokenCreation  = "СозданиеТокена"
	EventInitialBalance = "НачальныйБаланс"
	EventWalletCreation = "СозданиеКошелька"
	EventGenesisBlock   = "ГенезисБлок"

	// События транзакций
	EventTransaction    = "Транзакция"
	EventTransfer       = "Перевод"
	EventContractDeploy = "ДеплойКонтракта"
	EventContractCall   = "ВызовКонтракта"

	// События ошибок
	EventError   = "Ошибка"
	EventWarning = "Предупреждение"
)

// Event представляет событие в блокчейне
type Event struct {
	Type        string         `json:"type"`
	Contract    string         `json:"contract"`
	FromAddress string         `json:"from_address,omitempty"`
	ToAddress   string         `json:"to_address,omitempty"`
	Amount      string         `json:"amount,omitempty"`
	Timestamp   time.Time      `json:"timestamp"`
	TxHash      string         `json:"tx_hash,omitempty"`
	Error       string         `json:"error,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// CreateEvent создает новое событие в блокчейне
func CreateEvent(ctx context.Context, pool *pgxpool.Pool, event *Event) error {
	metadata, err := json.Marshal(event.Metadata)
	if err != nil {
		return fmt.Errorf("ошибка сериализации метаданных: %w", err)
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO events (
			type, contract, from_address, to_address, amount,
			timestamp, tx_hash, error, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		event.Type, event.Contract, event.FromAddress, event.ToAddress,
		event.Amount, event.Timestamp, event.TxHash, event.Error, metadata)

	if err != nil {
		return fmt.Errorf("ошибка создания события: %w", err)
	}

	return nil
}

// LogFirstLaunch создает событие первого запуска блокчейна
func LogFirstLaunch(ctx context.Context, pool *pgxpool.Pool, wallet *Wallet) error {
	event := &Event{
		Type:        EventFirstLaunch,
		Contract:    "SYSTEM",
		FromAddress: string(wallet.Address),
		Timestamp:   time.Now(),
		Metadata: map[string]any{
			"version":     "1.0.0",
			"wallet":      string(wallet.Address),
			"description": "Первый запуск блокчейна ГАНИМЕД",
		},
	}

	return CreateEvent(ctx, pool, event)
}

// LogTokenCreation создает событие создания токена
func LogTokenCreation(ctx context.Context, pool *pgxpool.Pool, token *Token) error {
	event := &Event{
		Type:        EventTokenCreation,
		Contract:    token.Address,
		FromAddress: token.Owner,
		Timestamp:   time.Now(),
		Metadata: map[string]any{
			"symbol":      token.Symbol,
			"name":        token.Name,
			"decimals":    token.Decimals,
			"totalSupply": token.TotalSupply,
		},
	}

	return CreateEvent(ctx, pool, event)
}

// LogInitialBalance создает событие начисления начального баланса
func LogInitialBalance(ctx context.Context, pool *pgxpool.Pool, address string, symbol string, amount string) error {
	event := &Event{
		Type:      EventInitialBalance,
		Contract:  "SYSTEM",
		ToAddress: address,
		Amount:    amount,
		Timestamp: time.Now(),
		Metadata: map[string]any{
			"symbol": symbol,
			"amount": amount,
		},
	}

	return CreateEvent(ctx, pool, event)
}

// LogError создает событие ошибки
func LogError(ctx context.Context, pool *pgxpool.Pool, err error, metadata map[string]any) error {
	event := &Event{
		Type:      EventError,
		Contract:  "SYSTEM",
		Timestamp: time.Now(),
		Error:     err.Error(),
		Metadata:  metadata,
	}

	return CreateEvent(ctx, pool, event)
}
