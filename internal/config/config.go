package config

import (
	"context"
	"log"
	"os"

	"admin-aws/internal/models"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/joho/godotenv"
)

var (
	EC2Client   *ec2.Client
	ServersList map[string]models.ServerConfig
)

func GetEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return fallback
}

func InitConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Println("File .env not found, using default values.")
	}
	kamateraIP := GetEnv("KAMATERA_PUBLIC_IP", "0.0.0.0")
	doToken := GetEnv("DIGITALOCEAN_TOKEN", "")

	ServersList = map[string]models.ServerConfig{
		"my-windows-server": {
			ID:         "my-windows-server",
			Provider:   "Local",
			Platform:   "Windows",
			MacAddress: GetEnv("WINDOWS_MAC", "00:25:90:9A:4A:C0"),
			AgentURL:   GetEnv("WINDOWS_AGENT_URL", "http://127.0.0.1:8082"),
			WoLTargets: []string{
				GetEnv("WINDOWS_WOL_DOMAIN", "e7dd0f5572ff.sn.mynetname.net:9"),
				GetEnv("WINDOWS_WOL_LOCAL", "255.255.255.255:9"),
			},
		},
		"kamatera-server-01": {
			ID:       GetEnv("KAMATERA_SERVER_ID", "9f36db08-e0c9-4dfa-b2b3-d36ba11f5015"),
			Provider: "Kamatera",
			Platform: "Linux",
			AgentURL: "http://" + kamateraIP + ":8082",
		},
		"digitalocean-server-01": {
			ID:       GetEnv("DIGITALOCEAN_DROPLET_ID", "12345678"),
			Provider: "DigitalOcean",
			Platform: "Linux",
			AgentURL: GetEnv("DIGITALOCEAN_AGENT_URL", "http://127.0.0.1:8083"),
		},
	}
	
	if doToken == "" {
		log.Println("DIGITALOCEAN_TOKEN not set, DO API operations will fail")
	}

	region := GetEnv("AWS_DEFAULT_REGION", "eu-north-1")
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		log.Println("AWS configuration not found:", err)
	} else {
		EC2Client = ec2.NewFromConfig(cfg)
	}
}
