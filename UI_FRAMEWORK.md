# admin:admin UI Framework Documentation

This document explains how to create custom UIs using the admin:admin UI framework. The framework provides a set of reusable components and builders to create consistent, clean user interfaces.

## Table of Contents

1. [Overview](#overview)
2. [Screen Builder](#screen-builder)
3. [Section Builder](#section-builder)
4. [Button Builder](#button-builder)
5. [Form Builder](#form-builder)
6. [UI Components](#ui-components)
7. [Creating Custom Screens](#creating-custom-screens)
8. [Best Practices](#best-practices)

---

## Overview

The UI framework is located in `internal/ui/` and consists of the following files:

| File | Description |
|------|-------------|
| `base.go` | Core screen and section builders |
| `forms.go` | Form input builders |
| `components.go` | Reusable UI components (cards, badges, tables) |
| `role_select.go` | Role selection screen |
| `admin_dashboard.go` | Admin dashboard screen |
| `worker_dashboard.go` | Worker dashboard screen |

---

## Screen Builder

The `ScreenBuilder` helps create consistent screen layouts with titles and sections.

### Basic Usage

```go
import "adminadmin/internal/ui"

func MyCustomScreen() fyne.CanvasObject {
    return ui.NewScreenBuilder("admin:admin - My Screen").
        AddSection(mySection1).
        AddSeparator().
        AddSection(mySection2).
        Build()
}
```

### Methods

| Method | Description |
|--------|-------------|
| `NewScreenBuilder(title)` | Creates a new builder with a title |
| `AddSection(object)` | Adds a section to the screen |
| `AddSeparator()` | Adds a visual separator line |
| `Build()` | Creates a border-layout screen |
| `BuildCentered()` | Creates a centered screen layout |

### Example: Centered Screen

```go
func WelcomeScreen(onContinue func()) fyne.CanvasObject {
    buttons := ui.NewButtonGroup().
        AddPrimaryButton("Continue", onContinue).
        Build()
    
    return ui.NewScreenBuilder("admin:admin - Welcome").
        AddSection(ui.NewSection("").
            AddLabel("Welcome to admin:admin!").
            AddLabel("Click Continue to proceed.").
            Build()).
        AddSeparator().
        AddSection(buttons).
        BuildCentered()
}
```

---

## Section Builder

The `SectionBuilder` creates content sections with optional titles.

### Basic Usage

```go
section := ui.NewSection("Connection Status").
    AddLabel("Status: Connected").
    AddLabel("IP: 192.168.1.100").
    Build()
```

### Methods

| Method | Description |
|--------|-------------|
| `NewSection(title)` | Creates a section (empty string = no title) |
| `AddItem(object)` | Adds any Fyne canvas object |
| `AddLabel(text)` | Adds a text label |
| `AddBoldLabel(text)` | Adds a bold text label |
| `Build()` | Creates the section container |

### Example: Info Section

```go
func createDeviceInfoSection(hostname, os, arch string) fyne.CanvasObject {
    return ui.NewSection("Device Information").
        AddLabel("Hostname: " + hostname).
        AddLabel("OS: " + os).
        AddLabel("Architecture: " + arch).
        Build()
}
```

---

## Button Builder

The `ButtonBuilder` creates groups of styled buttons.

### Basic Usage

```go
buttons := ui.NewButtonGroup().
    AddPrimaryButton("Connect", onConnect).
    AddButton("Refresh", onRefresh).
    AddSecondaryButton("Back", onBack).
    Build()
```

### Methods

| Method | Description |
|--------|-------------|
| `NewButtonGroup()` | Creates a new button builder |
| `AddPrimaryButton(label, onClick)` | Adds a high-importance button |
| `AddSecondaryButton(label, onClick)` | Adds a medium-importance button |
| `AddButton(label, onClick)` | Adds a normal button |
| `AddDisabledButton(label)` | Adds a disabled button |
| `Build()` | Creates vertical button layout |
| `BuildHorizontal()` | Creates horizontal button layout |

### Example: Action Buttons

```go
func createActionButtons(isConnected bool, onDisconnect, onRefresh func()) fyne.CanvasObject {
    group := ui.NewButtonGroup()
    
    if isConnected {
        group.AddPrimaryButton("Disconnect", onDisconnect)
    } else {
        group.AddDisabledButton("Disconnect")
    }
    
    group.AddButton("Refresh", onRefresh)
    
    return group.Build()
}
```

---

## Form Builder

The `FormBuilder` creates input forms with labels and validation.

### Basic Usage

```go
form := ui.NewFormBuilder().
    AddField("IP Address", "192.168.1.100").
    AddField("Port", "9876").
    Build()

// Get values later
ip := form.GetValue("IP Address")
port := form.GetValue("Port")
```

### Methods

| Method | Description |
|--------|-------------|
| `NewFormBuilder()` | Creates a new form builder |
| `AddField(label, placeholder)` | Adds a text input field |
| `AddPasswordField(label, placeholder)` | Adds a password input field |
| `AddMultiLineField(label, placeholder)` | Adds a multi-line text field |
| `GetField(label)` | Gets a field by label |
| `GetValue(label)` | Gets the value of a field |
| `DisableAll()` | Disables all fields |
| `EnableAll()` | Enables all fields |
| `ClearAll()` | Clears all field values |
| `Build()` | Creates the form layout |

### Example: Login Form

```go
func createLoginForm(onSubmit func(user, pass string)) fyne.CanvasObject {
    form := ui.NewFormBuilder().
        AddField("Username", "Enter username").
        AddPasswordField("Password", "Enter password")
    
    submitBtn := widget.NewButton("Login", func() {
        onSubmit(form.GetValue("Username"), form.GetValue("Password"))
    })
    
    return container.NewVBox(
        form.Build(),
        submitBtn,
    )
}
```

---

## UI Components

### Status Indicator

```go
// Creates "✓ Connected" or "✗ Not Connected"
status := ui.CreateStatusIndicator(isConnected)
```

### Info Card

```go
info := map[string]string{
    "Hostname": "PC-001",
    "OS":       "Windows",
    "Arch":     "amd64",
}
card := ui.CreateInfoCard("Device Info", info)
```

### Status Badge

```go
badge := ui.CreateStatusBadge("Online", ui.StatusSuccess)
badge := ui.CreateStatusBadge("Offline", ui.StatusDanger)
badge := ui.CreateStatusBadge("Warning", ui.StatusWarning)
```

### Cards

```go
content := widget.NewLabel("Card content goes here")
card := ui.CreateCard("Card Title", content)
```

### Icon Labels

```go
label := ui.CreateIconLabel("✓", "Task completed")
label := ui.CreateIconLabel("⚠", "Warning message")
```

### Loading Indicator

```go
loading := ui.CreateLoadingIndicator("Connecting...")
```

### Progress Section

```go
// CPU usage at 45%
progress := ui.CreateProgressSection("CPU Usage", 45.0, 100.0)
```

### Tables

```go
table := ui.NewTableBuilder("Name", "Status", "IP").
    AddRow("Worker-1", "Online", "192.168.1.101").
    AddRow("Worker-2", "Offline", "192.168.1.102").
    Build()
```

---

## Creating Custom Screens

### Step 1: Create a New File

Create a new file in `internal/ui/`, for example `my_screen.go`:

```go
package ui

import (
    "fyne.io/fyne/v2"
)

func NewMyCustomScreen(onAction func(), onBack func()) fyne.CanvasObject {
    // Your screen code here
}
```

### Step 2: Build the Screen

Use the builders to construct your screen:

```go
package ui

import (
    "fyne.io/fyne/v2"
)

func NewMyCustomScreen(data string, onAction func(), onBack func()) fyne.CanvasObject {
    // Create sections
    infoSection := NewSection("Information").
        AddLabel("Data: " + data).
        AddLabel("Status: Active").
        Build()
    
    // Create buttons
    buttons := NewButtonGroup().
        AddPrimaryButton("Do Action", onAction).
        AddButton("Back", onBack).
        Build()
    
    // Build the screen
    return NewScreenBuilder("admin:admin - My Custom Screen").
        AddSection(infoSection).
        AddSeparator().
        AddSection(buttons).
        Build()
}
```

### Step 3: Wire It Up in app.go

Add a method in `internal/application/app.go`:

```go
func (a *App) showMyCustomScreen() {
    content := ui.NewMyCustomScreen(
        a.state.GetSomeData(),
        func() { a.doSomeAction() },
        func() { a.backToRoleSelection() },
    )
    a.window.SetContent(content)
}
```

---

## Best Practices

### 1. Naming Conventions

- Screen functions: `New<Name>Screen()` or `New<Name>Dashboard()`
- Use descriptive names: `NewWorkerDashboard()`, `NewConnectionScreen()`

### 2. Callback Pattern

Pass callbacks as function parameters to keep UI decoupled from logic:

```go
func NewMyScreen(
    onConnect func(ip string),  // Takes IP as parameter
    onDisconnect func(),        // No parameters
    onBack func(),
) fyne.CanvasObject
```

### 3. State Management

Pass state through parameters, don't import state directly in UI:

```go
// Good - state passed as parameter
func NewDashboard(isConnected bool, deviceName string) fyne.CanvasObject

// Avoid - importing state directly
func NewDashboard() fyne.CanvasObject {
    state := someGlobalState  // Don't do this
}
```

### 4. Consistent Titles

Always prefix screen titles with "admin:admin - ":

```go
NewScreenBuilder("admin:admin - Settings")
NewScreenBuilder("admin:admin - Connection Status")
```

### 5. Separator Usage

Use separators to visually group related content:

```go
return NewScreenBuilder("admin:admin - Dashboard").
    AddSection(headerSection).
    AddSeparator().         // Separates header from content
    AddSection(contentSection).
    AddSeparator().         // Separates content from actions
    AddSection(buttonSection).
    Build()
```

---

## File Structure

Recommended organization for UI files:

```
internal/ui/
├── base.go              # Core builders (ScreenBuilder, SectionBuilder, ButtonBuilder)
├── forms.go             # Form input builders
├── components.go        # Reusable UI components (cards, badges, tables)
├── role_select.go       # Role selection screen
├── admin_dashboard.go   # Admin dashboard screen
├── worker_dashboard.go  # Worker dashboard screen
└── <your_screen>.go     # Your custom screens (e.g., settings.go, logs.go)
```

### Adding a New Screen File

1. Create a new `.go` file in `internal/ui/`
2. Use `package ui` at the top
3. Import the required Fyne packages
4. Create your screen function using the builders
5. Wire it up in `internal/application/app.go`

---

## Complete Example

Here's a complete example of a settings screen:

```go
// settings_screen.go
package ui

import (
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/widget"
)

type SettingsConfig struct {
    Port        int
    Timeout     int
    AutoConnect bool
}

func NewSettingsScreen(
    config SettingsConfig,
    onSave func(SettingsConfig),
    onBack func(),
) fyne.CanvasObject {
    
    // Network settings section
    portEntry := widget.NewEntry()
    portEntry.SetText(fmt.Sprintf("%d", config.Port))
    
    timeoutEntry := widget.NewEntry()
    timeoutEntry.SetText(fmt.Sprintf("%d", config.Timeout))
    
    networkSection := NewSection("Network Settings").
        AddBoldLabel("Port").
        AddItem(portEntry).
        AddBoldLabel("Timeout (seconds)").
        AddItem(timeoutEntry).
        Build()
    
    // Options section
    autoConnectCheck := widget.NewCheck("Auto-connect on startup", nil)
    autoConnectCheck.Checked = config.AutoConnect
    
    optionsSection := NewSection("Options").
        AddItem(autoConnectCheck).
        Build()
    
    // Buttons
    buttons := NewButtonGroup().
        AddPrimaryButton("Save", func() {
            port, _ := strconv.Atoi(portEntry.Text)
            timeout, _ := strconv.Atoi(timeoutEntry.Text)
            onSave(SettingsConfig{
                Port:        port,
                Timeout:     timeout,
                AutoConnect: autoConnectCheck.Checked,
            })
        }).
        AddButton("Cancel", onBack).
        Build()
    
    return NewScreenBuilder("admin:admin - Settings").
        AddSection(networkSection).
        AddSeparator().
        AddSection(optionsSection).
        AddSeparator().
        AddSection(buttons).
        Build()
}
```

---

## Questions?

If you need help with the UI framework, check the existing screens in:
- `role_select.go` - Simple centered layout
- `admin_dashboard.go` - Complex dashboard with forms and state
- `worker_dashboard.go` - System info display

Happy coding! 🚀

