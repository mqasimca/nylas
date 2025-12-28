// Package components provides reusable Bubble Tea components.
package components

import (
	"fmt"
	"slices"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/tui2/styles"
)

// CalendarViewMode represents the calendar display mode.
type CalendarViewMode int

const (
	// CalendarMonthView shows the full month.
	CalendarMonthView CalendarViewMode = iota
	// CalendarWeekView shows a single week.
	CalendarWeekView
	// CalendarAgendaView shows a list of upcoming events.
	CalendarAgendaView
)

// CalendarGrid is a component that displays a month calendar grid.
type CalendarGrid struct {
	theme        *styles.Theme
	events       []domain.Event
	eventsByDate map[string][]domain.Event
	currentMonth time.Time
	selectedDate time.Time
	width        int
	height       int
	viewMode     CalendarViewMode
	timezone     *time.Location
	workingHours *WorkingHours
	showWeekends bool
	firstDayMon  bool // If true, week starts on Monday; otherwise Sunday
}

// WorkingHours represents working hours configuration.
type WorkingHours struct {
	StartHour int // 0-23
	EndHour   int // 0-23
	Days      []time.Weekday
}

// DefaultWorkingHours returns default working hours (9-17, Mon-Fri).
func DefaultWorkingHours() *WorkingHours {
	return &WorkingHours{
		StartHour: 9,
		EndHour:   17,
		Days:      []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday},
	}
}

// IsWorkingDay returns true if the given weekday is a working day.
func (w *WorkingHours) IsWorkingDay(day time.Weekday) bool {
	for _, d := range w.Days {
		if d == day {
			return true
		}
	}
	return false
}

// IsWorkingHour returns true if the given hour is within working hours.
func (w *WorkingHours) IsWorkingHour(hour int) bool {
	return hour >= w.StartHour && hour < w.EndHour
}

// NewCalendarGrid creates a new calendar grid component.
func NewCalendarGrid(theme *styles.Theme) *CalendarGrid {
	now := time.Now()
	return &CalendarGrid{
		theme:        theme,
		eventsByDate: make(map[string][]domain.Event),
		currentMonth: time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()),
		selectedDate: now,
		viewMode:     CalendarMonthView, // Default to month view (Google Calendar-style)
		timezone:     time.Local,
		workingHours: DefaultWorkingHours(),
		showWeekends: true,
		firstDayMon:  true, // ISO week (Monday first)
	}
}

// SetEvents sets the events to display.
func (c *CalendarGrid) SetEvents(events []domain.Event) {
	c.events = events
	c.eventsByDate = make(map[string][]domain.Event)

	for i := range events {
		evt := &events[i]
		var key string

		// Handle all-day events specially - use the date string directly
		// to avoid timezone conversion issues
		if evt.When.IsAllDay() && evt.When.Date != "" {
			key = evt.When.Date
		} else if evt.When.IsAllDay() && evt.When.StartDate != "" {
			key = evt.When.StartDate
		} else {
			startDate := evt.When.StartDateTime()
			if startDate.IsZero() {
				continue
			}
			// Convert to grid's timezone for timed events
			if c.timezone != nil {
				startDate = startDate.In(c.timezone)
			}
			key = startDate.Format("2006-01-02")
		}

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

// GetEventsForDate returns events for a specific date.
func (c *CalendarGrid) GetEventsForDate(date time.Time) []domain.Event {
	key := date.Format("2006-01-02")
	return c.eventsByDate[key]
}

// GetEventCount returns the total number of events.
func (c *CalendarGrid) GetEventCount() int {
	return len(c.events)
}

// GetSelectedDate returns the currently selected date.
func (c *CalendarGrid) GetSelectedDate() time.Time {
	return c.selectedDate
}

// SetSelectedDate sets the selected date.
func (c *CalendarGrid) SetSelectedDate(date time.Time) {
	c.selectedDate = date
	// Also update current month to show the selected date
	c.currentMonth = time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
}

// GetCurrentMonth returns the currently displayed month.
func (c *CalendarGrid) GetCurrentMonth() time.Time {
	return c.currentMonth
}

// SetCurrentMonth sets the current month to display.
func (c *CalendarGrid) SetCurrentMonth(month time.Time) {
	c.currentMonth = time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, month.Location())
}

// SetTimezone sets the timezone for displaying events.
func (c *CalendarGrid) SetTimezone(tz *time.Location) {
	c.timezone = tz
	// Re-index events with new timezone
	c.SetEvents(c.events)
}

// GetTimezone returns the current timezone.
func (c *CalendarGrid) GetTimezone() *time.Location {
	return c.timezone
}

// SetWorkingHours sets the working hours configuration.
func (c *CalendarGrid) SetWorkingHours(wh *WorkingHours) {
	c.workingHours = wh
}

// GetWorkingHours returns the working hours configuration.
func (c *CalendarGrid) GetWorkingHours() *WorkingHours {
	return c.workingHours
}

// SetShowWeekends sets whether to show weekends.
func (c *CalendarGrid) SetShowWeekends(show bool) {
	c.showWeekends = show
}

// SetFirstDayMonday sets whether the week starts on Monday.
func (c *CalendarGrid) SetFirstDayMonday(monday bool) {
	c.firstDayMon = monday
}

// SetSize sets the width and height of the calendar grid.
func (c *CalendarGrid) SetSize(width, height int) {
	c.width = width
	c.height = height
}

// SetViewMode sets the calendar view mode.
func (c *CalendarGrid) SetViewMode(mode CalendarViewMode) {
	c.viewMode = mode
}

// GetViewMode returns the current view mode.
func (c *CalendarGrid) GetViewMode() CalendarViewMode {
	return c.viewMode
}

// NextMonth moves to the next month.
func (c *CalendarGrid) NextMonth() {
	c.currentMonth = c.currentMonth.AddDate(0, 1, 0)
}

// PrevMonth moves to the previous month.
func (c *CalendarGrid) PrevMonth() {
	c.currentMonth = c.currentMonth.AddDate(0, -1, 0)
}

// NextDay selects the next day.
func (c *CalendarGrid) NextDay() {
	c.selectedDate = c.selectedDate.AddDate(0, 0, 1)
	// Update month if needed
	if c.selectedDate.Month() != c.currentMonth.Month() || c.selectedDate.Year() != c.currentMonth.Year() {
		c.currentMonth = time.Date(c.selectedDate.Year(), c.selectedDate.Month(), 1, 0, 0, 0, 0, c.selectedDate.Location())
	}
}

// PrevDay selects the previous day.
func (c *CalendarGrid) PrevDay() {
	c.selectedDate = c.selectedDate.AddDate(0, 0, -1)
	// Update month if needed
	if c.selectedDate.Month() != c.currentMonth.Month() || c.selectedDate.Year() != c.currentMonth.Year() {
		c.currentMonth = time.Date(c.selectedDate.Year(), c.selectedDate.Month(), 1, 0, 0, 0, 0, c.selectedDate.Location())
	}
}

// NextWeek moves selection to next week.
func (c *CalendarGrid) NextWeek() {
	c.selectedDate = c.selectedDate.AddDate(0, 0, 7)
	if c.selectedDate.Month() != c.currentMonth.Month() || c.selectedDate.Year() != c.currentMonth.Year() {
		c.currentMonth = time.Date(c.selectedDate.Year(), c.selectedDate.Month(), 1, 0, 0, 0, 0, c.selectedDate.Location())
	}
}

// PrevWeek moves selection to previous week.
func (c *CalendarGrid) PrevWeek() {
	c.selectedDate = c.selectedDate.AddDate(0, 0, -7)
	if c.selectedDate.Month() != c.currentMonth.Month() || c.selectedDate.Year() != c.currentMonth.Year() {
		c.currentMonth = time.Date(c.selectedDate.Year(), c.selectedDate.Month(), 1, 0, 0, 0, 0, c.selectedDate.Location())
	}
}

// GoToToday jumps to today's date.
func (c *CalendarGrid) GoToToday() {
	now := time.Now()
	c.selectedDate = now
	c.currentMonth = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
}

// IsToday returns true if the given date is today.
func (c *CalendarGrid) IsToday(date time.Time) bool {
	now := time.Now()
	return date.Year() == now.Year() && date.Month() == now.Month() && date.Day() == now.Day()
}

// IsSelected returns true if the given date is the selected date.
func (c *CalendarGrid) IsSelected(date time.Time) bool {
	return date.Year() == c.selectedDate.Year() &&
		date.Month() == c.selectedDate.Month() &&
		date.Day() == c.selectedDate.Day()
}

// IsCurrentMonth returns true if the given date is in the current displayed month.
func (c *CalendarGrid) IsCurrentMonth(date time.Time) bool {
	return date.Year() == c.currentMonth.Year() && date.Month() == c.currentMonth.Month()
}

// HasEvents returns true if the given date has events.
func (c *CalendarGrid) HasEvents(date time.Time) bool {
	key := date.Format("2006-01-02")
	return len(c.eventsByDate[key]) > 0
}

// EventCountForDate returns the number of events for a given date.
func (c *CalendarGrid) EventCountForDate(date time.Time) int {
	key := date.Format("2006-01-02")
	return len(c.eventsByDate[key])
}

// Update handles messages for the calendar grid.
func (c *CalendarGrid) Update(msg tea.Msg) (*CalendarGrid, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case msg.Code == tea.KeyLeft || msg.Text == "h":
			c.PrevDay()
		case msg.Code == tea.KeyRight || msg.Text == "l":
			c.NextDay()
		case msg.Code == tea.KeyUp || msg.Text == "k":
			c.PrevWeek()
		case msg.Code == tea.KeyDown || msg.Text == "j":
			c.NextWeek()
		case msg.Text == "[":
			c.PrevMonth()
		case msg.Text == "]":
			c.NextMonth()
		case msg.Text == "t":
			c.GoToToday()
		}
	}
	return c, nil
}

// View renders the calendar grid.
func (c *CalendarGrid) View() string {
	switch c.viewMode {
	case CalendarWeekView:
		return c.renderWeekView()
	case CalendarAgendaView:
		return c.renderAgendaView()
	default:
		return c.renderMonthView()
	}
}

// renderMonthView renders the month grid with clean cells like Google Calendar.
func (c *CalendarGrid) renderMonthView() string {
	var b strings.Builder

	// Calculate dimensions
	numCols := 7
	if !c.showWeekends {
		numCols = 5
	}

	// Cell dimensions - use full available width
	cellWidth := c.width / numCols
	if cellWidth < 10 {
		cellWidth = 10
	}

	// Calculate cell height based on available space
	// Reserve: header (2), day names (1), bottom margin (1)
	reservedLines := 4
	availableHeight := c.height - reservedLines
	days := c.getMonthDays()
	numWeeks := (len(days) + numCols - 1) / numCols
	if numWeeks == 0 {
		numWeeks = 1
	}
	cellHeight := availableHeight / numWeeks
	if cellHeight < 4 {
		cellHeight = 4
	}
	if cellHeight > 7 {
		cellHeight = 7
	}

	// Header with month/year and navigation hints
	header := fmt.Sprintf("← %s →", c.currentMonth.Format("January 2006"))
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(c.theme.Primary).
		Width(c.width).
		Align(lipgloss.Center)
	b.WriteString(headerStyle.Render(header))
	b.WriteString("\n\n")

	// Day names header
	dayNames := c.getDayNamesShort()
	dayHeaderStyle := lipgloss.NewStyle().
		Foreground(c.theme.Secondary).
		Width(cellWidth).
		Align(lipgloss.Center).
		Bold(true)

	for _, name := range dayNames {
		b.WriteString(dayHeaderStyle.Render(name))
	}
	b.WriteString("\n")

	// Generate calendar grid
	for week := 0; week < numWeeks; week++ {
		// Build each line of the week row
		weekCells := make([]string, numCols)
		for day := 0; day < numCols; day++ {
			idx := week*numCols + day
			if idx < len(days) {
				weekCells[day] = c.renderCleanCell(days[idx], cellWidth, cellHeight)
			} else {
				weekCells[day] = c.renderEmptyCellClean(cellWidth, cellHeight)
			}
		}
		// Join cells horizontally
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, weekCells...))
		b.WriteString("\n")
	}

	return b.String()
}

// getDayNamesShort returns short day names for the header.
func (c *CalendarGrid) getDayNamesShort() []string {
	if c.firstDayMon {
		if c.showWeekends {
			return []string{"MON", "TUE", "WED", "THU", "FRI", "SAT", "SUN"}
		}
		return []string{"MON", "TUE", "WED", "THU", "FRI"}
	}
	if c.showWeekends {
		return []string{"SUN", "MON", "TUE", "WED", "THU", "FRI", "SAT"}
	}
	return []string{"MON", "TUE", "WED", "THU", "FRI"}
}

// renderEmptyCellClean renders an empty cell without content.
func (c *CalendarGrid) renderEmptyCellClean(width, height int) string {
	cellStyle := lipgloss.NewStyle().
		Width(width).
		Height(height).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(c.theme.Dimmed.GetForeground())

	return cellStyle.Render("")
}

// renderCleanCell renders a single day cell with clean formatting.
func (c *CalendarGrid) renderCleanCell(date time.Time, width, height int) string {
	isSelected := c.IsSelected(date)
	isToday := c.IsToday(date)
	isCurrentMonth := c.IsCurrentMonth(date)
	events := c.GetEventsForDate(date)

	// Cell style with border
	cellStyle := lipgloss.NewStyle().
		Width(width).
		Height(height).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(c.theme.Dimmed.GetForeground()).
		Padding(0, 1)

	// Selected cell gets highlighted background
	if isSelected {
		cellStyle = cellStyle.
			Background(c.theme.Primary).
			BorderForeground(c.theme.Primary)
	}

	var content strings.Builder
	contentWidth := width - 4 // Account for border and padding

	// Line 1: Day number
	dayNum := fmt.Sprintf("%d", date.Day())
	dayStyle := lipgloss.NewStyle()

	if isToday {
		dayStyle = dayStyle.Bold(true).Foreground(c.theme.Success)
		if isSelected {
			dayStyle = dayStyle.Background(c.theme.Success).Foreground(lipgloss.Color("#000000"))
		}
	} else if !isCurrentMonth {
		dayStyle = dayStyle.Foreground(c.theme.Dimmed.GetForeground())
	} else if isSelected {
		dayStyle = dayStyle.Foreground(lipgloss.Color("#000000")).Bold(true)
	}

	content.WriteString(dayStyle.Render(dayNum))
	content.WriteString("\n")

	// Line 2: Event dots
	if len(events) > 0 {
		dots := c.renderEventDotsClean(events, contentWidth, isSelected)
		content.WriteString(dots)
		content.WriteString("\n")

		// Line 3+: Event titles (up to 2)
		maxTitles := height - 3
		if maxTitles > 2 {
			maxTitles = 2
		}
		for i := 0; i < maxTitles && i < len(events); i++ {
			title := events[i].Title
			if title == "" {
				title = "(No title)"
			}
			// Truncate title to fit
			if len(title) > contentWidth-1 {
				title = title[:contentWidth-2] + "…"
			}

			titleStyle := lipgloss.NewStyle()
			if isSelected {
				titleStyle = titleStyle.Foreground(lipgloss.Color("#000000"))
			} else {
				titleStyle = titleStyle.Foreground(c.theme.Secondary)
			}
			content.WriteString(titleStyle.Render(title))
			if i < maxTitles-1 && i < len(events)-1 {
				content.WriteString("\n")
			}
		}
	}

	return cellStyle.Render(content.String())
}

// renderEventDotsClean renders colored dots for events.
func (c *CalendarGrid) renderEventDotsClean(events []domain.Event, maxWidth int, isSelected bool) string {
	var dots strings.Builder

	maxDots := min(len(events), maxWidth/2)
	if maxDots > 5 {
		maxDots = 5
	}

	for i := 0; i < maxDots; i++ {
		evt := events[i]
		dotColor := c.theme.Primary
		if evt.Status == "cancelled" {
			dotColor = c.theme.Error
		} else if !evt.Busy {
			dotColor = c.theme.Success
		}

		dotStyle := lipgloss.NewStyle().Foreground(dotColor)
		if isSelected {
			dotStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
		}
		dots.WriteString(dotStyle.Render("●"))
		if i < maxDots-1 {
			dots.WriteString(" ")
		}
	}

	// Show "+N" if more events
	if len(events) > maxDots {
		moreStyle := lipgloss.NewStyle().Foreground(c.theme.Secondary)
		if isSelected {
			moreStyle = moreStyle.Foreground(lipgloss.Color("#000000"))
		}
		dots.WriteString(moreStyle.Render(fmt.Sprintf("+%d", len(events)-maxDots)))
	}

	return dots.String()
}

// renderWeekView renders a Google Calendar-style week view with time slots.
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
