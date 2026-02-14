package ui

import (
	"fmt"
	"image/color"
	"regexp"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// Terminal color scheme (purple theme)
var (
	termBgColor      = color.NRGBA{R: 18, G: 12, B: 28, A: 255}    // Deep purple-black
	termFgColor      = color.NRGBA{R: 220, G: 215, B: 230, A: 255} // Light purple-white
	termPromptColor  = color.NRGBA{R: 180, G: 100, B: 255, A: 255} // Bright purple
	termAccentColor  = color.NRGBA{R: 138, G: 43, B: 226, A: 255}  // BlueViolet
	termSuccessColor = color.NRGBA{R: 100, G: 220, B: 150, A: 255} // Green
	termErrorColor   = color.NRGBA{R: 255, G: 100, B: 120, A: 255} // Red
	termWarningColor = color.NRGBA{R: 255, G: 200, B: 100, A: 255} // Yellow
	termInfoColor    = color.NRGBA{R: 100, G: 180, B: 255, A: 255} // Blue
	termDimColor     = color.NRGBA{R: 120, G: 110, B: 140, A: 255} // Dim purple
	termBorderColor  = color.NRGBA{R: 80, G: 50, B: 120, A: 255}   // Purple border
	termInputBg      = color.NRGBA{R: 30, G: 20, B: 45, A: 255}    // Slightly lighter input bg
	termHeaderBg     = color.NRGBA{R: 45, G: 30, B: 65, A: 255}    // Header background
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

	// If no tabs left, close the window
	if len(w.tabs) == 0 {
		w.window.Close()
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
	// Output area - custom rich text display with color support
	outputText := widget.NewRichText()
	outputText.Wrapping = fyne.TextWrapWord

	outputScroll := container.NewVScroll(outputText)
	outputScroll.SetMinSize(fyne.NewSize(750, 350))

	// Command history
	var history []string
	historyIndex := 0
	_ = historyIndex // Used in future for up/down arrow navigation

	// Output lines storage (raw text with ANSI codes stripped for copy)
	var outputSegments []widget.RichTextSegment
	var outputMu sync.Mutex

	// Function to update output display
	updateOutput := func() {
		if app := fyne.CurrentApp(); app != nil {
			if drv := app.Driver(); drv != nil {
				drv.DoFromGoroutine(func() {
					outputMu.Lock()
					segs := make([]widget.RichTextSegment, len(outputSegments))
					copy(segs, outputSegments)
					outputMu.Unlock()

					outputText.Segments = segs
					outputText.Refresh()
					outputScroll.ScrollToBottom()
				}, false)
				return
			}
		}
	}

	// Helper to add line with proper styling
	addLine := func(text string, isPrompt bool, isError bool) {
		outputMu.Lock()
		defer outputMu.Unlock()

		if isPrompt {
			// Prompt line: [hostname] $ command
			outputSegments = append(outputSegments, &widget.TextSegment{
				Text: "[",
				Style: widget.RichTextStyle{
					Inline:    true,
					TextStyle: fyne.TextStyle{Monospace: true},
				},
			})
			outputSegments = append(outputSegments, &widget.TextSegment{
				Text: hostname,
				Style: widget.RichTextStyle{
					Inline:    true,
					TextStyle: fyne.TextStyle{Monospace: true, Bold: true},
				},
			})
			outputSegments = append(outputSegments, &widget.TextSegment{
				Text: "] $ ",
				Style: widget.RichTextStyle{
					Inline:    true,
					TextStyle: fyne.TextStyle{Monospace: true},
				},
			})
			// The actual command
			parts := strings.SplitN(text, "$ ", 2)
			if len(parts) > 1 {
				outputSegments = append(outputSegments, &widget.TextSegment{
					Text: parts[1] + "\n",
					Style: widget.RichTextStyle{
						Inline:    true,
						TextStyle: fyne.TextStyle{Monospace: true, Bold: true},
					},
				})
			} else {
				outputSegments = append(outputSegments, &widget.TextSegment{
					Text:  "\n",
					Style: widget.RichTextStyle{Inline: true},
				})
			}
		} else {
			// Regular output line
			style := fyne.TextStyle{Monospace: true}
			if isError {
				style.Bold = true
			}
			outputSegments = append(outputSegments, &widget.TextSegment{
				Text: text + "\n",
				Style: widget.RichTextStyle{
					Inline:    true,
					TextStyle: style,
				},
			})
		}
	}

	// Add welcome message with styling
	addLine("╔══════════════════════════════════════════════════════════════╗", false, false)
	addLine(fmt.Sprintf("║  SSH Session: %s", padRight(hostname+" ("+ip+")", 48)+"║"), false, false)
	addLine(fmt.Sprintf("║  Started: %s", padRight(time.Now().Format("2006-01-02 15:04:05"), 51)+"║"), false, false)
	addLine("╠══════════════════════════════════════════════════════════════╣", false, false)
	addLine("║  Commands: help, clear, exit                                 ║", false, false)
	addLine("╚══════════════════════════════════════════════════════════════╝", false, false)
	addLine("", false, false)
	updateOutput()

	// Command input - custom styled
	cmdEntry := widget.NewEntry()
	cmdEntry.SetPlaceHolder("Type command and press Enter...")

	// Execute command function
	executeCmd := func() {
		cmd := strings.TrimSpace(cmdEntry.Text)
		if cmd == "" {
			return
		}

		// Add to history
		history = append(history, cmd)
		historyIndex = len(history)

		// Show command in output with prompt styling
		addLine(fmt.Sprintf("[%s] $ %s", hostname, cmd), true, false)

		// Clear input
		cmdEntry.SetText("")

		// Handle built-in commands
		switch strings.ToLower(cmd) {
		case "clear", "cls":
			outputMu.Lock()
			outputSegments = []widget.RichTextSegment{}
			outputMu.Unlock()
			updateOutput()
			return
		case "exit", "quit":
			addLine("Session terminated.", false, false)
			updateOutput()
			if onClose != nil {
				time.AfterFunc(500*time.Millisecond, onClose)
			}
			return
		case "help":
			addLine("┌─────────────────────────────────────┐", false, false)
			addLine("│         Available Commands          │", false, false)
			addLine("├─────────────────────────────────────┤", false, false)
			addLine("│  clear, cls  │ Clear terminal       │", false, false)
			addLine("│  exit, quit  │ Close session        │", false, false)
			addLine("│  help        │ Show this help       │", false, false)
			addLine("├─────────────────────────────────────┤", false, false)
			addLine("│  Other commands run on remote host  │", false, false)
			addLine("└─────────────────────────────────────┘", false, false)
			addLine("", false, false)
			updateOutput()
			return
		}

		// Execute on remote
		if onCommand != nil {
			// Show loading indicator
			addLine("Executing...", false, false)
			updateOutput()

			go func() {
				result := onCommand(cmd)

				// Remove "Executing..." line
				outputMu.Lock()
				if len(outputSegments) > 0 {
					outputSegments = outputSegments[:len(outputSegments)-1]
				}
				outputMu.Unlock()

				if result != "" {
					result = stripANSI(result)
					lines := strings.Split(strings.TrimRight(result, "\n\r"), "\n")
					for _, line := range lines {
						isErr := strings.Contains(strings.ToLower(line), "error") ||
							strings.Contains(strings.ToLower(line), "failed")
						addLine(line, false, isErr)
					}
				}
				addLine("", false, false)
				updateOutput()
			}()
		}
	}

	// Handle Enter key
	cmdEntry.OnSubmitted = func(s string) {
		executeCmd()
	}

	// Background
	bg := canvas.NewRectangle(termBgColor)

	// Input background
	inputBg := canvas.NewRectangle(termInputBg)
	inputBg.CornerRadius = 4

	// Prompt label with purple color
	promptLabel := canvas.NewText(fmt.Sprintf("[%s] $", hostname), termPromptColor)
	promptLabel.TextStyle = fyne.TextStyle{Monospace: true, Bold: true}
	promptLabel.TextSize = 14

	// Input row with styled background
	inputRow := container.NewBorder(nil, nil,
		container.NewPadded(promptLabel),
		nil,
		cmdEntry,
	)

	inputContainer := container.NewStack(
		inputBg,
		container.NewPadded(inputRow),
	)

	// Close button - styled
	closeBtn := widget.NewButton("✕ Close", func() {
		if onClose != nil {
			onClose()
		}
	})
	closeBtn.Importance = widget.DangerImportance

	// Header with gradient-like effect
	headerBg := canvas.NewRectangle(termHeaderBg)
	headerBg.CornerRadius = 6

	headerIcon := canvas.NewText("⌘", termPromptColor)
	headerIcon.TextSize = 18
	headerIcon.TextStyle = fyne.TextStyle{Bold: true}

	headerTitle := canvas.NewText(fmt.Sprintf("SSH: %s", hostname), termFgColor)
	headerTitle.TextSize = 14
	headerTitle.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}

	headerIP := canvas.NewText(ip, termDimColor)
	headerIP.TextSize = 12
	headerIP.TextStyle = fyne.TextStyle{Monospace: true}

	headerLeft := container.NewHBox(
		headerIcon,
		widget.NewSeparator(),
		headerTitle,
		headerIP,
	)

	headerRight := container.NewHBox(
		closeBtn,
	)

	headerContent := container.NewBorder(nil, nil, headerLeft, headerRight)
	header := container.NewStack(headerBg, container.NewPadded(headerContent))

	// Output container with border effect
	outputBorder := canvas.NewRectangle(termBorderColor)
	outputBorder.CornerRadius = 4
	outputBorder.StrokeWidth = 1
	outputBorder.StrokeColor = termBorderColor

	outputContainer := container.NewStack(
		outputBorder,
		container.NewPadded(outputScroll),
	)

	// Main layout with spacing
	mainContent := container.NewBorder(
		container.NewVBox(header, widget.NewSeparator()),
		container.NewVBox(widget.NewSeparator(), inputContainer),
		nil, nil,
		outputContainer,
	)

	return container.NewStack(bg, container.NewPadded(mainContent))
}

// stripANSI removes ANSI escape codes from a string
func stripANSI(str string) string {
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]|\x1b\][^\x07]*\x07`)
	return ansiRegex.ReplaceAllString(str, "")
}

// padRight pads a string to the right with spaces
func padRight(s string, length int) string {
	if len(s) >= length {
		return s[:length]
	}
	return s + strings.Repeat(" ", length-len(s))
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
