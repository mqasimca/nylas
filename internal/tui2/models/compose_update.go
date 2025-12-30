package models

import (
	"context"
	"crypto/sha256"
	"fmt"
	"regexp"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/mqasimca/nylas/internal/domain"
)

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
