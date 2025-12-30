// Package components provides reusable UI components.
package components

import (
	"strings"
	"testing"

	"github.com/mqasimca/nylas/internal/tui2/styles"
)

func TestNewStatusBar(t *testing.T) {
	theme := styles.GetTheme("k9s")
	email := "test@example.com"

	statusBar := NewStatusBar(theme, email)

	if statusBar == nil {
		t.Fatal("NewStatusBar returned nil")
	}

	if statusBar.email != email {
		t.Errorf("email = %q, want %q", statusBar.email, email)
	}

	if statusBar.theme != theme {
		t.Error("theme not set correctly")
	}

	if !statusBar.online {
		t.Error("expected online to be true by default")
	}
}

func TestStatusBar_SetWidth(t *testing.T) {
	theme := styles.GetTheme("k9s")
	statusBar := NewStatusBar(theme, "test@example.com")

	statusBar.SetWidth(100)

	if statusBar.width != 100 {
		t.Errorf("width = %d, want 100", statusBar.width)
	}
}

func TestStatusBar_SetOnline(t *testing.T) {
	theme := styles.GetTheme("k9s")
	statusBar := NewStatusBar(theme, "test@example.com")

	statusBar.SetOnline(false)

	if statusBar.online {
		t.Error("expected online to be false")
	}
}

func TestStatusBar_SetCounts(t *testing.T) {
	theme := styles.GetTheme("k9s")
	statusBar := NewStatusBar(theme, "test@example.com")

	statusBar.SetUnreadCount(42)
	statusBar.SetEventCount(5)

	if statusBar.unreadCount != 42 {
		t.Errorf("unreadCount = %d, want 42", statusBar.unreadCount)
	}

	if statusBar.eventCount != 5 {
		t.Errorf("eventCount = %d, want 5", statusBar.eventCount)
	}
}

func TestStatusBar_View(t *testing.T) {
	theme := styles.GetTheme("k9s")
	statusBar := NewStatusBar(theme, "test@example.com")
	statusBar.SetWidth(200)
	statusBar.SetUnreadCount(10)
	statusBar.SetEventCount(3)

	view := statusBar.View()

	if view == "" {
		t.Error("View returned empty string")
	}

	// k9s style: should contain email
	if !strings.Contains(view, "test@example.com") {
		t.Error("View should contain email")
	}
}

func TestFooterBar_SetBindings(t *testing.T) {
	theme := styles.GetTheme("k9s")
	footer := NewFooterBar(theme, "test@example.com")

	bindings := []KeyBinding{
		{Key: "a", Description: "Action"},
		{Key: "b", Description: "Button"},
	}

	footer.SetBindings(bindings)

	if len(footer.bindings) != 2 {
		t.Errorf("bindings length = %d, want 2", len(footer.bindings))
	}
}

func TestFooterBar_View(t *testing.T) {
	theme := styles.GetTheme("k9s")
	footer := NewFooterBar(theme, "test@example.com")
	footer.SetWidth(100)

	view := footer.View()

	if view == "" {
		t.Error("View returned empty string")
	}

	// k9s style footer: :command | ?:help | ^c:quit
	if !strings.Contains(view, "command") {
		t.Error("View should contain 'command'")
	}
}

// k9s-style Tests

func TestStatusBar_K9sStyle(t *testing.T) {
	theme := styles.GetTheme("k9s")
	statusBar := NewStatusBar(theme, "test@example.com")
	statusBar.SetWidth(200)

	view := statusBar.View()

	// Should contain email
	if !strings.Contains(view, "test@example.com") {
		t.Error("Status bar should contain email")
	}

	// Should contain pipe separator
	if !strings.Contains(view, "│") {
		t.Error("Status bar should contain pipe separator")
	}
}

func TestStatusBar_LiveStatus(t *testing.T) {
	theme := styles.GetTheme("k9s")

	t.Run("online status", func(t *testing.T) {
		statusBar := NewStatusBar(theme, "test@example.com")
		statusBar.SetWidth(200)
		statusBar.SetOnline(true)

		view := statusBar.View()

		// Should contain status dot
		if !strings.Contains(view, "●") {
			t.Error("Status bar should contain status dot")
		}

		// Should contain Live text
		if !strings.Contains(view, "Live") {
			t.Error("Status bar should display Live status")
		}
	})

	t.Run("offline status", func(t *testing.T) {
		statusBar := NewStatusBar(theme, "test@example.com")
		statusBar.SetWidth(200)
		statusBar.SetOnline(false)

		view := statusBar.View()

		// Should contain Offline text
		if !strings.Contains(view, "Offline") {
			t.Error("Status bar should display Offline status")
		}
	})
}

func TestStatusBar_Time(t *testing.T) {
	theme := styles.GetTheme("k9s")
	statusBar := NewStatusBar(theme, "test@example.com")
	statusBar.SetWidth(200)
	statusBar.Update() // Update time

	view := statusBar.View()

	// Should not be empty
	if view == "" {
		t.Error("Status bar view should not be empty")
	}
}

func TestStatusBar_ZeroWidth(t *testing.T) {
	theme := styles.GetTheme("k9s")
	statusBar := NewStatusBar(theme, "test@example.com")
	statusBar.SetWidth(0)

	view := statusBar.View()

	// Should return empty string when width is 0
	if view != "" {
		t.Error("Status bar should return empty string when width is 0")
	}
}

func TestFooterBar_K9sStyle(t *testing.T) {
	theme := styles.GetTheme("k9s")
	footer := NewFooterBar(theme, "test@example.com")
	footer.SetWidth(150)

	view := footer.View()

	// k9s style: :command | ?:help | ^c:quit
	if !strings.Contains(view, "command") {
		t.Error("Footer should contain 'command'")
	}

	if !strings.Contains(view, "help") {
		t.Error("Footer should contain 'help'")
	}

	if !strings.Contains(view, "quit") {
		t.Error("Footer should contain 'quit'")
	}

	// Should contain pipe separators
	if !strings.Contains(view, "│") {
		t.Error("Footer should contain pipe separator")
	}
}

func TestFooterBar_ZeroWidth(t *testing.T) {
	theme := styles.GetTheme("k9s")
	footer := NewFooterBar(theme, "test@example.com")
	footer.SetWidth(0)

	view := footer.View()

	// Should return empty string when width is 0
	if view != "" {
		t.Error("Footer should return empty string when width is 0")
	}
}
