package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"image/color"
)

// Theme colors for admin:admin
var (
	ColorPrimary   = color.NRGBA{R: 66, G: 133, B: 244, A: 255}  // Blue
	ColorSuccess   = color.NRGBA{R: 52, G: 168, B: 83, A: 255}   // Green
	ColorWarning   = color.NRGBA{R: 251, G: 188, B: 4, A: 255}   // Yellow
	ColorDanger    = color.NRGBA{R: 234, G: 67, B: 53, A: 255}   // Red
	ColorSecondary = color.NRGBA{R: 128, G: 128, B: 128, A: 255} // Gray
)

// StatusType represents different status states
type StatusType int

const (
	StatusInfo StatusType = iota
	StatusSuccess
	StatusWarning
	StatusDanger
)

// CreateColoredLabel creates a label with a colored background
func CreateColoredLabel(text string, bgColor color.Color) fyne.CanvasObject {
	label := widget.NewLabel(text)
	bg := canvas.NewRectangle(bgColor)
	return container.NewStack(bg, container.NewPadded(label))
}

// CreateStatusBadge creates a status badge with appropriate styling
func CreateStatusBadge(text string, status StatusType) fyne.CanvasObject {
	var bgColor color.Color
	switch status {
	case StatusSuccess:
		bgColor = ColorSuccess
	case StatusWarning:
		bgColor = ColorWarning
	case StatusDanger:
		bgColor = ColorDanger
	default:
		bgColor = ColorPrimary
	}
	return CreateColoredLabel(text, bgColor)
}

// CreateCard creates a card-like container with padding
func CreateCard(title string, content fyne.CanvasObject) fyne.CanvasObject {
	titleLabel := widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	return container.NewVBox(
		titleLabel,
		widget.NewSeparator(),
		container.NewPadded(content),
	)
}

// CreateIconLabel creates a label with an icon prefix
func CreateIconLabel(icon, text string) *widget.Label {
	return widget.NewLabel(icon + " " + text)
}

// TableBuilder helps create table-like displays
type TableBuilder struct {
	headers []string
	rows    [][]string
}

// NewTableBuilder creates a new table builder
func NewTableBuilder(headers ...string) *TableBuilder {
	return &TableBuilder{
		headers: headers,
		rows:    make([][]string, 0),
	}
}

// AddRow adds a row to the table
func (tb *TableBuilder) AddRow(values ...string) *TableBuilder {
	tb.rows = append(tb.rows, values)
	return tb
}

// Build creates a simple table display
func (tb *TableBuilder) Build() fyne.CanvasObject {
	section := NewSection("")

	// Header row
	headerLabels := make([]fyne.CanvasObject, len(tb.headers))
	for i, h := range tb.headers {
		label := widget.NewLabel(h)
		label.TextStyle = fyne.TextStyle{Bold: true}
		headerLabels[i] = label
	}
	section.AddItem(container.NewHBox(headerLabels...))
	section.AddItem(widget.NewSeparator())

	// Data rows
	for _, row := range tb.rows {
		rowLabels := make([]fyne.CanvasObject, len(row))
		for i, cell := range row {
			rowLabels[i] = widget.NewLabel(cell)
		}
		section.AddItem(container.NewHBox(rowLabels...))
	}

	return section.Build()
}

// ProgressSection creates a section with a progress indicator
func CreateProgressSection(title string, value float64, max float64) fyne.CanvasObject {
	section := NewSection(title)

	progress := widget.NewProgressBar()
	progress.Max = max
	progress.SetValue(value)

	section.AddItem(progress)
	return section.Build()
}

// CreateLoadingIndicator creates a loading spinner with optional text
func CreateLoadingIndicator(text string) fyne.CanvasObject {
	progress := widget.NewProgressBarInfinite()
	if text == "" {
		return progress
	}
	return container.NewVBox(
		widget.NewLabel(text),
		progress,
	)
}
