package ui

import (
	"adminadmin/internal/state"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func NewAdminDashboard(appState *state.AppState, onConnect func(string), onDisconnect func(), onBack func(), onRefresh func()) fyne.CanvasObject {
	title := widget.NewLabelWithStyle(
		"admin:admin - Admin Panel",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	// Connection input section
	connectionLabel := widget.NewLabel("Worker IP Address")
	connectionLabel.TextStyle = fyne.TextStyle{Bold: true}

	ipEntry := widget.NewEntry()
	ipEntry.SetPlaceHolder("192.168.1.100 note: only local network is supported")

	connectButton := widget.NewButton("Connect to Worker", func() {
		if ipEntry.Text != "" {
			onConnect(ipEntry.Text)
		}
	})
	connectButton.Importance = widget.HighImportance

	if appState.IsConnected() {
		ipEntry.Disable()
		connectButton.Disable()
	}

	connectionInputSection := container.NewVBox(
		connectionLabel,
		ipEntry,
		connectButton,
	)

	// Connection status section
	connectionStatusLabel := widget.NewLabel("Connection Status")
	connectionStatusLabel.TextStyle = fyne.TextStyle{Bold: true}

	var statusText string
	if appState.IsConnected() {
		statusText = "✓ Connected"
	} else {
		statusText = "✗ Not Connected"
	}
	statusValue := widget.NewLabel(statusText)

	connectionSection := container.NewVBox(
		connectionStatusLabel,
		statusValue,
	)

	// Device info section
	deviceInfoLabel := widget.NewLabel("Connected Device Info")
	deviceInfoLabel.TextStyle = fyne.TextStyle{Bold: true}

	hostnameLabel := widget.NewLabel("Hostname: N/A")
	osLabel := widget.NewLabel("OS: N/A")
	archLabel := widget.NewLabel("Architecture: N/A")
	ipLabel := widget.NewLabel("IP Address: N/A")

	if device := appState.GetConnectedDevice(); device != nil {
		hostnameLabel.SetText("Hostname: " + device.Hostname)
		osLabel.SetText("OS: " + device.OS)
		archLabel.SetText("Architecture: " + device.Architecture)
		ipLabel.SetText("IP Address: " + device.IPAddress)
	}

	deviceInfoSection := container.NewVBox(
		deviceInfoLabel,
		hostnameLabel,
		osLabel,
		archLabel,
		ipLabel,
	)

	// Buttons section
	disconnectButton := widget.NewButton("Disconnect", onDisconnect)
	if !appState.IsConnected() {
		disconnectButton.Disable()
	}

	refreshButton := widget.NewButton("Refresh", onRefresh)
	backButton := widget.NewButton("Back to Role Selection", onBack)

	buttonContainer := container.NewVBox(
		disconnectButton,
		refreshButton,
		backButton,
	)

	content := container.NewVBox(
		title,
		widget.NewSeparator(),
		connectionInputSection,
		widget.NewSeparator(),
		connectionSection,
		widget.NewSeparator(),
		deviceInfoSection,
		widget.NewSeparator(),
		buttonContainer,
	)

	return container.NewBorder(nil, nil, nil, nil, content)
}
