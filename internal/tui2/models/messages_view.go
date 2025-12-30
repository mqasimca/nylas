package models

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m *MessageList) View() tea.View {
	// Show advanced search dialog if open
	if m.searchMode == SearchModeAdvanced && m.searchDialog != nil {
		return tea.NewView(m.searchDialog.View())
	}

	if m.err != nil {
		return tea.NewView(m.theme.Error_.Render(fmt.Sprintf("Error: %v\n\nPress 'q' to go back", m.err)))
	}

	// Build header
	header := m.theme.Title.Render("Messages") + " " +
		m.theme.Subtitle.Render(fmt.Sprintf("(%s)", m.global.Email))

	// Show loading spinner if loading
	if m.loading {
		header += " " + m.spinner.View()
	}

	// Show search status if active
	if m.searchMode == SearchModeActive && !m.searchQuery.IsEmpty() {
		searchInfo := fmt.Sprintf(" [Searching: %s - %d results]", m.search.Value(), len(m.threads))
		searchStyle := lipgloss.NewStyle().Foreground(m.theme.Secondary).Italic(true)
		header += searchStyle.Render(searchInfo)
	}

	// Build search bar (only shown when in search mode)
	var searchBar string
	if m.searchMode != SearchModeOff {
		searchBar = m.search.ViewInline() + "\n"
	}

	// Build help text
	var help string
	switch m.searchMode {
	case SearchModeInput:
		help = m.theme.Help.Render("Enter: search  Esc: cancel  ?: advanced  | from: to: subject: is:unread has:attachment")
	case SearchModeActive:
		help = m.theme.Help.Render("/: search  ?: advanced  Esc: clear  c: compose  Tab: switch pane  Ctrl+R: refresh")
	default:
		help = m.theme.Help.Render("/: search  ?: advanced  c: compose  r: reply  a: reply all  f: forward  Tab: pane  Ctrl+R: refresh  esc: back")
	}

	// Build layout
	layoutView := m.layout.View()

	// Join all sections with single newlines to maximize space
	return tea.NewView(header + "\n" + searchBar + layoutView + "\n" + help)
}

// fetchMessages fetches the message list.
