package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

const MasterServerURL = "http://E7dd0f5572ff.sn.mynetname.net:8082/report-metrics"

var (
	DeviceName     string
	ServerLocation string
	PublicIP       string
)

func initStaticInfo() {
	host, err := os.Hostname()
	if err == nil {
		DeviceName = host
	} else {
		DeviceName = "Unknown Device"
	}

	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("http://ip-api.com/json/")
	if err == nil {
		defer resp.Body.Close()
		var data map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&data)

		country, _ := data["country"].(string)
		city, _ := data["city"].(string)
		ip, _ := data["query"].(string)

		if country != "" && city != "" {
			ServerLocation = fmt.Sprintf("%s, %s", country, city)
			PublicIP = ip
		} else {
			ServerLocation = "Unknown Location"
			PublicIP = "Local/Unknown"
		}
	} else {
		ServerLocation = "Offline"
		PublicIP = "Offline"
	}
}

func getInstanceID() string {
	return DeviceName
}

func MathRound(val float64) float64 {
	return float64(int(val*100)) / 100
}

func shutdownHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received shutdown command!")
	w.WriteHeader(http.StatusOK)

	if runtime.GOOS == "windows" {
		exec.Command("cmd", "/c", "shutdown", "/s", "/t", "0").Run()
	} else {
		exec.Command("sudo", "shutdown", "-h", "now").Run()
	}
}

func collectAndSendMetrics(nodeID string) {
	cpuP, _ := cpu.Percent(time.Second, false)
	vMem, _ := mem.VirtualMemory()

	diskPath := "/"
	if runtime.GOOS == "windows" {
		diskPath = "C:"
	}
	d, _ := disk.Usage(diskPath)

	start := time.Now()
	latency := int64(0)
	packetLoss := "0"

	client := http.Client{Timeout: 2 * time.Second}
	respPing, err := client.Get("http://www.google.com")
	if err == nil {
		latency = time.Since(start).Milliseconds()
		respPing.Body.Close()
	} else {
		packetLoss = "100"
	}

	cpuVal := 0.0
	if len(cpuP) > 0 {
		cpuVal = cpuP[0]
	}

	metrics := map[string]interface{}{
		"instance_id": nodeID,
		"device_name": DeviceName,
		"location":    ServerLocation,
		"public_ip":   PublicIP,
		"os":          runtime.GOOS,
		"time":        time.Now().Format("15:04:05"),
		"cpu":         MathRound(cpuVal),
		"ram":         MathRound(vMem.UsedPercent),
		"disk":        fmt.Sprintf("%.2f", d.UsedPercent),
		"ping":        latency,
		"packet_loss": packetLoss,
	}

	jsonData, err := json.Marshal(metrics)
	if err == nil {
		http.Post(MasterServerURL, "application/json", bytes.NewBuffer(jsonData))
	}
}

func main() {
	log.Println("Initializing agent...")
	initStaticInfo()

	nodeID := getInstanceID()
	log.Printf("Agent started. Device: %s. Location: %s. OS: %s", DeviceName, ServerLocation, runtime.GOOS)

	http.HandleFunc("/shutdown", shutdownHandler)

	go func() {
		log.Println("Agent listening for commands on local port :8081")
		log.Fatal(http.ListenAndServe(":8081", nil))
	}()

	for {
		collectAndSendMetrics(nodeID)
		time.Sleep(2 * time.Second)
	}
}