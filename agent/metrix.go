package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

func getInstanceID() string {
	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://169.254.169.254/latest/meta-data/instance-id")
	if err != nil {
		return "dev-local-node"
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	return string(body)
}

func main() {
	masterURL := "http://13.62.214.254:8080/report-metrics"
	myID := getInstanceID()

	log.Printf("Agent started for ID: %s", myID)

	for {
		cpuP, _ := cpu.Percent(time.Second, false)
		vMem, _ := mem.VirtualMemory()

		metrics := map[string]interface{}{
			"instance_id": myID,
			"cpu":         MathRound(cpuP[0]),
			"ram":         MathRound(vMem.UsedPercent),
			"time":        time.Now().Format("15:04:05"),
		}

		jsonData, _ := json.Marshal(metrics)
		http.Post(masterURL, "application/json", bytes.NewBuffer(jsonData))

		time.Sleep(2 * time.Second)
	}
}

func MathRound(val float64) float64 {
	return float64(int(val*100)) / 100
}