package audit

import (
	"GND/vm"
	"strings"
)

// SecurityIssue — структура для хранения информации об обнаруженной проблеме
type SecurityIssue struct {
	Title       string
	Description string
	Severity    string // "critical", "high", "medium", "low", "info"
	Location    string // адрес, имя контракта, функция и т.д.
}

// Проверка на переполнение/underflow (эмуляция для uint64)
func CheckOverflow(a, b uint64) bool {
	return a+b < a
}

// Проверка на повторный вход (reentrancy) — эмуляция: ищем вызовы внешних контрактов до изменения баланса
func CheckReentrancy(contractCode string) bool {
	// Примитивная эвристика: если есть вызов внешнего контракта до изменения storage
	return strings.Contains(contractCode, "call.value") && !strings.Contains(contractCode, "storage[") // упрощённо
}

// Проверка наличия открытых функций без модификаторов доступа
func CheckPublicFunctions(contractCode string) []string {
	var publicFuncs []string
	lines := strings.Split(contractCode, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "function") && !strings.Contains(line, "private") && !strings.Contains(line, "internal") {
			publicFuncs = append(publicFuncs, line)
		}
	}
	return publicFuncs
}

// Проверка на отсутствие проверки владельца (owner) в критических функциях
func CheckOwnerCheck(contractCode string) bool {
	// Упрощённо: ищем require(owner == msg.sender)
	return strings.Contains(contractCode, "require(owner") || strings.Contains(contractCode, "onlyOwner")
}

// Проверка на использование устаревших конструкций (например, tx.origin)
func CheckDeprecatedUsage(contractCode string) bool {
	return strings.Contains(contractCode, "tx.origin")
}

// Проверка на наличие событий (event) для важных функций
func CheckEvents(contractCode string) bool {
	return strings.Contains(contractCode, "event")
}

// Комплексная проверка контракта
func RunSecurityChecks(contract vm.Contract) []SecurityIssue {
	var issues []SecurityIssue

	code := string(contract.Bytecode())

	if CheckOverflow(1<<63, 1<<63) {
		issues = append(issues, SecurityIssue{
			Title:       "Возможное переполнение",
			Description: "Обнаружено потенциальное переполнение при сложении uint64.",
			Severity:    "high",
			Location:    string(contract.Address()),
		})
	}

	if CheckReentrancy(code) {
		issues = append(issues, SecurityIssue{
			Title:       "Возможная атака повторного входа (reentrancy)",
			Description: "В коде контракта обнаружен вызов внешнего контракта до изменения storage.",
			Severity:    "critical",
			Location:    string(contract.Address()),
		})
	}

	publicFuncs := CheckPublicFunctions(code)
	for _, f := range publicFuncs {
		issues = append(issues, SecurityIssue{
			Title:       "Открытая функция без модификатора доступа",
			Description: "Функция может быть вызвана любым адресом: " + f,
			Severity:    "medium",
			Location:    string(contract.Address()),
		})
	}

	if !CheckOwnerCheck(code) {
		issues = append(issues, SecurityIssue{
			Title:       "Нет проверки владельца в критических функциях",
			Description: "В коде отсутствует проверка owner или onlyOwner.",
			Severity:    "high",
			Location:    string(contract.Address()),
		})
	}

	if CheckDeprecatedUsage(code) {
		issues = append(issues, SecurityIssue{
			Title:       "Использование устаревших конструкций",
			Description: "В коде найдено использование tx.origin.",
			Severity:    "low",
			Location:    string(contract.Address()),
		})
	}

	if !CheckEvents(code) {
		issues = append(issues, SecurityIssue{
			Title:       "Нет событий для отслеживания действий",
			Description: "В коде отсутствуют события (event) для важных функций.",
			Severity:    "info",
			Location:    string(contract.Address()),
		})
	}

	return issues
}
