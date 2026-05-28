//go:build linux || darwin

package main

import (
	"strings"

	"github.com/shirou/gopsutil/v3/host"
)

func getCPUTemperature() float64 {
	temps, err := host.SensorsTemperatures()
	if err == nil {
		for _, t := range temps {
			if strings.Contains(strings.ToLower(t.SensorKey), "cpu") || strings.Contains(strings.ToLower(t.SensorKey), "core") {
				return t.Temperature
			}
		}
	}
	return 0
}
