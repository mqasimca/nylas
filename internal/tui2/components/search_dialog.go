package components

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/mqasimca/nylas/internal/tui2/styles"
)

// SearchDialogField represents a field in the search dialog.
type SearchDialogField int

const (
	SearchDialogFieldFrom SearchDialogField = iota
	SearchDialogFieldTo
	SearchDialogFieldSubject
	SearchDialogFieldHas
	SearchDialogFieldAfter
	SearchDialogFieldBefore
	SearchDialogFieldIn
	SearchDialogFieldUnread
	SearchDialogFieldStarred
	SearchDialogFieldText
	SearchDialogFieldSearch // Search button
	SearchDialogFieldCancel // Cancel button
	searchDialogFieldCount
)

// SearchDialog is an advanced search dialog component.
type SearchDialog struct {
	theme *styles.Theme

	// Input fields
	fromInput    textinput.Model
	toInput      textinput.Model
	subjectInput textinput.Model
	hasInput     textinput.Model
	afterInput   textinput.Model
	beforeInput  textinput.Model
	inInput      textinput.Model
	textInput    textinput.Model

	// Toggle fields
	unread  *bool // nil = any, true = unread, false = read
	starred *bool // nil = any, true = starred, false = not starred

	// State
	focusedField SearchDialogField
	visible      bool
	width        int
	height       int
}

// SearchDialogSubmitMsg is sent when the search is submitted.
type SearchDialogSubmitMsg struct {
	Query string
}

// SearchDialogCancelMsg is sent when the dialog is cancelled.
type SearchDialogCancelMsg struct{}

// NewSearchDialog creates a new advanced search dialog.
func NewSearchDialog(theme *styles.Theme) *SearchDialog {
	d := &SearchDialog{
		theme:        theme,
		focusedField: SearchDialogFieldFrom,
		visible:      true,
	}

	// Create text inputs
	d.fromInput = createSearchInput("Email addresses to search in From field")
	d.toInput = createSearchInput("Email addresses to search in To field")
	d.subjectInput = createSearchInput("Text to search in subject line")
	d.hasInput = createSearchInput("attachment")
	d.afterInput = createSearchInput("YYYY-MM-DD")
	d.beforeInput = createSearchInput("YYYY-MM-DD")
	d.inInput = createSearchInput("Folder name (inbox, sent, drafts)")
	d.textInput = createSearchInput("Free text to search in body")

	// Focus first field
	d.fromInput.Focus()

	return d
}

// createSearchInput creates a text input for the search dialog.
func createSearchInput(placeholder string) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.CharLimit = 256
	return ti
}

// Init implements tea.Model.
func (d *SearchDialog) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements tea.Model.
func (d *SearchDialog) Update(msg tea.Msg) (*SearchDialog, tea.Cmd) {
	if !d.visible {
		return d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.Key()
		keyStr := msg.String()

		switch {
		case key.Code == tea.KeyEsc:
			d.visible = false
			return d, func() tea.Msg { return SearchDialogCancelMsg{} }

		case key.Code == tea.KeyTab, keyStr == "down":
			d.focusNext()
			return d, nil

		case keyStr == "shift+tab", keyStr == "up":
			d.focusPrev()
			return d, nil

		case key.Code == tea.KeyEnter:
			switch d.focusedField {
			case SearchDialogFieldSearch:
				query := d.BuildQuery()
				d.visible = false
				return d, func() tea.Msg { return SearchDialogSubmitMsg{Query: query} }
			case SearchDialogFieldCancel:
				d.visible = false
				return d, func() tea.Msg { return SearchDialogCancelMsg{} }
			case SearchDialogFieldUnread:
				d.toggleUnread()
				return d, nil
			case SearchDialogFieldStarred:
				d.toggleStarred()
				return d, nil
			default:
				// Move to next field on Enter for text inputs
				d.focusNext()
				return d, nil
			}

		case key.Code == tea.KeySpace, keyStr == " ":
			// Space toggles checkboxes
			switch d.focusedField {
			case SearchDialogFieldUnread:
				d.toggleUnread()
				return d, nil
			case SearchDialogFieldStarred:
				d.toggleStarred()
				return d, nil
			}
		}
	}

	// Update the focused text input
	var cmd tea.Cmd
	switch d.focusedField {
	case SearchDialogFieldFrom:
		d.fromInput, cmd = d.fromInput.Update(msg)
	case SearchDialogFieldTo:
		d.toInput, cmd = d.toInput.Update(msg)
	case SearchDialogFieldSubject:
		d.subjectInput, cmd = d.subjectInput.Update(msg)
	case SearchDialogFieldHas:
		d.hasInput, cmd = d.hasInput.Update(msg)
	case SearchDialogFieldAfter:
		d.afterInput, cmd = d.afterInput.Update(msg)
	case SearchDialogFieldBefore:
		d.beforeInput, cmd = d.beforeInput.Update(msg)
	case SearchDialogFieldIn:
		d.inInput, cmd = d.inInput.Update(msg)
	case SearchDialogFieldText:
		d.textInput, cmd = d.textInput.Update(msg)
	}

	return d, cmd
}

// toggleUnread cycles through: nil (any) -> true (unread) -> false (read) -> nil
func (d *SearchDialog) toggleUnread() {
	if d.unread == nil {
		t := true
		d.unread = &t
	} else if *d.unread {
		f := false
		d.unread = &f
	} else {
		d.unread = nil
	}
}

// toggleStarred cycles through: nil (any) -> true (starred) -> false (not starred) -> nil
func (d *SearchDialog) toggleStarred() {
	if d.starred == nil {
		t := true
		d.starred = &t
	} else if *d.starred {
		f := false
		d.starred = &f
	} else {
		d.starred = nil
	}
}

// focusNext moves focus to the next field.
func (d *SearchDialog) focusNext() {
	d.blurCurrent()
	d.focusedField = (d.focusedField + 1) % searchDialogFieldCount
	d.focusCurrent()
}

// focusPrev moves focus to the previous field.
func (d *SearchDialog) focusPrev() {
	d.blurCurrent()
	d.focusedField = (d.focusedField - 1 + searchDialogFieldCount) % searchDialogFieldCount
	d.focusCurrent()
}

// blurCurrent blurs the current field.
func (d *SearchDialog) blurCurrent() {
	switch d.focusedField {
	case SearchDialogFieldFrom:
		d.fromInput.Blur()
	case SearchDialogFieldTo:
		d.toInput.Blur()
	case SearchDialogFieldSubject:
		d.subjectInput.Blur()
	case SearchDialogFieldHas:
		d.hasInput.Blur()
	case SearchDialogFieldAfter:
		d.afterInput.Blur()
	case SearchDialogFieldBefore:
		d.beforeInput.Blur()
	case SearchDialogFieldIn:
		d.inInput.Blur()
	case SearchDialogFieldText:
		d.textInput.Blur()
	}
}

// focusCurrent focuses the current field.
func (d *SearchDialog) focusCurrent() {
	switch d.focusedField {
	case SearchDialogFieldFrom:
		d.fromInput.Focus()
	case SearchDialogFieldTo:
		d.toInput.Focus()
	case SearchDialogFieldSubject:
		d.subjectInput.Focus()
	case SearchDialogFieldHas:
		d.hasInput.Focus()
	case SearchDialogFieldAfter:
		d.afterInput.Focus()
	case SearchDialogFieldBefore:
		d.beforeInput.Focus()
	case SearchDialogFieldIn:
		d.inInput.Focus()
	case SearchDialogFieldText:
		d.textInput.Focus()
	}
}

// View implements tea.Model.
func (d *SearchDialog) View() string {
	if !d.visible {
		return ""
	}

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(d.theme.Primary).
		MarginBottom(1)
	title := titleStyle.Render("🔍 Advanced Search")

	// Build the form
	var rows []string

	// Helper to render a labeled field
	renderField := func(label string, input textinput.Model, focused bool) string {
		labelStyle := lipgloss.NewStyle().Width(12).Foreground(d.theme.Secondary)
		if focused {
			labelStyle = labelStyle.Bold(true).Foreground(d.theme.Primary)
		}

		inputWidth := d.width - 20
		if inputWidth < 30 {
			inputWidth = 30
		}

		inputStyle := lipgloss.NewStyle().Width(inputWidth)
		if focused {
			inputStyle = inputStyle.BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(d.theme.Primary).
				BorderLeft(true)
		}

		return lipgloss.JoinHorizontal(lipgloss.Left,
			labelStyle.Render(label+":"),
			inputStyle.Render(input.View()),
		)
	}

	// Helper to render a toggle field
	renderToggle := func(label string, value *bool, focused bool) string {
		labelStyle := lipgloss.NewStyle().Width(12).Foreground(d.theme.Secondary)
		if focused {
			labelStyle = labelStyle.Bold(true).Foreground(d.theme.Primary)
		}

		var indicator string
		if value == nil {
			indicator = "[ ] Any"
		} else if *value {
			indicator = "[✓] Yes"
		} else {
			indicator = "[✗] No"
		}

		indicatorStyle := lipgloss.NewStyle()
		if focused {
			indicatorStyle = indicatorStyle.Bold(true).Foreground(d.theme.Primary)
		}

		return lipgloss.JoinHorizontal(lipgloss.Left,
			labelStyle.Render(label+":"),
			indicatorStyle.Render(indicator),
		)
	}

	// Add all fields
	rows = append(rows, renderField("From", d.fromInput, d.focusedField == SearchDialogFieldFrom))
	rows = append(rows, renderField("To", d.toInput, d.focusedField == SearchDialogFieldTo))
	rows = append(rows, renderField("Subject", d.subjectInput, d.focusedField == SearchDialogFieldSubject))
	rows = append(rows, renderField("Has", d.hasInput, d.focusedField == SearchDialogFieldHas))
	rows = append(rows, renderField("After", d.afterInput, d.focusedField == SearchDialogFieldAfter))
	rows = append(rows, renderField("Before", d.beforeInput, d.focusedField == SearchDialogFieldBefore))
	rows = append(rows, renderField("Folder", d.inInput, d.focusedField == SearchDialogFieldIn))
	rows = append(rows, renderToggle("Unread", d.unread, d.focusedField == SearchDialogFieldUnread))
	rows = append(rows, renderToggle("Starred", d.starred, d.focusedField == SearchDialogFieldStarred))
	rows = append(rows, renderField("Text", d.textInput, d.focusedField == SearchDialogFieldText))

	// Spacer
	rows = append(rows, "")

	// Buttons
	searchStyle := lipgloss.NewStyle().
		Padding(0, 2).
		MarginRight(2)
	cancelStyle := lipgloss.NewStyle().
		Padding(0, 2)

	if d.focusedField == SearchDialogFieldSearch {
		searchStyle = searchStyle.
			Background(d.theme.Primary).
			Foreground(lipgloss.Color("#000000")).
			Bold(true)
	} else {
		searchStyle = searchStyle.
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(d.theme.Primary)
	}

	if d.focusedField == SearchDialogFieldCancel {
		cancelStyle = cancelStyle.
			Background(d.theme.Secondary).
			Foreground(lipgloss.Color("#000000")).
			Bold(true)
	} else {
		cancelStyle = cancelStyle.
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(d.theme.Secondary)
	}

	buttons := lipgloss.JoinHorizontal(lipgloss.Center,
		searchStyle.Render("Search"),
		cancelStyle.Render("Cancel"),
	)
	rows = append(rows, buttons)

	// Help text
	helpStyle := lipgloss.NewStyle().Foreground(d.theme.Dimmed.GetForeground()).MarginTop(1)
	rows = append(rows, helpStyle.Render("Tab: next field  Shift+Tab: previous  Enter: search/toggle  Esc: cancel"))

	// Preview of query
	query := d.BuildQuery()
	if query != "" {
		previewStyle := lipgloss.NewStyle().
			Foreground(d.theme.Info).
			Italic(true).
			MarginTop(1)
		rows = append(rows, previewStyle.Render("Query: "+query))
	}

	// Join all rows
	content := lipgloss.JoinVertical(lipgloss.Left, rows...)

	// Dialog box
	dialogStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(d.theme.Primary).
		Padding(1, 2).
		Width(d.width - 10)

	// Center the dialog
	dialog := dialogStyle.Render(title + "\n" + content)
	return lipgloss.Place(
		d.width,
		d.height,
		lipgloss.Center,
		lipgloss.Center,
		dialog,
	)
}

// BuildQuery builds the search query string from the form values.
func (d *SearchDialog) BuildQuery() string {
	var parts []string

	if v := strings.TrimSpace(d.fromInput.Value()); v != "" {
		parts = append(parts, "from:"+quoteIfNeeded(v))
	}
	if v := strings.TrimSpace(d.toInput.Value()); v != "" {
		parts = append(parts, "to:"+quoteIfNeeded(v))
	}
	if v := strings.TrimSpace(d.subjectInput.Value()); v != "" {
		parts = append(parts, "subject:"+quoteIfNeeded(v))
	}
	if v := strings.TrimSpace(d.hasInput.Value()); v != "" {
		parts = append(parts, "has:"+v)
	}
	if v := strings.TrimSpace(d.afterInput.Value()); v != "" {
		parts = append(parts, "after:"+v)
	}
	if v := strings.TrimSpace(d.beforeInput.Value()); v != "" {
		parts = append(parts, "before:"+v)
	}
	if v := strings.TrimSpace(d.inInput.Value()); v != "" {
		parts = append(parts, "in:"+v)
	}
	if d.unread != nil {
		if *d.unread {
			parts = append(parts, "is:unread")
		} else {
			parts = append(parts, "is:read")
		}
	}
	if d.starred != nil {
		if *d.starred {
			parts = append(parts, "is:starred")
		} else {
			parts = append(parts, "is:unstarred")
		}
	}
	if v := strings.TrimSpace(d.textInput.Value()); v != "" {
		// Free text goes at the end
		parts = append(parts, v)
	}

	return strings.Join(parts, " ")
}

// quoteIfNeeded adds quotes around a value if it contains spaces.
func quoteIfNeeded(s string) string {
	if strings.Contains(s, " ") {
		return `"` + s + `"`
	}
	return s
}

// SetSize sets the size of the dialog.
func (d *SearchDialog) SetSize(width, height int) {
	d.width = width
	d.height = height
	// Note: In Bubble Tea v2, textinput width is controlled by CharLimit
	// and visual width is determined by the terminal, not a Width property
}

// Show shows the dialog.
func (d *SearchDialog) Show() {
	d.visible = true
	d.focusCurrent()
}

// Hide hides the dialog.
func (d *SearchDialog) Hide() {
	d.visible = false
	d.blurCurrent()
}

// IsVisible returns whether the dialog is visible.
func (d *SearchDialog) IsVisible() bool {
	return d.visible
}

// Reset clears all fields.
func (d *SearchDialog) Reset() {
	d.fromInput.SetValue("")
	d.toInput.SetValue("")
	d.subjectInput.SetValue("")
	d.hasInput.SetValue("")
	d.afterInput.SetValue("")
	d.beforeInput.SetValue("")
	d.inInput.SetValue("")
	d.textInput.SetValue("")
	d.unread = nil
	d.starred = nil
	d.focusedField = SearchDialogFieldFrom
	d.focusCurrent()
}

// SetQuery parses an existing query and populates the form fields.
func (d *SearchDialog) SetQuery(query string) {
	parsed := ParseSearchQuery(query)

	if len(parsed.From) > 0 {
		d.fromInput.SetValue(strings.Join(parsed.From, ", "))
	}
	if len(parsed.To) > 0 {
		d.toInput.SetValue(strings.Join(parsed.To, ", "))
	}
	if len(parsed.Subject) > 0 {
		d.subjectInput.SetValue(strings.Join(parsed.Subject, " "))
	}
	if len(parsed.Has) > 0 {
		d.hasInput.SetValue(strings.Join(parsed.Has, ", "))
	}
	if parsed.After != "" {
		d.afterInput.SetValue(parsed.After)
	}
	if parsed.Before != "" {
		d.beforeInput.SetValue(parsed.Before)
	}
	if len(parsed.In) > 0 {
		d.inInput.SetValue(strings.Join(parsed.In, ", "))
	}
	if parsed.Text != "" {
		d.textInput.SetValue(parsed.Text)
	}

	// Handle is: operators
	for _, is := range parsed.Is {
		switch is {
		case "unread":
			t := true
			d.unread = &t
		case "read":
			f := false
			d.unread = &f
		case "starred":
			t := true
			d.starred = &t
		case "unstarred":
			f := false
			d.starred = &f
		}
	}
}
