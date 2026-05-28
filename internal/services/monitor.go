package services

import (
	"fmt"
	"net"
	"time"

	"admin-aws/internal/state"
	"admin-aws/internal/database"
)

func MonitorServers() {
	for {
		time.Sleep(30 * time.Second)

		state.MetricsMu.Lock()
		now := time.Now()

		for id, s := range state.LatestMetrics {
			if now.Sub(s.LastSeen) > 2*time.Minute {
				database.CreateAlert(id, "OFFLINE", "Server has not reported metrics for over 2 minutes")
				delete(state.LatestMetrics, id)
				continue
			}

			if cpuOpt, ok := s.Metrics["cpu_usage"]; ok {
				if cpu, isFloat := cpuOpt.(float64); isFloat && cpu > 90.0 {
					msg := fmt.Sprintf("Critical CPU usage: %.1f%%", cpu)
					database.CreateAlert(id, "OVERLOAD", msg)
				}
			}
		}
		state.MetricsMu.Unlock()
	}
}

func WakeOnLan(macAddr string, targets []string) error {
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
