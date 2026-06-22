package httpdelivery

import (
	"log"
	"net/http"
	"os/exec"
	"runtime"
)

func SetupRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", WithCache(10, Protected(GetInstances)))
	mux.HandleFunc("/logs", Protected(GetLogs)) 
	mux.HandleFunc("/alerts", WithCache(5, Protected(GetAlerts)))
	mux.HandleFunc("/system-metrics", WithCache(5, Protected(GetMetricsForFront)))
	
	mux.HandleFunc("/report-metrics", EnableCORS(ReceiveMetricsFromAgent))
	mux.HandleFunc("/log-event", EnableCORS(HandleLogEvent))
	
	mux.HandleFunc("/start", Protected(StartInstance))
	mux.HandleFunc("/stop", Protected(StopInstance))
	mux.HandleFunc("/kamatera-instances", WithCache(10, Protected(GetKamateraInstances)))
	
	mux.HandleFunc("/kill-process", Protected(HandleKillProcess))
	
	mux.HandleFunc("/shutdown", Protected(HandleShutdown))

	mux.HandleFunc("/ssh", Protected(HandleTerminal)) 
	mux.HandleFunc("/rdp-ws", Protected(HandleRDPWebSocket))
	mux.HandleFunc("/ws/live", Protected(HandleLiveWebSocket))
	mux.HandleFunc("/upload", Protected(HandleUpload))
	mux.HandleFunc("/servers", Protected(HandleServers))

	return mux
}

func HandleShutdown(w http.ResponseWriter, r *http.Request) {
	log.Println("Received shutdown command!")
	w.WriteHeader(http.StatusOK)

	if runtime.GOOS == "windows" {
		exec.Command("cmd", "/c", "shutdown", "/s", "/t", "0").Run()
	} else {
		exec.Command("sudo", "shutdown", "-h", "now").Run()
	}
}
