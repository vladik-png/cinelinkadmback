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
	"admin-aws/internal/models"
	"admin-aws/internal/services"
	"admin-aws/internal/state"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

func HandleLogEvent(w http.ResponseWriter, r *http.Request) {
	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	instanceID, _ := data["instance_id"].(string)
	action, _ := data["action"].(string)
	status, _ := data["status"].(string)
	details, _ := data["details"].(string)
	
	if instanceID != "" {
		database.LogEvent(instanceID, action, status, details)
	}
	w.WriteHeader(http.StatusOK)
}

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
		if srv.Provider == "Local" {
			err := services.WakeOnLan(srv.MacAddress, srv.WoLTargets)
			if err != nil {
				database.LogEvent(id, "START", "FAILED", err.Error())
				http.Error(w, err.Error(), 500)
				return
			}
			database.LogEvent(id, "START", "SUCCESS", "Wake-on-LAN packet sent")
		} else if srv.Provider == "DigitalOcean" {
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
			_, err := http.Get(srv.AgentURL + "/shutdown")
			if err != nil {
				database.LogEvent(id, "STOP", "FAILED", "Agent unavailable: "+err.Error())
				http.Error(w, err.Error(), 500)
				return
			}
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

func ReceiveMetricsFromAgent(w http.ResponseWriter, r *http.Request) {
	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id, ok := data["instance_id"].(string)
	if !ok {
		return
	}

	state.MetricsMu.Lock()
	state.LatestMetrics[id] = models.ServerState{
		LastSeen: time.Now(),
		Metrics:  data,
	}
	state.MetricsMu.Unlock()

	if database.DB != nil {
		database.DB.Model(&models.ServerAlert{}).Where("server_id = ? AND type = ? AND resolved = ?", id, "OFFLINE", false).Update("resolved", true)
	}
	w.WriteHeader(http.StatusOK)
}

func GetLogs(w http.ResponseWriter, r *http.Request) {
	var logs []models.ServerLog
	if database.DB != nil {
		database.DB.Order("created_at desc").Limit(50).Find(&logs)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

func GetAlerts(w http.ResponseWriter, r *http.Request) {
	var alerts []models.ServerAlert
	if database.DB != nil {
		database.DB.Where("resolved = ?", false).Order("created_at desc").Find(&alerts)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alerts)
}

func GetMetricsForFront(w http.ResponseWriter, r *http.Request) {
	state.MetricsMu.Lock()
	defer state.MetricsMu.Unlock()

	frontData := make(map[string]interface{})
	for id, s := range state.LatestMetrics {
		frontData[id] = s.Metrics
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(frontData)
}

func EnableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}
