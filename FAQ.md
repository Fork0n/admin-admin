# admin:admin FAQ & Troubleshooting

Quick answers to common questions and simple problem-solving.

**For unique/complex issues:** [GitHub Issues](https://github.com/yourusername/adminadmin/issues)

---

## Frequently Asked Questions

### General

**Q: Is admin:admin free?**  
A: Yes, completely free and open source.

**Q: Does it work over the internet?**  
A: Designed for local networks. Internet use requires VPN or port forwarding (not recommended for security reasons).

**Q: Can I monitor Windows from macOS/Linux?**  
A: Yes! Admin and Worker can run on different operating systems.

**Q: How many workers can I monitor?**  
A: Tested up to 50. Practical limit depends on your hardware.

**Q: Does it need admin/root privileges?**  
A: No, runs as normal user. May need admin for firewall rules only.

**Q: Is my data secure?**  
A: SSH uses encryption. Main protocol sends data in plain text over local network. Don't use over public internet without VPN.

**Q: Can I build from source?**  
A: Yes! See [BUILD.md](BUILD.md) for instructions.

---

### Installation & Setup

**Q: Do I need to install anything?**  
A: No installation needed. Just run the executable.

**Q: Why does Windows Defender flag it?**  
A: False positive for unsigned executables. Add to exclusions or build from source.

**Q: Which version should I use?**  
A: Always use the latest stable (v1.0.6).

**Q: How do I update?**  
A: Download new version and replace the old executable. That's it.

---

## Common Problems & Quick Fixes

### 🔴 Can't Connect to Worker

**Symptom:** "Connection failed" or timeout when connecting

**Quick Fixes:**
1. **Check firewall on Worker PC**
   ```powershell
   # Windows - run as Administrator
   New-NetFirewallRule -DisplayName "admin:admin" -Direction Inbound -Protocol TCP -LocalPort 9876,2222 -Action Allow
   ```
   
   ```bash
   # Linux
   sudo ufw allow 9876/tcp
   sudo ufw allow 2222/tcp
   ```

2. **Verify both PCs on same network**
   - Both must be on same Wi-Fi/LAN
   - Check IP addresses start with same prefix (e.g., both 192.168.1.x)

3. **Use correct IP address**
   - On Worker PC: Run `ipconfig` (Windows) or `ip addr` (Linux)
   - Use the `192.168.x.x` or `10.x.x.x` address
   - NOT `127.0.0.1` (unless testing on same PC)

4. **Test network connection**
   ```powershell
   # From Admin PC
   ping 192.168.x.x
   Test-NetConnection -ComputerName 192.168.x.x -Port 9876
   ```

**Still not working?** → [GitHub Issues](https://github.com/yourusername/adminadmin/issues)

---

### 🔴 SSH Connection Fails

**Symptom:** "SSH connection failed" or authentication error

**Quick Fixes:**
1. **Check credentials**
   - Default: username `admin`, password `admin`
   - Verify custom credentials if changed on Worker

2. **Check firewall (port 2222)**
   ```powershell
   # Windows
   New-NetFirewallRule -DisplayName "admin:admin SSH" -Direction Inbound -Protocol TCP -LocalPort 2222 -Action Allow
   ```

3. **Test SSH port**
   ```powershell
   Test-NetConnection -ComputerName 192.168.x.x -Port 2222
   ```

4. **Regenerate SSH keys**
   - Windows: Delete `%APPDATA%\adminadmin\ssh_host_key`
   - Linux: Delete `~/.config/adminadmin/ssh_host_key`
   - Restart Worker to auto-generate new keys

**Still not working?** → [GitHub Issues](https://github.com/yourusername/adminadmin/issues)

---

### 🟡 OpenGL Error (Rare)

**Symptom:** "WGL: the driver does not appear to support OpenGL"

**This should NOT happen in v1.0.6**, but if it does:

**Quick Fix:**
```powershell
# Force software rendering
$env:FYNE_DISABLE_HARDWARE_RENDERING="1"
Remove-Item Env:FYNE_FORCE_HARDWARE_RENDERING -ErrorAction SilentlyContinue
.\admin-admin-1.0.6-stable.exe
```

**If still failing:** → [GitHub Issues](https://github.com/yourusername/adminadmin/issues) with:
- OS version
- GPU model and driver version
- Full error output

---

### 🟢 GPU Shows "N/A" or 0%

**Symptom:** GPU monitoring displays N/A

**This is NORMAL for:**
- Integrated graphics (Intel HD Graphics)
- Virtual machines
- Older GPU models
- Some laptop GPUs

**Why?**
- Some hardware doesn't expose GPU usage data
- Requires vendor-specific drivers

**Solution:**
- Not a bug, just hardware limitation
- Use external tools (GPU-Z, MSI Afterburner) if you need GPU monitoring

---

### 🔴 Application Won't Start

**Symptom:** Double-clicking does nothing or immediate crash

**Quick Fixes:**
1. **Run from terminal to see errors**
   ```powershell
   # Windows
   .\admin-admin-1.0.6-stable.exe
   
   # Linux
   ./admin-admin-1.0.6-stable
   ```

2. **Check requirements**
   - 64-bit OS (32-bit not supported)
   - 100 MB RAM available
   - Windows 7+, Linux 3.2+, macOS 10.13+

3. **Verify not corrupted**
   ```powershell
   # Windows - check file hash
   Get-FileHash .\admin-admin-1.0.6-stable.exe
   ```

4. **Try as administrator** (Windows only)
   - Right-click → "Run as administrator"

**Still not working?** → [GitHub Issues](https://github.com/yourusername/adminadmin/issues) with full error output

---

### 🟡 High CPU or RAM Usage

**Symptom:** Application using too much resources

**Expected Usage:**
- CPU: <1% idle, 2-5% monitoring, 5-10% with SSH
- RAM: 60 MB + 10 MB per worker + 20 MB per SSH tab

**If higher:**
1. **Close unused SSH sessions**
   - Each tab uses ~20 MB RAM
   - Close tabs when done

2. **Disconnect unused workers**
   - Each worker adds ~10 MB RAM

3. **Check metrics update rate**
   - Should be 1 Hz (once per second)

4. **Try hardware rendering** (if you have good GPU)
   ```powershell
   $env:FYNE_FORCE_HARDWARE_RENDERING="1"
   .\admin-admin-1.0.6-stable.exe
   ```

---

### 🟡 Application Freezes or Lags

**Symptom:** UI becomes unresponsive or slow

**Quick Fixes:**
1. **Check number of connections**
   - Recommended max: 25 workers
   - Hard limit: ~50 workers

2. **Check network quality**
   - Poor Wi-Fi can cause lag
   - Use wired connection if possible

3. **Close background programs**
   - Free up system resources

4. **Restart application**
   - Sometimes fixes temporary glitches

**Persistent freezing?** → [GitHub Issues](https://github.com/yourusername/adminadmin/issues)

---

### 🔴 Wrong IP Address Shown on Worker

**Symptom:** Worker shows 169.254.x.x instead of 192.168.x.x

**Why?**
- `169.254.x.x` is Windows self-assigned IP (no DHCP)
- Means network not properly configured

**Quick Fixes:**
1. **Check network connection**
   - Ensure connected to Wi-Fi/LAN
   - Run `ipconfig /all` (Windows) to see all IPs

2. **Use the correct IP**
   - Look for `192.168.x.x` or `10.x.x.x` in ipconfig output
   - Ignore `169.254.x.x` address

3. **Restart network adapter**
   ```powershell
   # Windows
   ipconfig /release
   ipconfig /renew
   ```

---

### 🟢 SSH Terminal Doesn't Support vim/nano

**Symptom:** Interactive editors don't work properly

**Why?**
- SSH terminal doesn't support full TUI (Text User Interface) applications
- Designed for command execution, not interactive editors

**Workaround:**
- Use `notepad` (Windows) or `cat`/`echo` for simple edits
- For complex editing, use file sharing or local editors

**Not a bug:** Limitation of the simplified SSH terminal

---

### 🟡 Can't Copy Text from SSH Terminal

**Symptom:** Copy/paste not working

**Solution:**
- Select text with mouse
- Right-click → Copy (or Ctrl+C in some terminals)
- Terminal supports text selection and copy

---

## Performance Optimization

### Get Better Performance

1. **Use hardware rendering** (if you have good GPU)
   ```powershell
   $env:FYNE_FORCE_HARDWARE_RENDERING="1"
   .\admin-admin-1.0.6-stable.exe
   ```

2. **Use wired network** instead of Wi-Fi

3. **Close unused SSH sessions**

4. **Limit worker connections** to what you need

### Get Maximum Compatibility

1. **Use software rendering** (default)
   - Already enabled by default
   - Works on all systems

2. **Keep workers updated** to v1.0.6

3. **Allow firewall exceptions** on Worker PC

---

## Error Messages Explained

| Error Message | Meaning | Fix |
|---------------|---------|-----|
| "Connection failed: dial tcp timeout" | Can't reach Worker | Check firewall, verify IP, same network |
| "SSH connection failed" | Can't connect to SSH | Check port 2222, verify credentials |
| "Failed to connect to worker" | Network issue | Check Worker is running, firewall allowed |
| "WGL: driver does not support OpenGL" | Graphics driver issue | Force software rendering (see above) |
| "Address already in use" | Port 9876/2222 busy | Close other instance or restart PC |

---

## Platform-Specific Issues

### Windows

**Antivirus flags executable:**
- Add to exclusions
- Or build from source and sign

**Firewall popup:**
- Click "Allow access" when prompted
- Or manually add rule (see connection issues above)

### Linux

**Permission denied:**
```bash
chmod +x admin-admin-1.0.6-stable
```

**Firewall (Ubuntu/Debian):**
```bash
sudo ufw allow 9876/tcp
sudo ufw allow 2222/tcp
sudo ufw reload
```

### macOS

**"admin-admin cannot be opened because it is from an unidentified developer":**
```bash
# Remove quarantine flag
xattr -d com.apple.quarantine admin-admin-1.0.6-stable
```

---

## Still Having Issues?

### Before Opening GitHub Issue:

1. ✅ Read this FAQ
2. ✅ Check [QUICKSTART.md](QUICKSTART.md)
3. ✅ Verify you're using v1.0.6-stable
4. ✅ Run from terminal and capture error output

### Open GitHub Issue With:

- **Version:** `admin:admin v1.0.6-stable`
- **OS:** Windows 10 21H2, Ubuntu 22.04, etc.
- **Hardware:** CPU, RAM, GPU models
- **Full error output** from terminal
- **Steps to reproduce** the issue
- **Expected vs actual** behavior

**GitHub Issues:** [github.com/yourusername/adminadmin/issues](https://github.com/yourusername/adminadmin/issues)

---

## Additional Resources

- [QUICKSTART.md](QUICKSTART.md) - 2-minute setup guide
- [README.md](README.md) - Complete documentation
- [REQUIREMENTS.md](REQUIREMENTS.md) - System requirements
- [OPENGL_FIX.md](OPENGL_FIX.md) - Rendering modes
- [BUILD.md](BUILD.md) - Build from source

---

**Last Updated:** February 18, 2026  
**For Version:** admin:admin v1.0.6-stable

