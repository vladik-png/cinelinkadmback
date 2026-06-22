package httpdelivery

import (
	"log"
	"net/http"
	"os/exec"
	"runtime"
)

func SetupRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", EnableCORS(GetInstances))
	mux.HandleFunc("/logs", EnableCORS(GetLogs))
	mux.HandleFunc("/alerts", EnableCORS(GetAlerts))
	mux.HandleFunc("/system-metrics", EnableCORS(GetMetricsForFront))
	mux.HandleFunc("/report-metrics", EnableCORS(ReceiveMetricsFromAgent))
	mux.HandleFunc("/log-event", EnableCORS(HandleLogEvent))
	
	mux.HandleFunc("/start", EnableCORS(StartInstance))
	mux.HandleFunc("/stop", EnableCORS(StopInstance))
	mux.HandleFunc("/kamatera-instances", EnableCORS(GetKamateraInstances))
	
	mux.HandleFunc("/kill-process", EnableCORS(HandleKillProcess))
	
	mux.HandleFunc("/shutdown", EnableCORS(HandleShutdown))

	mux.HandleFunc("/ssh", HandleTerminal) 
	mux.HandleFunc("/rdp-ws", HandleRDPWebSocket)
	mux.HandleFunc("/ws/live", HandleLiveWebSocket)
	mux.HandleFunc("/upload", HandleUpload)
	mux.HandleFunc("/servers", HandleServers)

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
