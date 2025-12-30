package models

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/mqasimca/nylas/internal/domain"
)

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
