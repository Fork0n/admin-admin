# Firewall Setup for admin:admin

## The Connection Issue

If you're getting "connection timeout" or "connection refused" errors, the most common cause is Windows Firewall blocking port 9876.

## Quick Fix

Run this command **as Administrator** on the **Worker PC**:

```powershell
New-NetFirewallRule -DisplayName "admin:admin Worker" -Direction Inbound -Protocol TCP -LocalPort 9876 -Action Allow
```

## Manual Firewall Configuration

### Windows Firewall

**On Worker PC:**

1. Open Windows Defender Firewall with Advanced Security
   - Press Win + R
   - Type `wf.msc`
   - Press Enter

2. Click "Inbound Rules" in left panel

3. Click "New Rule..." in right panel

4. Select "Port" and click Next

5. Select "TCP" and enter "9876" in Specific local ports

6. Click Next

7. Select "Allow the connection"

8. Click Next

9. Check all profiles (Domain, Private, Public)

10. Click Next

11. Name it "admin:admin Worker Port"

12. Click Finish

### Alternative: Allow Program

1. Open Windows Defender Firewall with Advanced Security

2. Click "Inbound Rules" then "New Rule..."

3. Select "Program" and click Next

4. Browse to: `D:\adminadmin\bin\admin-admin.exe`

5. Select "Allow the connection"

6. Check all profiles

7. Name it "admin:admin"

8. Click Finish

## Verify Port is Open

### On Worker PC

Check if port 9876 is listening:

```powershell
netstat -an | findstr :9876
```

You should see:
```
TCP    0.0.0.0:9876           0.0.0.0:0              LISTENING
```

### On Admin PC

Test connection to worker:

```powershell
Test-NetConnection -ComputerName 192.168.0.67 -Port 9876
```

Should show:
```
TcpTestSucceeded : True
```

## Troubleshooting

### Worker Shows "LISTENING" but Admin Can't Connect

1. **Check firewall on Worker PC** - Most common issue
2. **Check network** - Both PCs must be on same network/subnet
3. **Check IP address** - Use `ipconfig` on Worker to verify
4. **Try localhost first** - Run both on same PC with 127.0.0.1

### Port Already in Use

```powershell
# Find what's using port 9876
netstat -ano | findstr :9876

# Kill the process (replace XXXX with PID from above)
taskkill /F /PID XXXX
```

### Antivirus Blocking

Some antivirus software blocks network connections. Try:
1. Temporarily disable antivirus
2. Add exception for admin:admin executable
3. Add exception for port 9876

## Network Requirements

### Same Network
- Both PCs connected to same router/switch
- Both on same subnet (e.g., 192.168.0.x)
- No VPN interfering

### Different Networks (Advanced)
- Port forwarding required on worker's router
- Forward external port 9876 to worker's local IP:9876
- Use worker's public IP to connect
- Firewall must allow WAN access

## Testing

### Test 1: Same PC
```
Worker: 127.0.0.1
Admin: Connect to 127.0.0.1
```
If this works, network stack is OK.

### Test 2: Local Network
```
Worker: ipconfig (get IP like 192.168.0.67)
Admin: Connect to 192.168.0.67
```
If this fails but Test 1 works, it's firewall.

### Test 3: Firewall Check
```powershell
# Disable Windows Firewall temporarily (not recommended for production)
Set-NetFirewallProfile -Profile Domain,Public,Private -Enabled False

# Try connecting

# Re-enable firewall
Set-NetFirewallProfile -Profile Domain,Public,Private -Enabled True
```

If connection works with firewall off, add the firewall rule above.

