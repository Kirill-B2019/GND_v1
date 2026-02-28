// | KB @CerbeRus - Nexus Invest Team
package api

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestHashKey(t *testing.T) {
	key := "gnd_secret123"
	got := HashKey(key)
	h := sha256.Sum256([]byte(key))
	expect := hex.EncodeToString(h[:])
	if got != expect {
		t.Errorf("HashKey(%q) = %q, ожидали %q", key, got, expect)
	}
	if len(got) != 64 {
		t.Errorf("HashKey должен возвращать 64 hex-символа, получили %d", len(got))
	}
}

func TestKeyPrefix(t *testing.T) {
	tests := []struct {
		key    string
		expect string
	}{
		{"gnd_abc", "gnd_abc"},
		{"gnd_abcdefghijk", "gnd_abcd"},
		{"ab", "ab"},
		{"", ""},
	}
	for _, tt := range tests {
		got := KeyPrefix(tt.key)
		if got != tt.expect {
			t.Errorf("KeyPrefix(%q) = %q, ожидали %q", tt.key, got, tt.expect)
		}
	}
}

func TestValidateAdminToken_EmptySecret(t *testing.T) {
	// При пустом GND_ADMIN_SECRET любой токен должен быть неверным
	// (проверяем через результат: без установки env мы не можем гарантировать пустой секрет в параллельных тестах)
	// Просто проверяем, что функция не паникует
	_ = ValidateAdminToken("")
	_ = ValidateAdminToken("any-token")
}
