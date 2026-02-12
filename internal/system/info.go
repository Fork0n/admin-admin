package system

import (
	"os"
	"runtime"
)

type SystemInfo struct {
	Hostname  string
	OS        string
	Arch      string
	GoVersion string
	CPUUsage  float64
	RAMUsage  float64
}

func GetLocalSystemInfo() SystemInfo {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	return SystemInfo{
		Hostname:  hostname,
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		GoVersion: runtime.Version(),
		CPUUsage:  0.0,
		RAMUsage:  0.0,
	}
}
