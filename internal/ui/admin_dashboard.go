package ui

import (
	"adminadmin/internal/state"
	"adminadmin/internal/system"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"sync"
)

// AdminConnectScreen shows the connection screen before connecting
func NewAdminConnectScreen(onConnect func(string), onBack func()) fyne.CanvasObject {
	title := widget.NewLabelWithStyle(
		"admin:admin",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	subtitle := widget.NewLabelWithStyle(
		"Connect to Worker PC",
		fyne.TextAlignCenter,
		fyne.TextStyle{},
	)

	// IP input - wider entry
	ipEntry := widget.NewEntry()
	ipEntry.SetPlaceHolder("Enter Worker IP (e.g., 192.168.1.100)")

	connectButton := widget.NewButton("Connect", func() {
		if ipEntry.Text != "" {
			onConnect(ipEntry.Text)
		}
	})
	connectButton.Importance = widget.HighImportance

	backButton := widget.NewButton("Back", onBack)

	// Fixed layout: title, subtitle, separator, label, ip input, connect, separator, back
	content := container.NewVBox(
		title,
		subtitle,
		widget.NewSeparator(),
		widget.NewLabel("Worker IP Address:"),
		container.NewGridWrap(fyne.NewSize(300, 40), ipEntry),
		connectButton,
		widget.NewSeparator(),
		backButton,
	)

	return container.NewCenter(content)
}

// AdminDashboardController manages the admin dashboard with persistent gauges
type AdminDashboardController struct {
	mu sync.RWMutex

	// Callbacks
	onDisconnect   func()
	onBack         func()
	onAddWorker    func()
	onSelectWorker func(string)
	onSSH          func(string)

	// State
	appState *state.AppState

	// Cached UI elements (reused across updates)
	cpuGauge *Gauge
	ramGauge *Gauge
	gpuGauge *Gauge

	// Labels that need updating
	ramDetailsLabel *widget.Label
	uptimeLabel     *widget.Label

	// Current worker ID being displayed
	currentWorkerID string

	// Root container
	rootContainer fyne.CanvasObject
}

// NewAdminDashboardController creates a new dashboard controller
func NewAdminDashboardController(
	appState *state.AppState,
	onDisconnect func(),
	onBack func(),
	onAddWorker func(),
	onSelectWorker func(string),
	onSSH func(string),
) *AdminDashboardController {
	ctrl := &AdminDashboardController{
		appState:       appState,
		onDisconnect:   onDisconnect,
		onBack:         onBack,
		onAddWorker:    onAddWorker,
		onSelectWorker: onSelectWorker,
		onSSH:          onSSH,
	}

	// Create persistent gauges
	ctrl.cpuGauge = NewGauge("CPU")
	ctrl.ramGauge = NewGauge("RAM")
	ctrl.gpuGauge = NewGauge("GPU")

	// Create persistent labels
	ctrl.ramDetailsLabel = widget.NewLabel("")
	ctrl.uptimeLabel = widget.NewLabel("")

	return ctrl
}

// runOnMain safely runs a function on the main UI thread
func (ctrl *AdminDashboardController) runOnMain(fn func()) {
	if drv := fyne.CurrentApp().Driver(); drv != nil {
		drv.DoFromGoroutine(fn, false)
	} else {
		fn()
	}
}

// GetContent returns the dashboard UI, creating or updating as needed
func (ctrl *AdminDashboardController) GetContent() fyne.CanvasObject {
	ctrl.mu.Lock()
	defer ctrl.mu.Unlock()

	workers := ctrl.appState.GetConnectedDevicesList()

	if len(workers) == 0 {
		return container.NewCenter(widget.NewLabel("No workers connected"))
	}

	selectedID := ctrl.appState.GetSelectedWorkerID()
	device := ctrl.appState.GetSelectedWorker()

	// Update gauge values (they will animate smoothly from current to new)
	if device != nil {
		ctrl.cpuGauge.SetValue(device.CPUUsage)
		ctrl.ramGauge.SetValue(device.RAMUsage)
		ctrl.gpuGauge.SetValue(device.GPUUsage)

		ctrl.ramDetailsLabel.SetText(fmt.Sprintf("RAM: %s / %s",
			system.FormatBytes(device.RAMUsed),
			system.FormatBytes(device.RAMTotal)))
		ctrl.uptimeLabel.SetText(fmt.Sprintf("Uptime: %s", system.FormatUptime(device.Uptime)))
	}

	// Only rebuild UI if worker selection changed or first time
	if ctrl.rootContainer == nil || ctrl.currentWorkerID != selectedID {
		ctrl.currentWorkerID = selectedID
		ctrl.rootContainer = ctrl.buildFullUI(workers, device, selectedID)
	}

	return ctrl.rootContainer
}

// UpdateMetricsOnly updates only the gauge values without rebuilding UI
func (ctrl *AdminDashboardController) UpdateMetricsOnly() {
	ctrl.mu.RLock()
	device := ctrl.appState.GetSelectedWorker()
	ctrl.mu.RUnlock()

	if device != nil {
		// Update gauges - SetValue is thread-safe (has its own locking)
		ctrl.cpuGauge.SetValue(device.CPUUsage)
		ctrl.ramGauge.SetValue(device.RAMUsage)
		ctrl.gpuGauge.SetValue(device.GPUUsage)

		// Update labels on main thread
		ramText := fmt.Sprintf("RAM: %s / %s",
			system.FormatBytes(device.RAMUsed),
			system.FormatBytes(device.RAMTotal))
		ctrl.runOnMain(func() {
			ctrl.ramDetailsLabel.SetText(ramText)
		})
	}
}

// ForceRebuild forces a complete UI rebuild (for worker list changes)
func (ctrl *AdminDashboardController) ForceRebuild() {
	ctrl.mu.Lock()
	ctrl.rootContainer = nil
	ctrl.currentWorkerID = ""
	ctrl.mu.Unlock()
}

// buildFullUI creates the complete dashboard UI
func (ctrl *AdminDashboardController) buildFullUI(workers []*state.DeviceInfo, device *state.DeviceInfo, selectedID string) fyne.CanvasObject {
	title := widget.NewLabelWithStyle(
		"admin:admin - Admin Panel",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	workerCountLabel := widget.NewLabel(fmt.Sprintf("Connected Workers: %d", len(workers)))

	// Worker list (left side)
	workerList := container.NewVBox()

	for _, worker := range workers {
		w := worker
		workerBtn := widget.NewButton(fmt.Sprintf("%s (%s)", w.Hostname, w.IPAddress), func() {
			ctrl.onSelectWorker(w.ID)
		})
		if w.ID == selectedID {
			workerBtn.Importance = widget.HighImportance
		}
		workerList.Add(workerBtn)
	}

	addWorkerBtn := widget.NewButton("+ Add Worker", ctrl.onAddWorker)
	workerList.Add(widget.NewSeparator())
	workerList.Add(addWorkerBtn)

	workerListContainer := container.NewVBox(
		widget.NewLabelWithStyle("Workers", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		workerList,
	)

	// Selected worker details (right side)
	var detailsContent fyne.CanvasObject

	if device != nil {
		detailsContent = ctrl.buildWorkerDetailsView(device)
	} else {
		detailsContent = widget.NewLabel("Select a worker from the list")
	}

	// Main content split
	split := container.NewHSplit(
		container.NewVScroll(workerListContainer),
		container.NewVScroll(detailsContent),
	)
	split.SetOffset(0.25)

	// Bottom buttons
	disconnectButton := widget.NewButton("Disconnect All", ctrl.onDisconnect)
	disconnectButton.Importance = widget.DangerImportance
	backButton := widget.NewButton("Back to Role Selection", ctrl.onBack)
	buttonSection := container.NewHBox(disconnectButton, backButton)

	content := container.NewBorder(
		container.NewVBox(title, workerCountLabel, widget.NewSeparator()),
		container.NewVBox(widget.NewSeparator(), buttonSection),
		nil, nil,
		split,
	)

	return content
}

// buildWorkerDetailsView creates the detailed view using the persistent gauges
func (ctrl *AdminDashboardController) buildWorkerDetailsView(device *state.DeviceInfo) fyne.CanvasObject {
	deviceHeader := widget.NewLabelWithStyle(
		device.Hostname,
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	osLabel := widget.NewLabel(fmt.Sprintf("OS: %s (%s)", device.OS, device.Architecture))
	ipLabel := widget.NewLabel(fmt.Sprintf("IP: %s", device.IPAddress))

	infoSection := container.NewVBox(
		deviceHeader,
		osLabel,
		ipLabel,
		ctrl.uptimeLabel,
	)

	// Use the persistent gauges
	gaugesRow := container.NewGridWithColumns(3,
		ctrl.cpuGauge,
		ctrl.ramGauge,
		ctrl.gpuGauge,
	)

	gpuLabel := widget.NewLabel(fmt.Sprintf("GPU: %s", device.GPUName))

	// SSH Button
	sshButton := widget.NewButton("Open SSH Terminal", func() {
		ctrl.onSSH(device.IPAddress)
	})
	sshButton.Importance = widget.MediumImportance

	return container.NewVBox(
		infoSection,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Resource Usage", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		gaugesRow,
		ctrl.ramDetailsLabel,
		gpuLabel,
		widget.NewSeparator(),
		sshButton,
	)
}

// Legacy function for compatibility - creates a new controller each time (old behavior)
func NewAdminDashboard(appState *state.AppState, onDisconnect func(), onBack func(), onAddWorker func(), onSelectWorker func(string), onSSH func(string)) fyne.CanvasObject {
	// For backwards compatibility, but this won't have smooth gauge animations
	// Use AdminDashboardController for proper behavior
	ctrl := NewAdminDashboardController(appState, onDisconnect, onBack, onAddWorker, onSelectWorker, onSSH)
	return ctrl.GetContent()
}

// NewSSHDialog creates an SSH connection dialog
func NewSSHDialog(workerIP string, onConnect func(ip, user, password string)) fyne.CanvasObject {
	title := widget.NewLabelWithStyle(
		"SSH Connection",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	ipLabel := widget.NewLabel(fmt.Sprintf("Connecting to: %s", workerIP))

	userEntry := widget.NewEntry()
	userEntry.SetPlaceHolder("Username")
	userEntry.SetText("admin")

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Password")

	connectBtn := widget.NewButton("Connect", func() {
		onConnect(workerIP, userEntry.Text, passwordEntry.Text)
	})
	connectBtn.Importance = widget.HighImportance

	return container.NewVBox(
		title,
		ipLabel,
		widget.NewSeparator(),
		widget.NewLabel("Username:"),
		userEntry,
		widget.NewLabel("Password:"),
		passwordEntry,
		connectBtn,
	)
}
