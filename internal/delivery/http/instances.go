package httpdelivery

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"admin-aws/internal/config"
	"admin-aws/internal/database"
	"admin-aws/internal/services"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

func GetKamateraInstances(w http.ResponseWriter, r *http.Request) {
	clientID := config.GetEnv("KAMATERA_CLIENT_ID", "")
	secretKey := config.GetEnv("KAMATERA_SECRET_KEY", "")

	if clientID == "" || secretKey == "" {
		http.Error(w, "Kamatera credentials not found", http.StatusInternalServerError)
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://console.kamatera.com/service/server", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	req.Header.Set("AuthClientId", clientID)
	req.Header.Set("AuthSecret", secretKey)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to reach Kamatera API", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Error reading response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func StartInstance(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	if srv, exists := config.ServersList[id]; exists {
		switch srv.Provider {
		case "Local":
			err := services.WakeOnLan(srv.MacAddress, srv.WoLTargets)
			if err != nil {
				database.LogEvent(id, "START", "FAILED", err.Error())
				http.Error(w, err.Error(), 500)
				return
			}
			database.LogEvent(id, "START", "SUCCESS", "Wake-on-LAN packet sent")
		case "DigitalOcean":
			token := config.GetEnv("DIGITALOCEAN_TOKEN", "")
			if token == "" {
				database.LogEvent(id, "START", "FAILED", "DigitalOcean token missing")
				http.Error(w, "DigitalOcean token missing", 500)
				return
			}
			req, _ := http.NewRequest("POST", "https://api.digitalocean.com/v2/droplets/"+srv.ID+"/actions", strings.NewReader(`{"type":"power_on"}`))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)
			
			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Do(req)
			if err != nil || resp.StatusCode >= 400 {
				errMsg := "Unknown error"
				if err != nil { errMsg = err.Error() } else { errMsg = resp.Status }
				database.LogEvent(id, "START", "FAILED", "DigitalOcean API Error: " + errMsg)
				http.Error(w, errMsg, 500)
				return
			}
			database.LogEvent(id, "START", "SUCCESS", "DigitalOcean droplet started")
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	if config.EC2Client != nil {
		_, err := config.EC2Client.StartInstances(context.TODO(), &ec2.StartInstancesInput{InstanceIds: []string{id}})
		if err != nil {
			database.LogEvent(id, "START", "FAILED", "AWS EC2 Error: "+err.Error())
			http.Error(w, err.Error(), 500)
			return
		}
		database.LogEvent(id, "START", "SUCCESS", "AWS Instance started")
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Error(w, "Server ID not found", 404)
}

func StopInstance(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	if srv, exists := config.ServersList[id]; exists {
		switch srv.Provider {
		case "Local":
			resp, err := http.Get(srv.AgentURL + "/shutdown")
			if err != nil {
				database.LogEvent(id, "STOP", "FAILED", "Agent unavailable: "+err.Error())
				http.Error(w, err.Error(), 500)
				return
			}
			resp.Body.Close()
			database.LogEvent(id, "STOP", "SUCCESS", "Shutdown command sent to Agent")
		case "DigitalOcean":
			token := config.GetEnv("DIGITALOCEAN_TOKEN", "")
			if token == "" {
				database.LogEvent(id, "STOP", "FAILED", "DigitalOcean token missing")
				http.Error(w, "DigitalOcean token missing", 500)
				return
			}
			req, _ := http.NewRequest("POST", "https://api.digitalocean.com/v2/droplets/"+srv.ID+"/actions", strings.NewReader(`{"type":"power_off"}`))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)
			
			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Do(req)
			if err != nil || resp.StatusCode >= 400 {
				errMsg := "Unknown error"
				if err != nil { errMsg = err.Error() } else { errMsg = resp.Status }
				database.LogEvent(id, "STOP", "FAILED", "DigitalOcean API Error: " + errMsg)
				http.Error(w, errMsg, 500)
				return
			}
			database.LogEvent(id, "STOP", "SUCCESS", "DigitalOcean droplet stopped")
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	if config.EC2Client != nil {
		_, err := config.EC2Client.StopInstances(context.TODO(), &ec2.StopInstancesInput{InstanceIds: []string{id}})
		if err != nil {
			database.LogEvent(id, "STOP", "FAILED", "AWS EC2 Error: "+err.Error())
			http.Error(w, err.Error(), 500)
			return
		}
		database.LogEvent(id, "STOP", "SUCCESS", "AWS Instance stopped")
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Error(w, "Server ID not found", 404)
}

func GetInstances(w http.ResponseWriter, r *http.Request) {
	instances := make([]map[string]interface{}, 0)

	for _, srv := range config.ServersList {
		instances = append(instances, map[string]interface{}{
			"InstanceId": srv.ID,
			"Provider":   srv.Provider,
			"Platform":   srv.Platform,
			"State":      "unknown",
		})
	}

	if config.EC2Client != nil {
		resp, err := config.EC2Client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{})
		if err == nil {
			for _, res := range resp.Reservations {
				for _, inst := range res.Instances {
					instances = append(instances, map[string]interface{}{
						"InstanceId": *inst.InstanceId,
						"Provider":   "AWS",
						"Platform":   "Linux/Windows",
						"State":      inst.State.Name,
					})
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(instances)
}
