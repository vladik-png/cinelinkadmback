package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
	"os"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

var (
	nodesHistory = make(map[string]map[string]interface{})
	historyLock  sync.RWMutex
)

func getInstanceID() string {
	hostname, err := os.Hostname()
    if err != nil {
        return "my-local-pc"
    }
    return hostname
}

func MathRound(val float64) float64 {
	return float64(int(val*100)) / 100
}

func systemMetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	historyLock.RLock()
	defer historyLock.RUnlock()
	json.NewEncoder(w).Encode(nodesHistory)
}

func updateMetrics(nodeID string) {
	cpuP, _ := cpu.Percent(time.Second, false)
	vMem, _ := mem.VirtualMemory()
	d, _ := disk.Usage("D:")

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

	metrics := map[string]interface{}{
		"time":        time.Now().Format("15:04:05"),
		"cpu":         MathRound(cpuP[0]),
		"ram":         MathRound(vMem.UsedPercent),
		"disk":        fmt.Sprintf("%.2f", d.UsedPercent),
		"ping":        latency,
		"packet_loss": packetLoss,
	}

	historyLock.Lock()
	nodesHistory[nodeID] = metrics
	historyLock.Unlock()
}

func main() {
	nodeID := getInstanceID()
	log.Printf("Agent started for ID: %s", nodeID)

	http.HandleFunc("/system-metrics", systemMetricsHandler)
	go func() {
		log.Println("Server listening on :8081")
		log.Fatal(http.ListenAndServe(":8081", nil))
	}()

	for {
		updateMetrics(nodeID)
		time.Sleep(2 * time.Second)
	}
}

