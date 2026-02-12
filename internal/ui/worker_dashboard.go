package ui

import (
	"adminadmin/internal/network"
	"adminadmin/internal/system"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func NewWorkerDashboard(onBack func()) fyne.CanvasObject {
	title := widget.NewLabelWithStyle(
		"Worker Node",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	// Server info
	serverInfoLabel := widget.NewLabel("Server Status")
	serverInfoLabel.TextStyle = fyne.TextStyle{Bold: true}

	portLabel := widget.NewLabel(fmt.Sprintf("✓ Listening on port: %d", network.DefaultWorkerPort))
	portLabel.TextStyle = fyne.TextStyle{Bold: true}

	instructionLabel := widget.NewLabel("Admins can connect to this worker using your IP address")

	serverInfoSection := container.NewVBox(
		serverInfoLabel,
		portLabel,
		instructionLabel,
	)

	// System info
	systemInfoLabel := widget.NewLabel("System Information")
	systemInfoLabel.TextStyle = fyne.TextStyle{Bold: true}

	sysInfo := system.GetLocalSystemInfo()

	hostnameLabel := widget.NewLabel("Hostname: " + sysInfo.Hostname)
	osLabel := widget.NewLabel("OS: " + sysInfo.OS)
	archLabel := widget.NewLabel("Architecture: " + sysInfo.Arch)
	goVersionLabel := widget.NewLabel("Go Runtime: " + sysInfo.GoVersion)
	cpuLabel := widget.NewLabel(fmt.Sprintf("CPU Usage: %.2f%%", sysInfo.CPUUsage))
	ramLabel := widget.NewLabel(fmt.Sprintf("RAM Usage: %.2f%%", sysInfo.RAMUsage))

	systemInfoSection := container.NewVBox(
		systemInfoLabel,
		hostnameLabel,
		osLabel,
		archLabel,
		goVersionLabel,
		cpuLabel,
		ramLabel,
	)

	backButton := widget.NewButton("Back to Role Selection", onBack)

	buttonContainer := container.NewVBox(
		backButton,
	)

	content := container.NewVBox(
		title,
		widget.NewSeparator(),
		serverInfoSection,
		widget.NewSeparator(),
		systemInfoSection,
		widget.NewSeparator(),
		buttonContainer,
	)

	return container.NewBorder(nil, nil, nil, nil, content)
}
