// api/websocket.go

package api

import (
	"GND/core"
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"net/http"
	"sync"
	"time"
)

var (
	clients  = make(map[*Client]bool) // было: map[*websocket.Conn]bool
	hubMutex sync.RWMutex
)

type Client struct {
	Conn *websocket.Conn
	// Дополнительные поля, например:
	Addr          string
	Subscriptions map[string]bool // например, подписка на адреса
}

// Структура сообщения для клиента
type WSMessage struct {
	Type string      `json:"type"` // "block", "tx", "event"
	Data interface{} `json:"data"`
}

// Менеджер всех подключений
type WSManager struct {
	clients map[*websocket.Conn]bool
	lock    sync.Mutex
}

var wsManager = &WSManager{
	clients: make(map[*websocket.Conn]bool),
}

// Upgrader для WebSocket
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// В продакшене реализуйте строгую проверку домена!
		return true
	},
}

// Запуск WebSocket-сервера с портом из конфига
func StartWebSocketServer(blockchain *core.Blockchain, addr string) {
	http.HandleFunc("/ws", wsHandler)
	log.Printf("WebSocket сервер запущен на %s", addr)
	go broadcastBlocks(blockchain)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Ошибка запуска WebSocket сервера: %v", err)
	}
}

// Обработчик подключения клиента
func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Ошибка апгрейда WS:", err)
		return
	}

	client := &Client{
		Conn:          conn,
		Addr:          conn.RemoteAddr().String(),
		Subscriptions: make(map[string]bool),
	}

	hubMutex.Lock()
	clients[client] = true
	hubMutex.Unlock()

	log.Println("WS: новое подключение")
	go wsReadLoop(client)
}

// Чтение сообщений от клиента (например, подписка на адрес/тип событий)
func wsReadLoop(client *Client) {
	defer func() {
		hubMutex.Lock()
		delete(clients, client)
		hubMutex.Unlock()

		if err := client.Conn.Close(); err != nil {
			log.Printf("WS: ошибка при закрытии соединения: %v", err)
		}
		log.Println("WS: отключение клиента")
	}()
	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			log.Printf("Ошибка чтения сообщения: %v", err)
			break
		}
		fmt.Printf("Получено сообщение: %s\n", message)
	}
}

// Отправка сообщения всем клиентам
func wsBroadcast(msg WSMessage) {
	wsManager.lock.Lock()
	defer wsManager.lock.Unlock()
	for conn := range wsManager.clients {
		if err := conn.WriteJSON(msg); err != nil {
			log.Println("WS: ошибка отправки:", err)
			if err := conn.Close(); err != nil {
				log.Println("WS: ошибка при закрытии соединения:", err)
			}
			delete(wsManager.clients, conn)
		}
	}
}

// Периодически слать последний блок (или по событию)
func broadcastBlocks(blockchain *core.Blockchain) {
	if blockchain == nil {
		log.Println("Blockchain is nil")
		return
	}

	hubMutex.RLock()
	defer hubMutex.RUnlock()

	for client := range clients {
		if client.Conn != nil {
			err := client.Conn.WriteJSON(blockchain.LatestBlock())
			if err != nil {
				log.Printf("Error sending block to client: %v", err)
				client.Conn.Close()
				delete(clients, client)
			}
		}
	}
}

// Пример отправки новой транзакции в WebSocket
func NotifyNewTx(tx *core.Transaction) {
	msg := WSMessage{Type: "tx", Data: tx}
	wsBroadcast(msg)
}

// Пример отправки события контракта
func NotifyContractEvent(event interface{}) {
	msg := WSMessage{Type: "event", Data: event}
	wsBroadcast(msg)
}
func SaveContractEventToDB(ctx context.Context, pool *pgxpool.Pool, event WSMessage) error {
	txTimestamp, ok := event.Data.(map[string]interface{})["time"].(time.Time)
	if !ok {
		return fmt.Errorf("неверный формат времени")
	}

	_, err := pool.Exec(ctx, `
		INSERT INTO logs (
			tx_id, tx_timestamp, contract_id, event, data, timestamp
		) VALUES ($1, $2, $3, $4, $5, $6)`,
		event.Data.(map[string]interface{})["txHash"], // Используйте ID или Hash
		txTimestamp,
		event.Data.(map[string]interface{})["contract"],
		event.Type,
		event.Data,
		txTimestamp)

	return err
}
