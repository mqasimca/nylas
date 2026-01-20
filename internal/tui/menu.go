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

	keyColor := m.styles.Hex(m.styles.MenuKeyFg)
	numKeyColor := m.styles.Hex(m.styles.MenuNumKeyFg)
	descColor := m.styles.Hex(m.styles.MenuDescFg)
	muted := m.styles.Hex(m.styles.BorderColor)

	var parts []string
	for _, h := range m.hints {
		// k9s style: numeric keys get fuchsia color, letter keys get dodgerblue
		kc := keyColor
		if len(h.Key) == 1 && h.Key[0] >= '0' && h.Key[0] <= '9' {
			kc = numKeyColor
		}
		part := fmt.Sprintf("[%s::d]<%s>[-::-][%s::d]%s[-::-]", kc, h.Key, descColor, h.Desc)
		parts = append(parts, part)
	}

	_, _ = fmt.Fprintf(m, "%s", strings.Join(parts, fmt.Sprintf(" [%s::d]│[-::-] ", muted)))
}
