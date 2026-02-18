# admin:admin

A Go-based desktop application with GUI for managing Admin and Worker nodes using the Fyne framework with TCP networking support and SSH remote access.

**Maximum Compatibility** - Uses software rendering by default for universal device support. See [OPENGL_FIX.md](OPENGL_FIX.md) for details.

## Table of Contents

- [Quick Start](#quick-start)
- [Requirements](#requirements)
- [Building](#building)
- [Features](#features)
- [SSH Remote Access](#ssh-remote-access)
- [Networking](#networking)
- [Verbose Logging](#verbose-logging)
- [Project Structure](#project-structure)
- [Development](#development)
- [Troubleshooting](#troubleshooting)
- [Technology Stack](#technology-stack)

**Additional Documentation:**
- [QUICKSTART.md](QUICKSTART.md) - 2-minute setup guide for new users
- [FAQ.md](FAQ.md) - Frequently asked questions and troubleshooting
- [REQUIREMENTS.md](REQUIREMENTS.md) - Detailed system requirements and compatibility
- [OPENGL_FIX.md](OPENGL_FIX.md) - OpenGL compatibility and rendering modes
- [BUILD.md](BUILD.md) - Platform-specific build instructions
- [FIREWALL.md](FIREWALL.md) - Firewall configuration guide

## Quick Start

### Run the Application
```powershell
.\bin\admin-admin.exe
```

### Build the Application
```powershell
.\build.ps1 -v "1.0"
```

### Connect Two PCs

**Worker PC:**
1. Run `.\bin\admin-admin.exe`
2. Click "Worker PC"
3. Note the displayed IP address and port

**Admin PC:**
1. Run `.\bin\admin-admin.exe`
2. Click "Admin PC"
3. Enter Worker's IP address
4. Click "Connect"

## Requirements

### Minimum System Requirements

#### Hardware
- **CPU**: x86-64 processor (Intel/AMD), 1 GHz or faster
  - Single core sufficient
  - Dual-core recommended for smoother performance
- **RAM**: 100 MB available memory
  - Typical usage: 50-80 MB
  - Each SSH session adds ~20 MB
  - Each worker connection adds ~10 MB (Admin side)
- **Storage**: 30 MB free disk space
  - Executable: ~24 MB
  - Runtime cache and keys: ~5 MB
- **Network**: Ethernet or Wi-Fi adapter
  - Minimum 1 Mbps for metrics streaming
  - 10 Mbps recommended for multiple workers
- **Display**: Any resolution (minimum tested: 800x600)
  - Recommended: 1280x720 or higher
- **GPU**: Not required
  - Software rendering enabled by default
  - Optional GPU acceleration available

#### Software
- **Operating System**:
  - **Windows**: 7 SP1 or later (10/11 recommended)
  - **Linux**: Kernel 3.2+ with X11 or Wayland
  - **macOS**: 10.13 (High Sierra) or later
- **Graphics**: No OpenGL required (as much as I tested, but some reports are caused by "Fyne error:
  window creation error; Cause: APIUnavailable: WGL. The driver does not appear to support OpenGL)
  - CPU-based software rendering handles all UI
  - Works on VMs, RDP, headless systems with display
- **Network**: Standard TCP/IP stack (included in all modern OS)
- **Dependencies**: None - fully self-contained binary

#### Tested Configurations
- ✅ Windows 10/11 (physical machines)
- ✅ Windows VMs (VirtualBox, VMware, Hyper-V)
- ✅ Ubuntu 20.04+ / Debian 11+
- ✅ WSL2 with X server
- ✅ Remote Desktop (RDP, VNC)
- ✅ Machines without dedicated GPU
- ✅ Systems with integrated graphics only

### Build Requirements (Developers Only)
- Go 1.21 or later
- MSYS2 with MinGW64 (installed at `D:\msys2`) for Windows builds
- CGO enabled for GUI support
- GCC compiler (mingw-w64 on Windows, gcc on Linux, Xcode on macOS)

### Compatibility
admin:admin uses CPU-based software rendering by default, ensuring compatibility on:
- All modern Windows, Linux, and macOS systems
- Virtual machines (VMware, VirtualBox, Hyper-V, etc.)
- Remote Desktop (RDP, VNC, etc.)
- Older hardware without OpenGL support
- Systems with basic or outdated graphics drivers

Users with modern GPUs can optionally enable hardware rendering for slightly better performance (see [OPENGL_FIX.md](OPENGL_FIX.md)).

## Building

### Using Build Script (Recommended)
```powershell
# Interactive - will prompt for version
.\build.ps1

# With version specified
.\build.ps1 -v "1.0.0"
.\build.ps1 -v "dev-0.8"
```

### Manual Build
```powershell
$env:Path += ";D:\msys2\mingw64\bin"
$env:CGO_ENABLED = 1
go build -o bin/admin-admin.exe ./cmd/app
```

## Features

### Role Selection
- Choose between Admin PC and Worker PC modes
- Clean, centered UI with purple theme

### Admin Mode
- Connect to multiple remote workers via IP address
- View real-time resource monitoring (CPU, RAM, GPU)
- Radial gauge displays with smooth animations
- SSH terminal access to worker machines
- Disconnect from worker nodes
- Return to role selection

### Worker Mode
- TCP server listening on port 9876
- SSH server on port 2222
- Automatically sends system info when admin connects
- Real-time metrics streaming (1 Hz update rate)
- Display local IP and port for easy connection

### Resource Monitoring
- **CPU Usage**: Real-time CPU utilization percentage
- **RAM Usage**: Memory usage with total/used display
- **GPU Usage**: Graphics card utilization (NVIDIA, AMD, Intel)
- **System Uptime**: Time since last boot
- **Network Info**: Local IP address

## SSH Remote Access

admin:admin includes built-in SSH functionality for remote command execution.

### SSH Server (Worker Side)

When you select "Worker PC", an SSH server automatically starts:
- **Port**: 2222
- **Default Username**: `admin`
- **Default Password**: `admin`

You can customize credentials on the Worker's waiting screen before connecting.

The SSH host key is generated on first run and stored in:
- Windows: `%APPDATA%\adminadmin\ssh_host_key`

### Custom SSH Credentials

On the Worker PC, before an admin connects:
1. The waiting screen displays editable Username and Password fields
2. Modify these to set custom credentials
3. The SSH server will use these credentials for authentication

### Connecting via SSH from Admin Dashboard

1. Connect to a worker from the Admin dashboard
2. Click "Open SSH Terminal" button
3. Enter credentials:
   - **Username**: `admin` (default)
   - **Password**: `admin` (default)
4. Execute commands in the terminal interface

### SSH Terminal Features

- **Multi-tab support**: Open multiple SSH sessions in tabs
- **Command history**: Navigate previous commands
- **Copy support**: Select and copy terminal output
- **Built-in commands**: `clear`, `exit`, `help`
- **Auto-flags**: Package managers (winget, apt, choco) automatically get non-interactive flags

### Connecting via External SSH Client

You can also connect using any SSH client:

```powershell
# Using Windows OpenSSH
ssh admin@192.168.0.67 -p 2222

# Using PuTTY
# Host: 192.168.0.67
# Port: 2222
# Username: admin
# Password: admin
```

### SSH Security Notes

⚠️ **Important Security Considerations:**

1. Change the default credentials in production
2. SSH host keys are auto-generated and stored locally
3. The SSH server only runs when in Worker mode
4. Consider firewall rules to restrict SSH access

### Firewall Configuration for SSH

```powershell
# Allow SSH port (run as Administrator)
New-NetFirewallRule -DisplayName "admin:admin SSH" -Direction Inbound -Protocol TCP -LocalPort 2222 -Action Allow
```

## Networking

### Network Protocol

The system uses JSON-based TCP protocol on port 9876:

**Message Types:**
- `system_info`: Worker sends system information to Admin
- `metrics`: Real-time CPU/RAM/GPU updates (1 Hz)
- `admin_info`: Admin sends its hostname to Worker
- `ping/pong`: Keep-alive messages
- `disconnect`: Graceful disconnection

### Ports Used

| Port | Protocol | Purpose |
|------|----------|---------|
| 9876 | TCP | Main communication |
| 2222 | TCP | SSH remote access |

### Firewall Configuration

Allow the application through Windows Firewall:

```powershell
# Run as Administrator
# Main application port
New-NetFirewallRule -DisplayName "admin:admin Worker" -Direction Inbound -Protocol TCP -LocalPort 9876 -Action Allow

# SSH port
New-NetFirewallRule -DisplayName "admin:admin SSH" -Direction Inbound -Protocol TCP -LocalPort 2222 -Action Allow
```

## Verbose Logging

The application includes comprehensive console logging for debugging.

### What Gets Logged

**Application Lifecycle:**
```
=== APPLICATION STARTING ===
APP: Window created (900x600)
APP: Showing role selection screen
APP: Role selection screen displayed
=== APPLICATION SHUTTING DOWN ===
```

**Worker Mode:**
```
=== USER SELECTED: WORKER ROLE ===
APP: Creating worker server on port 9876...
=== WORKER: Starting server ===
SUCCESS: Worker server listening on port 9876
WORKER: Waiting for connection...
WORKER: New connection from 192.168.1.100:xxxxx
WORKER: Sending system info to admin...
```

**Admin Mode:**
```
=== USER SELECTED: ADMIN ROLE ===
=== CONNECTING TO WORKER: 192.168.1.50 ===
ADMIN: Attempting to connect to worker at 192.168.1.50:9876...
ADMIN: TCP connection established
ADMIN: Received message type: system_info
APP: Received device info update callback
```

**Disconnect:**
```
=== DISCONNECT REQUESTED ===
APP: Disconnecting admin client...
ADMIN: Sending disconnect message to worker...
ADMIN: Disconnected successfully
```

### Log Prefixes

- `MAIN:` - Main application entry point
- `APP:` - Application logic layer
- `ADMIN:` - Admin client operations
- `WORKER:` - Worker server operations

### Log Format

All logs include date, time with microseconds, source file, and line number:
```
2026/02/12 17:30:45.123456 app.go:26: === APPLICATION STARTING ===
```

### Saving Logs to File

```powershell
# Save logs
.\bin\control-system.exe > debug.log 2>&1

# Save and view simultaneously
.\bin\control-system.exe 2>&1 | Tee-Object -FilePath debug.log
```

## Project Structure

```
adminadmin/
├── bin/
│   └── control-system.exe     # Compiled executable
├── cmd/
│   └── app/
│       └── main.go             # Application entry point
├── internal/
│   ├── application/
│   │   └── app.go              # Application logic and navigation
│   ├── network/
│   │   ├── protocol.go         # Network protocol definitions
│   │   ├── worker.go           # Worker TCP server
│   │   └── admin.go            # Admin TCP client
│   ├── state/
│   │   └── state.go            # Application state management
│   ├── system/
│   │   └── info.go             # System information gathering
│   └── ui/
│       ├── admin_dashboard.go  # Admin interface
│       ├── role_select.go      # Role selection screen
│       └── worker_dashboard.go # Worker interface
├── build.ps1                   # Build script
├── run.ps1                     # Run script
├── go.mod                      # Go module definition
└── go.sum                      # Dependency checksums
```

## Architecture

### Design Principles

**Separation of Concerns:**
- main.go: Minimal bootstrap code, creates Fyne app and delegates to application package
- app.go: Manages window lifecycle, screen transitions, coordinates between UI and state
- state.go: Centralized state management with no UI dependencies
- system/info.go: System information gathering with no UI or state dependencies
- network/*: TCP networking layer for Admin-Worker communication
- UI files: Pure presentation logic, return fyne.CanvasObject, accept callbacks

**No Global Variables:**
All state is encapsulated in the AppState struct and passed through the application layer.

**Modular Design:**
Each component has a single responsibility:
- State management (state package)
- System information (system package)
- Networking (network package)
- UI rendering (ui package)
- Application coordination (application package)

### Components

**State Management:**
- Tracks current role (None, Admin, or Worker)
- Stores connected device information
- Provides connection state checking

**System Information:**
- Retrieves hostname, OS, architecture, Go runtime version
- CPU and RAM usage (placeholder values currently)

**Networking:**
- Worker TCP server listens on port 9876
- Admin TCP client connects to worker
- JSON-based message protocol
- Real-time system info exchange

**User Interface:**
- Role Selection: Clean centered layout with two role buttons
- Admin Dashboard: Connection input, status display, device info, control buttons
- Worker Dashboard: System info display, server status

## Development

### Run in Development Mode
```powershell
.\run.ps1
```

### Build for Production
```powershell
.\build.ps1
```

### Check for Issues
```powershell
go vet ./...
```

### Format Code
```powershell
go fmt ./...
```

### Run Tests
```powershell
go test ./...
```

### Clean Build
```powershell
Remove-Item -Force bin\admin-admin*.exe
.\build.ps1
```

## Troubleshooting

**For common issues and quick fixes, see [FAQ.md](FAQ.md)**

### Quick Reference

**Can't connect to worker?**
- Check firewall on Worker PC (ports 9876, 2222)
- Verify both PCs on same network  
- Use correct IP address (run `ipconfig` on Worker)
- See [FAQ.md](FAQ.md#-cant-connect-to-worker)

**SSH not working?**
- Default credentials: `admin` / `admin`
- Check port 2222 firewall rule
- See [FAQ.md](FAQ.md#-ssh-connection-fails)

**OpenGL error?** (Very rare in v1.0.6)
```powershell
$env:FYNE_DISABLE_HARDWARE_RENDERING="1"
.\admin-admin.exe
```
See [FAQ.md](FAQ.md#-opengl-error-rare) | [OPENGL_FIX.md](OPENGL_FIX.md)

**GPU shows N/A?**
- Normal for integrated graphics, VMs, older hardware
- See [FAQ.md](FAQ.md#-gpu-shows-na-or-0)

**Application won't start?**
- Run from terminal to see errors
- Check 64-bit OS, 100 MB RAM
- See [FAQ.md](FAQ.md#-application-wont-start)

**For complex/unique issues:**
- [GitHub Issues](https://github.com/yourusername/adminadmin/issues)

---

### Build Issues

**"gcc not found" error:**

**Solutions:**

**Windows:**
1. Install MinGW64, MSYS2, or TDM-GCC
2. Add compiler bin directory to system PATH
3. Restart terminal
4. Verify with: `gcc --version`
5. Run `.\build.ps1`

**Linux:**
```bash
# Debian/Ubuntu
sudo apt-get install gcc pkg-config libgl1-mesa-dev xorg-dev

# Fedora
sudo dnf install gcc libX11-devel libXcursor-devel libXrandr-devel

# Arch
sudo pacman -S gcc libx11 libxcursor libxrandr
```

**macOS:**
```bash
xcode-select --install
```

See BUILD.md for detailed platform-specific instructions.

### SSH Connection Issues

**Problem:** SSH terminal fails to connect

**Solutions:**
1. Verify credentials (default: `admin` / `admin`)
2. Check worker is in Worker mode
3. Check port 2222 is not blocked by firewall
4. Ensure SSH server started (check worker logs)
5. Try external SSH client to verify: `ssh admin@<worker-ip> -p 2222`

### GPU Shows as "N/A"

**Problem:** GPU usage always shows 0% or N/A

**Cause:** GPU monitoring requires vendor-specific libraries:
- NVIDIA: `nvidia-smi` must be in PATH
- AMD: Limited support on Windows
- Intel: Limited integrated GPU monitoring

**Solutions:**
1. Install GPU vendor tools (NVIDIA drivers include nvidia-smi)
2. GPU monitoring works best on NVIDIA GPUs
3. This is a known limitation, CPU/RAM monitoring works reliably

### Worker Port Already in Use

**Problem:** Port 9876 or 2222 already in use

**Solutions:**
1. Check if another instance is running:
   ```powershell
   netstat -ano | findstr "9876"
   netstat -ano | findstr "2222"
   ```
2. Kill the process using the port
3. Restart the application

### No System Info Displayed

**Problem:** Connected but no device info shows

**Solutions:**
1. Wait a few seconds for metrics to update
2. Check console logs for errors
3. Disconnect and reconnect
4. Restart both applications
5. Verify network connectivity

### Window Doesn't Appear

**Problem:** Application runs but window doesn't show

**Solutions:**
1. Check if application is running in background (Task Manager)
2. Kill any existing admin-admin.exe processes
3. Check for OpenGL errors in console
4. Software rendering should fix this (enabled by default)
5. Try running from command prompt to see error messages

### Build Errors on Linux

**Problem:** Build fails with `unknown field HideWindow`

**Cause:** Windows-specific code in cross-platform build

**Solution:** Use platform-specific builds:
```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o bin/admin-admin ./cmd/app

# Windows from Linux
GOOS=windows GOARCH=amd64 go build -o bin/admin-admin.exe ./cmd/app
```

## Technology Stack

- **Language:** Go 1.25
- **GUI Framework:** Fyne v2.5.3
- **Build System:** Go modules
- **Compiler:** GCC 15.2.0 (MSYS2 MinGW64)
- **Protocol:** TCP with JSON messaging
- **Default Port:** 9876

## Security Notes

This is a basic implementation for local network use. For production/internet use, add:

1. **Authentication:** Password or key-based auth
2. **Encryption:** TLS/SSL for network communication
3. **Authorization:** Role-based access control
4. **Input Validation:** Sanitize all inputs
5. **Rate Limiting:** Prevent DOS attacks

## Testing Checklist

- [ ] Application builds successfully
- [ ] Application launches without errors
- [ ] Role selection screen displays
- [ ] Worker mode starts TCP server on port 9876
- [ ] Admin mode shows IP input field
- [ ] Can connect using 127.0.0.1 on same PC
- [ ] Can connect using local IP on different PCs
- [ ] Connection status updates correctly
- [ ] Worker's system information displays correctly
- [ ] Disconnect button works
- [ ] Refresh button updates display
- [ ] Back to role selection works on both sides
- [ ] Console logs show detailed information

## Status

The application is fully operational with:
- Complete GUI interface
- TCP networking between Admin and Worker
- Real-time system information sharing
- Connection management
- Verbose console logging for debugging

Built with love by forkosssa, readme and code assistance by claude.

