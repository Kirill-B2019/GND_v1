// | KB @CerbeRus - Nexus Invest Team
// tokens/types.go

package tokens

import (
	"encoding/json"
	"errors"
	"math/big"
	"strings"
)

type TokenInfo struct {
	Name        string   `json:"name"`
	Symbol      string   `json:"symbol"`
	Decimals    uint8    `json:"decimals"`
	TotalSupply *big.Int `json:"totalSupply"`
	Address     string   `json:"address"`
	Standard    string   `json:"standard"`
	Bytecode    string   `json:"bytecode,omitempty"`
}

// Validate проверяет корректность данных токена
func (t *TokenInfo) Validate() error {
	if t.Name == "" {
		return errors.New("имя токена не может быть пустым")
	}
	if len(t.Name) > 32 {
		return errors.New("имя токена слишком длинное")
	}
	if t.Symbol == "" {
		return errors.New("символ токена не может быть пустым")
	}
	if len(t.Symbol) > 10 {
		return errors.New("символ токена слишком длинный")
	}
	if t.Decimals > 18 {
		return errors.New("количество десятичных знаков не может быть больше 18")
	}
	if t.TotalSupply == nil {
		return errors.New("totalSupply не может быть nil")
	}
	if t.TotalSupply.Sign() < 0 {
		return errors.New("totalSupply не может быть отрицательным")
	}
	if t.Address == "" {
		return errors.New("адрес токена не может быть пустым")
	}
	if !strings.HasPrefix(t.Address, "GNDct") {
		return errors.New("некорректный формат адреса токена")
	}
	return nil
}

// MarshalJSON реализует кастомную сериализацию для big.Int
func (t *TokenInfo) MarshalJSON() ([]byte, error) {
	type Alias TokenInfo
	return json.Marshal(&struct {
		TotalSupply string `json:"totalSupply"`
		Standard    string `json:"standard"`
		*Alias
	}{
		TotalSupply: t.TotalSupply.String(),
		Standard:    t.Standard,
		Alias:       (*Alias)(t),
	})
}

// UnmarshalJSON реализует кастомную десериализацию для big.Int
func (t *TokenInfo) UnmarshalJSON(data []byte) error {
	type Alias TokenInfo
	aux := &struct {
		TotalSupply string `json:"totalSupply"`
		Standard    string `json:"standard"`
		*Alias
	}{
		Alias: (*Alias)(t),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	t.TotalSupply = new(big.Int)
	t.TotalSupply.SetString(aux.TotalSupply, 10)
	t.Standard = aux.Standard
	return nil
}

type TransferEvent struct {
	From   string   `json:"from"`
	To     string   `json:"to"`
	Value  *big.Int `json:"value"`
	Token  string   `json:"token"`
	TxHash string   `json:"txHash"`
	Block  uint64   `json:"block"`
}

// Validate проверяет корректность данных события перевода
func (t *TransferEvent) Validate() error {
	if t.From == "" {
		return errors.New("адрес отправителя не может быть пустым")
	}
	if t.To == "" {
		return errors.New("адрес получателя не может быть пустым")
	}
	if t.Value == nil {
		return errors.New("значение перевода не может быть nil")
	}
	if t.Value.Sign() <= 0 {
		return errors.New("значение перевода должно быть положительным")
	}
	if t.Token == "" {
		return errors.New("адрес токена не может быть пустым")
	}
	if t.TxHash == "" {
		return errors.New("хеш транзакции не может быть пустым")
	}
	return nil
}
