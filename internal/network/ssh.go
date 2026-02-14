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
	"strings"
	"sync"

	"golang.org/x/crypto/ssh"
)

const (
	DefaultSSHPort     = 2222 // Use non-standard port to avoid conflicts
	DefaultSSHUsername = "admin"
	DefaultSSHPassword = "admin"
)

// SSHCredentials holds SSH login credentials
type SSHCredentials struct {
	Username string
	Password string
}

// SSHServer provides SSH access to the worker
type SSHServer struct {
	listener    net.Listener
	port        int
	quit        chan bool
	config      *ssh.ServerConfig
	mu          sync.Mutex
	running     bool
	credentials SSHCredentials
}

// NewSSHServer creates a new SSH server
func NewSSHServer(port int) *SSHServer {
	if port == 0 {
		port = DefaultSSHPort
	}
	return &SSHServer{
		port: port,
		quit: make(chan bool),
		credentials: SSHCredentials{
			Username: DefaultSSHUsername,
			Password: DefaultSSHPassword,
		},
	}
}

// SetCredentials sets custom SSH credentials
func (s *SSHServer) SetCredentials(username, password string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if username != "" {
		s.credentials.Username = username
	}
	if password != "" {
		s.credentials.Password = password
	}
}

// GetCredentials returns the current SSH credentials
func (s *SSHServer) GetCredentials() SSHCredentials {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.credentials
}

// Start starts the SSH server
func (s *SSHServer) Start(password string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("SSH server already running")
	}

	// Use provided password or default
	if password != "" {
		s.credentials.Password = password
	}

	// Generate or load host key
	hostKey, err := getOrCreateHostKey()
	if err != nil {
		return fmt.Errorf("failed to get host key: %w", err)
	}

	// Configure SSH server with credentials check
	expectedUser := s.credentials.Username
	expectedPass := s.credentials.Password
	s.config = &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			// Check username and password
			if c.User() == expectedUser && string(pass) == expectedPass {
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

	// For commands that might need auto-accept, modify them
	modifiedCmd := cmdStr

	if runtime.GOOS == "windows" {
		// Auto-accept for winget - add all necessary flags to avoid prompts
		lowerCmd := strings.ToLower(cmdStr)
		if strings.Contains(lowerCmd, "winget") {
			if strings.Contains(lowerCmd, "install") || strings.Contains(lowerCmd, "upgrade") {
				if !strings.Contains(lowerCmd, "--accept-source-agreements") {
					modifiedCmd += " --accept-source-agreements"
				}
				if !strings.Contains(lowerCmd, "--accept-package-agreements") {
					modifiedCmd += " --accept-package-agreements"
				}
				// Disable interactive mode
				if !strings.Contains(lowerCmd, "--disable-interactivity") {
					modifiedCmd += " --disable-interactivity"
				}
			}
		}
		// For choco
		if strings.Contains(lowerCmd, "choco") && strings.Contains(lowerCmd, "install") {
			if !strings.Contains(lowerCmd, "-y") {
				modifiedCmd += " -y"
			}
		}
		cmd = exec.Command("cmd.exe", "/c", modifiedCmd)
	} else {
		// For apt-get, add -y flag
		if strings.Contains(cmdStr, "apt-get install") && !strings.Contains(cmdStr, "-y") {
			modifiedCmd = strings.Replace(cmdStr, "apt-get install", "apt-get install -y", 1)
		}
		// For apt install
		if strings.Contains(cmdStr, "apt install") && !strings.Contains(cmdStr, "-y") {
			modifiedCmd = strings.Replace(cmdStr, "apt install", "apt install -y", 1)
		}
		// For yum/dnf
		if (strings.Contains(cmdStr, "yum install") || strings.Contains(cmdStr, "dnf install")) && !strings.Contains(cmdStr, "-y") {
			modifiedCmd = strings.Replace(strings.Replace(modifiedCmd, "yum install", "yum install -y", 1), "dnf install", "dnf install -y", 1)
		}
		cmd = exec.Command("/bin/sh", "-c", modifiedCmd)
	}

	// Use NUL or /dev/null for stdin to prevent "reading input" errors
	if runtime.GOOS == "windows" {
		devNull, err := os.Open("NUL")
		if err == nil {
			cmd.Stdin = devNull
			defer devNull.Close()
		} else {
			cmd.Stdin = strings.NewReader("")
		}
	} else {
		devNull, err := os.Open("/dev/null")
		if err == nil {
			cmd.Stdin = devNull
			defer devNull.Close()
		} else {
			cmd.Stdin = strings.NewReader("")
		}
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

	// Request a pseudo-terminal for better command compatibility
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // Disable echo
		ssh.TTY_OP_ISPEED: 14400, // Input speed
		ssh.TTY_OP_OSPEED: 14400, // Output speed
	}

	// Request PTY - this helps with some interactive commands
	if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
		// PTY failed, try without it
		log.Printf("SSH Client: PTY request failed, continuing without: %v", err)
	}

	// For commands that might prompt for input, we need to handle them differently
	// Run with CombinedOutput for simple commands
	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return string(output), fmt.Errorf("command failed: %w", err)
	}

	return string(output), nil
}

// ExecuteCommandWithInput executes a command and sends input to it
func (c *SSHClient) ExecuteCommandWithInput(cmd string, input string) (string, error) {
	if c.client == nil {
		return "", fmt.Errorf("not connected")
	}

	session, err := c.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// Set up stdin pipe
	stdin, err := session.StdinPipe()
	if err != nil {
		return "", fmt.Errorf("failed to get stdin: %w", err)
	}

	// Request a pseudo-terminal
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
		log.Printf("SSH Client: PTY request failed: %v", err)
	}

	// Get output
	output, err := session.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed to get stdout: %w", err)
	}
	stderr, err := session.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("failed to get stderr: %w", err)
	}

	// Start the command
	if err := session.Start(cmd); err != nil {
		return "", fmt.Errorf("failed to start command: %w", err)
	}

	// Send input if provided
	if input != "" {
		_, err = stdin.Write([]byte(input + "\n"))
		if err != nil {
			log.Printf("SSH Client: Failed to write input: %v", err)
		}
	}
	stdin.Close()

	// Read all output
	var result []byte
	buf := make([]byte, 4096)
	for {
		n, err := output.Read(buf)
		if n > 0 {
			result = append(result, buf[:n]...)
		}
		if err != nil {
			break
		}
	}

	// Also read stderr
	for {
		n, err := stderr.Read(buf)
		if n > 0 {
			result = append(result, buf[:n]...)
		}
		if err != nil {
			break
		}
	}

	session.Wait()
	return string(result), nil
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
