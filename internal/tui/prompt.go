package tui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// PromptMode represents the type of prompt.
type PromptMode int

const (
	PromptCommand PromptMode = iota
	PromptFilter
)

// Prompt handles command/filter input like k9s.
type Prompt struct {
	*tview.InputField
	styles    *Styles
	mode      PromptMode
	onCommand func(string)
	onFilter  func(string)
}

// NewPrompt creates a new prompt component.
func NewPrompt(styles *Styles, onCommand, onFilter func(string)) *Prompt {
	p := &Prompt{
		InputField: tview.NewInputField(),
		styles:     styles,
		onCommand:  onCommand,
		onFilter:   onFilter,
	}

	// k9s style prompt colors
	p.SetBackgroundColor(styles.BgColor)
	p.SetFieldBackgroundColor(styles.BgColor)
	p.SetFieldTextColor(styles.PromptFg)
	p.SetBorderPadding(0, 0, 1, 0)

	return p
}

// Activate activates the prompt in the given mode.
func (p *Prompt) Activate(mode PromptMode) {
	p.mode = mode
	p.SetText("")

	// k9s style: dodgerblue for command, filter indicator
	cmdColor := p.styles.Hex(p.styles.BorderColor)
	filterColor := p.styles.Hex(p.styles.FilterColor)

	if mode == PromptCommand {
		p.SetLabel(fmt.Sprintf("[%s::b]:[-::-]", cmdColor))
		p.SetLabelColor(p.styles.BorderColor)
	} else {
		p.SetLabel(fmt.Sprintf("[%s::b]/[-::-]", filterColor))
		p.SetLabelColor(p.styles.FilterColor)
	}
}

// HandleKey processes key events for the prompt.
func (p *Prompt) HandleKey(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEnter:
		text := p.GetText()
		if p.mode == PromptCommand && p.onCommand != nil {
			p.onCommand(text)
		} else if p.mode == PromptFilter && p.onFilter != nil {
			p.onFilter(text)
		}
		return nil

	case tcell.KeyEscape:
		if p.mode == PromptCommand && p.onCommand != nil {
			p.onCommand("")
		} else if p.mode == PromptFilter && p.onFilter != nil {
			p.onFilter("")
		}
		return nil
	}

	return event
}
