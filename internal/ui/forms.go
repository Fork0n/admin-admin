package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// FormField represents a form input field
type FormField struct {
	Label       string
	Placeholder string
	Entry       *widget.Entry
	Validator   func(string) error
}

// FormBuilder helps create input forms
type FormBuilder struct {
	fields []*FormField
}

// NewFormBuilder creates a new form builder
func NewFormBuilder() *FormBuilder {
	return &FormBuilder{
		fields: make([]*FormField, 0),
	}
}

// AddField adds a text input field
func (fb *FormBuilder) AddField(label, placeholder string) *FormBuilder {
	entry := widget.NewEntry()
	entry.SetPlaceHolder(placeholder)

	fb.fields = append(fb.fields, &FormField{
		Label:       label,
		Placeholder: placeholder,
		Entry:       entry,
	})
	return fb
}

// AddPasswordField adds a password input field
func (fb *FormBuilder) AddPasswordField(label, placeholder string) *FormBuilder {
	entry := widget.NewPasswordEntry()
	entry.SetPlaceHolder(placeholder)

	fb.fields = append(fb.fields, &FormField{
		Label:       label,
		Placeholder: placeholder,
		Entry:       entry,
	})
	return fb
}

// AddMultiLineField adds a multi-line text input field
func (fb *FormBuilder) AddMultiLineField(label, placeholder string) *FormBuilder {
	entry := widget.NewMultiLineEntry()
	entry.SetPlaceHolder(placeholder)

	fb.fields = append(fb.fields, &FormField{
		Label:       label,
		Placeholder: placeholder,
		Entry:       entry,
	})
	return fb
}

// GetField returns a field by label
func (fb *FormBuilder) GetField(label string) *FormField {
	for _, field := range fb.fields {
		if field.Label == label {
			return field
		}
	}
	return nil
}

// GetValue returns the value of a field by label
func (fb *FormBuilder) GetValue(label string) string {
	field := fb.GetField(label)
	if field != nil {
		return field.Entry.Text
	}
	return ""
}

// Build creates the form layout
func (fb *FormBuilder) Build() fyne.CanvasObject {
	section := NewSection("")

	for _, field := range fb.fields {
		label := widget.NewLabel(field.Label)
		label.TextStyle = fyne.TextStyle{Bold: true}
		section.AddItem(label)
		section.AddItem(field.Entry)
	}

	return section.Build()
}

// DisableAll disables all form fields
func (fb *FormBuilder) DisableAll() {
	for _, field := range fb.fields {
		field.Entry.Disable()
	}
}

// EnableAll enables all form fields
func (fb *FormBuilder) EnableAll() {
	for _, field := range fb.fields {
		field.Entry.Enable()
	}
}

// ClearAll clears all form fields
func (fb *FormBuilder) ClearAll() {
	for _, field := range fb.fields {
		field.Entry.SetText("")
	}
}
