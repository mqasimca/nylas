// Package models provides screen models for the TUI.
package models

import (
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/log"
	"github.com/mqasimca/nylas/internal/tui2/state"
	"github.com/mqasimca/nylas/internal/tui2/styles"
)

// LogLevel represents the severity of a log entry.
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

// LogEntry represents a single log entry.
type LogEntry struct {
	Time    time.Time
	Level   LogLevel
	Message string
	Context map[string]any
}

// DebugScreen displays application logs and debugging information.
type DebugScreen struct {
	global   *state.GlobalState
	theme    *styles.Theme
	viewport viewport.Model
	logs     []LogEntry
	logger   *log.Logger
	ready    bool
	width    int
	height   int
}

// NewDebugScreen creates a new debug screen.
func NewDebugScreen(global *state.GlobalState) *DebugScreen {
	theme := styles.GetTheme(global.Theme)

	// Create a custom logger that captures logs in memory
	logger := log.NewWithOptions(log.StandardLog().Writer(), log.Options{
		ReportCaller:    true,
		ReportTimestamp: true,
		TimeFormat:      time.Kitchen,
	})

	// Set level to debug
	logger.SetLevel(log.DebugLevel)

	d := &DebugScreen{
		global: global,
		theme:  theme,
		logger: logger,
		logs:   make([]LogEntry, 0),
	}

	// Add some initial logs
	d.addLog(LogLevelInfo, "Debug panel initialized")
	d.addLog(LogLevelDebug, fmt.Sprintf("Theme: %s", global.Theme))
	d.addLog(LogLevelDebug, fmt.Sprintf("Grant ID: %s", global.GrantID))
	d.addLog(LogLevelDebug, fmt.Sprintf("Email: %s", global.Email))

	return d
}

// Init implements tea.Model.
func (d *DebugScreen) Init() tea.Cmd {
	// Initialize viewport
	if d.global.WindowSize.Width > 0 && d.global.WindowSize.Height > 0 {
		d.width = d.global.WindowSize.Width
		d.height = d.global.WindowSize.Height
		d.viewport = viewport.New()
		d.viewport.SetWidth(d.width)
		d.viewport.SetHeight(d.height - 6)
		d.ready = true
		d.updateViewport()
	}

	return nil
}

// Update implements tea.Model.
func (d *DebugScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.global.SetWindowSize(msg.Width, msg.Height)
		d.width = msg.Width
		d.height = msg.Height

		if !d.ready {
			d.viewport = viewport.New()
			d.viewport.SetWidth(msg.Width)
			d.viewport.SetHeight(msg.Height - 6)
			d.ready = true
		} else {
			d.viewport.SetWidth(msg.Width)
			d.viewport.SetHeight(msg.Height - 6)
		}
		d.updateViewport()
		return d, nil

	case tea.KeyMsg:
		key := msg.Key()
		keyStr := msg.String()

		// Handle Esc key
		if key.Code == tea.KeyEsc {
			return d, func() tea.Msg { return BackMsg{} }
		}

		// Handle ctrl+c
		if keyStr == "ctrl+c" {
			return d, tea.Quit
		}

		// Handle 'c' to clear logs
		if keyStr == "c" {
			d.logs = make([]LogEntry, 0)
			d.addLog(LogLevelInfo, "Logs cleared")
			d.updateViewport()
			return d, nil
		}

		// Handle 't' to add test log entries
		if keyStr == "t" {
			d.addTestLogs()
			d.updateViewport()
			return d, nil
		}
	}

	// Update viewport
	if d.ready {
		d.viewport, cmd = d.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return d, tea.Batch(cmds...)
}

// View implements tea.Model.
func (d *DebugScreen) View() tea.View {
	if !d.ready {
		return tea.NewView("Initializing debug panel...")
	}

	var sections []string

	// Header
	header := d.theme.Title.Render("Debug Panel")
	logCount := d.theme.Subtitle.Render(fmt.Sprintf(" (%d logs)", len(d.logs)))
	sections = append(sections, header+logCount)
	sections = append(sections, "")

	// Viewport content
	sections = append(sections, d.viewport.View())
	sections = append(sections, "")

	// Help text
	help := d.theme.Help.Render("↑/↓: scroll • c: clear logs • t: add test logs • Esc: back")
	sections = append(sections, help)

	return tea.NewView(lipgloss.JoinVertical(lipgloss.Left, sections...))
}

// addLog adds a log entry.
func (d *DebugScreen) addLog(level LogLevel, message string) {
	entry := LogEntry{
		Time:    time.Now(),
		Level:   level,
		Message: message,
	}
	d.logs = append(d.logs, entry)
}

// addTestLogs adds some test log entries for demonstration.
func (d *DebugScreen) addTestLogs() {
	d.addLog(LogLevelDebug, "Debug message: checking cache")
	d.addLog(LogLevelInfo, "Info message: cache hit for key 'threads'")
	d.addLog(LogLevelWarn, "Warning message: rate limit approaching (80%)")
	d.addLog(LogLevelError, "Error message: failed to connect to API")
	d.addLog(LogLevelInfo, "Info message: retrying with backoff")
	d.addLog(LogLevelDebug, "Debug message: successfully reconnected")
}

// updateViewport updates the viewport content with all logs.
func (d *DebugScreen) updateViewport() {
	var content strings.Builder

	// System information
	content.WriteString(d.theme.Subtitle.Render("System Information"))
	content.WriteString("\n")
	content.WriteString(d.renderSystemInfo())
	content.WriteString("\n\n")

	// Log entries
	content.WriteString(d.theme.Subtitle.Render("Application Logs"))
	content.WriteString("\n\n")

	if len(d.logs) == 0 {
		content.WriteString(d.theme.Dimmed.Render("No logs yet. Press 't' to add test logs."))
	} else {
		// Show most recent logs first
		for i := len(d.logs) - 1; i >= 0; i-- {
			entry := d.logs[i]
			content.WriteString(d.formatLogEntry(entry))
			content.WriteString("\n")
		}
	}

	d.viewport.SetContent(content.String())
}

// renderSystemInfo renders system information.
func (d *DebugScreen) renderSystemInfo() string {
	var sb strings.Builder

	// Window size
	sb.WriteString(fmt.Sprintf("Window Size:  %dx%d\n", d.global.WindowSize.Width, d.global.WindowSize.Height))

	// Theme
	sb.WriteString(fmt.Sprintf("Theme:        %s\n", d.global.Theme))

	// Account
	sb.WriteString(fmt.Sprintf("Email:        %s\n", d.global.Email))
	sb.WriteString(fmt.Sprintf("Provider:     %s\n", d.global.Provider))
	sb.WriteString(fmt.Sprintf("Grant ID:     %s\n", truncateID(d.global.GrantID)))

	return sb.String()
}

// formatLogEntry formats a log entry for display.
func (d *DebugScreen) formatLogEntry(entry LogEntry) string {
	// Format timestamp
	timestamp := entry.Time.Format("15:04:05")
	timestampStyle := lipgloss.NewStyle().Foreground(d.theme.Dimmed.GetForeground())

	// Format level with color
	var levelStr string
	var levelStyle lipgloss.Style

	switch entry.Level {
	case LogLevelDebug:
		levelStr = "DEBUG"
		levelStyle = lipgloss.NewStyle().Foreground(d.theme.Info)
	case LogLevelInfo:
		levelStr = "INFO "
		levelStyle = lipgloss.NewStyle().Foreground(d.theme.Success)
	case LogLevelWarn:
		levelStr = "WARN "
		levelStyle = lipgloss.NewStyle().Foreground(d.theme.Warning)
	case LogLevelError:
		levelStr = "ERROR"
		levelStyle = lipgloss.NewStyle().Foreground(d.theme.Error).Bold(true)
	}

	// Format message
	messageStyle := lipgloss.NewStyle().Foreground(d.theme.Foreground)

	return fmt.Sprintf("%s [%s] %s",
		timestampStyle.Render(timestamp),
		levelStyle.Render(levelStr),
		messageStyle.Render(entry.Message),
	)
}

// truncateID truncates a long ID for display.
func truncateID(id string) string {
	if len(id) <= 20 {
		return id
	}
	return id[:8] + "..." + id[len(id)-8:]
}
