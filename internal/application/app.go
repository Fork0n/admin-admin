package application

import (
	"adminadmin/internal/network"
	"adminadmin/internal/state"
	"adminadmin/internal/ui"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"log"
)

type App struct {
	fyneApp      fyne.App
	window       fyne.Window
	state        *state.AppState
	adminClient  *network.AdminClient
	workerServer *network.WorkerServer
}

func NewApp(fyneApp fyne.App) *App {
	return &App{
		fyneApp: fyneApp,
		state:   state.NewAppState(),
	}
}

func (a *App) Run() {
	log.Println("=== APPLICATION STARTING ===")
	a.window = a.fyneApp.NewWindow("admin:admin")
	a.window.Resize(fyne.NewSize(800, 600))
	log.Println("APP: Window created (800x600)")

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
	a.window.SetContent(content)
	log.Println("APP: Role selection screen displayed")
}

func (a *App) selectAdminRole() {
	log.Println("=== USER SELECTED: ADMIN ROLE ===")
	a.state.SetRole(state.RoleAdmin)
	a.showAdminDashboard()
}

func (a *App) selectWorkerRole() {
	log.Println("=== USER SELECTED: WORKER ROLE ===")
	a.state.SetRole(state.RoleWorker)

	// Start worker server
	log.Printf("APP: Creating worker server on port %d...\n", network.DefaultWorkerPort)
	a.workerServer = network.NewWorkerServer(network.DefaultWorkerPort)
	if err := a.workerServer.Start(); err != nil {
		log.Printf("APP ERROR: Failed to start worker server: %v\n", err)
		// Show error dialog
		dialog.ShowError(err, a.window)
	} else {
		log.Println("APP: Worker server started successfully")
	}

	a.showWorkerDashboard()
}

func (a *App) showAdminDashboard() {
	log.Println("APP: Building admin dashboard UI...")
	content := ui.NewAdminDashboard(
		a.state,
		func(ip string) { a.connectToWorker(ip) },
		func() { a.disconnect() },
		func() { a.backToRoleSelection() },
		func() { a.refreshAdminDashboard() },
	)
	a.window.SetContent(content)
	log.Println("APP: Admin dashboard displayed")
}

func (a *App) showWorkerDashboard() {
	log.Println("APP: Building worker dashboard UI...")
	content := ui.NewWorkerDashboard(
		func() { a.backToRoleSelection() },
	)
	a.window.SetContent(content)
	log.Println("APP: Worker dashboard displayed")
}

func (a *App) disconnect() {
	log.Println("=== DISCONNECT REQUESTED ===")
	if a.adminClient != nil {
		log.Println("APP: Disconnecting admin client...")
		a.adminClient.Disconnect()
		a.adminClient = nil
		log.Println("APP: Admin client disconnected")
	} else {
		log.Println("APP: No active connection to disconnect")
	}
	a.state.ClearConnection()
	log.Println("APP: Refreshing dashboard after disconnect...")
	a.refreshAdminDashboard()
}

func (a *App) backToRoleSelection() {
	log.Println("=== RETURNING TO ROLE SELECTION ===")

	// Cleanup network resources
	if a.adminClient != nil {
		log.Println("APP: Cleaning up admin client...")
		a.adminClient.Disconnect()
		a.adminClient = nil
	}
	if a.workerServer != nil {
		log.Println("APP: Stopping worker server...")
		a.workerServer.Stop()
		a.workerServer = nil
		log.Println("APP: Worker server stopped")
	}

	a.state.SetRole(state.RoleNone)
	a.state.ClearConnection()
	log.Println("APP: State cleared, showing role selection...")
	a.showRoleSelection()
}

func (a *App) connectToWorker(ip string) {
	log.Printf("=== CONNECTING TO WORKER: %s ===\n", ip)

	// Create admin client with update callback
	log.Println("APP: Creating admin client...")
	a.adminClient = network.NewAdminClient(func(deviceInfo *state.DeviceInfo) {
		log.Println("APP: Received device info update callback")
		log.Printf("APP: Device - Hostname: %s, OS: %s, IP: %s\n",
			deviceInfo.Hostname, deviceInfo.OS, deviceInfo.IPAddress)
		a.state.SetConnectedDevice(deviceInfo)
		log.Println("APP: Refreshing admin dashboard with new data...")
		a.refreshAdminDashboard()
	})

	// Connect to worker
	log.Printf("APP: Initiating connection to %s:%d...\n", ip, network.DefaultWorkerPort)
	if err := a.adminClient.Connect(ip, network.DefaultWorkerPort); err != nil {
		log.Printf("APP ERROR: Connection failed: %v\n", err)
		// Show error dialog to user
		dialog.ShowError(err, a.window)
		a.adminClient = nil
		return
	}
	log.Println("APP: Connection initiated successfully")
}

func (a *App) refreshAdminDashboard() {
	log.Println("APP: Refreshing admin dashboard...")
	// Refresh the admin dashboard with updated state
	a.showAdminDashboard()
	log.Println("APP: Admin dashboard refreshed")
}
