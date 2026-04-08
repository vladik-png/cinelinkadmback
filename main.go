package main

import (
	"log"
	"net/http"
)

func main() {
	initConfig()
	initDB()

	go monitorServers()

	http.HandleFunc("/", enableCORS(getInstances))
	http.HandleFunc("/logs", enableCORS(getLogs))
	http.HandleFunc("/alerts", enableCORS(getAlerts))
	http.HandleFunc("/system-metrics", enableCORS(getMetricsForFront))
	http.HandleFunc("/report-metrics", enableCORS(receiveMetricsFromAgent))
	http.HandleFunc("/start", enableCORS(startInstance))
	http.HandleFunc("/stop", enableCORS(stopInstance))

	log.Println("Master Server started on port :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}