package tui

import "github.com/gdamore/tcell/v2"

// Styles holds all TUI color definitions (k9s-style).
type Styles struct {
	// Base colors
	BgColor     tcell.Color
	FgColor     tcell.Color
	BorderColor tcell.Color

	// Logo
	LogoColor tcell.Color

	// Status
	InfoColor    tcell.Color
	WarnColor    tcell.Color
	ErrorColor   tcell.Color
	SuccessColor tcell.Color

	// Table
	TableHeaderFg  tcell.Color
	TableHeaderBg  tcell.Color
	TableRowFg     tcell.Color
	TableRowBg     tcell.Color
	TableSelectFg  tcell.Color
	TableSelectBg  tcell.Color
	TableCursorFg  tcell.Color
	TableCursorBg  tcell.Color

	// Crumbs
	CrumbFg       tcell.Color
	CrumbBg       tcell.Color
	CrumbActiveFg tcell.Color
	CrumbActiveBg tcell.Color

	// Menu
	MenuKeyFg  tcell.Color
	MenuDescFg tcell.Color

	// Prompt
	PromptBg      tcell.Color
	PromptFg      tcell.Color
	PromptBorder  tcell.Color
}

// DefaultStyles returns the default k9s-like color scheme.
func DefaultStyles() *Styles {
	return &Styles{
		// Base - dark background
		BgColor:     tcell.ColorDefault,
		FgColor:     tcell.NewRGBColor(200, 200, 200),
		BorderColor: tcell.NewRGBColor(80, 80, 80),

		// Logo - cyan like k9s
		LogoColor: tcell.NewRGBColor(0, 215, 255),

		// Status colors
		InfoColor:    tcell.NewRGBColor(0, 215, 255),
		WarnColor:    tcell.NewRGBColor(255, 175, 0),
		ErrorColor:   tcell.NewRGBColor(255, 85, 85),
		SuccessColor: tcell.NewRGBColor(0, 255, 0),

		// Table
		TableHeaderFg:  tcell.NewRGBColor(255, 175, 0),
		TableHeaderBg:  tcell.ColorDefault,
		TableRowFg:     tcell.NewRGBColor(200, 200, 200),
		TableRowBg:     tcell.ColorDefault,
		TableSelectFg:  tcell.ColorBlack,
		TableSelectBg:  tcell.NewRGBColor(0, 215, 255),
		TableCursorFg:  tcell.ColorBlack,
		TableCursorBg:  tcell.NewRGBColor(255, 175, 0),

		// Crumbs
		CrumbFg:       tcell.NewRGBColor(200, 200, 200),
		CrumbBg:       tcell.NewRGBColor(50, 50, 50),
		CrumbActiveFg: tcell.ColorBlack,
		CrumbActiveBg: tcell.NewRGBColor(0, 215, 255),

		// Menu
		MenuKeyFg:  tcell.NewRGBColor(255, 175, 0),
		MenuDescFg: tcell.NewRGBColor(150, 150, 150),

		// Prompt
		PromptBg:     tcell.ColorDefault,
		PromptFg:     tcell.NewRGBColor(200, 200, 200),
		PromptBorder: tcell.NewRGBColor(0, 215, 255),
	}
}

// FlashLevel represents the severity of a flash message.
type FlashLevel int

const (
	FlashInfo FlashLevel = iota
	FlashWarn
	FlashError
)
