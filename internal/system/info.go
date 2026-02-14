package system

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

// getHiddenWindowAttr returns syscall attributes to hide command windows on Windows
func getHiddenWindowAttr() *syscall.SysProcAttr {
	if runtime.GOOS == "windows" {
		return &syscall.SysProcAttr{HideWindow: true}
	}
	return nil
}

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
	// Get CPU usage (average over 200ms for faster updates)
	cpuPercent, err := cpu.Percent(200*time.Millisecond, false)
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

// getLocalIP returns the local IP address (prefers 192.168.x.x or 10.x.x.x addresses)
func getLocalIP() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "unknown"
	}

	var fallbackIP string
	var candidateIPs []string

	for _, iface := range interfaces {
		// Skip down, loopback, and virtual interfaces
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		// Skip common virtual adapter names
		nameLower := strings.ToLower(iface.Name)
		if strings.Contains(nameLower, "virtual") ||
			strings.Contains(nameLower, "vmware") ||
			strings.Contains(nameLower, "vbox") ||
			strings.Contains(nameLower, "docker") ||
			strings.Contains(nameLower, "vethernet") ||
			strings.Contains(nameLower, "hyper-v") {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ipnet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}

			ip := ipnet.IP.To4()
			if ip == nil {
				continue
			}

			// Skip loopback and link-local addresses (169.254.x.x)
			if ip.IsLoopback() || ip.IsLinkLocalUnicast() {
				continue
			}

			// Prefer private network addresses (192.168.x.x, 10.x.x.x, 172.16-31.x.x)
			if ip[0] == 192 && ip[1] == 168 {
				// Highest priority - return immediately
				return ip.String()
			}
			if ip[0] == 10 {
				candidateIPs = append(candidateIPs, ip.String())
			}
			if ip[0] == 172 && ip[1] >= 16 && ip[1] <= 31 {
				candidateIPs = append(candidateIPs, ip.String())
			}

			// Store as fallback if no private address found
			if fallbackIP == "" {
				fallbackIP = ip.String()
			}
		}
	}

	// Return first candidate IP if we found any 10.x.x.x or 172.x.x.x addresses
	if len(candidateIPs) > 0 {
		return candidateIPs[0]
	}

	if fallbackIP != "" {
		return fallbackIP
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
		cmd.SysProcAttr = getHiddenWindowAttr()
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
		cmd.SysProcAttr = getHiddenWindowAttr()
		output, err = cmd.Output()
		if err == nil {
			var gpuUtil float64
			_, err = fmt.Sscanf(strings.TrimSpace(string(output)), "%f", &gpuUtil)
			if err == nil {
				usage = gpuUtil
				return
			}
		}

		// Try PowerShell for GPU usage (works with most GPUs)
		psCmd := `(Get-Counter '\GPU Engine(*engtype_3D)\Utilization Percentage' -ErrorAction SilentlyContinue).CounterSamples | Measure-Object -Property CookedValue -Sum | Select-Object -ExpandProperty Sum`
		cmd = exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psCmd)
		cmd.SysProcAttr = getHiddenWindowAttr()
		output, err = cmd.Output()
		if err == nil {
			var gpuUtil float64
			_, err = fmt.Sscanf(strings.TrimSpace(string(output)), "%f", &gpuUtil)
			if err == nil {
				usage = gpuUtil
				return
			}
		}

		// Alternative: Try GPU 3D usage counter
		psCmd2 := `(Get-Counter '\GPU Engine(*)\Utilization Percentage' -ErrorAction SilentlyContinue).CounterSamples | Where-Object { $_.InstanceName -like '*engtype_3D*' } | Measure-Object -Property CookedValue -Average | Select-Object -ExpandProperty Average`
		cmd = exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psCmd2)
		cmd.SysProcAttr = getHiddenWindowAttr()
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
