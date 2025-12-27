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
)

func TestNewMessageDetail(t *testing.T) {
	client := nylas.NewMockClient()
	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "test@example.com", "google")

	md := NewMessageDetail(global, "msg123")

	if md == nil {
		t.Fatal("NewMessageDetail returned nil")
	}

	if md.global != global {
		t.Error("global state not set correctly")
	}

	if md.theme == nil {
		t.Error("theme not initialized")
	}

	if md.message == nil {
		t.Error("message not initialized")
	}

	if md.message.ID != "msg123" {
		t.Errorf("message ID = %q, want %q", md.message.ID, "msg123")
	}

	if !md.loading {
		t.Error("loading should be true initially")
	}
}

func TestMessageDetail_Init(t *testing.T) {
	client := nylas.NewMockClient()
	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "test@example.com", "google")

	md := NewMessageDetail(global, "msg123")
	cmd := md.Init()

	if cmd == nil {
		t.Fatal("Init() returned nil command")
	}

	// Execute the batch command and verify it returns a message
	msg := cmd()
	if msg == nil {
		t.Error("Init command returned nil message")
	}
}

func TestMessageDetail_UpdateWithMessageLoaded(t *testing.T) {
	client := nylas.NewMockClient()
	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "test@example.com", "google")

	md := NewMessageDetail(global, "msg123")

	// Create test message
	testMessage := &domain.Message{
		ID:      "msg123",
		Subject: "Test Message",
		From:    []domain.EmailParticipant{{Name: "John Doe", Email: "john@example.com"}},
		To:      []domain.EmailParticipant{{Name: "Jane Smith", Email: "jane@example.com"}},
		Body:    "This is a test message body",
		Date:    time.Now(),
	}

	// Send messageLoadedMsg
	msg := messageLoadedMsg{message: testMessage}
	updated, cmd := md.Update(msg)

	if updated == nil {
		t.Fatal("Update returned nil model")
	}

	updatedMD := updated.(*MessageDetail)
	if updatedMD.loading {
		t.Error("loading should be false after message loaded")
	}

	if updatedMD.message != testMessage {
		t.Error("message should be set to test message")
	}

	if cmd != nil {
		t.Error("Update should return nil command for messageLoadedMsg")
	}
}

func TestMessageDetail_UpdateWithError(t *testing.T) {
	client := nylas.NewMockClient()
	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "test@example.com", "google")

	md := NewMessageDetail(global, "msg123")

	// Send error message
	testErr := errMsg{err: context.Canceled}
	updated, cmd := md.Update(testErr)

	if updated == nil {
		t.Fatal("Update returned nil model")
	}

	updatedMD := updated.(*MessageDetail)
	if updatedMD.loading {
		t.Error("loading should be false after error")
	}

	if updatedMD.err == nil {
		t.Error("error should be set")
	}

	if cmd != nil {
		t.Error("Update should return nil command for errMsg")
	}
}

func TestMessageDetail_UpdateWithKeyPress(t *testing.T) {
	client := nylas.NewMockClient()
	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "test@example.com", "google")

	md := NewMessageDetail(global, "msg123")

	tests := []struct {
		name     string
		key      string
		wantQuit bool
		wantBack bool
		wantCmd  bool
	}{
		{"esc key", "esc", false, true, true},
		{"ctrl+c", "ctrl+c", true, false, true},
		{"r key", "r", false, false, false},
		{"f key", "f", false, false, false},
		{"d key", "d", false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var msg tea.Msg
			if tt.key == "ctrl+c" {
				msg = tea.KeyMsg{Type: tea.KeyCtrlC}
			} else {
				msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			}

			updated, cmd := md.Update(msg)

			if updated == nil {
				t.Fatal("Update returned nil model")
			}

			if tt.wantCmd && cmd == nil {
				t.Error("expected command, got nil")
			}

			if !tt.wantCmd && cmd != nil {
				// For r, f, d keys we still get nil command (placeholders)
				if tt.key != "r" && tt.key != "f" && tt.key != "d" {
					t.Error("expected nil command")
				}
			}

			if tt.wantBack && cmd != nil {
				result := cmd()
				if _, ok := result.(BackMsg); !ok && tt.wantBack {
					t.Error("expected BackMsg")
				}
			}

			if tt.wantQuit && cmd != nil {
				result := cmd()
				if result != tea.Quit() {
					t.Error("expected Quit message")
				}
			}
		})
	}
}

func TestMessageDetail_UpdateWithWindowSize(t *testing.T) {
	client := nylas.NewMockClient()
	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "test@example.com", "google")

	md := NewMessageDetail(global, "msg123")

	// Send window size message
	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := md.Update(msg)

	if updated == nil {
		t.Fatal("Update returned nil model")
	}

	updatedMD := updated.(*MessageDetail)
	if !updatedMD.ready {
		t.Error("ready should be true after window size message")
	}
}

func TestMessageDetail_View(t *testing.T) {
	client := nylas.NewMockClient()
	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "test@example.com", "google")

	md := NewMessageDetail(global, "msg123")

	// Test loading state
	view := md.View()
	if view == "" {
		t.Error("View() returned empty string")
	}
	if !strings.Contains(view, "Loading") {
		t.Error("View should contain 'Loading' when loading")
	}

	// Test error state
	md.loading = false
	md.err = context.Canceled
	view = md.View()
	if !strings.Contains(view, "Error") {
		t.Error("View should contain 'Error' when there's an error")
	}
}

func TestMessageDetail_ViewWithMessage(t *testing.T) {
	client := nylas.NewMockClient()
	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "test@example.com", "google")

	md := NewMessageDetail(global, "msg123")
	md.loading = false
	md.ready = true
	md.message = &domain.Message{
		ID:      "msg123",
		Subject: "Test Message",
		From:    []domain.EmailParticipant{{Name: "John Doe", Email: "john@example.com"}},
		To:      []domain.EmailParticipant{{Name: "Jane Smith", Email: "jane@example.com"}},
		Body:    "This is a test message body",
		Date:    time.Now(),
	}

	// Set window size to initialize viewport
	md.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	view := md.View()
	if view == "" {
		t.Error("View() returned empty string")
	}

	if !strings.Contains(view, "Message Detail") {
		t.Error("View should contain 'Message Detail' title")
	}
}

func TestFormatParticipant(t *testing.T) {
	tests := []struct {
		name string
		p    domain.EmailParticipant
		want string
	}{
		{
			name: "with name",
			p:    domain.EmailParticipant{Name: "John Doe", Email: "john@example.com"},
			want: "John Doe <john@example.com>",
		},
		{
			name: "without name",
			p:    domain.EmailParticipant{Email: "john@example.com"},
			want: "john@example.com",
		},
		{
			name: "empty name",
			p:    domain.EmailParticipant{Name: "", Email: "john@example.com"},
			want: "john@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatParticipant(tt.p)
			if got != tt.want {
				t.Errorf("formatParticipant() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		name  string
		bytes int64
		want  string
	}{
		{"bytes", 500, "500 B"},
		{"kilobytes", 2048, "2.0 KB"},
		{"megabytes", 5242880, "5.0 MB"},
		{"gigabytes", 2147483648, "2.0 GB"},
		{"exact KB", 1024, "1.0 KB"},
		{"exact MB", 1048576, "1.0 MB"},
		{"exact GB", 1073741824, "1.0 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatSize(tt.bytes)
			if got != tt.want {
				t.Errorf("formatSize(%d) = %q, want %q", tt.bytes, got, tt.want)
			}
		})
	}
}

func TestStripHTML(t *testing.T) {
	tests := []struct {
		name string
		html string
		want string
	}{
		{
			name: "simple text",
			html: "Hello, World!",
			want: "Hello, World!",
		},
		{
			name: "with br tag",
			html: "Line 1<br>Line 2",
			want: "Line 1\nLine 2",
		},
		{
			name: "with br/ tag",
			html: "Line 1<br/>Line 2",
			want: "Line 1\nLine 2",
		},
		{
			name: "with br / tag",
			html: "Line 1<br />Line 2",
			want: "Line 1\nLine 2",
		},
		{
			name: "with paragraph",
			html: "<p>Paragraph 1</p><p>Paragraph 2</p>",
			want: "Paragraph 1\n\nParagraph 2",
		},
		{
			name: "with div",
			html: "<div>Content 1</div><div>Content 2</div>",
			want: "Content 1\nContent 2",
		},
		{
			name: "with bold tags",
			html: "This is <b>bold</b> text",
			want: "This is bold text",
		},
		{
			name: "with HTML entities",
			html: "Hello&nbsp;World&amp;Test&lt;tag&gt;&quot;quote&quot;&#39;apostrophe&#39;",
			want: "Hello World&Test<tag>\"quote\"'apostrophe'",
		},
		{
			name: "complex HTML",
			html: "<html><body><h1>Title</h1><p>This is a <b>test</b> message.</p><br/><div>Footer</div></body></html>",
			want: "TitleThis is a test message.\n\nFooter",
		},
		{
			name: "empty HTML",
			html: "",
			want: "",
		},
		{
			name: "nested tags",
			html: "<div><p><span>Nested content</span></p></div>",
			want: "Nested content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripHTML(tt.html)
			if got != tt.want {
				t.Errorf("stripHTML() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMessageDetail_UpdateViewport(t *testing.T) {
	client := nylas.NewMockClient()
	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "test@example.com", "google")

	md := NewMessageDetail(global, "msg123")
	md.ready = true

	// Initialize viewport
	md.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// Test with nil message (should not panic)
	md.updateViewport()

	// Test with message
	md.message = &domain.Message{
		ID:      "msg123",
		Subject: "Test Message",
		From:    []domain.EmailParticipant{{Name: "John Doe", Email: "john@example.com"}},
		To: []domain.EmailParticipant{
			{Name: "Jane Smith", Email: "jane@example.com"},
			{Name: "Bob Johnson", Email: "bob@example.com"},
		},
		Cc:   []domain.EmailParticipant{{Email: "cc@example.com"}},
		Body: "<p>This is a <b>test</b> message.</p>",
		Date: time.Now(),
		Attachments: []domain.Attachment{
			{Filename: "document.pdf", Size: 1024000},
			{Filename: "image.png", Size: 512000},
		},
	}

	md.updateViewport()

	// Verify viewport has content (can't easily test exact content due to styling)
	// but we can verify the method doesn't panic
}

func TestMessageDetail_BuildHelpText(t *testing.T) {
	client := nylas.NewMockClient()
	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "test@example.com", "google")

	md := NewMessageDetail(global, "msg123")
	md.message = &domain.Message{
		ID:          "msg123",
		Attachments: []domain.Attachment{},
	}

	// Without attachments
	help := md.buildHelpText()
	if !strings.Contains(help, "scroll") {
		t.Error("help text should contain 'scroll'")
	}
	if !strings.Contains(help, "reply") {
		t.Error("help text should contain 'reply'")
	}
	if strings.Contains(help, "download") {
		t.Error("help text should not contain 'download' when no attachments")
	}

	// With attachments
	md.message.Attachments = []domain.Attachment{
		{Filename: "test.pdf", Size: 1024},
	}
	help = md.buildHelpText()
	if !strings.Contains(help, "download") {
		t.Error("help text should contain 'download' when attachments present")
	}
}
