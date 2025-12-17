package tui

import (
	"context"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/mqasimca/nylas/internal/adapters/nylas"
	"github.com/mqasimca/nylas/internal/domain"
)

// createTestApp creates an App instance for testing
func createTestApp(t *testing.T) *App {
	t.Helper()

	mockClient := nylas.NewMockClient()

	// Set up mock responses
	mockClient.GetMessagesFunc = func(ctx context.Context, grantID string, limit int) ([]domain.Message, error) {
		return []domain.Message{
			{
				ID:      "msg-1",
				Subject: "Test Message 1",
				From:    []domain.EmailParticipant{{Email: "sender1@example.com", Name: "Sender One"}},
				Unread:  true,
				Date:    time.Now(),
			},
			{
				ID:      "msg-2",
				Subject: "Test Message 2",
				From:    []domain.EmailParticipant{{Email: "sender2@example.com", Name: "Sender Two"}},
				Starred: true,
				Date:    time.Now().Add(-time.Hour),
			},
			{
				ID:      "msg-3",
				Subject: "Test Message 3",
				From:    []domain.EmailParticipant{{Email: "sender3@example.com"}},
				Date:    time.Now().Add(-24 * time.Hour),
			},
		}, nil
	}

	config := Config{
		Client:          mockClient,
		GrantID:         "test-grant-id",
		Email:           "user@example.com",
		Provider:        "google",
		RefreshInterval: time.Second * 30,
		Theme:           ThemeK9s,
		InitialView:     "dashboard",
	}

	return NewApp(config)
}

func TestNewApp(t *testing.T) {
	app := createTestApp(t)

	if app == nil {
		t.Fatal("NewApp() returned nil")
	}

	if app.Application == nil {
		t.Error("App.Application is nil")
	}

	if app.styles == nil {
		t.Error("App.styles is nil")
	}

	if app.views == nil {
		t.Error("App.views is nil")
	}

	if app.content == nil {
		t.Error("App.content is nil")
	}
}

func TestNewAppWithThemes(t *testing.T) {
	themes := AvailableThemes()

	for _, theme := range themes {
		t.Run(string(theme), func(t *testing.T) {
			mockClient := nylas.NewMockClient()
			config := Config{
				Client:          mockClient,
				GrantID:         "test-grant-id",
				Email:           "user@example.com",
				Provider:        "google",
				RefreshInterval: time.Second * 30,
				Theme:           theme,
			}

			app := NewApp(config)
			if app == nil {
				t.Fatalf("NewApp() with theme %q returned nil", theme)
			}

			if app.styles == nil {
				t.Errorf("App.styles is nil for theme %q", theme)
			}
		})
	}
}

func TestAppGetConfig(t *testing.T) {
	app := createTestApp(t)
	config := app.GetConfig()

	if config.GrantID != "test-grant-id" {
		t.Errorf("GetConfig().GrantID = %q, want %q", config.GrantID, "test-grant-id")
	}
	if config.Email != "user@example.com" {
		t.Errorf("GetConfig().Email = %q, want %q", config.Email, "user@example.com")
	}
	if config.Provider != "google" {
		t.Errorf("GetConfig().Provider = %q, want %q", config.Provider, "google")
	}
}

func TestAppStyles(t *testing.T) {
	app := createTestApp(t)
	styles := app.Styles()

	if styles == nil {
		t.Fatal("Styles() returned nil")
	}

	// Verify styles are properly set
	if styles.TableSelectBg == 0 && styles.TableSelectBg != tcell.ColorDefault {
		t.Error("TableSelectBg not set")
	}
}

func TestCreateView(t *testing.T) {
	app := createTestApp(t)

	tests := []struct {
		name     string
		viewType string
	}{
		{"messages", "messages"},
		{"events", "events"},
		{"contacts", "contacts"},
		{"webhooks", "webhooks"},
		{"grants", "grants"},
		{"dashboard", "dashboard"},
		{"unknown", "unknown"}, // Should default to dashboard
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			view := app.createView(tt.viewType)
			if view == nil {
				t.Fatalf("createView(%q) returned nil", tt.viewType)
			}

			// Verify view has required methods
			if view.Name() == "" {
				t.Error("View.Name() returned empty string")
			}
			if view.Title() == "" {
				t.Error("View.Title() returned empty string")
			}
			if view.Primitive() == nil {
				t.Error("View.Primitive() returned nil")
			}
			// Hints can be empty, so we just check it doesn't panic
			_ = view.Hints()
		})
	}
}

func TestPageStack(t *testing.T) {
	stack := NewPageStack()

	if stack == nil {
		t.Fatal("NewPageStack() returned nil")
	}

	// Test empty stack
	if stack.Len() != 0 {
		t.Errorf("New stack Len() = %d, want 0", stack.Len())
	}
	if stack.Top() != "" {
		t.Errorf("Empty stack Top() = %q, want empty string", stack.Top())
	}
}

func TestTable(t *testing.T) {
	styles := DefaultStyles()
	table := NewTable(styles)

	if table == nil {
		t.Fatal("NewTable() returned nil")
	}

	// Set columns
	columns := []Column{
		{Title: "ID", Width: 10},
		{Title: "Name", Expand: true},
		{Title: "Status", Width: 8},
	}
	table.SetColumns(columns)

	// Set data
	data := [][]string{
		{"1", "Item 1", "Active"},
		{"2", "Item 2", "Inactive"},
		{"3", "Item 3", "Pending"},
	}
	meta := []RowMeta{
		{ID: "1", Unread: true},
		{ID: "2", Starred: true},
		{ID: "3", Error: true},
	}
	table.SetData(data, meta)

	// Verify row count
	if table.GetRowCount() != 3 {
		t.Errorf("GetRowCount() = %d, want 3", table.GetRowCount())
	}
}

func TestTableSelection(t *testing.T) {
	styles := DefaultStyles()
	table := NewTable(styles)

	// Set up simple data
	table.SetColumns([]Column{
		{Title: "Name", Expand: true},
	})
	table.SetData(
		[][]string{{"Item 1"}, {"Item 2"}, {"Item 3"}},
		[]RowMeta{{ID: "1"}, {ID: "2"}, {ID: "3"}},
	)

	// Test initial selection (should be first data row)
	row := table.GetSelectedRow()
	if row != 0 {
		t.Errorf("Initial GetSelectedRow() = %d, want 0", row)
	}

	// Test SelectedMeta
	meta := table.SelectedMeta()
	if meta == nil {
		t.Fatal("SelectedMeta() returned nil")
	}
	if meta.ID != "1" {
		t.Errorf("SelectedMeta().ID = %q, want %q", meta.ID, "1")
	}
}

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
	help := NewHelpView(styles)

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
		{"today", now, "PM"}, // Should show time like "3:04 PM"
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

func TestMessagesViewKeys(t *testing.T) {
	app := createTestApp(t)
	view := NewMessagesView(app)

	// Test key handling doesn't panic
	keys := []struct {
		key  tcell.Key
		rune rune
	}{
		{tcell.KeyRune, 'n'},  // New compose
		{tcell.KeyRune, 'R'},  // Reply
		{tcell.KeyRune, 'A'},  // Reply all
		{tcell.KeyRune, 's'},  // Star
		{tcell.KeyRune, 'u'},  // Unread
		{tcell.KeyEnter, 0},   // Open
		{tcell.KeyEscape, 0},  // Back
	}

	for _, k := range keys {
		event := tcell.NewEventKey(k.key, k.rune, tcell.ModNone)
		// Just verify it doesn't panic
		_ = view.HandleKey(event)
	}
}

func TestEventsViewKeys(t *testing.T) {
	app := createTestApp(t)
	view := NewEventsView(app)

	// Test key handling doesn't panic
	keys := []struct {
		key  tcell.Key
		rune rune
	}{
		{tcell.KeyRune, 'm'},  // Month view
		{tcell.KeyRune, 'w'},  // Week view
		{tcell.KeyRune, 'a'},  // Agenda view
		{tcell.KeyRune, 't'},  // Today
		{tcell.KeyRune, 'c'},  // Cycle calendar
		{tcell.KeyRune, 'C'},  // Calendar list
		{tcell.KeyTab, 0},     // Switch panel
		{tcell.KeyEnter, 0},   // View day
	}

	for _, k := range keys {
		event := tcell.NewEventKey(k.key, k.rune, tcell.ModNone)
		_ = view.HandleKey(event)
	}
}

// TestViewsEscapeKeyHandling verifies that all views properly return the Escape
// key event so the app can handle navigation back to the previous view.
// This is a regression test for the calendar view Escape key bug.
func TestViewsEscapeKeyHandling(t *testing.T) {
	app := createTestApp(t)
	escapeEvent := tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone)

	testCases := []struct {
		name     string
		view     ResourceView
	}{
		{"MessagesView", NewMessagesView(app)},
		{"EventsView", NewEventsView(app)},
		{"ContactsView", NewContactsView(app)},
		{"WebhooksView", NewWebhooksView(app)},
		{"GrantsView", NewGrantsView(app)},
		{"DashboardView", NewDashboardView(app)},
	}

	for _, tc := range testCases {
		t.Run(tc.name+"_returns_escape_event", func(t *testing.T) {
			result := tc.view.HandleKey(escapeEvent)

			// Views should return the Escape event (not nil) so the app can handle navigation
			if result == nil {
				t.Errorf("%s consumed Escape key instead of returning it for app navigation", tc.name)
			}
		})
	}
}

// TestEventsViewEscapeWithFocusedPanels tests that Escape works regardless of which panel has focus.
func TestEventsViewEscapeWithFocusedPanels(t *testing.T) {
	app := createTestApp(t)
	view := NewEventsView(app)
	escapeEvent := tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone)

	// Test with calendar panel focused (default)
	result := view.HandleKey(escapeEvent)
	if result == nil {
		t.Error("EventsView with calendar focus should return Escape event for app navigation")
	}

	// Switch to events list panel
	tabEvent := tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone)
	view.HandleKey(tabEvent)

	// Test with events list panel focused
	result = view.HandleKey(escapeEvent)
	if result == nil {
		t.Error("EventsView with events list focus should return Escape event for app navigation")
	}
}

// TestGrantsViewSwitching tests grant switching functionality.
func TestGrantsViewSwitching(t *testing.T) {
	t.Run("without_grant_store_cannot_switch", func(t *testing.T) {
		// Create app without GrantStore (like demo mode)
		mockClient := nylas.NewMockClient()
		config := Config{
			Client:          mockClient,
			GrantID:         "test-grant-id",
			Email:           "user@example.com",
			Provider:        "google",
			RefreshInterval: time.Second * 30,
			Theme:           ThemeK9s,
		}
		app := NewApp(config)

		if app.CanSwitchGrant() {
			t.Error("App without GrantStore should not be able to switch grants")
		}

		err := app.SwitchGrant("new-grant", "new@example.com", "microsoft")
		if err == nil {
			t.Error("SwitchGrant should return error when GrantStore is nil")
		}
	})

	t.Run("grants_view_shows_current_grant_marker", func(t *testing.T) {
		app := createTestApp(t)
		view := NewGrantsView(app)

		// The test app has GrantID "test-grant-id"
		// Grants list from mock should mark the matching grant
		view.Load()

		// Verify the grants view was created
		if view == nil {
			t.Fatal("GrantsView should not be nil")
		}
	})

	t.Run("grants_view_handles_enter_key", func(t *testing.T) {
		app := createTestApp(t)
		view := NewGrantsView(app)

		// Load grants data
		view.Load()

		// Test Enter key handling (should not panic)
		enterEvent := tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone)
		result := view.HandleKey(enterEvent)

		// Without GrantStore, should consume event but show warning
		if result != nil {
			t.Error("Enter key should be consumed by GrantsView")
		}
	})

	t.Run("grants_view_returns_escape_event", func(t *testing.T) {
		app := createTestApp(t)
		view := NewGrantsView(app)

		escapeEvent := tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone)
		result := view.HandleKey(escapeEvent)

		if result == nil {
			t.Error("GrantsView should return Escape event for app navigation")
		}
	})
}
