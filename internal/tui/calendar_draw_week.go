package tui

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/mqasimca/nylas/internal/cli/common"
	"github.com/mqasimca/nylas/internal/domain"
)

// CalendarViewMode represents the calendar display mode.

func (c *CalendarView) drawWeekView(screen tcell.Screen, x, y, width, height int) {
	titleColor := c.styles.TitleFg
	headerColor := c.styles.InfoColor
	borderColor := c.styles.BorderColor
	todayColor := c.styles.SuccessColor
	selectedBg := c.styles.TableSelectBg
	mutedColor := c.styles.BorderColor
	eventBg := c.styles.InfoColor

	// Get week start (Sunday)
	weekday := int(c.selectedDate.Weekday())
	weekStart := c.selectedDate.AddDate(0, 0, -weekday)
	today := time.Now()

	// Draw calendar name
	calName := "No Calendar"
	if cal := c.GetCurrentCalendar(); cal != nil {
		calName = cal.Name
		if len(calName) > 25 {
			calName = calName[:22] + "..."
		}
	}
	calText := fmt.Sprintf("📅 %s", calName)
	for i, ch := range calText {
		screen.SetContent(x+1+i, y, ch, nil, tcell.StyleDefault.Foreground(headerColor))
	}

	// Draw week range header
	weekRange := fmt.Sprintf("%s - %s", weekStart.Format(common.ShortDate), weekStart.AddDate(0, 0, 6).Format(common.DisplayDateFormat))
	headerText := fmt.Sprintf("◀  %s  ▶", weekRange)
	headerX := x + (width-len(headerText))/2
	for i, ch := range headerText {
		screen.SetContent(headerX+i, y, ch, nil, tcell.StyleDefault.Foreground(titleColor).Bold(true))
	}
	y += 2

	// Time column width and day column width
	timeColWidth := 7 // "10 AM "
	dayColWidth := (width - timeColWidth) / 7

	// Draw day headers
	for day := 0; day < 7; day++ {
		date := weekStart.AddDate(0, 0, day)
		colX := x + timeColWidth + day*dayColWidth

		isToday := date.Year() == today.Year() && date.YearDay() == today.YearDay()
		isSelected := date.Year() == c.selectedDate.Year() && date.YearDay() == c.selectedDate.YearDay()

		// Day name and number
		dayHeader := date.Format("Mon")
		dayNum := date.Format("2")

		headerStyle := tcell.StyleDefault.Foreground(headerColor)
		numStyle := tcell.StyleDefault.Foreground(mutedColor)
		if isToday {
			headerStyle = tcell.StyleDefault.Foreground(todayColor).Bold(true)
			numStyle = tcell.StyleDefault.Foreground(todayColor).Bold(true)
		}
		if isSelected {
			headerStyle = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(selectedBg).Bold(true)
			numStyle = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(selectedBg).Bold(true)
		}

		// Draw day name
		for j, ch := range dayHeader {
			if colX+j < x+width {
				screen.SetContent(colX+j, y, ch, nil, headerStyle)
			}
		}
		// Draw day number
		for j, ch := range dayNum {
			if colX+4+j < x+width {
				screen.SetContent(colX+4+j, y, ch, nil, numStyle)
			}
		}
	}
	y += 1

	// Draw separator line
	for i := 0; i < width; i++ {
		screen.SetContent(x+i, y, '─', nil, tcell.StyleDefault.Foreground(borderColor))
	}
	y += 1

	// Time slots from 7 AM to 9 PM (15 hours)
	startHour := 7
	endHour := 22
	availableHeight := height - 5 // Account for headers

	// Calculate rows per hour (minimum 1)
	hoursToShow := endHour - startHour
	rowsPerHour := availableHeight / hoursToShow
	if rowsPerHour < 1 {
		rowsPerHour = 1
	}

	// Build event placement map for each day
	type eventPlacement struct {
		event    *domain.Event
		startRow int
		endRow   int
	}
	dayEvents := make([][]eventPlacement, 7)

	for day := 0; day < 7; day++ {
		date := weekStart.AddDate(0, 0, day)
		dateKey := date.Format("2006-01-02")
		events := c.eventsByDate[dateKey]

		for i := range events {
			evt := &events[i]
			if evt.When.IsAllDay() {
				// All-day events at the top
				dayEvents[day] = append(dayEvents[day], eventPlacement{
					event:    evt,
					startRow: -1, // Special marker for all-day
					endRow:   -1,
				})
				continue
			}

			startTime := evt.When.StartDateTime()
			endTime := evt.When.EndDateTime()

			// Calculate row positions
			startMinutes := startTime.Hour()*60 + startTime.Minute()
			endMinutes := endTime.Hour()*60 + endTime.Minute()

			startRow := ((startMinutes - startHour*60) * rowsPerHour) / 60
			endRow := ((endMinutes - startHour*60) * rowsPerHour) / 60

			if startRow < 0 {
				startRow = 0
			}
			if endRow <= startRow {
				endRow = startRow + 1
			}

			dayEvents[day] = append(dayEvents[day], eventPlacement{
				event:    evt,
				startRow: startRow,
				endRow:   endRow,
			})
		}
	}

	// Draw time slots and grid
	for hour := startHour; hour < endHour; hour++ {
		rowY := y + (hour-startHour)*rowsPerHour

		if rowY >= y+availableHeight {
			break
		}

		// Draw time label
		timeLabel := fmt.Sprintf("%2d %s", hour%12, "AM")
		if hour == 0 || hour == 12 {
			timeLabel = fmt.Sprintf("%2d %s", 12, "AM")
		}
		if hour >= 12 {
			timeLabel = fmt.Sprintf("%2d %s", hour%12, "PM")
			if hour == 12 {
				timeLabel = fmt.Sprintf("%2d %s", 12, "PM")
			}
		}

		for i, ch := range timeLabel {
			screen.SetContent(x+i, rowY, ch, nil, tcell.StyleDefault.Foreground(mutedColor))
		}

		// Draw hour separator line
		for i := timeColWidth; i < width; i++ {
			screen.SetContent(x+i, rowY, '·', nil, tcell.StyleDefault.Foreground(borderColor))
		}
	}

	// Draw events
	for day := 0; day < 7; day++ {
		date := weekStart.AddDate(0, 0, day)
		colX := x + timeColWidth + day*dayColWidth
		isSelected := date.Year() == c.selectedDate.Year() && date.YearDay() == c.selectedDate.YearDay()

		for _, ep := range dayEvents[day] {
			if ep.startRow == -1 {
				// All-day event - show at top
				title := "▪ " + ep.event.Title
				if len(title) > dayColWidth-1 {
					title = title[:dayColWidth-2] + "…"
				}
				evtStyle := tcell.StyleDefault.Foreground(eventBg)
				for j, ch := range title {
					if colX+j < colX+dayColWidth-1 {
						screen.SetContent(colX+j, y-2, ch, nil, evtStyle)
					}
				}
				continue
			}

			// Draw timed event block
			eventY := y + ep.startRow
			eventHeight := ep.endRow - ep.startRow
			if eventHeight < 1 {
				eventHeight = 1
			}

			// Event background
			evtBgStyle := tcell.StyleDefault.Background(eventBg).Foreground(tcell.ColorBlack)
			if isSelected {
				evtBgStyle = tcell.StyleDefault.Background(selectedBg).Foreground(tcell.ColorBlack)
			}

			// Draw event block
			for row := 0; row < eventHeight && eventY+row < y+availableHeight; row++ {
				for col := 0; col < dayColWidth-1; col++ {
					screen.SetContent(colX+col, eventY+row, ' ', nil, evtBgStyle)
				}
			}

			// Draw event title
			title := ep.event.Title
			if len(title) > dayColWidth-2 {
				title = title[:dayColWidth-3] + "…"
			}

			// Draw time on first row
			timeStr := ep.event.When.StartDateTime().Format("3:04")
			for j, ch := range timeStr {
				if colX+j < colX+dayColWidth-1 && eventY < y+availableHeight {
					screen.SetContent(colX+j, eventY, ch, nil, evtBgStyle.Bold(true))
				}
			}

			// Draw title on second row if space
			if eventHeight > 1 && eventY+1 < y+availableHeight {
				for j, ch := range title {
					if colX+j < colX+dayColWidth-1 {
						screen.SetContent(colX+j, eventY+1, ch, nil, evtBgStyle)
					}
				}
			} else if eventHeight == 1 {
				// Compact: show time and truncated title
				compact := timeStr + " " + title
				if len(compact) > dayColWidth-2 {
					compact = compact[:dayColWidth-3] + "…"
				}
				for j, ch := range compact {
					if colX+j < colX+dayColWidth-1 {
						screen.SetContent(colX+j, eventY, ch, nil, evtBgStyle)
					}
				}
			}
		}

		// Draw column separator
		for row := 0; row < availableHeight; row++ {
			screen.SetContent(colX+dayColWidth-1, y+row, '│', nil, tcell.StyleDefault.Foreground(borderColor))
		}
	}
}
