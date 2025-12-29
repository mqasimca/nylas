// Package models provides screen models for the TUI.
package models

import (
	"context"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/tui2/state"
	"github.com/mqasimca/nylas/internal/tui2/styles"
)

// ContactsScreen displays and manages contacts.
type ContactsScreen struct {
	global   *state.GlobalState
	theme    *styles.Theme
	list     list.Model
	contacts []domain.Contact
	loading  bool
	err      error
	width    int
	height   int
}

// contactItem implements list.Item interface.
type contactItem struct {
	contact domain.Contact
}

func (i contactItem) Title() string {
	if i.contact.GivenName != "" || i.contact.Surname != "" {
		return fmt.Sprintf("%s %s", i.contact.GivenName, i.contact.Surname)
	}
	// Fallback to first email
	if len(i.contact.Emails) > 0 {
		return i.contact.Emails[0].Email
	}
	return "Unknown Contact"
}

func (i contactItem) Description() string {
	var parts []string

	// Add primary email
	if len(i.contact.Emails) > 0 {
		parts = append(parts, i.contact.Emails[0].Email)
	}

	// Add company if available
	if i.contact.CompanyName != "" {
		parts = append(parts, i.contact.CompanyName)
	}

	// Add phone if available
	if len(i.contact.PhoneNumbers) > 0 {
		parts = append(parts, i.contact.PhoneNumbers[0].Number)
	}

	if len(parts) == 0 {
		return "No details available"
	}

	return strings.Join(parts, " • ")
}

func (i contactItem) FilterValue() string {
	return i.Title() + " " + i.Description()
}

// contactsLoadedMsg is sent when contacts are loaded.
type contactsLoadedMsg struct {
	contacts []domain.Contact
	err      error
}

// NewContactsScreen creates a new contacts screen.
func NewContactsScreen(global *state.GlobalState) *ContactsScreen {
	theme := styles.GetTheme(global.Theme)

	// Create list with custom styling
	delegate := list.NewDefaultDelegate()

	// Style the delegate
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().
		Foreground(theme.Background).
		Background(theme.Primary).
		Bold(true)

	delegate.Styles.SelectedDesc = lipgloss.NewStyle().
		Foreground(theme.Background).
		Background(theme.Primary)

	delegate.Styles.NormalTitle = lipgloss.NewStyle().
		Foreground(theme.Foreground)

	delegate.Styles.NormalDesc = lipgloss.NewStyle().
		Foreground(theme.Dimmed.GetForeground())

	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = "Contacts"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = theme.Title
	l.Styles.TitleBar = lipgloss.NewStyle().Background(theme.Background)

	return &ContactsScreen{
		global:  global,
		theme:   theme,
		list:    l,
		loading: true,
	}
}

// Init implements tea.Model.
func (c *ContactsScreen) Init() tea.Cmd {
	return c.loadContacts()
}

// loadContacts loads contacts from the API.
func (c *ContactsScreen) loadContacts() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Get contacts from client
		contacts, err := c.global.Client.GetContacts(ctx, c.global.GrantID, nil)
		if err != nil {
			return contactsLoadedMsg{err: err}
		}

		return contactsLoadedMsg{contacts: contacts}
	}
}

// Update implements tea.Model.
func (c *ContactsScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.global.SetWindowSize(msg.Width, msg.Height)
		c.width = msg.Width
		c.height = msg.Height
		c.list.SetSize(msg.Width, msg.Height-6)

	case contactsLoadedMsg:
		c.loading = false
		if msg.err != nil {
			c.err = msg.err
		} else {
			c.contacts = msg.contacts
			// Convert to list items
			items := make([]list.Item, len(msg.contacts))
			for i, contact := range msg.contacts {
				items[i] = contactItem{contact: contact}
			}
			c.list.SetItems(items)
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			return c, func() tea.Msg { return BackMsg{} }
		case "ctrl+c":
			return c, tea.Quit
		case "r":
			// Refresh
			c.loading = true
			return c, c.loadContacts()
		case "enter":
			// View contact details
			if c.list.SelectedItem() != nil {
				item := c.list.SelectedItem().(contactItem)
				return c, c.viewContactDetail(item.contact)
			}
		}
	}

	// Update list
	var cmd tea.Cmd
	c.list, cmd = c.list.Update(msg)
	return c, cmd
}

// viewContactDetail shows contact detail view.
func (c *ContactsScreen) viewContactDetail(contact domain.Contact) tea.Cmd {
	// TODO: Implement contact detail view
	// For now, just show a status message
	c.global.SetStatus(fmt.Sprintf("Viewing contact: %s", contact.GivenName), 0)
	return nil
}

// View implements tea.Model.
func (c *ContactsScreen) View() tea.View {
	if c.loading {
		loadingMsg := lipgloss.NewStyle().
			Foreground(c.theme.Primary).
			Padding(2).
			Render("Loading contacts...")
		return tea.NewView(loadingMsg)
	}

	if c.err != nil {
		errorMsg := lipgloss.NewStyle().
			Foreground(c.theme.Error).
			Padding(2).
			Render(fmt.Sprintf("Error loading contacts: %v", c.err))
		return tea.NewView(errorMsg)
	}

	// Show list
	listView := c.list.View()

	// Add help text
	help := c.theme.Help.Render("↑/↓: Navigate • Enter: View • r: Refresh • /: Search • Esc: Back")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		listView,
		"",
		help,
	)

	return tea.NewView(content)
}
