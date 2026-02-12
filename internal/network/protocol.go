package network

import (
	"encoding/json"
)

// MessageType defines the type of network message
type MessageType string

const (
	MsgTypeSystemInfo MessageType = "system_info"
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

// SystemInfoPayload contains system information
type SystemInfoPayload struct {
	Hostname     string  `json:"hostname"`
	OS           string  `json:"os"`
	Architecture string  `json:"architecture"`
	IPAddress    string  `json:"ip_address"`
	CPUUsage     float64 `json:"cpu_usage"`
	RAMUsage     float64 `json:"ram_usage"`
}

// CommandPayload contains a command to execute
type CommandPayload struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}
