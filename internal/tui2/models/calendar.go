package models

import (
	"time"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/tui2/components"
	"github.com/mqasimca/nylas/internal/tui2/state"
	"github.com/mqasimca/nylas/internal/tui2/styles"
)

// CalendarScreenMode represents the current mode of the calendar screen.
type CalendarScreenMode int

const (
	CalendarModeView CalendarScreenMode = iota
	CalendarModeEventForm
	CalendarModeConfirmDelete
	CalendarModeAvailability
)

// CalendarScreen is the calendar view screen.
type CalendarScreen struct {
	global *state.GlobalState
	theme  *styles.Theme

	calendarGrid       *components.CalendarGrid
	spinner            spinner.Model
	calendarList       list.Model
	eventForm          *components.EventForm
	confirmDialog      *components.ConfirmDialog
	availabilityDialog *components.AvailabilityDialog

	calendars        []domain.Calendar
	events           []domain.Event
	selectedCalendar *domain.Calendar
	selectedEvent    *domain.Event // Currently selected event for edit/delete
	selectedEventIdx int           // Index of selected event in day's events

	mode            CalendarScreenMode
	loading         bool
	loadingEvents   bool
	calendarsLoaded bool
	err             error
	width           int
	height          int
	timezone        *time.Location
}

// NewCalendarScreen creates a new calendar screen.
func NewCalendarScreen(global *state.GlobalState) *CalendarScreen {
	theme := styles.GetTheme(global.Theme)

	// Create spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(theme.Primary)

	// Create calendar grid
	grid := components.NewCalendarGrid(theme)

	// Create calendar list
	delegate := list.NewDefaultDelegate()
	calList := list.New(nil, delegate, 0, 0)
	calList.SetShowTitle(false)
	calList.SetShowStatusBar(false)
	calList.SetFilteringEnabled(false)
	calList.SetShowPagination(false)
	calList.SetShowHelp(false)

	// Initialize sizes if available
	if global.WindowSize.Width > 0 && global.WindowSize.Height > 0 {
		grid.SetSize(global.WindowSize.Width-30, global.WindowSize.Height-6)
		calList.SetSize(28, global.WindowSize.Height-6)
	}

	return &CalendarScreen{
		global:       global,
		theme:        theme,
		calendarGrid: grid,
		spinner:      s,
		calendarList: calList,
		loading:      true,
		mode:         CalendarModeView,
		timezone:     time.Local,
	}
}

// SetTimezone sets the timezone for the calendar screen.
func (m *CalendarScreen) SetTimezone(tz *time.Location) {
	if tz != nil {
		m.timezone = tz
		m.calendarGrid.SetTimezone(tz)
	}
}

// Init implements tea.Model.
func (m *CalendarScreen) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.fetchCalendars(),
	)
}

// Update implements tea.Model.
