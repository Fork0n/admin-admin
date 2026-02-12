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
	a.window.Resize(fyne.NewSize(500, 400))
	log.Println("APP: Window created (500x400)")

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

	a.showWorkerWaitingScreen()
}

func (a *App) showAdminConnectScreen() {
	log.Println("APP: Building admin connect screen UI...")
	content := ui.NewAdminConnectScreen(
		func(ip string) { a.connectToWorker(ip) },
		func() { a.backToRoleSelection() },
	)
	a.window.SetContent(content)
	log.Println("APP: Admin connect screen displayed")
}

func (a *App) showAdminDashboard() {
	log.Println("APP: Building admin dashboard UI...")
	content := ui.NewAdminDashboard(
		a.state,
		func() { a.disconnect() },
		func() { a.backToRoleSelection() },
	)
	a.window.SetContent(content)
	log.Println("APP: Admin dashboard displayed")
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
	)
	a.window.SetContent(content)
	log.Println("APP: Worker waiting screen displayed")
}

func (a *App) showWorkerConnectedScreen() {
	log.Println("APP: Building worker connected screen UI...")
	content := ui.NewWorkerConnectedScreen(
		a.state,
		func() { a.backToRoleSelection() },
	)
	a.window.SetContent(content)
	log.Println("APP: Worker connected screen displayed")
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
	log.Println("APP: Returning to connect screen...")
	a.showAdminConnectScreen()
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

	// Create admin client with update callbacks
	log.Println("APP: Creating admin client...")
	a.adminClient = network.NewAdminClient(
		// onUpdate - full device info received
		func(deviceInfo *state.DeviceInfo) {
			log.Println("APP: Received device info update callback")
			log.Printf("APP: Device - Hostname: %s, OS: %s, IP: %s\n",
				deviceInfo.Hostname, deviceInfo.OS, deviceInfo.IPAddress)
			a.state.SetConnectedDevice(deviceInfo)
			log.Println("APP: Showing admin dashboard with device data...")
			a.showAdminDashboard()
		},
		// onMetricsUpdate - real-time metrics
		func(cpuUsage, ramUsage, gpuUsage float64) {
			a.state.UpdateDeviceMetrics(cpuUsage, ramUsage, gpuUsage)
			// Refresh the dashboard to show new metrics
			a.showAdminDashboard()
		},
	)

	// Connect to worker
	log.Printf("APP: Initiating connection to %s:%d...\n", ip, network.DefaultWorkerPort)
	if err := a.adminClient.Connect(ip, network.DefaultWorkerPort); err != nil {
		log.Printf("APP ERROR: Connection failed: %v\n", err)
		dialog.ShowError(err, a.window)
		a.adminClient = nil
		return
	}
	log.Println("APP: Connection initiated successfully")
}
