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
	statusBar.SetWidth(100)
	statusBar.SetUnreadCount(10)
	statusBar.SetEventCount(3)

	view := statusBar.View()

	if view == "" {
		t.Error("View returned empty string")
	}

	// Check that view contains expected elements
	if !strings.Contains(view, "Nylas CLI") {
		t.Error("View should contain app name")
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
	footer.SetBindings([]KeyBinding{
		{Key: "a", Description: "Action"},
	})

	view := footer.View()

	if view == "" {
		t.Error("View returned empty string")
	}

	if !strings.Contains(view, "test@example.com") {
		t.Error("View should contain email")
	}
}

// Glossy Enhancement Tests

func TestStatusBar_GlossyAppInfo(t *testing.T) {
	theme := styles.GetTheme("k9s")
	statusBar := NewStatusBar(theme, "test@example.com")
	statusBar.SetWidth(150)

	view := statusBar.View()

	// Should contain sparkle emoji
	if !strings.Contains(view, "✨") {
		t.Error("Status bar should contain sparkle emoji for glossy effect")
	}

	// Should contain app name and version
	if !strings.Contains(view, "Nylas CLI") {
		t.Error("Status bar should contain app name")
	}

	if !strings.Contains(view, "v2.0") {
		t.Error("Status bar should contain version")
	}
}

func TestStatusBar_GlossyBadges(t *testing.T) {
	theme := styles.GetTheme("k9s")
	statusBar := NewStatusBar(theme, "test@example.com")
	statusBar.SetWidth(150)
	statusBar.SetUnreadCount(42)
	statusBar.SetEventCount(7)

	view := statusBar.View()

	// Should contain inbox emoji
	if !strings.Contains(view, "📥") {
		t.Error("Status bar should contain inbox emoji")
	}

	// Should contain calendar emoji
	if !strings.Contains(view, "📅") {
		t.Error("Status bar should contain calendar emoji")
	}

	// Should contain counts (though they may be ANSI-styled)
	if !strings.Contains(view, "42") {
		t.Error("Status bar should display unread count")
	}

	if !strings.Contains(view, "7") {
		t.Error("Status bar should display event count")
	}
}

func TestStatusBar_GlossyStatus(t *testing.T) {
	theme := styles.GetTheme("k9s")

	t.Run("online status", func(t *testing.T) {
		statusBar := NewStatusBar(theme, "test@example.com")
		statusBar.SetWidth(150)
		statusBar.SetOnline(true)

		view := statusBar.View()

		// Should contain status dot
		if !strings.Contains(view, "●") {
			t.Error("Status bar should contain status dot")
		}

		// Should contain ONLINE text
		if !strings.Contains(view, "ONLINE") {
			t.Error("Status bar should display ONLINE status")
		}
	})

	t.Run("offline status", func(t *testing.T) {
		statusBar := NewStatusBar(theme, "test@example.com")
		statusBar.SetWidth(150)
		statusBar.SetOnline(false)

		view := statusBar.View()

		// Should contain OFFLINE text
		if !strings.Contains(view, "OFFLINE") {
			t.Error("Status bar should display OFFLINE status")
		}
	})
}

func TestStatusBar_GlossyClock(t *testing.T) {
	theme := styles.GetTheme("k9s")
	statusBar := NewStatusBar(theme, "test@example.com")
	statusBar.SetWidth(150)
	statusBar.Update() // Update time

	view := statusBar.View()

	// Should contain clock emoji
	if !strings.Contains(view, "⏰") {
		t.Error("Status bar should contain clock emoji")
	}

	// Should contain time (HH:MM:SS format)
	// We can't check exact time, but we can verify it's not empty
	if view == "" {
		t.Error("Status bar view should not be empty")
	}
}

func TestStatusBar_GlossySeparators(t *testing.T) {
	theme := styles.GetTheme("k9s")
	statusBar := NewStatusBar(theme, "test@example.com")
	statusBar.SetWidth(150)

	view := statusBar.View()

	// Should contain diamond separator
	if !strings.Contains(view, "◆") {
		t.Error("Status bar should contain diamond separator for glossy effect")
	}
}

func TestStatusBar_MinimalView(t *testing.T) {
	theme := styles.GetTheme("k9s")
	statusBar := NewStatusBar(theme, "test@example.com")
	statusBar.SetWidth(20) // Very small width

	view := statusBar.View()

	// Should still return content (minimal version)
	if view == "" {
		t.Error("Status bar should show minimal version when width is small")
	}

	// Should at least contain app name in minimal version
	if !strings.Contains(view, "Nylas CLI") {
		t.Error("Minimal status bar should contain app name")
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

func TestFooterBar_GlossyBadges(t *testing.T) {
	theme := styles.GetTheme("k9s")
	footer := NewFooterBar(theme, "test@example.com")
	footer.SetWidth(150)
	footer.SetBindings([]KeyBinding{
		{Key: "a", Description: "Air"},
		{Key: "c", Description: "Calendar"},
		{Key: "p", Description: "People"},
	})

	view := footer.View()

	// Should contain all keys
	if !strings.Contains(view, "Air") {
		t.Error("Footer should contain Air binding")
	}

	if !strings.Contains(view, "Calendar") {
		t.Error("Footer should contain Calendar binding")
	}

	if !strings.Contains(view, "People") {
		t.Error("Footer should contain People binding")
	}

	// Should contain bullet separators
	if !strings.Contains(view, "•") {
		t.Error("Footer should contain bullet separator")
	}
}

func TestFooterBar_GlossyUserBadge(t *testing.T) {
	theme := styles.GetTheme("k9s")
	footer := NewFooterBar(theme, "user@example.com")
	footer.SetWidth(150)

	view := footer.View()

	// Should contain user emoji
	if !strings.Contains(view, "👤") {
		t.Error("Footer should contain user emoji badge")
	}

	// Should contain email
	if !strings.Contains(view, "user@example.com") {
		t.Error("Footer should display user email")
	}
}

func TestFooterBar_AlternatingColors(t *testing.T) {
	theme := styles.GetTheme("k9s")
	footer := NewFooterBar(theme, "test@example.com")
	footer.SetWidth(200)
	footer.SetBindings([]KeyBinding{
		{Key: "a", Description: "First"},
		{Key: "b", Description: "Second"},
		{Key: "c", Description: "Third"},
		{Key: "d", Description: "Fourth"},
	})

	view := footer.View()

	// Should contain all bindings
	if !strings.Contains(view, "First") {
		t.Error("Footer should contain First binding")
	}

	if !strings.Contains(view, "Second") {
		t.Error("Footer should contain Second binding")
	}

	if !strings.Contains(view, "Third") {
		t.Error("Footer should contain Third binding")
	}

	if !strings.Contains(view, "Fourth") {
		t.Error("Footer should contain Fourth binding")
	}
}

func TestFooterBar_MinimalView(t *testing.T) {
	theme := styles.GetTheme("k9s")
	footer := NewFooterBar(theme, "verylongemail@example.com")
	footer.SetWidth(30) // Very small width
	footer.SetBindings([]KeyBinding{
		{Key: "a", Description: "Action"},
	})

	view := footer.View()

	// Should still return content (minimal version)
	if view == "" {
		t.Error("Footer should show minimal version when width is small")
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

func TestFooterBar_EmptyBindings(t *testing.T) {
	theme := styles.GetTheme("k9s")
	footer := NewFooterBar(theme, "test@example.com")
	footer.SetWidth(100)
	footer.SetBindings([]KeyBinding{})

	view := footer.View()

	// Should still display email even with no bindings
	if !strings.Contains(view, "test@example.com") {
		t.Error("Footer should display email even with empty bindings")
	}

	// Should contain user emoji
	if !strings.Contains(view, "👤") {
		t.Error("Footer should contain user emoji even with empty bindings")
	}
}
