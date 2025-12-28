package components

import (
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/tui2/styles"
)

// AvailabilityDialogField represents a field in the availability dialog.
type AvailabilityDialogField int

const (
	AvailabilityFieldParticipants AvailabilityDialogField = iota
	AvailabilityFieldStartDate
	AvailabilityFieldEndDate
	AvailabilityFieldDuration
	AvailabilityFieldCheck  // Check button
	AvailabilityFieldCancel // Cancel button
	availabilityFieldCount
)

// AvailabilityDialog is a dialog for checking availability.
type AvailabilityDialog struct {
	theme *styles.Theme

	// Input fields
	participantsInput textinput.Model
	startDateInput    textinput.Model
	endDateInput      textinput.Model
	durationInput     textinput.Model

	// State
	focusedField AvailabilityDialogField
	visible      bool
	width        int
	height       int

	// Results
	loading    bool
	results    []domain.AvailableSlot
	err        error
	timezone   *time.Location
	calendarID string // The calendar ID to check availability against
}

// AvailabilityCheckMsg is sent when user wants to check availability.
type AvailabilityCheckMsg struct {
	Request *domain.AvailabilityRequest
}

// AvailabilityResultsMsg contains the availability results.
type AvailabilityResultsMsg struct {
	Slots []domain.AvailableSlot
}

// AvailabilityCancelMsg is sent when the dialog is cancelled.
type AvailabilityCancelMsg struct{}

// AvailabilitySelectSlotMsg is sent when a slot is selected.
type AvailabilitySelectSlotMsg struct {
	Slot domain.AvailableSlot
}

// NewAvailabilityDialog creates a new availability check dialog.
func NewAvailabilityDialog(theme *styles.Theme) *AvailabilityDialog {
	d := &AvailabilityDialog{
		theme:        theme,
		focusedField: AvailabilityFieldParticipants,
		visible:      true,
		timezone:     time.Local,
	}

	// Create text inputs
	d.participantsInput = createAvailInput("email@example.com, email2@example.com")
	d.startDateInput = createAvailInput("YYYY-MM-DD")
	d.endDateInput = createAvailInput("YYYY-MM-DD")
	d.durationInput = createAvailInput("30")

	// Set defaults
	now := time.Now()
	d.startDateInput.SetValue(now.Format("2006-01-02"))
	d.endDateInput.SetValue(now.AddDate(0, 0, 7).Format("2006-01-02"))
	d.durationInput.SetValue("30")

	// Focus first field
	d.participantsInput.Focus()

	return d
}

// createAvailInput creates a text input for the availability dialog.
func createAvailInput(placeholder string) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.CharLimit = 256
	return ti
}

// Init implements tea.Model.
func (d *AvailabilityDialog) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements tea.Model.
func (d *AvailabilityDialog) Update(msg tea.Msg) (*AvailabilityDialog, tea.Cmd) {
	if !d.visible {
		return d, nil
	}

	switch msg := msg.(type) {
	case AvailabilityResultsMsg:
		d.loading = false
		d.results = msg.Slots
		return d, nil

	case tea.KeyMsg:
		key := msg.Key()
		keyStr := msg.String()

		switch {
		case key.Code == tea.KeyEsc:
			d.visible = false
			return d, func() tea.Msg { return AvailabilityCancelMsg{} }

		case key.Code == tea.KeyTab, keyStr == "down":
			d.focusNext()
			return d, nil

		case keyStr == "shift+tab", keyStr == "up":
			d.focusPrev()
			return d, nil

		case key.Code == tea.KeyEnter:
			switch d.focusedField {
			case AvailabilityFieldCheck:
				req, err := d.buildRequest()
				if err != nil {
					d.err = err
					return d, nil
				}
				d.loading = true
				d.err = nil
				return d, func() tea.Msg { return AvailabilityCheckMsg{Request: req} }
			case AvailabilityFieldCancel:
				d.visible = false
				return d, func() tea.Msg { return AvailabilityCancelMsg{} }
			default:
				// Move to next field on Enter for text inputs
				d.focusNext()
				return d, nil
			}

		case keyStr == "1", keyStr == "2", keyStr == "3", keyStr == "4", keyStr == "5":
			// Quick selection of results
			if len(d.results) > 0 {
				idx := int(keyStr[0] - '1')
				if idx >= 0 && idx < len(d.results) {
					d.visible = false
					return d, func() tea.Msg { return AvailabilitySelectSlotMsg{Slot: d.results[idx]} }
				}
			}
		}
	}

	// Update the focused text input
	var cmd tea.Cmd
	switch d.focusedField {
	case AvailabilityFieldParticipants:
		d.participantsInput, cmd = d.participantsInput.Update(msg)
	case AvailabilityFieldStartDate:
		d.startDateInput, cmd = d.startDateInput.Update(msg)
	case AvailabilityFieldEndDate:
		d.endDateInput, cmd = d.endDateInput.Update(msg)
	case AvailabilityFieldDuration:
		d.durationInput, cmd = d.durationInput.Update(msg)
	}

	return d, cmd
}

// focusNext moves focus to the next field.
func (d *AvailabilityDialog) focusNext() {
	d.blurCurrent()
	d.focusedField = (d.focusedField + 1) % availabilityFieldCount
	d.focusCurrent()
}

// focusPrev moves focus to the previous field.
func (d *AvailabilityDialog) focusPrev() {
	d.blurCurrent()
	d.focusedField = (d.focusedField - 1 + availabilityFieldCount) % availabilityFieldCount
	d.focusCurrent()
}

// blurCurrent blurs the current field.
func (d *AvailabilityDialog) blurCurrent() {
	switch d.focusedField {
	case AvailabilityFieldParticipants:
		d.participantsInput.Blur()
	case AvailabilityFieldStartDate:
		d.startDateInput.Blur()
	case AvailabilityFieldEndDate:
		d.endDateInput.Blur()
	case AvailabilityFieldDuration:
		d.durationInput.Blur()
	}
}

// focusCurrent focuses the current field.
func (d *AvailabilityDialog) focusCurrent() {
	switch d.focusedField {
	case AvailabilityFieldParticipants:
		d.participantsInput.Focus()
	case AvailabilityFieldStartDate:
		d.startDateInput.Focus()
	case AvailabilityFieldEndDate:
		d.endDateInput.Focus()
	case AvailabilityFieldDuration:
		d.durationInput.Focus()
	}
}

// View implements tea.Model.
func (d *AvailabilityDialog) View() string {
	if !d.visible {
		return ""
	}

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(d.theme.Primary).
		MarginBottom(1)
	title := titleStyle.Render("📅 Check Availability")

	// Build the form
	var rows []string

	// Helper to render a labeled field
	renderField := func(label string, input textinput.Model, focused bool) string {
		labelStyle := lipgloss.NewStyle().Width(14).Foreground(d.theme.Secondary)
		if focused {
			labelStyle = labelStyle.Bold(true).Foreground(d.theme.Primary)
		}

		inputWidth := d.width - 25
		if inputWidth < 30 {
			inputWidth = 30
		}

		inputStyle := lipgloss.NewStyle().Width(inputWidth)
		if focused {
			inputStyle = inputStyle.BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(d.theme.Primary).
				BorderLeft(true)
		}

		return lipgloss.JoinHorizontal(lipgloss.Left,
			labelStyle.Render(label+":"),
			inputStyle.Render(input.View()),
		)
	}

	// Add all fields
	rows = append(rows, renderField("Participants", d.participantsInput, d.focusedField == AvailabilityFieldParticipants))
	rows = append(rows, renderField("Start Date", d.startDateInput, d.focusedField == AvailabilityFieldStartDate))
	rows = append(rows, renderField("End Date", d.endDateInput, d.focusedField == AvailabilityFieldEndDate))
	rows = append(rows, renderField("Duration (min)", d.durationInput, d.focusedField == AvailabilityFieldDuration))

	// Error display
	if d.err != nil {
		errStyle := lipgloss.NewStyle().Foreground(d.theme.Error).MarginTop(1)
		rows = append(rows, errStyle.Render("Error: "+d.err.Error()))
	}

	// Spacer
	rows = append(rows, "")

	// Buttons
	checkStyle := lipgloss.NewStyle().
		Padding(0, 2).
		MarginRight(2)
	cancelStyle := lipgloss.NewStyle().
		Padding(0, 2)

	if d.focusedField == AvailabilityFieldCheck {
		checkStyle = checkStyle.
			Background(d.theme.Primary).
			Foreground(lipgloss.Color("#000000")).
			Bold(true)
	} else {
		checkStyle = checkStyle.
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(d.theme.Primary)
	}

	if d.focusedField == AvailabilityFieldCancel {
		cancelStyle = cancelStyle.
			Background(d.theme.Secondary).
			Foreground(lipgloss.Color("#000000")).
			Bold(true)
	} else {
		cancelStyle = cancelStyle.
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(d.theme.Secondary)
	}

	checkLabel := "Check"
	if d.loading {
		checkLabel = "Checking..."
	}
	buttons := lipgloss.JoinHorizontal(lipgloss.Center,
		checkStyle.Render(checkLabel),
		cancelStyle.Render("Cancel"),
	)
	rows = append(rows, buttons)

	// Results section
	if len(d.results) > 0 {
		rows = append(rows, "")
		resultHeader := lipgloss.NewStyle().
			Bold(true).
			Foreground(d.theme.Success).
			MarginTop(1)
		rows = append(rows, resultHeader.Render(fmt.Sprintf("✓ Found %d available slots:", len(d.results))))
		rows = append(rows, "")

		// Show up to 5 results
		maxResults := 5
		if len(d.results) < maxResults {
			maxResults = len(d.results)
		}

		for i := 0; i < maxResults; i++ {
			slot := d.results[i]
			start := time.Unix(slot.StartTime, 0)
			end := time.Unix(slot.EndTime, 0)
			if d.timezone != nil {
				start = start.In(d.timezone)
				end = end.In(d.timezone)
			}

			slotStyle := lipgloss.NewStyle().Foreground(d.theme.Info)
			slotText := fmt.Sprintf("[%d] %s %s - %s",
				i+1,
				start.Format("Mon Jan 2"),
				start.Format("3:04 PM"),
				end.Format("3:04 PM"),
			)
			rows = append(rows, slotStyle.Render(slotText))
		}

		if len(d.results) > maxResults {
			moreStyle := lipgloss.NewStyle().Foreground(d.theme.Dimmed.GetForeground())
			rows = append(rows, moreStyle.Render(fmt.Sprintf("... and %d more", len(d.results)-maxResults)))
		}

		selectHint := lipgloss.NewStyle().
			Foreground(d.theme.Secondary).
			Italic(true).
			MarginTop(1)
		rows = append(rows, selectHint.Render("Press 1-5 to select a slot and create an event"))
	}

	// Help text
	helpStyle := lipgloss.NewStyle().Foreground(d.theme.Dimmed.GetForeground()).MarginTop(1)
	rows = append(rows, helpStyle.Render("Tab: next field  Shift+Tab: previous  Enter: check/submit  Esc: cancel"))

	// Join all rows
	content := lipgloss.JoinVertical(lipgloss.Left, rows...)

	// Dialog box
	dialogWidth := d.width - 10
	if dialogWidth < 60 {
		dialogWidth = 60
	}
	dialogStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(d.theme.Primary).
		Padding(1, 2).
		Width(dialogWidth)

	// Center the dialog
	dialog := dialogStyle.Render(title + "\n" + content)
	return lipgloss.Place(
		d.width,
		d.height,
		lipgloss.Center,
		lipgloss.Center,
		dialog,
	)
}

// buildRequest builds an AvailabilityRequest from the form values.
func (d *AvailabilityDialog) buildRequest() (*domain.AvailabilityRequest, error) {
	// Parse participants
	participantsStr := strings.TrimSpace(d.participantsInput.Value())
	if participantsStr == "" {
		return nil, fmt.Errorf("participants are required")
	}

	var participants []domain.AvailabilityParticipant
	for _, email := range strings.Split(participantsStr, ",") {
		email = strings.TrimSpace(email)
		if email != "" {
			participants = append(participants, domain.AvailabilityParticipant{
				Email: email,
			})
		}
	}

	if len(participants) == 0 {
		return nil, fmt.Errorf("at least one participant email is required")
	}

	// Parse dates
	startDateStr := strings.TrimSpace(d.startDateInput.Value())
	endDateStr := strings.TrimSpace(d.endDateInput.Value())

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid start date format (use YYYY-MM-DD)")
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid end date format (use YYYY-MM-DD)")
	}

	// Set times to beginning and end of day
	startTime := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 9, 0, 0, 0, d.timezone)
	endTime := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 17, 0, 0, 0, d.timezone)

	if endTime.Before(startTime) {
		return nil, fmt.Errorf("end date must be after start date")
	}

	// Parse duration
	durationStr := strings.TrimSpace(d.durationInput.Value())
	var duration int
	if _, err := fmt.Sscanf(durationStr, "%d", &duration); err != nil || duration <= 0 {
		return nil, fmt.Errorf("duration must be a positive number of minutes")
	}

	return &domain.AvailabilityRequest{
		StartTime:       startTime.Unix(),
		EndTime:         endTime.Unix(),
		DurationMinutes: duration,
		Participants:    participants,
		IntervalMinutes: 15, // Default interval
	}, nil
}

// SetSize sets the size of the dialog.
func (d *AvailabilityDialog) SetSize(width, height int) {
	d.width = width
	d.height = height
	// Note: In Bubble Tea v2, textinput width is controlled by CharLimit
	// and visual width is determined by the terminal, not a Width property
}

// SetTimezone sets the timezone for displaying results.
func (d *AvailabilityDialog) SetTimezone(tz *time.Location) {
	if tz != nil {
		d.timezone = tz
	}
}

// SetCalendarID sets the calendar ID for availability checking.
func (d *AvailabilityDialog) SetCalendarID(calendarID string) {
	d.calendarID = calendarID
}

// SetDateRange sets the date range for availability checking.
func (d *AvailabilityDialog) SetDateRange(start, end time.Time) {
	d.startDateInput.SetValue(start.Format("2006-01-02"))
	d.endDateInput.SetValue(end.Format("2006-01-02"))
}

// SetResults sets the availability results.
func (d *AvailabilityDialog) SetResults(slots []domain.AvailableSlot) {
	d.results = slots
	d.loading = false
}

// SetError sets an error message.
func (d *AvailabilityDialog) SetError(err error) {
	d.err = err
	d.loading = false
}

// Show shows the dialog.
func (d *AvailabilityDialog) Show() {
	d.visible = true
	d.focusCurrent()
}

// Hide hides the dialog.
func (d *AvailabilityDialog) Hide() {
	d.visible = false
	d.blurCurrent()
}

// IsVisible returns whether the dialog is visible.
func (d *AvailabilityDialog) IsVisible() bool {
	return d.visible
}

// Reset clears all fields and results.
func (d *AvailabilityDialog) Reset() {
	d.participantsInput.SetValue("")
	now := time.Now()
	d.startDateInput.SetValue(now.Format("2006-01-02"))
	d.endDateInput.SetValue(now.AddDate(0, 0, 7).Format("2006-01-02"))
	d.durationInput.SetValue("30")
	d.results = nil
	d.err = nil
	d.loading = false
	d.focusedField = AvailabilityFieldParticipants
	d.focusCurrent()
}
