package models

import (
	"context"
	"crypto/sha256"
	"fmt"
	"regexp"
	"strings"
	"time"

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/tui2/state"
	"github.com/mqasimca/nylas/internal/tui2/styles"
)

// ComposeMode represents the mode of composition
type ComposeMode int

const (
	ComposeModeNew ComposeMode = iota
	ComposeModeReply
	ComposeModeReplyAll
	ComposeModeForward
	ComposeModeDraft
)

// ComposeData contains the data needed to initialize the compose screen
type ComposeData struct {
	Mode    ComposeMode
	Message *domain.Message
	Draft   *domain.Draft
}

// SaveStatus represents the draft save status
type SaveStatus int

const (
	SaveStatusNone SaveStatus = iota
	SaveStatusSaving
	SaveStatusSaved
	SaveStatusError
	SaveStatusUnsaved
)

// Compose is the email composition screen
type Compose struct {
	global *state.GlobalState
	theme  *styles.Theme

	// Compose context
	mode       ComposeMode
	replyToMsg *domain.Message
	draftID    string

	// Form fields (Bubbles components)
	toInput      textinput.Model
	ccInput      textinput.Model
	bccInput     textinput.Model
	subjectInput textinput.Model
	bodyInput    textarea.Model

	// Focus management
	focusIndex int // 0=to, 1=cc, 2=bcc, 3=subject, 4=body

	// State
	sending          bool
	savingDraft      bool
	validationErrors map[string]string
	showCc           bool
	showBcc          bool

	// Autosave
	lastSavedHash    string
	isDirty          bool
	lastSaveTime     time.Time
	saveStatus       SaveStatus
	autosaveEnabled  bool
	autosaveInterval time.Duration

	// Window dimensions
	width  int
	height int

	// Cursor repositioning (workaround for bubbles v0.21.0 viewport bug)
	needsCursorReposition bool
	hasReceivedWindowSize bool
}

// AutosaveTickMsg is sent by the autosave timer
type AutosaveTickMsg time.Time

// DraftSavedMsg is sent when a draft save completes
type DraftSavedMsg struct {
	draftID string
	err     error
}

// MessageSentMsg is sent when a message is successfully sent
type MessageSentMsg struct {
	message *domain.Message
}

// SendErrorMsg is sent when sending fails
type SendErrorMsg struct {
	err error
}

// NewCompose creates a new compose screen
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
func (c *Compose) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.width = msg.Width
		c.height = msg.Height
		c.updateFieldSizes()
		c.hasReceivedWindowSize = true

		// Reposition cursor to top if needed (workaround for bubbles v0.21.0 viewport bug)
		// We do this after WindowSizeMsg so the viewport is properly initialized
		if c.needsCursorReposition {
			c.needsCursorReposition = false
			// Move cursor to top by moving up from current position
			currentLine := c.bodyInput.Line()
			for i := 0; i < currentLine; i++ {
				c.bodyInput.CursorUp()
			}
			c.bodyInput.CursorStart()
		}
		return c, nil

	case tea.KeyMsg:
		// Handle special keys using msg.String() for modifier combos
		// Note: key.Text is empty for modifier combos in v2
		key := msg.Key()
		keyStr := msg.String()

		// Handle Esc key
		if key.Code == tea.KeyEsc {
			if c.isDirty {
				c.global.SetStatus("Unsaved changes - save with Ctrl+S or discard with Ctrl+Q", 1)
				return c, nil
			}
			return c, func() tea.Msg { return BackMsg{} }
		}

		// Handle Tab key (with Shift modifier check)
		switch keyStr {
		case "tab":
			c.cycleFocusForward()
			return c, nil
		case "shift+tab":
			c.cycleFocusBackward()
			return c, nil
		}

		// Handle Ctrl+key combinations using msg.String()
		switch keyStr {
		case "ctrl+a":
			// Send message with Ctrl+A
			if c.sending {
				return c, nil
			}
			return c, c.send()

		case "ctrl+c", "ctrl+q":
			// Ctrl+C or Ctrl+Q - quit (let app.go handle it)
			if c.isDirty {
				c.global.SetStatus("Unsaved changes - press again to quit", 1)
				return c, nil
			}
			// Don't consume the message - let it bubble up to app.go
			// Fall through to updateFocusedField

		case "ctrl+s":
			// Manual save
			if c.savingDraft {
				return c, nil
			}
			return c, c.saveDraft()

		case "ctrl+t":
			// Toggle Cc
			c.showCc = !c.showCc
			return c, nil

		case "ctrl+b":
			// Toggle Bcc
			c.showBcc = !c.showBcc
			return c, nil
		}

		// Update focused field
		cmd := c.updateFocusedField(msg)
		cmds = append(cmds, cmd)

		// Check for dirty state
		currentHash := c.computeContentHash()
		if currentHash != c.lastSavedHash {
			c.isDirty = true
			if c.saveStatus != SaveStatusSaving {
				c.saveStatus = SaveStatusUnsaved
			}
		}

	case AutosaveTickMsg:
		// Save if dirty
		if c.isDirty && !c.savingDraft {
			cmds = append(cmds, c.performAutosave())
		}

		// Schedule next tick
		cmds = append(cmds, c.startAutosaveTimer())

	case DraftSavedMsg:
		c.savingDraft = false

		if msg.err != nil {
			c.saveStatus = SaveStatusError
			c.global.SetStatus(fmt.Sprintf("Draft save failed: %v", msg.err), 1)
			return c, nil
		}

		// Validate draft was actually created/updated
		if msg.draftID == "" {
			c.saveStatus = SaveStatusError
			c.global.SetStatus("Draft save failed: no draft ID returned", 1)
			return c, nil
		}

		// Update draft ID on first save
		if c.draftID == "" {
			c.draftID = msg.draftID
		}

		// Mark as saved
		c.lastSavedHash = c.computeContentHash()
		c.isDirty = false
		c.saveStatus = SaveStatusSaved
		c.lastSaveTime = time.Now()

		// Clear any error status
		c.global.SetStatus("", 0)

		return c, nil

	case MessageSentMsg:
		c.sending = false
		c.global.SetStatus("Message sent successfully!", 0)

		// Delete draft if one was created
		var cmds []tea.Cmd
		if c.draftID != "" {
			cmds = append(cmds, c.deleteDraft())
		}

		// Navigate back after brief delay
		cmds = append(cmds, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return BackMsg{}
		}))

		return c, tea.Batch(cmds...)

	case SendErrorMsg:
		c.sending = false
		c.saveStatus = SaveStatusNone
		errStr := msg.err.Error()
		// Truncate long error messages
		if len(errStr) > 100 {
			errStr = errStr[:97] + "..."
		}
		c.global.SetStatus(fmt.Sprintf("Send failed: %s", errStr), 1)
		return c, nil
	}

	return c, tea.Batch(cmds...)
}

// updateFocusedField updates the currently focused field
func (c *Compose) updateFocusedField(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd

	switch c.focusIndex {
	case 0: // To
		c.toInput, cmd = c.toInput.Update(msg)
	case 1: // Cc
		if c.showCc {
			c.ccInput, cmd = c.ccInput.Update(msg)
		}
	case 2: // Bcc
		if c.showBcc {
			c.bccInput, cmd = c.bccInput.Update(msg)
		}
	case 3: // Subject
		c.subjectInput, cmd = c.subjectInput.Update(msg)
	case 4: // Body
		c.bodyInput, cmd = c.bodyInput.Update(msg)
	}

	return cmd
}

// cycleFocusForward moves focus to the next field
func (c *Compose) cycleFocusForward() {
	// Blur current field
	c.blurAll()

	// Find next focusable field
	for i := 0; i < 5; i++ {
		c.focusIndex = (c.focusIndex + 1) % 5

		// Skip hidden fields
		if c.focusIndex == 1 && !c.showCc {
			continue
		}
		if c.focusIndex == 2 && !c.showBcc {
			continue
		}

		break
	}

	// Focus new field
	c.focusField(c.focusIndex)
}

// cycleFocusBackward moves focus to the previous field
func (c *Compose) cycleFocusBackward() {
	// Blur current field
	c.blurAll()

	// Find previous focusable field
	for i := 0; i < 5; i++ {
		c.focusIndex = (c.focusIndex - 1 + 5) % 5

		// Skip hidden fields
		if c.focusIndex == 1 && !c.showCc {
			continue
		}
		if c.focusIndex == 2 && !c.showBcc {
			continue
		}

		break
	}

	// Focus new field
	c.focusField(c.focusIndex)
}

// blurAll blurs all fields
func (c *Compose) blurAll() {
	c.toInput.Blur()
	c.ccInput.Blur()
	c.bccInput.Blur()
	c.subjectInput.Blur()
	c.bodyInput.Blur()
}

// focusField focuses a specific field
func (c *Compose) focusField(index int) {
	switch index {
	case 0:
		c.toInput.Focus()
	case 1:
		c.ccInput.Focus()
	case 2:
		c.bccInput.Focus()
	case 3:
		c.subjectInput.Focus()
	case 4:
		c.bodyInput.Focus()
	}
}

// validate validates the form
func (c *Compose) validate() bool {
	c.validationErrors = make(map[string]string)

	// Validate To field
	toValue := strings.TrimSpace(c.toInput.Value())
	if toValue == "" {
		c.validationErrors["to"] = "At least one recipient is required"
	} else {
		// Parse and validate email addresses
		recipients := c.parseRecipients(toValue)
		if len(recipients) == 0 {
			c.validationErrors["to"] = "Invalid email address format"
		}
	}

	// Validate Subject (optional but warn if empty)
	subjectValue := strings.TrimSpace(c.subjectInput.Value())
	if subjectValue == "" {
		c.validationErrors["subject"] = "Subject is empty (optional)"
	}

	// Body is optional

	return len(c.validationErrors) == 0 || (len(c.validationErrors) == 1 && c.validationErrors["subject"] != "")
}

// send sends the message
func (c *Compose) send() tea.Cmd {
	// Validate first
	if !c.validate() {
		// Show specific validation errors
		if errMsg, ok := c.validationErrors["to"]; ok {
			c.global.SetStatus(fmt.Sprintf("Cannot send: %s", errMsg), 1)
		} else {
			c.global.SetStatus("Please fix validation errors", 1)
		}
		return nil
	}

	c.sending = true
	c.saveStatus = SaveStatusNone // Clear draft status while sending
	c.global.SetStatus("Sending message...", 0)

	return func() tea.Msg {
		c.global.RateLimiter.Wait()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Build request
		req := c.buildSendRequest()

		// Send message
		message, err := c.global.Client.SendMessage(ctx, c.global.GrantID, req)
		if err != nil {
			return SendErrorMsg{err: err}
		}

		return MessageSentMsg{message: message}
	}
}

// saveDraft manually saves a draft
func (c *Compose) saveDraft() tea.Cmd {
	c.savingDraft = true
	c.saveStatus = SaveStatusSaving
	c.global.SetStatus("Saving draft...", 0)

	return c.performAutosave()
}

// performAutosave performs the autosave operation
func (c *Compose) performAutosave() tea.Cmd {
	return func() tea.Msg {
		c.global.RateLimiter.Wait()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		req := c.buildDraftRequest()

		var draft *domain.Draft
		var err error

		if c.draftID == "" {
			draft, err = c.global.Client.CreateDraft(ctx, c.global.GrantID, req)
		} else {
			draft, err = c.global.Client.UpdateDraft(ctx, c.global.GrantID, c.draftID, req)
		}

		if err != nil {
			return DraftSavedMsg{err: err}
		}

		return DraftSavedMsg{draftID: draft.ID, err: nil}
	}
}

// deleteDraft deletes the saved draft after message is sent
func (c *Compose) deleteDraft() tea.Cmd {
	return func() tea.Msg {
		c.global.RateLimiter.Wait()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Delete the draft
		err := c.global.Client.DeleteDraft(ctx, c.global.GrantID, c.draftID)
		if err != nil {
			// Log error but don't fail - message was already sent
			// Just return nil, user won't see this error
			return nil
		}

		return nil
	}
}

// buildSendRequest builds a SendMessageRequest
func (c *Compose) buildSendRequest() *domain.SendMessageRequest {
	req := &domain.SendMessageRequest{
		Subject: strings.TrimSpace(c.subjectInput.Value()),
		Body:    c.bodyInput.Value(),
	}

	// Parse To
	toValue := strings.TrimSpace(c.toInput.Value())
	if toValue != "" {
		req.To = c.parseRecipients(toValue)
	}

	// Parse Cc
	if c.showCc {
		ccValue := strings.TrimSpace(c.ccInput.Value())
		if ccValue != "" {
			req.Cc = c.parseRecipients(ccValue)
		}
	}

	// Parse Bcc
	if c.showBcc {
		bccValue := strings.TrimSpace(c.bccInput.Value())
		if bccValue != "" {
			req.Bcc = c.parseRecipients(bccValue)
		}
	}

	// Set reply-to if replying
	if c.mode == ComposeModeReply || c.mode == ComposeModeReplyAll {
		if c.replyToMsg != nil {
			req.ReplyToMsgID = c.replyToMsg.ID
		}
	}

	return req
}

// buildDraftRequest builds a CreateDraftRequest
func (c *Compose) buildDraftRequest() *domain.CreateDraftRequest {
	req := &domain.CreateDraftRequest{
		Subject: strings.TrimSpace(c.subjectInput.Value()),
		Body:    c.bodyInput.Value(),
	}

	// Parse To
	toValue := strings.TrimSpace(c.toInput.Value())
	if toValue != "" {
		req.To = c.parseRecipients(toValue)
	}

	// Parse Cc
	if c.showCc {
		ccValue := strings.TrimSpace(c.ccInput.Value())
		if ccValue != "" {
			req.Cc = c.parseRecipients(ccValue)
		}
	}

	// Parse Bcc
	if c.showBcc {
		bccValue := strings.TrimSpace(c.bccInput.Value())
		if bccValue != "" {
			req.Bcc = c.parseRecipients(bccValue)
		}
	}

	// NOTE: ReplyToMsgID is only set in send request, not draft request
	// Drafts don't support reply-to-message-id in Nylas API

	return req
}

// parseRecipients parses a comma-separated list of email addresses
func (c *Compose) parseRecipients(input string) []domain.EmailParticipant {
	var recipients []domain.EmailParticipant

	// Split by comma
	parts := strings.Split(input, ",")

	// Regex to match "Name <email>" or just "email"
	re := regexp.MustCompile(`^([^<>]+)<([^<>]+)>$|^([^<>,\s]+)$`)

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		matches := re.FindStringSubmatch(part)
		if matches == nil {
			continue
		}

		if matches[1] != "" && matches[2] != "" {
			// "Name <email>" format
			recipients = append(recipients, domain.EmailParticipant{
				Name:  strings.TrimSpace(matches[1]),
				Email: strings.TrimSpace(matches[2]),
			})
		} else if matches[3] != "" {
			// Just "email" format
			recipients = append(recipients, domain.EmailParticipant{
				Email: strings.TrimSpace(matches[3]),
			})
		}
	}

	return recipients
}

// computeContentHash computes a hash of the current form content
func (c *Compose) computeContentHash() string {
	content := fmt.Sprintf("%s|%s|%s|%s|%s",
		c.toInput.Value(),
		c.ccInput.Value(),
		c.bccInput.Value(),
		c.subjectInput.Value(),
		c.bodyInput.Value(),
	)

	hash := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", hash)
}

// updateFieldSizes updates field sizes based on window dimensions
func (c *Compose) updateFieldSizes() {
	maxWidth := c.width - 20
	if maxWidth < 40 {
		maxWidth = 40
	}

	c.toInput.SetWidth(maxWidth)
	c.ccInput.SetWidth(maxWidth)
	c.bccInput.SetWidth(maxWidth)
	c.subjectInput.SetWidth(maxWidth)
	c.bodyInput.SetWidth(maxWidth)

	bodyHeight := c.height - 15
	if bodyHeight < 5 {
		bodyHeight = 5
	}
	c.bodyInput.SetHeight(bodyHeight)
}

// View renders the compose screen
func (c *Compose) View() tea.View {
	var sb strings.Builder

	// Title
	title := "Compose New Message"
	switch c.mode {
	case ComposeModeReply:
		title = "Reply"
	case ComposeModeReplyAll:
		title = "Reply All"
	case ComposeModeForward:
		title = "Forward"
	case ComposeModeDraft:
		title = "Edit Draft"
	}

	sb.WriteString(c.theme.Title.Render(title))
	sb.WriteString("\n\n")

	// To field
	sb.WriteString(c.renderField("To:", c.toInput.View(), c.validationErrors["to"], c.focusIndex == 0))

	// Cc field (if shown)
	if c.showCc {
		sb.WriteString(c.renderField("Cc:", c.ccInput.View(), "", c.focusIndex == 1))
	}

	// Bcc field (if shown)
	if c.showBcc {
		sb.WriteString(c.renderField("Bcc:", c.bccInput.View(), "", c.focusIndex == 2))
	}

	// Subject field
	sb.WriteString(c.renderField("Subject:", c.subjectInput.View(), c.validationErrors["subject"], c.focusIndex == 3))

	// Body field
	sb.WriteString(c.renderBodyField())

	// Status line
	sb.WriteString("\n")
	sb.WriteString(c.renderStatusLine())

	// Help text
	sb.WriteString("\n")
	sb.WriteString(c.renderHelp())

	return tea.NewView(sb.String())
}

// renderField renders a form field
func (c *Compose) renderField(label, value, errMsg string, focused bool) string {
	var sb strings.Builder

	// Label
	labelStyle := c.theme.Help
	if focused {
		labelStyle = lipgloss.NewStyle().Foreground(c.theme.Primary).Bold(true)
	}
	sb.WriteString(labelStyle.Render(fmt.Sprintf("%-10s", label)))

	// Value
	sb.WriteString(value)

	// Error message
	if errMsg != "" {
		sb.WriteString("  ")
		sb.WriteString(c.theme.Error_.Render(errMsg))
	}

	sb.WriteString("\n")
	return sb.String()
}

// renderBodyField renders the body field
func (c *Compose) renderBodyField() string {
	var sb strings.Builder

	// Label
	labelStyle := c.theme.Help
	if c.focusIndex == 4 {
		labelStyle = lipgloss.NewStyle().Foreground(c.theme.Primary).Bold(true)
	}
	sb.WriteString(labelStyle.Render("Body:"))
	sb.WriteString("\n")

	// Body text area
	sb.WriteString(c.bodyInput.View())

	return sb.String()
}

// renderStatusLine renders the status line with save status
func (c *Compose) renderStatusLine() string {
	var parts []string

	// Send status
	if c.sending {
		primaryStyle := lipgloss.NewStyle().Foreground(c.theme.Primary)
		parts = append(parts, primaryStyle.Render("⏳ Sending..."))
	}

	// Save status
	saveStatus := c.renderSaveStatus()
	if saveStatus != "" {
		parts = append(parts, saveStatus)
	}

	return strings.Join(parts, "  ")
}

// renderSaveStatus renders the save status indicator
func (c *Compose) renderSaveStatus() string {
	primaryStyle := lipgloss.NewStyle().Foreground(c.theme.Primary)
	warningStyle := lipgloss.NewStyle().Foreground(c.theme.Warning)
	successStyle := lipgloss.NewStyle().Foreground(c.theme.Success)

	switch c.saveStatus {
	case SaveStatusSaving:
		return primaryStyle.Render("⏳ Saving draft...")
	case SaveStatusSaved:
		elapsed := time.Since(c.lastSaveTime)
		if c.isDirty {
			return warningStyle.Render("● Unsaved changes")
		}
		return successStyle.Render(fmt.Sprintf("✓ Draft saved %s", formatElapsed(elapsed)))
	case SaveStatusError:
		return c.theme.Error_.Render("✗ Save failed (Ctrl+S to retry)")
	case SaveStatusUnsaved:
		return warningStyle.Render("● Unsaved changes")
	default:
		return ""
	}
}

// formatElapsed formats an elapsed duration
func formatElapsed(d time.Duration) string {
	if d < time.Minute {
		return "just now"
	}
	if d < time.Hour {
		mins := int(d.Minutes())
		if mins == 1 {
			return "1 min ago"
		}
		return fmt.Sprintf("%d mins ago", mins)
	}
	hours := int(d.Hours())
	if hours == 1 {
		return "1 hour ago"
	}
	return fmt.Sprintf("%d hours ago", hours)
}

// renderHelp renders the help text
func (c *Compose) renderHelp() string {
	helps := []string{
		"Tab: next field",
		"Ctrl+A: send",
		"Ctrl+S: save draft",
	}

	if !c.showCc {
		helps = append(helps, "Ctrl+T: show Cc")
	}
	if !c.showBcc {
		helps = append(helps, "Ctrl+B: show Bcc")
	}

	helps = append(helps, "Esc: cancel")

	return c.theme.Help.Render(strings.Join(helps, "  "))
}
