package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type AuthMessage struct {
	Type string `json:"type"`
	Host string `json:"host"`
	User string `json:"user"`
	Pass string `json:"pass"`
}

const serversFile = "servers.json"

var fileMutex sync.Mutex

func connectSSH(host, user, pass string) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(pass),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	return ssh.Dial("tcp", host+":22", config)
}

func handleTerminal(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer ws.Close()

	_, authData, err := ws.ReadMessage()
	if err != nil {
		return
	}

	var auth AuthMessage
	if err := json.Unmarshal(authData, &auth); err != nil || auth.Type != "auth" {
		ws.WriteMessage(websocket.TextMessage, []byte("\r\n\x1b[31m[Error]\x1b[0m Invalid auth data\r\n"))
		return
	}

	ws.WriteMessage(websocket.TextMessage, []byte("\x1b[33mConnecting to "+auth.Host+"...\x1b[0m\r\n"))

	sshClient, err := connectSSH(auth.Host, auth.User, auth.Pass)
	if err != nil {
		ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("\r\n\x1b[31m[Error]\x1b[0m Connection failed: %v\r\n", err)))
		return
	}
	defer sshClient.Close()

	session, err := sshClient.NewSession()
	if err != nil {
		return
	}
	defer session.Close()

	modes := ssh.TerminalModes{ssh.ECHO: 1, ssh.TTY_OP_ISPEED: 14400, ssh.TTY_OP_OSPEED: 14400}
	session.RequestPty("xterm-256color", 40, 80, modes)

	stdin, _ := session.StdinPipe()
	stdout, _ := session.StdoutPipe()
	stderr, _ := session.StderrPipe()
	session.Shell()

	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := stdout.Read(buf)
			if err != nil {
				return
			}
			ws.WriteMessage(websocket.TextMessage, buf[:n])
		}
	}()

	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := stderr.Read(buf)
			if err != nil {
				return
			}
			ws.WriteMessage(websocket.TextMessage, buf[:n])
		}
	}()

	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			break
		}
		stdin.Write(message)
	}
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == "OPTIONS" {
		return
	}

	r.ParseMultipartForm(50 << 20)

	host := r.FormValue("host")
	user := r.FormValue("user")
	pass := r.FormValue("pass")
	remoteDir := r.FormValue("remoteDir")

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "File error", http.StatusBadRequest)
		return
	}
	defer file.Close()

	sshClient, err := connectSSH(host, user, pass)
	if err != nil {
		http.Error(w, "SSH error", http.StatusInternalServerError)
		return
	}
	defer sshClient.Close()

	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		http.Error(w, "SFTP error", http.StatusInternalServerError)
		return
	}
	defer sftpClient.Close()

	dstFile, err := sftpClient.Create(remoteDir + header.Filename)
	if err != nil {
		http.Error(w, "File creation failed", http.StatusInternalServerError)
		return
	}
	defer dstFile.Close()

	io.Copy(dstFile, file)
	w.WriteHeader(http.StatusOK)
}

func handleServers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		return
	}

	fileMutex.Lock()
	defer fileMutex.Unlock()

	if r.Method == "GET" {
		data, err := os.ReadFile(serversFile)
		if err != nil {
			if os.IsNotExist(err) {
				w.Write([]byte("[]"))
				return
			}
			http.Error(w, "Error reading file", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
		return
	}

	if r.Method == "POST" {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		err = os.WriteFile(serversFile, body, 0600)
		if err != nil {
			http.Error(w, "Error writing file", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}
}

func main() {
	http.HandleFunc("/ssh", handleTerminal)
	http.HandleFunc("/upload", handleUpload)
	http.HandleFunc("/servers", handleServers)
	fmt.Println("Proxy running on :8085")
	log.Fatal(http.ListenAndServe(":8085", nil))
}
