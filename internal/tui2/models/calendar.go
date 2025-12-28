package models

import (
	"context"
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/tui2/components"
	"github.com/mqasimca/nylas/internal/tui2/state"
	"github.com/mqasimca/nylas/internal/tui2/styles"
)

// CalendarScreenMode represents the current mode of the calendar screen.
type CalendarScreenMode int

const (
	CalendarModeView CalendarScreenMode = iota
	CalendarModeEventForm
	CalendarModeConfirmDelete
	CalendarModeAvailability
)

// CalendarScreen is the calendar view screen.
type CalendarScreen struct {
	global *state.GlobalState
	theme  *styles.Theme

	calendarGrid       *components.CalendarGrid
	spinner            spinner.Model
	calendarList       list.Model
	eventForm          *components.EventForm
	confirmDialog      *components.ConfirmDialog
	availabilityDialog *components.AvailabilityDialog

	calendars        []domain.Calendar
	events           []domain.Event
	selectedCalendar *domain.Calendar
	selectedEvent    *domain.Event // Currently selected event for edit/delete
	selectedEventIdx int           // Index of selected event in day's events

	mode            CalendarScreenMode
	loading         bool
	loadingEvents   bool
	calendarsLoaded bool
	err             error
	width           int
	height          int
	timezone        *time.Location
}

// NewCalendarScreen creates a new calendar screen.
func NewCalendarScreen(global *state.GlobalState) *CalendarScreen {
	theme := styles.GetTheme(global.Theme)

	// Create spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(theme.Primary)

	// Create calendar grid
	grid := components.NewCalendarGrid(theme)

	// Create calendar list
	delegate := list.NewDefaultDelegate()
	calList := list.New(nil, delegate, 0, 0)
	calList.SetShowTitle(false)
	calList.SetShowStatusBar(false)
	calList.SetFilteringEnabled(false)
	calList.SetShowPagination(false)
	calList.SetShowHelp(false)

	// Initialize sizes if available
	if global.WindowSize.Width > 0 && global.WindowSize.Height > 0 {
		grid.SetSize(global.WindowSize.Width-30, global.WindowSize.Height-6)
		calList.SetSize(28, global.WindowSize.Height-6)
	}

	return &CalendarScreen{
		global:       global,
		theme:        theme,
		calendarGrid: grid,
		spinner:      s,
		calendarList: calList,
		loading:      true,
		mode:         CalendarModeView,
		timezone:     time.Local,
	}
}

// SetTimezone sets the timezone for the calendar screen.
func (m *CalendarScreen) SetTimezone(tz *time.Location) {
	if tz != nil {
		m.timezone = tz
		m.calendarGrid.SetTimezone(tz)
	}
}

// Init implements tea.Model.
func (m *CalendarScreen) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.fetchCalendars(),
	)
}

// Update implements tea.Model.
func (m *CalendarScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	// Handle mode-specific updates first
	switch m.mode {
	case CalendarModeEventForm:
		return m.updateEventForm(msg)
	case CalendarModeConfirmDelete:
		return m.updateConfirmDelete(msg)
	case CalendarModeAvailability:
		return m.updateAvailability(msg)
	}

	// View mode updates
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.Key()
		keyStr := msg.String()

		// Handle escape - go back
		if key.Code == tea.KeyEsc {
			return m, func() tea.Msg { return BackMsg{} }
		}

		// Handle ctrl+c
		if keyStr == "ctrl+c" {
			return m, tea.Quit
		}

		// Handle ctrl+r (refresh)
		if keyStr == "ctrl+r" {
			m.loading = true
			return m, tea.Batch(m.spinner.Tick, m.fetchCalendars())
		}

		// View mode switching
		switch keyStr {
		case "m":
			m.calendarGrid.SetViewMode(components.CalendarMonthView)
			return m, nil
		case "w":
			m.calendarGrid.SetViewMode(components.CalendarWeekView)
			return m, nil
		case "g":
			m.calendarGrid.SetViewMode(components.CalendarAgendaView)
			return m, nil
		case "t":
			m.calendarGrid.GoToToday()
			// Reload events for new date range
			if m.selectedCalendar != nil {
				m.loadingEvents = true
				return m, tea.Batch(m.spinner.Tick, m.fetchEventsForMonth(m.selectedCalendar.ID, m.calendarGrid.GetCurrentMonth()))
			}
			return m, nil
		case "n":
			// Create new event
			return m.openEventForm(components.EventFormCreate, nil)
		case "e":
			// Edit selected event
			events := m.calendarGrid.GetEventsForDate(m.calendarGrid.GetSelectedDate())
			if len(events) > 0 && m.selectedEventIdx < len(events) {
				return m.openEventForm(components.EventFormEdit, &events[m.selectedEventIdx])
			} else if len(events) > 0 {
				return m.openEventForm(components.EventFormEdit, &events[0])
			}
			m.global.SetStatus("No event selected to edit", 1)
			return m, nil
		case "d":
			// Delete selected event
			events := m.calendarGrid.GetEventsForDate(m.calendarGrid.GetSelectedDate())
			if len(events) > 0 && m.selectedEventIdx < len(events) {
				return m.openDeleteConfirmation(&events[m.selectedEventIdx])
			} else if len(events) > 0 {
				return m.openDeleteConfirmation(&events[0])
			}
			m.global.SetStatus("No event selected to delete", 1)
			return m, nil
		case "J":
			// Select next event in the day (Shift+J)
			events := m.calendarGrid.GetEventsForDate(m.calendarGrid.GetSelectedDate())
			if len(events) > 0 {
				m.selectedEventIdx = (m.selectedEventIdx + 1) % len(events)
			}
			return m, nil
		case "K":
			// Select previous event in the day (Shift+K)
			events := m.calendarGrid.GetEventsForDate(m.calendarGrid.GetSelectedDate())
			if len(events) > 0 {
				m.selectedEventIdx = (m.selectedEventIdx - 1 + len(events)) % len(events)
			}
			return m, nil
		case "A":
			// Check availability (Shift+A)
			return m.openAvailabilityDialog()
		}

		// Delegate to calendar grid for navigation
		m.calendarGrid, cmd = m.calendarGrid.Update(msg)
		cmds = append(cmds, cmd)

		// Reset event selection when date changes
		if keyStr == "h" || keyStr == "l" || keyStr == "j" || keyStr == "k" ||
			key.Code == tea.KeyLeft || key.Code == tea.KeyRight ||
			key.Code == tea.KeyUp || key.Code == tea.KeyDown {
			m.selectedEventIdx = 0
		}

	case tea.WindowSizeMsg:
		m.global.SetWindowSize(msg.Width, msg.Height)
		m.width = msg.Width
		m.height = msg.Height
		m.updateSizes()
		return m, nil

	case calendarsLoadedMsg:
		m.calendarsLoaded = true
		m.loading = false
		m.calendars = msg.calendars
		m.updateCalendarList()

		// Auto-select primary calendar
		for i := range m.calendars {
			if m.calendars[i].IsPrimary {
				m.selectedCalendar = &m.calendars[i]
				break
			}
		}
		// If no primary, select first
		if m.selectedCalendar == nil && len(m.calendars) > 0 {
			m.selectedCalendar = &m.calendars[0]
		}

		// Fetch events for selected calendar
		if m.selectedCalendar != nil {
			m.loadingEvents = true
			return m, tea.Batch(m.spinner.Tick, m.fetchEventsForMonth(m.selectedCalendar.ID, m.calendarGrid.GetCurrentMonth()))
		}
		return m, nil

	case eventsLoadedMsg:
		m.loadingEvents = false
		m.events = msg.events
		m.calendarGrid.SetEvents(m.events)
		m.global.SetStatus(fmt.Sprintf("Loaded %d events", len(m.events)), 0)
		m.selectedEventIdx = 0
		return m, nil

	case eventCreatedMsg:
		m.loading = false
		m.global.SetStatus(fmt.Sprintf("Event '%s' created", msg.event.Title), 0)
		// Refresh events
		if m.selectedCalendar != nil {
			m.loadingEvents = true
			return m, tea.Batch(m.spinner.Tick, m.fetchEventsForMonth(m.selectedCalendar.ID, m.calendarGrid.GetCurrentMonth()))
		}
		return m, nil

	case eventUpdatedMsg:
		m.loading = false
		m.global.SetStatus(fmt.Sprintf("Event '%s' updated", msg.event.Title), 0)
		// Refresh events
		if m.selectedCalendar != nil {
			m.loadingEvents = true
			return m, tea.Batch(m.spinner.Tick, m.fetchEventsForMonth(m.selectedCalendar.ID, m.calendarGrid.GetCurrentMonth()))
		}
		return m, nil

	case eventDeletedMsg:
		m.loading = false
		m.global.SetStatus("Event deleted", 0)
		// Refresh events
		if m.selectedCalendar != nil {
			m.loadingEvents = true
			return m, tea.Batch(m.spinner.Tick, m.fetchEventsForMonth(m.selectedCalendar.ID, m.calendarGrid.GetCurrentMonth()))
		}
		return m, nil

	case errMsg:
		m.err = msg.err
		m.loading = false
		m.loadingEvents = false
		return m, nil

	case spinner.TickMsg:
		if m.loading || m.loadingEvents {
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case components.CalendarMonthChangedMsg:
		// Reload events when month changes
		if m.selectedCalendar != nil {
			m.loadingEvents = true
			return m, tea.Batch(m.spinner.Tick, m.fetchEventsForMonth(m.selectedCalendar.ID, msg.Month))
		}
	}

	return m, tea.Batch(cmds...)
}

// updateEventForm handles updates when in event form mode.
func (m *CalendarScreen) updateEventForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case components.EventFormSubmitMsg:
		m.mode = CalendarModeView
		m.loading = true
		if msg.Mode == components.EventFormCreate {
			return m, tea.Batch(m.spinner.Tick, m.createEvent(msg.Request))
		}
		// Edit mode
		return m, tea.Batch(m.spinner.Tick, m.updateEvent(msg.EventID, msg.Request))

	case components.EventFormCancelMsg:
		m.mode = CalendarModeView
		m.eventForm = nil
		return m, nil

	case tea.WindowSizeMsg:
		m.global.SetWindowSize(msg.Width, msg.Height)
		m.width = msg.Width
		m.height = msg.Height
		if m.eventForm != nil {
			m.eventForm.SetSize(msg.Width-10, msg.Height-10)
		}
		return m, nil
	}

	// Forward to event form
	if m.eventForm != nil {
		var cmd tea.Cmd
		m.eventForm, cmd = m.eventForm.Update(msg)
		return m, cmd
	}

	return m, nil
}

// updateConfirmDelete handles updates when in confirm delete mode.
func (m *CalendarScreen) updateConfirmDelete(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case components.ConfirmDialogMsg:
		m.mode = CalendarModeView
		if msg.Result == components.ConfirmDialogResultConfirm {
			if eventID, ok := msg.Data.(string); ok {
				m.loading = true
				return m, tea.Batch(m.spinner.Tick, m.deleteEvent(eventID))
			}
		}
		m.confirmDialog = nil
		return m, nil

	case tea.WindowSizeMsg:
		m.global.SetWindowSize(msg.Width, msg.Height)
		m.width = msg.Width
		m.height = msg.Height
		if m.confirmDialog != nil {
			m.confirmDialog.SetSize(msg.Width, msg.Height)
		}
		return m, nil
	}

	// Forward to confirm dialog
	if m.confirmDialog != nil {
		var cmd tea.Cmd
		m.confirmDialog, cmd = m.confirmDialog.Update(msg)
		return m, cmd
	}

	return m, nil
}

// openEventForm opens the event form for creating or editing.
func (m *CalendarScreen) openEventForm(mode components.EventFormMode, event *domain.Event) (tea.Model, tea.Cmd) {
	if m.selectedCalendar == nil {
		m.global.SetStatus("No calendar selected", 1)
		return m, nil
	}
	if m.selectedCalendar.ReadOnly {
		m.global.SetStatus("Cannot modify events on read-only calendar", 1)
		return m, nil
	}

	m.eventForm = components.NewEventForm(m.theme, mode)
	m.eventForm.SetTimezone(m.timezone)
	m.eventForm.SetSize(m.width-10, m.height-10)

	if mode == components.EventFormCreate {
		// Set the selected date as default
		m.eventForm.SetDate(m.calendarGrid.GetSelectedDate())
	} else if event != nil {
		m.eventForm.SetEvent(event)
		m.selectedEvent = event
	}

	m.mode = CalendarModeEventForm
	return m, m.eventForm.Init()
}

// openDeleteConfirmation opens the delete confirmation dialog.
func (m *CalendarScreen) openDeleteConfirmation(event *domain.Event) (tea.Model, tea.Cmd) {
	if m.selectedCalendar == nil {
		m.global.SetStatus("No calendar selected", 1)
		return m, nil
	}
	if m.selectedCalendar.ReadOnly {
		m.global.SetStatus("Cannot delete events on read-only calendar", 1)
		return m, nil
	}
	if event.ReadOnly {
		m.global.SetStatus("Cannot delete read-only event", 1)
		return m, nil
	}

	title := event.Title
	if title == "" {
		title = "(No title)"
	}

	m.confirmDialog = components.NewConfirmDialog(
		m.theme,
		"Delete Event",
		fmt.Sprintf("Are you sure you want to delete '%s'?", title),
	)
	m.confirmDialog.SetData(event.ID)
	m.confirmDialog.SetButtonLabels("Delete", "Cancel")
	m.confirmDialog.SetSize(m.width, m.height)
	m.selectedEvent = event

	m.mode = CalendarModeConfirmDelete
	return m, nil
}

// openAvailabilityDialog opens the availability check dialog.
func (m *CalendarScreen) openAvailabilityDialog() (tea.Model, tea.Cmd) {
	m.availabilityDialog = components.NewAvailabilityDialog(m.theme)
	m.availabilityDialog.SetTimezone(m.timezone)
	m.availabilityDialog.SetSize(m.width, m.height)

	// Set date range to selected week
	selectedDate := m.calendarGrid.GetSelectedDate()
	startDate := selectedDate
	endDate := selectedDate.AddDate(0, 0, 7)
	m.availabilityDialog.SetDateRange(startDate, endDate)

	if m.selectedCalendar != nil {
		m.availabilityDialog.SetCalendarID(m.selectedCalendar.ID)
	}

	m.mode = CalendarModeAvailability
	return m, m.availabilityDialog.Init()
}

// updateAvailability handles updates when in availability check mode.
func (m *CalendarScreen) updateAvailability(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case components.AvailabilityCheckMsg:
		// Perform the availability check
		return m, m.checkAvailability(msg.Request)

	case components.AvailabilityResultsMsg:
		// Update the dialog with results
		if m.availabilityDialog != nil {
			m.availabilityDialog.SetResults(msg.Slots)
		}
		return m, nil

	case components.AvailabilityCancelMsg:
		m.mode = CalendarModeView
		m.availabilityDialog = nil
		return m, nil

	case components.AvailabilitySelectSlotMsg:
		// User selected a slot - create an event for it
		m.mode = CalendarModeView
		m.availabilityDialog = nil
		return m.createEventFromSlot(msg.Slot)

	case availabilityLoadedMsg:
		if m.availabilityDialog != nil {
			m.availabilityDialog.SetResults(msg.slots)
		}
		return m, nil

	case errMsg:
		if m.availabilityDialog != nil {
			m.availabilityDialog.SetError(msg.err)
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.global.SetWindowSize(msg.Width, msg.Height)
		m.width = msg.Width
		m.height = msg.Height
		if m.availabilityDialog != nil {
			m.availabilityDialog.SetSize(msg.Width, msg.Height)
		}
		return m, nil
	}

	// Forward to availability dialog
	if m.availabilityDialog != nil {
		var cmd tea.Cmd
		m.availabilityDialog, cmd = m.availabilityDialog.Update(msg)
		return m, cmd
	}

	return m, nil
}

// checkAvailability performs the availability check via the API.
func (m *CalendarScreen) checkAvailability(req *domain.AvailabilityRequest) tea.Cmd {
	return func() tea.Msg {
		m.global.RateLimiter.Wait()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		resp, err := m.global.Client.GetAvailability(ctx, req)
		if err != nil {
			return errMsg{err}
		}
		return availabilityLoadedMsg{slots: resp.Data.TimeSlots}
	}
}

// createEventFromSlot opens the event form pre-filled with the selected slot.
func (m *CalendarScreen) createEventFromSlot(slot domain.AvailableSlot) (tea.Model, tea.Cmd) {
	if m.selectedCalendar == nil {
		m.global.SetStatus("No calendar selected", 1)
		return m, nil
	}
	if m.selectedCalendar.ReadOnly {
		m.global.SetStatus("Cannot create events on read-only calendar", 1)
		return m, nil
	}

	m.eventForm = components.NewEventForm(m.theme, components.EventFormCreate)
	m.eventForm.SetTimezone(m.timezone)
	m.eventForm.SetSize(m.width-10, m.height-10)

	// Set the slot time
	startTime := time.Unix(slot.StartTime, 0)
	endTime := time.Unix(slot.EndTime, 0)
	if m.timezone != nil {
		startTime = startTime.In(m.timezone)
		endTime = endTime.In(m.timezone)
	}
	m.eventForm.SetDate(startTime)
	m.eventForm.SetTimeRange(startTime, endTime)

	m.mode = CalendarModeEventForm
	return m, m.eventForm.Init()
}

// availabilityLoadedMsg contains loaded availability slots.
type availabilityLoadedMsg struct {
	slots []domain.AvailableSlot
}

// View implements tea.Model.
func (m *CalendarScreen) View() tea.View {
	// Show event form if in form mode
	if m.mode == CalendarModeEventForm && m.eventForm != nil {
		return tea.NewView(m.eventForm.View())
	}

	// Show confirm dialog if in confirm mode
	if m.mode == CalendarModeConfirmDelete && m.confirmDialog != nil {
		return tea.NewView(m.confirmDialog.View())
	}

	// Show availability dialog if in availability mode
	if m.mode == CalendarModeAvailability && m.availabilityDialog != nil {
		return tea.NewView(m.availabilityDialog.View())
	}

	if m.err != nil {
		return tea.NewView(m.theme.Error_.Render(fmt.Sprintf("Error: %v\n\nPress Esc to go back", m.err)))
	}

	// Build header
	header := m.theme.Title.Render("Calendar")
	if m.selectedCalendar != nil {
		header += " " + m.theme.Subtitle.Render(fmt.Sprintf("(%s)", m.selectedCalendar.Name))
	}

	// Show loading spinner
	if m.loading || m.loadingEvents {
		header += " " + m.spinner.View()
	}

	// View mode indicator
	var viewMode string
	switch m.calendarGrid.GetViewMode() {
	case components.CalendarMonthView:
		viewMode = "Month"
	case components.CalendarWeekView:
		viewMode = "Week"
	case components.CalendarAgendaView:
		viewMode = "Agenda"
	}
	viewIndicator := m.theme.KeyBinding.Render(fmt.Sprintf(" [%s view]", viewMode))
	header += viewIndicator

	// Timezone indicator
	if m.timezone != nil && m.timezone != time.Local {
		tzIndicator := m.theme.Dimmed.Render(fmt.Sprintf(" [%s]", m.timezone.String()))
		header += tzIndicator
	}

	// Calculate layout dimensions
	scheduleWidth := 35
	if m.width < 100 {
		scheduleWidth = 0 // Hide schedule panel on narrow screens
	}
	gridWidth := m.width - scheduleWidth - 2
	if gridWidth < 50 {
		gridWidth = m.width
		scheduleWidth = 0
	}

	// Update grid size for side panel layout
	gridHeight := m.height - 6 // Reserve space for header and help
	if gridHeight < 10 {
		gridHeight = 10
	}
	m.calendarGrid.SetSize(gridWidth, gridHeight)

	// Calendar grid
	calendarView := m.calendarGrid.View()
	calendarLines := len(strings.Split(calendarView, "\n"))

	// Build main content with side-by-side layout
	var mainContent string
	if scheduleWidth > 0 {
		// Today's schedule panel - auto-sized and vertically centered
		schedulePanel := m.renderTodaySchedule(scheduleWidth, calendarLines)

		// Center the schedule panel vertically against the calendar
		centeredSchedule := lipgloss.Place(
			scheduleWidth,
			calendarLines,
			lipgloss.Center,
			lipgloss.Center,
			schedulePanel,
		)

		mainContent = lipgloss.JoinHorizontal(lipgloss.Top, calendarView, centeredSchedule)
	} else {
		mainContent = calendarView
	}

	// Help text
	help := m.theme.Help.Render("n: new  e: edit  d: delete  A: availability  m/w/g: views  t: today  [/]: month  h/l/j/k: nav  Ctrl+R: refresh  Esc: back")

	return tea.NewView(header + "\n\n" + mainContent + "\n" + help)
}

// updateSizes updates component sizes.
func (m *CalendarScreen) updateSizes() {
	if m.width > 0 && m.height > 0 {
		// Calendar grid takes most of the space
		gridHeight := m.height - 8 // Reserve space for header and help
		if gridHeight < 10 {
			gridHeight = 10
		}
		m.calendarGrid.SetSize(m.width, gridHeight)
	}
}

// updateCalendarList updates the calendar list.
func (m *CalendarScreen) updateCalendarList() {
	items := make([]list.Item, len(m.calendars))
	for i, cal := range m.calendars {
		items[i] = calendarItem{calendar: cal}
	}
	m.calendarList.SetItems(items)
}

// fetchCalendars fetches the list of calendars.
func (m *CalendarScreen) fetchCalendars() tea.Cmd {
	return func() tea.Msg {
		m.global.RateLimiter.Wait()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		calendars, err := m.global.Client.GetCalendars(ctx, m.global.GrantID)
		if err != nil {
			return errMsg{err}
		}
		return calendarsLoadedMsg{calendars}
	}
}

// fetchEventsForMonth fetches events for the given month.
func (m *CalendarScreen) fetchEventsForMonth(calendarID string, month time.Time) tea.Cmd {
	return func() tea.Msg {
		m.global.RateLimiter.Wait()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Calculate start and end of month with buffer for overflow
		startOfMonth := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, month.Location())
		endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second)

		// Add buffer to catch events from previous/next month that might show on the calendar
		start := startOfMonth.AddDate(0, 0, -7)
		end := endOfMonth.AddDate(0, 0, 7)

		params := &domain.EventQueryParams{
			Start: start.Unix(),
			End:   end.Unix(),
			Limit: 200,
		}

		events, err := m.global.Client.GetEvents(ctx, m.global.GrantID, calendarID, params)
		if err != nil {
			return errMsg{err}
		}
		return eventsLoadedMsg{events}
	}
}

// renderTodaySchedule renders the "Today's Schedule" panel for the right side.
// The panel auto-sizes based on content and will be vertically centered.
// maxHeight is used to limit the number of events shown.
func (m *CalendarScreen) renderTodaySchedule(width, maxHeight int) string {
	var lines []string

	// Header: "Today's Schedule" with date
	selectedDate := m.calendarGrid.GetSelectedDate()
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Foreground)

	headerText := "📅 Schedule"
	if m.calendarGrid.IsToday(selectedDate) {
		headerText = "📅 Today"
	}

	lines = append(lines, headerStyle.Render(headerText))
	dateStyle := lipgloss.NewStyle().Foreground(m.theme.Secondary)
	lines = append(lines, dateStyle.Render(selectedDate.Format("Monday, Jan 2")))
	lines = append(lines, "") // Empty line after header

	// Get events for selected date
	events := m.calendarGrid.GetEventsForDate(selectedDate)

	if len(events) == 0 {
		dimStyle := lipgloss.NewStyle().Foreground(m.theme.Dimmed.GetForeground())
		lines = append(lines, dimStyle.Render("No events scheduled"))
		lines = append(lines, "")
		lines = append(lines, dimStyle.Render("Press 'n' to create"))
	} else {
		// Calculate max events based on available height
		// Each event card is roughly 4-5 lines, reserve 6 for header/footer
		availableForEvents := maxHeight - 8
		maxEvents := availableForEvents / 5
		if maxEvents < 2 {
			maxEvents = 2
		}
		if maxEvents > 8 {
			maxEvents = 8 // Cap at 8 events for readability
		}
		if maxEvents > len(events) {
			maxEvents = len(events)
		}

		// Event count indicator
		countStyle := lipgloss.NewStyle().Foreground(m.theme.Dimmed.GetForeground())
		if len(events) > 1 {
			lines = append(lines, countStyle.Render(fmt.Sprintf("%d events", len(events))))
			lines = append(lines, "")
		}

		// Render each event as a card
		for i := 0; i < maxEvents; i++ {
			evt := events[i]
			card := m.renderEventCard(evt, width-6, i == m.selectedEventIdx)
			cardLines := strings.Split(card, "\n")
			lines = append(lines, cardLines...)
			// Add spacing between cards
			if i < maxEvents-1 {
				lines = append(lines, "")
			}
		}

		// Show scroll indicator if more events
		if len(events) > maxEvents {
			lines = append(lines, "")
			scrollStyle := lipgloss.NewStyle().
				Foreground(m.theme.Secondary).
				Bold(true)
			lines = append(lines, scrollStyle.Render(fmt.Sprintf("↓ %d more (J/K to scroll)", len(events)-maxEvents)))
		}
	}

	// Join all lines into content
	content := strings.Join(lines, "\n")

	// Panel style - auto-sized based on content
	panelStyle := lipgloss.NewStyle().
		Width(width-2).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Primary).
		Padding(0, 1)

	return panelStyle.Render(content)
}

// renderEventCard renders a single event as a card for the schedule panel.
func (m *CalendarScreen) renderEventCard(evt domain.Event, width int, selected bool) string {
	var b strings.Builder

	// Event card style with colored left border
	bgColor := m.theme.Primary
	if evt.Status == "cancelled" {
		bgColor = m.theme.Error
	} else if !evt.Busy {
		bgColor = m.theme.Success
	}

	cardStyle := lipgloss.NewStyle().
		Width(width).
		BorderStyle(lipgloss.NormalBorder()).
		BorderLeft(true).
		BorderForeground(bgColor).
		Padding(0, 1)

	if selected {
		cardStyle = cardStyle.Background(m.theme.Primary).Foreground(lipgloss.Color("#000000"))
	}

	// Time range
	timeStyle := lipgloss.NewStyle().Foreground(m.theme.Secondary)
	if selected {
		timeStyle = timeStyle.Foreground(lipgloss.Color("#333333"))
	}

	var timeStr string
	if evt.When.IsAllDay() {
		timeStr = "All day"
	} else {
		startTime := evt.When.StartDateTime()
		endTime := evt.When.EndDateTime()
		if m.timezone != nil {
			startTime = startTime.In(m.timezone)
			endTime = endTime.In(m.timezone)
		}
		timeStr = fmt.Sprintf("%s - %s", startTime.Format("3:04 PM"), endTime.Format("3:04 PM"))
	}
	b.WriteString(timeStyle.Render(timeStr))
	b.WriteString("\n")

	// Title
	title := evt.Title
	if title == "" {
		title = "(No title)"
	}
	titleStyle := lipgloss.NewStyle().Bold(true)
	if selected {
		titleStyle = titleStyle.Foreground(lipgloss.Color("#000000"))
	}
	if evt.Status == "cancelled" {
		titleStyle = titleStyle.Strikethrough(true)
	}
	// Truncate title
	maxLen := width - 2
	if len(title) > maxLen && maxLen > 0 {
		title = title[:maxLen-1] + "…"
	}
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n")

	// Conferencing indicator
	if evt.Conferencing != nil && evt.Conferencing.Details != nil && evt.Conferencing.Details.URL != "" {
		confStyle := lipgloss.NewStyle().Foreground(m.theme.Info)
		if selected {
			confStyle = confStyle.Foreground(lipgloss.Color("#333333"))
		}
		provider := "Video call"
		if evt.Conferencing.Provider != "" {
			provider = evt.Conferencing.Provider
		}
		b.WriteString(confStyle.Render("📹 " + provider))
	}

	return cardStyle.Render(b.String())
}

// renderEventSummary renders a single event summary.
func (m *CalendarScreen) renderEventSummary(evt domain.Event) string {
	var b strings.Builder

	// Time
	timeStyle := lipgloss.NewStyle().Foreground(m.theme.Secondary)
	if evt.When.IsAllDay() {
		b.WriteString(timeStyle.Render("  All day  "))
	} else {
		startTime := evt.When.StartDateTime()
		endTime := evt.When.EndDateTime()
		timeStr := fmt.Sprintf("  %s - %s  ", startTime.Format("3:04 PM"), endTime.Format("3:04 PM"))
		b.WriteString(timeStyle.Render(timeStr))
	}

	// Title
	title := evt.Title
	if title == "" {
		title = "(No title)"
	}
	titleStyle := lipgloss.NewStyle()
	if evt.Status == "cancelled" {
		titleStyle = titleStyle.Strikethrough(true).Foreground(m.theme.Dimmed.GetForeground())
	}
	b.WriteString(titleStyle.Render(title))

	// Busy indicator - show "Free" when not busy
	if !evt.Busy {
		freeStyle := lipgloss.NewStyle().Foreground(m.theme.Success)
		b.WriteString(freeStyle.Render(" (Free)"))
	}

	// Location
	if evt.Location != "" {
		locStyle := lipgloss.NewStyle().Foreground(m.theme.Dimmed.GetForeground())
		b.WriteString(locStyle.Render(fmt.Sprintf(" 📍 %s", truncate(evt.Location, 30))))
	}

	// Conferencing
	if evt.Conferencing != nil && evt.Conferencing.Details != nil && evt.Conferencing.Details.URL != "" {
		confStyle := lipgloss.NewStyle().Foreground(m.theme.Info)
		b.WriteString(confStyle.Render(" 📹"))
	}

	// Participants count
	if len(evt.Participants) > 1 {
		partStyle := lipgloss.NewStyle().Foreground(m.theme.Dimmed.GetForeground())
		b.WriteString(partStyle.Render(fmt.Sprintf(" 👥 %d", len(evt.Participants))))
	}

	return b.String()
}

// createEvent creates a new event via the API.
func (m *CalendarScreen) createEvent(req *domain.CreateEventRequest) tea.Cmd {
	return func() tea.Msg {
		m.global.RateLimiter.Wait()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		event, err := m.global.Client.CreateEvent(ctx, m.global.GrantID, m.selectedCalendar.ID, req)
		if err != nil {
			return errMsg{err}
		}
		return eventCreatedMsg{event}
	}
}

// updateEvent updates an existing event via the API.
func (m *CalendarScreen) updateEvent(eventID string, req *domain.CreateEventRequest) tea.Cmd {
	return func() tea.Msg {
		m.global.RateLimiter.Wait()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Convert CreateEventRequest to UpdateEventRequest
		updateReq := &domain.UpdateEventRequest{
			Title:       &req.Title,
			Description: &req.Description,
			Location:    &req.Location,
			When:        &req.When,
			Busy:        &req.Busy,
		}

		event, err := m.global.Client.UpdateEvent(ctx, m.global.GrantID, m.selectedCalendar.ID, eventID, updateReq)
		if err != nil {
			return errMsg{err}
		}
		return eventUpdatedMsg{event}
	}
}

// deleteEvent deletes an event via the API.
func (m *CalendarScreen) deleteEvent(eventID string) tea.Cmd {
	return func() tea.Msg {
		m.global.RateLimiter.Wait()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err := m.global.Client.DeleteEvent(ctx, m.global.GrantID, m.selectedCalendar.ID, eventID)
		if err != nil {
			return errMsg{err}
		}
		return eventDeletedMsg{}
	}
}

// Message types

type calendarsLoadedMsg struct {
	calendars []domain.Calendar
}

type eventsLoadedMsg struct {
	events []domain.Event
}

type eventCreatedMsg struct {
	event *domain.Event
}

type eventUpdatedMsg struct {
	event *domain.Event
}

type eventDeletedMsg struct{}

// calendarItem is a list item for calendar selection.
type calendarItem struct {
	calendar domain.Calendar
}

func (i calendarItem) Title() string {
	name := i.calendar.Name
	if i.calendar.IsPrimary {
		name += " ★"
	}
	return name
}

func (i calendarItem) Description() string {
	if i.calendar.Description != "" {
		return i.calendar.Description
	}
	return ""
}

func (i calendarItem) FilterValue() string {
	return i.calendar.Name
}
