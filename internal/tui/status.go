package tui

import (
	"fmt"
	"time"

	"github.com/rivo/tview"
)

// StatusIndicator shows status info like k9s.
type StatusIndicator struct {
	*tview.TextView
	styles      *Styles
	config      Config
	flashMsg    string
	flashLevel  FlashLevel
	flashExpiry time.Time
	refreshSec  int
	isLive      bool
}

// NewStatusIndicator creates a new status indicator.
func NewStatusIndicator(styles *Styles, config Config) *StatusIndicator {
	s := &StatusIndicator{
		TextView:   tview.NewTextView(),
		styles:     styles,
		config:     config,
		refreshSec: int(config.RefreshInterval.Seconds()),
		isLive:     true,
	}

	s.SetDynamicColors(true)
	s.SetBackgroundColor(styles.BgColor)
	s.SetTextAlign(tview.AlignRight)
	s.SetBorderPadding(0, 0, 0, 1)

	s.render()
	return s
}

// Update refreshes the status display.
func (s *StatusIndicator) Update() {
	s.refreshSec--
	if s.refreshSec <= 0 {
		s.refreshSec = int(s.config.RefreshInterval.Seconds())
	}
	s.render()
}

// Flash shows a temporary message.
func (s *StatusIndicator) Flash(level FlashLevel, msg string) {
	s.flashMsg = msg
	s.flashLevel = level
	s.flashExpiry = time.Now().Add(3 * time.Second)
	s.render()
}

// ToggleLive toggles the live refresh status.
func (s *StatusIndicator) ToggleLive() {
	s.isLive = !s.isLive
	s.render()
}

func (s *StatusIndicator) render() {
	s.Clear()

	// Check for flash message
	if s.flashMsg != "" && time.Now().Before(s.flashExpiry) {
		var color string
		switch s.flashLevel {
		case FlashError:
			color = colorToHex(s.styles.ErrorColor)
		case FlashWarn:
			color = colorToHex(s.styles.WarnColor)
		default:
			color = colorToHex(s.styles.InfoColor)
		}
		fmt.Fprintf(s, "[%s]%s[-]", color, s.flashMsg)
		return
	}

	s.flashMsg = ""

	// Time
	timeStr := time.Now().Format("3:04:05 PM")
	muted := colorToHex(s.styles.BorderColor)
	info := colorToHex(s.styles.InfoColor)
	success := colorToHex(s.styles.SuccessColor)
	warn := colorToHex(s.styles.WarnColor)

	// Refresh countdown
	refreshStr := fmt.Sprintf("<%ds>", s.refreshSec)

	// Live indicator
	var liveStr string
	if s.isLive {
		liveStr = fmt.Sprintf("[%s]●[-] [%s]Live[-]", success, muted)
	} else {
		liveStr = fmt.Sprintf("[%s]○[-] [%s]Paused[-]", muted, muted)
	}

	// Account info
	email := s.config.Email
	provider := s.config.Provider
	grantID := s.config.GrantID
	if len(grantID) > 12 {
		grantID = grantID[:12] + "..."
	}

	// Format: email | provider | grant   time <refresh> live  ?:help q:quit
	fmt.Fprintf(s, "[%s]%s[-] [%s]│[-] [%s]%s[-] [%s]│[-] [%s]%s[-]   %s [%s]%s[-] %s  [%s]?[-][::]help [%s]q[-][::]quit",
		info, email,
		muted,
		success, provider,
		muted,
		muted, grantID,
		timeStr,
		warn, refreshStr,
		liveStr,
		warn,
		warn,
	)
}
