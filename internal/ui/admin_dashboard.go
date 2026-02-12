package ui

import (
	"adminadmin/internal/state"
	"adminadmin/internal/system"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
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

	// IP input
	ipEntry := widget.NewEntry()
	ipEntry.SetPlaceHolder("Enter Worker IP (e.g., 192.168.1.100)")

	connectButton := widget.NewButton("Connect", func() {
		if ipEntry.Text != "" {
			onConnect(ipEntry.Text)
		}
	})
	connectButton.Importance = widget.HighImportance

	backButton := widget.NewButton("Back", onBack)

	content := container.NewVBox(
		title,
		subtitle,
		widget.NewSeparator(),
		widget.NewLabel("Worker IP Address:"),
		ipEntry,
		connectButton,
		widget.NewSeparator(),
		backButton,
	)

	return container.NewCenter(content)
}

// AdminDashboard shows the connected dashboard with device info and gauges
func NewAdminDashboard(appState *state.AppState, onDisconnect func(), onBack func(), onAddWorker func(), onSelectWorker func(string), onSSH func(string)) fyne.CanvasObject {
	workers := appState.GetConnectedDevicesList()

	if len(workers) == 0 {
		return container.NewCenter(widget.NewLabel("No workers connected"))
	}

	title := widget.NewLabelWithStyle(
		"admin:admin - Admin Panel",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	// Worker count
	workerCountLabel := widget.NewLabel(fmt.Sprintf("Connected Workers: %d", len(workers)))

	// Worker list (left side)
	workerList := container.NewVBox()
	selectedID := appState.GetSelectedWorkerID()

	for _, worker := range workers {
		w := worker // Capture for closure
		workerBtn := widget.NewButton(fmt.Sprintf("%s (%s)", w.Hostname, w.IPAddress), func() {
			onSelectWorker(w.ID)
		})
		if w.ID == selectedID {
			workerBtn.Importance = widget.HighImportance
		}
		workerList.Add(workerBtn)
	}

	// Add worker button
	addWorkerBtn := widget.NewButton("+ Add Worker", onAddWorker)
	workerList.Add(widget.NewSeparator())
	workerList.Add(addWorkerBtn)

	workerListContainer := container.NewVBox(
		widget.NewLabelWithStyle("Workers", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		workerList,
	)

	// Selected worker details (right side)
	device := appState.GetSelectedWorker()
	var detailsContent fyne.CanvasObject

	if device != nil {
		detailsContent = createWorkerDetailsView(device, onSSH)
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
	disconnectButton := widget.NewButton("Disconnect All", onDisconnect)
	disconnectButton.Importance = widget.DangerImportance
	backButton := widget.NewButton("Back to Role Selection", onBack)
	buttonSection := container.NewHBox(disconnectButton, backButton)

	content := container.NewBorder(
		container.NewVBox(title, workerCountLabel, widget.NewSeparator()),
		container.NewVBox(widget.NewSeparator(), buttonSection),
		nil, nil,
		split,
	)

	return content
}

// createWorkerDetailsView creates the detailed view for a selected worker
func createWorkerDetailsView(device *state.DeviceInfo, onSSH func(string)) fyne.CanvasObject {
	// Device Info Header
	deviceHeader := widget.NewLabelWithStyle(
		device.Hostname,
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	osLabel := widget.NewLabel(fmt.Sprintf("OS: %s (%s)", device.OS, device.Architecture))
	ipLabel := widget.NewLabel(fmt.Sprintf("IP: %s", device.IPAddress))
	uptimeLabel := widget.NewLabel(fmt.Sprintf("Uptime: %s", system.FormatUptime(device.Uptime)))

	infoSection := container.NewVBox(
		deviceHeader,
		osLabel,
		ipLabel,
		uptimeLabel,
	)

	// Gauges Section - Speedometer style
	cpuGauge := CreateSpeedometerGauge("CPU", device.CPUUsage, "%")
	ramGauge := CreateSpeedometerGauge("RAM", device.RAMUsage, "%")
	gpuGauge := CreateSpeedometerGauge("GPU", device.GPUUsage, "%")

	gaugesRow := container.NewGridWithColumns(3,
		cpuGauge,
		ramGauge,
		gpuGauge,
	)

	// RAM Details
	ramDetails := widget.NewLabel(fmt.Sprintf("RAM: %s / %s",
		system.FormatBytes(device.RAMUsed),
		system.FormatBytes(device.RAMTotal)))

	// GPU Info
	gpuLabel := widget.NewLabel(fmt.Sprintf("GPU: %s", device.GPUName))

	// SSH Button
	sshButton := widget.NewButton("Open SSH Terminal", func() {
		onSSH(device.IPAddress)
	})
	sshButton.Importance = widget.MediumImportance

	return container.NewVBox(
		infoSection,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Resource Usage", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		gaugesRow,
		ramDetails,
		gpuLabel,
		widget.NewSeparator(),
		sshButton,
	)
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
