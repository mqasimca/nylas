package tui

import (
	"fmt"

	"github.com/rivo/tview"
)

// Logo displays the application logo.
type Logo struct {
	*tview.TextView
	styles *Styles
}

// NewLogo creates a new logo component.
func NewLogo(styles *Styles) *Logo {
	l := &Logo{
		TextView: tview.NewTextView(),
		styles:   styles,
	}

	l.SetDynamicColors(true)
	l.SetBackgroundColor(styles.BgColor)
	l.SetTextAlign(tview.AlignLeft)
	l.SetBorderPadding(0, 0, 1, 0)

	l.render()
	return l
}

func (l *Logo) render() {
	l.Clear()
	color := l.styles.Hex(l.styles.LogoColor)
	_, _ = fmt.Fprintf(l, "[%s::b] NYLAS [-::-]", color)
}
