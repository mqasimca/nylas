package components

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/mqasimca/nylas/internal/tui2/styles"
)

func TestNewSearchDialog(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewSearchDialog(theme)

	if dialog == nil {
		t.Fatal("expected dialog to be created")
	}
	if !dialog.IsVisible() {
		t.Error("expected dialog to be visible by default")
	}
	if dialog.focusedField != SearchDialogFieldFrom {
		t.Errorf("expected first field to be From, got %d", dialog.focusedField)
	}
}

func TestSearchDialog_BuildQuery_Empty(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewSearchDialog(theme)

	query := dialog.BuildQuery()
	if query != "" {
		t.Errorf("expected empty query, got %q", query)
	}
}

func TestSearchDialog_BuildQuery_SingleField(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewSearchDialog(theme)

	dialog.fromInput.SetValue("alice@example.com")
	query := dialog.BuildQuery()

	if query != "from:alice@example.com" {
		t.Errorf("expected 'from:alice@example.com', got %q", query)
	}
}

func TestSearchDialog_BuildQuery_MultipleFields(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewSearchDialog(theme)

	dialog.fromInput.SetValue("alice@example.com")
	dialog.subjectInput.SetValue("meeting")
	dialog.hasInput.SetValue("attachment")

	query := dialog.BuildQuery()

	if !strings.Contains(query, "from:alice@example.com") {
		t.Error("expected query to contain from:")
	}
	if !strings.Contains(query, "subject:meeting") {
		t.Error("expected query to contain subject:")
	}
	if !strings.Contains(query, "has:attachment") {
		t.Error("expected query to contain has:")
	}
}

func TestSearchDialog_BuildQuery_QuotesSpaces(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewSearchDialog(theme)

	dialog.subjectInput.SetValue("team meeting notes")

	query := dialog.BuildQuery()
	expected := `subject:"team meeting notes"`

	if query != expected {
		t.Errorf("expected %q, got %q", expected, query)
	}
}

func TestSearchDialog_BuildQuery_Toggles(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewSearchDialog(theme)

	// Test unread toggle
	t.Run("unread true", func(t *testing.T) {
		unread := true
		dialog.unread = &unread
		dialog.starred = nil

		query := dialog.BuildQuery()
		if query != "is:unread" {
			t.Errorf("expected 'is:unread', got %q", query)
		}
	})

	t.Run("unread false", func(t *testing.T) {
		read := false
		dialog.unread = &read
		dialog.starred = nil

		query := dialog.BuildQuery()
		if query != "is:read" {
			t.Errorf("expected 'is:read', got %q", query)
		}
	})

	t.Run("starred true", func(t *testing.T) {
		dialog.unread = nil
		starred := true
		dialog.starred = &starred

		query := dialog.BuildQuery()
		if query != "is:starred" {
			t.Errorf("expected 'is:starred', got %q", query)
		}
	})
}

func TestSearchDialog_BuildQuery_DateFilters(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewSearchDialog(theme)

	dialog.afterInput.SetValue("2024-01-01")
	dialog.beforeInput.SetValue("2024-12-31")

	query := dialog.BuildQuery()

	if !strings.Contains(query, "after:2024-01-01") {
		t.Error("expected query to contain after:")
	}
	if !strings.Contains(query, "before:2024-12-31") {
		t.Error("expected query to contain before:")
	}
}

func TestSearchDialog_BuildQuery_FreeText(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewSearchDialog(theme)

	dialog.fromInput.SetValue("alice@example.com")
	dialog.textInput.SetValue("important")

	query := dialog.BuildQuery()

	// Free text should be at the end
	if !strings.HasSuffix(query, "important") {
		t.Errorf("expected query to end with free text, got %q", query)
	}
	if !strings.HasPrefix(query, "from:") {
		t.Error("expected query to start with from:")
	}
}

func TestSearchDialog_ToggleUnread(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewSearchDialog(theme)

	// Initial state: nil (any)
	if dialog.unread != nil {
		t.Error("expected unread to be nil initially")
	}

	// First toggle: true (unread)
	dialog.toggleUnread()
	if dialog.unread == nil || !*dialog.unread {
		t.Error("expected unread to be true after first toggle")
	}

	// Second toggle: false (read)
	dialog.toggleUnread()
	if dialog.unread == nil || *dialog.unread {
		t.Error("expected unread to be false after second toggle")
	}

	// Third toggle: nil (any)
	dialog.toggleUnread()
	if dialog.unread != nil {
		t.Error("expected unread to be nil after third toggle")
	}
}

func TestSearchDialog_ToggleStarred(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewSearchDialog(theme)

	// Initial state: nil (any)
	if dialog.starred != nil {
		t.Error("expected starred to be nil initially")
	}

	// First toggle: true (starred)
	dialog.toggleStarred()
	if dialog.starred == nil || !*dialog.starred {
		t.Error("expected starred to be true after first toggle")
	}

	// Second toggle: false (not starred)
	dialog.toggleStarred()
	if dialog.starred == nil || *dialog.starred {
		t.Error("expected starred to be false after second toggle")
	}

	// Third toggle: nil (any)
	dialog.toggleStarred()
	if dialog.starred != nil {
		t.Error("expected starred to be nil after third toggle")
	}
}

func TestSearchDialog_Update_TabNavigation(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewSearchDialog(theme)

	// Tab should move to next field
	tabKey := tea.KeyPressMsg{Code: tea.KeyTab}
	dialog, _ = dialog.Update(tabKey)

	if dialog.focusedField != SearchDialogFieldTo {
		t.Errorf("expected focus on To after tab, got %d", dialog.focusedField)
	}

	// Tab again
	dialog, _ = dialog.Update(tabKey)
	if dialog.focusedField != SearchDialogFieldSubject {
		t.Errorf("expected focus on Subject after tab, got %d", dialog.focusedField)
	}
}

func TestSearchDialog_Update_ShiftTabNavigation(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewSearchDialog(theme)

	// Move to To field first
	dialog.focusedField = SearchDialogFieldTo
	dialog.focusCurrent()

	// Shift+Tab should go back to From
	shiftTabKey := tea.KeyPressMsg{Text: "shift+tab"}
	dialog, _ = dialog.Update(shiftTabKey)

	if dialog.focusedField != SearchDialogFieldFrom {
		t.Errorf("expected focus on From after shift+tab, got %d", dialog.focusedField)
	}
}

func TestSearchDialog_Update_EscapeCancel(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewSearchDialog(theme)

	escKey := tea.KeyPressMsg{Code: tea.KeyEsc}
	_, cmd := dialog.Update(escKey)

	if cmd == nil {
		t.Fatal("expected cmd to be returned")
	}

	msg := cmd()
	if _, ok := msg.(SearchDialogCancelMsg); !ok {
		t.Errorf("expected SearchDialogCancelMsg, got %T", msg)
	}

	if dialog.IsVisible() {
		t.Error("expected dialog to be hidden after escape")
	}
}

func TestSearchDialog_Update_EnterSubmit(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewSearchDialog(theme)

	// Set some search criteria
	dialog.fromInput.SetValue("test@example.com")

	// Focus on search button
	dialog.focusedField = SearchDialogFieldSearch

	enterKey := tea.KeyPressMsg{Code: tea.KeyEnter}
	_, cmd := dialog.Update(enterKey)

	if cmd == nil {
		t.Fatal("expected cmd to be returned")
	}

	msg := cmd()
	result, ok := msg.(SearchDialogSubmitMsg)
	if !ok {
		t.Fatalf("expected SearchDialogSubmitMsg, got %T", msg)
	}

	if result.Query != "from:test@example.com" {
		t.Errorf("expected query 'from:test@example.com', got %q", result.Query)
	}
}

func TestSearchDialog_Update_EnterToggle(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewSearchDialog(theme)

	// Focus on unread field
	dialog.focusedField = SearchDialogFieldUnread

	// Press enter to toggle
	enterKey := tea.KeyPressMsg{Code: tea.KeyEnter}
	dialog, _ = dialog.Update(enterKey)

	if dialog.unread == nil || !*dialog.unread {
		t.Error("expected unread to be true after enter on unread field")
	}
}

func TestSearchDialog_Update_SpaceToggle(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewSearchDialog(theme)

	// Focus on starred field
	dialog.focusedField = SearchDialogFieldStarred
	dialog.focusCurrent()

	// Press space to toggle (use KeySpace code)
	spaceKey := tea.KeyPressMsg{Code: tea.KeySpace}
	dialog, _ = dialog.Update(spaceKey)

	if dialog.starred == nil || !*dialog.starred {
		t.Error("expected starred to be true after space on starred field")
	}
}

func TestSearchDialog_View(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewSearchDialog(theme)
	dialog.SetSize(80, 40)

	view := dialog.View()

	if !strings.Contains(view, "Advanced Search") {
		t.Error("expected 'Advanced Search' in view")
	}
	if !strings.Contains(view, "From") {
		t.Error("expected 'From' field in view")
	}
	if !strings.Contains(view, "To") {
		t.Error("expected 'To' field in view")
	}
	if !strings.Contains(view, "Subject") {
		t.Error("expected 'Subject' field in view")
	}
	if !strings.Contains(view, "Search") {
		t.Error("expected 'Search' button in view")
	}
	if !strings.Contains(view, "Cancel") {
		t.Error("expected 'Cancel' button in view")
	}
}

func TestSearchDialog_View_Hidden(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewSearchDialog(theme)
	dialog.Hide()

	view := dialog.View()

	if view != "" {
		t.Error("expected empty view when dialog is hidden")
	}
}

func TestSearchDialog_View_QueryPreview(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewSearchDialog(theme)
	dialog.SetSize(80, 40)

	dialog.fromInput.SetValue("alice@example.com")

	view := dialog.View()

	if !strings.Contains(view, "Query:") {
		t.Error("expected query preview in view")
	}
	if !strings.Contains(view, "from:alice@example.com") {
		t.Error("expected query content in preview")
	}
}

func TestSearchDialog_ShowHide(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewSearchDialog(theme)

	dialog.Hide()
	if dialog.IsVisible() {
		t.Error("expected dialog to be hidden")
	}

	dialog.Show()
	if !dialog.IsVisible() {
		t.Error("expected dialog to be visible")
	}
}

func TestSearchDialog_Reset(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewSearchDialog(theme)

	// Set some values
	dialog.fromInput.SetValue("alice@example.com")
	dialog.subjectInput.SetValue("meeting")
	unread := true
	dialog.unread = &unread
	dialog.focusedField = SearchDialogFieldSubject

	// Reset
	dialog.Reset()

	if dialog.fromInput.Value() != "" {
		t.Error("expected from to be cleared")
	}
	if dialog.subjectInput.Value() != "" {
		t.Error("expected subject to be cleared")
	}
	if dialog.unread != nil {
		t.Error("expected unread to be nil")
	}
	if dialog.focusedField != SearchDialogFieldFrom {
		t.Error("expected focus to be reset to From")
	}
}

func TestSearchDialog_SetQuery(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewSearchDialog(theme)

	dialog.SetQuery("from:alice@example.com subject:meeting is:unread has:attachment after:2024-01-01")

	if dialog.fromInput.Value() != "alice@example.com" {
		t.Errorf("expected from 'alice@example.com', got %q", dialog.fromInput.Value())
	}
	if dialog.subjectInput.Value() != "meeting" {
		t.Errorf("expected subject 'meeting', got %q", dialog.subjectInput.Value())
	}
	if dialog.hasInput.Value() != "attachment" {
		t.Errorf("expected has 'attachment', got %q", dialog.hasInput.Value())
	}
	if dialog.afterInput.Value() != "2024-01-01" {
		t.Errorf("expected after '2024-01-01', got %q", dialog.afterInput.Value())
	}
	if dialog.unread == nil || !*dialog.unread {
		t.Error("expected unread to be true")
	}
}

func TestSearchDialog_Init(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewSearchDialog(theme)

	cmd := dialog.Init()
	if cmd == nil {
		t.Error("expected Init to return a command for text input blink")
	}
}

func TestSearchDialog_FolderField(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewSearchDialog(theme)

	dialog.inInput.SetValue("inbox")

	query := dialog.BuildQuery()
	if query != "in:inbox" {
		t.Errorf("expected 'in:inbox', got %q", query)
	}
}

func TestSearchDialog_Update_NotVisibleNoOp(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewSearchDialog(theme)
	dialog.Hide()

	tabKey := tea.KeyPressMsg{Code: tea.KeyTab}
	_, cmd := dialog.Update(tabKey)

	if cmd != nil {
		t.Error("expected no cmd when dialog is hidden")
	}
}
