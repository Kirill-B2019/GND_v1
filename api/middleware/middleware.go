// | KB @CerbeRus - Nexus Invest Team
package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// ErrorResponse структура для ответов с ошибками
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// sendError отправляет JSON ответ с ошибкой
func sendError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   http.StatusText(code),
		Code:    code,
		Message: message,
	})
}

// CORS middleware для обработки CORS заголовков
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Logger middleware для логирования запросов
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Создаем ResponseWriter для перехвата статуса ответа
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(rw, r)

		duration := time.Since(start)
		log.Printf(
			"[%s] %s %s %d %v",
			r.Method,
			r.RequestURI,
			r.RemoteAddr,
			rw.statusCode,
			duration,
		)
	})
}

// responseWriter для перехвата статуса ответа
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Recovery middleware для обработки паники
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Паника: %v", err)
				sendError(w, http.StatusInternalServerError, "Внутренняя ошибка сервера")
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// RateLimiter middleware для ограничения количества запросов
type RateLimiter struct {
	ips   map[string]*rate.Limiter
	mu    *sync.RWMutex
	rate  rate.Limit
	burst int
}

func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
	return &RateLimiter{
		ips:   make(map[string]*rate.Limiter),
		mu:    &sync.RWMutex{},
		rate:  r,
		burst: b,
	}
}

func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.ips[ip]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.ips[ip] = limiter
	}

	return limiter
}

func RateLimit(next http.Handler) http.Handler {
	limiter := NewRateLimiter(rate.Limit(100), 100) // 100 запросов в секунду, burst 100

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if !limiter.getLimiter(ip).Allow() {
			sendError(w, http.StatusTooManyRequests, "Слишком много запросов. Пожалуйста, подождите.")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Auth middleware для проверки API ключа
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			sendError(w, http.StatusUnauthorized, "Требуется API ключ")
			return
		}

		// TODO: Добавить проверку API ключа в базе данных
		// Временная заглушка для тестирования
		if apiKey != "ganymede-demo-key" {
			sendError(w, http.StatusUnauthorized, "Неверный API ключ")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// ValidateContentType middleware для проверки Content-Type
func ValidateContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" || r.Method == "PUT" {
			contentType := r.Header.Get("Content-Type")
			if contentType != "application/json" {
				sendError(w, http.StatusUnsupportedMediaType, "Требуется Content-Type: application/json")
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// Экспортируем middleware функции с правильными именами
var (
	LoggerMiddleware              = Logger
	CORSMiddleware                = CORS
	RateLimitMiddleware           = RateLimit
	AuthMiddleware                = Auth
	ValidateContentTypeMiddleware = ValidateContentType
)

func validateAPIKey(key string) bool {
	// Для примера: разрешен только ключ "ganymede-demo-key"
	return key == "ganymede-demo-key"
}
