// Package browser provides browser opening functionality.
package browser

import (
	"os/exec"
	"runtime"
	"syscall"
)

// DefaultBrowser opens URLs in the system default browser.
type DefaultBrowser struct{}

// NewDefaultBrowser creates a new DefaultBrowser.
func NewDefaultBrowser() *DefaultBrowser {
	return &DefaultBrowser{}
}

// Open opens a URL in the default browser.
// On Linux, it ensures the browser is started in its own process group
// so that Ctrl+C doesn't kill the browser when stopping the CLI.
func (b *DefaultBrowser) Open(url string) error {
	return openURL(url)
}

// openURL opens a URL in the default browser with proper process isolation.
func openURL(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		// Use xdg-open on Linux
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		// Use open on macOS
		cmd = exec.Command("open", url)
	case "windows":
		// Use start on Windows
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		// Fallback to xdg-open
		cmd = exec.Command("xdg-open", url)
	}

	// On Unix systems, start the browser in its own process group.
	// This prevents SIGINT (Ctrl+C) from propagating to the browser
	// when the user stops the CLI.
	if runtime.GOOS != "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setpgid: true,
		}
	}

	return cmd.Start()
}

// MockBrowser is a mock implementation for testing.
type MockBrowser struct {
	OpenCalled bool
	LastURL    string
	OpenFunc   func(url string) error
}

// NewMockBrowser creates a new MockBrowser.
func NewMockBrowser() *MockBrowser {
	return &MockBrowser{}
}

// Open records the call and optionally calls the custom function.
func (m *MockBrowser) Open(url string) error {
	m.OpenCalled = true
	m.LastURL = url
	if m.OpenFunc != nil {
		return m.OpenFunc(url)
	}
	return nil
}
