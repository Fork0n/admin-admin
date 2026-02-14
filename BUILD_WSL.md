# Building admin:admin from WSL

This guide explains how to build Linux binaries using WSL (Windows Subsystem for Linux) while keeping your codebase on Windows.

## Prerequisites

1. **WSL2 installed** with a Linux distribution (Ubuntu recommended)
2. **Go installed in WSL** (not Windows Go)
3. **Required dependencies** for Fyne

## Step 1: Install WSL (if not already)

```powershell
# In PowerShell as Administrator
wsl --install -d Ubuntu
```

## Step 2: Install Go in WSL

```bash
# In WSL terminal
sudo apt update
sudo apt install -y golang-go

# Or install latest Go manually:
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Verify
go version
```

## Step 3: Install Fyne Dependencies in WSL

```bash
# Required for Fyne GUI apps
sudo apt install -y \
    gcc \
    libgl1-mesa-dev \
    xorg-dev \
    libxcursor-dev \
    libxrandr-dev \
    libxinerama-dev \
    libxi-dev \
    libxxf86vm-dev
```

## Step 4: Access Windows Files from WSL

Your Windows drives are mounted under `/mnt/` in WSL:

```bash
# Navigate to the project
cd /mnt/d/adminadmin

# List files to verify
ls -la
```

## Step 5: Build the Project

### Option A: Using the build script

```bash
cd /mnt/d/adminadmin
chmod +x build.sh
./build.sh -o l -v "1.0.0-linux"
```

### Option B: Manual build

```bash
cd /mnt/d/adminadmin
export CGO_ENABLED=1
export GOOS=linux
export GOARCH=amd64

# Create bin directory if needed
mkdir -p bin

# Build
go build -ldflags="-s -w" -o bin/admin-admin-linux ./cmd/app
```

## Step 6: Verify the Build

```bash
# Check the binary
ls -lh bin/admin-admin-linux
file bin/admin-admin-linux

# Output should show: ELF 64-bit LSB executable, x86-64
```

## Quick Reference

| Action | Command |
|--------|---------|
| Open WSL | `wsl` in PowerShell |
| Go to project | `cd /mnt/d/adminadmin` |
| Build Linux | `./build.sh -o l -v "1.0.0"` |
| Build silent | `./build.sh -s -o l -v "1.0.0"` |

## Troubleshooting

### "permission denied" on build.sh
```bash
chmod +x build.sh
```

### "go: command not found"
```bash
sudo apt install golang-go
# or add to PATH if installed manually
export PATH=$PATH:/usr/local/go/bin
```

### OpenGL/graphics errors
```bash
sudo apt install -y libgl1-mesa-dev xorg-dev
```

### Slow file access on /mnt/
This is normal - WSL accessing Windows filesystem is slower than native Linux filesystem. For faster builds, you can:
1. Copy files to WSL native filesystem (`~/projects/adminadmin`)
2. Build there
3. Copy binary back to Windows

```bash
# One-time setup
mkdir -p ~/projects
cp -r /mnt/d/adminadmin ~/projects/

# Build in WSL native fs (faster)
cd ~/projects/adminadmin
./build.sh -o l -v "1.0.0"

# Copy binary back to Windows
cp bin/admin-admin-1.0.0 /mnt/d/adminadmin/bin/
```

## Building All Platforms

| Platform | Where to Build | Command |
|----------|---------------|---------|
| Windows | Windows (PowerShell) | `.\build.ps1 -o w -v "1.0.0"` |
| Linux | WSL or Linux | `./build.sh -o l -v "1.0.0"` |
| macOS | macOS only | `./build.sh -o m -v "1.0.0"` |

