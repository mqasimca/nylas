package tui

import (
	"fmt"
	"slices"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/rivo/tview"
)

// CalendarViewMode represents the calendar display mode.
type CalendarViewMode int

const (
	CalendarMonthView CalendarViewMode = iota
	CalendarWeekView
	CalendarAgendaView
)

// CalendarView displays a Google Calendar-style calendar.
type CalendarView struct {
	*tview.Box
	app              *App
	styles           *Styles
	events           []domain.Event
	eventsByDate     map[string][]domain.Event
	calendars        []domain.Calendar
	calendarID       string
	calendarIndex    int
	currentMonth     time.Time
	selectedDate     time.Time
	viewMode         CalendarViewMode
	cellWidth        int
	cellHeight       int
	headerHeight     int
	onDateSelect     func(time.Time)
	onEventSelect    func(*domain.Event)
	onCalendarChange func(string) // Callback when calendar changes
}

// NewCalendarView creates a new calendar view.
func NewCalendarView(app *App) *CalendarView {
	now := time.Now()
	c := &CalendarView{
		Box:          tview.NewBox(),
		app:          app,
		styles:       app.styles,
		eventsByDate: make(map[string][]domain.Event),
		currentMonth: time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()),
		selectedDate: now,
		viewMode:     CalendarMonthView,
		cellWidth:    14,
		cellHeight:   4,
		headerHeight: 3,
	}

	c.SetBackgroundColor(app.styles.BgColor)
	return c
}

// SetEvents sets the events to display.
func (c *CalendarView) SetEvents(events []domain.Event) {
	c.events = events
	c.eventsByDate = make(map[string][]domain.Event)

	for i := range events {
		evt := &events[i]
		startDate := evt.When.StartDateTime()
		if startDate.IsZero() {
			continue
		}
		key := startDate.Format("2006-01-02")
		c.eventsByDate[key] = append(c.eventsByDate[key], *evt)
	}

	// Sort events by time within each day
	for key := range c.eventsByDate {
		slices.SortFunc(c.eventsByDate[key], func(a, b domain.Event) int {
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
	}
}

// SetCalendars sets the available calendars.
func (c *CalendarView) SetCalendars(calendars []domain.Calendar) {
	c.calendars = calendars
	if len(calendars) > 0 {
		// Find primary calendar or use first
		for i, cal := range calendars {
			if cal.IsPrimary {
				c.calendarIndex = i
				c.calendarID = cal.ID
				return
			}
		}
		c.calendarIndex = 0
		c.calendarID = calendars[0].ID
	}
}

// GetCalendars returns all available calendars.
func (c *CalendarView) GetCalendars() []domain.Calendar {
	return c.calendars
}

// GetCurrentCalendar returns the currently selected calendar.
func (c *CalendarView) GetCurrentCalendar() *domain.Calendar {
	if c.calendarIndex >= 0 && c.calendarIndex < len(c.calendars) {
		return &c.calendars[c.calendarIndex]
	}
	return nil
}

// GetCurrentCalendarID returns the current calendar ID.
func (c *CalendarView) GetCurrentCalendarID() string {
	return c.calendarID
}

// NextCalendar switches to the next calendar.
func (c *CalendarView) NextCalendar() {
	if len(c.calendars) == 0 {
		return
	}
	c.calendarIndex = (c.calendarIndex + 1) % len(c.calendars)
	c.calendarID = c.calendars[c.calendarIndex].ID
	if c.onCalendarChange != nil {
		c.onCalendarChange(c.calendarID)
	}
}

// PrevCalendar switches to the previous calendar.
func (c *CalendarView) PrevCalendar() {
	if len(c.calendars) == 0 {
		return
	}
	c.calendarIndex--
	if c.calendarIndex < 0 {
		c.calendarIndex = len(c.calendars) - 1
	}
	c.calendarID = c.calendars[c.calendarIndex].ID
	if c.onCalendarChange != nil {
		c.onCalendarChange(c.calendarID)
	}
}

// SetCalendarByIndex sets the current calendar by index.
func (c *CalendarView) SetCalendarByIndex(index int) {
	if index >= 0 && index < len(c.calendars) {
		c.calendarIndex = index
		c.calendarID = c.calendars[index].ID
		if c.onCalendarChange != nil {
			c.onCalendarChange(c.calendarID)
		}
	}
}

// SetOnCalendarChange sets the callback for when calendar changes.
func (c *CalendarView) SetOnCalendarChange(handler func(string)) {
	c.onCalendarChange = handler
}

// SetOnDateSelect sets the callback for date selection.
func (c *CalendarView) SetOnDateSelect(handler func(time.Time)) {
	c.onDateSelect = handler
}

// SetOnEventSelect sets the callback for event selection.
func (c *CalendarView) SetOnEventSelect(handler func(*domain.Event)) {
	c.onEventSelect = handler
}

// NextMonth moves to the next month.
func (c *CalendarView) NextMonth() {
	c.currentMonth = c.currentMonth.AddDate(0, 1, 0)
}

// PrevMonth moves to the previous month.
func (c *CalendarView) PrevMonth() {
	c.currentMonth = c.currentMonth.AddDate(0, -1, 0)
}

// NextWeek moves to the next week.
func (c *CalendarView) NextWeek() {
	c.selectedDate = c.selectedDate.AddDate(0, 0, 7)
	c.updateCurrentMonth()
}

// PrevWeek moves to the previous week.
func (c *CalendarView) PrevWeek() {
	c.selectedDate = c.selectedDate.AddDate(0, 0, -7)
	c.updateCurrentMonth()
}

// GoToToday navigates to today.
func (c *CalendarView) GoToToday() {
	now := time.Now()
	c.selectedDate = now
	c.currentMonth = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
}

// ToggleViewMode cycles through view modes.
func (c *CalendarView) ToggleViewMode() {
	c.viewMode = (c.viewMode + 1) % 3
}

// SetViewMode sets the view mode.
func (c *CalendarView) SetViewMode(mode CalendarViewMode) {
	c.viewMode = mode
}

// GetSelectedDate returns the currently selected date.
func (c *CalendarView) GetSelectedDate() time.Time {
	return c.selectedDate
}

// GetEventsForDate returns events for a specific date.
func (c *CalendarView) GetEventsForDate(date time.Time) []domain.Event {
	key := date.Format("2006-01-02")
	return c.eventsByDate[key]
}

func (c *CalendarView) updateCurrentMonth() {
	if c.selectedDate.Month() != c.currentMonth.Month() || c.selectedDate.Year() != c.currentMonth.Year() {
		c.currentMonth = time.Date(c.selectedDate.Year(), c.selectedDate.Month(), 1, 0, 0, 0, 0, c.selectedDate.Location())
	}
}

// Draw renders the calendar.
func (c *CalendarView) Draw(screen tcell.Screen) {
	c.Box.DrawForSubclass(screen, c)
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

				// Draw event dot and title
				screen.SetContent(cellX+1, cellY+i+1, '•', nil, tcell.StyleDefault.Foreground(c.styles.InfoColor))
				for j, ch := range title {
					if cellX+2+j < cellX+c.cellWidth-1 {
						screen.SetContent(cellX+2+j, cellY+i+1, ch, nil, evtStyle)
					}
				}
			}
		}
	}
}

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
	weekRange := fmt.Sprintf("%s - %s", weekStart.Format("Jan 2"), weekStart.AddDate(0, 0, 6).Format("Jan 2, 2006"))
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
func (c *CalendarView) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return c.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		switch event.Key() {
		case tcell.KeyLeft:
			c.selectedDate = c.selectedDate.AddDate(0, 0, -1)
			c.updateCurrentMonth()
		case tcell.KeyRight:
			c.selectedDate = c.selectedDate.AddDate(0, 0, 1)
			c.updateCurrentMonth()
		case tcell.KeyUp:
			c.selectedDate = c.selectedDate.AddDate(0, 0, -7)
			c.updateCurrentMonth()
		case tcell.KeyDown:
			c.selectedDate = c.selectedDate.AddDate(0, 0, 7)
			c.updateCurrentMonth()
		case tcell.KeyPgUp:
			c.PrevMonth()
		case tcell.KeyPgDn:
			c.NextMonth()
		case tcell.KeyEnter:
			if c.onDateSelect != nil {
				c.onDateSelect(c.selectedDate)
			}
		case tcell.KeyRune:
			switch event.Rune() {
			case 'h':
				c.selectedDate = c.selectedDate.AddDate(0, 0, -1)
				c.updateCurrentMonth()
			case 'l':
				c.selectedDate = c.selectedDate.AddDate(0, 0, 1)
				c.updateCurrentMonth()
			case 'k':
				c.selectedDate = c.selectedDate.AddDate(0, 0, -7)
				c.updateCurrentMonth()
			case 'j':
				c.selectedDate = c.selectedDate.AddDate(0, 0, 7)
				c.updateCurrentMonth()
			case 'H':
				c.PrevMonth()
			case 'L':
				c.NextMonth()
			case 't':
				c.GoToToday()
			case 'm':
				c.SetViewMode(CalendarMonthView)
			case 'w':
				c.SetViewMode(CalendarWeekView)
			case 'a':
				c.SetViewMode(CalendarAgendaView)
			case 'c':
				c.NextCalendar()
			case 'C':
				c.PrevCalendar()
			}
		}
	})
}

// MouseHandler returns the mouse handler for the calendar.
func (c *CalendarView) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return c.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
		if !c.InRect(event.Position()) {
			return false, nil
		}

		switch action {
		case tview.MouseLeftClick:
			setFocus(c)
			x, y, _, _ := c.GetInnerRect()
			mouseX, mouseY := event.Position()

			if c.viewMode == CalendarMonthView {
				// Calculate which day was clicked
				headerOffset := 4 // Header + day names + separator
				if mouseY < y+headerOffset {
					// Check if clicked on navigation
					return true, nil
				}

				// Calculate clicked cell
				col := (mouseX - x) / c.cellWidth
				row := (mouseY - y - headerOffset) / c.cellHeight

				if col >= 0 && col < 7 && row >= 0 && row < 6 {
					// Calculate the date
					firstDay := c.currentMonth
					firstWeekday := int(firstDay.Weekday())
					startDate := firstDay.AddDate(0, 0, -firstWeekday)
					clickedDate := startDate.AddDate(0, 0, row*7+col)

					c.selectedDate = clickedDate
					c.updateCurrentMonth()

					if c.onDateSelect != nil {
						c.onDateSelect(c.selectedDate)
					}
				}
			}
			return true, nil

		case tview.MouseLeftDoubleClick:
			if c.onDateSelect != nil {
				c.onDateSelect(c.selectedDate)
			}
			return true, nil

		case tview.MouseScrollUp:
			c.PrevMonth()
			return true, nil

		case tview.MouseScrollDown:
			c.NextMonth()
			return true, nil
		}

		return false, nil
	})
}

// Focus is called when this primitive receives focus.
func (c *CalendarView) Focus(delegate func(p tview.Primitive)) {
	c.Box.Focus(delegate)
}

// HasFocus returns whether or not this primitive has focus.
func (c *CalendarView) HasFocus() bool {
	return c.Box.HasFocus()
}
