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

// UpdateGrant updates the displayed grant info.
func (s *StatusIndicator) UpdateGrant(email, provider, grantID string) {
	s.config.Email = email
	s.config.Provider = provider
	s.config.GrantID = grantID
	s.render()
}

func (s *StatusIndicator) render() {
	s.Clear()

	// Check for flash message
	// Use cached Hex() method for colors
	st := s.styles
	if s.flashMsg != "" && time.Now().Before(s.flashExpiry) {
		var color string
		switch s.flashLevel {
		case FlashError:
			color = st.Hex(st.ErrorColor)
		case FlashWarn:
			color = st.Hex(st.WarnColor)
		default:
			color = st.Hex(st.InfoColor)
		}
		_, _ = fmt.Fprintf(s, "[%s::b]%s[-::-]", color, s.flashMsg)
		return
	}

	s.flashMsg = ""

	// k9s style colors - use cached Hex() method
	muted := st.Hex(st.BorderColor)
	info := st.Hex(st.InfoColor)
	section := st.Hex(st.InfoSectionFg)
	success := st.Hex(st.SuccessColor)
	highlight := st.Hex(st.HighlightColor)

	// Time
	timeStr := time.Now().Format("15:04:05")

	// Refresh countdown
	refreshStr := fmt.Sprintf("<%ds>", s.refreshSec)

	// Live indicator - k9s style
	var liveStr string
	if s.isLive {
		liveStr = fmt.Sprintf("[%s]●[-][%s::d] Live[-::-]", success, section)
	} else {
		liveStr = fmt.Sprintf("[%s]○[-][%s::d] Paused[-::-]", muted, muted)
	}

	// Account info
	email := s.config.Email
	provider := s.config.Provider
	grantID := s.config.GrantID

	// k9s style: info | section | section   time <refresh> live
	_, _ = fmt.Fprintf(s, "[%s]%s[-] [%s::d]│[-::-] [%s]%s[-] [%s::d]│[-::-] [%s::d]%s[-::-]   [%s]%s[-] [%s]%s[-] %s",
		info, email,
		muted,
		success, provider,
		muted,
		muted, grantID,
		section, timeStr,
		highlight, refreshStr,
		liveStr,
	)
}
