// | KB @CerbeRus - Nexus Invest Team
// api/auth.go

package api

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ValidateAPIKey проверяет ключ из заголовка X-API-Key.
// Поддерживается: запись в api_keys по полю key (legacy), по key_hash (SHA-256 hex).
// Отключённые ключи (disabled = true) и просроченные не принимаются.
func ValidateAPIKey(ctx context.Context, pool *pgxpool.Pool, key string) bool {
	if key == "" || pool == nil {
		return false
	}
	// Legacy: ключ в открытом виде в колонке key
	var exists bool
	err := pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM public.api_keys
			WHERE key = $1 AND COALESCE(disabled, false) = false
			AND (expires_at IS NULL OR expires_at > NOW()))`, key).Scan(&exists)
	if err == nil && exists {
		return true
	}
	// Новый формат: ключ проверяется по хешу
	hash := sha256.Sum256([]byte(key))
	hashHex := hex.EncodeToString(hash[:])
	err = pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM public.api_keys
			WHERE key_hash = $1 AND COALESCE(disabled, false) = false
			AND (expires_at IS NULL OR expires_at > NOW()))`, hashHex).Scan(&exists)
	return err == nil && exists
}

// HashKey возвращает SHA-256 хеш ключа в hex (для сохранения в api_keys.key_hash).
func HashKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}

// KeyPrefix возвращает префикс ключа для отображения (первые 8 символов или до 8).
func KeyPrefix(key string) string {
	const prefixLen = 8
	if len(key) <= prefixLen {
		return key
	}
	return key[:prefixLen]
}

// ValidateAdminToken проверяет заголовок X-Admin-Token против GND_ADMIN_SECRET (constant-time).
// Если GND_ADMIN_SECRET не задан, возвращает false.
func ValidateAdminToken(headerToken string) bool {
	secret := os.Getenv("GND_ADMIN_SECRET")
	if secret == "" || headerToken == "" {
		return false
	}
	return subtleEqual(headerToken, secret)
}

func subtleEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
