package httpdelivery

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type RDPTunnelRequest struct {
	Host string `json:"host"`
	User string `json:"user"`
	Pass string `json:"pass"`
}

func HandleRDPWebSocket(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade RDP websocket: %v", err)
		return
	}
	defer ws.Close()

	_, authData, err := ws.ReadMessage()
	if err != nil {
		return
	}

	var auth AuthMessage
	if err := json.Unmarshal(authData, &auth); err != nil || auth.Host == "" {
		var rdpReq RDPTunnelRequest
		if err := json.Unmarshal(authData, &rdpReq); err == nil && rdpReq.Host != "" {
			auth.Host = rdpReq.Host
			auth.User = rdpReq.User
			auth.Pass = rdpReq.Pass
		} else {
			ws.WriteMessage(websocket.TextMessage, []byte(`{"error": "Invalid auth data"}`))
			return
		}
	}

	sshClient, err := connectSSH(auth.Host, auth.User, auth.Pass)
	if err != nil {
		ws.WriteMessage(websocket.TextMessage, []byte(`{"error": "SSH connection failed"}`))
		return
	}
	defer sshClient.Close()

	remoteConn, err := sshClient.Dial("tcp", "127.0.0.1:3389")
	if err != nil {
		ws.WriteMessage(websocket.TextMessage, []byte(`{"error": "Failed to dial remote RDP port"}`))
		return
	}
	defer remoteConn.Close()

	ws.WriteMessage(websocket.TextMessage, []byte(`{"status": "connected"}`))

	var wsMutex sync.Mutex
	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			msgType, msg, err := ws.ReadMessage()
			if err != nil {
				break
			}
			if msgType == websocket.BinaryMessage || msgType == websocket.TextMessage {
				_, err = remoteConn.Write(msg)
				if err != nil {
					break
				}
			}
		}
	}()

	go func() {
		buf := make([]byte, 32768)
		for {
			n, err := remoteConn.Read(buf)
			if err != nil {
				break
			}
			
			wsMutex.Lock()
			err = ws.WriteMessage(websocket.BinaryMessage, buf[:n])
			wsMutex.Unlock()
			
			if err != nil {
				break
			}
		}
		ws.Close()
	}()

	<-done
	log.Printf("RDP WebSocket session closed for %s", auth.Host)
}
