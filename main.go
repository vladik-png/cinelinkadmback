package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/joho/godotenv"
)

var (
	ec2Client     *ec2.Client
	latestMetrics = make(map[string]map[string]interface{})
	metricsMu     sync.Mutex
)

func main() {
	godotenv.Load()
	region := os.Getenv("AWS_DEFAULT_REGION")
	if region == "" { region = "eu-north-1" }

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil { log.Fatal(err) }
	ec2Client = ec2.NewFromConfig(cfg)

	http.HandleFunc("/", enableCORS(index))
	http.HandleFunc("/system-metrics", enableCORS(getMetricsForFront))

	http.HandleFunc("/report-metrics", enableCORS(receiveMetricsFromAgent))

	http.HandleFunc("/start", enableCORS(startInstance))
	http.HandleFunc("/stop", enableCORS(stopInstance))

	log.Println("Master Server on for :8082")
	http.ListenAndServe(":8082", nil)
}

func receiveMetricsFromAgent(w http.ResponseWriter, r *http.Request) {
	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		return
	}
	id, _ := data["instance_id"].(string)
	
	metricsMu.Lock()
	latestMetrics[id] = data
	metricsMu.Unlock()
	w.WriteHeader(http.StatusOK)
}

func getMetricsForFront(w http.ResponseWriter, r *http.Request) {
	metricsMu.Lock()
	defer metricsMu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(latestMetrics)
}

func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" { w.WriteHeader(http.StatusOK); return }
		next(w, r)
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	resp, err := ec2Client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{})
	if err != nil { http.Error(w, err.Error(), 500); return }
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp.Reservations)
}

func startInstance(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	ec2Client.StartInstances(context.TODO(), &ec2.StartInstancesInput{InstanceIds: []string{id}})
	w.WriteHeader(http.StatusOK)
}

func stopInstance(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	ec2Client.StopInstances(context.TODO(), &ec2.StopInstancesInput{InstanceIds: []string{id}})
	w.WriteHeader(http.StatusOK)
}