package network

import (
	"adminadmin/internal/state"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

// AdminClient represents an admin client that connects to worker nodes
type AdminClient struct {
	conn            net.Conn
	connected       bool
	onUpdate        func(*state.DeviceInfo)
	onMetricsUpdate func(cpuUsage, ramUsage, gpuUsage float64)
}

// contains is a helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// NewAdminClient creates a new admin client
func NewAdminClient(onUpdate func(*state.DeviceInfo), onMetricsUpdate func(cpuUsage, ramUsage, gpuUsage float64)) *AdminClient {
	return &AdminClient{
		onUpdate:        onUpdate,
		onMetricsUpdate: onMetricsUpdate,
	}
}

// Connect connects to a worker node
func (a *AdminClient) Connect(address string, port int) error {
	addr := fmt.Sprintf("%s:%d", address, port)

	log.Printf("ADMIN: Attempting to connect to worker at %s...\n", addr)
	log.Println("ADMIN: If this takes a long time, check firewall settings on the Worker PC")
	log.Println("ADMIN: See FIREWALL.md for setup instructions")

	// Use 60 second timeout for slow networks or first-time connections
	// This is increased from 30s because firewall prompts can take time
	dialer := net.Dialer{
		Timeout:   60 * time.Second,
		KeepAlive: 15 * time.Second,
	}

	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		log.Printf("ADMIN ERROR: Failed to connect to %s: %v\n", addr, err)
		log.Println("ADMIN: ========== TROUBLESHOOTING ==========")
		log.Println("ADMIN: Possible causes:")
		log.Println("  1. Worker not running or not in Worker mode")
		log.Println("  2. Firewall blocking port 9876 on Worker PC")
		log.Println("  3. Wrong IP address")
		log.Println("  4. Different network/subnet")
		log.Println("  5. Antivirus blocking the connection")
		log.Println("")
		log.Println("ADMIN: Quick Fix (run on Worker PC as Admin):")
		log.Println("  New-NetFirewallRule -DisplayName \"admin:admin Worker\" -Direction Inbound -Protocol TCP -LocalPort 9876 -Action Allow")
		log.Println("ADMIN: ======================================")

		// Provide more helpful error message
		errMsg := err.Error()
		if contains(errMsg, "i/o timeout") || contains(errMsg, "timeout") {
			return fmt.Errorf("connection timeout - check if port 9876 is open on Worker PC firewall. Run on Worker PC as Admin: New-NetFirewallRule -DisplayName \"admin:admin Worker\" -Direction Inbound -Protocol TCP -LocalPort 9876 -Action Allow")
		} else if contains(errMsg, "connection refused") {
			return fmt.Errorf("connection refused - make sure the Worker application is running in Worker mode")
		} else if contains(errMsg, "no route to host") {
			return fmt.Errorf("no route to host - check if both PCs are on the same network")
		}
		return fmt.Errorf("failed to connect to worker: %w", err)
	}

	// Set TCP keep-alive to detect dead connections
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(15 * time.Second)
	}

	a.conn = conn
	a.connected = true
	log.Printf("ADMIN: TCP connection established to %s\n", addr)

	// Send admin info to worker
	a.sendAdminInfo()

	// Start listening for updates
	log.Println("ADMIN: Starting receive updates goroutine...")
	go a.receiveUpdates()

	log.Printf("ADMIN: Successfully connected to worker at %s\n", addr)
	return nil
}

// sendAdminInfo sends admin hostname to the worker
func (a *AdminClient) sendAdminInfo() {
	hostname, _ := os.Hostname()
	payload := AdminInfoPayload{
		Hostname: hostname,
	}
	payloadBytes, _ := json.Marshal(payload)
	msg := Message{
		Type:    MsgTypeAdminInfo,
		Payload: payloadBytes,
	}
	encoder := json.NewEncoder(a.conn)
	encoder.Encode(msg)
	log.Printf("ADMIN: Sent admin info (hostname: %s)\n", hostname)
}

// Disconnect disconnects from the worker node
func (a *AdminClient) Disconnect() error {
	if !a.connected || a.conn == nil {
		return nil
	}

	// Send disconnect message
	msg := Message{Type: MsgTypeDisconnect}
	encoder := json.NewEncoder(a.conn)
	encoder.Encode(msg)

	a.connected = false
	return a.conn.Close()
}

// IsConnected returns whether the client is connected
func (a *AdminClient) IsConnected() bool {
	return a.connected
}

// SendPing sends a ping to the worker
func (a *AdminClient) SendPing() error {
	if !a.connected {
		return fmt.Errorf("not connected")
	}

	msg := Message{Type: MsgTypePing}
	encoder := json.NewEncoder(a.conn)
	return encoder.Encode(msg)
}

func (a *AdminClient) receiveUpdates() {
	log.Println("ADMIN: Receive updates goroutine started")

	defer func() {
		log.Println("ADMIN: Receive updates goroutine ending")
		a.connected = false
		if a.conn != nil {
			a.conn.Close()
		}
	}()

	decoder := json.NewDecoder(a.conn)
	log.Println("ADMIN: Waiting for messages from worker...")

	for a.connected {
		var msg Message
		if err := decoder.Decode(&msg); err != nil {
			log.Printf("ADMIN: Connection error: %v\n", err)
			return
		}

		log.Printf("ADMIN: Received message type: %s\n", msg.Type)

		switch msg.Type {
		case MsgTypeSystemInfo:
			log.Println("ADMIN: Processing system info message...")
			var payload SystemInfoPayload
			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				log.Printf("ADMIN ERROR: Error parsing system info: %v\n", err)
				continue
			}

			log.Printf("ADMIN: System Info Received - Hostname: %s, OS: %s, IP: %s\n",
				payload.Hostname, payload.OS, payload.IPAddress)

			deviceInfo := &state.DeviceInfo{
				Hostname:      payload.Hostname,
				OS:            payload.OS,
				Architecture:  payload.Architecture,
				IPAddress:     payload.IPAddress,
				CPUUsage:      payload.CPUUsage,
				RAMUsage:      payload.RAMUsage,
				RAMTotal:      payload.RAMTotal,
				RAMUsed:       payload.RAMUsed,
				GPUName:       payload.GPUName,
				GPUUsage:      payload.GPUUsage,
				InternetSpeed: payload.InternetSpeed,
				Uptime:        payload.Uptime,
			}

			if a.onUpdate != nil {
				log.Println("ADMIN: Calling onUpdate callback...")
				a.onUpdate(deviceInfo)
			} else {
				log.Println("ADMIN WARNING: No onUpdate callback set")
			}

		case MsgTypeMetrics:
			var payload MetricsPayload
			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				log.Printf("ADMIN ERROR: Error parsing metrics: %v\n", err)
				continue
			}

			if a.onMetricsUpdate != nil {
				a.onMetricsUpdate(payload.CPUUsage, payload.RAMUsage, payload.GPUUsage)
			}

		case MsgTypePong:
			log.Println("ADMIN: Received pong from worker")

		default:
			log.Printf("ADMIN: Unknown message type: %s\n", msg.Type)
		}
	}

	log.Println("ADMIN: Exiting receive loop (not connected)")
}
