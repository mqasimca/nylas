package models

import (
	"context"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mqasimca/nylas/internal/adapters/keyring"
	"github.com/mqasimca/nylas/internal/adapters/nylas"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/tui2/state"
	"github.com/mqasimca/nylas/internal/tui2/styles"
)

func TestNewMessageList(t *testing.T) {
	client := nylas.NewMockClient()
	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "test@example.com", "google")

	ml := NewMessageList(global)

	if ml == nil {
		t.Fatal("NewMessageList returned nil")
	}

	if ml.global != global {
		t.Error("global state not set correctly")
	}

	if ml.theme == nil {
		t.Error("theme not initialized")
	}

	if ml.layout == nil {
		t.Error("layout not initialized")
	}

	if !ml.loading {
		t.Error("loading should be true initially")
	}
}

func TestMessageList_Init(t *testing.T) {
	client := nylas.NewMockClient()
	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "test@example.com", "google")

	ml := NewMessageList(global)
	cmd := ml.Init()

	if cmd == nil {
		t.Fatal("Init() returned nil command")
	}

	// Execute the batch command and verify it returns messages
	msg := cmd()
	if msg == nil {
		t.Error("Init command returned nil message")
	}
}

func TestMessageList_UpdateWithMessages(t *testing.T) {
	client := nylas.NewMockClient()
	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "test@example.com", "google")

	ml := NewMessageList(global)

	// Create test messages
	testMessages := []domain.Message{
		{
			ID:      "msg1",
			Subject: "Test Message 1",
			From:    []domain.EmailParticipant{{Name: "John Doe", Email: "john@example.com"}},
			Date:    time.Now(),
		},
		{
			ID:      "msg2",
			Subject: "Test Message 2",
			From:    []domain.EmailParticipant{{Name: "Jane Smith", Email: "jane@example.com"}},
			Date:    time.Now().Add(-24 * time.Hour),
		},
	}

	// Send messagesLoadedMsg
	msg := messagesLoadedMsg{messages: testMessages}
	updated, cmd := ml.Update(msg)

	if updated == nil {
		t.Fatal("Update returned nil model")
	}

	updatedML := updated.(*MessageList)
	if updatedML.loading {
		t.Error("loading should be false after messages loaded")
	}

	if len(updatedML.messages) != len(testMessages) {
		t.Errorf("messages count = %d, want %d", len(updatedML.messages), len(testMessages))
	}

	if cmd != nil {
		t.Error("Update should return nil command for messagesLoadedMsg")
	}
}

func TestMessageList_UpdateWithError(t *testing.T) {
	client := nylas.NewMockClient()
	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "test@example.com", "google")

	ml := NewMessageList(global)

	// Send error message
	testErr := errMsg{err: context.Canceled}
	updated, cmd := ml.Update(testErr)

	if updated == nil {
		t.Fatal("Update returned nil model")
	}

	updatedML := updated.(*MessageList)
	if updatedML.loading {
		t.Error("loading should be false after error")
	}

	if updatedML.err == nil {
		t.Error("error should be set")
	}

	if cmd != nil {
		t.Error("Update should return nil command for errMsg")
	}
}

func TestMessageList_UpdateWithKeyPress(t *testing.T) {
	client := nylas.NewMockClient()
	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "test@example.com", "google")

	ml := NewMessageList(global)

	// Test that Update doesn't panic with various key messages
	// Note: Actual key handling is tested in integration tests
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}
	updated, _ := ml.Update(msg)

	if updated == nil {
		t.Fatal("Update returned nil model")
	}
}

func TestMessageList_View(t *testing.T) {
	client := nylas.NewMockClient()
	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "test@example.com", "google")

	ml := NewMessageList(global)

	view := ml.View()
	if view == "" {
		t.Error("View() returned empty string")
	}

	// View should contain "Messages" title
	if !strings.Contains(view, "Messages") {
		t.Error("View should contain 'Messages' title")
	}

	// View should contain email address
	if !strings.Contains(view, "test@example.com") {
		t.Error("View should contain email address")
	}
}

func TestMessageList_ViewWithError(t *testing.T) {
	client := nylas.NewMockClient()
	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "test@example.com", "google")

	ml := NewMessageList(global)
	ml.err = context.Canceled

	view := ml.View()
	if view == "" {
		t.Error("View() returned empty string")
	}

	// View should contain error message
	if !strings.Contains(view, "Error") {
		t.Error("View should contain 'Error' when there's an error")
	}
}

func TestFormatDate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name string
		time time.Time
		want string
	}{
		{"just now", now, "just now"},
		{"1 minute ago", now.Add(-1 * time.Minute), "1m ago"},
		{"30 minutes ago", now.Add(-30 * time.Minute), "30m ago"},
		{"1 hour ago", now.Add(-1 * time.Hour), "1h ago"},
		{"5 hours ago", now.Add(-5 * time.Hour), "5h ago"},
		{"1 day ago", now.Add(-24 * time.Hour), "1d ago"},
		{"3 days ago", now.Add(-3 * 24 * time.Hour), "3d ago"},
		{"10 days ago", now.Add(-10 * 24 * time.Hour), now.Add(-10 * 24 * time.Hour).Format("Jan 2")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDate(tt.time)
			if got != tt.want {
				t.Errorf("formatDate() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name  string
		input string
		max   int
		want  string
	}{
		{"short string", "Hello", 10, "Hello"},
		{"exact length", "Hello", 5, "Hello"},
		{"truncate needed", "Hello World", 8, "Hello..."},
		{"truncate with max 3", "Hello", 3, "..."},
		{"empty string", "", 10, ""},
		{"max zero", "Hello", 0, ""},
		{"max one", "Hello", 1, "H"},
		{"max two", "Hello", 2, "He"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncate(tt.input, tt.max)
			if got != tt.want {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.max, got, tt.want)
			}
		})
	}
}

func TestMessageList_ShowMessagePreview(t *testing.T) {
	client := nylas.NewMockClient()
	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "test@example.com", "google")

	ml := NewMessageList(global)
	ml.theme = styles.DefaultTheme()

	// Set up test messages
	ml.messages = []domain.Message{
		{
			ID:      "msg1",
			Subject: "Test Message",
			From:    []domain.EmailParticipant{{Name: "John Doe", Email: "john@example.com"}},
			To:      []domain.EmailParticipant{{Name: "Jane Smith", Email: "jane@example.com"}},
			Body:    "This is a test message body",
			Date:    time.Now(),
		},
	}

	// Set size to avoid nil viewport issues
	ml.layout.SetSize(120, 40)

	// Show preview
	ml.showMessagePreview("msg1")

	// We can't easily test the preview content directly,
	// but we can verify the method doesn't panic
}

func TestMessageList_ShowMessagePreview_NotFound(t *testing.T) {
	client := nylas.NewMockClient()
	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "test@example.com", "google")

	ml := NewMessageList(global)
	ml.theme = styles.DefaultTheme()
	ml.layout.SetSize(120, 40)

	// Show preview for non-existent message
	ml.showMessagePreview("nonexistent")

	// Should not panic
}
