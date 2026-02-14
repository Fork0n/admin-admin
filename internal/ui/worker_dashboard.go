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
// onCredentialsChange is called when SSH credentials are updated
func NewWorkerWaitingScreen(localIP string, port int, onBack func(), onCredentialsChange func(username, password string)) fyne.CanvasObject {
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

	// SSH Credentials Section
	sshHeader := widget.NewLabelWithStyle("SSH Credentials", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("Username")
	usernameEntry.SetText(network.DefaultSSHUsername)

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Password")
	passwordEntry.SetText(network.DefaultSSHPassword)

	// Update credentials when changed
	updateCredentials := func() {
		if onCredentialsChange != nil {
			onCredentialsChange(usernameEntry.Text, passwordEntry.Text)
		}
	}

	usernameEntry.OnChanged = func(s string) { updateCredentials() }
	passwordEntry.OnChanged = func(s string) { updateCredentials() }

	sshPortLabel := widget.NewLabel(fmt.Sprintf("SSH Port: %d", network.DefaultSSHPort))

	sshSection := container.NewVBox(
		sshHeader,
		widget.NewLabel("Username:"),
		usernameEntry,
		widget.NewLabel("Password:"),
		passwordEntry,
		sshPortLabel,
	)

	backButton := widget.NewButton("Back to Role Selection", onBack)

	content := container.NewVBox(
		title,
		widget.NewSeparator(),
		waitingLabel,
		widget.NewSeparator(),
		infoSection,
		widget.NewSeparator(),
		sshSection,
		widget.NewSeparator(),
		backButton,
	)

	return container.NewCenter(content)
}

// WorkerConnectedScreen shows the screen when admin is connected
// Returns the content and a flag indicating this should use a compact window
func NewWorkerConnectedScreen(appState *state.AppState, onBack func()) fyne.CanvasObject {
	admin := appState.GetConnectedAdmin()
	adminName := "Unknown"
	if admin != nil {
		adminName = admin.Hostname
	}

	// Compact status display
	statusIcon := widget.NewLabelWithStyle("✓", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	connectedLabel := widget.NewLabelWithStyle(
		fmt.Sprintf("Connected to: %s", adminName),
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	sshLabel := widget.NewLabel(fmt.Sprintf("SSH available on port %d", network.DefaultSSHPort))
	sshLabel.Alignment = fyne.TextAlignCenter

	backButton := widget.NewButton("Disconnect", onBack)
	backButton.Importance = widget.DangerImportance

	content := container.NewVBox(
		container.NewHBox(statusIcon, connectedLabel),
		sshLabel,
		backButton,
	)

	return container.NewPadded(content)
}

// WorkerConnectedScreenCompact returns true to indicate compact mode is preferred
func WorkerConnectedScreenCompact() bool {
	return true
}

// NewWorkerDashboard creates the worker dashboard (legacy, for compatibility)
func NewWorkerDashboard(onBack func()) fyne.CanvasObject {
	return NewWorkerWaitingScreen("", network.DefaultWorkerPort, onBack, nil)
}
