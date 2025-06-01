package utils

import (
	"math/rand"
	_ "strings"
	"time"
)

// Возвращает случайный вариант "GN" с разным регистром букв
func randomCaseGN() string {
	rand.Seed(time.Now().UnixNano())
	options := []string{"GN_", "Gn_", "gN_", "gn_"}
	return options[rand.Intn(len(options))]
}
