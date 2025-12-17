// Package browser provides browser opening functionality.
package browser

import (
	"github.com/pkg/browser"
)

// DefaultBrowser opens URLs in the system default browser.
type DefaultBrowser struct{}

// NewDefaultBrowser creates a new DefaultBrowser.
func NewDefaultBrowser() *DefaultBrowser {
	return &DefaultBrowser{}
}

// Open opens a URL in the default browser.
func (b *DefaultBrowser) Open(url string) error {
	return browser.OpenURL(url)
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
