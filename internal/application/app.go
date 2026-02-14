package application

import (
	"adminadmin/internal/network"
	"adminadmin/internal/state"
	"adminadmin/internal/ui"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"log"
	"sync"
)

type App struct {
	fyneApp      fyne.App
	window       fyne.Window
	state        *state.AppState
	adminClients map[string]*network.AdminClient // Multiple connections by IP
	clientsMu    sync.RWMutex
	workerServer *network.WorkerServer
	sshServer    *network.SSHServer
	sshPassword  string

	// Dashboard controller (persistent for smooth gauge animations)
	dashboardCtrl *ui.AdminDashboardController

	// SSH terminal window (separate window with tabs)
	sshTerminalWindow *ui.SSHTerminalWindow
}

func NewApp(fyneApp fyne.App) *App {
	return &App{
		fyneApp:      fyneApp,
		state:        state.NewAppState(),
		adminClients: make(map[string]*network.AdminClient),
		sshPassword:  network.DefaultSSHPassword, // Default SSH password: admin
	}
}

// runOnMain safely runs a function on the main UI thread
func (a *App) runOnMain(fn func()) {
	if drv := fyne.CurrentApp().Driver(); drv != nil {
		drv.DoFromGoroutine(fn, false)
	} else {
		fn() // Fallback to direct execution
	}
}

func (a *App) Run() {
	log.Println("=== APPLICATION STARTING ===")

	// Apply purple theme
	a.fyneApp.Settings().SetTheme(ui.NewPurpleTheme())

	a.window = a.fyneApp.NewWindow("admin:admin")
	a.window.Resize(fyne.NewSize(900, 600))
	log.Println("APP: Window created (900x600)")

	a.showRoleSelection()

	log.Println("APP: Showing window and entering main loop...")
	a.window.ShowAndRun()
	log.Println("=== APPLICATION SHUTTING DOWN ===")
}

func (a *App) showRoleSelection() {
	log.Println("APP: Showing role selection screen")
	content := ui.NewRoleSelectScreen(
		func() { a.selectAdminRole() },
		func() { a.selectWorkerRole() },
	)
	a.runOnMain(func() {
		a.window.SetContent(content)
	})
	log.Println("APP: Role selection screen displayed")
}

func (a *App) selectAdminRole() {
	log.Println("=== USER SELECTED: ADMIN ROLE ===")
	a.state.SetRole(state.RoleAdmin)
	a.showAdminConnectScreen()
}

func (a *App) selectWorkerRole() {
	log.Println("=== USER SELECTED: WORKER ROLE ===")
	a.state.SetRole(state.RoleWorker)

	// Start worker server
	log.Printf("APP: Creating worker server on port %d...\n", network.DefaultWorkerPort)
	a.workerServer = network.NewWorkerServer(network.DefaultWorkerPort)

	// Set callbacks for admin connection events
	a.workerServer.SetCallbacks(
		func(hostname string) {
			log.Printf("APP: Admin connected: %s\n", hostname)
			a.state.SetConnectedAdmin(&state.AdminInfo{Hostname: hostname})
			a.showWorkerConnectedScreen()
		},
		func() {
			log.Println("APP: Admin disconnected")
			a.state.ClearConnection()
			a.showWorkerWaitingScreen()
		},
	)

	if err := a.workerServer.Start(); err != nil {
		log.Printf("APP ERROR: Failed to start worker server: %v\n", err)
		dialog.ShowError(err, a.window)
	} else {
		log.Println("APP: Worker server started successfully")
	}

	// Start SSH server
	a.sshServer = network.NewSSHServer(network.DefaultSSHPort)
	if err := a.sshServer.Start(a.sshPassword); err != nil {
		log.Printf("APP WARNING: Failed to start SSH server: %v\n", err)
	} else {
		log.Printf("APP: SSH server started on port %d\n", network.DefaultSSHPort)
	}

	a.showWorkerWaitingScreen()
}

func (a *App) showAdminConnectScreen() {
	log.Println("APP: Building admin connect screen UI...")
	content := ui.NewAdminConnectScreen(
		func(ip string) { a.connectToWorker(ip) },
		func() { a.backToRoleSelection() },
	)
	a.runOnMain(func() {
		a.window.SetContent(content)
	})
	log.Println("APP: Admin connect screen displayed")
}

func (a *App) showAdminDashboard() {
	log.Println("APP: Building admin dashboard UI...")

	// Create controller if not exists
	if a.dashboardCtrl == nil {
		a.dashboardCtrl = ui.NewAdminDashboardController(
			a.state,
			func() { a.disconnectAll() },
			func() { a.backToRoleSelection() },
			func() { a.showAddWorkerDialog() }, // Add worker dialog
			func(id string) { a.selectWorker(id) },
			func(ip string) { a.showSSHDialog(ip) },
		)
	}

	content := a.dashboardCtrl.GetContent()
	a.runOnMain(func() {
		a.window.SetContent(content)
	})
	log.Println("APP: Admin dashboard displayed")
}

// showAddWorkerDialog shows a dialog to add a new worker without leaving the dashboard
func (a *App) showAddWorkerDialog() {
	log.Println("APP: Showing add worker dialog")

	ipEntry := widget.NewEntry()
	ipEntry.SetPlaceHolder("Enter Worker IP (e.g., 192.168.1.100)")

	formItems := []*widget.FormItem{
		widget.NewFormItem("Worker IP", ipEntry),
	}

	dialog.ShowForm(
		"Add Worker",
		"Connect",
		"Cancel",
		formItems,
		func(ok bool) {
			if ok && ipEntry.Text != "" {
				a.connectToWorker(ipEntry.Text)
			}
			// Cancel just closes the dialog, doesn't affect existing connections
		},
		a.window,
	)
}

// updateDashboardMetrics updates only the gauge values without rebuilding UI
func (a *App) updateDashboardMetrics() {
	if a.dashboardCtrl != nil {
		a.dashboardCtrl.UpdateMetricsOnly()
	}
}

func (a *App) showWorkerWaitingScreen() {
	log.Println("APP: Building worker waiting screen UI...")
	localIP := ""
	if a.workerServer != nil {
		localIP = a.workerServer.GetLocalIP()
	}
	content := ui.NewWorkerWaitingScreen(
		localIP,
		network.DefaultWorkerPort,
		func() { a.backToRoleSelection() },
		func(username, password string) {
			// Update SSH credentials when user changes them
			if a.sshServer != nil {
				a.sshServer.SetCredentials(username, password)
				log.Printf("APP: SSH credentials updated - username: %s\n", username)
			}
		},
	)
	a.runOnMain(func() {
		a.window.SetContent(content)
	})
	log.Println("APP: Worker waiting screen displayed")
}

func (a *App) showWorkerConnectedScreen() {
	log.Println("APP: Building worker connected screen UI...")
	content := ui.NewWorkerConnectedScreen(
		a.state,
		func() { a.backToRoleSelection() },
	)
	a.runOnMain(func() {
		a.window.SetContent(content)
		// Make window compact when connected
		a.window.Resize(fyne.NewSize(350, 120))
	})
	log.Println("APP: Worker connected screen displayed")
}

func (a *App) selectWorker(id string) {
	log.Printf("APP: Selecting worker: %s\n", id)
	a.state.SetSelectedWorker(id)
	a.showAdminDashboard()
}

func (a *App) disconnectAll() {
	log.Println("=== DISCONNECT ALL REQUESTED ===")
	a.clientsMu.Lock()
	for ip, client := range a.adminClients {
		log.Printf("APP: Disconnecting from %s...\n", ip)
		client.Disconnect()
		delete(a.adminClients, ip)
	}
	a.clientsMu.Unlock()
	a.state.ClearConnection()
	log.Println("APP: All connections closed")
	a.showAdminConnectScreen()
}

func (a *App) backToRoleSelection() {
	log.Println("=== RETURNING TO ROLE SELECTION ===")

	// Cleanup dashboard controller
	a.dashboardCtrl = nil

	// Cleanup SSH terminal window
	if a.sshTerminalWindow != nil {
		a.sshTerminalWindow.Close()
		a.sshTerminalWindow = nil
	}

	// Cleanup admin clients
	a.clientsMu.Lock()
	for ip, client := range a.adminClients {
		log.Printf("APP: Cleaning up connection to %s...\n", ip)
		client.Disconnect()
		delete(a.adminClients, ip)
	}
	a.clientsMu.Unlock()

	// Cleanup worker server
	if a.workerServer != nil {
		log.Println("APP: Stopping worker server...")
		a.workerServer.Stop()
		a.workerServer = nil
		log.Println("APP: Worker server stopped")
	}

	// Cleanup SSH server
	if a.sshServer != nil {
		log.Println("APP: Stopping SSH server...")
		a.sshServer.Stop()
		a.sshServer = nil
		log.Println("APP: SSH server stopped")
	}

	// Restore window size
	a.runOnMain(func() {
		a.window.Resize(fyne.NewSize(900, 600))
	})

	a.state.SetRole(state.RoleNone)
	a.state.ClearConnection()
	log.Println("APP: State cleared, showing role selection...")
	a.showRoleSelection()
}

func (a *App) connectToWorker(ip string) {
	log.Printf("=== CONNECTING TO WORKER: %s ===\n", ip)

	// Check if already connected
	a.clientsMu.RLock()
	if _, exists := a.adminClients[ip]; exists {
		a.clientsMu.RUnlock()
		log.Printf("APP: Already connected to %s\n", ip)
		dialog.ShowInformation("Already Connected", fmt.Sprintf("Already connected to %s", ip), a.window)
		return
	}
	a.clientsMu.RUnlock()

	// Create admin client with update callbacks
	log.Println("APP: Creating admin client...")
	client := network.NewAdminClient(
		// onUpdate - full device info received
		func(deviceInfo *state.DeviceInfo) {
			log.Println("APP: Received device info update callback")
			log.Printf("APP: Device - Hostname: %s, OS: %s, IP: %s\n",
				deviceInfo.Hostname, deviceInfo.OS, deviceInfo.IPAddress)
			deviceInfo.ID = ip // Use IP as ID
			a.state.AddConnectedDevice(deviceInfo)
			// Force rebuild since we have a new worker
			if a.dashboardCtrl != nil {
				a.dashboardCtrl.ForceRebuild()
			}
			log.Println("APP: Showing admin dashboard with device data...")
			a.showAdminDashboard()
		},
		// onMetricsUpdate - real-time metrics (just update values, don't rebuild)
		func(cpuUsage, ramUsage, gpuUsage float64) {
			a.state.UpdateDeviceMetricsByID(ip, cpuUsage, ramUsage, gpuUsage)
			// Only update gauges if this is the selected worker
			if a.state.GetSelectedWorkerID() == ip {
				a.updateDashboardMetrics()
			}
		},
	)

	// Connect to worker
	log.Printf("APP: Initiating connection to %s:%d...\n", ip, network.DefaultWorkerPort)
	if err := client.Connect(ip, network.DefaultWorkerPort); err != nil {
		log.Printf("APP ERROR: Connection failed: %v\n", err)
		dialog.ShowError(err, a.window)
		return
	}

	// Store the client
	a.clientsMu.Lock()
	a.adminClients[ip] = client
	a.clientsMu.Unlock()

	log.Println("APP: Connection initiated successfully")
}

func (a *App) showSSHDialog(workerIP string) {
	log.Printf("APP: Showing SSH dialog for %s\n", workerIP)

	// Get hostname for the worker
	hostname := workerIP
	device := a.state.GetConnectedDeviceByID(workerIP)
	if device != nil {
		hostname = device.Hostname
	}

	userEntry := widget.NewEntry()
	userEntry.SetPlaceHolder("Username")
	userEntry.SetText(network.DefaultSSHUsername) // Default: admin

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Password")
	passwordEntry.SetText(network.DefaultSSHPassword) // Default: admin

	formItems := []*widget.FormItem{
		widget.NewFormItem("Username", userEntry),
		widget.NewFormItem("Password", passwordEntry),
	}

	dialog.ShowForm(
		fmt.Sprintf("SSH to %s (%s)", hostname, workerIP),
		"Connect",
		"Cancel",
		formItems,
		func(ok bool) {
			if ok {
				a.connectSSH(workerIP, hostname, userEntry.Text, passwordEntry.Text)
			}
		},
		a.window,
	)
}

func (a *App) connectSSH(ip, hostname, user, password string) {
	log.Printf("APP: Connecting SSH to %s as %s\n", ip, user)

	sshClient := network.NewSSHClient()
	err := sshClient.Connect(ip, network.DefaultSSHPort, user, password)
	if err != nil {
		dialog.ShowError(fmt.Errorf("SSH connection failed: %v", err), a.window)
		return
	}

	// Create SSH terminal window if not exists
	if a.sshTerminalWindow == nil {
		a.sshTerminalWindow = ui.NewSSHTerminalWindow(a.fyneApp, func() {
			// Cleanup when window closes
			a.sshTerminalWindow = nil
		})
	}

	// Add tab for this connection
	tabID := ip
	a.sshTerminalWindow.AddTab(tabID, hostname, ip, func(cmd string) string {
		output, err := sshClient.ExecuteCommand(cmd)
		if err != nil {
			return fmt.Sprintf("Error: %v\n%s", err, output)
		}
		return output
	})

	// Show the window
	a.sshTerminalWindow.Show()
}
