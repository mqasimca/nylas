package tui

import "github.com/gdamore/tcell/v2"

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
func AmberStyles() *Styles {
	amber := tcell.NewRGBColor(255, 129, 0)       // #ff8100 - main amber
	amberBright := tcell.NewRGBColor(255, 180, 0) // brighter amber
	amberDim := tcell.NewRGBColor(180, 90, 0)     // dimmer amber
	amberDark := tcell.NewRGBColor(100, 50, 0)    // dark amber

	return &Styles{
		BgColor:     tcell.ColorBlack,
		FgColor:     amber,
		BorderColor: amberDim,
		FocusColor:  amberBright,

		LogoColor: amberBright,

		InfoColor:      amberBright,
		InfoSectionFg:  amber,
		WarnColor:      amberBright,
		ErrorColor:     tcell.NewRGBColor(255, 50, 50),
		SuccessColor:   amberBright,
		PendingColor:   amberDim,
		HighlightColor: amberBright,
		CompletedColor: amberDim,

		TableHeaderFg:    amberBright,
		TableHeaderBg:    tcell.ColorDefault,
		TableSorterColor: amberBright,
		TableRowFg:       amber,
		TableRowBg:       tcell.ColorDefault,
		TableSelectFg:    tcell.ColorBlack,
		TableSelectBg:    amber,
		TableCursorFg:    tcell.ColorBlack,
		TableCursorBg:    amberBright,
		TableMarkColor:   amberBright,

		CrumbFg:       tcell.ColorBlack,
		CrumbBg:       amberDim,
		CrumbActiveFg: tcell.ColorBlack,
		CrumbActiveBg: amber,

		MenuKeyFg:    amberBright,
		MenuNumKeyFg: amber,
		MenuDescFg:   amberDim,

		PromptBg:     tcell.ColorBlack,
		PromptFg:     amber,
		PromptBorder: amberDim,
		SuggestColor: amberDim,

		TitleFg:        amberBright,
		TitleHighlight: amber,
		CounterColor:   amberDim,
		FilterColor:    amberDark,
	}
}

// GreenStyles returns the classic green phosphor CRT theme.
// Based on cool-retro-term Monochrome Green: #0ccc68
func GreenStyles() *Styles {
	green := tcell.NewRGBColor(12, 204, 104)       // #0ccc68 - main green
	greenBright := tcell.NewRGBColor(50, 255, 130) // brighter green
	greenDim := tcell.NewRGBColor(8, 140, 70)      // dimmer green
	greenDark := tcell.NewRGBColor(4, 80, 40)      // dark green

	return &Styles{
		BgColor:     tcell.ColorBlack,
		FgColor:     green,
		BorderColor: greenDim,
		FocusColor:  greenBright,

		LogoColor: greenBright,

		InfoColor:      greenBright,
		InfoSectionFg:  green,
		WarnColor:      tcell.NewRGBColor(255, 200, 0),
		ErrorColor:     tcell.NewRGBColor(255, 50, 50),
		SuccessColor:   greenBright,
		PendingColor:   greenDim,
		HighlightColor: greenBright,
		CompletedColor: greenDim,

		TableHeaderFg:    greenBright,
		TableHeaderBg:    tcell.ColorDefault,
		TableSorterColor: greenBright,
		TableRowFg:       green,
		TableRowBg:       tcell.ColorDefault,
		TableSelectFg:    tcell.ColorBlack,
		TableSelectBg:    green,
		TableCursorFg:    tcell.ColorBlack,
		TableCursorBg:    greenBright,
		TableMarkColor:   greenBright,

		CrumbFg:       tcell.ColorBlack,
		CrumbBg:       greenDim,
		CrumbActiveFg: tcell.ColorBlack,
		CrumbActiveBg: green,

		MenuKeyFg:    greenBright,
		MenuNumKeyFg: green,
		MenuDescFg:   greenDim,

		PromptBg:     tcell.ColorBlack,
		PromptFg:     green,
		PromptBorder: greenDim,
		SuggestColor: greenDim,

		TitleFg:        greenBright,
		TitleHighlight: green,
		CounterColor:   greenDim,
		FilterColor:    greenDark,
	}
}

// AppleIIStyles returns the Apple ][ style green theme.
// Based on cool-retro-term Apple ][: #00d56d
func AppleIIStyles() *Styles {
	green := tcell.NewRGBColor(0, 213, 109)        // #00d56d - Apple II green
	greenBright := tcell.NewRGBColor(50, 255, 150) // brighter
	greenDim := tcell.NewRGBColor(0, 150, 75)      // dimmer
	greenDark := tcell.NewRGBColor(0, 90, 45)      // dark

	return &Styles{
		BgColor:     tcell.ColorBlack,
		FgColor:     green,
		BorderColor: greenDim,
		FocusColor:  greenBright,

		LogoColor: greenBright,

		InfoColor:      greenBright,
		InfoSectionFg:  green,
		WarnColor:      tcell.NewRGBColor(255, 200, 0),
		ErrorColor:     tcell.NewRGBColor(255, 50, 50),
		SuccessColor:   greenBright,
		PendingColor:   greenDim,
		HighlightColor: greenBright,
		CompletedColor: greenDim,

		TableHeaderFg:    greenBright,
		TableHeaderBg:    tcell.ColorDefault,
		TableSorterColor: greenBright,
		TableRowFg:       green,
		TableRowBg:       tcell.ColorDefault,
		TableSelectFg:    tcell.ColorBlack,
		TableSelectBg:    green,
		TableCursorFg:    tcell.ColorBlack,
		TableCursorBg:    greenBright,
		TableMarkColor:   greenBright,

		CrumbFg:       tcell.ColorBlack,
		CrumbBg:       greenDim,
		CrumbActiveFg: tcell.ColorBlack,
		CrumbActiveBg: green,

		MenuKeyFg:    greenBright,
		MenuNumKeyFg: green,
		MenuDescFg:   greenDim,

		PromptBg:     tcell.ColorBlack,
		PromptFg:     green,
		PromptBorder: greenDim,
		SuggestColor: greenDim,

		TitleFg:        greenBright,
		TitleHighlight: green,
		CounterColor:   greenDim,
		FilterColor:    greenDark,
	}
}

// VintageStyles returns the vintage neon green theme.
// Based on cool-retro-term Vintage: #00ff3e
func VintageStyles() *Styles {
	neonGreen := tcell.NewRGBColor(0, 255, 62) // #00ff3e - bright neon green
	neonGreenBright := tcell.NewRGBColor(100, 255, 130)
	neonGreenDim := tcell.NewRGBColor(0, 180, 40)
	neonGreenDark := tcell.NewRGBColor(0, 100, 25)

	return &Styles{
		BgColor:     tcell.ColorBlack,
		FgColor:     neonGreen,
		BorderColor: neonGreenDim,
		FocusColor:  neonGreenBright,

		LogoColor: neonGreenBright,

		InfoColor:      neonGreenBright,
		InfoSectionFg:  neonGreen,
		WarnColor:      tcell.NewRGBColor(255, 200, 0),
		ErrorColor:     tcell.NewRGBColor(255, 50, 50),
		SuccessColor:   neonGreenBright,
		PendingColor:   neonGreenDim,
		HighlightColor: neonGreenBright,
		CompletedColor: neonGreenDim,

		TableHeaderFg:    neonGreenBright,
		TableHeaderBg:    tcell.ColorDefault,
		TableSorterColor: neonGreenBright,
		TableRowFg:       neonGreen,
		TableRowBg:       tcell.ColorDefault,
		TableSelectFg:    tcell.ColorBlack,
		TableSelectBg:    neonGreen,
		TableCursorFg:    tcell.ColorBlack,
		TableCursorBg:    neonGreenBright,
		TableMarkColor:   neonGreenBright,

		CrumbFg:       tcell.ColorBlack,
		CrumbBg:       neonGreenDim,
		CrumbActiveFg: tcell.ColorBlack,
		CrumbActiveBg: neonGreen,

		MenuKeyFg:    neonGreenBright,
		MenuNumKeyFg: neonGreen,
		MenuDescFg:   neonGreenDim,

		PromptBg:     tcell.ColorBlack,
		PromptFg:     neonGreen,
		PromptBorder: neonGreenDim,
		SuggestColor: neonGreenDim,

		TitleFg:        neonGreenBright,
		TitleHighlight: neonGreen,
		CounterColor:   neonGreenDim,
		FilterColor:    neonGreenDark,
	}
}

// IBMDOSStyles returns the IBM DOS white theme.
// Based on cool-retro-term IBM Dos: #ffffff on black
func IBMDOSStyles() *Styles {
	white := tcell.ColorWhite
	gray := tcell.NewRGBColor(180, 180, 180)
	darkGray := tcell.NewRGBColor(100, 100, 100)
	blue := tcell.NewRGBColor(0, 100, 200)

	return &Styles{
		BgColor:     tcell.ColorBlack,
		FgColor:     white,
		BorderColor: gray,
		FocusColor:  white,

		LogoColor: white,

		InfoColor:      white,
		InfoSectionFg:  gray,
		WarnColor:      tcell.NewRGBColor(255, 200, 0),
		ErrorColor:     tcell.NewRGBColor(255, 50, 50),
		SuccessColor:   tcell.NewRGBColor(50, 255, 50),
		PendingColor:   gray,
		HighlightColor: white,
		CompletedColor: darkGray,

		TableHeaderFg:    white,
		TableHeaderBg:    tcell.ColorDefault,
		TableSorterColor: white,
		TableRowFg:       gray,
		TableRowBg:       tcell.ColorDefault,
		TableSelectFg:    tcell.ColorBlack,
		TableSelectBg:    white,
		TableCursorFg:    tcell.ColorBlack,
		TableCursorBg:    white,
		TableMarkColor:   white,

		CrumbFg:       tcell.ColorBlack,
		CrumbBg:       darkGray,
		CrumbActiveFg: tcell.ColorBlack,
		CrumbActiveBg: white,

		MenuKeyFg:    white,
		MenuNumKeyFg: gray,
		MenuDescFg:   darkGray,

		PromptBg:     tcell.ColorBlack,
		PromptFg:     white,
		PromptBorder: gray,
		SuggestColor: darkGray,

		TitleFg:        white,
		TitleHighlight: gray,
		CounterColor:   darkGray,
		FilterColor:    blue,
	}
}

// FuturisticStyles returns the steel blue futuristic theme.
// Based on cool-retro-term Futuristic: #729fcf
func FuturisticStyles() *Styles {
	steelBlue := tcell.NewRGBColor(114, 159, 207) // #729fcf
	steelBlueBright := tcell.NewRGBColor(150, 190, 240)
	steelBlueDim := tcell.NewRGBColor(80, 110, 150)
	steelBlueDark := tcell.NewRGBColor(50, 70, 100)
	cyan := tcell.NewRGBColor(0, 220, 220)

	return &Styles{
		BgColor:     tcell.ColorBlack,
		FgColor:     steelBlue,
		BorderColor: steelBlueDim,
		FocusColor:  cyan,

		LogoColor: cyan,

		InfoColor:      steelBlueBright,
		InfoSectionFg:  steelBlue,
		WarnColor:      tcell.NewRGBColor(255, 200, 0),
		ErrorColor:     tcell.NewRGBColor(255, 50, 50),
		SuccessColor:   cyan,
		PendingColor:   steelBlueDim,
		HighlightColor: cyan,
		CompletedColor: steelBlueDim,

		TableHeaderFg:    steelBlueBright,
		TableHeaderBg:    tcell.ColorDefault,
		TableSorterColor: cyan,
		TableRowFg:       steelBlue,
		TableRowBg:       tcell.ColorDefault,
		TableSelectFg:    tcell.ColorBlack,
		TableSelectBg:    steelBlue,
		TableCursorFg:    tcell.ColorBlack,
		TableCursorBg:    cyan,
		TableMarkColor:   cyan,

		CrumbFg:       tcell.ColorBlack,
		CrumbBg:       steelBlueDim,
		CrumbActiveFg: tcell.ColorBlack,
		CrumbActiveBg: steelBlue,

		MenuKeyFg:    cyan,
		MenuNumKeyFg: steelBlue,
		MenuDescFg:   steelBlueDim,

		PromptBg:     tcell.ColorBlack,
		PromptFg:     steelBlue,
		PromptBorder: steelBlueDim,
		SuggestColor: steelBlueDim,

		TitleFg:        cyan,
		TitleHighlight: steelBlue,
		CounterColor:   steelBlueDim,
		FilterColor:    steelBlueDark,
	}
}

// MatrixStyles returns the Matrix-inspired green rain theme.
func MatrixStyles() *Styles {
	matrixGreen := tcell.NewRGBColor(0, 255, 0)      // Pure green
	matrixBright := tcell.NewRGBColor(150, 255, 150) // Bright green
	matrixDim := tcell.NewRGBColor(0, 150, 0)        // Dim green
	matrixDark := tcell.NewRGBColor(0, 80, 0)        // Dark green

	return &Styles{
		BgColor:     tcell.ColorBlack,
		FgColor:     matrixGreen,
		BorderColor: matrixDim,
		FocusColor:  matrixBright,

		LogoColor: matrixBright,

		InfoColor:      matrixBright,
		InfoSectionFg:  matrixGreen,
		WarnColor:      tcell.NewRGBColor(255, 255, 0),
		ErrorColor:     tcell.NewRGBColor(255, 0, 0),
		SuccessColor:   matrixBright,
		PendingColor:   matrixDim,
		HighlightColor: matrixBright,
		CompletedColor: matrixDim,

		TableHeaderFg:    matrixBright,
		TableHeaderBg:    tcell.ColorDefault,
		TableSorterColor: matrixBright,
		TableRowFg:       matrixGreen,
		TableRowBg:       tcell.ColorDefault,
		TableSelectFg:    tcell.ColorBlack,
		TableSelectBg:    matrixGreen,
		TableCursorFg:    tcell.ColorBlack,
		TableCursorBg:    matrixBright,
		TableMarkColor:   matrixBright,

		CrumbFg:       tcell.ColorBlack,
		CrumbBg:       matrixDim,
		CrumbActiveFg: tcell.ColorBlack,
		CrumbActiveBg: matrixGreen,

		MenuKeyFg:    matrixBright,
		MenuNumKeyFg: matrixGreen,
		MenuDescFg:   matrixDim,

		PromptBg:     tcell.ColorBlack,
		PromptFg:     matrixGreen,
		PromptBorder: matrixDim,
		SuggestColor: matrixDim,

		TitleFg:        matrixBright,
		TitleHighlight: matrixGreen,
		CounterColor:   matrixDim,
		FilterColor:    matrixDark,
	}
}

// NortonStyles returns the classic Norton Commander DOS file manager theme.
// Blue background with cyan text - the iconic DOS look.
func NortonStyles() *Styles {
	// Classic DOS/CGA 16-color palette
	dosBlue := tcell.NewRGBColor(0, 0, 170)          // #0000AA - DOS blue background
	dosCyan := tcell.NewRGBColor(0, 170, 170)        // #00AAAA - Cyan text
	dosLightCyan := tcell.NewRGBColor(85, 255, 255)  // #55FFFF - Light cyan
	dosYellow := tcell.NewRGBColor(255, 255, 85)     // #FFFF55 - Yellow highlight
	dosWhite := tcell.NewRGBColor(255, 255, 255)     // #FFFFFF - Bright white
	dosLightGray := tcell.NewRGBColor(170, 170, 170) // #AAAAAA - Light gray
	dosRed := tcell.NewRGBColor(255, 85, 85)         // #FF5555 - Light red
	dosGreen := tcell.NewRGBColor(85, 255, 85)       // #55FF55 - Light green

	return &Styles{
		BgColor:     dosBlue,
		FgColor:     dosLightCyan,
		BorderColor: dosCyan,
		FocusColor:  dosYellow,

		LogoColor: dosYellow,

		InfoColor:      dosLightCyan,
		InfoSectionFg:  dosCyan,
		WarnColor:      dosYellow,
		ErrorColor:     dosRed,
		SuccessColor:   dosGreen,
		PendingColor:   dosLightGray,
		HighlightColor: dosYellow,
		CompletedColor: dosLightGray,

		TableHeaderFg:    dosYellow,
		TableHeaderBg:    dosBlue,
		TableSorterColor: dosYellow,
		TableRowFg:       dosLightCyan,
		TableRowBg:       dosBlue,
		TableSelectFg:    dosBlue,
		TableSelectBg:    dosCyan,
		TableCursorFg:    dosBlue,
		TableCursorBg:    dosYellow,
		TableMarkColor:   dosYellow,

		CrumbFg:       dosBlue,
		CrumbBg:       dosCyan,
		CrumbActiveFg: dosBlue,
		CrumbActiveBg: dosYellow,

		MenuKeyFg:    dosYellow,
		MenuNumKeyFg: dosLightCyan,
		MenuDescFg:   dosCyan,

		PromptBg:     dosBlue,
		PromptFg:     dosLightCyan,
		PromptBorder: dosCyan,
		SuggestColor: dosCyan,

		TitleFg:        dosYellow,
		TitleHighlight: dosWhite,
		CounterColor:   dosLightGray,
		FilterColor:    dosCyan,
	}
}

// FlashLevel represents the severity of a flash message.
type FlashLevel int

const (
	FlashInfo FlashLevel = iota
	FlashWarn
	FlashError
)
