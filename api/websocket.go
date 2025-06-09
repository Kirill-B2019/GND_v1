// api/websocket.go

package api

import (
	"GND/core"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	// Время ожидания для записи сообщения
	writeWait = 10 * time.Second

	// Время ожидания для чтения следующего pong сообщения
	pongWait = 60 * time.Second

	// Период отправки ping сообщений
	pingPeriod = (pongWait * 9) / 10

	// Максимальный размер сообщения
	maxMessageSize = 512 * 1024 // 512KB
)

var (
	clients  = make(map[*Client]bool) // было: map[*websocket.Conn]bool
	hubMutex sync.RWMutex
)

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

// Hub поддерживает активные соединения и рассылку сообщений
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mutex      sync.RWMutex
}

// Client представляет подключенного клиента
type Client struct {
	hub           *Hub
	conn          *websocket.Conn
	send          chan []byte
	subscriptions map[string]bool
}

var hub = &Hub{
	clients:    make(map[*Client]bool),
	broadcast:  make(chan []byte),
	register:   make(chan *Client),
	unregister: make(chan *Client),
}

// Запуск WebSocket сервера
func StartWebSocketServer(blockchain *core.Blockchain, mempool *core.Mempool, cfg *core.Config) {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Ошибка обновления соединения: %v", err)
			return
		}
		defer conn.Close()

		// Регистрация клиента
		client := &Client{
			hub:           hub,
			conn:          conn,
			send:          make(chan []byte, 256),
			subscriptions: make(map[string]bool),
		}
		hub.register <- client

		// Запуск горутин для чтения и записи
		go client.writePump()
		go client.readPump(blockchain, mempool)
	})

	addr := cfg.Server.WS.WSAddr
	log.Printf("Запуск WebSocket сервера на %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Ошибка запуска WebSocket сервера: %v", err)
	}
}

// Обработка сообщений от клиента
func (c *Client) readPump(blockchain *core.Blockchain, mempool *core.Mempool) {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Ошибка чтения: %v", err)
			}
			break
		}

		// Обработка JSON-RPC запроса
		var request struct {
			JSONRPC string        `json:"jsonrpc"`
			ID      interface{}   `json:"id"`
			Method  string        `json:"method"`
			Params  []interface{} `json:"params"`
		}

		if err := json.Unmarshal(message, &request); err != nil {
			c.send <- []byte(`{"jsonrpc":"2.0","error":{"code":-32700,"message":"Ошибка разбора JSON"},"id":null}`)
			continue
		}

		// Обработка методов
		switch request.Method {
		case "gnd_subscribe":
			if len(request.Params) < 1 {
				c.send <- []byte(`{"jsonrpc":"2.0","error":{"code":-32602,"message":"Неверное количество параметров"},"id":null}`)
				continue
			}
			subscription := request.Params[0].(string)
			c.subscriptions[subscription] = true
			c.send <- []byte(fmt.Sprintf(`{"jsonrpc":"2.0","result":"%s","id":%v}`, subscription, request.ID))

		case "gnd_unsubscribe":
			if len(request.Params) < 1 {
				c.send <- []byte(`{"jsonrpc":"2.0","error":{"code":-32602,"message":"Неверное количество параметров"},"id":null}`)
				continue
			}
			subscription := request.Params[0].(string)
			delete(c.subscriptions, subscription)
			c.send <- []byte(fmt.Sprintf(`{"jsonrpc":"2.0","result":true,"id":%v}`, request.ID))

		default:
			c.send <- []byte(fmt.Sprintf(`{"jsonrpc":"2.0","error":{"code":-32601,"message":"Метод не найден"},"id":%v}`, request.ID))
		}
	}
}

// Отправка сообщений клиенту
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
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
		if client.conn != nil {
			err := client.conn.WriteJSON(blockchain.LatestBlock())
			if err != nil {
				log.Printf("Error sending block to client: %v", err)
				client.conn.Close()
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

func sendWebSocketError(conn *websocket.Conn, code int, message string, data interface{}) {
	response := struct {
		JSONRPC string `json:"jsonrpc"`
		Error   struct {
			Code    int         `json:"code"`
			Message string      `json:"message"`
			Data    interface{} `json:"data"`
		} `json:"error"`
		ID interface{} `json:"id"`
	}{
		JSONRPC: "2.0",
		Error: struct {
			Code    int         `json:"code"`
			Message string      `json:"message"`
			Data    interface{} `json:"data"`
		}{
			Code:    code,
			Message: message,
			Data:    data,
		},
		ID: nil,
	}

	if err := conn.WriteJSON(response); err != nil {
		log.Println("WS: ошибка отправки:", err)
		if err := conn.Close(); err != nil {
			log.Println("WS: ошибка при закрытии соединения:", err)
		}
	}
}

func sendWebSocketResponse(conn *websocket.Conn, id interface{}, result interface{}) {
	response := struct {
		JSONRPC string      `json:"jsonrpc"`
		Result  interface{} `json:"result"`
		ID      interface{} `json:"id"`
	}{
		JSONRPC: "2.0",
		Result:  result,
		ID:      id,
	}

	if err := conn.WriteJSON(response); err != nil {
		log.Println("WS: ошибка отправки:", err)
		if err := conn.Close(); err != nil {
			log.Println("WS: ошибка при закрытии соединения:", err)
		}
	}
}
