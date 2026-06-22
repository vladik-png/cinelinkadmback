package main

import (
	"log"
	"net/http"

	"admin-aws/internal/agent"
	"admin-aws/internal/config"
	"admin-aws/internal/database"
httpdelivery "admin-aws/internal/delivery/http"
	"admin-aws/internal/services"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	config.InitConfig()
	database.InitDB()

	go services.MonitorServers()

	agent.Start()

	router := httpdelivery.SetupRouter()

	port := ":8081"
	log.Printf("Unified Backend Server started on port %s", port)
	log.Fatal(http.ListenAndServe(port, router))
}
