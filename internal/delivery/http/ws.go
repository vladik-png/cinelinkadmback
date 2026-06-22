package httpdelivery

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type WSHub struct {
	clients map[*websocket.Conn]bool
	mu      sync.Mutex
}

var LiveHub = &WSHub{
	clients: make(map[*websocket.Conn]bool),
}

func (h *WSHub) register(conn *websocket.Conn) {
	h.mu.Lock()
	h.clients[conn] = true
	h.mu.Unlock()
	log.Println("New Live WebSocket client connected")
}

func (h *WSHub) unregister(conn *websocket.Conn) {
	h.mu.Lock()
	if _, ok := h.clients[conn]; ok {
		delete(h.clients, conn)
		conn.Close()
		log.Println("Live WebSocket client disconnected")
	}
	h.mu.Unlock()
}

func (h *WSHub) Broadcast(msgType string, data interface{}) {
	payload := map[string]interface{}{
		"type": msgType,
		"data": data,
	}
	bytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal WS broadcast: %v", err)
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	for conn := range h.clients {
		err := conn.WriteMessage(websocket.TextMessage, bytes)
		if err != nil {
			conn.Close()
			delete(h.clients, conn)
		}
	}
}

func HandleLiveWebSocket(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade Live websocket: %v", err)
		return
	}

	LiveHub.register(ws)
	defer LiveHub.unregister(ws)

	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			break
		}
	}
}
