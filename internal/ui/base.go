// Package ui provides the UI framework for admin:admin application.
// This package contains reusable UI components and screen builders.
package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// Screen represents a UI screen with a title and content
type Screen struct {
	Title   string
	Content fyne.CanvasObject
}

// ScreenBuilder helps construct screens with consistent styling
type ScreenBuilder struct {
	title    string
	sections []fyne.CanvasObject
}

// NewScreenBuilder creates a new screen builder with a title
func NewScreenBuilder(title string) *ScreenBuilder {
	return &ScreenBuilder{
		title:    title,
		sections: make([]fyne.CanvasObject, 0),
	}
}

// AddSection adds a section to the screen
func (sb *ScreenBuilder) AddSection(section fyne.CanvasObject) *ScreenBuilder {
	sb.sections = append(sb.sections, section)
	return sb
}

// AddSeparator adds a visual separator
func (sb *ScreenBuilder) AddSeparator() *ScreenBuilder {
	sb.sections = append(sb.sections, widget.NewSeparator())
	return sb
}

// Build creates the final screen layout
func (sb *ScreenBuilder) Build() fyne.CanvasObject {
	titleLabel := widget.NewLabelWithStyle(
		sb.title,
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	elements := []fyne.CanvasObject{titleLabel, widget.NewSeparator()}
	elements = append(elements, sb.sections...)

	content := container.NewVBox(elements...)
	return container.NewBorder(nil, nil, nil, nil, content)
}

// BuildCentered creates a centered screen layout
func (sb *ScreenBuilder) BuildCentered() fyne.CanvasObject {
	titleLabel := widget.NewLabelWithStyle(
		sb.title,
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	elements := []fyne.CanvasObject{titleLabel, widget.NewSeparator()}
	elements = append(elements, sb.sections...)

	content := container.NewVBox(elements...)
	return container.NewCenter(content)
}

// SectionBuilder helps construct screen sections
type SectionBuilder struct {
	title    string
	items    []fyne.CanvasObject
	hasTitle bool
}

// NewSection creates a new section builder
func NewSection(title string) *SectionBuilder {
	return &SectionBuilder{
		title:    title,
		items:    make([]fyne.CanvasObject, 0),
		hasTitle: title != "",
	}
}

// AddItem adds an item to the section
func (s *SectionBuilder) AddItem(item fyne.CanvasObject) *SectionBuilder {
	s.items = append(s.items, item)
	return s
}

// AddLabel adds a text label
func (s *SectionBuilder) AddLabel(text string) *SectionBuilder {
	s.items = append(s.items, widget.NewLabel(text))
	return s
}

// AddBoldLabel adds a bold text label
func (s *SectionBuilder) AddBoldLabel(text string) *SectionBuilder {
	label := widget.NewLabel(text)
	label.TextStyle = fyne.TextStyle{Bold: true}
	s.items = append(s.items, label)
	return s
}

// Build creates the section container
func (s *SectionBuilder) Build() fyne.CanvasObject {
	elements := make([]fyne.CanvasObject, 0)

	if s.hasTitle {
		titleLabel := widget.NewLabel(s.title)
		titleLabel.TextStyle = fyne.TextStyle{Bold: true}
		elements = append(elements, titleLabel)
	}

	elements = append(elements, s.items...)
	return container.NewVBox(elements...)
}

// ButtonBuilder helps create styled buttons
type ButtonBuilder struct {
	buttons []*widget.Button
}

// NewButtonGroup creates a new button builder
func NewButtonGroup() *ButtonBuilder {
	return &ButtonBuilder{
		buttons: make([]*widget.Button, 0),
	}
}

// AddPrimaryButton adds a high-importance button
func (bb *ButtonBuilder) AddPrimaryButton(label string, onClick func()) *ButtonBuilder {
	btn := widget.NewButton(label, onClick)
	btn.Importance = widget.HighImportance
	bb.buttons = append(bb.buttons, btn)
	return bb
}

// AddSecondaryButton adds a medium-importance button
func (bb *ButtonBuilder) AddSecondaryButton(label string, onClick func()) *ButtonBuilder {
	btn := widget.NewButton(label, onClick)
	btn.Importance = widget.MediumImportance
	bb.buttons = append(bb.buttons, btn)
	return bb
}

// AddButton adds a normal button
func (bb *ButtonBuilder) AddButton(label string, onClick func()) *ButtonBuilder {
	btn := widget.NewButton(label, onClick)
	bb.buttons = append(bb.buttons, btn)
	return bb
}

// AddDisabledButton adds a disabled button
func (bb *ButtonBuilder) AddDisabledButton(label string) *ButtonBuilder {
	btn := widget.NewButton(label, func() {})
	btn.Disable()
	bb.buttons = append(bb.buttons, btn)
	return bb
}

// Build creates a vertical button container
func (bb *ButtonBuilder) Build() fyne.CanvasObject {
	items := make([]fyne.CanvasObject, len(bb.buttons))
	for i, btn := range bb.buttons {
		items[i] = btn
	}
	return container.NewVBox(items...)
}

// BuildHorizontal creates a horizontal button container
func (bb *ButtonBuilder) BuildHorizontal() fyne.CanvasObject {
	items := make([]fyne.CanvasObject, len(bb.buttons))
	for i, btn := range bb.buttons {
		items[i] = btn
	}
	return container.NewHBox(items...)
}

// CreateInfoCard creates an info display card with label-value pairs
func CreateInfoCard(title string, info map[string]string) fyne.CanvasObject {
	section := NewSection(title)
	for key, value := range info {
		section.AddLabel(key + ": " + value)
	}
	return section.Build()
}

// CreateStatusIndicator creates a status indicator with icon
func CreateStatusIndicator(connected bool) *widget.Label {
	var text string
	if connected {
		text = "✓ Connected"
	} else {
		text = "✗ Not Connected"
	}
	label := widget.NewLabel(text)
	label.TextStyle = fyne.TextStyle{Bold: true}
	return label
}
