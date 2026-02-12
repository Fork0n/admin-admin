package ui

import (
	"adminadmin/internal/network"
	"adminadmin/internal/state"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// WorkerWaitingScreen shows the screen when waiting for admin connection
func NewWorkerWaitingScreen(localIP string, port int, onBack func()) fyne.CanvasObject {
	title := widget.NewLabelWithStyle(
		"admin:admin - Worker Node",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	waitingLabel := widget.NewLabelWithStyle(
		"Waiting for Admin connection...",
		fyne.TextAlignCenter,
		fyne.TextStyle{Italic: true},
	)

	// Connection Info
	ipLabel := widget.NewLabelWithStyle(
		fmt.Sprintf("Local IP: %s", localIP),
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	portLabel := widget.NewLabelWithStyle(
		fmt.Sprintf("Port: %d", port),
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	instructionLabel := widget.NewLabel("Give this IP to the Admin to connect")

	infoSection := container.NewVBox(
		widget.NewLabelWithStyle("Connection Info", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		ipLabel,
		portLabel,
		instructionLabel,
	)

	backButton := widget.NewButton("Back to Role Selection", onBack)

	content := container.NewVBox(
		title,
		widget.NewSeparator(),
		waitingLabel,
		widget.NewSeparator(),
		infoSection,
		widget.NewSeparator(),
		backButton,
	)

	return container.NewCenter(content)
}

// WorkerConnectedScreen shows the screen when admin is connected
func NewWorkerConnectedScreen(appState *state.AppState, onBack func()) fyne.CanvasObject {
	title := widget.NewLabelWithStyle(
		"admin:admin - Worker Node",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	admin := appState.GetConnectedAdmin()
	adminName := "Unknown"
	if admin != nil {
		adminName = admin.Hostname
	}

	connectedLabel := widget.NewLabelWithStyle(
		fmt.Sprintf("Connected to: %s", adminName),
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	statusLabel := widget.NewLabel("✓ Admin is monitoring this device")

	backButton := widget.NewButton("Back to Role Selection", onBack)

	content := container.NewVBox(
		title,
		widget.NewSeparator(),
		connectedLabel,
		statusLabel,
		widget.NewSeparator(),
		backButton,
	)

	return container.NewCenter(content)
}

// NewWorkerDashboard creates the worker dashboard (legacy, for compatibility)
func NewWorkerDashboard(onBack func()) fyne.CanvasObject {
	return NewWorkerWaitingScreen("", network.DefaultWorkerPort, onBack)
}
