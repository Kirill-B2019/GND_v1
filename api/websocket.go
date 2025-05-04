package api

import (
	_ "encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"GND/core"
	// "ganymede/vm" // если нужно слать события контрактов
	// "ganymede/tokens"
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
		return true // Для продакшена - добавить проверку домена!
	},
}

// Запуск WebSocket-сервера
func StartWebSocketServer(bc *core.Blockchain /*, другие модули */) {
	http.HandleFunc("/ws", wsHandler)
	log.Println("WebSocket сервер запущен на /ws")
	go broadcastBlocks(bc)
	// Можно добавить broadcastTxs, broadcastEvents и т.д.
	http.ListenAndServe(":8090", nil) // порт вынести в конфиг
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
	// Можно обрабатывать входящие сообщения клиента (например, подписки)
	go wsReadLoop(conn)
}

// Чтение сообщений от клиента (например, подписка на адрес/тип событий)
func wsReadLoop(conn *websocket.Conn) {
	defer func() {
		wsManager.lock.Lock()
		delete(wsManager.clients, conn)
		wsManager.lock.Unlock()
		conn.Close()
		log.Println("WS: отключение клиента")
	}()
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
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
			conn.Close()
			delete(wsManager.clients, conn)
		}
	}
}

// Пример: периодически слать последний блок (или по событию)
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

// Можно реализовать broadcastTxs, broadcastEvents и подписки по фильтрам

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
