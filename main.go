package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/joho/godotenv"
)

var ec2Client *ec2.Client

func main() {
	godotenv.Load()
	region := os.Getenv("AWS_DEFAULT_REGION")
	if region == "" {
		region = "eu-north-1"
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
	)
	if err != nil {
		log.Fatal(err)
	}

	ec2Client = ec2.NewFromConfig(cfg)

	http.HandleFunc("/", enableCORS(index))
	http.HandleFunc("/info", enableCORS(getInfo))
	http.HandleFunc("/start", enableCORS(startInstance))
	http.HandleFunc("/stop", enableCORS(stopInstance))

	log.Println("Server running on :8080")
	http.ListenAndServe(":8080", nil)
}

func getInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// Повертаємо саме твій регіон
	region := os.Getenv("AWS_DEFAULT_REGION")
	json.NewEncoder(w).Encode(map[string]string{"region": region})
}

func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	resp, err := ec2Client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{})
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
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