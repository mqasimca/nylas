package tui

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
)

// CalendarViewMode represents the calendar display mode.

func (c *CalendarView) drawMonthView(screen tcell.Screen, x, y, width, height int) {
	titleColor := c.styles.TitleFg
	headerColor := c.styles.InfoColor
	borderColor := c.styles.BorderColor
	todayColor := c.styles.SuccessColor
	selectedBg := c.styles.TableSelectBg
	eventColor := c.styles.FgColor
	mutedColor := c.styles.BorderColor

	// Draw current calendar name
	calName := "No Calendar"
	if cal := c.GetCurrentCalendar(); cal != nil {
		calName = cal.Name
		if len(calName) > 30 {
			calName = calName[:27] + "..."
		}
	}
	calText := fmt.Sprintf("📅 %s [c]hange", calName)
	for i, ch := range calText {
		style := tcell.StyleDefault.Foreground(headerColor)
		if ch == 'c' && i > 0 {
			style = tcell.StyleDefault.Foreground(titleColor).Bold(true)
		}
		screen.SetContent(x+1+i, y, ch, nil, style)
	}

	// Draw view mode indicator on the right
	modeText := "[M]onth [W]eek [A]genda"
	modeX := x + width - len(modeText) - 1
	for i, ch := range modeText {
		style := tcell.StyleDefault.Foreground(mutedColor)
		if ch == 'M' || ch == 'W' || ch == 'A' {
			style = tcell.StyleDefault.Foreground(headerColor).Bold(true)
		}
		screen.SetContent(modeX+i, y, ch, nil, style)
	}
	y += 1

	// Draw month/year header
	monthYear := c.currentMonth.Format("January 2006")
	headerText := fmt.Sprintf("◀  %s  ▶", monthYear)
	headerX := x + (width-len(headerText))/2
	for i, ch := range headerText {
		style := tcell.StyleDefault.Foreground(titleColor).Bold(true)
		screen.SetContent(headerX+i, y, ch, nil, style)
	}

	y += 2

	// Calculate cell dimensions based on available space
	c.cellWidth = width / 7
	if c.cellWidth < 12 {
		c.cellWidth = 12
	}
	// 6 weeks max in a month view, plus 1 row for day headers
	c.cellHeight = (height - 6) / 6
	if c.cellHeight < 4 {
		c.cellHeight = 4
	}

	// Draw day headers
	days := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	for i, day := range days {
		dx := x + i*c.cellWidth + (c.cellWidth-len(day))/2
		style := tcell.StyleDefault.Foreground(headerColor).Bold(true)
		for j, ch := range day {
			screen.SetContent(dx+j, y, ch, nil, style)
		}
	}
	y += 1

	// Draw separator line
	for i := 0; i < width; i++ {
		screen.SetContent(x+i, y, '─', nil, tcell.StyleDefault.Foreground(borderColor))
	}
	y += 1

	// Calculate first day of month
	firstDay := c.currentMonth
	firstWeekday := int(firstDay.Weekday())

	// Start from the first cell (might be previous month)
	startDate := firstDay.AddDate(0, 0, -firstWeekday)

	// Draw calendar grid (6 weeks to cover all cases)
	today := time.Now()
	for week := 0; week < 6; week++ {
		for day := 0; day < 7; day++ {
			date := startDate.AddDate(0, 0, week*7+day)
			cellX := x + day*c.cellWidth
			cellY := y + week*c.cellHeight

			// Determine cell style
			isCurrentMonth := date.Month() == c.currentMonth.Month()
			isToday := date.Year() == today.Year() && date.YearDay() == today.YearDay()
			isSelected := date.Year() == c.selectedDate.Year() && date.YearDay() == c.selectedDate.YearDay()

			// Draw cell background for selected
			if isSelected {
				for cy := 0; cy < c.cellHeight && cellY+cy < y+height-4; cy++ {
					for cx := 0; cx < c.cellWidth-1; cx++ {
						screen.SetContent(cellX+cx, cellY+cy, ' ', nil, tcell.StyleDefault.Background(selectedBg))
					}
				}
			}

			// Draw day number
			dayNum := fmt.Sprintf("%d", date.Day())
			dayStyle := tcell.StyleDefault.Foreground(eventColor)
			if !isCurrentMonth {
				dayStyle = tcell.StyleDefault.Foreground(mutedColor)
			}
			if isToday {
				dayStyle = tcell.StyleDefault.Foreground(todayColor).Bold(true)
			}
			if isSelected {
				dayStyle = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(selectedBg).Bold(true)
			}

			// Center the day number in the cell
			dayX := cellX + 1
			for j, ch := range dayNum {
				screen.SetContent(dayX+j, cellY, ch, nil, dayStyle)
			}

			// Draw today indicator
			if isToday && !isSelected {
				screen.SetContent(dayX+len(dayNum)+1, cellY, '●', nil, tcell.StyleDefault.Foreground(todayColor))
			}

			// Draw events for this day
			dateKey := date.Format("2006-01-02")
			events := c.eventsByDate[dateKey]
			maxEvents := c.cellHeight - 1
			if maxEvents > 3 {
				maxEvents = 3
			}

			for i, evt := range events {
				if i >= maxEvents {
					// Show "+N more" indicator
					more := len(events) - maxEvents + 1
					moreText := fmt.Sprintf("+%d more", more)
					moreStyle := tcell.StyleDefault.Foreground(mutedColor)
					if isSelected {
						moreStyle = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(selectedBg)
					}
					for j, ch := range moreText {
						if cellX+1+j < cellX+c.cellWidth-1 {
							screen.SetContent(cellX+1+j, cellY+i+1, ch, nil, moreStyle)
						}
					}
					break
				}

				// Truncate event title to fit cell
				title := evt.Title
				maxLen := c.cellWidth - 3
				if len(title) > maxLen {
					title = title[:maxLen-1] + "…"
				}

				// Event color indicator
				evtStyle := tcell.StyleDefault.Foreground(eventColor)
				if isSelected {
					evtStyle = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(selectedBg)
				}
				if !isCurrentMonth {
					evtStyle = tcell.StyleDefault.Foreground(mutedColor)
				}

				// Draw event dot and title with calendar color
				dotColor := c.getCalendarColor()
				if dotColor == tcell.ColorDefault {
					dotColor = c.styles.InfoColor
				}
				screen.SetContent(cellX+1, cellY+i+1, '•', nil, tcell.StyleDefault.Foreground(dotColor))
				for j, ch := range title {
					if cellX+2+j < cellX+c.cellWidth-1 {
						screen.SetContent(cellX+2+j, cellY+i+1, ch, nil, evtStyle)
					}
				}
			}
		}
	}
}
