package tui

import "github.com/rivo/tview"

// DetailViewConfig holds optional configuration for detail text views.
type DetailViewConfig struct {
	Title      string // Optional title displayed in border
	Border     bool   // Whether to show border (default false for messages, true for others)
	Scrollable bool   // Whether view is scrollable (default true)
}

// NewStyledDetailView creates a pre-configured text view for detail panels.
// This consolidates the common pattern:
//
//	detail := tview.NewTextView()
//	detail.SetDynamicColors(true)
//	detail.SetBackgroundColor(v.app.styles.BgColor)
//	detail.SetBorder(true)
//	detail.SetBorderColor(v.app.styles.FocusColor)
//	detail.SetTitle(" Title ")
//	detail.SetTitleColor(v.app.styles.TitleFg)
//	detail.SetBorderPadding(1, 1, 2, 2)
//	detail.SetScrollable(true)
//
// Usage:
//
//	detail := NewStyledDetailView(v.app.styles, DetailViewConfig{Title: "Event", Border: true})
func NewStyledDetailView(styles *Styles, cfg DetailViewConfig) *tview.TextView {
	detail := tview.NewTextView()
	detail.SetDynamicColors(true)
	detail.SetBackgroundColor(styles.BgColor)
	detail.SetBorderPadding(1, 1, 2, 2)

	if cfg.Scrollable {
		detail.SetScrollable(true)
	}

	if cfg.Border {
		detail.SetBorder(true)
		detail.SetBorderColor(styles.FocusColor)
		detail.SetTitleColor(styles.TitleFg)
		if cfg.Title != "" {
			detail.SetTitle(" " + cfg.Title + " ")
		}
	}

	return detail
}

// ListViewConfig holds optional configuration for styled lists.
type ListViewConfig struct {
	Title             string // Title displayed in border
	ShowSecondaryText bool   // Whether to show secondary text line
	HighlightFullLine bool   // Whether selection highlights full line
	UseTableSelectBg  bool   // Use TableSelectBg instead of FocusColor for selection
}

// NewStyledList creates a pre-configured list view.
// This consolidates the common pattern:
//
//	list := tview.NewList()
//	list.SetBackgroundColor(styles.BgColor)
//	list.SetMainTextColor(styles.FgColor)
//	list.SetSecondaryTextColor(styles.InfoColor)
//	list.SetSelectedBackgroundColor(styles.FocusColor)
//	list.SetSelectedTextColor(styles.BgColor)
//	list.SetBorder(true)
//	list.SetBorderColor(styles.BorderColor)
//	list.SetTitle(" Title ")
//	list.SetTitleColor(styles.TitleFg)
//	list.ShowSecondaryText(false)
//
// Usage:
//
//	list := NewStyledList(v.app.styles, ListViewConfig{Title: "Participants", ShowSecondaryText: true})
func NewStyledList(styles *Styles, cfg ListViewConfig) *tview.List {
	list := tview.NewList()
	list.SetBackgroundColor(styles.BgColor)
	list.SetMainTextColor(styles.FgColor)
	list.SetSecondaryTextColor(styles.InfoColor)
	list.SetBorder(true)
	list.SetBorderColor(styles.BorderColor)
	list.SetTitleColor(styles.TitleFg)

	if cfg.Title != "" {
		list.SetTitle(" " + cfg.Title + " ")
	}

	// Selection colors
	if cfg.UseTableSelectBg {
		list.SetSelectedBackgroundColor(styles.TableSelectBg)
		list.SetSelectedTextColor(styles.TableSelectFg)
	} else {
		list.SetSelectedBackgroundColor(styles.FocusColor)
		list.SetSelectedTextColor(styles.BgColor)
	}

	list.ShowSecondaryText(cfg.ShowSecondaryText)

	if cfg.HighlightFullLine {
		list.SetHighlightFullLine(true)
	}

	return list
}

// NewStyledInfoPanel creates a text view styled for info panels (smaller padding).
// Common pattern used for settings, status, and timeline panels.
func NewStyledInfoPanel(styles *Styles, title string) *tview.TextView {
	panel := tview.NewTextView()
	panel.SetDynamicColors(true)
	panel.SetBackgroundColor(styles.BgColor)
	panel.SetBorder(true)
	panel.SetBorderColor(styles.BorderColor)
	panel.SetTitleColor(styles.TitleFg)
	panel.SetBorderPadding(0, 0, 1, 1)

	if title != "" {
		panel.SetTitle(" " + title + " ")
	}

	return panel
}

// ColorCache provides cached hex color strings for a Styles instance.
// This avoids repeated colorToHex() calls during rendering.
type ColorCache struct {
	Title   string // TitleFg
	Key     string // FgColor
	Value   string // InfoSectionFg
	Muted   string // BorderColor
	Info    string // InfoColor
	Hint    string // InfoColor (alias for clarity)
	Success string // SuccessColor
	Error   string // ErrorColor
	Warn    string // WarnColor
}

// NewColorCache creates a ColorCache from styles, using the cached Hex() method.
func NewColorCache(styles *Styles) *ColorCache {
	return &ColorCache{
		Title:   styles.Hex(styles.TitleFg),
		Key:     styles.Hex(styles.FgColor),
		Value:   styles.Hex(styles.InfoSectionFg),
		Muted:   styles.Hex(styles.BorderColor),
		Info:    styles.Hex(styles.InfoColor),
		Hint:    styles.Hex(styles.InfoColor),
		Success: styles.Hex(styles.SuccessColor),
		Error:   styles.Hex(styles.ErrorColor),
		Warn:    styles.Hex(styles.WarnColor),
	}
}
