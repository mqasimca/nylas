package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/rivo/tview"
)

// ComposeMode indicates the type of compose action.
type ComposeMode int

const (
	ComposeModeNew ComposeMode = iota
	ComposeModeReply
	ComposeModeReplyAll
	ComposeModeForward
)

// ComposeView provides an email compose form.
type ComposeView struct {
	*tview.Flex
	app          *App
	form         *tview.Form
	bodyInput    *tview.TextArea
	toInput      *tview.InputField
	ccInput      *tview.InputField
	subjectInput *tview.InputField
	mode         ComposeMode
	replyToMsg   *domain.Message
	onSent       func()
	onCancel     func()
}

// NewComposeView creates a new compose email view.
func NewComposeView(app *App, mode ComposeMode, replyTo *domain.Message) *ComposeView {
	c := &ComposeView{
		Flex:       tview.NewFlex(),
		app:        app,
		mode:       mode,
		replyToMsg: replyTo,
	}

	c.SetDirection(tview.FlexRow)
	c.SetBackgroundColor(app.styles.BgColor)
	c.SetBorder(true)
	c.SetBorderColor(app.styles.FocusColor)

	// Set title based on mode
	switch mode {
	case ComposeModeReply:
		c.SetTitle(" Reply ")
	case ComposeModeReplyAll:
		c.SetTitle(" Reply All ")
	case ComposeModeForward:
		c.SetTitle(" Forward ")
	default:
		c.SetTitle(" New Message ")
	}
	c.SetTitleColor(app.styles.TitleFg)

	c.buildForm()
	return c
}

func (c *ComposeView) buildForm() {
	// Create form
	c.form = tview.NewForm()
	c.form.SetBackgroundColor(c.app.styles.BgColor)
	c.form.SetFieldBackgroundColor(tcell.ColorDarkSlateGray)
	c.form.SetFieldTextColor(c.app.styles.FgColor)
	c.form.SetLabelColor(c.app.styles.InfoColor)
	c.form.SetButtonBackgroundColor(c.app.styles.TableSelectBg)
	c.form.SetButtonTextColor(c.app.styles.TableSelectFg)
	c.form.SetBorderPadding(1, 1, 2, 2)

	// To field
	c.toInput = tview.NewInputField()
	c.toInput.SetLabel("To: ")
	c.toInput.SetFieldWidth(60)
	c.toInput.SetFieldBackgroundColor(tcell.ColorDarkSlateGray)
	c.toInput.SetFieldTextColor(c.app.styles.FgColor)
	c.toInput.SetLabelColor(c.app.styles.InfoColor)

	// CC field
	c.ccInput = tview.NewInputField()
	c.ccInput.SetLabel("Cc: ")
	c.ccInput.SetFieldWidth(60)
	c.ccInput.SetFieldBackgroundColor(tcell.ColorDarkSlateGray)
	c.ccInput.SetFieldTextColor(c.app.styles.FgColor)
	c.ccInput.SetLabelColor(c.app.styles.InfoColor)

	// Subject field
	c.subjectInput = tview.NewInputField()
	c.subjectInput.SetLabel("Subject: ")
	c.subjectInput.SetFieldWidth(60)
	c.subjectInput.SetFieldBackgroundColor(tcell.ColorDarkSlateGray)
	c.subjectInput.SetFieldTextColor(c.app.styles.FgColor)
	c.subjectInput.SetLabelColor(c.app.styles.InfoColor)

	// Body text area
	c.bodyInput = tview.NewTextArea()
	c.bodyInput.SetBackgroundColor(tcell.ColorDarkSlateGray)
	c.bodyInput.SetTextStyle(tcell.StyleDefault.Foreground(c.app.styles.FgColor).Background(tcell.ColorDarkSlateGray))
	c.bodyInput.SetBorder(true)
	c.bodyInput.SetBorderColor(c.app.styles.BorderColor)
	c.bodyInput.SetTitle(" Message Body (Ctrl+S to send, Esc to cancel) ")
	c.bodyInput.SetTitleColor(c.app.styles.InfoSectionFg)
	c.bodyInput.SetPlaceholder("Type your message here...")

	// Pre-fill fields for reply/forward
	if c.replyToMsg != nil {
		c.prefillForReply()
	}

	// Header section with To, Cc, Subject
	headerForm := tview.NewFlex().SetDirection(tview.FlexRow)
	headerForm.SetBackgroundColor(c.app.styles.BgColor)

	// Build To row
	toRow := tview.NewFlex().SetDirection(tview.FlexColumn)
	toRow.AddItem(c.toInput, 0, 1, true)
	headerForm.AddItem(toRow, 1, 0, true)

	// Build Cc row
	ccRow := tview.NewFlex().SetDirection(tview.FlexColumn)
	ccRow.AddItem(c.ccInput, 0, 1, false)
	headerForm.AddItem(ccRow, 1, 0, false)

	// Build Subject row
	subjectRow := tview.NewFlex().SetDirection(tview.FlexColumn)
	subjectRow.AddItem(c.subjectInput, 0, 1, false)
	headerForm.AddItem(subjectRow, 1, 0, false)

	// Button bar
	buttonBar := tview.NewFlex().SetDirection(tview.FlexColumn)
	buttonBar.SetBackgroundColor(c.app.styles.BgColor)

	sendBtn := tview.NewButton("Send (Ctrl+S)")
	sendBtn.SetBackgroundColor(c.app.styles.SuccessColor)
	sendBtn.SetLabelColor(tcell.ColorBlack)
	sendBtn.SetSelectedFunc(c.send)

	cancelBtn := tview.NewButton("Cancel (Esc)")
	cancelBtn.SetBackgroundColor(c.app.styles.BorderColor)
	cancelBtn.SetLabelColor(c.app.styles.FgColor)
	cancelBtn.SetSelectedFunc(c.cancel)

	spacer := tview.NewBox().SetBackgroundColor(c.app.styles.BgColor)
	buttonBar.AddItem(spacer, 0, 1, false)
	buttonBar.AddItem(sendBtn, 16, 0, false)
	buttonBar.AddItem(spacer, 2, 0, false)
	buttonBar.AddItem(cancelBtn, 16, 0, false)
	buttonBar.AddItem(spacer, 0, 1, false)

	// Layout
	c.AddItem(headerForm, 4, 0, true)
	c.AddItem(c.bodyInput, 0, 1, false)
	c.AddItem(buttonBar, 1, 0, false)

	// Set up key handling
	c.SetInputCapture(c.handleKey)
}

func (c *ComposeView) prefillForReply() {
	msg := c.replyToMsg

	// Set To field
	if c.mode == ComposeModeReply || c.mode == ComposeModeReplyAll {
		// Reply to the sender
		if len(msg.From) > 0 {
			c.toInput.SetText(msg.From[0].Email)
		}

		// For Reply All, add other recipients to CC
		if c.mode == ComposeModeReplyAll {
			var ccEmails []string
			for _, to := range msg.To {
				// Don't include the current user
				if to.Email != c.app.config.Email {
					ccEmails = append(ccEmails, to.Email)
				}
			}
			for _, cc := range msg.Cc {
				if cc.Email != c.app.config.Email {
					ccEmails = append(ccEmails, cc.Email)
				}
			}
			if len(ccEmails) > 0 {
				c.ccInput.SetText(strings.Join(ccEmails, ", "))
			}
		}
	}

	// Set subject
	subject := msg.Subject
	if c.mode == ComposeModeReply || c.mode == ComposeModeReplyAll {
		if !strings.HasPrefix(strings.ToLower(subject), "re:") {
			subject = "Re: " + subject
		}
	} else if c.mode == ComposeModeForward {
		if !strings.HasPrefix(strings.ToLower(subject), "fwd:") {
			subject = "Fwd: " + subject
		}
	}
	c.subjectInput.SetText(subject)

	// Set body with quoted original message (Gmail-style format)
	var body strings.Builder
	body.WriteString("\n\n")

	// Gmail-style attribution line
	if len(msg.From) > 0 {
		from := msg.From[0]
		if from.Name != "" {
			body.WriteString(fmt.Sprintf("On %s %s <%s> wrote:\n",
				msg.Date.Format("Mon, Jan 2, 2006 at 3:04 PM"),
				from.Name,
				from.Email))
		} else {
			body.WriteString(fmt.Sprintf("On %s %s wrote:\n",
				msg.Date.Format("Mon, Jan 2, 2006 at 3:04 PM"),
				from.Email))
		}
	}

	// Quote the original message body with > prefix on each line
	originalBody := msg.Body
	if originalBody == "" {
		originalBody = msg.Snippet
	}
	// Strip HTML if present
	originalBody = stripHTMLForQuote(originalBody)

	// Add > prefix to each line
	lines := strings.Split(originalBody, "\n")
	for _, line := range lines {
		body.WriteString("> ")
		body.WriteString(line)
		body.WriteString("\n")
	}

	c.bodyInput.SetText(body.String(), false) // cursor at beginning for top-posting
}

func formatParticipants(participants []domain.EmailParticipant) string {
	var parts []string
	for _, p := range participants {
		if p.Name != "" {
			parts = append(parts, fmt.Sprintf("%s <%s>", p.Name, p.Email))
		} else {
			parts = append(parts, p.Email)
		}
	}
	return strings.Join(parts, ", ")
}

func (c *ComposeView) handleKey(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEscape:
		c.cancel()
		return nil

	case tcell.KeyCtrlS:
		c.send()
		return nil

	case tcell.KeyTab:
		// Cycle focus between fields
		c.cycleFocus()
		return nil

	case tcell.KeyBacktab:
		// Cycle focus backwards
		c.cycleFocusBackward()
		return nil
	}

	return event
}

func (c *ComposeView) cycleFocus() {
	focused := c.app.GetFocus()

	switch focused {
	case c.toInput:
		c.app.SetFocus(c.ccInput)
	case c.ccInput:
		c.app.SetFocus(c.subjectInput)
	case c.subjectInput:
		c.app.SetFocus(c.bodyInput)
	case c.bodyInput:
		c.app.SetFocus(c.toInput)
	default:
		c.app.SetFocus(c.toInput)
	}
}

func (c *ComposeView) cycleFocusBackward() {
	focused := c.app.GetFocus()

	switch focused {
	case c.toInput:
		c.app.SetFocus(c.bodyInput)
	case c.ccInput:
		c.app.SetFocus(c.toInput)
	case c.subjectInput:
		c.app.SetFocus(c.ccInput)
	case c.bodyInput:
		c.app.SetFocus(c.subjectInput)
	default:
		c.app.SetFocus(c.toInput)
	}
}

// SetOnSent sets the callback for when an email is sent successfully.
func (c *ComposeView) SetOnSent(handler func()) {
	c.onSent = handler
}

// SetOnCancel sets the callback for when compose is cancelled.
func (c *ComposeView) SetOnCancel(handler func()) {
	c.onCancel = handler
}

// Focus sets focus to the To field.
func (c *ComposeView) Focus(delegate func(p tview.Primitive)) {
	c.app.SetFocus(c.toInput)
}

func (c *ComposeView) send() {
	// Validate fields
	to := strings.TrimSpace(c.toInput.GetText())
	if to == "" {
		c.app.Flash(FlashError, "To field is required")
		return
	}

	subject := strings.TrimSpace(c.subjectInput.GetText())
	body := c.bodyInput.GetText()

	// Convert plain text body to HTML for proper rendering in email clients
	htmlBody := convertToHTML(body)

	// Parse recipients
	toRecipients := parseRecipients(to)
	if len(toRecipients) == 0 {
		c.app.Flash(FlashError, "Invalid recipient email")
		return
	}

	// Parse CC
	var ccRecipients []domain.EmailParticipant
	cc := strings.TrimSpace(c.ccInput.GetText())
	if cc != "" {
		ccRecipients = parseRecipients(cc)
	}

	// Build request
	req := &domain.SendMessageRequest{
		To:      toRecipients,
		Cc:      ccRecipients,
		Subject: subject,
		Body:    htmlBody,
	}

	// If replying, include the reply-to message ID for threading
	if c.replyToMsg != nil && (c.mode == ComposeModeReply || c.mode == ComposeModeReplyAll) {
		req.ReplyToMsgID = c.replyToMsg.ID
	}

	// Send asynchronously
	c.app.Flash(FlashInfo, "Sending message...")

	go func() {
		// Email send operations should complete within 30 seconds
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_, err := c.app.config.Client.SendMessage(ctx, c.app.config.GrantID, req)
		if err != nil {
			c.app.QueueUpdateDraw(func() {
				c.app.Flash(FlashError, "Failed to send: %v", err)
			})
			return
		}

		c.app.QueueUpdateDraw(func() {
			c.app.Flash(FlashInfo, "Message sent successfully!")
			if c.onSent != nil {
				c.onSent()
			}
		})
	}()
}

func (c *ComposeView) cancel() {
	if c.onCancel != nil {
		c.onCancel()
	}
}

// parseRecipients parses a comma-separated list of email addresses.
func parseRecipients(input string) []domain.EmailParticipant {
	var recipients []domain.EmailParticipant

	parts := strings.Split(input, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Parse "Name <email>" format
		var name, email string
		if idx := strings.Index(part, "<"); idx != -1 {
			name = strings.TrimSpace(part[:idx])
			end := strings.Index(part, ">")
			if end > idx {
				email = strings.TrimSpace(part[idx+1 : end])
			}
		} else {
			email = part
		}

		if email != "" && strings.Contains(email, "@") {
			recipients = append(recipients, domain.EmailParticipant{
				Name:  name,
				Email: email,
			})
		}
	}

	return recipients
}

// stripHTMLForQuote removes HTML tags for quoting in replies.
func stripHTMLForQuote(s string) string {
	return stripHTMLForTUI(s)
}

// convertToHTML converts plain text to HTML for proper email rendering.
func convertToHTML(text string) string {
	// Escape HTML special characters
	escaped := strings.ReplaceAll(text, "&", "&amp;")
	escaped = strings.ReplaceAll(escaped, "<", "&lt;")
	escaped = strings.ReplaceAll(escaped, ">", "&gt;")

	// Convert newlines to <br> tags
	escaped = strings.ReplaceAll(escaped, "\n", "<br>\n")

	// Wrap in basic HTML structure
	return fmt.Sprintf(`<div style="font-family: Arial, sans-serif; font-size: 14px;">%s</div>`, escaped)
}
