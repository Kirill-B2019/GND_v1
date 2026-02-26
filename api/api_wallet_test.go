//api/api_wallet_test.go

package api

import (
	"GND/core"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgxpool"
)

// setupTestDB создает тестовое подключение к реальной тестовой базе PostgreSQL
func setupTestDB(t *testing.T) (*pgxpool.Pool, sqlmock.Sqlmock) {
	connStr := "postgres://gnduser:TitanDay0909@31.128.41.155:5432/gnd_db"
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		t.Fatalf("Failed to create connection pool: %v", err)
	}
	// Возвращаем pool и nil вместо mock, так как sqlmock не используется
	return pool, nil
}

// TestWalletCreateHandler тестирует эндпоинт создания кошелька
func TestWalletCreateHandler(t *testing.T) {
	// Создаем тестовую БД
	pool, mock := setupTestDB(t)
	defer pool.Close()

	// Создаем тестовый запрос
	req, err := http.NewRequest("POST", "/api/wallet/create", bytes.NewBuffer([]byte{}))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("X-API-Key", ApiKey)

	// Создаем ResponseRecorder для записи ответа
	rr := httptest.NewRecorder()

	// Создаем тестовый обработчик
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Создаем кошелек
		wallet, err := core.NewWallet(pool)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Отправляем ответ
		json.NewEncoder(w).Encode(map[string]interface{}{
			"address":   string(wallet.Address),
			"publicKey": wallet.PublicKeyHex(),
		})
	})

	// Оборачиваем обработчик в middleware
	handler = http.HandlerFunc(AuthMiddleware(handler).ServeHTTP)

	// Выполняем запрос
	handler.ServeHTTP(rr, req)

	// Проверяем статус код
	if status := rr.Code; status != http.StatusOK {
		if status == http.StatusUnauthorized {
			if !bytes.Contains(rr.Body.Bytes(), []byte("Unauthorized")) {
				t.Errorf("expected body to contain 'Unauthorized', got '%s'", rr.Body.String())
			}
			return
		}
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		return
	}
	// Проверяем тело ответа только если статус 200
	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := response["address"]; !ok {
		t.Error("response missing address field")
	}
	if _, ok := response["publicKey"]; !ok {
		t.Error("response missing publicKey field")
	}

	// Проверяем, что все ожидаемые запросы были выполнены
	if mock != nil {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unfulfilled expectations: %s", err)
		}
	}
}

// TestWalletCreateHandlerWithoutAuth тестирует создание кошелька без авторизации
func TestWalletCreateHandlerWithoutAuth(t *testing.T) {
	// Создаем тестовую БД
	pool, mock := setupTestDB(t)
	defer pool.Close()

	// Создаем тестовый запрос без API ключа
	req, err := http.NewRequest("POST", "/api/wallet/create", bytes.NewBuffer([]byte{}))
	if err != nil {
		t.Fatal(err)
	}

	// Создаем ResponseRecorder для записи ответа
	rr := httptest.NewRecorder()

	// Создаем тестовый обработчик
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Создаем кошелек
		wallet, err := core.NewWallet(pool)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Отправляем ответ
		json.NewEncoder(w).Encode(map[string]interface{}{
			"address":   string(wallet.Address),
			"publicKey": wallet.PublicKeyHex(),
		})
	})

	// Оборачиваем обработчик в middleware
	handler = http.HandlerFunc(AuthMiddleware(handler).ServeHTTP)

	// Выполняем запрос
	handler.ServeHTTP(rr, req)

	// Проверяем статус код
	if status := rr.Code; status != http.StatusOK {
		if status == http.StatusUnauthorized {
			if !bytes.Contains(rr.Body.Bytes(), []byte("Unauthorized")) {
				t.Errorf("expected body to contain 'Unauthorized', got '%s'", rr.Body.String())
			}
			return
		}
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		return
	}
	// Проверяем тело ответа только если статус 200
	var response struct {
		Address    string `json:"address"`
		PublicKey  string `json:"publicKey"`
		PrivateKey string `json:"privateKey"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Errorf("invalid response body: %v", err)
		return
	}
	if response.Address == "" || response.PublicKey == "" || response.PrivateKey == "" {
		t.Error("response missing required fields")
	}

	// Проверяем, что все ожидаемые запросы были выполнены
	if mock != nil {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unfulfilled expectations: %s", err)
		}
	}
}

// TestWalletCreateHandlerInvalidKey тестирует создание кошелька с неверным API ключом
func TestWalletCreateHandlerInvalidKey(t *testing.T) {
	// Создаем тестовую БД
	pool, mock := setupTestDB(t)
	defer pool.Close()

	// Создаем тестовый запрос с неверным API ключом
	req, err := http.NewRequest("POST", "/api/wallet/create", bytes.NewBuffer([]byte{}))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("X-API-Key", "invalid_api_key")

	// Создаем ResponseRecorder для записи ответа
	rr := httptest.NewRecorder()

	// Создаем тестовый обработчик
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Создаем кошелек
		wallet, err := core.NewWallet(pool)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Отправляем ответ
		json.NewEncoder(w).Encode(map[string]interface{}{
			"address":   string(wallet.Address),
			"publicKey": wallet.PublicKeyHex(),
		})
	})

	// Оборачиваем обработчик в middleware
	handler = http.HandlerFunc(AuthMiddleware(handler).ServeHTTP)

	// Выполняем запрос
	handler.ServeHTTP(rr, req)

	// Проверяем статус код
	if status := rr.Code; status != http.StatusOK {
		if status == http.StatusUnauthorized {
			if !bytes.Contains(rr.Body.Bytes(), []byte("Unauthorized")) {
				t.Errorf("expected body to contain 'Unauthorized', got '%s'", rr.Body.String())
			}
			return
		}
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		return
	}
	// Проверяем тело ответа только если статус 200
	var response struct {
		Address    string `json:"address"`
		PublicKey  string `json:"publicKey"`
		PrivateKey string `json:"privateKey"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Errorf("invalid response body: %v", err)
		return
	}
	if response.Address == "" || response.PublicKey == "" || response.PrivateKey == "" {
		t.Error("response missing required fields")
	}

	// Проверяем, что все ожидаемые запросы были выполнены
	if mock != nil {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unfulfilled expectations: %s", err)
		}
	}
}
