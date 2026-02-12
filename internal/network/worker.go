package network

import (
	"adminadmin/internal/system"
	"encoding/json"
	"fmt"
	"log"
	"net"
)

const (
	DefaultWorkerPort = 9876
)

// WorkerServer represents a worker node server
type WorkerServer struct {
	listener net.Listener
	port     int
	quit     chan bool
	sysInfo  system.SystemInfo
}

// NewWorkerServer creates a new worker server
func NewWorkerServer(port int) *WorkerServer {
	if port == 0 {
		port = DefaultWorkerPort
	}
	return &WorkerServer{
		port: port,
		quit: make(chan bool),
	}
}

// Start starts the worker server
func (w *WorkerServer) Start() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", w.port))
	if err != nil {
		return fmt.Errorf("failed to start worker server: %w", err)
	}
	w.listener = listener

	log.Printf("Worker server listening on port %d\n", w.port)

	go w.acceptConnections()
	return nil
}

// Stop stops the worker server
func (w *WorkerServer) Stop() error {
	close(w.quit)
	if w.listener != nil {
		return w.listener.Close()
	}
	return nil
}

// GetPort returns the port the server is listening on
func (w *WorkerServer) GetPort() int {
	return w.port
}

func (w *WorkerServer) acceptConnections() {
	for {
		select {
		case <-w.quit:
			return
		default:
			conn, err := w.listener.Accept()
			if err != nil {
				select {
				case <-w.quit:
					return
				default:
					log.Printf("Error accepting connection: %v\n", err)
					continue
				}
			}
			go w.handleConnection(conn)
		}
	}
}

func (w *WorkerServer) handleConnection(conn net.Conn) {
	defer conn.Close()
	log.Printf("Admin connected from: %s\n", conn.RemoteAddr())

	// Send system info immediately upon connection
	w.sendSystemInfo(conn)

	// Keep connection alive and handle incoming messages
	decoder := json.NewDecoder(conn)
	for {
		var msg Message
		if err := decoder.Decode(&msg); err != nil {
			log.Printf("Connection closed: %v\n", err)
			return
		}

		switch msg.Type {
		case MsgTypePing:
			w.sendPong(conn)
		case MsgTypeDisconnect:
			log.Println("Disconnect requested by admin")
			return
		case MsgTypeCommand:
			// Handle commands (future implementation)
			log.Println("Received command (not implemented)")
		}
	}
}

func (w *WorkerServer) sendSystemInfo(conn net.Conn) {
	log.Println("WORKER: Gathering system information...")
	w.sysInfo = system.GetLocalSystemInfo()

	// Get local IP address
	localIP := getLocalIP()
	log.Printf("WORKER: Local IP detected: %s\n", localIP)

	payload := SystemInfoPayload{
		Hostname:     w.sysInfo.Hostname,
		OS:           w.sysInfo.OS,
		Architecture: w.sysInfo.Arch,
		IPAddress:    localIP,
		CPUUsage:     w.sysInfo.CPUUsage,
		RAMUsage:     w.sysInfo.RAMUsage,
	}

	log.Printf("WORKER: System Info - Hostname: %s, OS: %s, Arch: %s\n",
		payload.Hostname, payload.OS, payload.Architecture)

	payloadBytes, _ := json.Marshal(payload)
	msg := Message{
		Type:    MsgTypeSystemInfo,
		Payload: payloadBytes,
	}

	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(msg); err != nil {
		log.Printf("WORKER ERROR: Failed to send system info: %v\n", err)
	} else {
		log.Println("WORKER: System info sent successfully")
	}
}

func (w *WorkerServer) sendPong(conn net.Conn) {
	msg := Message{Type: MsgTypePong}
	encoder := json.NewEncoder(conn)
	encoder.Encode(msg)
}

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
