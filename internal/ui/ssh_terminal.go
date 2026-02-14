package ui

import (
	"fmt"
	"image/color"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// SSHTerminal represents a terminal-like SSH interface
type SSHTerminal struct {
	widget.BaseWidget

	mu           sync.RWMutex
	history      []string            // Command history
	historyIndex int                 // Current position in history
	outputLines  []string            // Terminal output buffer
	maxLines     int                 // Maximum lines to keep
	currentDir   string              // Simulated current directory
	hostname     string              // Remote hostname
	connected    bool                // Connection status
	onCommand    func(string) string // Callback to execute commands
}

// NewSSHTerminal creates a new SSH terminal widget
func NewSSHTerminal(hostname string, onCommand func(string) string) *SSHTerminal {
	t := &SSHTerminal{
		history:      make([]string, 0),
		historyIndex: -1,
		outputLines:  make([]string, 0),
		maxLines:     500,
		currentDir:   "~",
		hostname:     hostname,
		connected:    true,
		onCommand:    onCommand,
	}
	t.ExtendBaseWidget(t)

	// Add welcome message
	t.appendOutput(fmt.Sprintf("Connected to %s", hostname))
	t.appendOutput(fmt.Sprintf("SSH session started at %s", time.Now().Format("2006-01-02 15:04:05")))
	t.appendOutput("Type 'help' for available commands, 'exit' to close")
	t.appendOutput("")

	return t
}

func (t *SSHTerminal) appendOutput(line string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.outputLines = append(t.outputLines, line)
	if len(t.outputLines) > t.maxLines {
		t.outputLines = t.outputLines[len(t.outputLines)-t.maxLines:]
	}
}

func (t *SSHTerminal) getOutput() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return strings.Join(t.outputLines, "\n")
}

func (t *SSHTerminal) executeCommand(cmd string) {
	if cmd == "" {
		return
	}

	// Add to history
	t.mu.Lock()
	t.history = append(t.history, cmd)
	t.historyIndex = len(t.history)
	t.mu.Unlock()

	// Show command in output
	prompt := fmt.Sprintf("[%s] $ %s", t.hostname, cmd)
	t.appendOutput(prompt)

	// Handle built-in commands
	switch strings.ToLower(strings.TrimSpace(cmd)) {
	case "clear", "cls":
		t.mu.Lock()
		t.outputLines = []string{}
		t.mu.Unlock()
		return
	case "exit", "quit":
		t.appendOutput("Session closed.")
		t.connected = false
		return
	case "help":
		t.appendOutput("Available commands:")
		t.appendOutput("  clear/cls  - Clear terminal")
		t.appendOutput("  exit/quit  - Close session")
		t.appendOutput("  help       - Show this help")
		t.appendOutput("  Any other command will be executed on the remote system")
		t.appendOutput("")
		return
	}

	// Execute on remote
	if t.onCommand != nil {
		result := t.onCommand(cmd)
		if result != "" {
			// Split result into lines
			lines := strings.Split(result, "\n")
			for _, line := range lines {
				t.appendOutput(line)
			}
		}
	}
	t.appendOutput("")
}

// CreateRenderer implements fyne.Widget
func (t *SSHTerminal) CreateRenderer() fyne.WidgetRenderer {
	return &sshTerminalRenderer{
		terminal: t,
	}
}

type sshTerminalRenderer struct {
	terminal *SSHTerminal
	objects  []fyne.CanvasObject
}

func (r *sshTerminalRenderer) Layout(size fyne.Size)        {}
func (r *sshTerminalRenderer) MinSize() fyne.Size           { return fyne.NewSize(400, 300) }
func (r *sshTerminalRenderer) Refresh()                     {}
func (r *sshTerminalRenderer) Objects() []fyne.CanvasObject { return r.objects }
func (r *sshTerminalRenderer) Destroy()                     {}

// SSHTab represents a single SSH connection tab
type SSHTab struct {
	ID       string
	Hostname string
	IP       string
	Terminal *SSHTerminal
}

// SSHTerminalWindow manages the SSH terminal window with tabs
type SSHTerminalWindow struct {
	window  fyne.Window
	tabs    map[string]*SSHTab
	tabBar  *container.AppTabs
	mu      sync.RWMutex
	onClose func()
}

// NewSSHTerminalWindow creates a new SSH terminal window
func NewSSHTerminalWindow(app fyne.App, onClose func()) *SSHTerminalWindow {
	w := &SSHTerminalWindow{
		tabs:    make(map[string]*SSHTab),
		onClose: onClose,
	}

	w.window = app.NewWindow("admin:admin - SSH Terminal")
	w.window.Resize(fyne.NewSize(800, 500))
	w.window.SetOnClosed(func() {
		if w.onClose != nil {
			w.onClose()
		}
	})

	// Create tabs container
	w.tabBar = container.NewAppTabs()
	w.tabBar.SetTabLocation(container.TabLocationTop)

	// Placeholder when no tabs
	placeholder := container.NewCenter(
		widget.NewLabel("No SSH sessions. Connect to a worker first."),
	)

	w.window.SetContent(container.NewStack(placeholder, w.tabBar))

	return w
}

// AddTab adds a new SSH tab
func (w *SSHTerminalWindow) AddTab(id, hostname, ip string, onCommand func(string) string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Generate unique tab ID with timestamp to allow multiple tabs per worker
	tabID := fmt.Sprintf("%s-%d", id, time.Now().UnixNano())
	sessionNum := 1

	// Count existing sessions for this worker
	for existingID := range w.tabs {
		if strings.HasPrefix(existingID, id+"-") {
			sessionNum++
		}
	}

	// Create display name with session number if multiple
	displayName := hostname
	if sessionNum > 1 {
		displayName = fmt.Sprintf("%s (%d)", hostname, sessionNum)
	}

	// Create terminal content
	terminalContent := createTerminalUI(hostname, ip, onCommand, func() {
		w.RemoveTab(tabID)
	})

	// Create tab
	tab := container.NewTabItem(displayName, terminalContent)
	w.tabBar.Append(tab)
	w.tabBar.Select(tab)

	// Store tab info
	w.tabs[tabID] = &SSHTab{
		ID:       tabID,
		Hostname: displayName,
		IP:       ip,
	}

	// Update window content
	w.window.SetContent(w.tabBar)
}

// RemoveTab removes an SSH tab
func (w *SSHTerminalWindow) RemoveTab(id string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	tab, exists := w.tabs[id]
	if !exists {
		return
	}

	// Find and remove the tab
	for i, item := range w.tabBar.Items {
		if item.Text == tab.Hostname {
			w.tabBar.Remove(item)
			if len(w.tabBar.Items) > 0 && i > 0 {
				w.tabBar.SelectIndex(i - 1)
			}
			break
		}
	}

	delete(w.tabs, id)

	// If no tabs left, show placeholder
	if len(w.tabs) == 0 {
		placeholder := container.NewCenter(
			widget.NewLabel("No SSH sessions. Connect to a worker first."),
		)
		w.window.SetContent(container.NewStack(placeholder, w.tabBar))
	}
}

// Show shows the SSH terminal window
func (w *SSHTerminalWindow) Show() {
	w.window.Show()
}

// Close closes the SSH terminal window
func (w *SSHTerminalWindow) Close() {
	w.window.Close()
}

// HasTab checks if a tab exists
func (w *SSHTerminalWindow) HasTab(id string) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	_, exists := w.tabs[id]
	return exists
}

// createTerminalUI creates the terminal-like UI for a single tab
func createTerminalUI(hostname, ip string, onCommand func(string) string, onClose func()) fyne.CanvasObject {
	// Terminal colors
	bgColor := color.NRGBA{R: 20, G: 15, B: 30, A: 255}       // Dark purple-black
	promptColor := color.NRGBA{R: 138, G: 43, B: 226, A: 255} // Purple

	// Output area - use Entry with MultiLine for copy support
	outputEntry := widget.NewMultiLineEntry()
	outputEntry.Wrapping = fyne.TextWrapWord
	outputEntry.TextStyle = fyne.TextStyle{Monospace: true}

	// Make it read-only but selectable (we'll manage content ourselves)
	outputEntry.Disable()

	outputScroll := container.NewVScroll(outputEntry)
	outputScroll.SetMinSize(fyne.NewSize(780, 380))

	// Command history
	var history []string
	var historyIndex int
	_ = historyIndex // Will be used for up/down arrow navigation in future

	// Output lines storage
	var outputLines []string
	var outputMu sync.Mutex

	// Function to update output display
	updateOutput := func() {
		// Run UI updates on main thread
		if app := fyne.CurrentApp(); app != nil {
			if drv := app.Driver(); drv != nil {
				drv.DoFromGoroutine(func() {
					outputMu.Lock()
					text := strings.Join(outputLines, "\n")
					outputMu.Unlock()

					outputEntry.Enable()
					outputEntry.SetText(text)
					outputEntry.Disable()

					// Scroll to bottom
					outputScroll.ScrollToBottom()
				}, false)
				return
			}
		}
	}

	// Add welcome message
	outputMu.Lock()
	outputLines = append(outputLines, fmt.Sprintf("Connected to %s (%s)", hostname, ip))
	outputLines = append(outputLines, fmt.Sprintf("SSH session started at %s", time.Now().Format("2006-01-02 15:04:05")))
	outputLines = append(outputLines, "Type 'help' for available commands, 'exit' to close")
	outputLines = append(outputLines, "")
	outputMu.Unlock()
	updateOutput()

	// Command input
	cmdEntry := widget.NewEntry()
	cmdEntry.SetPlaceHolder("Enter command...")

	// Execute command function
	executeCmd := func() {
		cmd := strings.TrimSpace(cmdEntry.Text)
		if cmd == "" {
			return
		}

		// Add to history
		history = append(history, cmd)
		historyIndex = len(history)

		// Show command in output
		prompt := fmt.Sprintf("[%s] $ %s", hostname, cmd)
		outputMu.Lock()
		outputLines = append(outputLines, prompt)
		outputMu.Unlock()

		// Clear input
		cmdEntry.SetText("")

		// Handle built-in commands
		switch strings.ToLower(cmd) {
		case "clear", "cls":
			outputMu.Lock()
			outputLines = []string{}
			outputMu.Unlock()
			updateOutput()
			return
		case "exit", "quit":
			outputMu.Lock()
			outputLines = append(outputLines, "Session closed.")
			outputMu.Unlock()
			updateOutput()
			if onClose != nil {
				onClose()
			}
			return
		case "help":
			outputMu.Lock()
			outputLines = append(outputLines, "Available commands:")
			outputLines = append(outputLines, "  clear/cls  - Clear terminal")
			outputLines = append(outputLines, "  exit/quit  - Close session")
			outputLines = append(outputLines, "  help       - Show this help")
			outputLines = append(outputLines, "  Any other command will be executed on the remote system")
			outputLines = append(outputLines, "")
			outputMu.Unlock()
			updateOutput()
			return
		}

		// Execute on remote
		if onCommand != nil {
			go func() {
				result := onCommand(cmd)
				outputMu.Lock()
				if result != "" {
					lines := strings.Split(strings.TrimRight(result, "\n"), "\n")
					outputLines = append(outputLines, lines...)
				}
				outputLines = append(outputLines, "")
				outputMu.Unlock()
				updateOutput()
			}()
		}
	}

	// Handle Enter key
	cmdEntry.OnSubmitted = func(s string) {
		executeCmd()
	}

	// Background
	bg := canvas.NewRectangle(bgColor)

	// Prompt label
	promptLabel := canvas.NewText(fmt.Sprintf("[%s] $", hostname), promptColor)
	promptLabel.TextStyle = fyne.TextStyle{Monospace: true}

	// Input row with prompt
	inputRow := container.NewBorder(nil, nil,
		container.NewHBox(promptLabel),
		nil,
		cmdEntry,
	)

	// Close button
	closeBtn := widget.NewButton("✕ Close Tab", func() {
		if onClose != nil {
			onClose()
		}
	})

	// Header
	headerLabel := widget.NewLabelWithStyle(
		fmt.Sprintf("SSH: %s (%s)", hostname, ip),
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true, Monospace: true},
	)
	header := container.NewBorder(nil, nil, headerLabel, closeBtn)

	// Main layout
	content := container.NewBorder(
		header,
		inputRow,
		nil, nil,
		outputScroll,
	)

	return container.NewStack(bg, container.NewPadded(content))
}

// TerminalEntry is a custom entry with key handling
type TerminalEntry struct {
	widget.Entry
	onKeyUp   func()
	onKeyDown func()
}

func NewTerminalEntry() *TerminalEntry {
	e := &TerminalEntry{}
	e.ExtendBaseWidget(e)
	return e
}

func (e *TerminalEntry) TypedKey(key *fyne.KeyEvent) {
	switch key.Name {
	case fyne.KeyUp:
		if e.onKeyUp != nil {
			e.onKeyUp()
			return
		}
	case fyne.KeyDown:
		if e.onKeyDown != nil {
			e.onKeyDown()
			return
		}
	}
	e.Entry.TypedKey(key)
}

func (e *TerminalEntry) KeyDown(key *fyne.KeyEvent) {
	if key.Name == fyne.KeyReturn || key.Name == fyne.KeyEnter {
		if e.OnSubmitted != nil {
			e.OnSubmitted(e.Text)
		}
		return
	}
	e.Entry.KeyDown(key)
}

// Ensure TerminalEntry implements desktop.Keyable
var _ desktop.Keyable = (*TerminalEntry)(nil)
