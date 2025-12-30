package models

import (
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/tui2/state"
	"github.com/mqasimca/nylas/internal/tui2/styles"
)

func NewCompose(global *state.GlobalState, data ComposeData) *Compose {
	theme := styles.GetTheme(global.Theme)

	c := &Compose{
		global:           global,
		theme:            theme,
		mode:             data.Mode,
		replyToMsg:       data.Message,
		validationErrors: make(map[string]string),
		autosaveEnabled:  true,
		autosaveInterval: 30 * time.Second,
		showCc:           false,
		showBcc:          false,
	}

	// Initialize form fields
	c.toInput = textinput.New()
	c.toInput.Placeholder = "recipient@example.com"
	c.toInput.CharLimit = 500
	c.toInput.SetWidth(80)

	c.ccInput = textinput.New()
	c.ccInput.Placeholder = "cc@example.com"
	c.ccInput.CharLimit = 500
	c.ccInput.SetWidth(80)

	c.bccInput = textinput.New()
	c.bccInput.Placeholder = "bcc@example.com"
	c.bccInput.CharLimit = 500
	c.bccInput.SetWidth(80)

	c.subjectInput = textinput.New()
	c.subjectInput.Placeholder = "Subject"
	c.subjectInput.CharLimit = 500
	c.subjectInput.SetWidth(80)

	c.bodyInput = textarea.New()
	c.bodyInput.Placeholder = "Type your message here..."
	c.bodyInput.CharLimit = 50000
	c.bodyInput.SetWidth(80)
	c.bodyInput.SetHeight(10)

	// Pre-fill based on mode
	c.prefillFields(data)

	// Focus on first field
	c.focusIndex = 0
	c.toInput.Focus()

	return c
}

// prefillFields pre-fills form fields based on compose mode
func (c *Compose) prefillFields(data ComposeData) {
	switch c.mode {
	case ComposeModeReply, ComposeModeReplyAll:
		if data.Message != nil {
			// Set subject with "Re:" prefix
			subject := data.Message.Subject
			if !strings.HasPrefix(strings.ToLower(subject), "re:") {
				subject = "Re: " + subject
			}
			c.subjectInput.SetValue(subject)

			// Set To field to original sender
			if len(data.Message.From) > 0 {
				c.toInput.SetValue(c.formatParticipant(data.Message.From[0]))
			}

			// For reply all, add Cc recipients
			if c.mode == ComposeModeReplyAll {
				if len(data.Message.To) > 0 || len(data.Message.Cc) > 0 {
					c.showCc = true
					var ccList []string
					for _, p := range data.Message.To {
						ccList = append(ccList, c.formatParticipant(p))
					}
					for _, p := range data.Message.Cc {
						ccList = append(ccList, c.formatParticipant(p))
					}
					c.ccInput.SetValue(strings.Join(ccList, ", "))
				}
			}

			// Set body with quoted original
			c.bodyInput.SetValue(c.buildQuotedBody(data.Message))
			// Mark that we need to reposition cursor to top (will be done in Update after WindowSizeMsg)
			c.needsCursorReposition = true
		}

	case ComposeModeForward:
		if data.Message != nil {
			// Set subject with "Fwd:" prefix
			subject := data.Message.Subject
			if !strings.HasPrefix(strings.ToLower(subject), "fwd:") {
				subject = "Fwd: " + subject
			}
			c.subjectInput.SetValue(subject)

			// Set body with forwarded content
			c.bodyInput.SetValue(c.buildForwardedBody(data.Message))
			// Mark that we need to reposition cursor to top (will be done in Update after WindowSizeMsg)
			c.needsCursorReposition = true
		}

	case ComposeModeDraft:
		if data.Draft != nil {
			c.draftID = data.Draft.ID
			// Pre-fill from draft
			if len(data.Draft.To) > 0 {
				var toList []string
				for _, p := range data.Draft.To {
					toList = append(toList, c.formatParticipant(p))
				}
				c.toInput.SetValue(strings.Join(toList, ", "))
			}
			if len(data.Draft.Cc) > 0 {
				c.showCc = true
				var ccList []string
				for _, p := range data.Draft.Cc {
					ccList = append(ccList, c.formatParticipant(p))
				}
				c.ccInput.SetValue(strings.Join(ccList, ", "))
			}
			if len(data.Draft.Bcc) > 0 {
				c.showBcc = true
				var bccList []string
				for _, p := range data.Draft.Bcc {
					bccList = append(bccList, c.formatParticipant(p))
				}
				c.bccInput.SetValue(strings.Join(bccList, ", "))
			}
			c.subjectInput.SetValue(data.Draft.Subject)
			c.bodyInput.SetValue(data.Draft.Body)
		}
	}

	// Compute initial hash
	c.lastSavedHash = c.computeContentHash()
}

// formatParticipant formats an email participant for display
func (c *Compose) formatParticipant(p domain.EmailParticipant) string {
	if p.Name != "" {
		return fmt.Sprintf("%s <%s>", p.Name, p.Email)
	}
	return p.Email
}

// buildQuotedBody builds a quoted reply body
// Cursor starts at top, quoted content below (like Gmail)
func (c *Compose) buildQuotedBody(msg *domain.Message) string {
	var sb strings.Builder

	// Leave space at top for user to type reply
	sb.WriteString("\n\n\n")

	// Add quote header
	sender := "Unknown"
	if len(msg.From) > 0 {
		sender = msg.From[0].Email
		if msg.From[0].Name != "" {
			sender = msg.From[0].Name
		}
	}

	dateStr := msg.Date.Format("Mon, Jan 2, 2006 at 3:04 PM")
	sb.WriteString(fmt.Sprintf("On %s, %s wrote:\n", dateStr, sender))

	// Quote the body
	body := msg.Body
	if body == "" {
		body = msg.Snippet
	}

	// Strip HTML tags if present
	body = c.stripHTMLBasic(body)

	// Strip email signatures (lines after "--")
	body = c.stripSignature(body)

	// Strip existing quote markers and attribution lines to prevent nesting
	// This ensures we only quote the NEW content from the last reply
	body = c.stripExistingQuotes(body)

	lines := strings.Split(body, "\n")
	for _, line := range lines {
		sb.WriteString("> ")
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	return sb.String()
}

// buildForwardedBody builds a forwarded message body
func (c *Compose) buildForwardedBody(msg *domain.Message) string {
	var sb strings.Builder

	// Leave space at top for user to add context
	sb.WriteString("\n\n\n")
	sb.WriteString("---------- Forwarded message ---------\n")

	if len(msg.From) > 0 {
		sb.WriteString(fmt.Sprintf("From: %s\n", c.formatParticipant(msg.From[0])))
	}
	sb.WriteString(fmt.Sprintf("Date: %s\n", msg.Date.Format("Mon, Jan 2, 2006 at 3:04 PM")))
	sb.WriteString(fmt.Sprintf("Subject: %s\n", msg.Subject))
	if len(msg.To) > 0 {
		var toList []string
		for _, p := range msg.To {
			toList = append(toList, c.formatParticipant(p))
		}
		sb.WriteString(fmt.Sprintf("To: %s\n", strings.Join(toList, ", ")))
	}
	sb.WriteString("\n")

	body := msg.Body
	if body == "" {
		body = msg.Snippet
	}

	// Strip HTML tags if present
	body = c.stripHTMLBasic(body)

	sb.WriteString(body)

	return sb.String()
}

// stripHTMLBasic performs basic HTML tag stripping
func (c *Compose) stripHTMLBasic(html string) string {
	// Replace common block elements with line breaks
	s := html
	s = strings.ReplaceAll(s, "<br>", "\n")
	s = strings.ReplaceAll(s, "<br/>", "\n")
	s = strings.ReplaceAll(s, "<br />", "\n")
	s = strings.ReplaceAll(s, "</p>", "\n\n")
	s = strings.ReplaceAll(s, "</div>", "\n")

	// Remove all remaining HTML tags
	inTag := false
	var result strings.Builder
	for _, ch := range s {
		if ch == '<' {
			inTag = true
			continue
		}
		if ch == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result.WriteRune(ch)
		}
	}

	cleaned := result.String()

	// Decode common HTML entities
	cleaned = strings.ReplaceAll(cleaned, "&nbsp;", " ")
	cleaned = strings.ReplaceAll(cleaned, "&amp;", "&")
	cleaned = strings.ReplaceAll(cleaned, "&lt;", "<")
	cleaned = strings.ReplaceAll(cleaned, "&gt;", ">")
	cleaned = strings.ReplaceAll(cleaned, "&quot;", "\"")
	cleaned = strings.ReplaceAll(cleaned, "&#39;", "'")

	// Clean up excessive newlines
	cleaned = strings.ReplaceAll(cleaned, "\n\n\n", "\n\n")

	return strings.TrimSpace(cleaned)
}

// stripSignature removes email signatures from message body
// Signatures typically start with "--" on its own line
func (c *Compose) stripSignature(body string) string {
	lines := strings.Split(body, "\n")
	var result []string

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Standard email signature delimiter is "-- " or just "--"
		if trimmed == "--" || trimmed == "-- " {
			// Found signature delimiter, stop here
			// Keep lines before this point
			result = lines[:i]
			break
		}
	}

	// If no signature delimiter found, return original
	if len(result) == 0 {
		return body
	}

	return strings.Join(result, "\n")
}

// stripExistingQuotes removes existing quote markers and attribution lines
// to prevent nested "On ... wrote:" headers when replying to replies
func (c *Compose) stripExistingQuotes(body string) string {
	lines := strings.Split(body, "\n")
	var result []string

	for _, line := range lines {
		// Skip lines that start with quote markers ("> ")
		if strings.HasPrefix(line, ">") {
			continue
		}

		// Skip attribution lines (e.g., "On Dec 27, 2025 at 3:05 PM, John wrote:")
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "On ") && strings.Contains(trimmed, " wrote:") {
			continue
		}

		// Keep other lines
		result = append(result, line)
	}

	return strings.TrimSpace(strings.Join(result, "\n"))
}

// Init initializes the compose screen
func (c *Compose) Init() tea.Cmd {
	var cmds []tea.Cmd
	cmds = append(cmds, textarea.Blink)
	cmds = append(cmds, c.startAutosaveTimer())

	return tea.Batch(cmds...)
}

// startAutosaveTimer starts the autosave timer
func (c *Compose) startAutosaveTimer() tea.Cmd {
	if !c.autosaveEnabled {
		return nil
	}

	return tea.Tick(c.autosaveInterval, func(t time.Time) tea.Msg {
		return AutosaveTickMsg(t)
	})
}

// Update handles messages
