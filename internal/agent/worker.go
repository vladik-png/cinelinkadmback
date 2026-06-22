package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

var (
	DeviceName     string
	ServerLocation string
	PublicIP       string
	MasterURL      string
	InstanceID     string
)

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return fallback
}

func sendLog(action, status, details string) {
	logURL := strings.Replace(MasterURL, "/report-metrics", "/log-event", 1)
	
	payload := map[string]interface{}{
		"instance_id": InstanceID,
		"component":   "AGENT",
		"action":      action,
		"status":      status,
		"details":     details,
	}

	jsonData, err := json.Marshal(payload)
	if err == nil {
		client := http.Client{Timeout: 3 * time.Second}
		resp, err := client.Post(logURL, "application/json", bytes.NewBuffer(jsonData))
		if err == nil {
			resp.Body.Close()
		}
	}
}

func initStaticInfo() {
	hostName, err := os.Hostname()
	if err == nil {
		DeviceName = hostName
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

func MathRound(val float64) float64 {
	return float64(int(val*100)) / 100
}

func collectAndSendMetrics() {
	cpuP, err := cpu.Percent(time.Second, false)
	if err != nil {
		sendLog("METRICS_ERROR", "ERROR", "Failed to get CPU percent: "+err.Error())
	}
	
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
		"instance_id": InstanceID,
		"device_name": DeviceName,
		"location":    ServerLocation,
		"public_ip":   PublicIP,
		"os":          runtime.GOOS,
		"time":        time.Now().Format("15:04:05"),
		"cpu_usage":   MathRound(cpuVal),
		"cpu_temp":    MathRound(getCPUTemperature()),
		"ram":         MathRound(float64(vMem.Total) / (1024 * 1024 * 1024)),
		"disk":        fmt.Sprintf("%.2f", d.UsedPercent),
		"ping":        latency,
		"packet_loss": packetLoss,
	}

	jsonData, err := json.Marshal(metrics)
	if err == nil {
		resp, err := http.Post(MasterURL, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			log.Println("Failed to send metrics:", err)
		} else {
			resp.Body.Close()
		}
	}
}

func Start() {
	log.Println("Initializing internal agent worker...")

	MasterURL = getEnv("MASTER_URL", "http://127.0.0.1:8081/report-metrics")
	InstanceID = getEnv("INSTANCE_ID", "")

	initStaticInfo()

	if InstanceID == "" || InstanceID == "local-pc" {
		InstanceID = DeviceName
	}
	
	sendLog("BOOT_COMPLETE", "SUCCESS", fmt.Sprintf("Agent worker started on %s (%s)", DeviceName, runtime.GOOS))

	log.Printf("Agent ID: %s. Sending metrics internally to: %s", InstanceID, MasterURL)

	go func() {
		for {
			collectAndSendMetrics()
			time.Sleep(2 * time.Second)
		}
	}()
}
