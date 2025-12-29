package tui

import (
	"context"
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/rivo/tview"
)

// InboundView displays inbound inboxes and their messages.
type InboundView struct {
	app           *App
	layout        *tview.Flex
	inboxList     *Table
	messageList   *Table
	inboxes       []domain.InboundInbox
	messages      []domain.InboundMessage
	selectedInbox *domain.InboundInbox
	focusedPanel  int // 0 = inbox list, 1 = message list
	showingDetail bool
	name          string
	title         string
}

// NewInboundView creates a new inbound view.
func NewInboundView(app *App) *InboundView {
	v := &InboundView{
		app:   app,
		name:  "inbound",
		title: "Inbound",
	}

	// Create inbox list table
	v.inboxList = NewTable(app.styles)
	v.inboxList.SetColumns([]Column{
		{Title: "", Width: 3},
		{Title: "EMAIL", Expand: true},
		{Title: "STATUS", Width: 10},
		{Title: "CREATED", Width: 15},
	})

	// Create message list table
	v.messageList = NewTable(app.styles)
	v.messageList.SetColumns([]Column{
		{Title: "", Width: 3},
		{Title: "FROM", Width: 25},
		{Title: "SUBJECT", Expand: true},
		{Title: "DATE", Width: 12},
	})

	// Create split layout: Inboxes (top) | Messages (bottom)
	v.layout = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(v.inboxList, 10, 0, true).
		AddItem(v.messageList, 0, 1, false)

	// Set up selection callback for inbox list
	v.inboxList.SetOnSelect(func(meta *RowMeta) {
		if inbox, ok := meta.Data.(*domain.InboundInbox); ok {
			v.selectedInbox = inbox
			v.loadMessages(inbox.ID)
		}
	})

	// Set up double-click to view message
	v.messageList.SetOnDoubleClick(func(meta *RowMeta) {
		if msg, ok := meta.Data.(*domain.InboundMessage); ok {
			v.showMessageDetail(msg)
		}
	})

	return v
}

func (v *InboundView) Name() string               { return v.name }
func (v *InboundView) Title() string              { return v.title }
func (v *InboundView) Primitive() tview.Primitive { return v.layout }
func (v *InboundView) Filter(string)              {}

func (v *InboundView) Hints() []Hint {
	return []Hint{
		{Key: "enter", Desc: "select/view"},
		{Key: "Tab", Desc: "switch panel"},
		{Key: "r", Desc: "refresh"},
	}
}

func (v *InboundView) Load() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Load inbound inboxes
	inboxes, err := v.app.config.Client.ListInboundInboxes(ctx)
	if err != nil {
		v.app.Flash(FlashError, "Failed to load inboxes: %v", err)
		return
	}

	v.inboxes = inboxes
	v.renderInboxes()

	if len(inboxes) == 0 {
		v.app.Flash(FlashInfo, "No inbound inboxes found. Create one with: nylas inbound create <name>")
		return
	}

	// Select first inbox and load its messages
	v.selectedInbox = &inboxes[0]
	v.loadMessages(inboxes[0].ID)

	v.app.Flash(FlashInfo, "Found %d inbound inbox(es)", len(inboxes))
}

func (v *InboundView) Refresh() {
	v.Load()
}

func (v *InboundView) renderInboxes() {
	var data [][]string
	var meta []RowMeta

	for _, inbox := range v.inboxes {
		status := "active"
		if inbox.GrantStatus != "valid" {
			status = inbox.GrantStatus
		}

		created := formatDate(inbox.CreatedAt.Time)

		// Mark selected inbox
		marker := ""
		if v.selectedInbox != nil && inbox.ID == v.selectedInbox.ID {
			marker = ">"
		}

		data = append(data, []string{
			marker,
			inbox.Email,
			status,
			created,
		})

		// Create a copy for closure
		i := inbox
		meta = append(meta, RowMeta{
			ID:    inbox.ID,
			Data:  &i,
			Error: inbox.GrantStatus != "valid",
		})
	}

	v.inboxList.SetData(data, meta)
}

func (v *InboundView) loadMessages(inboxID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	messages, err := v.app.config.Client.GetInboundMessages(ctx, inboxID, &domain.MessageQueryParams{Limit: 50})
	if err != nil {
		v.app.Flash(FlashError, "Failed to load messages: %v", err)
		return
	}

	v.messages = messages
	v.renderMessages()
}

func (v *InboundView) renderMessages() {
	var data [][]string
	var meta []RowMeta

	for _, msg := range v.messages {
		from := ""
		if len(msg.From) > 0 {
			from = msg.From[0].Name
			if from == "" {
				from = msg.From[0].Email
			}
		}

		date := formatDate(msg.Date)

		data = append(data, []string{
			"",
			from,
			msg.Subject,
			date,
		})

		// Create a copy for closure
		m := msg
		meta = append(meta, RowMeta{
			ID:      msg.ID,
			Data:    &m,
			Unread:  msg.Unread,
			Starred: msg.Starred,
		})
	}

	v.messageList.SetData(data, meta)
}

func (v *InboundView) HandleKey(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEscape:
		if v.showingDetail {
			v.closeDetail()
			return nil
		}
		return event

	case tcell.KeyTab:
		// Switch focus between inbox list and message list
		v.focusedPanel = (v.focusedPanel + 1) % 2
		if v.focusedPanel == 0 {
			v.app.SetFocus(v.inboxList)
		} else {
			v.app.SetFocus(v.messageList)
		}
		return nil

	case tcell.KeyEnter:
		if v.focusedPanel == 0 {
			// Select inbox and load messages
			if meta := v.inboxList.SelectedMeta(); meta != nil {
				if inbox, ok := meta.Data.(*domain.InboundInbox); ok {
					v.selectedInbox = inbox
					v.renderInboxes() // Update marker
					v.loadMessages(inbox.ID)
					// Switch to message panel
					v.focusedPanel = 1
					v.app.SetFocus(v.messageList)
				}
			}
		} else {
			// View message detail
			if meta := v.messageList.SelectedMeta(); meta != nil {
				if msg, ok := meta.Data.(*domain.InboundMessage); ok {
					v.showMessageDetail(msg)
				}
			}
		}
		return nil
	}

	return event
}

func (v *InboundView) showMessageDetail(msg *domain.InboundMessage) {
	detail := tview.NewTextView()
	detail.SetDynamicColors(true)
	detail.SetBackgroundColor(v.app.styles.BgColor)
	detail.SetBorderPadding(1, 1, 2, 2)
	detail.SetScrollable(true)

	// Colors
	title := colorToHex(v.app.styles.TitleFg)
	key := colorToHex(v.app.styles.FgColor)
	value := colorToHex(v.app.styles.InfoSectionFg)
	muted := colorToHex(v.app.styles.BorderColor)

	// Format sender
	from := ""
	if len(msg.From) > 0 {
		from = msg.From[0].String()
	}

	// Format recipients
	var to []string
	for _, t := range msg.To {
		to = append(to, t.String())
	}

	_, _ = fmt.Fprintf(detail, "[%s::b]%s[-::-]\n", title, msg.Subject)
	_, _ = fmt.Fprintf(detail, "[%s]────────────────────────────────────────[-]\n", muted)
	_, _ = fmt.Fprintf(detail, "[%s]From:[-] [%s]%s[-]\n", key, value, from)
	if len(to) > 0 {
		_, _ = fmt.Fprintf(detail, "[%s]To:[-] [%s]%s[-]\n", key, value, strings.Join(to, ", "))
	}
	_, _ = fmt.Fprintf(detail, "[%s]Date:[-] [%s]%s[-]\n", key, value, msg.Date.Format("Mon, Jan 2, 2006 3:04 PM"))
	_, _ = fmt.Fprintf(detail, "[%s]ID:[-] [%s]%s[-]\n", key, value, msg.ID)
	_, _ = fmt.Fprintf(detail, "[%s]────────────────────────────────────────[-]\n\n", muted)

	// Body
	body := msg.Body
	if body == "" {
		body = msg.Snippet
	}
	body = stripHTMLForTUI(body)
	_, _ = fmt.Fprintf(detail, "[%s]%s[-]\n\n", value, tview.Escape(body))

	_, _ = fmt.Fprintf(detail, "[%s]Press Esc to go back[-]", muted)

	v.app.PushDetail("inbound-message-detail", detail)
	v.showingDetail = true
}

func (v *InboundView) closeDetail() {
	v.app.PopDetail()
	v.showingDetail = false
	if v.focusedPanel == 0 {
		v.app.SetFocus(v.inboxList)
	} else {
		v.app.SetFocus(v.messageList)
	}
}

// formatDate formats a time for display in the UI.
func formatDate(t time.Time) string {
	now := time.Now()
	if t.Year() == now.Year() && t.YearDay() == now.YearDay() {
		return t.Format("3:04 PM")
	}
	if t.Year() == now.Year() {
		return t.Format("Jan 2")
	}
	return t.Format("Jan 2, 06")
}

// formatFileSize formats a file size in bytes to a human-readable string.
func formatFileSize(size int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case size >= GB:
		return fmt.Sprintf("%.1f GB", float64(size)/float64(GB))
	case size >= MB:
		return fmt.Sprintf("%.1f MB", float64(size)/float64(MB))
	case size >= KB:
		return fmt.Sprintf("%.1f KB", float64(size)/float64(KB))
	default:
		return fmt.Sprintf("%d B", size)
	}
}

// stripHTMLForTUI removes HTML tags from a string for terminal display.
func stripHTMLForTUI(s string) string {
	// Remove style and script tags and their contents
	s = removeTagWithContent(s, "style")
	s = removeTagWithContent(s, "script")
	s = removeTagWithContent(s, "head")

	// Replace block-level elements with newlines before stripping tags
	blockTags := []string{"br", "p", "div", "tr", "li", "h1", "h2", "h3", "h4", "h5", "h6"}
	for _, tag := range blockTags {
		s = strings.ReplaceAll(s, "<"+tag+">", "\n")
		s = strings.ReplaceAll(s, "<"+tag+"/>", "\n")
		s = strings.ReplaceAll(s, "<"+tag+" />", "\n")
		s = strings.ReplaceAll(s, "</"+tag+">", "\n")
		s = strings.ReplaceAll(s, "<"+strings.ToUpper(tag)+">", "\n")
		s = strings.ReplaceAll(s, "</"+strings.ToUpper(tag)+">", "\n")
	}

	// Strip remaining HTML tags
	var result strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			result.WriteRune(r)
		}
	}

	// Decode HTML entities
	text := html.UnescapeString(result.String())

	// Clean up whitespace
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	// Collapse multiple spaces
	for strings.Contains(text, "  ") {
		text = strings.ReplaceAll(text, "  ", " ")
	}

	// Collapse multiple newlines
	for strings.Contains(text, "\n\n\n") {
		text = strings.ReplaceAll(text, "\n\n\n", "\n\n")
	}

	// Trim spaces from each line
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}
	text = strings.Join(lines, "\n")

	return strings.TrimSpace(text)
}

// removeTagWithContent removes an HTML tag and all its content.
func removeTagWithContent(s, tag string) string {
	result := s
	for {
		lower := strings.ToLower(result)
		startIdx := strings.Index(lower, "<"+tag)
		if startIdx == -1 {
			break
		}
		endTag := "</" + tag + ">"
		endIdx := strings.Index(lower[startIdx:], endTag)
		if endIdx == -1 {
			closeIdx := strings.Index(result[startIdx:], ">")
			if closeIdx == -1 {
				break
			}
			result = result[:startIdx] + result[startIdx+closeIdx+1:]
		} else {
			result = result[:startIdx] + result[startIdx+endIdx+len(endTag):]
		}
	}
	return result
}
