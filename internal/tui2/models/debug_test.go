// Package models provides screen models for the TUI.
package models

import (
	"strings"
	"testing"

	"github.com/mqasimca/nylas/internal/tui2/state"
)

func TestNewDebugScreen(t *testing.T) {
	global := state.NewGlobalState(nil, nil, "test-grant-id", "test@example.com", "gmail")
	global.Theme = "k9s"

	d := NewDebugScreen(global)

	if d == nil {
		t.Fatal("NewDebugScreen returned nil")
	}

	if d.global != global {
		t.Error("global state not set correctly")
	}

	if d.theme == nil {
		t.Error("theme not initialized")
	}

	if d.logger == nil {
		t.Error("logger not initialized")
	}

	// Check that initial logs were added
	if len(d.logs) == 0 {
		t.Error("should have initial logs")
	}

	// Verify initial log entries
	foundInitLog := false
	for _, log := range d.logs {
		if strings.Contains(log.Message, "Debug panel initialized") {
			foundInitLog = true
			break
		}
	}
	if !foundInitLog {
		t.Error("should have initialization log")
	}
}

func TestDebugAddLog(t *testing.T) {
	global := state.NewGlobalState(nil, nil, "test-grant", "test@example.com", "gmail")
	global.Theme = "k9s"

	d := NewDebugScreen(global)
	initialCount := len(d.logs)

	// Add a debug log
	d.addLog(LogLevelDebug, "Test debug message")

	if len(d.logs) != initialCount+1 {
		t.Errorf("expected %d logs, got %d", initialCount+1, len(d.logs))
	}

	lastLog := d.logs[len(d.logs)-1]
	if lastLog.Message != "Test debug message" {
		t.Errorf("log message = %q, want %q", lastLog.Message, "Test debug message")
	}

	if lastLog.Level != LogLevelDebug {
		t.Errorf("log level = %v, want %v", lastLog.Level, LogLevelDebug)
	}
}

func TestDebugAddTestLogs(t *testing.T) {
	global := state.NewGlobalState(nil, nil, "test-grant", "test@example.com", "gmail")
	global.Theme = "k9s"

	d := NewDebugScreen(global)
	initialCount := len(d.logs)

	// Add test logs
	d.addTestLogs()

	if len(d.logs) <= initialCount {
		t.Error("addTestLogs should add logs")
	}

	// Verify different log levels are present
	var hasDebug, hasInfo, hasWarn, hasError bool
	for _, log := range d.logs {
		switch log.Level {
		case LogLevelDebug:
			hasDebug = true
		case LogLevelInfo:
			hasInfo = true
		case LogLevelWarn:
			hasWarn = true
		case LogLevelError:
			hasError = true
		}
	}

	if !hasDebug || !hasInfo || !hasWarn || !hasError {
		t.Error("addTestLogs should add logs of all levels")
	}
}

func TestDebugFormatLogEntry(t *testing.T) {
	global := state.NewGlobalState(nil, nil, "test-grant", "test@example.com", "gmail")
	global.Theme = "k9s"

	d := NewDebugScreen(global)

	tests := []struct {
		name  string
		level LogLevel
	}{
		{"debug level", LogLevelDebug},
		{"info level", LogLevelInfo},
		{"warn level", LogLevelWarn},
		{"error level", LogLevelError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := LogEntry{
				Level:   tt.level,
				Message: "Test message",
			}

			formatted := d.formatLogEntry(entry)
			if formatted == "" {
				t.Error("formatLogEntry should return non-empty string")
			}

			// Should contain the message
			if !strings.Contains(formatted, "Test message") {
				t.Error("formatted entry should contain message")
			}
		})
	}
}

func TestDebugRenderSystemInfo(t *testing.T) {
	global := state.NewGlobalState(nil, nil, "test-grant-id-123", "test@example.com", "gmail")
	global.Theme = "cyberpunk"
	global.SetWindowSize(100, 50)

	d := NewDebugScreen(global)

	info := d.renderSystemInfo()

	if info == "" {
		t.Error("renderSystemInfo should return non-empty string")
	}

	// Check for expected content
	expectedContent := []string{
		"Window Size:",
		"Theme:",
		"Email:",
		"Provider:",
		"Grant ID:",
	}

	for _, expected := range expectedContent {
		if !strings.Contains(info, expected) {
			t.Errorf("system info should contain %q", expected)
		}
	}
}

func TestTruncateID(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want string
	}{
		{
			name: "short ID",
			id:   "short",
			want: "short",
		},
		{
			name: "exactly 20 chars",
			id:   "12345678901234567890",
			want: "12345678901234567890",
		},
		{
			name: "long ID",
			id:   "this-is-a-very-long-grant-id-that-needs-truncation",
			want: "this-is-...uncation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateID(tt.id)
			if got != tt.want {
				t.Errorf("truncateID(%q) = %q, want %q", tt.id, got, tt.want)
			}
		})
	}
}
