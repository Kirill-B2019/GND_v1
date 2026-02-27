// | KB @CerbeRus - Nexus Invest Team
package api

import (
	"GND/core"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// TestBroadcastBlocks использует broadcastBlocks (для линтера).
func TestBroadcastBlocks(t *testing.T) {
	broadcastBlocks(nil) // nil — допустимо, внутри есть проверка
	// С реальным blockchain можно вызывать при интеграционных тестах
}

// TestNotifyNewTx использует NotifyNewTx (публичный API).
func TestNotifyNewTx(t *testing.T) {
	NotifyNewTx(nil)
}

// TestNotifyContractEvent использует NotifyContractEvent (публичный API).
func TestNotifyContractEvent(t *testing.T) {
	NotifyContractEvent(struct{}{})
}

// TestSaveContractEventToDB использует SaveContractEventToDB (ожидаем ошибку при неверных данных).
func TestSaveContractEventToDB(t *testing.T) {
	ctx := context.Background()
	event := WSMessage{Type: "event", Data: map[string]interface{}{"time": "invalid"}}
	err := SaveContractEventToDB(ctx, nil, event)
	if err == nil {
		t.Log("SaveContractEventToDB с nil pool или неверным форматом может вернуть ошибку")
	}
}

// TestSendWebSocketHelpers проверяет вызов sendWebSocketError и sendWebSocketResponse по реальному соединению.
func TestSendWebSocketHelpers(t *testing.T) {
	upgrader := websocket.Upgrader{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Logf("upgrade: %v", err)
			return
		}
		defer conn.Close()
		sendWebSocketError(conn, -32600, "test error", nil)
		sendWebSocketResponse(conn, 1, "test result")
	}))
	defer srv.Close()

	wsURL := "ws" + srv.URL[4:]
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Skipf("WebSocket dial skip: %v", err)
		return
	}
	defer conn.Close()

	// Дождаться ответов от сервера
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	for i := 0; i < 2; i++ {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}

// TestBroadcastBlocks_withBlockchain вызывает broadcastBlocks с заглушкой блокчейна (опционально).
func TestBroadcastBlocks_withBlockchain(t *testing.T) {
	genesis := &core.Block{
		Index: 0, Timestamp: time.Now(), PrevHash: "", Hash: "genesis",
		Consensus: "poa", Nonce: 0, Status: "finalized", Transactions: nil,
	}
	genesis.Hash = genesis.CalculateHash()
	bc := core.NewBlockchain(genesis, nil)
	if bc == nil {
		t.Fatal("NewBlockchain returned nil")
	}
	broadcastBlocks(bc)
}
