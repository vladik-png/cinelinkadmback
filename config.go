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

	serversList = map[string]ServerConfig{
		"my-windows-server": {
			ID:         "my-windows-server",
			Provider:   "Local",
			Platform:   "Windows",
			MacAddress: "00:25:90:9A:4A:C0",
			AgentURL:   "http://127.0.0.1:8081",
			WoLTargets: []string{"e7dd0f5572ff.sn.mynetname.net:9", "255.255.255.255:9"},
		},
	}
)

func initConfig() {
	godotenv.Load()
	region := os.Getenv("AWS_DEFAULT_REGION")
	if region == "" {
		region = "eu-north-1"
	}
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		log.Println("AWS config not found (Local Mode Only)")
	} else {
		ec2Client = ec2.NewFromConfig(cfg)
		log.Println("AWS EC2 Client initialized")
	}
}