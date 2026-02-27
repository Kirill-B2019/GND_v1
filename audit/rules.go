// | KB @CerbeRus - Nexus Invest Team
package audit

import (
	"strings"
)

// RuleType — тип правила
type RuleType string

const (
	RuleBlacklist RuleType = "blacklist"
	RuleWhitelist RuleType = "whitelist"
	RuleLimit     RuleType = "limit"
	RulePattern   RuleType = "pattern"
	RuleCustom    RuleType = "custom"
)

// Rule — структура одного правила аудита
type Rule struct {
	Name        string
	Type        RuleType
	Description string
	// Для blacklist/whitelist
	Addresses []string
	// Для лимитов
	MinValue uint64
	MaxValue uint64
	// Для паттернов
	Pattern string
	// Для кастомных правил
	CustomFunc func(tx interface{}) bool
	Enabled    bool
}

//TODO реализовать запись контрактов в конфиг JSON и чтение из конфига в листы

// Пример глобальных списков (могут быть загружены из конфига)
var Blacklist = []string{
	"GNDscam1...", "GNDscam2...",
}
var Whitelist = []string{
	"GNDct123...", "GNDct456...",
}

// Пример правил
var DefaultRules = []Rule{
	{
		Name:        "Blacklist addresses",
		Type:        RuleBlacklist,
		Addresses:   Blacklist,
		Description: "Запрещённые адреса для перевода средств",
		Enabled:     true,
	},
	{
		Name:        "Whitelist contracts",
		Type:        RuleWhitelist,
		Addresses:   Whitelist,
		Description: "Разрешённые контракты для вызова",
		Enabled:     true,
	},
	{
		Name:        "Limit transfer value",
		Type:        RuleLimit,
		MinValue:    1,
		MaxValue:    1_000_000_000,
		Description: "Ограничение суммы перевода",
		Enabled:     true,
	},
	{
		Name:        "Pattern: contract prefix",
		Type:        RulePattern,
		Pattern:     "GNDct",
		Description: "Адреса контрактов должны начинаться с GNDct",
		Enabled:     true,
	},
}

// Проверка адреса по blacklist
func IsBlacklisted(address string) bool {
	for _, a := range Blacklist {
		if strings.EqualFold(address, a) {
			return true
		}
	}
	return false
}

// Проверка адреса по whitelist
func IsWhitelisted(address string) bool {
	for _, a := range Whitelist {
		if strings.EqualFold(address, a) {
			return true
		}
	}
	return false
}

// Проверка суммы перевода по лимитам
func IsWithinLimits(value uint64, min uint64, max uint64) bool {
	return value >= min && value <= max
}

// Проверка по паттерну (например, префикс адреса)
func MatchesPattern(address, pattern string) bool {
	return strings.HasPrefix(address, pattern)
}
