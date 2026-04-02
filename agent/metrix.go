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


const MasterServerURL = "http://127.0.0.1:8082/report-metrics"

func getInstanceID() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "my-local-pc"
	}

	if runtime.GOOS == "windows" {
		return "my-windows-server"
	}
	return hostname
}

func MathRound(val float64) float64 {
	return float64(int(val*100)) / 100
}

func shutdownHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Отримано команду на вимкнення!")
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
        "os":          runtime.GOOS,
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
	nodeID := getInstanceID()
	log.Printf("Агент запущено. ID: %s. ОС: %s", nodeID, runtime.GOOS)

	http.HandleFunc("/shutdown", shutdownHandler)
	
	go func() {
		log.Println("Агент слухає команди на порту :8081")
		log.Fatal(http.ListenAndServe(":8081", nil))
	}()

	for {
		collectAndSendMetrics(nodeID)
		time.Sleep(2 * time.Second)
	}
}