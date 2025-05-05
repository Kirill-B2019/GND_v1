package tokens

import (
	"errors"
)

// TokenStandard определяет поддерживаемые стандарты токенов
type TokenStandard string

const (
	StandardERC20  TokenStandard = "erc20"
	StandardTRC20  TokenStandard = "trc20"
	StandardCustom TokenStandard = "custom"
)

// TokenMeta содержит метаданные токена
type TokenMeta struct {
	Address     string        // Уникальный адрес токена/контракта
	Owner       string        // Владелец токена
	Standard    TokenStandard // Стандарт: erc20, trc20, custom
	Name        string        // Название токена
	Symbol      string        // Символ токена
	Decimals    uint8         // Количество знаков после запятой
	Description string        // Описание
}

// TokenInterface - универсальный интерфейс для токенов (ERC-20, TRC-20, custom)
type TokenInterface interface {
	Meta() TokenMeta

	Name() string
	Symbol() string
	Decimals() uint8
	TotalSupply() uint64
	BalanceOf(address string) uint64
	Transfer(from, to string, amount uint64) error

	// Для ERC20/TRC20
	Approve(owner, spender string, amount uint64) error
	Allowance(owner, spender string) uint64
	TransferFrom(spender, from, to string, amount uint64) error

	// Для кастомных токенов
	CustomMethod(method string, args ...interface{}) (interface{}, error)
}

// Базовая реализация кастомного токена (может быть расширена)
type BaseToken struct {
	meta TokenMeta
}

func (t *BaseToken) Meta() TokenMeta {
	return t.meta
}

func (t *BaseToken) Name() string {
	return t.meta.Name
}

func (t *BaseToken) Symbol() string {
	return t.meta.Symbol
}

func (t *BaseToken) Decimals() uint8 {
	return t.meta.Decimals
}

// Для кастомных токенов эти методы должны быть реализованы в конкретном типе
func (t *BaseToken) TotalSupply() uint64 {
	return 0
}
func (t *BaseToken) BalanceOf(address string) uint64 {
	return 0
}
func (t *BaseToken) Transfer(from, to string, amount uint64) error {
	return errors.New("not implemented")
}
func (t *BaseToken) Approve(owner, spender string, amount uint64) error {
	return errors.New("not implemented")
}
func (t *BaseToken) Allowance(owner, spender string) uint64 {
	return 0
}
func (t *BaseToken) TransferFrom(spender, from, to string, amount uint64) error {
	return errors.New("not implemented")
}
func (t *BaseToken) CustomMethod(method string, args ...interface{}) (interface{}, error) {
	return nil, errors.New("not implemented")
}
