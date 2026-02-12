# Cross-Platform Build Instructions

This project includes build scripts for Windows, Linux, and macOS.

## Prerequisites

### All Platforms
- Go 1.25 or later
- CGO-compatible C compiler

### Windows
Install one of:
- **MSYS2 with MinGW64** (recommended)
  - Download from: https://www.msys2.org/
  - Install packages: `pacman -S mingw-w64-x86_64-gcc`
  - Add to PATH: `C:\msys64\mingw64\bin` (or your install location)
  
- **TDM-GCC**
  - Download from: https://jmeubank.github.io/tdm-gcc/
  - Add to PATH: `C:\TDM-GCC-64\bin`

- **MinGW-w64**
  - Download from: https://www.mingw-w64.org/
  - Add bin directory to PATH

### Linux
```bash
# Debian/Ubuntu
sudo apt-get install gcc pkg-config libgl1-mesa-dev xorg-dev

# Fedora
sudo dnf install gcc libX11-devel libXcursor-devel libXrandr-devel libXinerama-devel mesa-libGL-devel libXi-devel

# Arch
sudo pacman -S gcc libx11 libxcursor libxrandr libxinerama mesa libxi
```

### macOS
```bash
# Install Xcode Command Line Tools
xcode-select --install
```

## Building
`
### Windows`
```powershell
.\build.ps1
```

### Linux/macOS
```bash
chmod +x build.sh
./build.sh
```

## Running

### Windows
```powershell
.\run.ps1
# Or run the executable directly:
.\bin\control-system.exe
```

### Linux/macOS
```bash
chmod +x run.sh
./run.sh
# Or run the executable directly:
./bin/control-system
```

## Verifying Your Setup

Check if GCC is in your PATH:

### Windows
```powershell
gcc --version
```

### Linux/macOS
```bash
gcc --version
```

If you see version information, you're ready to build!

## Silent Mode (No Console)

### Windows
Build without console window:
```powershell
$env:CGO_ENABLED = 1
go build -ldflags="-H windowsgui" -o bin/control-system.exe .\cmd\app
```

### Linux/macOS
No special flags needed - terminal apps don't show console by default.

## Troubleshooting

### "gcc not found"
- **Windows:** Make sure GCC is installed and its bin directory is in your PATH
- **Linux:** Install gcc and required development libraries
- **macOS:** Run `xcode-select --install`

### "CGO_ENABLED not set"
The build scripts automatically set this. If building manually, ensure:
```bash
export CGO_ENABLED=1  # Linux/macOS
$env:CGO_ENABLED = 1  # Windows PowerShell
```

### Build works but executable doesn't run (Linux)
Make sure the executable has execute permissions:
```bash
chmod +x bin/control-system
```

