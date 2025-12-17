package tui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
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
	color := colorToHex(l.styles.LogoColor)
	fmt.Fprintf(l, "[%s::b] NYLAS [-::-]", color)
}

// colorToHex converts a tcell.Color to a hex string for tview tags.
func colorToHex(c tcell.Color) string {
	r, g, b := c.RGB()
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}
