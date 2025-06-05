//api/rest.go

package api

import (
	"GND/core"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"net/http"
	"strings"
	_ "sync"
)

func StartRESTServer(
	bc *core.Blockchain,
	mempool *core.Mempool,
	config *core.Config,
	pool *pgxpool.Pool,
) {
	mux := http.NewServeMux()

	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")

		// Получаем адрес ноды из конфига
		nodeAddr := config.Server.REST.RESTAddr

		// Формируем строку с описанием монет
		coinsInfo := ""
		for _, coin := range config.Coins {
			coinsInfo += fmt.Sprintf(
				"Монета: %s (%s), знаков после запятой: %d\n",
				coin.Name, coin.Symbol, coin.Decimals,
			)
			// Если есть описание, добавьте его:
			if coin.Description != "" {
				coinsInfo += fmt.Sprintf("Описание: %s\n", coin.Description)
			}
		}

		// Формируем итоговое сообщение
		msg := fmt.Sprintf(
			"Привет, это ГАНИМЕД.\nМой API версии 1.0\n"+
				"Порт node: %s\n\n"+
				"Доступные монеты:\n%s",
			nodeAddr, coinsInfo,
		)

		w.Write([]byte(msg))
	})

	mux.HandleFunc("/block/latest", func(w http.ResponseWriter, r *http.Request) {
		block := bc.LatestBlock()
		json.NewEncoder(w).Encode(block)
	})

	mux.HandleFunc("/tx/send", func(w http.ResponseWriter, r *http.Request) {
		var tx core.Transaction

		// 1. Проверка Content-Type
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Неподдерживаемый тип. JSON только", http.StatusUnsupportedMediaType)
			return
		}

		// 2. Декодирование с проверкой ошибок
		if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
			http.Error(w, fmt.Sprintf("Недопустимый JSON-файл: %v", err), http.StatusBadRequest)
			return
		}

		// 3. Валидация адресов
		if !core.ValidateAddress(tx.From) {
			http.Error(w, "Неверный адрес отправителя", http.StatusBadRequest)
			return
		}

		if !core.ValidateAddress(tx.To) {
			http.Error(w, "Неверный адрес получателя", http.StatusBadRequest)
			return
		}

		// 4. Добавление в мемпул с обработкой ошибок
		if err := mempool.Add(&tx); err != nil {
			http.Error(w, fmt.Sprintf("Транзакция отклонена: %v", err), http.StatusBadRequest)
			return
		}

		// 5. Генерация ответа
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)

		// 6. Проверка ошибок кодирования
		if err := json.NewEncoder(w).Encode(map[string]string{
			"txHash": tx.Hash,
			"status": "pending",
		}); err != nil {
			log.Printf("Не удалось закодировать ответ: %v", err)
		}
	})

	mux.HandleFunc("/api/wallet/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path[len("/api/wallet/"):]
		parts := strings.Split(path, "/")
		if len(parts) != 2 || parts[1] != "balance" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		address := parts[0]
		if !core.ValidateAddress(address) {
			http.Error(w, "invalid address", http.StatusBadRequest)
			return
		}

		balances := make([]map[string]interface{}, 0)
		for _, coin := range config.Coins {
			balance := bc.State.GetBalance(core.Address(address), coin.Symbol)
			balances = append(balances, map[string]interface{}{
				"symbol":   coin.Symbol,
				"name":     coin.Name,
				"decimals": coin.Decimals,
				"balance":  balance.String(),
			})
		}

		resp := map[string]interface{}{
			"address":  address,
			"balances": balances,
		}
		json.NewEncoder(w).Encode(resp)
	})

	// Подключение обработчика генерации кошелька с middleware
	mux.Handle("/api/wallet/create",
		AuthMiddleware( // Проверка API-ключа (см. middleware.go)
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Разрешаем только POST-запросы для безопасности
				if r.Method != http.MethodPost {
					http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
					return
				}

				// Генерируем новый кошелёк через core.NewWallet()
				wallet, err := core.NewWallet(pool)
				if err != nil {
					http.Error(w, "ошибка генерации кошелька", http.StatusInternalServerError)
					return
				}

				// Получаем публичный ключ в hex-формате (см. core.Wallet)
				pubKeyHex := wallet.PublicKeyHex()

				// Формируем ответ (только публичные данные!)
				resp := map[string]interface{}{
					"address":   wallet.Address, // Адрес кошелька
					"publicKey": pubKeyHex,      // Публичный ключ в hex-формате
					//"privateKey": wallet.PrivateKeyHex(), // Не возвращайте приватный ключ в некастодиальных решениях!
				}

				// Устанавливаем заголовок Content-Type для JSON
				w.Header().Set("Content-Type", "application/json")
				// Кодируем и отправляем ответ
				if err := json.NewEncoder(w).Encode(resp); err != nil {
					log.Printf("Ошибка кодирования ответа: %v", err)
				}
			}),
		),
	)

	addr := config.Server.REST.RESTAddr
	log.Printf("REST сервер запущен на %s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Ошибка запуска REST сервера: %v", err)
	}
}
