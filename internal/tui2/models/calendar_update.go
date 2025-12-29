package models

import (
	"context"
	"fmt"
	"time"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/tui2/components"
)

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
