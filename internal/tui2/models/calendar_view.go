package models

import (
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/mqasimca/nylas/internal/tui2/components"
)

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
