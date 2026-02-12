package network

import (
	"adminadmin/internal/state"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"
)

// AdminClient represents an admin client that connects to worker nodes
type AdminClient struct {
	conn      net.Conn
	connected bool
	onUpdate  func(*state.DeviceInfo)
}

// NewAdminClient creates a new admin client
func NewAdminClient(onUpdate func(*state.DeviceInfo)) *AdminClient {
	return &AdminClient{
		onUpdate: onUpdate,
	}
}

// Connect connects to a worker node
func (a *AdminClient) Connect(address string, port int) error {
	addr := fmt.Sprintf("%s:%d", address, port)

	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to worker: %w", err)
	}

	a.conn = conn
	a.connected = true

	// Start listening for updates
	go a.receiveUpdates()

	log.Printf("Connected to worker at %s\n", addr)
	return nil
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
				Hostname:     payload.Hostname,
				OS:           payload.OS,
				Architecture: payload.Architecture,
				IPAddress:    payload.IPAddress,
			}

			if a.onUpdate != nil {
				log.Println("ADMIN: Calling onUpdate callback...")
				a.onUpdate(deviceInfo)
			} else {
				log.Println("ADMIN WARNING: No onUpdate callback set")
			}

		case MsgTypePong:
			log.Println("ADMIN: Received pong from worker")

		default:
			log.Printf("ADMIN: Unknown message type: %s\n", msg.Type)
		}
	}

	log.Println("ADMIN: Exiting receive loop (not connected)")
}
