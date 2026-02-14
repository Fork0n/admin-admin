package network

import (
	"adminadmin/internal/system"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

const (
	DefaultWorkerPort = 9876
)

// WorkerServer represents a worker node server
type WorkerServer struct {
	listener          net.Listener
	port              int
	quit              chan bool
	sysInfo           system.SystemInfo
	activeConn        net.Conn
	connMu            sync.Mutex
	onAdminConnect    func(hostname string)
	onAdminDisconnect func()
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

// SetCallbacks sets the callbacks for admin connection events
func (w *WorkerServer) SetCallbacks(onConnect func(hostname string), onDisconnect func()) {
	w.onAdminConnect = onConnect
	w.onAdminDisconnect = onDisconnect
}

// Start starts the worker server
func (w *WorkerServer) Start() error {
	log.Println("=== WORKER: Starting server ===")

	// Get local IP for logging
	localIP := getLocalIP()
	log.Printf("WORKER: Local IP address detected: %s\n", localIP)

	// Listen on all interfaces (0.0.0.0) to allow remote connections
	bindAddr := fmt.Sprintf("0.0.0.0:%d", w.port)
	listener, err := net.Listen("tcp", bindAddr)
	if err != nil {
		log.Printf("ERROR: Failed to start worker server: %v\n", err)
		log.Println("ERROR: Port 9876 might already be in use by another application")
		return fmt.Errorf("failed to start worker server: %w", err)
	}
	w.listener = listener

	log.Printf("SUCCESS: Worker server listening on %s (port %d)\n", bindAddr, w.port)
	log.Printf("SUCCESS: Admins can connect using IP: %s\n", localIP)
	log.Println("")
	log.Println("=== IMPORTANT: FIREWALL SETUP ===")
	log.Println("If admin cannot connect, run this command as Administrator:")
	log.Printf("  New-NetFirewallRule -DisplayName \"admin:admin Worker\" -Direction Inbound -Protocol TCP -LocalPort %d -Action Allow\n", w.port)
	log.Println("=================================")
	log.Println("")
	log.Println("Waiting for admin connections...")

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

// GetLocalIP returns the local IP address
func (w *WorkerServer) GetLocalIP() string {
	return getLocalIP()
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
	w.connMu.Lock()
	w.activeConn = conn
	w.connMu.Unlock()

	defer func() {
		conn.Close()
		w.connMu.Lock()
		w.activeConn = nil
		w.connMu.Unlock()
		if w.onAdminDisconnect != nil {
			w.onAdminDisconnect()
		}
	}()

	log.Printf("Admin connected from: %s\n", conn.RemoteAddr())

	// Send system info immediately upon connection
	w.sendSystemInfo(conn)

	// Start sending metrics updates every 2 seconds
	stopMetrics := make(chan bool)
	go w.sendMetricsLoop(conn, stopMetrics)

	// Keep connection alive and handle incoming messages
	decoder := json.NewDecoder(conn)
	for {
		var msg Message
		if err := decoder.Decode(&msg); err != nil {
			log.Printf("Connection closed: %v\n", err)
			close(stopMetrics)
			return
		}

		switch msg.Type {
		case MsgTypePing:
			w.sendPong(conn)
		case MsgTypeAdminInfo:
			var adminInfo AdminInfoPayload
			if err := json.Unmarshal(msg.Payload, &adminInfo); err == nil {
				log.Printf("WORKER: Admin identified as: %s\n", adminInfo.Hostname)
				if w.onAdminConnect != nil {
					w.onAdminConnect(adminInfo.Hostname)
				}
			}
		case MsgTypeDisconnect:
			log.Println("Disconnect requested by admin")
			close(stopMetrics)
			return
		case MsgTypeCommand:
			// Handle commands (future implementation)
			log.Println("Received command (not implemented)")
		}
	}
}

func (w *WorkerServer) sendMetricsLoop(conn net.Conn, stop chan bool) {
	ticker := time.NewTicker(1 * time.Second) // 1 Hz polling rate
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-w.quit:
			return
		case <-ticker.C:
			cpuUsage, ramUsage, gpuUsage := system.GetRealTimeMetrics()
			payload := MetricsPayload{
				CPUUsage: cpuUsage,
				RAMUsage: ramUsage,
				GPUUsage: gpuUsage,
			}
			payloadBytes, _ := json.Marshal(payload)
			msg := Message{
				Type:    MsgTypeMetrics,
				Payload: payloadBytes,
			}
			encoder := json.NewEncoder(conn)
			if err := encoder.Encode(msg); err != nil {
				log.Printf("WORKER: Failed to send metrics: %v\n", err)
				return
			}
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
		Hostname:      w.sysInfo.Hostname,
		OS:            w.sysInfo.OS,
		Architecture:  w.sysInfo.Arch,
		IPAddress:     localIP,
		CPUUsage:      w.sysInfo.CPUUsage,
		RAMUsage:      w.sysInfo.RAMUsage,
		RAMTotal:      w.sysInfo.RAMTotal,
		RAMUsed:       w.sysInfo.RAMUsed,
		GPUName:       w.sysInfo.GPUName,
		GPUUsage:      w.sysInfo.GPUUsage,
		InternetSpeed: w.sysInfo.InternetSpeed,
		Uptime:        w.sysInfo.Uptime,
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
