package api

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"GND/core"
	"github.com/gorilla/websocket"
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

// Запуск WebSocket-сервера с портом из конфига
func StartWebSocketServer(bc *core.Blockchain, config *core.Config) {
	http.HandleFunc("/ws", wsHandler)
	log.Printf("WebSocket сервер запущен на /ws (порт %d)", config.WsPort)
	go broadcastBlocks(bc)
	addr := fmt.Sprintf(":%d", config.WsPort)
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
	wsManager.lock.Lock()
	wsManager.clients[conn] = true
	wsManager.lock.Unlock()
	log.Println("WS: новое подключение")
	go wsReadLoop(conn)
}

// Чтение сообщений от клиента (например, подписка на адрес/тип событий)
func wsReadLoop(conn *websocket.Conn) {
	defer func() {
		wsManager.lock.Lock()
		delete(wsManager.clients, conn)
		wsManager.lock.Unlock()
		if err := conn.Close(); err != nil {
			log.Printf("WS: ошибка при закрытии соединения: %v", err)
		}
		log.Println("WS: отключение клиента")
	}()
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WS: ошибка чтения сообщения: %v", err)
			}
			break
		}
		// Здесь можно обработать команды подписки/фильтрации
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
func broadcastBlocks(bc *core.Blockchain) {
	lastHash := ""
	for {
		block := bc.LatestBlock()
		if block.Hash != lastHash {
			msg := WSMessage{Type: "block", Data: block}
			wsBroadcast(msg)
			lastHash = block.Hash
		}
		time.Sleep(2 * time.Second)
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
