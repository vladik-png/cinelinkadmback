package config

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"

	"admin-aws/internal/models"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/joho/godotenv"
)

var (
	EC2Client    *ec2.Client
	ServersList  map[string]models.ServerConfig
	ServersMutex sync.RWMutex
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

	hostName, err := os.Hostname()
	if err != nil {
		hostName = "local-pc"
	}

	ServersList = map[string]models.ServerConfig{
		hostName: {
			ID:         hostName,
			Provider:   "Local",
			Platform:   runtime.GOOS,
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
		log.Println("DIGITALOCEAN_TOKEN not set, DO API auto-discovery will fail")
	} else {
		log.Println("DIGITALOCEAN_TOKEN found, starting auto-discovery...")
		syncDigitalOceanServers()
		go func() {
			for {
				time.Sleep(5 * time.Minute)
				syncDigitalOceanServers()
			}
		}()
	}

	region := GetEnv("AWS_DEFAULT_REGION", "eu-north-1")
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		log.Println("AWS configuration not found:", err)
	} else {
		EC2Client = ec2.NewFromConfig(cfg)
	}
}

type DODropletsResponse struct {
	Droplets []struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		Networks struct {
			V4 []struct {
				IPAddress string `json:"ip_address"`
				Type      string `json:"type"`
			} `json:"v4"`
		} `json:"networks"`
	} `json:"droplets"`
}

func syncDigitalOceanServers() {
	doToken := GetEnv("DIGITALOCEAN_TOKEN", "")
	if doToken == "" {
		return
	}

	req, err := http.NewRequest("GET", "https://api.digitalocean.com/v2/droplets", nil)
	if err != nil {
		log.Println("Error creating DO request:", err)
		return
	}
	req.Header.Set("Authorization", "Bearer "+doToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error fetching DO droplets:", err)
		return
	}
	defer resp.Body.Close()

	var result DODropletsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Println("Error decoding DO droplets:", err)
		return
	}

	ServersMutex.Lock()
	for _, droplet := range result.Droplets {
		ip := ""
		for _, v4 := range droplet.Networks.V4 {
			if v4.Type == "public" {
				ip = v4.IPAddress
				break
			}
		}

		key := fmt.Sprintf("digitalocean-%d", droplet.ID)
		ServersList[key] = models.ServerConfig{
			ID:       fmt.Sprintf("%d", droplet.ID),
			Provider: "DigitalOcean",
			Platform: "Linux",
			AgentURL: fmt.Sprintf("http://%s:8083", ip),
		}
	}
	ServersMutex.Unlock()
	log.Printf("DigitalOcean sync complete: found %d droplets", len(result.Droplets))
}
