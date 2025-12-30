package tui

// ThemeConfig represents a customizable theme loaded from YAML (k9s-style).
type ThemeConfig struct {
	// Color definitions (like k9s anchors)
	Foreground string `yaml:"foreground"`
	Background string `yaml:"background"`
	Black      string `yaml:"black"`
	Red        string `yaml:"red"`
	Green      string `yaml:"green"`
	Yellow     string `yaml:"yellow"`
	Blue       string `yaml:"blue"`
	Magenta    string `yaml:"magenta"`
	Cyan       string `yaml:"cyan"`
	White      string `yaml:"white"`

	// K9s-style skin configuration
	K9s K9sSkin `yaml:"k9s"`
}

// K9sSkin represents the k9s skin configuration.
type K9sSkin struct {
	Body   BodyStyle   `yaml:"body"`
	Prompt PromptStyle `yaml:"prompt"`
	Info   InfoStyle   `yaml:"info"`
	Frame  FrameStyle  `yaml:"frame"`
	Views  ViewsStyle  `yaml:"views"`
}

// BodyStyle for general body colors.
type BodyStyle struct {
	FgColor   string `yaml:"fgColor"`
	BgColor   string `yaml:"bgColor"`
	LogoColor string `yaml:"logoColor"`
}

// PromptStyle for command prompt.
type PromptStyle struct {
	FgColor      string `yaml:"fgColor"`
	BgColor      string `yaml:"bgColor"`
	SuggestColor string `yaml:"suggestColor"`
}

// InfoStyle for info panel.
type InfoStyle struct {
	FgColor      string `yaml:"fgColor"`
	SectionColor string `yaml:"sectionColor"`
}

// FrameStyle for frame elements.
type FrameStyle struct {
	Border BorderStyle `yaml:"border"`
	Menu   MenuStyle   `yaml:"menu"`
	Crumbs CrumbsStyle `yaml:"crumbs"`
	Status StatusStyle `yaml:"status"`
	Title  TitleStyle  `yaml:"title"`
}

// BorderStyle for borders.
type BorderStyle struct {
	FgColor    string `yaml:"fgColor"`
	FocusColor string `yaml:"focusColor"`
}

// MenuStyle for menu.
type MenuStyle struct {
	FgColor     string `yaml:"fgColor"`
	KeyColor    string `yaml:"keyColor"`
	NumKeyColor string `yaml:"numKeyColor"`
}

// CrumbsStyle for breadcrumbs.
type CrumbsStyle struct {
	FgColor     string `yaml:"fgColor"`
	BgColor     string `yaml:"bgColor"`
	ActiveColor string `yaml:"activeColor"`
}

// StatusStyle for status indicators.
type StatusStyle struct {
	NewColor       string `yaml:"newColor"`
	ModifyColor    string `yaml:"modifyColor"`
	AddColor       string `yaml:"addColor"`
	PendingColor   string `yaml:"pendingColor"`
	ErrorColor     string `yaml:"errorColor"`
	HighlightColor string `yaml:"highlightColor"`
	KillColor      string `yaml:"killColor"`
	CompletedColor string `yaml:"completedColor"`
}

// TitleStyle for titles.
type TitleStyle struct {
	FgColor        string `yaml:"fgColor"`
	BgColor        string `yaml:"bgColor"`
	HighlightColor string `yaml:"highlightColor"`
	CounterColor   string `yaml:"counterColor"`
	FilterColor    string `yaml:"filterColor"`
}

// ViewsStyle for view-specific styles.
type ViewsStyle struct {
	Table TableStyle `yaml:"table"`
}

// TableStyle for table views.
type TableStyle struct {
	FgColor   string             `yaml:"fgColor"`
	BgColor   string             `yaml:"bgColor"`
	MarkColor string             `yaml:"markColor"`
	Header    TableHeaderStyle   `yaml:"header"`
	Selected  TableSelectedStyle `yaml:"selected"`
}

// TableHeaderStyle for table headers.
type TableHeaderStyle struct {
	FgColor     string `yaml:"fgColor"`
	BgColor     string `yaml:"bgColor"`
	SorterColor string `yaml:"sorterColor"`
}

// TableSelectedStyle for selected rows.
type TableSelectedStyle struct {
	FgColor string `yaml:"fgColor"`
	BgColor string `yaml:"bgColor"`
}

// ThemeLoadError provides detailed error information for theme loading failures.
