package tui

import (
	"fmt"
	"strings"

	"github.com/rivo/tview"
)

// Crumbs displays breadcrumb navigation like k9s.
type Crumbs struct {
	*tview.TextView
	styles *Styles
	path   string
}

// NewCrumbs creates a new crumbs component.
func NewCrumbs(styles *Styles) *Crumbs {
	c := &Crumbs{
		TextView: tview.NewTextView(),
		styles:   styles,
	}

	c.SetDynamicColors(true)
	c.SetBackgroundColor(styles.BgColor)
	c.SetTextAlign(tview.AlignLeft)
	c.SetBorderPadding(0, 0, 1, 0)

	return c
}

// SetPath sets the breadcrumb path.
func (c *Crumbs) SetPath(path string) {
	c.path = path
	c.render()
}

func (c *Crumbs) render() {
	c.Clear()

	if c.path == "" {
		return
	}

	// k9s style: crumb with background color
	// Active crumb: black text on orange background
	fg := c.styles.Hex(c.styles.CrumbActiveFg)
	bg := c.styles.Hex(c.styles.CrumbActiveBg)
	_, _ = fmt.Fprintf(c, "[%s:%s:b] :%s [-:-:-]", fg, bg, strings.ToLower(c.path))
}
