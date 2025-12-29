// Package components provides reusable Bubble Tea components.
package components

import (
	"slices"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/mqasimca/nylas/internal/domain"
)

func (c *CalendarGrid) renderAgendaView() string {
	var b strings.Builder

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(c.theme.Primary)
	b.WriteString(headerStyle.Render("Upcoming Events"))
	b.WriteString("\n\n")

	if len(c.events) == 0 {
		dimStyle := lipgloss.NewStyle().Foreground(c.theme.Dimmed.GetForeground())
		b.WriteString(dimStyle.Render("No upcoming events"))
		return b.String()
	}

	// Get events from selected date forward, sorted by date
	now := c.selectedDate
	var upcomingEvents []domain.Event
	for _, evt := range c.events {
		startDate := evt.When.StartDateTime()
		if !startDate.Before(now) || c.IsToday(startDate) {
			upcomingEvents = append(upcomingEvents, evt)
		}
	}

	slices.SortFunc(upcomingEvents, func(a, b domain.Event) int {
		aTime := a.When.StartDateTime()
		bTime := b.When.StartDateTime()
		if aTime.Before(bTime) {
			return -1
		}
		if aTime.After(bTime) {
			return 1
		}
		return 0
	})

	// Show up to 10 events
	maxEvents := 10
	if len(upcomingEvents) < maxEvents {
		maxEvents = len(upcomingEvents)
	}

	var currentDateStr string
	for i := 0; i < maxEvents; i++ {
		evt := upcomingEvents[i]
		dateStr := evt.When.StartDateTime().Format("Mon, Jan 2")
		if dateStr != currentDateStr {
			currentDateStr = dateStr
			dateStyle := lipgloss.NewStyle().Bold(true).Foreground(c.theme.Secondary)
			b.WriteString(dateStyle.Render(dateStr))
			b.WriteString("\n")
		}
		b.WriteString(c.renderEventLine(evt))
		b.WriteString("\n")
	}

	return b.String()
}

// renderEventLine renders a single event line.
func (c *CalendarGrid) renderEventLine(evt domain.Event) string {
	var b strings.Builder

	// Time
	startTime := evt.When.StartDateTime()
	if c.timezone != nil {
		startTime = startTime.In(c.timezone)
	}

	timeStyle := lipgloss.NewStyle().Foreground(c.theme.Secondary)
	if evt.When.IsAllDay() {
		b.WriteString(timeStyle.Render("All day"))
	} else {
		b.WriteString(timeStyle.Render(startTime.Format("3:04 PM")))
	}
	b.WriteString(" ")

	// Title
	titleStyle := lipgloss.NewStyle()
	if evt.Status == "cancelled" {
		titleStyle = titleStyle.Strikethrough(true).Foreground(c.theme.Dimmed.GetForeground())
	}
	title := evt.Title
	if title == "" {
		title = "(No title)"
	}
	b.WriteString(titleStyle.Render(title))

	// Location indicator
	if evt.Location != "" {
		locStyle := lipgloss.NewStyle().Foreground(c.theme.Dimmed.GetForeground())
		b.WriteString(locStyle.Render(" 📍"))
	}

	// Conferencing indicator
	if evt.Conferencing != nil && evt.Conferencing.Details != nil && evt.Conferencing.Details.URL != "" {
		confStyle := lipgloss.NewStyle().Foreground(c.theme.Dimmed.GetForeground())
		b.WriteString(confStyle.Render(" 📹"))
	}

	return b.String()
}

// getDayNames returns the day name headers based on settings.
func (c *CalendarGrid) getDayNames() []string {
	if c.firstDayMon {
		if c.showWeekends {
			return []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
		}
		return []string{"Mon", "Tue", "Wed", "Thu", "Fri"}
	}
	if c.showWeekends {
		return []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	}
	return []string{"Mon", "Tue", "Wed", "Thu", "Fri"}
}

// getMonthDays returns all days to display for the current month.
func (c *CalendarGrid) getMonthDays() []time.Time {
	// First day of the month
	firstDay := c.currentMonth

	// Find start of calendar (first day of week containing first of month)
	start := c.getWeekStart(firstDay)

	// Last day of the month
	lastDay := firstDay.AddDate(0, 1, -1)

	// Find end of calendar (last day of week containing last of month)
	end := c.getWeekEnd(lastDay)

	var days []time.Time
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		if !c.showWeekends && (d.Weekday() == time.Saturday || d.Weekday() == time.Sunday) {
			continue
		}
		days = append(days, d)
	}

	return days
}

// getWeekStart returns the start of the week containing the given date.
func (c *CalendarGrid) getWeekStart(date time.Time) time.Time {
	weekday := int(date.Weekday())
	if c.firstDayMon {
		// Monday = 0, Sunday = 6
		weekday = (weekday + 6) % 7
	}
	return date.AddDate(0, 0, -weekday)
}

// getWeekEnd returns the end of the week containing the given date.
func (c *CalendarGrid) getWeekEnd(date time.Time) time.Time {
	weekday := int(date.Weekday())
	if c.firstDayMon {
		// Monday = 0, Sunday = 6
		weekday = (weekday + 6) % 7
	}
	daysToEnd := 6 - weekday
	return date.AddDate(0, 0, daysToEnd)
}

// CalendarDateSelectedMsg is sent when a date is selected.
type CalendarDateSelectedMsg struct {
	Date time.Time
}

// CalendarEventSelectedMsg is sent when an event is selected.
type CalendarEventSelectedMsg struct {
	Event *domain.Event
}

// CalendarMonthChangedMsg is sent when the month changes.
type CalendarMonthChangedMsg struct {
	Month time.Time
}
