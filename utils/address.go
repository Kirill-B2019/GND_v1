package utils

import (
	"math/rand"
	"strings"
	"time"
)

// AddPrefix добавляет префикс "GN" с разным регистром букв
func AddPrefix(addr string) string {
	if hasAnyPrefix(addr, []string{"Gn_", "Gn_", "gN_", "gn_"}) {
		return addr // уже есть префикс в любом регистре
	}
	return randomCaseGN() + addr
}

// RemovePrefix удаляет префикс "GN" в любом регистре, если он есть
func RemovePrefix(addr string) string {
	prefixes := []string{"Gn_", "Gn_", "gN_", "gn_"}
	for _, p := range prefixes {
		if strings.HasPrefix(addr, p) {
			return addr[len(p):]
		}
	}
	return addr
}

// Проверяет наличие любого из префиксов в строке
func hasAnyPrefix(s string, prefixes []string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}

// Возвращает случайный вариант "GN" с разным регистром букв
func randomCaseGN() string {
	rand.Seed(time.Now().UnixNano())
	options := []string{"GN_", "Gn_", "gN_", "gn_"}
	return options[rand.Intn(len(options))]
}
