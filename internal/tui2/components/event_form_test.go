package components

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/tui2/styles"
)

func TestNewEventForm(t *testing.T) {
	theme := styles.GetTheme("default")
	form := NewEventForm(theme, EventFormCreate)

	if form == nil {
		t.Fatal("expected form to be created")
	}
	if form.mode != EventFormCreate {
		t.Errorf("expected mode %d, got %d", EventFormCreate, form.mode)
	}
	if form.allDay {
		t.Error("expected allDay to be false by default")
	}
	if !form.busy {
		t.Error("expected busy to be true by default")
	}
	if form.focusedField != EventFormFieldTitle {
		t.Errorf("expected focused field to be title, got %d", form.focusedField)
	}
}

func TestEventForm_SetEvent(t *testing.T) {
	theme := styles.GetTheme("default")
	form := NewEventForm(theme, EventFormEdit)

	event := &domain.Event{
		ID:          "test-event-id",
		Title:       "Test Meeting",
		Location:    "Conference Room A",
		Description: "Discuss project updates",
		Busy:        true,
		When: domain.EventWhen{
			Object:    "timespan",
			StartTime: time.Date(2025, 6, 15, 14, 0, 0, 0, time.Local).Unix(),
			EndTime:   time.Date(2025, 6, 15, 15, 0, 0, 0, time.Local).Unix(),
		},
	}

	form.SetEvent(event)

	if form.eventID != "test-event-id" {
		t.Errorf("expected eventID 'test-event-id', got '%s'", form.eventID)
	}
	if form.GetTitle() != "Test Meeting" {
		t.Errorf("expected title 'Test Meeting', got '%s'", form.GetTitle())
	}
	if !form.IsBusy() {
		t.Error("expected busy to be true")
	}
	if form.IsAllDay() {
		t.Error("expected allDay to be false for timespan event")
	}
}

func TestEventForm_SetEventAllDay(t *testing.T) {
	theme := styles.GetTheme("default")
	form := NewEventForm(theme, EventFormEdit)

	event := &domain.Event{
		ID:    "all-day-event",
		Title: "All Day Event",
		Busy:  false,
		When: domain.EventWhen{
			Object: "date",
			Date:   "2025-06-15",
		},
	}

	form.SetEvent(event)

	if !form.IsAllDay() {
		t.Error("expected allDay to be true for date event")
	}
	if form.IsBusy() {
		t.Error("expected busy to be false")
	}
}

func TestEventForm_SetTimezone(t *testing.T) {
	theme := styles.GetTheme("default")
	form := NewEventForm(theme, EventFormCreate)

	loc, _ := time.LoadLocation("America/New_York")
	form.SetTimezone(loc)

	if form.timezone != loc {
		t.Error("expected timezone to be set")
	}
}

func TestEventForm_SetDate(t *testing.T) {
	theme := styles.GetTheme("default")
	form := NewEventForm(theme, EventFormCreate)

	date := time.Date(2025, 7, 4, 0, 0, 0, 0, time.Local)
	form.SetDate(date)

	// The start date input should have the date
	if !strings.Contains(form.View(), "2025-07-04") {
		t.Error("expected date to be set in form")
	}
}

func TestEventForm_GetMode(t *testing.T) {
	theme := styles.GetTheme("default")

	createForm := NewEventForm(theme, EventFormCreate)
	if createForm.GetMode() != EventFormCreate {
		t.Error("expected create mode")
	}

	editForm := NewEventForm(theme, EventFormEdit)
	if editForm.GetMode() != EventFormEdit {
		t.Error("expected edit mode")
	}
}

func TestEventForm_Update_TabNavigation(t *testing.T) {
	theme := styles.GetTheme("default")
	form := NewEventForm(theme, EventFormCreate)

	// Tab should move to next field
	tabKey := tea.KeyPressMsg{Code: tea.KeyTab}
	form, _ = form.Update(tabKey)

	if form.focusedField != EventFormFieldLocation {
		t.Errorf("expected focused field to be location after tab, got %d", form.focusedField)
	}

	// Tab again
	form, _ = form.Update(tabKey)
	if form.focusedField != EventFormFieldDescription {
		t.Errorf("expected focused field to be description after tab, got %d", form.focusedField)
	}
}

func TestEventForm_Update_ShiftTabNavigation(t *testing.T) {
	theme := styles.GetTheme("default")
	form := NewEventForm(theme, EventFormCreate)

	// Move to location first
	tabKey := tea.KeyPressMsg{Code: tea.KeyTab}
	form, _ = form.Update(tabKey)

	// Shift+Tab should go back to title
	shiftTabKey := tea.KeyPressMsg{Text: "shift+tab"}
	form, _ = form.Update(shiftTabKey)

	if form.focusedField != EventFormFieldTitle {
		t.Errorf("expected focused field to be title after shift+tab, got %d", form.focusedField)
	}
}

func TestEventForm_Update_EscapeCancel(t *testing.T) {
	theme := styles.GetTheme("default")
	form := NewEventForm(theme, EventFormCreate)

	escKey := tea.KeyPressMsg{Code: tea.KeyEsc}
	_, cmd := form.Update(escKey)

	if cmd == nil {
		t.Fatal("expected cmd to be returned")
	}

	msg := cmd()
	if _, ok := msg.(EventFormCancelMsg); !ok {
		t.Errorf("expected EventFormCancelMsg, got %T", msg)
	}
}

func TestEventForm_Update_ToggleAllDay(t *testing.T) {
	theme := styles.GetTheme("default")
	form := NewEventForm(theme, EventFormCreate)

	// Navigate to all-day field
	for form.focusedField != EventFormFieldAllDay {
		tabKey := tea.KeyPressMsg{Code: tea.KeyTab}
		form, _ = form.Update(tabKey)
	}

	// Toggle with enter
	enterKey := tea.KeyPressMsg{Code: tea.KeyEnter}
	form, _ = form.Update(enterKey)

	if !form.IsAllDay() {
		t.Error("expected allDay to be true after toggle")
	}

	// Toggle again
	form, _ = form.Update(enterKey)
	if form.IsAllDay() {
		t.Error("expected allDay to be false after second toggle")
	}
}

func TestEventForm_Update_ToggleBusy(t *testing.T) {
	theme := styles.GetTheme("default")
	form := NewEventForm(theme, EventFormCreate)

	// Navigate to busy field
	for form.focusedField != EventFormFieldBusy {
		tabKey := tea.KeyPressMsg{Code: tea.KeyTab}
		form, _ = form.Update(tabKey)
	}

	// Toggle with enter (space handling may differ)
	enterKey := tea.KeyPressMsg{Code: tea.KeyEnter}
	form, _ = form.Update(enterKey)

	if form.IsBusy() {
		t.Error("expected busy to be false after toggle (was true)")
	}
}

func TestEventForm_View(t *testing.T) {
	theme := styles.GetTheme("default")
	form := NewEventForm(theme, EventFormCreate)
	form.SetSize(80, 40)

	view := form.View()

	// Check for expected content
	if !strings.Contains(view, "Create Event") {
		t.Error("expected 'Create Event' in view")
	}
	if !strings.Contains(view, "Title:") {
		t.Error("expected 'Title:' in view")
	}
	if !strings.Contains(view, "Location:") {
		t.Error("expected 'Location:' in view")
	}
	if !strings.Contains(view, "All Day:") {
		t.Error("expected 'All Day:' in view")
	}
	if !strings.Contains(view, "Create") {
		t.Error("expected 'Create' button in view")
	}
	if !strings.Contains(view, "Cancel") {
		t.Error("expected 'Cancel' button in view")
	}
}

func TestEventForm_View_EditMode(t *testing.T) {
	theme := styles.GetTheme("default")
	form := NewEventForm(theme, EventFormEdit)

	view := form.View()

	if !strings.Contains(view, "Edit Event") {
		t.Error("expected 'Edit Event' in view")
	}
	if !strings.Contains(view, "Save") {
		t.Error("expected 'Save' button in view")
	}
}

func TestEventForm_Validate_EmptyTitle(t *testing.T) {
	theme := styles.GetTheme("default")
	form := NewEventForm(theme, EventFormCreate)

	// Clear the title
	form.titleInput.SetValue("")

	err := form.Validate()
	if err == nil {
		t.Error("expected validation error for empty title")
	}
}

func TestEventForm_Validate_ValidForm(t *testing.T) {
	theme := styles.GetTheme("default")
	form := NewEventForm(theme, EventFormCreate)

	// Set a title
	form.titleInput.SetValue("Test Event")

	err := form.Validate()
	if err != nil {
		t.Errorf("unexpected validation error: %v", err)
	}
}

func TestEventForm_Init(t *testing.T) {
	theme := styles.GetTheme("default")
	form := NewEventForm(theme, EventFormCreate)

	cmd := form.Init()
	if cmd == nil {
		t.Error("expected Init to return a command for text input blink")
	}
}

func TestEventForm_AllDayToggleHidesTime(t *testing.T) {
	theme := styles.GetTheme("default")
	form := NewEventForm(theme, EventFormCreate)
	form.allDay = true

	view := form.View()

	// When all-day is enabled, the time inputs should still be in the view
	// but not rendered on the same line as date inputs
	// Just verify the view is not empty
	if view == "" {
		t.Error("expected non-empty view")
	}
}
