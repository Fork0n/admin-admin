package main

import (
	"adminadmin/internal/application"
	"fyne.io/fyne/v2/app"
	"log"
	"os"
)

// Version is set at build time via -ldflags
var Version = "dev"

func main() {
	// Enable software rendering fallback for systems without proper OpenGL drivers
	// This prevents "WGL: the driver does not appear to support OpenGL" errors on:
	// - Virtual machines (VMware, VirtualBox, Hyper-V)
	// - Remote Desktop (RDP, VNC)
	// - Older hardware or outdated drivers
	//
	// To force hardware rendering, set: FYNE_FORCE_HARDWARE_RENDERING=1
	// To force software rendering, set: FYNE_DISABLE_HARDWARE_RENDERING=1
	if os.Getenv("FYNE_FORCE_HARDWARE_RENDERING") == "" {
		// Use software rendering by default for maximum compatibility
		if os.Getenv("FYNE_DISABLE_HARDWARE_RENDERING") == "" {
			os.Setenv("FYNE_DISABLE_HARDWARE_RENDERING", "1")
		}
	} else {
		// User explicitly requested hardware rendering
		os.Setenv("FYNE_DISABLE_HARDWARE_RENDERING", "0")
	}

	// Configure logging
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	log.Println("=====================================")
	log.Printf("        admin:admin v%s", Version)
	log.Println("=====================================")

	fyneApp := app.New()
	log.Println("MAIN: Fyne application created")

	application.NewApp(fyneApp).Run()

	log.Println("=====================================")
	log.Println("  admin:admin Exited")
	log.Println("=====================================")
}
