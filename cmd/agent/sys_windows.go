//go:build windows

package main

import (
	"github.com/yusufpapurcu/wmi"
)

type Win32_Temperature struct {
	CurrentTemperature uint32
}

func getCPUTemperature() float64 {
	var dst []Win32_Temperature
	q := "SELECT CurrentTemperature FROM MSAcpi_ThermalZoneTemperature"
	err := wmi.Query(q, &dst)
	if err != nil || len(dst) == 0 {
		return 0
	}
	return (float64(dst[0].CurrentTemperature) - 2732.0) / 10.0
}
