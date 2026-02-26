// api/auth.go

package api

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ValidateAPIKey проверяет ключ из заголовка X-API-Key.
// Если pool != nil, в будущем можно проверять ключ по таблице api_keys.
// Пока допускается ключ из константы ApiKey (для тестов и внешних систем).
func ValidateAPIKey(ctx context.Context, pool *pgxpool.Pool, key string) bool {
	if key == "" {
		return false
	}
	if key == ApiKey {
		return true
	}
	if pool != nil {
		var exists bool
		err := pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM public.api_keys WHERE key = $1 AND (expires_at IS NULL OR expires_at > NOW()))", key).Scan(&exists)
		if err == nil && exists {
			return true
		}
	}
	return false
}
