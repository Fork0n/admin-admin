package network

import (
	"encoding/json"
)

// MessageType defines the type of network message
type MessageType string

const (
	MsgTypeSystemInfo MessageType = "system_info"
	MsgTypeMetrics    MessageType = "metrics"
	MsgTypeAdminInfo  MessageType = "admin_info"
	MsgTypeCommand    MessageType = "command"
	MsgTypePing       MessageType = "ping"
	MsgTypePong       MessageType = "pong"
	MsgTypeDisconnect MessageType = "disconnect"
)

// Message represents a network message
type Message struct {
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// SystemInfoPayload contains full system information
type SystemInfoPayload struct {
	Hostname      string  `json:"hostname"`
	OS            string  `json:"os"`
	Architecture  string  `json:"architecture"`
	IPAddress     string  `json:"ip_address"`
	CPUUsage      float64 `json:"cpu_usage"`
	RAMUsage      float64 `json:"ram_usage"`
	RAMTotal      uint64  `json:"ram_total"`
	RAMUsed       uint64  `json:"ram_used"`
	GPUName       string  `json:"gpu_name"`
	GPUUsage      float64 `json:"gpu_usage"`
	InternetSpeed string  `json:"internet_speed"`
	Uptime        uint64  `json:"uptime"`
}

// MetricsPayload contains real-time metrics update
type MetricsPayload struct {
	CPUUsage float64 `json:"cpu_usage"`
	RAMUsage float64 `json:"ram_usage"`
	GPUUsage float64 `json:"gpu_usage"`
}

// AdminInfoPayload contains admin device info sent to worker
type AdminInfoPayload struct {
	Hostname string `json:"hostname"`
}

// CommandPayload contains a command to execute
type CommandPayload struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}
