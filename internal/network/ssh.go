package network

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"

	"golang.org/x/crypto/ssh"
)

const (
	DefaultSSHPort = 2222 // Use non-standard port to avoid conflicts
)

// SSHServer provides SSH access to the worker
type SSHServer struct {
	listener net.Listener
	port     int
	quit     chan bool
	config   *ssh.ServerConfig
	mu       sync.Mutex
	running  bool
}

// NewSSHServer creates a new SSH server
func NewSSHServer(port int) *SSHServer {
	if port == 0 {
		port = DefaultSSHPort
	}
	return &SSHServer{
		port: port,
		quit: make(chan bool),
	}
}

// Start starts the SSH server
func (s *SSHServer) Start(password string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("SSH server already running")
	}

	// Generate or load host key
	hostKey, err := getOrCreateHostKey()
	if err != nil {
		return fmt.Errorf("failed to get host key: %w", err)
	}

	// Configure SSH server
	s.config = &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			if string(pass) == password {
				log.Printf("SSH: User %s authenticated successfully\n", c.User())
				return nil, nil
			}
			log.Printf("SSH: Failed authentication attempt for user %s\n", c.User())
			return nil, fmt.Errorf("password rejected")
		},
	}
	s.config.AddHostKey(hostKey)

	// Listen
	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", s.port))
	if err != nil {
		return fmt.Errorf("failed to start SSH server: %w", err)
	}
	s.listener = listener
	s.running = true

	log.Printf("SSH: Server listening on port %d\n", s.port)

	go s.acceptConnections()
	return nil
}

// Stop stops the SSH server
func (s *SSHServer) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	s.running = false
	close(s.quit)
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

// GetPort returns the SSH port
func (s *SSHServer) GetPort() int {
	return s.port
}

// IsRunning returns whether the SSH server is running
func (s *SSHServer) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

func (s *SSHServer) acceptConnections() {
	for {
		select {
		case <-s.quit:
			return
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				select {
				case <-s.quit:
					return
				default:
					log.Printf("SSH: Error accepting connection: %v\n", err)
					continue
				}
			}
			go s.handleConnection(conn)
		}
	}
}

func (s *SSHServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Perform SSH handshake
	sshConn, chans, reqs, err := ssh.NewServerConn(conn, s.config)
	if err != nil {
		log.Printf("SSH: Handshake failed: %v\n", err)
		return
	}
	defer sshConn.Close()

	log.Printf("SSH: New connection from %s (%s)\n", sshConn.RemoteAddr(), sshConn.ClientVersion())

	// Discard out-of-band requests
	go ssh.DiscardRequests(reqs)

	// Handle channels
	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}

		channel, requests, err := newChannel.Accept()
		if err != nil {
			log.Printf("SSH: Could not accept channel: %v\n", err)
			continue
		}

		go s.handleChannel(channel, requests)
	}
}

func (s *SSHServer) handleChannel(channel ssh.Channel, requests <-chan *ssh.Request) {
	defer channel.Close()

	for req := range requests {
		switch req.Type {
		case "shell":
			req.Reply(true, nil)
			s.startShell(channel)
			return
		case "exec":
			req.Reply(true, nil)
			if len(req.Payload) > 4 {
				cmdLen := int(req.Payload[0])<<24 | int(req.Payload[1])<<16 | int(req.Payload[2])<<8 | int(req.Payload[3])
				if cmdLen > 0 && cmdLen <= len(req.Payload)-4 {
					cmd := string(req.Payload[4 : 4+cmdLen])
					s.executeCommand(channel, cmd)
				}
			}
			return
		case "pty-req":
			req.Reply(true, nil)
		default:
			req.Reply(false, nil)
		}
	}
}

func (s *SSHServer) startShell(channel ssh.Channel) {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd.exe")
	} else {
		shell := os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/sh"
		}
		cmd = exec.Command(shell)
	}

	cmd.Stdin = channel
	cmd.Stdout = channel
	cmd.Stderr = channel

	if err := cmd.Start(); err != nil {
		log.Printf("SSH: Failed to start shell: %v\n", err)
		io.WriteString(channel, fmt.Sprintf("Failed to start shell: %v\r\n", err))
		return
	}

	cmd.Wait()
}

func (s *SSHServer) executeCommand(channel ssh.Channel, cmdStr string) {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd.exe", "/c", cmdStr)
	} else {
		cmd = exec.Command("/bin/sh", "-c", cmdStr)
	}

	cmd.Stdout = channel
	cmd.Stderr = channel

	if err := cmd.Run(); err != nil {
		io.WriteString(channel, fmt.Sprintf("Error: %v\r\n", err))
	}
}

// getHostKeyPath returns the path to store the SSH host key
func getHostKeyPath() string {
	// Store in user's config directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}
	keyDir := filepath.Join(configDir, "adminadmin")
	os.MkdirAll(keyDir, 0700)
	return filepath.Join(keyDir, "ssh_host_key")
}

// getOrCreateHostKey loads existing host key or generates a new one
func getOrCreateHostKey() (ssh.Signer, error) {
	keyPath := getHostKeyPath()

	// Try to load existing key
	keyData, err := os.ReadFile(keyPath)
	if err == nil {
		signer, err := ssh.ParsePrivateKey(keyData)
		if err == nil {
			log.Printf("SSH: Loaded existing host key from %s\n", keyPath)
			return signer, nil
		}
	}

	// Generate new key
	log.Println("SSH: Generating new host key...")
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	// Encode to PEM
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	// Save to file
	if err := os.WriteFile(keyPath, privateKeyPEM, 0600); err != nil {
		log.Printf("SSH: Warning - could not save host key: %v\n", err)
	} else {
		log.Printf("SSH: Host key saved to %s\n", keyPath)
	}

	// Parse and return
	signer, err := ssh.ParsePrivateKey(privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse generated key: %w", err)
	}

	return signer, nil
}

// SSHClient provides SSH client functionality for admin
type SSHClient struct {
	client  *ssh.Client
	session *ssh.Session
}

// NewSSHClient creates a new SSH client
func NewSSHClient() *SSHClient {
	return &SSHClient{}
}

// Connect connects to an SSH server
func (c *SSHClient) Connect(host string, port int, user, password string) error {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // For demo; use proper verification in production
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	c.client = client

	log.Printf("SSH Client: Connected to %s\n", addr)
	return nil
}

// ExecuteCommand executes a command on the remote server
func (c *SSHClient) ExecuteCommand(cmd string) (string, error) {
	if c.client == nil {
		return "", fmt.Errorf("not connected")
	}

	session, err := c.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return string(output), fmt.Errorf("command failed: %w", err)
	}

	return string(output), nil
}

// Close closes the SSH connection
func (c *SSHClient) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// IsConnected returns whether the client is connected
func (c *SSHClient) IsConnected() bool {
	return c.client != nil
}
