package tui

import (
	"fmt"
	"slices"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/mqasimca/nylas/internal/domain"
)

// CalendarViewMode represents the calendar display mode.

func (c *CalendarView) Draw(screen tcell.Screen) {
	c.DrawForSubclass(screen, c)
	x, y, width, height := c.GetInnerRect()

	switch c.viewMode {
	case CalendarMonthView:
		c.drawMonthView(screen, x, y, width, height)
	case CalendarWeekView:
		c.drawWeekView(screen, x, y, width, height)
	case CalendarAgendaView:
		c.drawAgendaView(screen, x, y, width, height)
	}
}

func (c *CalendarView) drawAgendaView(screen tcell.Screen, x, y, width, height int) {
	titleColor := c.styles.TitleFg
	headerColor := c.styles.InfoColor
	borderColor := c.styles.BorderColor
	eventColor := c.styles.FgColor
	mutedColor := c.styles.BorderColor
	todayColor := c.styles.SuccessColor

	// Draw header
	headerText := "Upcoming Events"
	headerX := x + (width-len(headerText))/2
	for i, ch := range headerText {
		style := tcell.StyleDefault.Foreground(titleColor).Bold(true)
		screen.SetContent(headerX+i, y, ch, nil, style)
	}
	y += 2

	// Collect and sort upcoming events
	type agendaItem struct {
		date  time.Time
		event domain.Event
	}

	var items []agendaItem
	today := time.Now()
	endDate := today.AddDate(0, 1, 0) // Next month

	for _, evt := range c.events {
		startDate := evt.When.StartDateTime()
		if startDate.After(today.AddDate(0, 0, -1)) && startDate.Before(endDate) {
			items = append(items, agendaItem{date: startDate, event: evt})
		}
	}

	slices.SortFunc(items, func(a, b agendaItem) int {
		if a.date.Before(b.date) {
			return -1
		}
		if a.date.After(b.date) {
			return 1
		}
		return 0
	})

	if len(items) == 0 {
		noEvt := "No upcoming events"
		for i, ch := range noEvt {
			screen.SetContent(x+i, y, ch, nil, tcell.StyleDefault.Foreground(mutedColor))
		}
		return
	}

	// Draw events grouped by date
	currentDate := ""
	row := 0
	for _, item := range items {
		if y+row >= y+height-2 {
			break
		}

		dateStr := item.date.Format("2006-01-02")
		if dateStr != currentDate {
			currentDate = dateStr

			// Draw date header
			isToday := item.date.Year() == today.Year() && item.date.YearDay() == today.YearDay()
			dateHeader := item.date.Format("Monday, January 2")
			if isToday {
				dateHeader = "Today - " + dateHeader
			}

			dateStyle := tcell.StyleDefault.Foreground(headerColor).Bold(true)
			if isToday {
				dateStyle = tcell.StyleDefault.Foreground(todayColor).Bold(true)
			}

			for i, ch := range dateHeader {
				screen.SetContent(x+i, y+row, ch, nil, dateStyle)
			}
			row++

			// Draw separator
			for i := 0; i < width; i++ {
				screen.SetContent(x+i, y+row, '─', nil, tcell.StyleDefault.Foreground(borderColor))
			}
			row++
		}

		// Draw event
		timeStr := "All day"
		if !item.event.When.IsAllDay() {
			timeStr = item.event.When.StartDateTime().Format("3:04 PM")
		}

		eventLine := fmt.Sprintf("  %s  %s", timeStr, item.event.Title)
		if len(eventLine) > width {
			eventLine = eventLine[:width-1] + "…"
		}

		for i, ch := range eventLine {
			style := tcell.StyleDefault.Foreground(eventColor)
			if i < 10 {
				style = tcell.StyleDefault.Foreground(mutedColor)
			}
			screen.SetContent(x+i, y+row, ch, nil, style)
		}
		row++
	}
}

// InputHandler returns the input handler for the calendar.
