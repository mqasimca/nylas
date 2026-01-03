package tui

import (
	"fmt"
	"sync"

	"github.com/gdamore/tcell/v2"
)

// ThemeName represents available themes.
type ThemeName string

const (
	ThemeK9s        ThemeName = "k9s"        // Default k9s style
	ThemeAmber      ThemeName = "amber"      // Amber phosphor CRT
	ThemeGreen      ThemeName = "green"      // Green phosphor CRT
	ThemeAppleII    ThemeName = "apple2"     // Apple ][ style
	ThemeVintage    ThemeName = "vintage"    // Vintage neon green
	ThemeIBMDOS     ThemeName = "ibm"        // IBM DOS white
	ThemeFuturistic ThemeName = "futuristic" // Steel blue futuristic
	ThemeMatrix     ThemeName = "matrix"     // Matrix green rain
	ThemeNorton     ThemeName = "norton"     // Norton Commander DOS style
)

// AvailableThemes returns all available theme names.
func AvailableThemes() []ThemeName {
	return []ThemeName{
		ThemeK9s,
		ThemeAmber,
		ThemeGreen,
		ThemeAppleII,
		ThemeVintage,
		ThemeIBMDOS,
		ThemeFuturistic,
		ThemeMatrix,
		ThemeNorton,
	}
}

// Styles holds all TUI color definitions (k9s-style).
type Styles struct {
	// Base colors
	BgColor     tcell.Color
	FgColor     tcell.Color
	BorderColor tcell.Color
	FocusColor  tcell.Color

	// Logo
	LogoColor tcell.Color

	// Status/Info section
	InfoColor      tcell.Color
	InfoSectionFg  tcell.Color
	WarnColor      tcell.Color
	ErrorColor     tcell.Color
	SuccessColor   tcell.Color
	PendingColor   tcell.Color
	HighlightColor tcell.Color
	CompletedColor tcell.Color

	// Table
	TableHeaderFg    tcell.Color
	TableHeaderBg    tcell.Color
	TableSorterColor tcell.Color
	TableRowFg       tcell.Color
	TableRowBg       tcell.Color
	TableSelectFg    tcell.Color
	TableSelectBg    tcell.Color
	TableCursorFg    tcell.Color
	TableCursorBg    tcell.Color
	TableMarkColor   tcell.Color

	// Crumbs
	CrumbFg       tcell.Color
	CrumbBg       tcell.Color
	CrumbActiveFg tcell.Color
	CrumbActiveBg tcell.Color

	// Menu
	MenuKeyFg    tcell.Color
	MenuNumKeyFg tcell.Color
	MenuDescFg   tcell.Color

	// Prompt
	PromptBg     tcell.Color
	PromptFg     tcell.Color
	PromptBorder tcell.Color
	SuggestColor tcell.Color

	// Title
	TitleFg        tcell.Color
	TitleHighlight tcell.Color
	CounterColor   tcell.Color
	FilterColor    tcell.Color

	// Hex color cache - lazily populated, thread-safe
	hexCache map[tcell.Color]string
	hexMu    sync.RWMutex
}

// Hex returns the cached hex string for a color.
// This avoids repeated fmt.Sprintf calls during rendering.
func (s *Styles) Hex(c tcell.Color) string {
	// Fast path: check cache with read lock
	s.hexMu.RLock()
	if hex, ok := s.hexCache[c]; ok {
		s.hexMu.RUnlock()
		return hex
	}
	s.hexMu.RUnlock()

	// Slow path: compute and cache with write lock
	s.hexMu.Lock()
	defer s.hexMu.Unlock()

	// Double-check after acquiring write lock
	if hex, ok := s.hexCache[c]; ok {
		return hex
	}

	// Initialize cache if needed
	if s.hexCache == nil {
		s.hexCache = make(map[tcell.Color]string, 32)
	}

	r, g, b := c.RGB()
	hex := fmt.Sprintf("#%02x%02x%02x", r, g, b)
	s.hexCache[c] = hex
	return hex
}

// DefaultStyles returns the default k9s-like color scheme.
// Based on k9s stock.yaml skin: https://github.com/derailed/k9s/blob/master/skins/stock.yaml
func DefaultStyles() *Styles {
	// k9s stock skin colors
	dodgerblue := tcell.NewRGBColor(30, 144, 255)    // #1E90FF
	orange := tcell.NewRGBColor(255, 165, 0)         // #FFA500
	aqua := tcell.NewRGBColor(0, 255, 255)           // #00FFFF
	steelblue := tcell.NewRGBColor(70, 130, 180)     // #4682B4
	fuchsia := tcell.NewRGBColor(255, 0, 255)        // #FF00FF
	cadetblue := tcell.NewRGBColor(95, 158, 160)     // #5F9EA0
	papayawhip := tcell.NewRGBColor(255, 239, 213)   // #FFEFD5
	darkorange := tcell.NewRGBColor(255, 140, 0)     // #FF8C00
	darkgoldenrod := tcell.NewRGBColor(184, 134, 11) // #B8860B
	orangered := tcell.NewRGBColor(255, 69, 0)       // #FF4500
	greenyellow := tcell.NewRGBColor(173, 255, 47)   // #ADFF2F

	return &Styles{
		// Base - k9s body colors
		BgColor:     tcell.ColorBlack,
		FgColor:     dodgerblue,
		BorderColor: dodgerblue,
		FocusColor:  aqua,

		// Logo - orange like k9s stock
		LogoColor: orange,

		// Status/Info colors - k9s frame.status
		InfoColor:      orange,
		InfoSectionFg:  tcell.ColorWhite,
		WarnColor:      darkorange,
		ErrorColor:     orangered,
		SuccessColor:   greenyellow,
		PendingColor:   darkorange,
		HighlightColor: aqua,
		CompletedColor: tcell.ColorGray,

		// Table - k9s views.table
		TableHeaderFg:    tcell.ColorWhite,
		TableHeaderBg:    tcell.ColorDefault,
		TableSorterColor: orange,
		TableRowFg:       dodgerblue,
		TableRowBg:       tcell.ColorDefault,
		TableSelectFg:    tcell.ColorBlack,
		TableSelectBg:    aqua,
		TableCursorFg:    tcell.ColorBlack,
		TableCursorBg:    aqua,
		TableMarkColor:   darkgoldenrod,

		// Crumbs - k9s frame.crumbs
		CrumbFg:       tcell.ColorBlack,
		CrumbBg:       steelblue,
		CrumbActiveFg: tcell.ColorBlack,
		CrumbActiveBg: orange,

		// Menu - k9s frame.menu
		MenuKeyFg:    dodgerblue,
		MenuNumKeyFg: fuchsia,
		MenuDescFg:   tcell.ColorWhite,

		// Prompt - k9s prompt
		PromptBg:     tcell.ColorBlack,
		PromptFg:     cadetblue,
		PromptBorder: dodgerblue,
		SuggestColor: dodgerblue,

		// Title - k9s frame.title
		TitleFg:        aqua,
		TitleHighlight: fuchsia,
		CounterColor:   papayawhip,
		FilterColor:    steelblue,
	}
}

// GetThemeStyles returns styles for the specified theme.
// It first checks for built-in themes, then looks for custom themes
// in ~/.config/nylas/themes/<name>.yaml
func GetThemeStyles(theme ThemeName) *Styles {
	styles, _ := GetThemeStylesWithError(theme)
	return styles
}

// GetThemeStylesWithError returns styles and any error that occurred while loading.
// This is useful for displaying error messages to users.
func GetThemeStylesWithError(theme ThemeName) (*Styles, error) {
	// Check built-in themes first
	switch theme {
	case ThemeAmber:
		return AmberStyles(), nil
	case ThemeGreen:
		return GreenStyles(), nil
	case ThemeAppleII:
		return AppleIIStyles(), nil
	case ThemeVintage:
		return VintageStyles(), nil
	case ThemeIBMDOS:
		return IBMDOSStyles(), nil
	case ThemeFuturistic:
		return FuturisticStyles(), nil
	case ThemeMatrix:
		return MatrixStyles(), nil
	case ThemeNorton:
		return NortonStyles(), nil
	case ThemeK9s, "":
		return DefaultStyles(), nil
	}

	// Try loading as a custom theme
	styles, err := LoadCustomTheme(string(theme))
	if err != nil {
		// Return default styles but also return the error so caller can display it
		return DefaultStyles(), err
	}

	return styles, nil
}

// IsBuiltInTheme checks if a theme name is a built-in theme.
func IsBuiltInTheme(theme ThemeName) bool {
	switch theme {
	case ThemeK9s, ThemeAmber, ThemeGreen, ThemeAppleII, ThemeVintage,
		ThemeIBMDOS, ThemeFuturistic, ThemeMatrix, ThemeNorton, "":
		return true
	}
	return false
}

// AmberStyles returns the classic amber phosphor CRT theme.
// Based on cool-retro-term Default Amber: #ff8100
