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
		{Key: "^C", Desc: "quit"},
	}
}

func (v *DashboardView) HandleKey(event *tcell.EventKey) *tcell.EventKey {
	return event
}

func (v *DashboardView) Load() {
	v.view.Clear()

	// k9s style colors
	title := colorToHex(v.app.styles.TitleFg)
	key := colorToHex(v.app.styles.MenuKeyFg)
	desc := colorToHex(v.app.styles.FgColor)
	muted := colorToHex(v.app.styles.BorderColor)

	resources := []struct {
		cmd  string
		name string
		desc string
	}{
		{":m", "Messages", "Email messages"},
		{":e", "Events", "Calendar events"},
		{":c", "Contacts", "Contacts"},
		{":i", "Inbound", "Inbound inboxes"},
		{":w", "Webhooks", "Webhooks"},
		{":ws", "Server", "Webhook server (local)"},
		{":g", "Grants", "Connected accounts"},
	}

	fmt.Fprintf(v.view, "[%s::b]Quick Navigation[-::-]\n\n", title)

	for _, r := range resources {
		fmt.Fprintf(v.view, "  [%s]%-6s[-]  [%s]%-12s[-]  [%s::d]%s[-::-]\n",
			key, r.cmd,
			desc, r.name,
			muted, r.desc,
		)
	}

	fmt.Fprintf(v.view, "\n[%s::d]Press : to enter command mode[-::-]", muted)
}

// ============================================================================
// Messages View (Thread-based)
// ============================================================================

// MessagesView displays email threads (conversations).
type MessagesView struct {
	*BaseTableView
	threads        []domain.Thread
	showingDetail  bool
	currentThread  *domain.Thread
	currentMessage *domain.Message // For reply functionality
}

// NewMessagesView creates a new messages view.
func NewMessagesView(app *App) *MessagesView {
	v := &MessagesView{
		BaseTableView: newBaseTableView(app, "messages", "Messages"),
	}

	v.hints = []Hint{
		{Key: "enter", Desc: "view"},
		{Key: "n", Desc: "compose"},
		{Key: "R", Desc: "reply"},
		{Key: "s", Desc: "star"},
		{Key: "u", Desc: "unread"},
		{Key: "r", Desc: "refresh"},
	}

	v.table.SetColumns([]Column{
		{Title: "", Width: 3},
		{Title: "FROM", Width: 25},
		{Title: "SUBJECT", Expand: true},
		{Title: "#", Width: 3},
		{Title: "DATE", Width: 12},
	})

	// Set up double-click to open thread
	v.table.SetOnDoubleClick(func(meta *RowMeta) {
		if thread, ok := meta.Data.(*domain.Thread); ok {
			v.showDetail(thread)
		}
	})

	return v
}

func (v *MessagesView) Load() {
	ctx := context.Background()
	// Default to INBOX folder to show only inbox threads
	params := &domain.ThreadQueryParams{
		Limit: 50,
		In:    []string{"INBOX"},
	}
	threads, err := v.app.config.Client.GetThreads(ctx, v.app.config.GrantID, params)
	if err != nil {
		v.app.Flash(FlashError, "Failed to load threads: %v", err)
		return
	}
	v.threads = threads
	v.render()
}

func (v *MessagesView) Refresh() {
	v.Load()
}

func (v *MessagesView) render() {
	var data [][]string
	var meta []RowMeta

	for _, thread := range v.threads {
		// Apply filter
		if v.filter != "" {
			subject := strings.ToLower(thread.Subject)
			participants := ""
			if len(thread.Participants) > 0 {
				participants = strings.ToLower(thread.Participants[0].Email)
			}
			if !strings.Contains(subject, strings.ToLower(v.filter)) &&
				!strings.Contains(participants, strings.ToLower(v.filter)) {
				continue
			}
		}

		// Get the primary participant (first one, typically the sender)
		from := ""
		if len(thread.Participants) > 0 {
			from = thread.Participants[0].Name
			if from == "" {
				from = thread.Participants[0].Email
			}
		}

		date := formatDate(thread.LatestMessageRecvDate)
		msgCount := fmt.Sprintf("%d", len(thread.MessageIDs))

		data = append(data, []string{
			"", // Status column
			from,
			thread.Subject,
			msgCount,
			date,
		})

		// Create a copy of thread for the closure
		t := thread
		meta = append(meta, RowMeta{
			ID:      thread.ID,
			Data:    &t,
			Unread:  thread.Unread,
			Starred: thread.Starred,
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
		// View thread detail
		if meta := v.table.SelectedMeta(); meta != nil {
			if thread, ok := meta.Data.(*domain.Thread); ok {
				v.showDetail(thread)
			}
		}
		return nil

	case tcell.KeyRune:
		switch event.Rune() {
		case 'n':
			// New compose
			v.showCompose(ComposeModeNew, nil)
			return nil
		case 'R':
			// Reply to selected or current message
			if v.showingDetail && v.currentMessage != nil {
				v.showCompose(ComposeModeReply, v.currentMessage)
			}
			return nil
		case 'A':
			// Reply All to selected or current message
			if v.showingDetail && v.currentMessage != nil {
				v.showCompose(ComposeModeReplyAll, v.currentMessage)
			}
			return nil
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

	thread, ok := meta.Data.(*domain.Thread)
	if !ok {
		return
	}

	go func() {
		ctx := context.Background()
		newStarred := !thread.Starred
		_, err := v.app.config.Client.UpdateThread(ctx, v.app.config.GrantID, thread.ID, &domain.UpdateMessageRequest{
			Starred: &newStarred,
		})
		if err != nil {
			v.app.Flash(FlashError, "Failed to update: %v", err)
			return
		}
		v.app.Flash(FlashInfo, "Thread starred")
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

	thread, ok := meta.Data.(*domain.Thread)
	if !ok {
		return
	}

	go func() {
		ctx := context.Background()
		unread := true
		_, err := v.app.config.Client.UpdateThread(ctx, v.app.config.GrantID, thread.ID, &domain.UpdateMessageRequest{
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

func (v *MessagesView) showDetail(thread *domain.Thread) {
	v.currentThread = thread

	detail := tview.NewTextView()
	detail.SetDynamicColors(true)
	detail.SetBackgroundColor(v.app.styles.BgColor)
	detail.SetBorderPadding(1, 1, 2, 2)
	detail.SetScrollable(true)

	// k9s style colors
	title := colorToHex(v.app.styles.TitleFg)
	key := colorToHex(v.app.styles.FgColor)
	value := colorToHex(v.app.styles.InfoSectionFg)
	muted := colorToHex(v.app.styles.BorderColor)
	hint := colorToHex(v.app.styles.InfoColor)

	// Format participants
	var participants []string
	for _, p := range thread.Participants {
		participants = append(participants, p.String())
	}

	// Show loading state first
	fmt.Fprintf(detail, "[%s::b]%s[-::-]\n", title, thread.Subject)
	fmt.Fprintf(detail, "[%s]Participants:[-] [%s]%s[-]\n", key, value, strings.Join(participants, ", "))
	fmt.Fprintf(detail, "[%s]Messages:[-] [%s]%d[-]\n\n", key, value, len(thread.MessageIDs))
	fmt.Fprintf(detail, "[%s]────────────────────────────────────────[-]\n\n", muted)
	fmt.Fprintf(detail, "[%s]Loading messages...[-]\n\n", muted)

	// Fetch all messages in the thread asynchronously
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Fetch each message in the thread
		var messages []*domain.Message
		for _, msgID := range thread.MessageIDs {
			msg, err := v.app.config.Client.GetMessage(ctx, v.app.config.GrantID, msgID)
			if err == nil {
				messages = append(messages, msg)
			}
		}

		v.app.QueueUpdateDraw(func() {
			detail.Clear()

			fmt.Fprintf(detail, "[%s::b]%s[-::-]\n", title, thread.Subject)
			fmt.Fprintf(detail, "[%s]Participants:[-] [%s]%s[-]\n", key, value, strings.Join(participants, ", "))
			fmt.Fprintf(detail, "[%s]Messages:[-] [%s]%d[-]\n\n", key, value, len(thread.MessageIDs))

			if len(messages) == 0 {
				fmt.Fprintf(detail, "[%s]────────────────────────────────────────[-]\n\n", muted)
				fmt.Fprintf(detail, "[%s]%s[-]\n\n", value, thread.Snippet)
			} else {
				// Display all messages in chronological order
				for i, msg := range messages {
					fmt.Fprintf(detail, "[%s]════════════════════════════════════════[-]\n", muted)

					from := ""
					if len(msg.From) > 0 {
						from = msg.From[0].String()
					}

					fmt.Fprintf(detail, "[%s]From:[-] [%s]%s[-]\n", key, value, from)
					fmt.Fprintf(detail, "[%s]Date:[-] [%s]%s[-]\n\n", key, value, msg.Date.Format("Mon, Jan 2, 2006 3:04 PM"))

					// Use full body, strip HTML for terminal display
					body := msg.Body
					if body == "" {
						body = msg.Snippet
					}
					body = stripHTMLForTUI(body)
					fmt.Fprintf(detail, "[%s]%s[-]\n\n", value, tview.Escape(body))

					// Store the last message for reply
					if i == len(messages)-1 {
						v.currentMessage = msg
					}
				}
			}

			fmt.Fprintf(detail, "[%s]R[-][%s::d]=reply  [-::-][%s]A[-][%s::d]=reply all  [-::-][%s]Esc[-][%s::d]=back[-::-]", hint, muted, hint, muted, hint, muted)
		})
	}()

	fmt.Fprintf(detail, "[%s]R[-][%s::d]=reply  [-::-][%s]A[-][%s::d]=reply all  [-::-][%s]Esc[-][%s::d]=back[-::-]", hint, muted, hint, muted, hint, muted)

	// Handle key events for reply actions in detail view
	detail.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			v.closeDetail()
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'R':
				if v.currentMessage != nil {
					v.showCompose(ComposeModeReply, v.currentMessage)
				}
				return nil
			case 'A':
				if v.currentMessage != nil {
					v.showCompose(ComposeModeReplyAll, v.currentMessage)
				}
				return nil
			}
		}
		return event
	})

	// Push detail onto the page stack
	v.app.PushDetail("thread-detail", detail)
	v.showingDetail = true
}

func (v *MessagesView) closeDetail() {
	v.app.PopDetail()
	v.showingDetail = false
	v.currentThread = nil
	v.currentMessage = nil
	v.app.SetFocus(v.table)
}

func (v *MessagesView) showCompose(mode ComposeMode, replyTo *domain.Message) {
	compose := NewComposeView(v.app, mode, replyTo)

	compose.SetOnSent(func() {
		v.app.PopDetail()
		// Refresh messages to show the sent message
		go func() {
			v.Load()
			v.app.QueueUpdateDraw(func() {})
		}()
	})

	compose.SetOnCancel(func() {
		v.app.PopDetail()
		if v.showingDetail {
			// Go back to message detail view - just set focus
		} else {
			v.app.SetFocus(v.table)
		}
	})

	v.app.PushDetail("compose", compose)
}

// ============================================================================
// Events View (Google Calendar Style)
// ============================================================================

// EventsView displays a Google Calendar-style calendar view.
type EventsView struct {
	app          *App
	layout       *tview.Flex
	calendar     *CalendarView
	eventsList   *tview.TextView
	events       []domain.Event
	calendars    []domain.Calendar
	name         string
	title        string
	focusedPanel int // 0 = calendar, 1 = events list
}

// NewEventsView creates a new calendar-style events view.
func NewEventsView(app *App) *EventsView {
	v := &EventsView{
		app:   app,
		name:  "events",
		title: "Calendar",
	}

	// Create calendar view
	v.calendar = NewCalendarView(app)
	v.calendar.SetOnDateSelect(v.onDateSelect)
	v.calendar.SetOnCalendarChange(v.onCalendarChange)

	// Create events list panel
	v.eventsList = tview.NewTextView()
	v.eventsList.SetDynamicColors(true)
	v.eventsList.SetBackgroundColor(app.styles.BgColor)
	v.eventsList.SetBorder(true)
	v.eventsList.SetBorderColor(app.styles.BorderColor)
	v.eventsList.SetTitle(" Events ")
	v.eventsList.SetTitleColor(app.styles.TitleFg)
	v.eventsList.SetBorderPadding(0, 0, 1, 1)

	// Create split layout: Calendar (left) | Events List (right)
	v.layout = tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(v.calendar, 0, 2, true).
		AddItem(v.eventsList, 0, 1, false)

	return v
}

func (v *EventsView) Name() string               { return v.name }
func (v *EventsView) Title() string              { return v.title }
func (v *EventsView) Primitive() tview.Primitive { return v.layout }
func (v *EventsView) Filter(string)              {}

func (v *EventsView) Hints() []Hint {
	return []Hint{
		{Key: "enter", Desc: "view day"},
		{Key: "c/C", Desc: "switch/list cal"},
		{Key: "m", Desc: "month"},
		{Key: "w", Desc: "week"},
		{Key: "a", Desc: "agenda"},
		{Key: "t", Desc: "today"},
		{Key: "H/L", Desc: "±month"},
		{Key: "r", Desc: "refresh"},
	}
}

func (v *EventsView) Load() {
	ctx := context.Background()

	// Get calendars first
	calendars, err := v.app.config.Client.GetCalendars(ctx, v.app.config.GrantID)
	if err != nil {
		v.app.Flash(FlashError, "Failed to load calendars: %v", err)
		return
	}

	v.calendars = calendars
	v.calendar.SetCalendars(calendars)

	if len(calendars) == 0 {
		v.app.Flash(FlashWarn, "No calendars found")
		return
	}

	// Load events for the current calendar
	v.loadEventsForCalendar(v.calendar.GetCurrentCalendarID())
}

func (v *EventsView) loadEventsForCalendar(calendarID string) {
	ctx := context.Background()

	// Get events from selected calendar (fetch 2 months range)
	now := time.Now()
	startTime := now.AddDate(0, -1, 0).Unix()
	endTime := now.AddDate(0, 2, 0).Unix()

	events, err := v.app.config.Client.GetEvents(ctx, v.app.config.GrantID, calendarID, &domain.EventQueryParams{
		Start:           startTime,
		End:             endTime,
		ExpandRecurring: true,
		Limit:           200,
	})
	if err != nil {
		v.app.Flash(FlashError, "Failed to load events: %v", err)
		return
	}

	v.events = events
	v.calendar.SetEvents(events)
	v.updateEventsList(v.calendar.GetSelectedDate())

	// Show calendar name in flash
	if cal := v.calendar.GetCurrentCalendar(); cal != nil {
		v.app.Flash(FlashInfo, "Calendar: %s (%d events)", cal.Name, len(events))
	}
}

func (v *EventsView) Refresh() { v.Load() }

func (v *EventsView) onCalendarChange(calendarID string) {
	// Reload events for the new calendar
	go func() {
		v.loadEventsForCalendar(calendarID)
		v.app.QueueUpdateDraw(func() {})
	}()
}

func (v *EventsView) onDateSelect(date time.Time) {
	v.updateEventsList(date)
}

func (v *EventsView) updateEventsList(date time.Time) {
	v.eventsList.Clear()

	events := v.calendar.GetEventsForDate(date)
	title := colorToHex(v.app.styles.TitleFg)
	info := colorToHex(v.app.styles.InfoColor)
	muted := colorToHex(v.app.styles.BorderColor)
	eventColor := colorToHex(v.app.styles.FgColor)
	success := colorToHex(v.app.styles.SuccessColor)

	// Header with date
	dateStr := date.Format("Monday, January 2, 2006")
	fmt.Fprintf(v.eventsList, "[%s::b]%s[-::-]\n\n", title, dateStr)

	if len(events) == 0 {
		fmt.Fprintf(v.eventsList, "[%s]No events scheduled[-]\n", muted)
		return
	}

	for i, evt := range events {
		// Time
		timeStr := "All day"
		if !evt.When.IsAllDay() {
			start := evt.When.StartDateTime()
			end := evt.When.EndDateTime()
			timeStr = fmt.Sprintf("%s - %s", start.Format("3:04 PM"), end.Format("3:04 PM"))
		}

		// Event entry
		fmt.Fprintf(v.eventsList, "[%s]%s[-]\n", info, timeStr)
		fmt.Fprintf(v.eventsList, "[%s::b]%s[-::-]\n", eventColor, evt.Title)

		// Location
		if evt.Location != "" {
			fmt.Fprintf(v.eventsList, "[%s]📍 %s[-]\n", muted, evt.Location)
		}

		// Status
		statusIcon := "✓"
		if evt.Status == "tentative" {
			statusIcon = "?"
		} else if evt.Status == "cancelled" {
			statusIcon = "✗"
		}
		fmt.Fprintf(v.eventsList, "[%s]%s %s[-]\n", success, statusIcon, evt.Status)

		// Separator between events
		if i < len(events)-1 {
			fmt.Fprintf(v.eventsList, "\n[%s]───────────────────[-]\n\n", muted)
		}
	}
}

func (v *EventsView) HandleKey(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEscape:
		// Let the app handle Escape for navigation
		return event

	case tcell.KeyTab:
		// Switch focus between calendar and events list
		v.focusedPanel = (v.focusedPanel + 1) % 2
		if v.focusedPanel == 0 {
			v.app.SetFocus(v.calendar)
		} else {
			v.app.SetFocus(v.eventsList)
		}
		return nil

	case tcell.KeyEnter:
		// Show detailed view of selected day's events
		v.showDayDetail()
		return nil

	case tcell.KeyRune:
		switch event.Rune() {
		case 'C': // Show calendar list
			v.showCalendarList()
			return nil
		}
	}

	// Pass to calendar if it has focus
	if v.focusedPanel == 0 {
		handler := v.calendar.InputHandler()
		if handler != nil {
			handler(event, func(p tview.Primitive) {})
			v.updateEventsList(v.calendar.GetSelectedDate())
			return nil
		}
	}

	return event
}

func (v *EventsView) showCalendarList() {
	calendars := v.calendar.GetCalendars()
	if len(calendars) == 0 {
		v.app.Flash(FlashWarn, "No calendars available")
		return
	}

	// Create a list view for calendar selection
	list := tview.NewList()
	list.SetBackgroundColor(v.app.styles.BgColor)
	list.SetBorder(true)
	list.SetBorderColor(v.app.styles.FocusColor)
	list.SetTitle(" Select Calendar ")
	list.SetTitleColor(v.app.styles.TitleFg)
	list.ShowSecondaryText(true)
	list.SetHighlightFullLine(true)
	list.SetSelectedBackgroundColor(v.app.styles.TableSelectBg)
	list.SetSelectedTextColor(v.app.styles.TableSelectFg)
	list.SetMainTextColor(v.app.styles.FgColor)
	list.SetSecondaryTextColor(v.app.styles.BorderColor)

	currentCal := v.calendar.GetCurrentCalendar()

	for i, cal := range calendars {
		name := cal.Name
		secondary := cal.ID
		if len(secondary) > 40 {
			secondary = secondary[:37] + "..."
		}

		// Mark primary and current
		if cal.IsPrimary {
			name = "★ " + name
		}
		if currentCal != nil && cal.ID == currentCal.ID {
			name = "● " + name
		}

		idx := i // Capture for closure
		list.AddItem(name, secondary, rune('1'+i), func() {
			v.calendar.SetCalendarByIndex(idx)
			v.app.PopDetail()
			v.app.SetFocus(v.calendar)
		})
	}

	// Handle escape to close
	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			v.app.PopDetail()
			v.app.SetFocus(v.calendar)
			return nil
		}
		return event
	})

	// Push the list as a detail view
	v.app.PushDetail("calendar-list", list)
	v.app.SetFocus(list)
}

func (v *EventsView) showDayDetail() {
	date := v.calendar.GetSelectedDate()
	events := v.calendar.GetEventsForDate(date)

	if len(events) == 0 {
		v.app.Flash(FlashInfo, "No events on %s", date.Format("Jan 2"))
		return
	}

	// Create detail view
	detail := tview.NewTextView()
	detail.SetDynamicColors(true)
	detail.SetBackgroundColor(v.app.styles.BgColor)
	detail.SetBorderPadding(1, 1, 2, 2)
	detail.SetScrollable(true)

	title := colorToHex(v.app.styles.TitleFg)
	info := colorToHex(v.app.styles.InfoColor)
	key := colorToHex(v.app.styles.FgColor)
	value := colorToHex(v.app.styles.InfoSectionFg)
	muted := colorToHex(v.app.styles.BorderColor)

	dateStr := date.Format("Monday, January 2, 2006")
	fmt.Fprintf(detail, "[%s::b]%s[-::-]\n", title, dateStr)
	fmt.Fprintf(detail, "[%s]%d event(s)[-]\n\n", muted, len(events))

	for i, evt := range events {
		fmt.Fprintf(detail, "[%s::b]%d. %s[-::-]\n", info, i+1, evt.Title)

		// Time
		timeStr := "All day"
		if !evt.When.IsAllDay() {
			start := evt.When.StartDateTime()
			end := evt.When.EndDateTime()
			timeStr = fmt.Sprintf("%s - %s", start.Format("3:04 PM"), end.Format("3:04 PM"))
		}
		fmt.Fprintf(detail, "[%s]Time:[-] [%s]%s[-]\n", key, value, timeStr)

		// Location
		if evt.Location != "" {
			fmt.Fprintf(detail, "[%s]Location:[-] [%s]%s[-]\n", key, value, evt.Location)
		}

		// Description
		if evt.Description != "" {
			desc := evt.Description
			if len(desc) > 200 {
				desc = desc[:200] + "..."
			}
			fmt.Fprintf(detail, "[%s]Description:[-] [%s]%s[-]\n", key, value, desc)
		}

		// Participants
		if len(evt.Participants) > 0 {
			fmt.Fprintf(detail, "[%s]Participants:[-]\n", key)
			for _, p := range evt.Participants {
				name := p.Name
				if name == "" {
					name = p.Email
				}
				status := p.Status
				if status == "" {
					status = "pending"
				}
				fmt.Fprintf(detail, "  [%s]• %s (%s)[-]\n", value, name, status)
			}
		}

		// Conferencing
		if evt.Conferencing != nil && evt.Conferencing.Details != nil && evt.Conferencing.Details.URL != "" {
			fmt.Fprintf(detail, "[%s]Meeting:[-] [%s]%s[-]\n", key, value, evt.Conferencing.Details.URL)
		}

		if i < len(events)-1 {
			fmt.Fprintf(detail, "\n[%s]════════════════════════════════════════[-]\n\n", muted)
		}
	}

	fmt.Fprintf(detail, "\n\n[%s::d]Press Esc to go back[-::-]", muted)

	// Push detail onto page stack
	v.app.PushDetail("day-detail", detail)
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

	// Different hints based on whether switching is available
	if app.CanSwitchGrant() {
		v.hints = []Hint{
			{Key: "enter", Desc: "switch"},
			{Key: "r", Desc: "refresh"},
		}
	} else {
		v.hints = []Hint{
			{Key: "r", Desc: "refresh"},
		}
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

	currentGrantID := v.app.config.GrantID

	for _, g := range v.grants {
		// Mark current/default grant with ★
		marker := ""
		if g.ID == currentGrantID {
			marker = "★"
		}

		data = append(data, []string{
			marker,
			g.Email,
			string(g.Provider),
			g.ID,
		})
		meta = append(meta, RowMeta{ID: g.ID, Data: &g})
	}

	v.table.SetData(data, meta)
}

// HandleKey handles key events for the grants view.
func (v *GrantsView) HandleKey(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEnter:
		// Switch to selected grant
		if !v.app.CanSwitchGrant() {
			v.app.Flash(FlashWarn, "Grant switching not available in demo mode")
			return nil
		}

		meta := v.table.SelectedMeta()
		if meta == nil || meta.Data == nil {
			return nil
		}

		grant, ok := meta.Data.(*domain.Grant)
		if !ok {
			return nil
		}

		// Check if already the current grant
		if grant.ID == v.app.config.GrantID {
			v.app.Flash(FlashInfo, "Already using this grant")
			return nil
		}

		// Switch to the selected grant
		if err := v.app.SwitchGrant(grant.ID, grant.Email, string(grant.Provider)); err != nil {
			v.app.Flash(FlashError, "Failed to switch: %v", err)
			return nil
		}

		v.app.Flash(FlashInfo, "Switched to %s", grant.Email)
		v.render() // Re-render to update the marker
		return nil

	case tcell.KeyEscape:
		return event
	}

	return event
}

// ============================================================================
// Inbound View
// ============================================================================

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
	ctx := context.Background()

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
	ctx := context.Background()

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

	fmt.Fprintf(detail, "[%s::b]%s[-::-]\n", title, msg.Subject)
	fmt.Fprintf(detail, "[%s]────────────────────────────────────────[-]\n", muted)
	fmt.Fprintf(detail, "[%s]From:[-] [%s]%s[-]\n", key, value, from)
	if len(to) > 0 {
		fmt.Fprintf(detail, "[%s]To:[-] [%s]%s[-]\n", key, value, strings.Join(to, ", "))
	}
	fmt.Fprintf(detail, "[%s]Date:[-] [%s]%s[-]\n", key, value, msg.Date.Format("Mon, Jan 2, 2006 3:04 PM"))
	fmt.Fprintf(detail, "[%s]ID:[-] [%s]%s[-]\n", key, value, msg.ID)
	fmt.Fprintf(detail, "[%s]────────────────────────────────────────[-]\n\n", muted)

	// Body
	body := msg.Body
	if body == "" {
		body = msg.Snippet
	}
	body = stripHTMLForTUI(body)
	fmt.Fprintf(detail, "[%s]%s[-]\n\n", value, tview.Escape(body))

	fmt.Fprintf(detail, "[%s]Press Esc to go back[-]", muted)

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

// ============================================================================
// Help View
// ============================================================================

// NewHelpView creates a help overlay.
func NewHelpView(styles *Styles) *tview.TextView {
	help := tview.NewTextView()
	help.SetDynamicColors(true)
	help.SetBackgroundColor(styles.BgColor)
	help.SetBorder(true)
	help.SetBorderColor(styles.FocusColor)
	help.SetTitle(" Help (vim-style) ")
	help.SetTitleColor(styles.TitleFg)
	help.SetScrollable(true)

	info := colorToHex(styles.FgColor)
	warn := colorToHex(styles.InfoColor)
	muted := colorToHex(styles.BorderColor)

	fmt.Fprintf(help, "\n")
	fmt.Fprintf(help, "  [%s::b]Vim Navigation[-::-]\n", warn)
	fmt.Fprintf(help, "  [%s]j/↓[-]       Move down\n", info)
	fmt.Fprintf(help, "  [%s]k/↑[-]       Move up\n", info)
	fmt.Fprintf(help, "  [%s]gg[-]        Go to first row\n", info)
	fmt.Fprintf(help, "  [%s]G[-]         Go to last row\n", info)
	fmt.Fprintf(help, "  [%s]Ctrl+d[-]    Half page down\n", info)
	fmt.Fprintf(help, "  [%s]Ctrl+u[-]    Half page up\n", info)
	fmt.Fprintf(help, "  [%s]Ctrl+f[-]    Full page down\n", info)
	fmt.Fprintf(help, "  [%s]Ctrl+b[-]    Full page up\n", info)
	fmt.Fprintf(help, "  [%s]:N[-]        Jump to row N (e.g. :5)\n", info)
	fmt.Fprintf(help, "\n")
	fmt.Fprintf(help, "  [%s::b]Vim Commands[-::-]\n", warn)
	fmt.Fprintf(help, "  [%s]:q[-]        Quit\n", info)
	fmt.Fprintf(help, "  [%s]:q![-]       Force quit\n", info)
	fmt.Fprintf(help, "  [%s]:wq[-]       Save and quit\n", info)
	fmt.Fprintf(help, "  [%s]:h[-]        Show help\n", info)
	fmt.Fprintf(help, "  [%s]:e <view>[-] Open view (e.g. :e messages)\n", info)
	fmt.Fprintf(help, "  [%s]/[-]         Filter/search\n", info)
	fmt.Fprintf(help, "\n")
	fmt.Fprintf(help, "  [%s::b]Actions[-::-]\n", warn)
	fmt.Fprintf(help, "  [%s]dd[-]        Delete item\n", info)
	fmt.Fprintf(help, "  [%s]x[-]         Delete item\n", info)
	fmt.Fprintf(help, "  [%s]:star[-]     Star message\n", info)
	fmt.Fprintf(help, "  [%s]:unread[-]   Mark as unread\n", info)
	fmt.Fprintf(help, "\n")
	fmt.Fprintf(help, "  [%s::b]Messages[-::-]\n", warn)
	fmt.Fprintf(help, "  [%s]n[-]         New/compose email\n", info)
	fmt.Fprintf(help, "  [%s]R[-]         Reply to message\n", info)
	fmt.Fprintf(help, "  [%s]A[-]         Reply all\n", info)
	fmt.Fprintf(help, "  [%s]s[-]         Toggle star\n", info)
	fmt.Fprintf(help, "  [%s]u[-]         Mark as unread\n", info)
	fmt.Fprintf(help, "  [%s]:reply[-]    Reply (command mode)\n", info)
	fmt.Fprintf(help, "  [%s]:compose[-]  New message (command mode)\n", info)
	fmt.Fprintf(help, "\n")
	fmt.Fprintf(help, "  [%s::b]Resource Navigation[-::-]\n", warn)
	fmt.Fprintf(help, "  [%s]:m[-]        Messages\n", info)
	fmt.Fprintf(help, "  [%s]:e[-]        Events/Calendar\n", info)
	fmt.Fprintf(help, "  [%s]:c[-]        Contacts\n", info)
	fmt.Fprintf(help, "  [%s]:i[-]        Inbound inboxes\n", info)
	fmt.Fprintf(help, "  [%s]:w[-]        Webhooks\n", info)
	fmt.Fprintf(help, "  [%s]:ws[-]       Webhook Server\n", info)
	fmt.Fprintf(help, "  [%s]:g[-]        Grants\n", info)
	fmt.Fprintf(help, "  [%s]:d[-]        Dashboard\n", info)
	fmt.Fprintf(help, "\n")
	fmt.Fprintf(help, "  [%s::b]General[-::-]\n", warn)
	fmt.Fprintf(help, "  [%s]r[-]         Refresh\n", info)
	fmt.Fprintf(help, "  [%s]?[-]         Show help\n", info)
	fmt.Fprintf(help, "  [%s]Esc[-]       Go back\n", info)
	fmt.Fprintf(help, "  [%s]Ctrl+C[-]    Quit\n", info)
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
