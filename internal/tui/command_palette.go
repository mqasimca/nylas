package tui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	maxSuggestions = 10 // Maximum number of suggestions to show
)

// CommandPalette provides an autocomplete dropdown for command input.
type CommandPalette struct {
	*tview.Flex
	app         *App
	registry    *CommandRegistry
	input       *tview.InputField
	dropdown    *tview.List
	suggestions []Command
	selected    int
	visible     bool
	parentCmd   string       // Parent command for sub-commands (e.g., "folder")
	onExecute   func(string) // Callback when command is executed
	onCancel    func()       // Callback when palette is cancelled
}

// NewCommandPalette creates a new command palette.
func NewCommandPalette(app *App, registry *CommandRegistry, onExecute func(string), onCancel func()) *CommandPalette {
	p := &CommandPalette{
		Flex:        tview.NewFlex(),
		app:         app,
		registry:    registry,
		suggestions: make([]Command, 0),
		onExecute:   onExecute,
		onCancel:    onCancel,
	}

	p.init()
	return p
}

func (p *CommandPalette) init() {
	styles := p.app.styles

	// Create input field
	p.input = tview.NewInputField()
	p.input.SetLabel(":")
	p.input.SetLabelColor(styles.BorderColor)
	p.input.SetFieldBackgroundColor(styles.BgColor)
	p.input.SetFieldTextColor(styles.PromptFg)
	p.input.SetBackgroundColor(styles.BgColor)
	p.input.SetPlaceholder("Type command...")
	p.input.SetPlaceholderTextColor(styles.BorderColor)

	// Handle text changes for autocomplete
	p.input.SetChangedFunc(func(text string) {
		p.updateSuggestions(text)
	})

	// Create dropdown list for suggestions
	p.dropdown = tview.NewList()
	p.dropdown.SetBackgroundColor(styles.BgColor)
	p.dropdown.SetMainTextColor(styles.FgColor)
	p.dropdown.SetSecondaryTextColor(styles.BorderColor)
	p.dropdown.SetSelectedTextColor(styles.BgColor)
	p.dropdown.SetSelectedBackgroundColor(styles.FocusColor)
	p.dropdown.SetHighlightFullLine(true)
	p.dropdown.ShowSecondaryText(true)
	p.dropdown.SetBorder(true)
	p.dropdown.SetBorderColor(styles.BorderColor)

	// Layout: input on top, dropdown below
	p.SetDirection(tview.FlexRow)
	p.AddItem(p.input, 1, 0, true)
	p.AddItem(p.dropdown, 0, 1, false)

	// Set up input capture for navigation
	p.input.SetInputCapture(p.handleInput)

	// Initial suggestions
	p.updateSuggestions("")
}

// Show activates the command palette.
func (p *CommandPalette) Show() {
	p.input.SetText("")
	p.parentCmd = ""
	p.selected = 0
	p.visible = true
	p.updateSuggestions("")
	p.app.SetFocus(p.input)
}

// Hide deactivates the command palette.
func (p *CommandPalette) Hide() {
	p.visible = false
}

// IsVisible returns true if the palette is visible.
func (p *CommandPalette) IsVisible() bool {
	return p.visible
}

func (p *CommandPalette) updateSuggestions(text string) {
	p.dropdown.Clear()
	text = strings.TrimSpace(text)

	// Check if we're typing a sub-command
	parts := strings.SplitN(text, " ", 2)
	if len(parts) > 1 && p.registry.HasSubCommands(parts[0]) {
		// Show sub-commands
		p.parentCmd = parts[0]
		subQuery := strings.TrimSpace(parts[1])
		p.suggestions = p.registry.SearchSubCommands(p.parentCmd, subQuery)
	} else if len(parts) == 1 && strings.HasSuffix(text, " ") {
		// Space after command - check for sub-commands
		cmdName := strings.TrimSpace(parts[0])
		if p.registry.HasSubCommands(cmdName) {
			p.parentCmd = cmdName
			p.suggestions = p.registry.GetSubCommands(cmdName)
		} else {
			p.parentCmd = ""
			p.suggestions = p.registry.Search(text)
		}
	} else {
		// Normal search
		p.parentCmd = ""
		p.suggestions = p.registry.Search(text)
	}

	// Limit suggestions
	if len(p.suggestions) > maxSuggestions {
		p.suggestions = p.suggestions[:maxSuggestions]
	}

	// Populate dropdown
	styles := p.app.styles
	cmdColor := styles.Hex(styles.InfoColor)
	aliasColor := styles.Hex(styles.BorderColor)

	for _, cmd := range p.suggestions {
		// Format: command name with aliases
		mainText := cmd.Name
		if p.parentCmd != "" {
			mainText = p.parentCmd + " " + cmd.Name
		}

		// Secondary text: description and aliases
		secondaryText := cmd.Description
		if len(cmd.Aliases) > 0 {
			secondaryText += fmt.Sprintf(" [%s](:%s)[-]", aliasColor, strings.Join(cmd.Aliases, ", :"))
		}
		_ = cmdColor // For future coloring

		p.dropdown.AddItem(mainText, secondaryText, 0, nil)
	}

	// Select first item
	if len(p.suggestions) > 0 {
		p.selected = 0
		p.dropdown.SetCurrentItem(0)
	}
}

func (p *CommandPalette) handleInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEscape:
		// Cancel and close
		if p.onCancel != nil {
			p.onCancel()
		}
		return nil

	case tcell.KeyEnter:
		// Execute command
		p.executeSelected()
		return nil

	case tcell.KeyTab:
		// Autocomplete with selected suggestion
		p.autocomplete()
		return nil

	case tcell.KeyDown, tcell.KeyCtrlN:
		// Move selection down
		p.moveSelection(1)
		return nil

	case tcell.KeyUp, tcell.KeyCtrlP:
		// Move selection up
		p.moveSelection(-1)
		return nil

	case tcell.KeyCtrlU:
		// Clear input
		p.input.SetText("")
		return nil

	case tcell.KeyBackspace, tcell.KeyBackspace2:
		// If input is empty, close palette
		if p.input.GetText() == "" {
			if p.onCancel != nil {
				p.onCancel()
			}
			return nil
		}
		// Otherwise let default handling work
		return event

	case tcell.KeyRune:
		// Check for space - might trigger sub-command mode
		if event.Rune() == ' ' {
			text := p.input.GetText()
			if p.registry.HasSubCommands(strings.TrimSpace(text)) {
				// Don't consume the space, let it be added
				return event
			}
		}
		return event
	}

	return event
}

func (p *CommandPalette) moveSelection(delta int) {
	if len(p.suggestions) == 0 {
		return
	}

	p.selected += delta
	if p.selected < 0 {
		p.selected = len(p.suggestions) - 1
	} else if p.selected >= len(p.suggestions) {
		p.selected = 0
	}

	p.dropdown.SetCurrentItem(p.selected)
}

func (p *CommandPalette) autocomplete() {
	if p.selected >= 0 && p.selected < len(p.suggestions) {
		cmd := p.suggestions[p.selected]
		if p.parentCmd != "" {
			p.input.SetText(p.parentCmd + " " + cmd.Name)
		} else {
			p.input.SetText(cmd.Name)
		}

		// Check if this command has sub-commands
		fullCmd := p.input.GetText()
		if p.registry.HasSubCommands(fullCmd) {
			p.input.SetText(fullCmd + " ")
			p.updateSuggestions(fullCmd + " ")
		}
	}
}

func (p *CommandPalette) executeSelected() {
	var cmd string

	// If there's selected text in input, use that
	inputText := strings.TrimSpace(p.input.GetText())
	if inputText != "" {
		cmd = inputText
	} else if p.selected >= 0 && p.selected < len(p.suggestions) {
		// Use selected suggestion
		suggestion := p.suggestions[p.selected]
		if p.parentCmd != "" {
			cmd = p.parentCmd + " " + suggestion.Name
		} else {
			cmd = suggestion.Name
		}
	}

	if cmd != "" && p.onExecute != nil {
		p.onExecute(cmd)
	}
}

// Focus sets focus to the input field.
func (p *CommandPalette) Focus(delegate func(tview.Primitive)) {
	delegate(p.input)
}

// GetInput returns the current input text.
func (p *CommandPalette) GetInput() string {
	return p.input.GetText()
}

// SetInput sets the input text.
func (p *CommandPalette) SetInput(text string) {
	p.input.SetText(text)
	p.updateSuggestions(text)
}
