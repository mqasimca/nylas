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

	// k9s style: show current resource with colon prefix
	info := colorToHex(c.styles.InfoColor)
	fmt.Fprintf(c, "[%s::b]:%s[-::-]", info, strings.ToLower(c.path))
}
