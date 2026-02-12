package system

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

type SystemInfo struct {
	Hostname      string
	OS            string
	Arch          string
	GoVersion     string
	CPUUsage      float64
	RAMUsage      float64
	RAMTotal      uint64
	RAMUsed       uint64
	GPUName       string
	GPUUsage      float64
	InternetSpeed string
	LocalIP       string
	Uptime        uint64
}

// GetLocalSystemInfo gets system information with real metrics
func GetLocalSystemInfo() SystemInfo {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	// Get CPU usage (average over 1 second)
	cpuPercent, err := cpu.Percent(time.Second, false)
	cpuUsage := 0.0
	if err == nil && len(cpuPercent) > 0 {
		cpuUsage = cpuPercent[0]
	}

	// Get RAM usage
	memInfo, err := mem.VirtualMemory()
	ramUsage := 0.0
	var ramTotal, ramUsed uint64
	if err == nil {
		ramUsage = memInfo.UsedPercent
		ramTotal = memInfo.Total
		ramUsed = memInfo.Used
	}

	// Get GPU info
	gpuName, gpuUsage := getGPUInfo()

	// Get local IP
	localIP := getLocalIP()

	// Get uptime
	uptime, _ := host.Uptime()

	return SystemInfo{
		Hostname:      hostname,
		OS:            getOSName(),
		Arch:          runtime.GOARCH,
		GoVersion:     runtime.Version(),
		CPUUsage:      cpuUsage,
		RAMUsage:      ramUsage,
		RAMTotal:      ramTotal,
		RAMUsed:       ramUsed,
		GPUName:       gpuName,
		GPUUsage:      gpuUsage,
		InternetSpeed: "N/A", // Will be measured on demand
		LocalIP:       localIP,
		Uptime:        uptime,
	}
}

// GetRealTimeMetrics gets only the dynamic metrics (CPU, RAM, GPU usage)
func GetRealTimeMetrics() (cpuUsage, ramUsage, gpuUsage float64) {
	// Get CPU usage (average over 500ms for faster updates)
	cpuPercent, err := cpu.Percent(500*time.Millisecond, false)
	if err == nil && len(cpuPercent) > 0 {
		cpuUsage = cpuPercent[0]
	}

	// Get RAM usage
	memInfo, err := mem.VirtualMemory()
	if err == nil {
		ramUsage = memInfo.UsedPercent
	}

	// Get GPU usage
	_, gpuUsage = getGPUInfo()

	return
}

// getOSName returns a human-readable OS name
func getOSName() string {
	switch runtime.GOOS {
	case "windows":
		return "Windows"
	case "darwin":
		return "macOS"
	case "linux":
		return "Linux"
	default:
		return runtime.GOOS
	}
}

// getLocalIP returns the local IP address
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "unknown"
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "unknown"
}

// getGPUInfo returns GPU name and usage (Windows only for now)
func getGPUInfo() (name string, usage float64) {
	name = "N/A"
	usage = 0.0

	if runtime.GOOS == "windows" {
		// Try to get GPU name using WMIC
		cmd := exec.Command("wmic", "path", "win32_VideoController", "get", "name")
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" && line != "Name" {
					name = line
					break
				}
			}
		}

		// Try to get GPU usage using nvidia-smi (for NVIDIA GPUs)
		cmd = exec.Command("nvidia-smi", "--query-gpu=utilization.gpu", "--format=csv,noheader,nounits")
		output, err = cmd.Output()
		if err == nil {
			var gpuUtil float64
			_, err = fmt.Sscanf(strings.TrimSpace(string(output)), "%f", &gpuUtil)
			if err == nil {
				usage = gpuUtil
			}
		}
	}

	return
}

// FormatBytes formats bytes to human readable string
func FormatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// FormatUptime formats uptime seconds to human readable string
func FormatUptime(seconds uint64) string {
	days := seconds / 86400
	hours := (seconds % 86400) / 3600
	minutes := (seconds % 3600) / 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}
