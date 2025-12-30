// Package components provides reusable Bubble Tea components.
package components

import (
	"slices"
	"time"

	tea "charm.land/bubbletea/v2"
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
