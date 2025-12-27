package models

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mqasimca/nylas/internal/adapters/keyring"
	"github.com/mqasimca/nylas/internal/adapters/nylas"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/tui2/components"
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

// Keyboard Navigation Tests

func TestMessageList_KeyboardNavigation_Esc(t *testing.T) {
	client := nylas.NewMockClient()
	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "test@example.com", "google")

	ml := NewMessageList(global)
	ml.loading = false // Set to not loading
	ml.layout.SetSize(120, 40)

	// Test ESC key - should return BackMsg
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	_, cmd := ml.Update(msg)

	if cmd == nil {
		t.Fatal("ESC should return a command")
	}

	result := cmd()
	if _, ok := result.(BackMsg); !ok {
		t.Errorf("ESC should return BackMsg, got %T", result)
	}
}

func TestMessageList_KeyboardNavigation_CtrlC(t *testing.T) {
	client := nylas.NewMockClient()
	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "test@example.com", "google")

	ml := NewMessageList(global)
	ml.loading = false
	ml.layout.SetSize(120, 40)

	// Test Ctrl+C - should return tea.Quit
	msg := tea.KeyMsg{Type: tea.KeyCtrlC}
	_, cmd := ml.Update(msg)

	if cmd == nil {
		t.Fatal("Ctrl+C should return a command")
	}

	result := cmd()
	if result != tea.Quit() {
		t.Error("Ctrl+C should return tea.Quit message")
	}
}

func TestMessageList_KeyboardNavigation_Tab(t *testing.T) {
	client := nylas.NewMockClient()
	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "test@example.com", "google")

	ml := NewMessageList(global)
	ml.loading = false
	ml.layout.SetSize(120, 40)

	// Get initial focused pane
	initialPane := ml.layout.GetFocused()

	// Test Tab key - should focus next pane
	msg := tea.KeyMsg{Type: tea.KeyTab}
	updated, _ := ml.Update(msg)

	updatedML := updated.(*MessageList)
	newPane := updatedML.layout.GetFocused()

	if newPane == initialPane {
		t.Error("Tab should change focused pane")
	}
}

func TestMessageList_KeyboardNavigation_ShiftTab(t *testing.T) {
	client := nylas.NewMockClient()
	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "test@example.com", "google")

	ml := NewMessageList(global)
	ml.loading = false
	ml.layout.SetSize(120, 40)

	// Get initial focused pane
	initialPane := ml.layout.GetFocused()

	// Test Shift+Tab key - should focus previous pane
	msg := tea.KeyMsg{Type: tea.KeyShiftTab}
	updated, _ := ml.Update(msg)

	updatedML := updated.(*MessageList)
	newPane := updatedML.layout.GetFocused()

	if newPane == initialPane {
		t.Error("Shift+Tab should change focused pane")
	}
}

func TestMessageList_KeyboardNavigation_H(t *testing.T) {
	client := nylas.NewMockClient()
	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "test@example.com", "google")

	ml := NewMessageList(global)
	ml.loading = false
	ml.layout.SetSize(120, 40)

	// Get initial focused pane
	initialPane := ml.layout.GetFocused()

	// Test 'h' key - should focus previous pane
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")}
	updated, _ := ml.Update(msg)

	updatedML := updated.(*MessageList)
	newPane := updatedML.layout.GetFocused()

	if newPane == initialPane {
		t.Error("'h' key should change focused pane")
	}
}

func TestMessageList_KeyboardNavigation_L(t *testing.T) {
	client := nylas.NewMockClient()
	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "test@example.com", "google")

	ml := NewMessageList(global)
	ml.loading = false
	ml.layout.SetSize(120, 40)

	// Get initial focused pane
	initialPane := ml.layout.GetFocused()

	// Test 'l' key - should focus next pane
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")}
	updated, _ := ml.Update(msg)

	updatedML := updated.(*MessageList)
	newPane := updatedML.layout.GetFocused()

	if newPane == initialPane {
		t.Error("'l' key should change focused pane")
	}
}

func TestMessageList_KeyboardNavigation_Refresh(t *testing.T) {
	client := nylas.NewMockClient()
	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "test@example.com", "google")

	ml := NewMessageList(global)
	ml.loading = false
	ml.layout.SetSize(120, 40)

	// Test 'r' key - should set loading and return fetch command
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")}
	updated, cmd := ml.Update(msg)

	updatedML := updated.(*MessageList)
	if !updatedML.loading {
		t.Error("'r' key should set loading to true")
	}

	if cmd == nil {
		t.Error("'r' key should return a fetch command")
	}
}

func TestMessageList_KeyboardNavigation_EnterOnMessagePane(t *testing.T) {
	client := nylas.NewMockClient()
	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "test@example.com", "google")

	ml := NewMessageList(global)
	ml.loading = false
	ml.layout.SetSize(120, 40)

	// Set up test messages
	ml.messages = []domain.Message{
		{ID: "msg1", Subject: "Test 1"},
		{ID: "msg2", Subject: "Test 2"},
	}
	ml.updateMessageTable()

	// Focus on message pane
	ml.layout.FocusPane(components.MessagePane)

	// Test Enter key - should navigate to message detail
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	_, cmd := ml.Update(msg)

	if cmd == nil {
		t.Fatal("Enter on MessagePane should return a command")
	}

	result := cmd()
	navMsg, ok := result.(NavigateMsg)
	if !ok {
		t.Fatalf("Enter should return NavigateMsg, got %T", result)
	}

	if navMsg.Screen != ScreenMessageDetail {
		t.Errorf("Navigate screen = %v, want ScreenMessageDetail", navMsg.Screen)
	}

	if navMsg.Data != "msg1" {
		t.Errorf("Navigate data = %v, want 'msg1'", navMsg.Data)
	}
}

func TestMessageList_KeyboardNavigation_EnterOnFolderPane(t *testing.T) {
	client := nylas.NewMockClient()
	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "test@example.com", "google")

	ml := NewMessageList(global)
	ml.loading = false
	ml.layout.SetSize(120, 40)

	// Set folders to loaded
	ml.foldersLoaded = true

	// Create a test folder item (not the special "show all" folder)
	testFolder := domain.Folder{
		ID:         "folder123",
		Name:       "Sent",
		TotalCount: 5,
	}

	ml.layout.SetFolders([]list.Item{
		components.FolderItem{Folder: testFolder},
	})

	// Focus on folder pane
	ml.layout.FocusPane(components.FolderPane)

	// Test Enter key - should set selectedFolderID and trigger fetch
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, cmd := ml.Update(msg)

	updatedML := updated.(*MessageList)
	if updatedML.selectedFolderID != "folder123" {
		t.Errorf("selectedFolderID = %q, want 'folder123'", updatedML.selectedFolderID)
	}

	if !updatedML.loading {
		t.Error("Enter on folder should set loading to true")
	}

	if cmd == nil {
		t.Error("Enter on folder should return a fetch command")
	}
}

func TestMessageList_KeyboardNavigation_LazyLoadFolders(t *testing.T) {
	client := nylas.NewMockClient()
	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "test@example.com", "google")

	ml := NewMessageList(global)
	ml.loading = false
	ml.layout.SetSize(120, 40)

	// Folders not loaded yet
	if ml.foldersLoaded {
		t.Fatal("Folders should not be loaded initially")
	}

	// Focus starts on MessagePane, navigate to FolderPane with Tab
	// This should trigger lazy loading of folders
	msg := tea.KeyMsg{Type: tea.KeyTab}
	updated, cmd := ml.Update(msg)

	updatedML := updated.(*MessageList)

	// If we landed on FolderPane, folders should start loading
	if updatedML.layout.GetFocused() == components.FolderPane {
		if !updatedML.loadingFolders {
			t.Error("Focusing FolderPane should trigger folder loading")
		}

		if cmd == nil {
			t.Error("Focusing FolderPane should return fetch folders command")
		}
	}
}

func TestMessageList_KeyboardNavigation_WindowResize(t *testing.T) {
	client := nylas.NewMockClient()
	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "test@example.com", "google")

	ml := NewMessageList(global)
	ml.loading = false

	// Initial size
	ml.layout.SetSize(100, 30)

	// Send window resize message
	msg := tea.WindowSizeMsg{Width: 150, Height: 50}
	updated, cmd := ml.Update(msg)

	if cmd != nil {
		t.Error("WindowSizeMsg should not return a command")
	}

	updatedML := updated.(*MessageList)

	// Verify global state was updated
	if updatedML.global.WindowSize.Width != 150 {
		t.Errorf("Window width = %d, want 150", updatedML.global.WindowSize.Width)
	}

	if updatedML.global.WindowSize.Height != 50 {
		t.Errorf("Window height = %d, want 50", updatedML.global.WindowSize.Height)
	}
}

func TestMessageList_KeyboardNavigation_ArrowKeys(t *testing.T) {
	client := nylas.NewMockClient()
	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "test@example.com", "google")

	ml := NewMessageList(global)
	ml.loading = false
	ml.layout.SetSize(120, 40)

	// Set up test messages
	ml.messages = []domain.Message{
		{ID: "msg1", Subject: "Test 1"},
		{ID: "msg2", Subject: "Test 2"},
		{ID: "msg3", Subject: "Test 3"},
	}
	ml.updateMessageTable()

	// Focus on message pane
	ml.layout.FocusPane(components.MessagePane)

	tests := []struct {
		name string
		key  tea.KeyType
	}{
		{"arrow_up", tea.KeyUp},
		{"arrow_down", tea.KeyDown},
		{"page_up", tea.KeyPgUp},
		{"page_down", tea.KeyPgDown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tea.KeyMsg{Type: tt.key}
			updated, cmd := ml.Update(msg)

			if updated == nil {
				t.Error("Update should return a model")
			}

			// Arrow keys should be passed to the layout
			// cmd may be nil or contain layout update commands
			_ = cmd
		})
	}
}
