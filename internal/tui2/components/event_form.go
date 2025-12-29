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

// EventFormMode represents the form mode.
type EventFormMode int

const (
	// EventFormCreate is for creating new events.
	EventFormCreate EventFormMode = iota
	// EventFormEdit is for editing existing events.
	EventFormEdit
)

// EventFormField represents which field is focused.
type EventFormField int

const (
	EventFormFieldTitle EventFormField = iota
	EventFormFieldLocation
	EventFormFieldDescription
	EventFormFieldStartDate
	EventFormFieldStartTime
	EventFormFieldEndDate
	EventFormFieldEndTime
	EventFormFieldAllDay
	EventFormFieldBusy
	EventFormFieldSubmit
	EventFormFieldCancel
)

const eventFormFieldCount = 11

// EventFormSubmitMsg is sent when the form is submitted.
type EventFormSubmitMsg struct {
	Mode    EventFormMode
	EventID string // Only for edit mode
	Request *domain.CreateEventRequest
}

// EventFormCancelMsg is sent when the form is cancelled.
type EventFormCancelMsg struct{}

// EventForm is a form for creating/editing events.
type EventForm struct {
	theme *styles.Theme
	mode  EventFormMode

	// Form fields
	titleInput       textinput.Model
	locationInput    textinput.Model
	descriptionInput textinput.Model
	startDateInput   textinput.Model
	startTimeInput   textinput.Model
	endDateInput     textinput.Model
	endTimeInput     textinput.Model

	// Toggle fields
	allDay bool
	busy   bool

	// State
	focusedField EventFormField
	eventID      string // For edit mode
	timezone     *time.Location

	// Sizing
	width  int
	height int
}

// NewEventForm creates a new event form.
func NewEventForm(theme *styles.Theme, mode EventFormMode) *EventForm {
	// Title input
	titleInput := textinput.New()
	titleInput.Placeholder = "Event title"
	titleInput.CharLimit = 200

	// Location input
	locationInput := textinput.New()
	locationInput.Placeholder = "Location (optional)"
	locationInput.CharLimit = 200

	// Description input
	descriptionInput := textinput.New()
	descriptionInput.Placeholder = "Description (optional)"
	descriptionInput.CharLimit = 2000

	// Date/time inputs with placeholders
	now := time.Now()
	// Round to next hour for start time
	startTime := now.Truncate(time.Hour).Add(time.Hour)
	// End time is 1 hour after start
	endTime := startTime.Add(time.Hour)

	startDateInput := textinput.New()
	startDateInput.Placeholder = "YYYY-MM-DD"
	startDateInput.CharLimit = 10
	startDateInput.SetValue(startTime.Format("2006-01-02"))

	startTimeInput := textinput.New()
	startTimeInput.Placeholder = "HH:MM"
	startTimeInput.CharLimit = 5
	startTimeInput.SetValue(startTime.Format("15:04"))

	endDateInput := textinput.New()
	endDateInput.Placeholder = "YYYY-MM-DD"
	endDateInput.CharLimit = 10
	endDateInput.SetValue(endTime.Format("2006-01-02"))

	endTimeInput := textinput.New()
	endTimeInput.Placeholder = "HH:MM"
	endTimeInput.CharLimit = 5
	endTimeInput.SetValue(endTime.Format("15:04"))

	form := &EventForm{
		theme:            theme,
		mode:             mode,
		titleInput:       titleInput,
		locationInput:    locationInput,
		descriptionInput: descriptionInput,
		startDateInput:   startDateInput,
		startTimeInput:   startTimeInput,
		endDateInput:     endDateInput,
		endTimeInput:     endTimeInput,
		allDay:           false,
		busy:             true,
		focusedField:     EventFormFieldTitle,
		timezone:         time.Local,
	}

	// Focus the title input
	form.updateFocus()

	return form
}

// SetEvent populates the form with an existing event (for edit mode).
func (f *EventForm) SetEvent(event *domain.Event) {
	if event == nil {
		return
	}

	f.eventID = event.ID
	f.titleInput.SetValue(event.Title)
	f.locationInput.SetValue(event.Location)
	f.descriptionInput.SetValue(event.Description)
	f.busy = event.Busy

	// Handle when
	if event.When.IsAllDay() {
		f.allDay = true
		if event.When.Date != "" {
			f.startDateInput.SetValue(event.When.Date)
			f.endDateInput.SetValue(event.When.Date)
		} else if event.When.StartDate != "" {
			f.startDateInput.SetValue(event.When.StartDate)
			if event.When.EndDate != "" {
				f.endDateInput.SetValue(event.When.EndDate)
			} else {
				f.endDateInput.SetValue(event.When.StartDate)
			}
		}
	} else {
		f.allDay = false
		startTime := event.When.StartDateTime()
		endTime := event.When.EndDateTime()
		if !startTime.IsZero() {
			if f.timezone != nil {
				startTime = startTime.In(f.timezone)
			}
			f.startDateInput.SetValue(startTime.Format("2006-01-02"))
			f.startTimeInput.SetValue(startTime.Format("15:04"))
		}
		if !endTime.IsZero() {
			if f.timezone != nil {
				endTime = endTime.In(f.timezone)
			}
			f.endDateInput.SetValue(endTime.Format("2006-01-02"))
			f.endTimeInput.SetValue(endTime.Format("15:04"))
		}
	}
}

// SetTimezone sets the timezone for the form.
func (f *EventForm) SetTimezone(tz *time.Location) {
	if tz != nil {
		f.timezone = tz
	}
}

// SetDate sets the initial date for the event.
func (f *EventForm) SetDate(date time.Time) {
	f.startDateInput.SetValue(date.Format("2006-01-02"))
	f.endDateInput.SetValue(date.Format("2006-01-02"))
}

// SetTimeRange sets the start and end time for the event.
func (f *EventForm) SetTimeRange(start, end time.Time) {
	f.startDateInput.SetValue(start.Format("2006-01-02"))
	f.startTimeInput.SetValue(start.Format("15:04"))
	f.endDateInput.SetValue(end.Format("2006-01-02"))
	f.endTimeInput.SetValue(end.Format("15:04"))
}

// SetSize sets the form size.
func (f *EventForm) SetSize(width, height int) {
	f.width = width
	f.height = height
	// Note: textinput width is controlled via CharLimit in Bubble Tea v2
	// The visual width is determined by the terminal, not a Width property
}

// Init implements tea.Model.
func (f *EventForm) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements tea.Model.
func (f *EventForm) Update(msg tea.Msg) (*EventForm, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.Key()
		keyStr := msg.String()

		// Handle escape - cancel form
		if key.Code == tea.KeyEsc {
			return f, func() tea.Msg { return EventFormCancelMsg{} }
		}

		// Handle tab navigation
		if key.Code == tea.KeyTab {
			f.focusedField = (f.focusedField + 1) % eventFormFieldCount
			f.updateFocus()
			return f, nil
		}

		// Handle shift+tab
		if keyStr == "shift+tab" {
			f.focusedField = (f.focusedField - 1 + eventFormFieldCount) % eventFormFieldCount
			f.updateFocus()
			return f, nil
		}

		// Handle enter
		if key.Code == tea.KeyEnter {
			switch f.focusedField {
			case EventFormFieldAllDay:
				f.allDay = !f.allDay
				return f, nil
			case EventFormFieldBusy:
				f.busy = !f.busy
				return f, nil
			case EventFormFieldSubmit:
				req, err := f.buildRequest()
				if err != nil {
					// TODO: Show error
					return f, nil
				}
				return f, func() tea.Msg {
					return EventFormSubmitMsg{
						Mode:    f.mode,
						EventID: f.eventID,
						Request: req,
					}
				}
			case EventFormFieldCancel:
				return f, func() tea.Msg { return EventFormCancelMsg{} }
			}
		}

		// Handle space for toggles
		if keyStr == " " {
			switch f.focusedField {
			case EventFormFieldAllDay:
				f.allDay = !f.allDay
				return f, nil
			case EventFormFieldBusy:
				f.busy = !f.busy
				return f, nil
			}
		}
	}

	// Update text inputs
	switch f.focusedField {
	case EventFormFieldTitle:
		f.titleInput, cmd = f.titleInput.Update(msg)
		cmds = append(cmds, cmd)
	case EventFormFieldLocation:
		f.locationInput, cmd = f.locationInput.Update(msg)
		cmds = append(cmds, cmd)
	case EventFormFieldDescription:
		f.descriptionInput, cmd = f.descriptionInput.Update(msg)
		cmds = append(cmds, cmd)
	case EventFormFieldStartDate:
		f.startDateInput, cmd = f.startDateInput.Update(msg)
		cmds = append(cmds, cmd)
	case EventFormFieldStartTime:
		f.startTimeInput, cmd = f.startTimeInput.Update(msg)
		cmds = append(cmds, cmd)
	case EventFormFieldEndDate:
		f.endDateInput, cmd = f.endDateInput.Update(msg)
		cmds = append(cmds, cmd)
	case EventFormFieldEndTime:
		f.endTimeInput, cmd = f.endTimeInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return f, tea.Batch(cmds...)
}

// View implements tea.Model.
func (f *EventForm) View() string {
	var b strings.Builder

	// Title
	title := "Create Event"
	if f.mode == EventFormEdit {
		title = "Edit Event"
	}
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(f.theme.Primary)
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n\n")

	labelStyle := lipgloss.NewStyle().Width(12)
	focusedLabelStyle := labelStyle.Foreground(f.theme.Primary).Bold(true)

	// Title field
	label := labelStyle.Render("Title:")
	if f.focusedField == EventFormFieldTitle {
		label = focusedLabelStyle.Render("Title:")
	}
	b.WriteString(label + " " + f.titleInput.View())
	b.WriteString("\n")

	// Location field
	label = labelStyle.Render("Location:")
	if f.focusedField == EventFormFieldLocation {
		label = focusedLabelStyle.Render("Location:")
	}
	b.WriteString(label + " " + f.locationInput.View())
	b.WriteString("\n")

	// Description field
	label = labelStyle.Render("Description:")
	if f.focusedField == EventFormFieldDescription {
		label = focusedLabelStyle.Render("Description:")
	}
	b.WriteString(label + " " + f.descriptionInput.View())
	b.WriteString("\n\n")

	// All-day toggle
	label = labelStyle.Render("All Day:")
	if f.focusedField == EventFormFieldAllDay {
		label = focusedLabelStyle.Render("All Day:")
	}
	checkbox := "[ ]"
	if f.allDay {
		checkbox = "[x]"
	}
	if f.focusedField == EventFormFieldAllDay {
		checkbox = lipgloss.NewStyle().Foreground(f.theme.Primary).Bold(true).Render(checkbox)
	}
	b.WriteString(label + " " + checkbox)
	b.WriteString("\n")

	// Start date/time
	label = labelStyle.Render("Start:")
	if f.focusedField == EventFormFieldStartDate || f.focusedField == EventFormFieldStartTime {
		label = focusedLabelStyle.Render("Start:")
	}
	if f.allDay {
		b.WriteString(label + " " + f.startDateInput.View())
	} else {
		b.WriteString(label + " " + f.startDateInput.View() + " " + f.startTimeInput.View())
	}
	b.WriteString("\n")

	// End date/time
	label = labelStyle.Render("End:")
	if f.focusedField == EventFormFieldEndDate || f.focusedField == EventFormFieldEndTime {
		label = focusedLabelStyle.Render("End:")
	}
	if f.allDay {
		b.WriteString(label + " " + f.endDateInput.View())
	} else {
		b.WriteString(label + " " + f.endDateInput.View() + " " + f.endTimeInput.View())
	}
	b.WriteString("\n\n")

	// Busy toggle
	label = labelStyle.Render("Busy:")
	if f.focusedField == EventFormFieldBusy {
		label = focusedLabelStyle.Render("Busy:")
	}
	checkbox = "[ ]"
	if f.busy {
		checkbox = "[x]"
	}
	if f.focusedField == EventFormFieldBusy {
		checkbox = lipgloss.NewStyle().Foreground(f.theme.Primary).Bold(true).Render(checkbox)
	}
	b.WriteString(label + " " + checkbox)
	b.WriteString("\n\n")

	// Buttons
	submitStyle := lipgloss.NewStyle().Padding(0, 2).Background(f.theme.Primary).Foreground(lipgloss.Color("#FFFFFF"))
	cancelStyle := lipgloss.NewStyle().Padding(0, 2).Border(lipgloss.NormalBorder()).BorderForeground(f.theme.Dimmed.GetForeground())

	if f.focusedField == EventFormFieldSubmit {
		submitStyle = submitStyle.Bold(true).Underline(true)
	}
	if f.focusedField == EventFormFieldCancel {
		cancelStyle = cancelStyle.Bold(true).BorderForeground(f.theme.Primary)
	}

	submitText := "Create"
	if f.mode == EventFormEdit {
		submitText = "Save"
	}

	b.WriteString(submitStyle.Render(submitText) + "  " + cancelStyle.Render("Cancel"))
	b.WriteString("\n\n")

	// Help text
	helpStyle := lipgloss.NewStyle().Foreground(f.theme.Dimmed.GetForeground())
	b.WriteString(helpStyle.Render("Tab: next field  Shift+Tab: previous  Enter: select  Esc: cancel"))

	return b.String()
}

// updateFocus updates the focus state of all inputs.
func (f *EventForm) updateFocus() {
	f.titleInput.Blur()
	f.locationInput.Blur()
	f.descriptionInput.Blur()
	f.startDateInput.Blur()
	f.startTimeInput.Blur()
	f.endDateInput.Blur()
	f.endTimeInput.Blur()

	switch f.focusedField {
	case EventFormFieldTitle:
		f.titleInput.Focus()
	case EventFormFieldLocation:
		f.locationInput.Focus()
	case EventFormFieldDescription:
		f.descriptionInput.Focus()
	case EventFormFieldStartDate:
		f.startDateInput.Focus()
	case EventFormFieldStartTime:
		f.startTimeInput.Focus()
	case EventFormFieldEndDate:
		f.endDateInput.Focus()
	case EventFormFieldEndTime:
		f.endTimeInput.Focus()
	}
}

// buildRequest builds a CreateEventRequest from the form data.
func (f *EventForm) buildRequest() (*domain.CreateEventRequest, error) {
	title := strings.TrimSpace(f.titleInput.Value())
	if title == "" {
		return nil, fmt.Errorf("title is required")
	}

	req := &domain.CreateEventRequest{
		Title:       title,
		Location:    strings.TrimSpace(f.locationInput.Value()),
		Description: strings.TrimSpace(f.descriptionInput.Value()),
		Busy:        f.busy,
	}

	// Parse dates and times
	if f.allDay {
		req.When = domain.EventWhen{
			Object:    "datespan",
			StartDate: f.startDateInput.Value(),
			EndDate:   f.endDateInput.Value(),
		}
		// Validate date format
		if _, err := time.Parse("2006-01-02", req.When.StartDate); err != nil {
			return nil, fmt.Errorf("invalid start date format (use YYYY-MM-DD)")
		}
		if _, err := time.Parse("2006-01-02", req.When.EndDate); err != nil {
			return nil, fmt.Errorf("invalid end date format (use YYYY-MM-DD)")
		}
	} else {
		// Parse start time
		startDateStr := f.startDateInput.Value()
		startTimeStr := f.startTimeInput.Value()
		startDateTime, err := time.ParseInLocation("2006-01-02 15:04", startDateStr+" "+startTimeStr, f.timezone)
		if err != nil {
			return nil, fmt.Errorf("invalid start date/time format")
		}

		// Parse end time
		endDateStr := f.endDateInput.Value()
		endTimeStr := f.endTimeInput.Value()
		endDateTime, err := time.ParseInLocation("2006-01-02 15:04", endDateStr+" "+endTimeStr, f.timezone)
		if err != nil {
			return nil, fmt.Errorf("invalid end date/time format")
		}

		if endDateTime.Before(startDateTime) {
			return nil, fmt.Errorf("end time must be after start time")
		}

		tzName := "UTC"
		if f.timezone != nil {
			tzName = f.timezone.String()
		}

		req.When = domain.EventWhen{
			Object:        "timespan",
			StartTime:     startDateTime.Unix(),
			EndTime:       endDateTime.Unix(),
			StartTimezone: tzName,
			EndTimezone:   tzName,
		}
	}

	return req, nil
}

// Validate validates the form.
func (f *EventForm) Validate() error {
	_, err := f.buildRequest()
	return err
}

// GetTitle returns the current title value.
func (f *EventForm) GetTitle() string {
	return f.titleInput.Value()
}

// IsAllDay returns true if the event is all-day.
func (f *EventForm) IsAllDay() bool {
	return f.allDay
}

// IsBusy returns true if the event is busy.
func (f *EventForm) IsBusy() bool {
	return f.busy
}

// GetMode returns the form mode.
func (f *EventForm) GetMode() EventFormMode {
	return f.mode
}
