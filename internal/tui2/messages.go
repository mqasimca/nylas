// Package tui2 provides a Bubble Tea-based terminal user interface for Nylas.
package tui2

import tea "github.com/charmbracelet/bubbletea"

// ScreenType represents different screens in the application.
type ScreenType int

const (
	// ScreenDashboard is the main dashboard view.
	ScreenDashboard ScreenType = iota
	// ScreenMessages is the email list view.
	ScreenMessages
	// ScreenMessageView is the email detail view.
	ScreenMessageView
	// ScreenCompose is the compose email view.
	ScreenCompose
	// ScreenCalendar is the calendar view.
	ScreenCalendar
	// ScreenEventForm is the event creation/edit view.
	ScreenEventForm
	// ScreenContacts is the contacts view.
	ScreenContacts
	// ScreenSettings is the settings view.
	ScreenSettings
	// ScreenHelp is the help view.
	ScreenHelp
)

// String returns the string representation of the screen type.
func (s ScreenType) String() string {
	switch s {
	case ScreenDashboard:
		return "Dashboard"
	case ScreenMessages:
		return "Messages"
	case ScreenMessageView:
		return "Message"
	case ScreenCompose:
		return "Compose"
	case ScreenCalendar:
		return "Calendar"
	case ScreenEventForm:
		return "Event"
	case ScreenContacts:
		return "Contacts"
	case ScreenSettings:
		return "Settings"
	case ScreenHelp:
		return "Help"
	default:
		return "Unknown"
	}
}

// NavigateMsg is sent to navigate to a new screen.
type NavigateMsg struct {
	Screen ScreenType
	Data   interface{}
}

// BackMsg is sent to go back to the previous screen.
type BackMsg struct{}

// QuitMsg is sent to quit the application.
type QuitMsg struct{}

// ErrMsg represents an error message.
type ErrMsg struct {
	Err error
}

// Error returns the error string.
func (e ErrMsg) Error() string {
	return e.Err.Error()
}

// StatusMsg represents a status update message.
type StatusMsg struct {
	Message string
	Level   StatusLevel
}

// StatusLevel represents the severity of a status message.
type StatusLevel int

const (
	// StatusInfo is an informational message.
	StatusInfo StatusLevel = iota
	// StatusSuccess is a success message.
	StatusSuccess
	// StatusWarning is a warning message.
	StatusWarning
	// StatusError is an error message.
	StatusError
)

// Navigation helpers

// NavigateCmd creates a command to navigate to a screen.
func NavigateCmd(screen ScreenType, data interface{}) tea.Cmd {
	return func() tea.Msg {
		return NavigateMsg{Screen: screen, Data: data}
	}
}

// BackCmd creates a command to go back.
func BackCmd() tea.Cmd {
	return func() tea.Msg {
		return BackMsg{}
	}
}

// QuitCmd creates a command to quit.
func QuitCmd() tea.Cmd {
	return func() tea.Msg {
		return QuitMsg{}
	}
}

// StatusCmd creates a command to show a status message.
func StatusCmd(message string, level StatusLevel) tea.Cmd {
	return func() tea.Msg {
		return StatusMsg{Message: message, Level: level}
	}
}
