package httpdelivery

import (
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

type RDPTunnelRequest struct {
	Host string `json:"host"`
	User string `json:"user"`
	Pass string `json:"pass"`
}

type RDPTunnelResponse struct {
	Status string `json:"status"`
	Port   int    `json:"port"`
	Error  string `json:"error,omitempty"`
}

func HandleRDPTunnel(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var host, user, pass string
	if r.Header.Get("Content-Type") == "application/json" {
		var req RDPTunnelRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
			host, user, pass = req.Host, req.User, req.Pass
		}
	} else {
		host = r.FormValue("host")
		user = r.FormValue("user")
		pass = r.FormValue("pass")
	}

	if host == "" || user == "" {
		json.NewEncoder(w).Encode(RDPTunnelResponse{Status: "error", Error: "Missing credentials"})
		return
	}

	sshClient, err := connectSSH(host, user, pass)
	if err != nil {
		json.NewEncoder(w).Encode(RDPTunnelResponse{Status: "error", Error: "SSH connection failed: " + err.Error()})
		return
	}

	listener, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		sshClient.Close()
		json.NewEncoder(w).Encode(RDPTunnelResponse{Status: "error", Error: "Failed to allocate local port"})
		return
	}

	port := listener.Addr().(*net.TCPAddr).Port

	go func() {
		defer sshClient.Close()
		defer listener.Close()

		listener.(*net.TCPListener).SetDeadline(time.Now().Add(2 * time.Minute))

		localConn, err := listener.Accept()
		if err != nil {
			log.Printf("RDP tunnel timeout or error for %s: %v", host, err)
			return
		}
		defer localConn.Close()

		listener.(*net.TCPListener).SetDeadline(time.Time{})

		remoteConn, err := sshClient.Dial("tcp", "127.0.0.1:3389")
		if err != nil {
			log.Printf("Failed to dial remote RDP port for %s: %v", host, err)
			return
		}
		defer remoteConn.Close()

		errc := make(chan error, 2)
		go func() {
			_, err := io.Copy(remoteConn, localConn)
			errc <- err
		}()
		go func() {
			_, err := io.Copy(localConn, remoteConn)
			errc <- err
		}()

		<-errc
		log.Printf("RDP tunnel closed for %s", host)
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(RDPTunnelResponse{Status: "success", Port: port})
}
