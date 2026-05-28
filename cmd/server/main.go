package main

import (
	"log"
	"net/http"

	"admin-aws/internal/config"
	"admin-aws/internal/database"
	httpdelivery "admin-aws/internal/delivery/http"
	"admin-aws/internal/services"
)

func main() {
	config.InitConfig()
	database.InitDB()

	go services.MonitorServers()

	http.HandleFunc("/", httpdelivery.EnableCORS(httpdelivery.GetInstances))
	http.HandleFunc("/logs", httpdelivery.EnableCORS(httpdelivery.GetLogs))
	http.HandleFunc("/alerts", httpdelivery.EnableCORS(httpdelivery.GetAlerts))
	http.HandleFunc("/system-metrics", httpdelivery.EnableCORS(httpdelivery.GetMetricsForFront))
	http.HandleFunc("/report-metrics", httpdelivery.EnableCORS(httpdelivery.ReceiveMetricsFromAgent))
	http.HandleFunc("/log-event", httpdelivery.EnableCORS(httpdelivery.HandleLogEvent))
	http.HandleFunc("/start", httpdelivery.EnableCORS(httpdelivery.StartInstance))
	http.HandleFunc("/stop", httpdelivery.EnableCORS(httpdelivery.StopInstance))
	http.HandleFunc("/kamatera-instances", httpdelivery.EnableCORS(httpdelivery.GetKamateraInstances))

	http.HandleFunc("/ssh", httpdelivery.HandleTerminal) 
	http.HandleFunc("/upload", httpdelivery.HandleUpload)
	http.HandleFunc("/servers", httpdelivery.HandleServers)

	log.Println("Unified Backend Server started on port :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
