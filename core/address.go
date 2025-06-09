// core/address.go

package core

// Address представляет адрес в блокчейне
type Address string

// String возвращает строковое представление адреса
func (a Address) String() string {
	return string(a)
}

// IsValid проверяет валидность адреса
func (a Address) IsValid() bool {
	return len(a) == 42 && a[:2] == "0x"
}
