package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
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


const (
	LocalInstanceID = "my-windows-server"
	LocalMacAddress = "00:11:22:33:44:55"
	LocalAgentURL   = "http://127.0.0.1:8081"
)

func main() {
	godotenv.Load()
	region := os.Getenv("AWS_DEFAULT_REGION")
	if region == "" {
		region = "eu-north-1"
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		log.Println("No AWS config found (will work only locally):", err)
	} else {
		ec2Client = ec2.NewFromConfig(cfg)
	}

	http.HandleFunc("/", enableCORS(index))
	http.HandleFunc("/system-metrics", enableCORS(getMetricsForFront))
	http.HandleFunc("/report-metrics", enableCORS(receiveMetricsFromAgent))
	http.HandleFunc("/start", enableCORS(startInstance))
	http.HandleFunc("/stop", enableCORS(stopInstance))

	log.Println("Master Server started on port :8082")
	http.ListenAndServe(":8082", nil)
}

// Wake-on-LAN функція
func wakeOnLan(macAddr string) error {
	hwAddr, err := net.ParseMAC(macAddr)
	if err != nil {
		return err
	}
	packet := make([]byte, 102)
	copy(packet[0:6], []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})
	for i := 1; i <= 16; i++ {
		copy(packet[i*6:i*6+6], hwAddr)
	}
	conn, err := net.Dial("udp", "255.255.255.255:9")
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = conn.Write(packet)
	return err
}

func receiveMetricsFromAgent(w http.ResponseWriter, r *http.Request) {
	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	instances := make([]map[string]interface{}, 0)

	instances = append(instances, map[string]interface{}{
		"InstanceId": LocalInstanceID,
		"Provider":   "Local",

	})

	if ec2Client != nil {
		resp, err := ec2Client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{})
		if err == nil {
			for _, res := range resp.Reservations {
				for _, inst := range res.Instances {
					instances = append(instances, map[string]interface{}{
						"InstanceId": *inst.InstanceId,
						"Provider":   "AWS",
						"State":      inst.State.Name,
					})
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(instances)
}

func startInstance(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	if id == LocalInstanceID {
		err := wakeOnLan(LocalMacAddress)
		if err != nil {
			http.Error(w, "Error sending Wake-on-LAN packet: "+err.Error(), 500)
			return
		}
		log.Println("Wake-on-LAN packet sent to", LocalMacAddress)
	} else if ec2Client != nil {
		_, err := ec2Client.StartInstances(context.TODO(), &ec2.StartInstancesInput{InstanceIds: []string{id}})
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
}

func stopInstance(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	if id == LocalInstanceID {
		_, err := http.Get(LocalAgentURL + "/shutdown")
		if err != nil {
			http.Error(w, "Failed to connect to agent: "+err.Error(), 500)
			return
		}
		log.Println("Shutdown command sent to local server")
	} else if ec2Client != nil {
		_, err := ec2Client.StopInstances(context.TODO(), &ec2.StopInstancesInput{InstanceIds: []string{id}})
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
}