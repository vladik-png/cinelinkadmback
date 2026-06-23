package httpdelivery

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"admin-aws/internal/config"
	"admin-aws/internal/services"

	"github.com/golang-jwt/jwt/v5"
)

func getUserIDFromRequest(r *http.Request) uint {
	secret := config.GetEnv("JWT_SECRET", "")
	authHeader := r.Header.Get("Authorization")
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == "" {
		tokenString = r.URL.Query().Get("token")
	}
	
	if tokenString != "" && secret != "" {
		token, _ := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			if idFloat, ok := claims["user_id"].(float64); ok {
				return uint(idFloat)
			}
			if idFloat, ok := claims["id"].(float64); ok {
				return uint(idFloat)
			}
		}
	}
	return 1
}

func sendJSONResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"data": data})
}

func sendJSONError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]interface{}{"error": msg})
}

func HandleGetUserChats(w http.ResponseWriter, r *http.Request) {
	employeeID := getUserIDFromRequest(r)
	chats, err := services.GetUserChats(employeeID)
	if err != nil {
		sendJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sendJSONResponse(w, chats)
}

func HandleGetChatDetails(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		sendJSONError(w, "invalid path", http.StatusBadRequest)
		return
	}
	chatID, err := strconv.ParseUint(parts[2], 10, 32)
	if err != nil {
		sendJSONError(w, "invalid chat id", http.StatusBadRequest)
		return
	}
	if r.Method == "DELETE" {
		err := services.DeleteChat(uint(chatID))
		if err != nil {
			sendJSONError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		sendJSONResponse(w, map[string]string{"status": "deleted"})
		return
	}

	chat, err := services.GetChatDetails(uint(chatID))
	if err != nil {
		sendJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sendJSONResponse(w, chat)
}

func HandleGetOrCreateChat(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		sendJSONError(w, "invalid path", http.StatusBadRequest)
		return
	}
	friendID, err := strconv.ParseUint(parts[3], 10, 32)
	if err != nil {
		sendJSONError(w, "invalid friend id", http.StatusBadRequest)
		return
	}
	employeeID := getUserIDFromRequest(r)
	
	chatID, err := services.GetOrCreateChat(employeeID, uint(friendID))
	if err != nil {
		sendJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sendJSONResponse(w, chatID)
}

func HandleChatMembers(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		sendJSONError(w, "invalid path", http.StatusBadRequest)
		return
	}
	chatID, err := strconv.ParseUint(parts[2], 10, 32)
	if err != nil {
		sendJSONError(w, "invalid chat id", http.StatusBadRequest)
		return
	}

	if r.Method == "GET" {
		members, err := services.GetChatMembers(uint(chatID))
		if err != nil {
			sendJSONError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		sendJSONResponse(w, members)
	} else if r.Method == "POST" {
		var body struct {
			UserID uint `json:"user_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			sendJSONError(w, "invalid body", http.StatusBadRequest)
			return
		}
		err := services.AddChatMember(uint(chatID), body.UserID)
		if err != nil {
			sendJSONError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		sendJSONResponse(w, map[string]string{"status": "ok"})
	} else if r.Method == "DELETE" {
		if len(parts) < 5 {
			sendJSONError(w, "invalid path", http.StatusBadRequest)
			return
		}
		userID, _ := strconv.ParseUint(parts[4], 10, 32)
		err := services.RemoveChatMember(uint(chatID), uint(userID))
		if err != nil {
			sendJSONError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		sendJSONResponse(w, map[string]string{"status": "ok"})
	}
}

func HandleChatMessages(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		sendJSONError(w, "invalid path", http.StatusBadRequest)
		return
	}
	chatID, err := strconv.ParseUint(parts[2], 10, 32)
	if err != nil {
		sendJSONError(w, "invalid chat id", http.StatusBadRequest)
		return
	}

	if r.Method == "GET" {
		messages, err := services.GetChatMessages(uint(chatID))
		if err != nil {
			sendJSONError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		sendJSONResponse(w, messages)
	} else if r.Method == "POST" {
		var body struct {
			Type    string `json:"type"`
			Content struct {
				ChatID      uint   `json:"chat_id"`
				UserID      uint   `json:"user_id"`
				Message     string `json:"message"`
				MessageType string `json:"message_type"`
			} `json:"content"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			sendJSONError(w, "invalid body", http.StatusBadRequest)
			return
		}
		fmt.Println("SEND MSG:", body.Content)
		msg, err := services.SendMessage(uint(chatID), body.Content.UserID, body.Content.Message, body.Content.MessageType)
		if err != nil {
			sendJSONError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		sendJSONResponse(w, msg)
	}
}
