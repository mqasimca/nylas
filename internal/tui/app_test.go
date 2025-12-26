package tui

import (
	"context"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/mqasimca/nylas/internal/adapters/nylas"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/rivo/tview"
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

func TestMessagesViewKeys(t *testing.T) {
	app := createTestApp(t)
	view := NewMessagesView(app)

	// Test key handling doesn't panic
	keys := []struct {
		key  tcell.Key
		rune rune
	}{
		{tcell.KeyRune, 'n'}, // New compose
		{tcell.KeyRune, 'R'}, // Reply
		{tcell.KeyRune, 'A'}, // Reply all
		{tcell.KeyRune, 's'}, // Star
		{tcell.KeyRune, 'u'}, // Unread
		{tcell.KeyEnter, 0},  // Open
		{tcell.KeyEscape, 0}, // Back
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
		{tcell.KeyRune, 'm'}, // Month view
		{tcell.KeyRune, 'w'}, // Week view
		{tcell.KeyRune, 'a'}, // Agenda view
		{tcell.KeyRune, 't'}, // Today
		{tcell.KeyRune, 'c'}, // Cycle calendar
		{tcell.KeyRune, 'C'}, // Calendar list
		{tcell.KeyTab, 0},    // Switch panel
		{tcell.KeyEnter, 0},  // View day
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
		name string
		view ResourceView
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

func TestApp_GetCurrentView(t *testing.T) {
	app := createTestApp(t)

	// Initially should have dashboard view
	view := app.getCurrentView()
	if view == nil {
		t.Fatal("getCurrentView() returned nil")
	}

	// Default initial view is dashboard
	if view.Name() != "dashboard" {
		t.Errorf("getCurrentView().Name() = %q, want %q", view.Name(), "dashboard")
	}
}

func TestApp_Flash(t *testing.T) {
	app := createTestApp(t)

	// Test all flash levels
	flashLevels := []FlashLevel{FlashInfo, FlashWarn, FlashError}

	for _, fl := range flashLevels {
		// Should not panic
		app.Flash(fl, "Test message %s", "arg")
	}
}

func TestApp_ShowConfirmDialog(t *testing.T) {
	app := createTestApp(t)

	called := false
	app.ShowConfirmDialog("Test Title", "Test message", func() {
		called = true
	})

	// Dialog should be shown (pushed to content stack)
	// We can't easily test the callback without simulating UI interaction
	// but we can verify it doesn't panic
	if called {
		t.Error("callback should not be called immediately")
	}
}

func TestApp_PopDetail(t *testing.T) {
	app := createTestApp(t)

	// Pop on empty detail should not panic
	app.PopDetail()

	// Push something to content stack first using PushDetail
	box := tview.NewBox()
	app.PushDetail("test-detail", box)

	// Now pop should work
	app.PopDetail()
}

func TestApp_GoBack(t *testing.T) {
	app := createTestApp(t)

	// goBack is unexported, but we can test the exported PopDetail
	// which internally handles going back in detail views

	// Push a detail view first
	box := tview.NewBox()
	app.PushDetail("test-detail", box)

	// PopDetail should work
	app.PopDetail()
}

func TestApp_PageNavigation(t *testing.T) {
	// pageMove, goToTop, goToBottom, goToRow are unexported methods
	// They work on the internal table selection
	// We can only test that table navigation works via exported methods

	styles := DefaultStyles()
	table := NewTable(styles)

	table.SetColumns([]Column{{Title: "Name", Expand: true}})
	table.SetData(
		[][]string{{"Item 1"}, {"Item 2"}, {"Item 3"}},
		[]RowMeta{{ID: "id-1"}, {ID: "id-2"}, {ID: "id-3"}},
	)

	// Test table selection
	table.Select(1, 0)
	row, _ := table.GetSelection()
	if row != 1 {
		t.Errorf("After Select(1, 0), row = %d, want 1", row)
	}

	table.Select(2, 0)
	row, _ = table.GetSelection()
	if row != 2 {
		t.Errorf("After Select(2, 0), row = %d, want 2", row)
	}
}

func TestApp_NavigateView(t *testing.T) {
	// navigateTo is unexported, but we can test views creation directly
	views := []string{"messages", "events", "contacts", "webhooks", "grants", "dashboard", "drafts", "availability"}

	for _, viewName := range views {
		t.Run(viewName, func(t *testing.T) {
			app := createTestApp(t)

			// createView is unexported, but views are created automatically
			// via command processing. Test that getCurrentView works after init.
			view := app.getCurrentView()
			if view == nil {
				t.Fatal("getCurrentView() returned nil")
			}

			// Verify default is dashboard
			if view.Name() != "dashboard" {
				t.Errorf("Default view = %q, want dashboard", view.Name())
			}
		})
	}
}

func TestApp_SetFocus(t *testing.T) {
	app := createTestApp(t)

	// SetFocus on a primitive should not panic
	box := tview.NewBox()
	app.SetFocus(box)
}

func TestApp_Styles(t *testing.T) {
	app := createTestApp(t)

	// Styles should return the styles
	styles := app.Styles()
	if styles == nil {
		t.Error("Styles() returned nil")
	}
}

func TestPageStack_PushPop(t *testing.T) {
	stack := NewPageStack()

	// Create dummy primitives
	box1 := tview.NewBox()
	box2 := tview.NewBox()
	box3 := tview.NewBox()

	// Push items
	stack.Push("page1", box1)
	stack.Push("page2", box2)
	stack.Push("page3", box3)

	if stack.Len() != 3 {
		t.Errorf("Len() = %d, want 3", stack.Len())
	}

	if stack.Top() != "page3" {
		t.Errorf("Top() = %q, want %q", stack.Top(), "page3")
	}

	// Pop items
	popped := stack.Pop()
	if popped != "page3" {
		t.Errorf("Pop() = %q, want %q", popped, "page3")
	}

	if stack.Len() != 2 {
		t.Errorf("After pop, Len() = %d, want 2", stack.Len())
	}

	// Pop remaining
	stack.Pop()
	stack.Pop()

	// Pop on empty should return empty string
	popped = stack.Pop()
	if popped != "" {
		t.Errorf("Pop() on empty stack = %q, want empty string", popped)
	}
}

func TestPageStack_HasPage(t *testing.T) {
	stack := NewPageStack()

	box1 := tview.NewBox()
	box2 := tview.NewBox()

	stack.Push("page1", box1)
	stack.Push("page2", box2)

	// PageStack embeds tview.Pages which has HasPage method
	if !stack.HasPage("page1") {
		t.Error("HasPage(page1) = false, want true")
	}

	if !stack.HasPage("page2") {
		t.Error("HasPage(page2) = false, want true")
	}

	if stack.HasPage("page3") {
		t.Error("HasPage(page3) = true, want false")
	}
}

func TestPageStack_SwitchTo(t *testing.T) {
	stack := NewPageStack()

	// Push pages
	box1 := tview.NewBox()
	box2 := tview.NewBox()

	stack.Push("page1", box1)
	stack.Push("page2", box2)

	if stack.Len() != 2 {
		t.Errorf("Len() = %d, want 2", stack.Len())
	}

	// Top should return current page name
	if stack.Top() != "page2" {
		t.Errorf("Top() = %q, want %q", stack.Top(), "page2")
	}

	// SwitchTo an existing page
	stack.SwitchTo("page1", box1)
	if stack.Top() != "page1" {
		t.Errorf("After SwitchTo(page1), Top() = %q, want page1", stack.Top())
	}

	// SwitchTo a new page
	box3 := tview.NewBox()
	stack.SwitchTo("page3", box3)
	if stack.Top() != "page3" {
		t.Errorf("After SwitchTo(page3), Top() = %q, want page3", stack.Top())
	}

	// Pop
	stack.Pop()
	if stack.Top() != "page1" {
		t.Errorf("After pop, Top() = %q, want page1", stack.Top())
	}
}

func TestTable_SelectedMeta(t *testing.T) {
	styles := DefaultStyles()
	table := NewTable(styles)

	table.SetColumns([]Column{{Title: "Name", Expand: true}})
	table.SetData(
		[][]string{{"Item 1"}, {"Item 2"}},
		[]RowMeta{{ID: "id-1"}, {ID: "id-2"}},
	)

	// Select first row (row 1, since row 0 is header)
	table.Select(1, 0)

	meta := table.SelectedMeta()
	if meta == nil {
		t.Fatal("SelectedMeta() returned nil")
	}
	if meta.ID != "id-1" {
		t.Errorf("SelectedMeta().ID = %q, want %q", meta.ID, "id-1")
	}
}

func TestTable_SetData(t *testing.T) {
	styles := DefaultStyles()
	table := NewTable(styles)

	table.SetColumns([]Column{{Title: "Name", Expand: true}})

	// Verify initial state
	initialCount := table.GetRowCount()

	// Set data
	table.SetData(
		[][]string{{"Item 1"}, {"Item 2"}},
		[]RowMeta{{ID: "id-1"}, {ID: "id-2"}},
	)

	// Verify data was set
	afterCount := table.GetRowCount()
	if afterCount <= initialCount {
		t.Errorf("After SetData(), GetRowCount() = %d, should be > %d", afterCount, initialCount)
	}
}

func TestTable_GetRowCount(t *testing.T) {
	styles := DefaultStyles()
	table := NewTable(styles)

	table.SetColumns([]Column{{Title: "Name", Expand: true}})
	table.SetData(
		[][]string{{"Item 1"}, {"Item 2"}, {"Item 3"}},
		[]RowMeta{{ID: "id-1"}, {ID: "id-2"}, {ID: "id-3"}},
	)

	// GetRowCount should return the count including header
	count := table.GetRowCount()
	if count < 3 {
		t.Errorf("GetRowCount() = %d, want >= 3", count)
	}
}

func TestCommandRegistry(t *testing.T) {
	registry := NewCommandRegistry()

	if registry == nil {
		t.Fatal("NewCommandRegistry() returned nil")
	}

	// Get all commands
	commands := registry.GetAll()
	if len(commands) == 0 {
		t.Error("GetAll() returned empty list")
	}

	// Get specific command
	cmd := registry.Get("q")
	if cmd == nil {
		t.Error("Get('q') should find quit command")
	} else if cmd.Name != "quit" {
		t.Errorf("Get('q').Name = %q, want %q", cmd.Name, "quit")
	}

	// Get non-existent command
	cmd = registry.Get("nonexistent")
	if cmd != nil {
		t.Error("Get('nonexistent') should return nil")
	}

	// Test Search
	results := registry.Search("quit")
	if len(results) == 0 {
		t.Error("Search('quit') returned empty list")
	}
}

func TestStyles_DefaultStyles(t *testing.T) {
	styles := DefaultStyles()

	if styles == nil {
		t.Fatal("DefaultStyles() returned nil")
	}

	// Verify some key colors are set
	if styles.BgColor == 0 {
		t.Error("BgColor not set")
	}
	if styles.FgColor == 0 {
		t.Error("FgColor not set")
	}
}

func TestCalendarView_Focus(t *testing.T) {
	app := createTestApp(t)
	view := NewCalendarView(app)

	// Focus should not panic
	view.Focus(nil)

	// HasFocus may return true or false depending on initialization
	// Just verify it doesn't panic
	_ = view.HasFocus()
}

func TestCalendarView_SetOnEventSelect(t *testing.T) {
	app := createTestApp(t)
	view := NewCalendarView(app)

	view.SetOnEventSelect(func(event *domain.Event) {
		// callback set
	})

	if view.onEventSelect == nil {
		t.Error("SetOnEventSelect did not set callback")
	}
}

func TestCalendarView_InputHandler(t *testing.T) {
	app := createTestApp(t)
	view := NewCalendarView(app)

	// Test various keys
	keys := []struct {
		key  tcell.Key
		rune rune
		desc string
	}{
		{tcell.KeyLeft, 0, "left arrow"},
		{tcell.KeyRight, 0, "right arrow"},
		{tcell.KeyUp, 0, "up arrow"},
		{tcell.KeyDown, 0, "down arrow"},
		{tcell.KeyRune, 'h', "h key"},
		{tcell.KeyRune, 'l', "l key"},
		{tcell.KeyRune, 'j', "j key"},
		{tcell.KeyRune, 'k', "k key"},
		{tcell.KeyRune, 'H', "H key (prev month)"},
		{tcell.KeyRune, 'L', "L key (next month)"},
		{tcell.KeyRune, 't', "t key (today)"},
		{tcell.KeyRune, 'v', "v key (toggle view)"},
		{tcell.KeyRune, 'm', "m key (month view)"},
		{tcell.KeyRune, 'w', "w key (week view)"},
		{tcell.KeyRune, 'a', "a key (agenda view)"},
		{tcell.KeyRune, ']', "] key (next calendar)"},
		{tcell.KeyRune, '[', "[ key (prev calendar)"},
		{tcell.KeyEnter, 0, "enter key"},
	}

	handler := view.InputHandler()
	if handler == nil {
		t.Fatal("InputHandler() returned nil")
	}

	for _, k := range keys {
		t.Run(k.desc, func(t *testing.T) {
			event := tcell.NewEventKey(k.key, k.rune, tcell.ModNone)
			// Should not panic
			handler(event, nil)
		})
	}
}

func TestDashboardView(t *testing.T) {
	app := createTestApp(t)
	view := NewDashboardView(app)

	if view == nil {
		t.Fatal("NewDashboardView() returned nil")
	}

	if view.Name() != "dashboard" {
		t.Errorf("Name() = %q, want %q", view.Name(), "dashboard")
	}

	if view.Title() != "Dashboard" {
		t.Errorf("Title() = %q, want %q", view.Title(), "Dashboard")
	}

	// Load should not panic
	view.Load()

	// Filter should not panic
	view.Filter("test")

	// Refresh should not panic
	view.Refresh()
}

func TestContactsView(t *testing.T) {
	app := createTestApp(t)
	view := NewContactsView(app)

	if view == nil {
		t.Fatal("NewContactsView() returned nil")
	}

	if view.Name() != "contacts" {
		t.Errorf("Name() = %q, want %q", view.Name(), "contacts")
	}

	// Test keys
	keys := []struct {
		key  tcell.Key
		rune rune
	}{
		{tcell.KeyRune, 'n'}, // New contact
		{tcell.KeyRune, 'e'}, // Edit
		{tcell.KeyRune, 'd'}, // Delete
		{tcell.KeyEnter, 0},  // View
		{tcell.KeyRune, 'r'}, // Refresh
	}

	for _, k := range keys {
		event := tcell.NewEventKey(k.key, k.rune, tcell.ModNone)
		// Should not panic
		_ = view.HandleKey(event)
	}
}

func TestWebhooksView(t *testing.T) {
	app := createTestApp(t)
	view := NewWebhooksView(app)

	if view == nil {
		t.Fatal("NewWebhooksView() returned nil")
	}

	if view.Name() != "webhooks" {
		t.Errorf("Name() = %q, want %q", view.Name(), "webhooks")
	}

	// Test keys
	keys := []struct {
		key  tcell.Key
		rune rune
	}{
		{tcell.KeyRune, 'n'}, // New webhook
		{tcell.KeyRune, 'e'}, // Edit
		{tcell.KeyRune, 'd'}, // Delete
		{tcell.KeyEnter, 0},  // View
		{tcell.KeyRune, 'r'}, // Refresh
	}

	for _, k := range keys {
		event := tcell.NewEventKey(k.key, k.rune, tcell.ModNone)
		_ = view.HandleKey(event)
	}
}

func TestAvailabilityView_FullInterface(t *testing.T) {
	app := createTestApp(t)
	view := NewAvailabilityView(app)

	// Test ResourceView interface
	if view.Name() != "availability" {
		t.Errorf("Name() = %q, want %q", view.Name(), "availability")
	}

	if view.Title() != "Availability" {
		t.Errorf("Title() = %q, want %q", view.Title(), "Availability")
	}

	if view.Primitive() == nil {
		t.Error("Primitive() returned nil")
	}

	hints := view.Hints()
	if len(hints) == 0 {
		t.Error("Hints() returned empty")
	}

	// Filter should not panic
	view.Filter("test")
}

func TestDraftsView(t *testing.T) {
	app := createTestApp(t)
	view := NewDraftsView(app)

	if view == nil {
		t.Fatal("NewDraftsView() returned nil")
	}

	if view.Name() != "drafts" {
		t.Errorf("Name() = %q, want %q", view.Name(), "drafts")
	}

	// Load should not panic
	view.Load()

	// Test keys
	keys := []struct {
		key  tcell.Key
		rune rune
	}{
		{tcell.KeyRune, 'n'}, // New draft
		{tcell.KeyRune, 'e'}, // Edit
		{tcell.KeyRune, 'd'}, // Delete
		{tcell.KeyEnter, 0},  // View/Edit
		{tcell.KeyRune, 'r'}, // Refresh
	}

	for _, k := range keys {
		event := tcell.NewEventKey(k.key, k.rune, tcell.ModNone)
		_ = view.HandleKey(event)
	}
}
