package main

import (
	"fmt"
	"net"
	"time"
)

func monitorServers() {
	for {
		time.Sleep(30 * time.Second)

		metricsMu.Lock()
		now := time.Now()

		for id, state := range latestMetrics {
			if now.Sub(state.LastSeen) > 2*time.Minute {
				createAlert(id, "OFFLINE", "Server has not reported metrics for over 2 minutes")
				delete(latestMetrics, id)
				continue
			}

			if cpuOpt, ok := state.Metrics["cpu_usage"]; ok {
				if cpu, isFloat := cpuOpt.(float64); isFloat && cpu > 90.0 {
					msg := fmt.Sprintf("Critical CPU usage: %.1f%%", cpu)
					createAlert(id, "OVERLOAD", msg)
				}
			}
		}
		metricsMu.Unlock()
	}
}

func wakeOnLan(macAddr string, targets []string) error {
	hwAddr, err := net.ParseMAC(macAddr)
	if err != nil {
		return err
	}
	packet := make([]byte, 102)
	copy(packet[0:6], []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})
	for i := 1; i <= 16; i++ {
		copy(packet[i*6:i*6+6], hwAddr)
	}

	for _, target := range targets {
		conn, err := net.Dial("udp", target)
		if err == nil {
			conn.Write(packet)
			conn.Close()
		}
	}
	return nil
}