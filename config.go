package main

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/joho/godotenv"
)

type ServerConfig struct {
	ID         string
	Provider   string
	Platform   string
	MacAddress string
	AgentURL   string
	WoLTargets []string
}

type ServerState struct {
	LastSeen time.Time
	Metrics  map[string]interface{}
}

var (
	ec2Client     *ec2.Client
	latestMetrics = make(map[string]ServerState)
	metricsMu     sync.Mutex
	serversList   map[string]ServerConfig
)

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return fallback
}

func initConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Println("File .env not found, using default values.")
	}

	serversList = map[string]ServerConfig{
		"my-windows-server": {
			ID:         "my-windows-server",
			Provider:   "Local",
			Platform:   "Windows",
			MacAddress: getEnv("WINDOWS_MAC", "00:25:90:9A:4A:C0"),
			AgentURL:   getEnv("WINDOWS_AGENT_URL", "http://127.0.0.1:8081"),
			WoLTargets: []string{
				getEnv("WINDOWS_WOL_DOMAIN", "e7dd0f5572ff.sn.mynetname.net:9"),
				getEnv("WINDOWS_WOL_LOCAL", "255.255.255.255:9"),
			},
		},
	}
	log.Println("Configuration loaded successfully")

	region := getEnv("AWS_DEFAULT_REGION", "eu-north-1")
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		log.Println("AWS configuration not found (only local mode will work):", err)
	} else {
		ec2Client = ec2.NewFromConfig(cfg)
		log.Println("AWS EC2 client initialized successfully")
	}
}