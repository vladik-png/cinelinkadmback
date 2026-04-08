package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

func startInstance(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	if srv, exists := serversList[id]; exists {
		if srv.Provider == "Local" {
			err := wakeOnLan(srv.MacAddress, srv.WoLTargets)
			if err != nil {
				logEvent(id, "START", "FAILED", err.Error())
				http.Error(w, err.Error(), 500)
				return
			}
			logEvent(id, "START", "SUCCESS", "Wake-on-LAN packet sent")
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	if ec2Client != nil {
		_, err := ec2Client.StartInstances(context.TODO(), &ec2.StartInstancesInput{InstanceIds: []string{id}})
		if err != nil {
			logEvent(id, "START", "FAILED", "AWS EC2 Error: "+err.Error())
			http.Error(w, err.Error(), 500)
			return
		}
		logEvent(id, "START", "SUCCESS", "AWS Instance started")
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Error(w, "Server ID not found", 404)
}

func stopInstance(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	if srv, exists := serversList[id]; exists {
		if srv.Provider == "Local" {
			_, err := http.Get(srv.AgentURL + "/shutdown")
			if err != nil {
				logEvent(id, "STOP", "FAILED", "Agent unavailable: "+err.Error())
				http.Error(w, err.Error(), 500)
				return
			}
			logEvent(id, "STOP", "SUCCESS", "Shutdown command sent to Agent")
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	if ec2Client != nil {
		_, err := ec2Client.StopInstances(context.TODO(), &ec2.StopInstancesInput{InstanceIds: []string{id}})
		if err != nil {
			logEvent(id, "STOP", "FAILED", "AWS EC2 Error: "+err.Error())
			http.Error(w, err.Error(), 500)
			return
		}
		logEvent(id, "STOP", "SUCCESS", "AWS Instance stopped")
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Error(w, "Server ID not found", 404)
}

func getInstances(w http.ResponseWriter, r *http.Request) {
	instances := make([]map[string]interface{}, 0)

	for _, srv := range serversList {
		instances = append(instances, map[string]interface{}{
			"InstanceId": srv.ID,
			"Provider":   srv.Provider,
			"Platform":   srv.Platform,
			"State":      "unknown",
		})
	}

	if ec2Client != nil {
		resp, err := ec2Client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{})
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

func receiveMetricsFromAgent(w http.ResponseWriter, r *http.Request) {
	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id, ok := data["instance_id"].(string)
	if !ok {
		return
	}

	metricsMu.Lock()
	latestMetrics[id] = ServerState{
		LastSeen: time.Now(),
		Metrics:  data,
	}
	metricsMu.Unlock()

	db.Model(&ServerAlert{}).Where("server_id = ? AND type = ? AND resolved = ?", id, "OFFLINE", false).Update("resolved", true)
	w.WriteHeader(http.StatusOK)
}

func getLogs(w http.ResponseWriter, r *http.Request) {
	var logs []ServerLog
	db.Order("created_at desc").Limit(50).Find(&logs)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

func getAlerts(w http.ResponseWriter, r *http.Request) {
	var alerts []ServerAlert
	db.Where("resolved = ?", false).Order("created_at desc").Find(&alerts)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alerts)
}

func getMetricsForFront(w http.ResponseWriter, r *http.Request) {
	metricsMu.Lock()
	defer metricsMu.Unlock()

	frontData := make(map[string]interface{})
	for id, state := range latestMetrics {
		frontData[id] = state.Metrics
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(frontData)
}