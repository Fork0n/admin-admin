package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func NewRoleSelectScreen(onAdminSelected func(), onWorkerSelected func()) fyne.CanvasObject {
	title := widget.NewLabelWithStyle(
		"Desktop Control System",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	subtitle := widget.NewLabelWithStyle(
		"Select Role",
		fyne.TextAlignCenter,
		fyne.TextStyle{},
	)

	adminButton := widget.NewButton("Admin PC", onAdminSelected)
	adminButton.Importance = widget.HighImportance

	workerButton := widget.NewButton("Worker PC", onWorkerSelected)
	workerButton.Importance = widget.HighImportance

	buttonContainer := container.NewVBox(
		adminButton,
		workerButton,
	)

	content := container.NewVBox(
		title,
		subtitle,
		widget.NewSeparator(),
		buttonContainer,
	)

	return container.NewCenter(content)
}
