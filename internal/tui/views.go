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

// ResourceView interface for all views.
type ResourceView interface {
	Name() string
	Title() string
	Primitive() tview.Primitive
	Hints() []Hint
	Load()
	Refresh()
	Filter(string)
	HandleKey(*tcell.EventKey) *tcell.EventKey
}

// BaseTableView provides common table view functionality.
type BaseTableView struct {
	app    *App
	table  *Table
	name   string
	title  string
	hints  []Hint
	filter string
}

func newBaseTableView(app *App, name, title string) *BaseTableView {
	return &BaseTableView{
		app:   app,
		table: NewTable(app.styles),
		name:  name,
		title: title,
	}
}

func (v *BaseTableView) Name() string               { return v.name }
func (v *BaseTableView) Title() string              { return v.title }
func (v *BaseTableView) Primitive() tview.Primitive { return v.table }
func (v *BaseTableView) Hints() []Hint              { return v.hints }
func (v *BaseTableView) Filter(f string)            { v.filter = f }

func (v *BaseTableView) HandleKey(event *tcell.EventKey) *tcell.EventKey {
	return event // Let table handle navigation
}

// ============================================================================
// Dashboard View
// ============================================================================

// DashboardView shows an overview.
type DashboardView struct {
	app   *App
	view  *tview.TextView
	name  string
	title string
}

// NewDashboardView creates a new dashboard view.
func NewDashboardView(app *App) *DashboardView {
	v := &DashboardView{
		app:   app,
		view:  tview.NewTextView(),
		name:  "dashboard",
		title: "Dashboard",
	}

	v.view.SetDynamicColors(true)
	v.view.SetBackgroundColor(app.styles.BgColor)
	v.view.SetBorderPadding(1, 1, 2, 2)

	return v
}

func (v *DashboardView) Name() string               { return v.name }
func (v *DashboardView) Title() string              { return v.title }
func (v *DashboardView) Primitive() tview.Primitive { return v.view }
func (v *DashboardView) Filter(string)              {}
func (v *DashboardView) Refresh()                   { v.Load() }

func (v *DashboardView) Hints() []Hint {
	return []Hint{
		{Key: ":", Desc: "command"},
		{Key: "?", Desc: "help"},
		{Key: "q", Desc: "quit"},
	}
}

func (v *DashboardView) HandleKey(event *tcell.EventKey) *tcell.EventKey {
	return event
}

func (v *DashboardView) Load() {
	v.view.Clear()

	info := colorToHex(v.app.styles.InfoColor)
	warn := colorToHex(v.app.styles.WarnColor)
	muted := colorToHex(v.app.styles.BorderColor)

	resources := []struct {
		cmd   string
		name  string
		desc  string
	}{
		{":m", "Messages", "Email messages"},
		{":e", "Events", "Calendar events"},
		{":c", "Contacts", "Contacts"},
		{":w", "Webhooks", "Webhooks"},
		{":g", "Grants", "Connected accounts"},
	}

	fmt.Fprintf(v.view, "[%s::b]Quick Navigation[-::-]\n\n", warn)

	for _, r := range resources {
		fmt.Fprintf(v.view, "  [%s]%-6s[-]  [%s]%-12s[-]  [%s]%s[-]\n",
			warn, r.cmd,
			info, r.name,
			muted, r.desc,
		)
	}

	fmt.Fprintf(v.view, "\n[%s]Press : to enter command mode[-]", muted)
}

// ============================================================================
// Messages View
// ============================================================================

// MessagesView displays email messages.
type MessagesView struct {
	*BaseTableView
	messages     []domain.Message
	showingDetail bool
}

// NewMessagesView creates a new messages view.
func NewMessagesView(app *App) *MessagesView {
	v := &MessagesView{
		BaseTableView: newBaseTableView(app, "messages", "Messages"),
	}

	v.hints = []Hint{
		{Key: "enter", Desc: "view"},
		{Key: "s", Desc: "star"},
		{Key: "u", Desc: "unread"},
		{Key: "r", Desc: "refresh"},
	}

	v.table.SetColumns([]Column{
		{Title: "", Width: 3},
		{Title: "FROM", Width: 25},
		{Title: "SUBJECT", Expand: true},
		{Title: "DATE", Width: 12},
	})

	return v
}

func (v *MessagesView) Load() {
	ctx := context.Background()
	msgs, err := v.app.config.Client.GetMessages(ctx, v.app.config.GrantID, 50)
	if err != nil {
		v.app.Flash(FlashError, "Failed to load messages: %v", err)
		return
	}
	v.messages = msgs
	v.render()
}

func (v *MessagesView) Refresh() {
	v.Load()
}

func (v *MessagesView) render() {
	var data [][]string
	var meta []RowMeta

	for _, msg := range v.messages {
		// Apply filter
		if v.filter != "" {
			subject := strings.ToLower(msg.Subject)
			from := ""
			if len(msg.From) > 0 {
				from = strings.ToLower(msg.From[0].Email)
			}
			if !strings.Contains(subject, strings.ToLower(v.filter)) &&
				!strings.Contains(from, strings.ToLower(v.filter)) {
				continue
			}
		}

		from := ""
		if len(msg.From) > 0 {
			from = msg.From[0].Name
			if from == "" {
				from = msg.From[0].Email
			}
		}

		date := formatDate(msg.Date)

		data = append(data, []string{
			"", // Status column
			from,
			msg.Subject,
			date,
		})

		meta = append(meta, RowMeta{
			ID:      msg.ID,
			Data:    &msg,
			Unread:  msg.Unread,
			Starred: msg.Starred,
		})
	}

	v.table.SetData(data, meta)
}

func (v *MessagesView) HandleKey(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEscape:
		// If showing detail, close it and return nil to indicate we handled it
		if v.showingDetail {
			v.closeDetail()
			return nil
		}
		// Otherwise, let app handle the Escape
		return event

	case tcell.KeyEnter:
		// View message detail
		if meta := v.table.SelectedMeta(); meta != nil {
			if msg, ok := meta.Data.(*domain.Message); ok {
				v.showDetail(msg)
			}
		}
		return nil

	case tcell.KeyRune:
		switch event.Rune() {
		case 's':
			v.toggleStar()
			return nil
		case 'u':
			v.markUnread()
			return nil
		}
	}

	return event
}

func (v *MessagesView) toggleStar() {
	meta := v.table.SelectedMeta()
	if meta == nil {
		return
	}

	msg, ok := meta.Data.(*domain.Message)
	if !ok {
		return
	}

	go func() {
		ctx := context.Background()
		newStarred := !msg.Starred
		_, err := v.app.config.Client.UpdateMessage(ctx, v.app.config.GrantID, msg.ID, &domain.UpdateMessageRequest{
			Starred: &newStarred,
		})
		if err != nil {
			v.app.Flash(FlashError, "Failed to update: %v", err)
			return
		}
		v.app.Flash(FlashInfo, "Message starred")
		v.app.QueueUpdateDraw(func() {
			v.Load()
		})
	}()
}

func (v *MessagesView) markUnread() {
	meta := v.table.SelectedMeta()
	if meta == nil {
		return
	}

	msg, ok := meta.Data.(*domain.Message)
	if !ok {
		return
	}

	go func() {
		ctx := context.Background()
		unread := true
		_, err := v.app.config.Client.UpdateMessage(ctx, v.app.config.GrantID, msg.ID, &domain.UpdateMessageRequest{
			Unread: &unread,
		})
		if err != nil {
			v.app.Flash(FlashError, "Failed to update: %v", err)
			return
		}
		v.app.Flash(FlashInfo, "Marked as unread")
		v.app.QueueUpdateDraw(func() {
			v.Load()
		})
	}()
}

func (v *MessagesView) showDetail(msg *domain.Message) {
	detail := tview.NewTextView()
	detail.SetDynamicColors(true)
	detail.SetBackgroundColor(v.app.styles.BgColor)
	detail.SetBorderPadding(1, 1, 2, 2)

	info := colorToHex(v.app.styles.InfoColor)
	muted := colorToHex(v.app.styles.BorderColor)

	from := ""
	if len(msg.From) > 0 {
		from = msg.From[0].String()
	}

	var tos []string
	for _, t := range msg.To {
		tos = append(tos, t.String())
	}

	fmt.Fprintf(detail, "[%s]From:[-] %s\n", muted, from)
	fmt.Fprintf(detail, "[%s]To:[-] %s\n", muted, strings.Join(tos, ", "))
	fmt.Fprintf(detail, "[%s]Subject:[-] [%s::b]%s[-::-]\n", muted, info, msg.Subject)
	fmt.Fprintf(detail, "[%s]Date:[-] %s\n\n", muted, msg.Date.Format("Mon, Jan 2, 2006 3:04 PM"))
	fmt.Fprintf(detail, "────────────────────────────────────────\n\n")
	fmt.Fprintf(detail, "%s\n\n", msg.Snippet)
	fmt.Fprintf(detail, "[%s]Press Esc to go back[-]", muted)

	// Push detail onto the page stack
	v.app.PushDetail("message-detail", detail)
	v.showingDetail = true
}

func (v *MessagesView) closeDetail() {
	v.app.PopDetail()
	v.showingDetail = false
	v.app.SetFocus(v.table)
}

// ============================================================================
// Events View
// ============================================================================

// EventsView displays calendar events.
type EventsView struct {
	*BaseTableView
	events []domain.Event
}

// NewEventsView creates a new events view.
func NewEventsView(app *App) *EventsView {
	v := &EventsView{
		BaseTableView: newBaseTableView(app, "events", "Events"),
	}

	v.hints = []Hint{
		{Key: "enter", Desc: "view"},
		{Key: "r", Desc: "refresh"},
	}

	v.table.SetColumns([]Column{
		{Title: "", Width: 3},
		{Title: "TITLE", Expand: true},
		{Title: "WHEN", Width: 30},
		{Title: "STATUS", Width: 12},
	})

	return v
}

func (v *EventsView) Load() {
	ctx := context.Background()

	// Get calendars first
	calendars, err := v.app.config.Client.GetCalendars(ctx, v.app.config.GrantID)
	if err != nil {
		v.app.Flash(FlashError, "Failed to load calendars: %v", err)
		return
	}

	if len(calendars) == 0 {
		v.table.SetData(nil, nil)
		return
	}

	// Get events from primary calendar
	events, err := v.app.config.Client.GetEvents(ctx, v.app.config.GrantID, calendars[0].ID, nil)
	if err != nil {
		v.app.Flash(FlashError, "Failed to load events: %v", err)
		return
	}

	v.events = events
	v.render()
}

func (v *EventsView) Refresh() { v.Load() }

func (v *EventsView) render() {
	var data [][]string
	var meta []RowMeta

	for _, evt := range v.events {
		when := formatEventTime(evt.When)
		data = append(data, []string{
			"",
			evt.Title,
			when,
			evt.Status,
		})
		meta = append(meta, RowMeta{ID: evt.ID, Data: &evt})
	}

	v.table.SetData(data, meta)
}

// ============================================================================
// Contacts View
// ============================================================================

// ContactsView displays contacts.
type ContactsView struct {
	*BaseTableView
	contacts []domain.Contact
}

// NewContactsView creates a new contacts view.
func NewContactsView(app *App) *ContactsView {
	v := &ContactsView{
		BaseTableView: newBaseTableView(app, "contacts", "Contacts"),
	}

	v.hints = []Hint{
		{Key: "enter", Desc: "view"},
		{Key: "r", Desc: "refresh"},
	}

	v.table.SetColumns([]Column{
		{Title: "", Width: 3},
		{Title: "NAME", Width: 30},
		{Title: "EMAIL", Expand: true},
		{Title: "COMPANY", Width: 25},
	})

	return v
}

func (v *ContactsView) Load() {
	ctx := context.Background()
	contacts, err := v.app.config.Client.GetContacts(ctx, v.app.config.GrantID, nil)
	if err != nil {
		v.app.Flash(FlashError, "Failed to load contacts: %v", err)
		return
	}
	v.contacts = contacts
	v.render()
}

func (v *ContactsView) Refresh() { v.Load() }

func (v *ContactsView) render() {
	var data [][]string
	var meta []RowMeta

	for _, c := range v.contacts {
		email := ""
		if len(c.Emails) > 0 {
			email = c.Emails[0].Email
		}

		name := c.GivenName
		if c.Surname != "" {
			name += " " + c.Surname
		}

		data = append(data, []string{
			"",
			name,
			email,
			c.CompanyName,
		})
		meta = append(meta, RowMeta{ID: c.ID, Data: &c})
	}

	v.table.SetData(data, meta)
}

// ============================================================================
// Webhooks View
// ============================================================================

// WebhooksView displays webhooks.
type WebhooksView struct {
	*BaseTableView
	webhooks []domain.Webhook
}

// NewWebhooksView creates a new webhooks view.
func NewWebhooksView(app *App) *WebhooksView {
	v := &WebhooksView{
		BaseTableView: newBaseTableView(app, "webhooks", "Webhooks"),
	}

	v.hints = []Hint{
		{Key: "enter", Desc: "view"},
		{Key: "r", Desc: "refresh"},
	}

	v.table.SetColumns([]Column{
		{Title: "", Width: 3},
		{Title: "TRIGGERS", Width: 30},
		{Title: "URL", Expand: true},
		{Title: "STATUS", Width: 12},
	})

	return v
}

func (v *WebhooksView) Load() {
	ctx := context.Background()
	webhooks, err := v.app.config.Client.ListWebhooks(ctx)
	if err != nil {
		v.app.Flash(FlashError, "Failed to load webhooks: %v", err)
		return
	}
	v.webhooks = webhooks
	v.render()
}

func (v *WebhooksView) Refresh() { v.Load() }

func (v *WebhooksView) render() {
	var data [][]string
	var meta []RowMeta

	for _, wh := range v.webhooks {
		triggers := strings.Join(wh.TriggerTypes, ", ")
		data = append(data, []string{
			"",
			triggers,
			wh.WebhookURL,
			wh.Status,
		})
		meta = append(meta, RowMeta{
			ID:    wh.ID,
			Data:  &wh,
			Error: wh.Status != "active",
		})
	}

	v.table.SetData(data, meta)
}

// ============================================================================
// Grants View
// ============================================================================

// GrantsView displays grants.
type GrantsView struct {
	*BaseTableView
	grants []domain.Grant
}

// NewGrantsView creates a new grants view.
func NewGrantsView(app *App) *GrantsView {
	v := &GrantsView{
		BaseTableView: newBaseTableView(app, "grants", "Grants"),
	}

	v.hints = []Hint{
		{Key: "enter", Desc: "view"},
		{Key: "r", Desc: "refresh"},
	}

	v.table.SetColumns([]Column{
		{Title: "", Width: 3},
		{Title: "EMAIL", Width: 35},
		{Title: "PROVIDER", Width: 15},
		{Title: "GRANT ID", Expand: true},
	})

	return v
}

func (v *GrantsView) Load() {
	ctx := context.Background()
	grants, err := v.app.config.Client.ListGrants(ctx)
	if err != nil {
		v.app.Flash(FlashError, "Failed to load grants: %v", err)
		return
	}
	v.grants = grants
	v.render()
}

func (v *GrantsView) Refresh() { v.Load() }

func (v *GrantsView) render() {
	var data [][]string
	var meta []RowMeta

	for _, g := range v.grants {
		data = append(data, []string{
			"",
			g.Email,
			string(g.Provider),
			g.ID,
		})
		meta = append(meta, RowMeta{ID: g.ID, Data: &g})
	}

	v.table.SetData(data, meta)
}

// ============================================================================
// Help View
// ============================================================================

// NewHelpView creates a help overlay.
func NewHelpView(styles *Styles) *tview.TextView {
	help := tview.NewTextView()
	help.SetDynamicColors(true)
	help.SetBackgroundColor(styles.BgColor)
	help.SetBorder(true)
	help.SetBorderColor(styles.InfoColor)
	help.SetTitle(" Help ")
	help.SetTitleColor(styles.InfoColor)

	info := colorToHex(styles.InfoColor)
	warn := colorToHex(styles.WarnColor)
	muted := colorToHex(styles.BorderColor)

	fmt.Fprintf(help, "\n")
	fmt.Fprintf(help, "  [%s::b]Navigation[-::-]\n", warn)
	fmt.Fprintf(help, "  [%s]↑/k[-]      Move up\n", info)
	fmt.Fprintf(help, "  [%s]↓/j[-]      Move down\n", info)
	fmt.Fprintf(help, "  [%s]g/Home[-]   Go to top\n", info)
	fmt.Fprintf(help, "  [%s]G/End[-]    Go to bottom\n", info)
	fmt.Fprintf(help, "  [%s]PgUp[-]     Page up\n", info)
	fmt.Fprintf(help, "  [%s]PgDn[-]     Page down\n", info)
	fmt.Fprintf(help, "  [%s]Esc[-]      Go back\n", info)
	fmt.Fprintf(help, "\n")
	fmt.Fprintf(help, "  [%s::b]Commands[-::-]\n", warn)
	fmt.Fprintf(help, "  [%s]:[-]        Enter command mode\n", info)
	fmt.Fprintf(help, "  [%s]/[-]        Filter\n", info)
	fmt.Fprintf(help, "  [%s]r[-]        Refresh\n", info)
	fmt.Fprintf(help, "  [%s]q[-]        Quit\n", info)
	fmt.Fprintf(help, "\n")
	fmt.Fprintf(help, "  [%s::b]Resources[-::-]\n", warn)
	fmt.Fprintf(help, "  [%s]:m[-]       Messages\n", info)
	fmt.Fprintf(help, "  [%s]:e[-]       Events\n", info)
	fmt.Fprintf(help, "  [%s]:c[-]       Contacts\n", info)
	fmt.Fprintf(help, "  [%s]:w[-]       Webhooks\n", info)
	fmt.Fprintf(help, "  [%s]:g[-]       Grants\n", info)
	fmt.Fprintf(help, "\n")
	fmt.Fprintf(help, "  [%s]Press any key to close[-]\n", muted)

	return help
}

// ============================================================================
// Helpers
// ============================================================================

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

func formatEventTime(when domain.EventWhen) string {
	if when.StartTime > 0 {
		start := time.Unix(when.StartTime, 0)
		end := time.Unix(when.EndTime, 0)
		return fmt.Sprintf("%s - %s", start.Format("Jan 2 3:04 PM"), end.Format("3:04 PM"))
	}
	if when.Date != "" {
		return when.Date
	}
	return ""
}
