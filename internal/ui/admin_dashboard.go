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
		"admin:admin - Admin Panel",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	subtitle := widget.NewLabelWithStyle(
		"Connect to a Worker PC",
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

// AdminDashboard shows the connected dashboard with device info
func NewAdminDashboard(appState *state.AppState, onDisconnect func(), onBack func()) fyne.CanvasObject {
	device := appState.GetConnectedDevice()
	if device == nil {
		return widget.NewLabel("No device connected")
	}

	title := widget.NewLabelWithStyle(
		"admin:admin - Admin Panel",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	// Device Info Section
	deviceNameLabel := widget.NewLabelWithStyle(
		fmt.Sprintf("Connected to: %s", device.Hostname),
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	osLabel := widget.NewLabel(fmt.Sprintf("OS: %s (%s)", device.OS, device.Architecture))
	ipLabel := widget.NewLabel(fmt.Sprintf("IP Address: %s", device.IPAddress))
	uptimeLabel := widget.NewLabel(fmt.Sprintf("Uptime: %s", system.FormatUptime(device.Uptime)))

	deviceSection := container.NewVBox(
		widget.NewLabelWithStyle("Device Information", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		deviceNameLabel,
		osLabel,
		ipLabel,
		uptimeLabel,
	)

	// Resource Usage Section
	cpuLabel := widget.NewLabel(fmt.Sprintf("CPU Usage: %.1f%%", device.CPUUsage))
	ramLabel := widget.NewLabel(fmt.Sprintf("RAM Usage: %.1f%% (%s / %s)",
		device.RAMUsage,
		system.FormatBytes(device.RAMUsed),
		system.FormatBytes(device.RAMTotal)))
	gpuNameLabel := widget.NewLabel(fmt.Sprintf("GPU: %s", device.GPUName))
	gpuUsageLabel := widget.NewLabel(fmt.Sprintf("GPU Usage: %.1f%%", device.GPUUsage))
	internetLabel := widget.NewLabel(fmt.Sprintf("Internet: %s", device.InternetSpeed))

	resourceSection := container.NewVBox(
		widget.NewLabelWithStyle("Resource Usage", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		cpuLabel,
		ramLabel,
		gpuNameLabel,
		gpuUsageLabel,
		internetLabel,
	)

	// Buttons
	disconnectButton := widget.NewButton("Disconnect", onDisconnect)
	disconnectButton.Importance = widget.DangerImportance

	backButton := widget.NewButton("Back to Role Selection", onBack)

	buttonSection := container.NewVBox(
		disconnectButton,
		backButton,
	)

	content := container.NewVBox(
		title,
		widget.NewSeparator(),
		deviceSection,
		widget.NewSeparator(),
		resourceSection,
		widget.NewSeparator(),
		buttonSection,
	)

	return container.NewBorder(nil, nil, nil, nil, content)
}
