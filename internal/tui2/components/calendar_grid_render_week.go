// Package components provides reusable Bubble Tea components.
package components

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/mqasimca/nylas/internal/domain"
)

func (c *CalendarGrid) renderWeekView() string {
	var b strings.Builder

	weekStart := c.getWeekStart(c.selectedDate)
	numDays := 7
	if !c.showWeekends {
		numDays = 5
	}

	// Time gutter width and day column calculations
	timeGutterWidth := 8 // "12:00 PM"
	availableWidth := c.width - timeGutterWidth - 2
	dayWidth := availableWidth / numDays
	if dayWidth < 10 {
		dayWidth = 10
	}

	// === HEADER: Day names with dates ===
	headerRow := strings.Repeat(" ", timeGutterWidth)
	for i := 0; i < numDays; i++ {
		date := weekStart.AddDate(0, 0, i)
		if !c.showWeekends && (date.Weekday() == time.Saturday || date.Weekday() == time.Sunday) {
			continue
		}

		dayName := date.Format("Mon")
		dayNum := fmt.Sprintf("%d", date.Day())
		cellContent := fmt.Sprintf("%s %s", dayName, dayNum)

		headerStyle := lipgloss.NewStyle().
			Width(dayWidth).
			Align(lipgloss.Center).
			Bold(true)

		if c.IsToday(date) {
			headerStyle = headerStyle.Foreground(c.theme.Success)
		} else if c.IsSelected(date) {
			headerStyle = headerStyle.Foreground(c.theme.Primary)
		} else {
			headerStyle = headerStyle.Foreground(c.theme.Foreground)
		}

		headerRow += headerStyle.Render(cellContent)
	}
	b.WriteString(headerRow)
	b.WriteString("\n")

	// === ALL-DAY EVENTS ROW ===
	allDayEvents := c.getAllDayEventsForWeek(weekStart, numDays)
	if len(allDayEvents) > 0 {
		allDayRow := c.renderAllDayRow(weekStart, numDays, timeGutterWidth, dayWidth, allDayEvents)
		b.WriteString(allDayRow)
		b.WriteString("\n")
	}

	// Separator line
	separatorWidth := c.width - 1
	if separatorWidth < 1 {
		separatorWidth = 1
	}
	separatorStyle := lipgloss.NewStyle().Foreground(c.theme.Dimmed.GetForeground())
	b.WriteString(separatorStyle.Render(strings.Repeat("─", separatorWidth)))
	b.WriteString("\n")

	// === TIME GRID: Hours 8 AM to 8 PM ===
	startHour := 8
	endHour := 20
	if c.workingHours != nil {
		startHour = max(c.workingHours.StartHour-1, 0)
		endHour = min(c.workingHours.EndHour+1, 23)
	}

	// Calculate how many hours we can show based on available height
	// Reserve lines for header (1), all-day row (1-2), separator (1), and margin (2)
	reservedLines := 5
	if len(allDayEvents) > 0 {
		reservedLines++
	}
	availableLines := c.height - reservedLines
	if availableLines < 4 {
		availableLines = 4
	}

	// Show fewer hours if space is limited
	hoursToShow := endHour - startHour
	if hoursToShow > availableLines {
		hoursToShow = availableLines
	}
	endHour = startHour + hoursToShow

	// Build event map for quick lookup: map[dayIndex][hour][]Event
	timedEvents := c.getTimedEventsForWeek(weekStart, numDays)

	timeStyle := lipgloss.NewStyle().
		Width(timeGutterWidth).
		Foreground(c.theme.Secondary).
		Align(lipgloss.Right)

	for hour := startHour; hour < endHour; hour++ {
		// Time label
		hourLabel := c.formatHour(hour)
		row := timeStyle.Render(hourLabel) + " "

		// Day columns
		for dayIdx := 0; dayIdx < numDays; dayIdx++ {
			date := weekStart.AddDate(0, 0, dayIdx)
			if !c.showWeekends && (date.Weekday() == time.Saturday || date.Weekday() == time.Sunday) {
				continue
			}

			dayKey := date.Format("2006-01-02")
			cellContent := c.renderTimeSlotCell(timedEvents[dayKey], hour, dayWidth, date)
			row += cellContent
		}

		b.WriteString(row)
		b.WriteString("\n")
	}

	return b.String()
}

// getAllDayEventsForWeek collects all-day events for the week.
func (c *CalendarGrid) getAllDayEventsForWeek(weekStart time.Time, numDays int) map[string][]domain.Event {
	allDayEvents := make(map[string][]domain.Event)
	for i := 0; i < numDays; i++ {
		date := weekStart.AddDate(0, 0, i)
		dayKey := date.Format("2006-01-02")
		events := c.GetEventsForDate(date)
		for _, evt := range events {
			if evt.When.IsAllDay() {
				allDayEvents[dayKey] = append(allDayEvents[dayKey], evt)
			}
		}
	}
	return allDayEvents
}

// getTimedEventsForWeek collects timed events for the week.
func (c *CalendarGrid) getTimedEventsForWeek(weekStart time.Time, numDays int) map[string][]domain.Event {
	timedEvents := make(map[string][]domain.Event)
	for i := 0; i < numDays; i++ {
		date := weekStart.AddDate(0, 0, i)
		dayKey := date.Format("2006-01-02")
		events := c.GetEventsForDate(date)
		for _, evt := range events {
			if !evt.When.IsAllDay() {
				timedEvents[dayKey] = append(timedEvents[dayKey], evt)
			}
		}
	}
	return timedEvents
}

// renderAllDayRow renders the all-day events section.
func (c *CalendarGrid) renderAllDayRow(weekStart time.Time, numDays, gutterWidth, dayWidth int, allDayEvents map[string][]domain.Event) string {
	allDayLabel := lipgloss.NewStyle().
		Width(gutterWidth).
		Foreground(c.theme.Secondary).
		Align(lipgloss.Right).
		Render("All day")

	row := allDayLabel + " "

	for i := 0; i < numDays; i++ {
		date := weekStart.AddDate(0, 0, i)
		if !c.showWeekends && (date.Weekday() == time.Saturday || date.Weekday() == time.Sunday) {
			continue
		}

		dayKey := date.Format("2006-01-02")
		events := allDayEvents[dayKey]

		cellStyle := lipgloss.NewStyle().
			Width(dayWidth).
			Align(lipgloss.Left)

		if len(events) > 0 {
			// Show first all-day event with colored block
			evt := events[0]
			title := evt.Title
			if title == "" {
				title = "(No title)"
			}
			// Truncate to fit
			maxLen := dayWidth - 2
			if len(title) > maxLen {
				title = title[:maxLen-1] + "…"
			}

			eventStyle := lipgloss.NewStyle().
				Background(c.theme.Primary).
				Foreground(lipgloss.Color("#FFFFFF")).
				Padding(0, 1)

			cellContent := eventStyle.Render(title)
			if len(events) > 1 {
				moreStyle := lipgloss.NewStyle().Foreground(c.theme.Secondary)
				cellContent += moreStyle.Render(fmt.Sprintf("+%d", len(events)-1))
			}
			row += cellStyle.Render(cellContent)
		} else {
			row += cellStyle.Render("")
		}
	}

	return row
}

// renderTimeSlotCell renders a single cell in the time grid.
func (c *CalendarGrid) renderTimeSlotCell(dayEvents []domain.Event, hour, width int, date time.Time) string {
	// Find events that overlap this hour
	var eventsAtHour []domain.Event
	for _, evt := range dayEvents {
		startTime := evt.When.StartDateTime()
		endTime := evt.When.EndDateTime()
		if c.timezone != nil {
			startTime = startTime.In(c.timezone)
			endTime = endTime.In(c.timezone)
		}

		evtStartHour := startTime.Hour()
		evtEndHour := endTime.Hour()
		if endTime.Minute() > 0 {
			evtEndHour++ // Round up if there are extra minutes
		}

		// Check if this event overlaps with this hour
		if hour >= evtStartHour && hour < evtEndHour {
			eventsAtHour = append(eventsAtHour, evt)
		}
	}

	cellStyle := lipgloss.NewStyle().Width(width)

	// Dim non-working hours
	isWorkingTime := true
	if c.workingHours != nil {
		isWorkingTime = c.workingHours.IsWorkingHour(hour) && c.workingHours.IsWorkingDay(date.Weekday())
	}

	if len(eventsAtHour) == 0 {
		// Empty cell - show a subtle indicator
		if c.IsSelected(date) && hour == time.Now().Hour() {
			// Current hour on selected day
			nowStyle := lipgloss.NewStyle().
				Width(width).
				Foreground(c.theme.Success)
			return nowStyle.Render("▸")
		}
		// Use dimmer dots for non-working hours
		if isWorkingTime {
			return cellStyle.Render(strings.Repeat("·", min(width-1, 3)))
		}
		dimStyle := cellStyle.Foreground(c.theme.Dimmed.GetForeground())
		return dimStyle.Render(strings.Repeat("·", min(width-1, 2)))
	}

	// Render event block
	evt := eventsAtHour[0]
	startTime := evt.When.StartDateTime()
	if c.timezone != nil {
		startTime = startTime.In(c.timezone)
	}

	// Only show event title on the start hour
	if startTime.Hour() == hour {
		title := evt.Title
		if title == "" {
			title = "(No title)"
		}
		// Truncate to fit
		maxLen := width - 2
		if len(title) > maxLen && maxLen > 0 {
			title = title[:max(maxLen-1, 0)] + "…"
		}

		// Color based on event properties
		bgColor := c.theme.Primary
		if evt.Status == "cancelled" {
			bgColor = c.theme.Error
		} else if !evt.Busy {
			bgColor = c.theme.Success
		}

		eventStyle := lipgloss.NewStyle().
			Background(bgColor).
			Foreground(lipgloss.Color("#FFFFFF")).
			Width(width - 1).
			MaxWidth(width - 1)

		return eventStyle.Render(title)
	}

	// Continuation of event - show continuation bar
	bgColor := c.theme.Primary
	if evt.Status == "cancelled" {
		bgColor = c.theme.Error
	} else if !evt.Busy {
		bgColor = c.theme.Success
	}

	contStyle := lipgloss.NewStyle().
		Background(bgColor).
		Width(width - 1).
		MaxWidth(width - 1)
	return contStyle.Render(strings.Repeat(" ", width-2))
}

// formatHour formats an hour (0-23) as a 12-hour time string.
func (c *CalendarGrid) formatHour(hour int) string {
	if hour == 0 {
		return "12 AM"
	} else if hour < 12 {
		return fmt.Sprintf("%d AM", hour)
	} else if hour == 12 {
		return "12 PM"
	} else {
		return fmt.Sprintf("%d PM", hour-12)
	}
}

// renderAgendaView renders an agenda list of upcoming events.
