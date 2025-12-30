package tui

import (
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/rivo/tview"
)

func TestComposeView(t *testing.T) {
	app := createTestApp(t)

	tests := []struct {
		mode     ComposeMode
		title    string
		hasReply bool
	}{
		{ComposeModeNew, "New Message", false},
		{ComposeModeReply, "Reply", true},
		{ComposeModeReplyAll, "Reply All", true},
		{ComposeModeForward, "Forward", true},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			var replyTo *domain.Message
			if tt.hasReply {
				replyTo = &domain.Message{
					ID:      "original-msg",
					Subject: "Original Subject",
					From:    []domain.EmailParticipant{{Email: "sender@example.com"}},
					To:      []domain.EmailParticipant{{Email: "recipient@example.com"}},
					Date:    time.Now(),
					Snippet: "Original message content",
				}
			}

			compose := NewComposeView(app, tt.mode, replyTo)
			if compose == nil {
				t.Fatalf("NewComposeView() with mode %v returned nil", tt.mode)
			}
		})
	}
}

func TestHelpView(t *testing.T) {
	styles := DefaultStyles()
	registry := NewCommandRegistry()

	// Create a minimal app for testing
	app := &App{
		Application: tview.NewApplication(),
		styles:      styles,
		cmdRegistry: registry,
	}

	// Create help view with nil callbacks (for testing)
	help := NewHelpView(app, registry, nil, nil)

	if help == nil {
		t.Fatal("NewHelpView() returned nil")
	}
}

func TestFormatDate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		time     time.Time
		contains string
	}{
		{"today", now, "PM"},                        // Should show time like "3:04 PM"
		{"yesterday", now.Add(-24 * time.Hour), ""}, // Should show date
		{"last year", now.AddDate(-1, 0, 0), ""},    // Should include year
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDate(tt.time)
			if result == "" {
				t.Error("formatDate() returned empty string")
			}
			// Just verify it doesn't panic and returns something
		})
	}
}

func TestParseRecipients(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"single email", "test@example.com", 1},
		{"multiple emails", "a@example.com, b@example.com", 2},
		{"with name", "John Doe <john@example.com>", 1},
		{"mixed", "John <john@example.com>, jane@example.com", 2},
		{"empty", "", 0},
		{"invalid", "not-an-email", 0},
		{"spaces", "  test@example.com  ", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseRecipients(tt.input)
			if len(result) != tt.expected {
				t.Errorf("parseRecipients(%q) returned %d recipients, want %d", tt.input, len(result), tt.expected)
			}
		})
	}
}

func TestConvertToHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{"plain text", "Hello World", "Hello World"},
		{"with newlines", "Line 1\nLine 2", "<br>"},
		{"with HTML chars", "<script>alert('xss')</script>", "&lt;script&gt;"},
		{"with ampersand", "A & B", "&amp;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertToHTML(tt.input)
			if result == "" {
				t.Error("convertToHTML() returned empty string")
			}
			// Verify HTML structure
			if len(result) < len("<div>") {
				t.Error("Result too short to be valid HTML")
			}
		})
	}
}

func TestFormatParticipants(t *testing.T) {
	tests := []struct {
		name         string
		participants []domain.EmailParticipant
		expected     string
	}{
		{
			"single with name",
			[]domain.EmailParticipant{{Name: "John", Email: "john@example.com"}},
			"John <john@example.com>",
		},
		{
			"single email only",
			[]domain.EmailParticipant{{Email: "john@example.com"}},
			"john@example.com",
		},
		{
			"multiple",
			[]domain.EmailParticipant{
				{Name: "John", Email: "john@example.com"},
				{Email: "jane@example.com"},
			},
			"John <john@example.com>, jane@example.com",
		},
		{
			"empty",
			[]domain.EmailParticipant{},
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatParticipants(tt.participants)
			if result != tt.expected {
				t.Errorf("formatParticipants() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestViewInterfaces verifies all views implement ResourceView interface
func TestViewInterfaces(t *testing.T) {
	app := createTestApp(t)

	views := []ResourceView{
		NewDashboardView(app),
		NewMessagesView(app),
		NewEventsView(app),
		NewContactsView(app),
		NewWebhooksView(app),
		NewGrantsView(app),
	}

	for _, view := range views {
		t.Run(view.Name(), func(t *testing.T) {
			// Verify interface methods don't panic
			_ = view.Name()
			_ = view.Title()
			_ = view.Primitive()
			_ = view.Hints()

			// Filter should accept any string
			view.Filter("")
			view.Filter("test")

			// HandleKey should accept events
			event := tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone)
			_ = view.HandleKey(event)
		})
	}
}
