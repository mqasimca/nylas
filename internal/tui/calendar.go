package tui

import (
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

// getCalendarColor returns the color for the current calendar.
func (c *CalendarView) getCalendarColor() tcell.Color {
	cal := c.GetCurrentCalendar()
	if cal == nil || cal.HexColor == "" {
		return tcell.ColorDefault
	}
	return parseHexColor(cal.HexColor)
}

// parseHexColor parses a hex color string and returns a tcell.Color.
func parseHexColor(hex string) tcell.Color {
	if hex == "" {
		return tcell.ColorDefault
	}

	// Remove # prefix if present
	if len(hex) > 0 && hex[0] == '#' {
		hex = hex[1:]
	}

	// Parse 6-character hex color
	if len(hex) != 6 {
		return tcell.ColorDefault
	}

	r := parseHexDigits(hex[0:2])
	g := parseHexDigits(hex[2:4])
	b := parseHexDigits(hex[4:6])

	// #nosec G115 -- r, g, b are from parseHexDigits which returns 0-255, no overflow possible
	return tcell.NewRGBColor(int32(r), int32(g), int32(b))
}

// parseHexDigits converts a 2-character hex string to int.
func parseHexDigits(hex string) int {
	if len(hex) != 2 {
		return 0
	}
	high := hexCharToInt(hex[0])
	low := hexCharToInt(hex[1])
	return high*16 + low
}

// hexCharToInt converts a single hex digit to int.
func hexCharToInt(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'f':
		return int(c - 'a' + 10)
	case c >= 'A' && c <= 'F':
		return int(c - 'A' + 10)
	default:
		return 0
	}
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
