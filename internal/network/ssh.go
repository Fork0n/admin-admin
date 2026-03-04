//the code lowkey might suck but it's not that bad

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
	"time"

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

// sessionState holds per-session persistent state (e.g. working directory)
type sessionState struct {
	cwd string
}

func newSessionState() *sessionState {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return &sessionState{cwd: home}
}

func (s *SSHServer) handleChannel(channel ssh.Channel, requests <-chan *ssh.Request) {
	defer channel.Close()

	state := newSessionState()

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
					s.executeCommand(channel, cmd, state)
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

func (s *SSHServer) executeCommand(channel ssh.Channel, cmdStr string, state *sessionState) {
	var cmd *exec.Cmd
	modifiedCmd := cmdStr

	if runtime.GOOS == "windows" {
		lowerCmd := strings.ToLower(cmdStr)
		if strings.Contains(lowerCmd, "winget") {
			if strings.Contains(lowerCmd, "install") || strings.Contains(lowerCmd, "upgrade") {
				if !strings.Contains(lowerCmd, "--accept-source-agreements") {
					modifiedCmd += " --accept-source-agreements"
				}
				if !strings.Contains(lowerCmd, "--accept-package-agreements") {
					modifiedCmd += " --accept-package-agreements"
				}
				if !strings.Contains(lowerCmd, "--disable-interactivity") {
					modifiedCmd += " --disable-interactivity"
				}
			}
		}
		if strings.Contains(lowerCmd, "choco") && strings.Contains(lowerCmd, "install") {
			if !strings.Contains(lowerCmd, "-y") {
				modifiedCmd += " -y"
			}
		}
		cmd = exec.Command("cmd.exe", "/c", modifiedCmd)
	} else {
		if strings.Contains(cmdStr, "apt-get install") && !strings.Contains(cmdStr, "-y") {
			modifiedCmd = strings.Replace(cmdStr, "apt-get install", "apt-get install -y", 1)
		}
		if strings.Contains(cmdStr, "apt install") && !strings.Contains(cmdStr, "-y") {
			modifiedCmd = strings.Replace(cmdStr, "apt install", "apt install -y", 1)
		}
		if (strings.Contains(cmdStr, "yum install") || strings.Contains(cmdStr, "dnf install")) && !strings.Contains(cmdStr, "-y") {
			modifiedCmd = strings.Replace(strings.Replace(modifiedCmd, "yum install", "yum install -y", 1), "dnf install", "dnf install -y", 1)
		}
		cmd = exec.Command("/bin/sh", "-c", modifiedCmd)
	}

	if runtime.GOOS == "windows" {
		if devNull, err := os.Open("NUL"); err == nil {
			cmd.Stdin = devNull
			defer devNull.Close()
		} else {
			cmd.Stdin = strings.NewReader("")
		}
	} else {
		if devNull, err := os.Open("/dev/null"); err == nil {
			cmd.Stdin = devNull
			defer devNull.Close()
		} else {
			cmd.Stdin = strings.NewReader("")
		}
	}

	cmd.Stdout = channel
	cmd.Stderr = channel

	exitStatus := 0
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(interface{ ExitStatus() int }); ok {
				exitStatus = status.ExitStatus()
			} else {
				exitStatus = 1
			}
		} else {
			io.WriteString(channel, fmt.Sprintf("Error: %v\r\n", err))
			exitStatus = 1
		}
	}

	type exitStatusMsg struct{ Status uint32 }
	channel.SendRequest("exit-status", false, ssh.Marshal(&exitStatusMsg{
		Status: uint32(exitStatus),
	}))
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

// SSHClient connects to a worker's SSH server and executes commands.
// Each command runs in its own exec session; working-directory state is
// tracked client-side and prepended to every command.
type SSHClient struct {
	client        *ssh.Client
	cwd           string // client-tracked working directory
	remoteWindows bool
	mu            sync.Mutex
}

// NewSSHClient creates a new SSH client.
func NewSSHClient() *SSHClient {
	return &SSHClient{}
}

// Connect dials the SSH server and detects the remote OS.
func (c *SSHClient) Connect(host string, port int, user, password string) error {
	config := &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}
	addr := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	c.client = client

	// Probe OS: "ver" only works on Windows cmd.exe
	if out, err2 := c.runExec("ver"); err2 == nil && strings.Contains(strings.ToLower(out), "windows") {
		c.remoteWindows = true
		log.Println("SSH Client: remote OS = Windows")
	} else {
		c.remoteWindows = false
		log.Println("SSH Client: remote OS = Unix/Linux")
	}

	// Fetch the starting working directory — only store if it looks like a real path
	var pwdOut string
	if c.remoteWindows {
		pwdOut, _ = c.runExec("cd")
	} else {
		pwdOut, _ = c.runExec("pwd")
	}
	candidate := strings.TrimSpace(strings.NewReplacer("\r", "", "\n", "").Replace(pwdOut))
	// Only accept it as a cwd if it looks like a real path (starts with / or a drive letter)
	if isValidPath(candidate) {
		c.cwd = candidate
	}
	log.Printf("SSH Client: initial cwd = %q\n", c.cwd)

	return nil
}

// runExec opens a fresh exec session, runs cmd, returns combined output.
// This is the raw transport — no cwd injection, no filtering.
func (c *SSHClient) runExec(cmd string) (string, error) {
	if c.client == nil {
		return "", fmt.Errorf("not connected")
	}
	session, err := c.client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	log.Printf("SSH runExec: sending: %q\n", cmd)
	out, err := session.CombinedOutput(cmd)
	result := string(out)
	log.Printf("SSH runExec: output=%q err=%v\n", result, err)

	// Ignore "exit without status" noise — the output is still valid
	if err != nil {
		es := err.Error()
		if strings.Contains(es, "exited without exit status") ||
			strings.Contains(es, "exit status") ||
			strings.Contains(es, "exit signal") {
			return result, nil
		}
		return result, err
	}
	return result, nil
}

// ExecuteCommand runs cmdStr on the remote host inside the tracked cwd.
// "cd" commands update the tracked cwd rather than being run directly.
func (c *SSHClient) ExecuteCommand(cmdStr string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.client == nil {
		return "", fmt.Errorf("not connected")
	}

	trimmed := strings.TrimSpace(cmdStr)
	lower := strings.ToLower(trimmed)

	// ── cd built-in ──────────────────────────────────────────────────────
	if lower == "cd" || strings.HasPrefix(lower, "cd ") || strings.HasPrefix(lower, "cd\t") {
		target := strings.TrimSpace(trimmed[2:])

		var probeCmd string
		if c.remoteWindows {
			if target == "" {
				// bare "cd" on Windows prints current dir
				probeCmd = "cd"
			} else {
				// change drive+dir, then print new dir
				probeCmd = fmt.Sprintf("cd /d %s && cd", target)
			}
		} else {
			if target == "" || target == "~" {
				probeCmd = "cd ~ && pwd"
			} else {
				probeCmd = fmt.Sprintf("cd %s && pwd", target)
			}
		}

		// Wrap with current cwd so relative paths resolve correctly
		out, err := c.runExec(c.wrapWithCwd(probeCmd))
		if err != nil {
			return out, err
		}
		// Take the last non-empty line — on Windows "cd" prints the path, possibly
		// with a trailing prompt line; on Linux pwd prints exactly one line.
		candidate := lastNonEmptyLine(out)
		if isValidPath(candidate) {
			c.cwd = candidate
			return candidate, nil
		}
		// cd failed (bad path etc.) — return the raw output as the error message
		return strings.TrimSpace(out), nil
	}

	// ── regular command ──────────────────────────────────────────────────
	full := c.wrapWithCwd(trimmed)
	return c.runExec(full)
}

// wrapWithCwd prepends a cd command so every exec runs in the tracked directory.
func (c *SSHClient) wrapWithCwd(cmd string) string {
	if c.cwd == "" {
		return cmd
	}
	if c.remoteWindows {
		// cmd.exe /c receives the whole string; cd /d switches drive+dir
		return fmt.Sprintf("cd /d %s && %s", c.cwd, cmd)
	}
	return fmt.Sprintf("cd %s && %s", c.cwd, cmd)
}

// lastNonEmptyLine returns the last non-whitespace line in s.
func lastNonEmptyLine(s string) string {
	s = strings.ReplaceAll(s, "\r", "")
	lines := strings.Split(strings.TrimSpace(s), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if t := strings.TrimSpace(lines[i]); t != "" {
			return t
		}
	}
	return ""
}

// isValidPath returns true if s looks like an absolute filesystem path.
func isValidPath(s string) bool {
	if s == "" {
		return false
	}
	// Unix absolute path
	if strings.HasPrefix(s, "/") {
		return true
	}
	// Windows: "C:", "D:", etc.
	if len(s) >= 2 && s[1] == ':' {
		return true
	}
	return false
}

// Close terminates the SSH connection.
func (c *SSHClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.client != nil {
		err := c.client.Close()
		c.client = nil
		return err
	}
	return nil
}

// IsConnected reports whether there is an active connection.
func (c *SSHClient) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.client != nil
}

// ExecuteCommandWithInput is kept for API compatibility.
func (c *SSHClient) ExecuteCommandWithInput(cmd, input string) (string, error) {
	return c.ExecuteCommand(cmd)
}
