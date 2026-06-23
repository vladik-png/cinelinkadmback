package httpdelivery

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"admin-aws/internal/database"
	"admin-aws/internal/models"
	"admin-aws/internal/services"

	"github.com/gorilla/websocket"
	"gorm.io/gorm/clause"
)

type WSHub struct {
	clients map[*websocket.Conn]uint
	mu      sync.RWMutex
}

var LiveHub = &WSHub{
	clients: make(map[*websocket.Conn]uint),
}

func (h *WSHub) register(conn *websocket.Conn) {
	h.mu.Lock()
	h.clients[conn] = 0
	h.mu.Unlock()
	log.Println("New Live WebSocket client connected")
}

func (h *WSHub) unregister(conn *websocket.Conn) {
	h.mu.Lock()
	employeeID := h.clients[conn]
	if _, ok := h.clients[conn]; ok {
		delete(h.clients, conn)
		conn.Close()
		log.Println("Live WebSocket client disconnected")
	}
	h.mu.Unlock()

	if employeeID != 0 {
		status := models.EmployeeStatus{
			EmployeeID: employeeID,
			IsOnline:   false,
			LastSeen:   time.Now().UTC(),
		}
		database.DB.Clauses(clause.OnConflict{UpdateAll: true}).Create(&status)

		h.Broadcast("user_status", map[string]interface{}{
			"employee_id": employeeID,
			"is_online":   false,
			"last_seen":   status.LastSeen.Format(time.RFC3339),
		})
	}
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

	h.mu.RLock()
	defer h.mu.RUnlock()
	for conn := range h.clients {
		err := conn.WriteMessage(websocket.TextMessage, bytes)
		if err != nil {
			go h.unregister(conn)
		}
	}
}

func (h *WSHub) SendToEmployee(employeeID uint, msgType string, data interface{}) {
	payload := map[string]interface{}{
		"type": msgType,
		"data": data,
	}
	bytes, err := json.Marshal(payload)
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()
	for conn, eid := range h.clients {
		if eid == employeeID {
			conn.WriteMessage(websocket.TextMessage, bytes)
		}
	}
}

type WSMessage struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
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
		_, msg, err := ws.ReadMessage()
		if err != nil {
			break
		}

		var payload WSMessage
		if err := json.Unmarshal(msg, &payload); err != nil {
			continue
		}

		switch payload.Type {
		case "auth":
			if eIDFloat, ok := payload.Data["employee_id"].(float64); ok {
				employeeID := uint(eIDFloat)
				LiveHub.mu.Lock()
				LiveHub.clients[ws] = employeeID
				LiveHub.mu.Unlock()

				status := models.EmployeeStatus{
					EmployeeID: employeeID,
					IsOnline:   true,
					LastSeen:   time.Now().UTC(),
				}
				database.DB.Clauses(clause.OnConflict{UpdateAll: true}).Create(&status)

				LiveHub.Broadcast("user_status", map[string]interface{}{
					"employee_id": employeeID,
					"is_online":   true,
				})
			}
		case "typing":
			chatIDFloat, _ := payload.Data["chat_id"].(float64)
			isTyping, _ := payload.Data["is_typing"].(bool)
			
			LiveHub.mu.RLock()
			senderID := LiveHub.clients[ws]
			LiveHub.mu.RUnlock()

			if senderID != 0 && chatIDFloat != 0 {
				members, err := services.GetChatMembers(uint(chatIDFloat))
				if err == nil {
					for _, m := range members {
						if m.EmployeeID != senderID {
							LiveHub.SendToEmployee(m.EmployeeID, "typing", map[string]interface{}{
								"chat_id":     uint(chatIDFloat),
								"employee_id": senderID,
								"is_typing":   isTyping,
							})
						}
					}
				}
			}
		case "seen":
			chatIDFloat, _ := payload.Data["chat_id"].(float64)
			msgIDFloat, _ := payload.Data["message_id"].(float64)

			LiveHub.mu.RLock()
			senderID := LiveHub.clients[ws]
			LiveHub.mu.RUnlock()

			if senderID != 0 && chatIDFloat != 0 && msgIDFloat != 0 {
				chatID := uint(chatIDFloat)
				msgID := uint(msgIDFloat)
				
				database.DB.Model(&models.CorporateChatMember{}).
					Where("chat_id = ? AND employee_id = ?", chatID, senderID).
					Update("last_seen_message_id", msgID)

				members, err := services.GetChatMembers(chatID)
				if err == nil {
					for _, m := range members {
						if m.EmployeeID != senderID {
							LiveHub.SendToEmployee(m.EmployeeID, "seen_update", map[string]interface{}{
								"chat_id":     chatID,
								"employee_id": senderID,
								"message_id":  msgID,
							})
						}
					}
				}
			}
		}
	}
}
