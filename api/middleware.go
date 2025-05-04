package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ====== АУТЕНТИФИКАЦИЯ ======

// AuthMiddleware реализует проверку API-ключа или токена
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-Key")
		if !validateAPIKey(apiKey) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Простая проверка API-ключа (можно заменить на JWT или OAuth2)
func validateAPIKey(key string) bool {
	// Для примера: разрешен только ключ "ganymede-demo-key"
	return key == "ganymede-demo-key"
}

// ====== ЛИМИТИРОВАНИЕ ======

// RateLimiter реализует лимитирование по IP/токену (простая версия)
type RateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int           // лимит запросов
	interval time.Duration // окно времени
}

// Новый лимитер (например, 100 запросов в 1 минуту)
func NewRateLimiter(limit int, interval time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		interval: interval,
	}
}

// Middleware для лимитирования
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := rl.key(r)
		rl.mu.Lock()
		now := time.Now()
		// Оставляем только свежие запросы
		reqs := rl.requests[key]
		var recent []time.Time
		for _, t := range reqs {
			if now.Sub(t) < rl.interval {
				recent = append(recent, t)
			}
		}
		// Проверяем лимит
		if len(recent) >= rl.limit {
			rl.mu.Unlock()
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		// Добавляем новый запрос
		rl.requests[key] = append(recent, now)
		rl.mu.Unlock()
		next.ServeHTTP(w, r)
	})
}

// Ключ для лимитирования - по IP или токену
func (rl *RateLimiter) key(r *http.Request) string {
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

// ====== АУДИТ ======

// AuditEntry - структура записи аудита
type AuditEntry struct {
	Time      time.Time `json:"time"`
	Method    string    `json:"method"`
	Path      string    `json:"path"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	Status    int       `json:"status"`
}

// AuditLogger - простой аудит-логгер (можно заменить на запись в БД/файл)
type AuditLogger struct {
	mu   sync.Mutex
	logs []AuditEntry
}

// Новый аудит-логгер
func NewAuditLogger() *AuditLogger {
	return &AuditLogger{
		logs: make([]AuditEntry, 0, 1000),
	}
}

// Middleware для аудита
func (al *AuditLogger) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Обертка для получения кода ответа
		lrw := &loggingResponseWriter{ResponseWriter: w, status: 200}
		next.ServeHTTP(lrw, r)
		entry := AuditEntry{
			Time:      time.Now(),
			Method:    r.Method,
			Path:      r.URL.Path,
			IP:        r.RemoteAddr,
			UserAgent: r.UserAgent(),
			Status:    lrw.status,
		}
		al.mu.Lock()
		al.logs = append(al.logs, entry)
		al.mu.Unlock()
		// Можно также отправлять в систему мониторинга или писать в файл
		log.Printf("[AUDIT] %s %s %s %d", entry.IP, entry.Method, entry.Path, entry.Status)
	})
}

// loggingResponseWriter - для перехвата статуса ответа
type loggingResponseWriter struct {
	http.ResponseWriter
	status int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.status = code
	lrw.ResponseWriter.WriteHeader(code)
}

// Экспорт аудит-лога (например, для админки)
func (al *AuditLogger) ExportJSON() ([]byte, error) {
	al.mu.Lock()
	defer al.mu.Unlock()
	return json.Marshal(al.logs)
}
