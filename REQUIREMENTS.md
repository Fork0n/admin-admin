# admin:admin System Requirements

## Quick Reference

### ✅ Will Run On

- Windows 7, 8, 10, 11
- Linux (Ubuntu, Debian, Fedora, Arch, etc.)
- macOS 10.13+
- Virtual Machines (VMware, VirtualBox, Hyper-V, QEMU)
- Remote Desktop sessions (RDP, VNC)
- Systems without GPU or OpenGL
- Low-end hardware (1 GHz CPU, 100 MB RAM)

### ❌ Will NOT Run On

- 32-bit systems (requires 64-bit OS)
- Windows XP/Vista (too old)
- macOS 10.12 or earlier
- ARM processors (currently x86-64 only)
- Headless systems without any display server

---

## Detailed Requirements

### Absolute Minimum

| Component | Requirement | Notes |
|-----------|-------------|-------|
| **CPU** | x86-64, 1 GHz | Any modern 64-bit processor |
| **RAM** | 100 MB free | App uses ~60 MB typically |
| **Disk** | 30 MB free | For executable and cache |
| **Network** | Any adapter | Wi-Fi or Ethernet |
| **OS** | 64-bit only | Windows 7+, Linux 3.2+, macOS 10.13+ |
| **GPU** | None | Software rendering included |

### Recommended

| Component | Recommendation | Why |
|-----------|---------------|-----|
| **CPU** | Dual-core 2 GHz | Smoother UI animations |
| **RAM** | 256 MB free | Comfortable with SSH sessions |
| **Network** | 10 Mbps+ | Better for multiple workers |
| **Display** | 1920x1080 | Optimal UI layout |
| **OS** | Win 10/11, Ubuntu 20.04+ | Best tested platforms |

---

## Performance Scaling

### Single Worker Connection
- **CPU**: <1% idle, 2-5% active monitoring
- **RAM**: 60 MB (Admin) + 55 MB (Worker)
- **Network**: ~5 KB/s continuous

### Multiple Workers (Admin Side)
- **10 Workers**: 160 MB RAM, ~50 KB/s
- **25 Workers**: 310 MB RAM, ~125 KB/s
- **50 Workers**: 560 MB RAM, ~250 KB/s (practical limit)

### SSH Sessions
- **Per Tab**: +20 MB RAM, +5 KB/s during active use
- **Recommended**: Close unused tabs
- **Limit**: No hard limit, but 5-10 concurrent sessions is practical

---

## Compatibility Matrix

### Operating Systems

| OS | Version | Status | Notes |
|---|---|---|---|
| **Windows 10** | 1809+ | ✅ Fully Supported | Recommended |
| **Windows 11** | All | ✅ Fully Supported | Recommended |
| **Windows 8.1** | All | ✅ Compatible | Tested |
| **Windows 7** | SP1+ | ⚠️ Compatible | End of life, use at own risk |
| **Ubuntu** | 20.04+ | ✅ Fully Supported | Recommended |
| **Debian** | 11+ | ✅ Fully Supported | Tested |
| **Fedora** | 35+ | ✅ Compatible | Should work |
| **Arch Linux** | Rolling | ✅ Compatible | Should work |
| **macOS** | 10.13-10.15 | ⚠️ Compatible | Older versions |
| **macOS** | 11+ | ✅ Compatible | Recommended |
| **Windows XP/Vista** | All | ❌ Not Supported | Too old |
| **macOS** | <10.13 | ❌ Not Supported | Too old |

### Virtualization

| Platform | Status | Notes |
|---|---|---|
| **VirtualBox** | ✅ Fully Supported | 3D acceleration not required |
| **VMware Workstation** | ✅ Fully Supported | Works with/without GPU passthrough |
| **Hyper-V** | ✅ Fully Supported | Windows VMs tested |
| **QEMU/KVM** | ✅ Compatible | Should work |
| **Parallels (Mac)** | ✅ Compatible | Should work |
| **WSL2** | ⚠️ With X Server | Requires VcXsrv, Xming, or X410 |
| **Docker** | ❌ Not Recommended | GUI apps in containers are complex |

### Remote Access

| Method | Status | Notes |
|---|---|---|
| **RDP (Windows)** | ✅ Fully Supported | Software rendering works perfectly |
| **VNC** | ✅ Fully Supported | All VNC servers compatible |
| **TeamViewer** | ✅ Compatible | Should work |
| **AnyDesk** | ✅ Compatible | Should work |
| **Chrome Remote Desktop** | ✅ Compatible | Should work |
| **SSH X11 Forwarding** | ⚠️ Slow | Works but laggy over network |

### Graphics Hardware

| GPU Type | Status | Notes |
|---|---|---|
| **No GPU** | ✅ Fully Supported | Software rendering handles everything |
| **Integrated (Intel HD)** | ✅ Fully Supported | GPU monitoring may show N/A |
| **NVIDIA** | ✅ Fully Supported | GPU monitoring works with drivers |
| **AMD** | ✅ Fully Supported | GPU monitoring works with drivers |
| **Intel Iris/Xe** | ✅ Fully Supported | Modern integrated graphics |
| **Very Old GPU** | ✅ Compatible | Software rendering bypasses GPU issues |

---

## Network Requirements

### Ports

| Port | Protocol | Direction | Purpose | Required |
|---|---|---|---|---|
| **9876** | TCP | Inbound (Worker) | Main communication | Yes |
| **2222** | TCP | Inbound (Worker) | SSH access | Optional |

### Bandwidth

| Scenario | Bandwidth | Latency | Notes |
|---|---|---|---|
| **Single Worker** | ~5 KB/s | <100ms | Comfortable |
| **10 Workers** | ~50 KB/s | <100ms | Smooth |
| **50 Workers** | ~250 KB/s | <100ms | Near limit |
| **LAN (Gigabit)** | ✅ Ideal | <1ms | Best experience |
| **Wi-Fi (100 Mbps)** | ✅ Great | <10ms | Recommended |
| **Wi-Fi (10 Mbps)** | ⚠️ OK | <50ms | May lag with many workers |
| **Internet (WAN)** | ⚠️ Not Recommended | Varies | Security risk, use VPN |

---

## Firewall Configuration

### Worker PC (Required)

**Windows:**
```powershell
New-NetFirewallRule -DisplayName "admin:admin Main" -Direction Inbound -Protocol TCP -LocalPort 9876 -Action Allow
New-NetFirewallRule -DisplayName "admin:admin SSH" -Direction Inbound -Protocol TCP -LocalPort 2222 -Action Allow
```

**Linux (UFW):**
```bash
sudo ufw allow 9876/tcp
sudo ufw allow 2222/tcp
```

**Linux (iptables):**
```bash
sudo iptables -A INPUT -p tcp --dport 9876 -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 2222 -j ACCEPT
```

### Admin PC

Usually no configuration needed (outbound connections allowed by default).

---

## Disk Space Breakdown

| Item | Size | Purpose |
|---|---|---|
| Executable | ~24 MB | Application binary |
| SSH Host Key | ~2 KB | Worker SSH server |
| Config Cache | ~1 MB | Fyne UI preferences |
| Logs (if saved) | Variable | Optional debug output |
| **Total** | **~30 MB** | Maximum installation size |

---

## Special Considerations

### Running on Very Old Hardware

If your hardware is extremely limited:

1. **Close unnecessary programs** before running
2. **Limit worker connections** (Admin mode)
3. **Avoid multiple SSH sessions**
4. **Use hardware rendering** if you have a working GPU:
   ```powershell
   $env:FYNE_FORCE_HARDWARE_RENDERING="1"
   .\admin-admin.exe
   ```

### Running on High-End Hardware

If you want maximum performance:

1. **Enable hardware rendering**:
   ```powershell
   $env:FYNE_FORCE_HARDWARE_RENDERING="1"
   .\admin-admin.exe
   ```
2. **Increase worker limit** (modify source code if needed)
3. **Use wired network** instead of Wi-Fi
4. **Ensure good network infrastructure** (Gigabit switch)

### Running in Enterprise Environments

- **Antivirus**: May flag as unknown executable (false positive)
  - Add to whitelist if needed
- **Firewall**: Corporate firewalls may block ports 9876/2222
  - Request IT to allow these ports
- **Group Policy**: May restrict unsigned executables
  - Request IT exception or code signing
- **Network Policies**: May block P2P communication
  - Verify local network communication is allowed

---

## Troubleshooting Requirements

### "System Requirements Not Met"

If you see this error (shouldn't happen with v1.0.6):

1. **Check OS version**:
   ```powershell
   # Windows
   Get-ComputerInfo | Select-Object WindowsVersion, OsArchitecture
   ```

2. **Verify 64-bit OS**:
   - admin:admin requires x86-64 architecture
   - 32-bit systems are not supported

3. **Check available RAM**:
   ```powershell
   # Windows
   Get-ComputerInfo | Select-Object CsTotalPhysicalMemory
   ```

### Performance Issues

If app is slow or laggy:

1. **Check CPU usage** (Task Manager / htop)
2. **Close unused SSH tabs**
3. **Reduce number of workers**
4. **Try hardware rendering** (if you have GPU)
5. **Check network latency**

### Out of Memory

If app crashes with OOM errors:

- **Close other applications**
- **Reduce worker connections**
- **Close SSH sessions**
- **Upgrade RAM** (100 MB is absolute minimum)

---

## Version Requirements

### Current Version: v1.0.6-stable

**Minimum Versions for Cross-Compatibility:**
- Admin v1.0.2+ can connect to Worker v1.0.2+
- All versions use same network protocol
- SSH keys compatible across versions

**Recommended:**
- Use same version on all machines
- Always use latest stable (v1.0.6)

---

## Summary

### ✅ YES, It Will Work If You Have:

- 64-bit OS (Windows 7+, Linux 3.2+, macOS 10.13+)
- 1 GHz CPU and 100 MB RAM
- Network adapter
- 30 MB disk space

### ❌ NO, It Won't Work If:

- 32-bit OS
- ARM processor (Raspberry Pi, Apple M1 without Rosetta)
- No display server (headless server)
- Less than 100 MB RAM

### ⚠️ Special Setup Needed:

- WSL requires X server (VcXsrv, X410)
- Corporate networks may need firewall exceptions
- Very old OS may need updates

---

**Last Updated:** February 18, 2026  
**For Version:** admin:admin v1.0.6-stable  
**See Also:** README.md, RELEASE_LOG.md, OPENGL_FIX.md

