package components

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/tui2/styles"
)

func TestNewAvailabilityDialog(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewAvailabilityDialog(theme)

	if dialog == nil {
		t.Fatal("expected dialog to be created")
	}
	if !dialog.IsVisible() {
		t.Error("expected dialog to be visible by default")
	}
	if dialog.focusedField != AvailabilityFieldParticipants {
		t.Errorf("expected first field to be Participants, got %d", dialog.focusedField)
	}
}

func TestAvailabilityDialog_DefaultValues(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewAvailabilityDialog(theme)

	// Check default duration
	if dialog.durationInput.Value() != "30" {
		t.Errorf("expected default duration '30', got %q", dialog.durationInput.Value())
	}

	// Check that start date is set to today
	today := time.Now().Format("2006-01-02")
	if dialog.startDateInput.Value() != today {
		t.Errorf("expected start date %q, got %q", today, dialog.startDateInput.Value())
	}

	// Check that end date is set to 7 days from now
	endDate := time.Now().AddDate(0, 0, 7).Format("2006-01-02")
	if dialog.endDateInput.Value() != endDate {
		t.Errorf("expected end date %q, got %q", endDate, dialog.endDateInput.Value())
	}
}

func TestAvailabilityDialog_BuildRequest_Valid(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewAvailabilityDialog(theme)

	dialog.participantsInput.SetValue("alice@example.com, bob@example.com")
	dialog.startDateInput.SetValue("2025-01-15")
	dialog.endDateInput.SetValue("2025-01-22")
	dialog.durationInput.SetValue("60")

	req, err := dialog.buildRequest()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(req.Participants) != 2 {
		t.Errorf("expected 2 participants, got %d", len(req.Participants))
	}
	if req.Participants[0].Email != "alice@example.com" {
		t.Errorf("expected first participant 'alice@example.com', got %q", req.Participants[0].Email)
	}
	if req.DurationMinutes != 60 {
		t.Errorf("expected duration 60, got %d", req.DurationMinutes)
	}
}

func TestAvailabilityDialog_BuildRequest_EmptyParticipants(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewAvailabilityDialog(theme)

	dialog.participantsInput.SetValue("")

	_, err := dialog.buildRequest()
	if err == nil {
		t.Error("expected error for empty participants")
	}
	if !strings.Contains(err.Error(), "participants") {
		t.Errorf("expected error about participants, got: %v", err)
	}
}

func TestAvailabilityDialog_BuildRequest_InvalidStartDate(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewAvailabilityDialog(theme)

	dialog.participantsInput.SetValue("alice@example.com")
	dialog.startDateInput.SetValue("invalid")
	dialog.endDateInput.SetValue("2025-01-22")

	_, err := dialog.buildRequest()
	if err == nil {
		t.Error("expected error for invalid start date")
	}
	if !strings.Contains(err.Error(), "start date") {
		t.Errorf("expected error about start date, got: %v", err)
	}
}

func TestAvailabilityDialog_BuildRequest_InvalidEndDate(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewAvailabilityDialog(theme)

	dialog.participantsInput.SetValue("alice@example.com")
	dialog.startDateInput.SetValue("2025-01-15")
	dialog.endDateInput.SetValue("invalid")

	_, err := dialog.buildRequest()
	if err == nil {
		t.Error("expected error for invalid end date")
	}
	if !strings.Contains(err.Error(), "end date") {
		t.Errorf("expected error about end date, got: %v", err)
	}
}

func TestAvailabilityDialog_BuildRequest_EndBeforeStart(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewAvailabilityDialog(theme)

	dialog.participantsInput.SetValue("alice@example.com")
	dialog.startDateInput.SetValue("2025-01-22")
	dialog.endDateInput.SetValue("2025-01-15")

	_, err := dialog.buildRequest()
	if err == nil {
		t.Error("expected error when end date is before start date")
	}
	if !strings.Contains(err.Error(), "after") {
		t.Errorf("expected error about date order, got: %v", err)
	}
}

func TestAvailabilityDialog_BuildRequest_InvalidDuration(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewAvailabilityDialog(theme)

	dialog.participantsInput.SetValue("alice@example.com")
	dialog.startDateInput.SetValue("2025-01-15")
	dialog.endDateInput.SetValue("2025-01-22")
	dialog.durationInput.SetValue("invalid")

	_, err := dialog.buildRequest()
	if err == nil {
		t.Error("expected error for invalid duration")
	}
	if !strings.Contains(err.Error(), "duration") {
		t.Errorf("expected error about duration, got: %v", err)
	}
}

func TestAvailabilityDialog_BuildRequest_ZeroDuration(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewAvailabilityDialog(theme)

	dialog.participantsInput.SetValue("alice@example.com")
	dialog.startDateInput.SetValue("2025-01-15")
	dialog.endDateInput.SetValue("2025-01-22")
	dialog.durationInput.SetValue("0")

	_, err := dialog.buildRequest()
	if err == nil {
		t.Error("expected error for zero duration")
	}
}

func TestAvailabilityDialog_Update_TabNavigation(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewAvailabilityDialog(theme)

	// Tab should move to next field
	tabKey := tea.KeyPressMsg{Code: tea.KeyTab}
	dialog, _ = dialog.Update(tabKey)

	if dialog.focusedField != AvailabilityFieldStartDate {
		t.Errorf("expected focus on StartDate after tab, got %d", dialog.focusedField)
	}

	// Tab again
	dialog, _ = dialog.Update(tabKey)
	if dialog.focusedField != AvailabilityFieldEndDate {
		t.Errorf("expected focus on EndDate after tab, got %d", dialog.focusedField)
	}
}

func TestAvailabilityDialog_Update_ShiftTabNavigation(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewAvailabilityDialog(theme)

	// Move to StartDate first
	dialog.focusedField = AvailabilityFieldStartDate
	dialog.focusCurrent()

	// Shift+Tab should go back to Participants
	shiftTabKey := tea.KeyPressMsg{Text: "shift+tab"}
	dialog, _ = dialog.Update(shiftTabKey)

	if dialog.focusedField != AvailabilityFieldParticipants {
		t.Errorf("expected focus on Participants after shift+tab, got %d", dialog.focusedField)
	}
}

func TestAvailabilityDialog_Update_EscapeCancel(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewAvailabilityDialog(theme)

	escKey := tea.KeyPressMsg{Code: tea.KeyEsc}
	_, cmd := dialog.Update(escKey)

	if cmd == nil {
		t.Fatal("expected cmd to be returned")
	}

	msg := cmd()
	if _, ok := msg.(AvailabilityCancelMsg); !ok {
		t.Errorf("expected AvailabilityCancelMsg, got %T", msg)
	}

	if dialog.IsVisible() {
		t.Error("expected dialog to be hidden after escape")
	}
}

func TestAvailabilityDialog_Update_EnterCheck(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewAvailabilityDialog(theme)

	// Set valid values
	dialog.participantsInput.SetValue("alice@example.com")
	dialog.startDateInput.SetValue("2025-01-15")
	dialog.endDateInput.SetValue("2025-01-22")
	dialog.durationInput.SetValue("30")

	// Focus on check button
	dialog.focusedField = AvailabilityFieldCheck

	enterKey := tea.KeyPressMsg{Code: tea.KeyEnter}
	dialog, cmd := dialog.Update(enterKey)

	if cmd == nil {
		t.Fatal("expected cmd to be returned")
	}

	if !dialog.loading {
		t.Error("expected loading to be true")
	}

	msg := cmd()
	result, ok := msg.(AvailabilityCheckMsg)
	if !ok {
		t.Fatalf("expected AvailabilityCheckMsg, got %T", msg)
	}

	if result.Request == nil {
		t.Error("expected request to be set")
	}
}

func TestAvailabilityDialog_Update_SlotSelection(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewAvailabilityDialog(theme)

	// Set some results
	now := time.Now()
	dialog.results = []domain.AvailableSlot{
		{StartTime: now.Unix(), EndTime: now.Add(30 * time.Minute).Unix()},
		{StartTime: now.Add(time.Hour).Unix(), EndTime: now.Add(90 * time.Minute).Unix()},
	}

	// Press "1" to select first slot
	oneKey := tea.KeyPressMsg{Text: "1"}
	_, cmd := dialog.Update(oneKey)

	if cmd == nil {
		t.Fatal("expected cmd to be returned")
	}

	msg := cmd()
	result, ok := msg.(AvailabilitySelectSlotMsg)
	if !ok {
		t.Fatalf("expected AvailabilitySelectSlotMsg, got %T", msg)
	}

	if result.Slot.StartTime != dialog.results[0].StartTime {
		t.Error("expected first slot to be selected")
	}
}

func TestAvailabilityDialog_View(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewAvailabilityDialog(theme)
	dialog.SetSize(80, 40)

	view := dialog.View()

	if !strings.Contains(view, "Check Availability") {
		t.Error("expected 'Check Availability' in view")
	}
	if !strings.Contains(view, "Participants") {
		t.Error("expected 'Participants' field in view")
	}
	if !strings.Contains(view, "Start Date") {
		t.Error("expected 'Start Date' field in view")
	}
	if !strings.Contains(view, "End Date") {
		t.Error("expected 'End Date' field in view")
	}
	if !strings.Contains(view, "Duration") {
		t.Error("expected 'Duration' field in view")
	}
	if !strings.Contains(view, "Check") {
		t.Error("expected 'Check' button in view")
	}
	if !strings.Contains(view, "Cancel") {
		t.Error("expected 'Cancel' button in view")
	}
}

func TestAvailabilityDialog_View_Hidden(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewAvailabilityDialog(theme)
	dialog.Hide()

	view := dialog.View()

	if view != "" {
		t.Error("expected empty view when dialog is hidden")
	}
}

func TestAvailabilityDialog_View_WithResults(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewAvailabilityDialog(theme)
	dialog.SetSize(80, 50)

	// Set some results
	now := time.Now()
	dialog.results = []domain.AvailableSlot{
		{StartTime: now.Unix(), EndTime: now.Add(30 * time.Minute).Unix()},
		{StartTime: now.Add(time.Hour).Unix(), EndTime: now.Add(90 * time.Minute).Unix()},
	}

	view := dialog.View()

	if !strings.Contains(view, "Found 2 available slots") {
		t.Error("expected results count in view")
	}
	if !strings.Contains(view, "[1]") {
		t.Error("expected slot selection hint in view")
	}
}

func TestAvailabilityDialog_View_WithError(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewAvailabilityDialog(theme)
	dialog.SetSize(80, 40)

	dialog.SetError(nil) // This should set err to nil

	// Now set a real error
	dialog.SetError(fmt.Errorf("test error"))

	// The view should handle errors gracefully
	view := dialog.View()
	if view == "" {
		t.Error("expected non-empty view even with error")
	}
	if !strings.Contains(view, "test error") {
		t.Error("expected error message in view")
	}
}

func TestAvailabilityDialog_ShowHide(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewAvailabilityDialog(theme)

	dialog.Hide()
	if dialog.IsVisible() {
		t.Error("expected dialog to be hidden")
	}

	dialog.Show()
	if !dialog.IsVisible() {
		t.Error("expected dialog to be visible")
	}
}

func TestAvailabilityDialog_Reset(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewAvailabilityDialog(theme)

	// Set some values
	dialog.participantsInput.SetValue("alice@example.com")
	dialog.durationInput.SetValue("60")
	now := time.Now()
	dialog.results = []domain.AvailableSlot{
		{StartTime: now.Unix(), EndTime: now.Add(30 * time.Minute).Unix()},
	}
	dialog.focusedField = AvailabilityFieldDuration

	// Reset
	dialog.Reset()

	if dialog.participantsInput.Value() != "" {
		t.Error("expected participants to be cleared")
	}
	if dialog.durationInput.Value() != "30" {
		t.Error("expected duration to be reset to default")
	}
	if len(dialog.results) != 0 {
		t.Error("expected results to be cleared")
	}
	if dialog.focusedField != AvailabilityFieldParticipants {
		t.Error("expected focus to be reset to Participants")
	}
}

func TestAvailabilityDialog_SetTimezone(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewAvailabilityDialog(theme)

	loc, _ := time.LoadLocation("America/New_York")
	dialog.SetTimezone(loc)

	if dialog.timezone != loc {
		t.Error("expected timezone to be set")
	}
}

func TestAvailabilityDialog_SetDateRange(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewAvailabilityDialog(theme)

	start := time.Date(2025, 6, 1, 0, 0, 0, 0, time.Local)
	end := time.Date(2025, 6, 15, 0, 0, 0, 0, time.Local)
	dialog.SetDateRange(start, end)

	if dialog.startDateInput.Value() != "2025-06-01" {
		t.Errorf("expected start date '2025-06-01', got %q", dialog.startDateInput.Value())
	}
	if dialog.endDateInput.Value() != "2025-06-15" {
		t.Errorf("expected end date '2025-06-15', got %q", dialog.endDateInput.Value())
	}
}

func TestAvailabilityDialog_Init(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewAvailabilityDialog(theme)

	cmd := dialog.Init()
	if cmd == nil {
		t.Error("expected Init to return a command for text input blink")
	}
}

func TestAvailabilityDialog_Update_NotVisibleNoOp(t *testing.T) {
	theme := styles.GetTheme("default")
	dialog := NewAvailabilityDialog(theme)
	dialog.Hide()

	tabKey := tea.KeyPressMsg{Code: tea.KeyTab}
	_, cmd := dialog.Update(tabKey)

	if cmd != nil {
		t.Error("expected no cmd when dialog is hidden")
	}
}
