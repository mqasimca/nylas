package tui

import (
	"fmt"
	"strings"

	"github.com/rivo/tview"
)

// Hint represents a keyboard shortcut hint.
type Hint struct {
	Key  string
	Desc string
}

// Menu displays keyboard hints at the bottom like k9s.
type Menu struct {
	*tview.TextView
	styles *Styles
	hints  []Hint
}

// NewMenu creates a new menu component.
func NewMenu(styles *Styles) *Menu {
	m := &Menu{
		TextView: tview.NewTextView(),
		styles:   styles,
	}

	m.SetDynamicColors(true)
	m.SetBackgroundColor(styles.BgColor)
	m.SetTextAlign(tview.AlignLeft)
	m.SetBorderPadding(0, 0, 1, 0)

	return m
}

// SetHints sets the keyboard hints to display.
func (m *Menu) SetHints(hints []Hint) {
	m.hints = hints
	m.render()
}

func (m *Menu) render() {
	m.Clear()

	if len(m.hints) == 0 {
		return
	}

	keyColor := colorToHex(m.styles.MenuKeyFg)
	descColor := colorToHex(m.styles.MenuDescFg)
	muted := colorToHex(m.styles.BorderColor)

	var parts []string
	for _, h := range m.hints {
		part := fmt.Sprintf("[%s]<%s>[-][%s]%s[-]", keyColor, h.Key, descColor, h.Desc)
		parts = append(parts, part)
	}

	fmt.Fprintf(m, "%s", strings.Join(parts, fmt.Sprintf(" [%s]│[-] ", muted)))
}
