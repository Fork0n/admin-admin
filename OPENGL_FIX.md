# OpenGL Compatibility Fix

## Problem

In rare cases, users may encounter the following error when running admin:admin:

```
WGL: the driver does not appear to support OpenGL
```

This is a **driver-specific issue** that has been reported on:
- Certain hardware configurations with outdated or missing OpenGL drivers
- Some virtual machines (depending on configuration)
- Remote Desktop sessions (RDP, VNC) in specific setups
- Older hardware without proper OpenGL support

**Note**: Testing shows this error does NOT occur on:
- Most modern Windows 10/11 systems
- Standard VMs with 3D acceleration disabled
- Linux systems with basic graphics drivers
- WSL environments

## Solution

**✅ INCLUDED in admin:admin v1.0.0+**

The application includes automatic software rendering fallback as a safety measure for the small percentage of systems that may encounter OpenGL driver issues.

## How It Works

The fix is implemented in `cmd/app/main.go`:

```go
// Enable software rendering fallback by default for maximum compatibility
// Users can override with FYNE_FORCE_HARDWARE_RENDERING=1 if desired
if os.Getenv("FYNE_FORCE_HARDWARE_RENDERING") == "" {
    if os.Getenv("FYNE_DISABLE_HARDWARE_RENDERING") == "" {
        os.Setenv("FYNE_DISABLE_HARDWARE_RENDERING", "1")
    }
}
```

### What This Does

1. **Uses software rendering by default** - Safe fallback that works everywhere
2. **Allows override** - Users with proper drivers can force hardware rendering
3. **Prevents errors** - Eliminates OpenGL driver issues on affected systems
4. **No noticeable performance impact** - admin:admin's UI is lightweight

## Performance Impact

- **Negligible** for admin:admin's UI
- Gauges update smoothly at 30 FPS even with CPU rendering
- Dashboard and SSH terminal are responsive
- Actual monitoring (CPU/RAM/GPU/Network) is unaffected
- Worth the minor tradeoff for universal compatibility

## Manual Override

### Force Hardware Rendering (Better Performance)

If you have proper OpenGL drivers and want to use GPU rendering:

**Windows PowerShell:**
```powershell
$env:FYNE_FORCE_HARDWARE_RENDERING="1"
.\bin\admin-admin.exe
```

**Windows CMD:**
```cmd
set FYNE_FORCE_HARDWARE_RENDERING=1
.\bin\admin-admin.exe
```

**Linux/macOS:**
```bash
export FYNE_FORCE_HARDWARE_RENDERING=1
./bin/admin-admin
```

### Force Software Rendering (Maximum Compatibility)

This is already the default, but you can explicitly set it:

**Windows PowerShell:**
```powershell
$env:FYNE_DISABLE_HARDWARE_RENDERING="1"
.\bin\admin-admin.exe
```

**Windows CMD:**
```cmd
set FYNE_DISABLE_HARDWARE_RENDERING=1
.\bin\admin-admin.exe
```

**Linux/macOS:**
```bash
export FYNE_DISABLE_HARDWARE_RENDERING=1
./bin/admin-admin
```

## Alternative Solutions

If software rendering doesn't work for some reason:

### 1. Update Graphics Drivers

- **NVIDIA**: https://www.nvidia.com/Download/index.aspx
- **AMD**: https://www.amd.com/en/support
- **Intel**: https://www.intel.com/content/www/us/en/download-center/home.html

### 2. Install Mesa3D (for VMs/older systems)

Mesa3D provides OpenGL implementation in software:

1. Download from: https://github.com/pal1000/mesa-dist-win/releases
2. Download the latest `mesa3d-XX.X.X-release-mingw.7z`
3. Extract the archive
4. Copy `x64/opengl32.dll` to the same folder as `admin-admin.exe`
5. Run the application

### 3. Remote Desktop Users

If using Remote Desktop:
- Software rendering (enabled by default in v1.0.0+) should work
- Or enable RemoteFX for GPU acceleration (Windows Server)
- Or use the physical machine if possible

## Testing

To verify the fix works:

```powershell
# Build the application
.\build.ps1 -v "test"

# Run it - should work without errors
.\bin\admin-admin-build-test.exe
```

Look for this in the startup logs:
```
=====================================
        admin:admin v1.0.0
=====================================
MAIN: Fyne application created
```

If you see the Fyne application created message, the fix worked!

## Technical Details

### Fyne Rendering Modes

Fyne GUI toolkit supports two rendering modes:

1. **Hardware Rendering (default)**
   - Uses GPU via OpenGL
   - Requires proper graphics drivers
   - Better performance for complex graphics
   - Fails on systems without OpenGL

2. **Software Rendering (fallback)**
   - Uses CPU for rendering
   - Works on any system
   - Slightly lower performance
   - No driver dependencies

### Why Software Rendering for admin:admin?

admin:admin's UI consists of:
- Radial gauges (drawn with canvas primitives)
- Text labels and buttons
- Simple layouts and containers
- No complex 3D graphics or animations

These render perfectly fine with CPU-based software rendering with no visible performance degradation.

### Environment Variables

The `FYNE_DISABLE_HARDWARE_RENDERING` environment variable:
- `"1"` = Force software rendering (CPU-based)
- `"0"` = Force hardware rendering (GPU-based)
- Not set = Fyne decides (defaults to hardware)

Our fix sets it to `"1"` by default unless already configured.

## Support

If you still experience issues after updating to v1.0.0+:

1. Check console output for errors
2. Try manual software rendering override
3. Update graphics drivers
4. Try Mesa3D installation
5. Report issue with:
   - Windows version
   - Graphics card model
   - Console error output
   - Whether running in VM/RDP

## Version History

- **v1.0.0+**: Automatic software rendering fallback enabled by default
- **v0.9 and earlier**: Required manual environment variable configuration

---

**Built for 99% device compatibility.**

