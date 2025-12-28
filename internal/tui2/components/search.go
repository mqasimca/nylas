// Package components provides reusable Bubble Tea components.
package components

import (
	"regexp"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/mqasimca/nylas/internal/tui2/styles"
)

// SearchQuery represents a parsed search query with operators.
type SearchQuery struct {
	Raw     string   // Original query string
	Text    string   // Free text (not in operators)
	From    []string // from: operator values
	To      []string // to: operator values
	Subject []string // subject: operator values
	Has     []string // has: operator values (e.g., "attachment")
	Is      []string // is: operator values (e.g., "unread", "starred")
	After   string   // after: date filter (YYYY-MM-DD)
	Before  string   // before: date filter (YYYY-MM-DD)
	In      []string // in: folder filter
}

// IsEmpty returns true if the query has no search criteria.
func (q *SearchQuery) IsEmpty() bool {
	return q.Raw == "" && q.Text == "" &&
		len(q.From) == 0 && len(q.To) == 0 &&
		len(q.Subject) == 0 && len(q.Has) == 0 &&
		len(q.Is) == 0 && q.After == "" && q.Before == "" &&
		len(q.In) == 0
}

// ToNativeQuery converts the parsed query to a Gmail-style search string.
func (q *SearchQuery) ToNativeQuery() string {
	var parts []string

	// Add operator values
	for _, v := range q.From {
		parts = append(parts, "from:"+v)
	}
	for _, v := range q.To {
		parts = append(parts, "to:"+v)
	}
	for _, v := range q.Subject {
		parts = append(parts, "subject:"+v)
	}
	for _, v := range q.Has {
		parts = append(parts, "has:"+v)
	}
	for _, v := range q.Is {
		parts = append(parts, "is:"+v)
	}
	if q.After != "" {
		parts = append(parts, "after:"+q.After)
	}
	if q.Before != "" {
		parts = append(parts, "before:"+q.Before)
	}
	for _, v := range q.In {
		parts = append(parts, "in:"+v)
	}

	// Add free text at the end
	if q.Text != "" {
		parts = append(parts, q.Text)
	}

	return strings.Join(parts, " ")
}

// Search is a search input component.
type Search struct {
	input textinput.Model
	theme *styles.Theme
	query *SearchQuery
	width int

	// Callback message when search changes
	OnChange func(query *SearchQuery) tea.Msg
}

// NewSearch creates a new search component.
func NewSearch(theme *styles.Theme) *Search {
	input := textinput.New()
	input.Placeholder = "Search... (from:, to:, subject:, is:unread, has:attachment)"
	input.Prompt = "/ "
	input.CharLimit = 256

	// Set styles using v2 API
	inputStyles := textinput.DefaultDarkStyles()
	inputStyles.Focused.Prompt = lipgloss.NewStyle().Foreground(theme.Primary).Bold(true)
	inputStyles.Focused.Text = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	inputStyles.Focused.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	inputStyles.Blurred.Prompt = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	inputStyles.Blurred.Text = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	inputStyles.Blurred.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	input.SetStyles(inputStyles)

	return &Search{
		input: input,
		theme: theme,
		query: &SearchQuery{},
	}
}

// Focus focuses the search input.
func (s *Search) Focus() tea.Cmd {
	return s.input.Focus()
}

// Blur unfocuses the search input.
func (s *Search) Blur() {
	s.input.Blur()
}

// Focused returns true if the search input is focused.
func (s *Search) Focused() bool {
	return s.input.Focused()
}

// SetWidth sets the width of the search input.
func (s *Search) SetWidth(width int) {
	s.width = width
	inputWidth := width - 4 // Account for prompt and padding
	if inputWidth < 20 {
		inputWidth = 20
	}
	s.input.SetWidth(inputWidth)
}

// Value returns the current search value.
func (s *Search) Value() string {
	return s.input.Value()
}

// SetValue sets the search value.
func (s *Search) SetValue(value string) {
	s.input.SetValue(value)
	s.query = ParseSearchQuery(value)
}

// Query returns the parsed search query.
func (s *Search) Query() *SearchQuery {
	return s.query
}

// Reset clears the search input.
func (s *Search) Reset() {
	s.input.Reset()
	s.query = &SearchQuery{}
}

// Update handles messages for the search component.
func (s *Search) Update(msg tea.Msg) (*Search, tea.Cmd) {
	var cmd tea.Cmd

	// Track old value for change detection
	oldValue := s.input.Value()

	// Update the input
	s.input, cmd = s.input.Update(msg)

	// If value changed, parse the new query
	newValue := s.input.Value()
	if oldValue != newValue {
		s.query = ParseSearchQuery(newValue)

		// Emit change callback if set
		if s.OnChange != nil {
			changeMsg := s.OnChange(s.query)
			if changeMsg != nil {
				return s, tea.Batch(cmd, func() tea.Msg { return changeMsg })
			}
		}
	}

	return s, cmd
}

// View renders the search component.
func (s *Search) View() string {
	// Build the search bar
	style := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(s.theme.Primary).
		Padding(0, 1).
		Width(s.width - 2)

	return style.Render(s.input.View())
}

// ViewInline renders the search component inline (without border).
func (s *Search) ViewInline() string {
	return s.input.View()
}

// ParseSearchQuery parses a Gmail-style search query.
func ParseSearchQuery(query string) *SearchQuery {
	result := &SearchQuery{Raw: query}

	if query == "" {
		return result
	}

	// Regular expression to match operators
	// Matches: operator:value or operator:"value with spaces"
	operatorRegex := regexp.MustCompile(`(\w+):("([^"]+)"|([^\s]+))`)

	// Find all operator matches
	matches := operatorRegex.FindAllStringSubmatch(query, -1)

	// Track positions of operators to extract free text
	matchPositions := operatorRegex.FindAllStringIndex(query, -1)

	// Process each operator match
	for _, match := range matches {
		operator := strings.ToLower(match[1])
		// Value is either quoted (match[3]) or unquoted (match[4])
		value := match[3]
		if value == "" {
			value = match[4]
		}

		switch operator {
		case "from":
			result.From = append(result.From, value)
		case "to":
			result.To = append(result.To, value)
		case "subject":
			result.Subject = append(result.Subject, value)
		case "has":
			result.Has = append(result.Has, value)
		case "is":
			result.Is = append(result.Is, value)
		case "after":
			result.After = value
		case "before":
			result.Before = value
		case "in":
			result.In = append(result.In, value)
		}
	}

	// Extract free text (parts not covered by operators)
	if len(matchPositions) == 0 {
		result.Text = strings.TrimSpace(query)
	} else {
		var freeTextParts []string
		lastEnd := 0

		for _, pos := range matchPositions {
			if pos[0] > lastEnd {
				part := strings.TrimSpace(query[lastEnd:pos[0]])
				if part != "" {
					freeTextParts = append(freeTextParts, part)
				}
			}
			lastEnd = pos[1]
		}

		// Check for text after last operator
		if lastEnd < len(query) {
			part := strings.TrimSpace(query[lastEnd:])
			if part != "" {
				freeTextParts = append(freeTextParts, part)
			}
		}

		result.Text = strings.Join(freeTextParts, " ")
	}

	return result
}

// SearchChangedMsg is sent when the search query changes.
type SearchChangedMsg struct {
	Query *SearchQuery
}

// SearchSubmitMsg is sent when the user submits the search (presses Enter).
type SearchSubmitMsg struct {
	Query *SearchQuery
}

// SearchCancelMsg is sent when the user cancels the search (presses Esc).
type SearchCancelMsg struct{}

// HighlightMatches highlights matching text in the given string based on the search query.
// It highlights free text matches in the subject/body and from: matches in sender fields.
func (q *SearchQuery) HighlightMatches(text string, field string, highlightStyle lipgloss.Style) string {
	if q.IsEmpty() || text == "" {
		return text
	}

	// Get search terms based on field type
	var searchTerms []string

	switch field {
	case "from":
		searchTerms = append(searchTerms, q.From...)
	case "to":
		searchTerms = append(searchTerms, q.To...)
	case "subject":
		searchTerms = append(searchTerms, q.Subject...)
	}

	// Always include free text for all fields
	if q.Text != "" {
		// Split free text into words for better matching
		words := strings.Fields(q.Text)
		searchTerms = append(searchTerms, words...)
	}

	if len(searchTerms) == 0 {
		return text
	}

	// Build a regex pattern for all search terms (case insensitive)
	var patterns []string
	for _, term := range searchTerms {
		// Escape regex special characters
		escaped := regexp.QuoteMeta(term)
		patterns = append(patterns, escaped)
	}

	pattern := "(?i)(" + strings.Join(patterns, "|") + ")"
	re, err := regexp.Compile(pattern)
	if err != nil {
		return text
	}

	// Replace matches with highlighted version
	result := re.ReplaceAllStringFunc(text, func(match string) string {
		return highlightStyle.Render(match)
	})

	return result
}
