# admin:admin Release Log

**Current Version:** v1.0.6-stable (February 18, 2026)  
**Status:** ✅ Production Ready

---

## Version History

### v1.0.6-stable (Current - RECOMMENDED) ✅

**Release Date:** February 18, 2026

**What's New:**
- Enhanced OpenGL compatibility with configurable rendering modes
- Added `FYNE_FORCE_HARDWARE_RENDERING` environment variable for power users
- Software rendering enabled by default for maximum compatibility
- Improved documentation and troubleshooting guides

**Why Upgrade:**
- Works on 99%+ of systems (VMs, RDP, older hardware)
- Better performance control options
- All features from v1.0.2 with enhanced stability

**Breaking Changes:** None  
**Migration:** Direct replacement, no config changes needed

---

### v1.0.4-quickpatch ⚠️

**Release Date:** February 18, 2026  
**Status:** Superseded by v1.0.6

**Changes:**
- Initial OpenGL compatibility fix (software rendering only)
- Addressed "WGL: driver does not support OpenGL" errors

**Issue:** No option to use hardware rendering (fixed in v1.0.6)

---

### v1.0.2 (Original Release) ⚠️

**Release Date:** February 2026  
**Status:** Legacy - upgrade recommended

**Features:**
- Multi-worker support
- SSH terminal with multi-tab interface
- Real-time CPU/RAM/GPU monitoring with radial gauges
- Purple-themed UI
- Custom SSH credentials
- TCP networking (port 9876)
- SSH access (port 2222)

**Known Issue:** May crash on systems with missing OpenGL drivers

---

## System Requirements

### Minimum
- **CPU:** x86-64, 1 GHz, single core
- **RAM:** 100 MB available (~60 MB typical usage)
- **Disk:** 30 MB free space
- **OS:** Windows 7+, Linux 3.2+, macOS 10.13+
- **Network:** Any adapter (1 Mbps minimum)
- **GPU:** Not required (software rendering)

### Recommended
- **CPU:** Dual-core, 2 GHz
- **RAM:** 256 MB available
- **Network:** 10 Mbps
- **OS:** Windows 10/11, Ubuntu 20.04+, macOS 11+

### Compatibility
✅ Virtual Machines | ✅ Remote Desktop | ✅ No GPU | ✅ Integrated Graphics  
❌ 32-bit OS | ❌ ARM processors | ❌ Windows XP/Vista

---

## Performance

**Typical Usage:**
- CPU: <1% idle, 2-5% monitoring
- RAM: 60 MB (Admin), 55 MB (Worker)
- Network: ~5 KB/s per worker

**Scaling:**
- Up to 50 concurrent workers
- Each worker: +10 MB RAM
- Each SSH session: +20 MB RAM

---

## Quick Start

### Installation
1. Download binary for your OS
2. Run executable (no installation needed)
3. Allow firewall on Worker PC:
   ```powershell
   # Windows
   New-NetFirewallRule -DisplayName "admin:admin" -Direction Inbound -Protocol TCP -LocalPort 9876,2222 -Action Allow
   
   # Linux
   sudo ufw allow 9876/tcp && sudo ufw allow 2222/tcp
   ```

### Usage
1. **Worker PC:** Select "Worker PC", note IP address
2. **Admin PC:** Select "Admin PC", enter Worker IP, click Connect

See [QUICKSTART.md](QUICKSTART.md) for detailed guide.

---

## Upgrade Guide

**From v1.0.2 or v1.0.4 to v1.0.6:**
1. Download v1.0.6-stable binary
2. Replace old executable
3. Restart application

**Benefits:** Better compatibility, performance options, bug fixes  
**Compatibility:** Network protocol unchanged, versions can interoperate

---

## Configuration Options

### Default Behavior (Maximum Compatibility)
Software rendering enabled automatically - works on all systems.

### Force Hardware Rendering (Better Performance)
```powershell
$env:FYNE_FORCE_HARDWARE_RENDERING="1"
.\admin-admin-1.0.6-stable.exe
```

### Force Software Rendering (Already Default)
```powershell
$env:FYNE_DISABLE_HARDWARE_RENDERING="1"
.\admin-admin-1.0.6-stable.exe
```

---

## Version Comparison

| Feature | v1.0.2 | v1.0.4 | v1.0.6 |
|---------|:------:|:------:|:------:|
| Multi-worker | ✅ | ✅ | ✅ |
| SSH terminal | ✅ | ✅ | ✅ |
| Resource monitoring | ✅ | ✅ | ✅ |
| OpenGL fallback | ❌ | ✅ | ✅ |
| Hardware rendering option | - | ❌ | ✅ |
| VM/RDP compatibility | ⚠️ | ✅ | ✅ |

---

## Support

**Need Help?**
- **Common issues** → [FAQ.md](FAQ.md)
- **Complex/unique issues** → [GitHub Issues](https://github.com/yourusername/adminadmin/issues)

**Before Reporting:**
1. Check [FAQ.md](FAQ.md) and [QUICKSTART.md](QUICKSTART.md)
2. Verify you're using v1.0.6-stable
3. Run from terminal to capture error output

**Include in Bug Reports:**
- Version, OS, hardware specs
- Full error output from terminal
- Steps to reproduce

---

## Built With

- Go 1.21+ | Fyne v2.7.2 | gopsutil v3.24.5 | golang.org/x/crypto | love <3

**Platforms:** Windows 7+ | Linux 3.2+ | macOS 10.13+

---

**Last Updated:** February 18, 2026  
**Documentation:** README.md | QUICKSTART.md | FAQ.md | REQUIREMENTS.md

