# admin:admin - Quick Start Guide

> **Version:** 1.0.6-stable (February 2026)  
> **Status:** ✅ Production Ready

---

## What is admin:admin?

A lightweight remote administration tool for monitoring and managing Windows/Linux/macOS computers over a local network.

**Key Features:**
- 📊 Real-time CPU, RAM, GPU monitoring
- 🔧 Remote SSH terminal access
- 🌐 Multi-device support
- 🎨 Clean purple-themed UI
- ⚡ Lightweight (24 MB, uses ~60 MB RAM)

---

## Installation (30 seconds)

### 1. Download
Get the latest binary for your OS:
- `admin-admin-1.0.6-stable.exe` (Windows)
- `admin-admin-1.0.6-stable` (Linux/macOS)

### 2. Run
**Windows:** Double-click the `.exe`

**Linux/macOS:**
```bash
chmod +x admin-admin-1.0.6-stable
./admin-admin-1.0.6-stable
```

### 3. Done!
No installation, no dependencies, no configuration needed.

---

## Usage (2 minutes)

### Setup Worker PC (Computer to Monitor)

1. Run the application
2. Click **"Worker PC"**
3. Note the IP address shown (e.g., `192.168.1.50`)
4. Allow firewall access if prompted

**That's it!** The worker is now waiting for connections.

### Setup Admin PC (Computer that Monitors)

1. Run the application
2. Click **"Admin PC"**
3. Enter the Worker's IP address (e.g., `192.168.1.50`)
4. Click **"Connect"**

**Done!** You're now monitoring the worker in real-time.

---

## Firewall (If Connection Fails)

### Worker PC Only:

**Windows:**
```powershell
New-NetFirewallRule -DisplayName "admin:admin" -Direction Inbound -Protocol TCP -LocalPort 9876,2222 -Action Allow
```

**Linux:**
```bash
sudo ufw allow 9876/tcp
sudo ufw allow 2222/tcp
```

**Admin PC:** No firewall changes needed.

---

## SSH Access (Optional)

### Connect via SSH:

1. On Admin dashboard, click **"Open SSH Terminal"**
2. Enter credentials:
   - Username: `admin`
   - Password: `admin`
3. Type commands and press Enter

### Change Credentials:

On the Worker's waiting screen, edit the Username/Password fields before Admin connects.

---

## System Requirements

### ✅ Works On:
- Windows 7, 8, 10, 11
- Linux (any modern distro)
- macOS 10.13+
- Virtual machines
- Remote Desktop
- Computers without GPU

### 📋 Needs:
- 64-bit OS
- 100 MB RAM
- 30 MB disk space
- Network connection

---

## Troubleshooting

### Can't Connect?

1. **Check IP Address:**
   - On Worker PC, run `ipconfig` (Windows) or `ip addr` (Linux)
   - Use the `192.168.x.x` or `10.x.x.x` address

2. **Check Firewall:**
   - Temporarily disable to test
   - Add firewall rule (see above)

3. **Same Network?**
   - Both PCs must be on same Wi-Fi/LAN
   - Won't work over internet without VPN

**More solutions:** [FAQ.md](FAQ.md#-cant-connect-to-worker)

### Application Won't Start?

**Run from terminal to see errors:**
```powershell
.\admin-admin-1.0.6-stable.exe
```

**Common issues:**
- Need 64-bit OS (32-bit not supported)
- Antivirus blocking (add exception)
- Corrupted download (re-download)

**More solutions:** [FAQ.md](FAQ.md#-application-wont-start)

### OpenGL Error?

**This should NOT happen in v1.0.6**, but if it does:

```powershell
$env:FYNE_DISABLE_HARDWARE_RENDERING="1"
.\admin-admin-1.0.6-stable.exe
```

**More solutions:** [FAQ.md](FAQ.md#-opengl-error-rare)

### GPU Shows N/A?

**This is normal** for:
- Integrated graphics (Intel HD)
- Virtual machines
- Older GPUs

**Not a bug** - some hardware doesn't expose GPU usage.

**More info:** [FAQ.md](FAQ.md#-gpu-shows-na-or-0)

---

## Performance

**Typical Usage:**
- CPU: <1% idle, 2-5% monitoring
- RAM: 60 MB (single worker)
- Network: 5 KB/s per worker

**Scales to:**
- Up to 50 workers simultaneously
- Multiple SSH sessions
- Low network overhead

---

## Tips & Tricks

### Multiple Workers
- Connect to many workers from one Admin PC
- Each worker adds ~10 MB RAM
- Disconnect unused workers to free resources

### SSH Sessions
- Open multiple tabs for different tasks
- Close tabs when done (saves ~20 MB each)
- Window auto-closes when all tabs closed

### Performance
- Use hardware rendering for better GPU utilization:
  ```powershell
  $env:FYNE_FORCE_HARDWARE_RENDERING="1"
  .\admin-admin.exe
  ```
- Default software rendering works great for most users

### Security
- Change SSH credentials (default is `admin`/`admin`)
- Use local network only (not internet)
- Consider VPN for remote access

---

## Documentation

**Quick Reference:**
- **README.md** - Complete documentation
- **REQUIREMENTS.md** - Detailed system requirements
- **OPENGL_FIX.md** - Rendering modes and compatibility
- **FIREWALL.md** - Firewall configuration
- **BUILD.md** - Building from source

**For Issues:**
- Check RELEASE_LOG.md (gitignored, for developers)
- Report bugs with full error output

---

## Version Info

**Current:** v1.0.6-stable (February 18, 2026)

**What's New:**
- ✅ Enhanced OpenGL compatibility
- ✅ User-selectable rendering mode
- ✅ Improved documentation
- ✅ Bug fixes

**Upgrade from v1.0.2/v1.0.4:**
- Just replace the executable
- No configuration changes needed
- Fully backward compatible

---

## Common Questions

### Q: Is it free?
**A:** Yes, completely free.

### Q: Does it work over internet?
**A:** Designed for local networks. Works over internet with VPN or port forwarding (not recommended for security).

### Q: Can I monitor Windows from macOS?
**A:** Yes! Admin and Worker can be different OSes.

### Q: How many workers can I monitor?
**A:** Tested up to 50. Practical limit depends on your hardware.

### Q: Does it need admin/root privileges?
**A:** No, runs as normal user. May need admin for firewall rules.

### Q: Is my data secure?
**A:** SSH uses encryption. Main protocol sends data in plain text over local network. Don't use over public internet.

### Q: Can I build from source?
**A:** Yes! See BUILD.md for instructions.

---

## Support

**Before asking for help:**
1. Read the troubleshooting section above
2. Check README.md and REQUIREMENTS.md
3. Run from terminal to see error messages

**When reporting issues, include:**
- Version (v1.0.6-stable)
- OS and version
- Full error output from terminal
- Steps to reproduce

---

## License & Credits

**Built with:**
- Go 1.21+
- Fyne GUI framework
- gopsutil for system monitoring
- golang.org/x/crypto for SSH

**Platforms:**
- Windows 7+
- Linux (kernel 3.2+)
- macOS 10.13+

---

**That's it!** You're ready to use admin:admin. For complete documentation, see [README.md](README.md).

---

**Last Updated:** February 18, 2026  
**Current Version:** v1.0.6-stable  
**Status:** Production Ready

