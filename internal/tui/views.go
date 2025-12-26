package tui

import (
	"context"
	"fmt"
	"html"
	"io"
	"os"
	"path/filepath"
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

// AttachmentInfo holds attachment metadata for download reference.
type AttachmentInfo struct {
	MessageID  string
	Attachment domain.Attachment
}

// MessagesView displays email threads (conversations).
type MessagesView struct {
	*BaseTableView
	threads         []domain.Thread
	showingDetail   bool
	currentThread   *domain.Thread
	currentMessage  *domain.Message // For reply functionality
	attachments     []AttachmentInfo // All attachments in current thread
	folderPanel     *FolderPanel
	currentFolderID string
	currentFolder   string // Display name for current folder
	showingFolders  bool
	layout          *tview.Flex // Main layout with optional folder panel
}

// NewMessagesView creates a new messages view.
func NewMessagesView(app *App) *MessagesView {
	v := &MessagesView{
		BaseTableView:   newBaseTableView(app, "messages", "Inbox"),
		currentFolder:   "Inbox",
		currentFolderID: "", // Will use INBOX by default in Load()
	}

	// Create folder panel with callback for folder selection
	v.folderPanel = NewFolderPanel(app, func(folder *domain.Folder) {
		v.currentFolderID = folder.ID
		v.currentFolder = folder.Name
		if folder.SystemFolder != "" {
			v.currentFolder = v.folderPanel.getSystemFolderName(folder.SystemFolder)
		}
		v.title = v.currentFolder
		v.showingFolders = false
		v.updateLayout()
		app.SetFocus(v.table)
		v.Load()
	})

	// Create layout
	v.layout = tview.NewFlex()
	v.updateLayout()

	v.hints = []Hint{
		{Key: "enter", Desc: "view"},
		{Key: "n", Desc: "compose"},
		{Key: "R", Desc: "reply"},
		{Key: "s", Desc: "star"},
		{Key: "u", Desc: "unread"},
		{Key: "F", Desc: "folders"},
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Load folders if not already loaded
	if len(v.folderPanel.folders) == 0 {
		v.folderPanel.Load()
	}

	// Build folder filter - use folder ID if set, otherwise default to INBOX
	var folderFilter []string
	if v.currentFolderID != "" {
		folderFilter = []string{v.currentFolderID}
	} else {
		folderFilter = []string{"INBOX"}
	}

	params := &domain.ThreadQueryParams{
		Limit: 50,
		In:    folderFilter,
	}
	threads, err := v.app.config.Client.GetThreads(ctx, v.app.config.GrantID, params)
	if err != nil {
		v.app.Flash(FlashError, "Failed to load threads: %v", err)
		return
	}
	v.threads = threads
	v.render()
}

// updateLayout rebuilds the layout based on folder panel visibility.
func (v *MessagesView) updateLayout() {
	v.layout.Clear()
	if v.showingFolders {
		v.layout.AddItem(v.folderPanel, 30, 0, true)
		v.layout.AddItem(v.table, 0, 1, false)
	} else {
		v.layout.AddItem(v.table, 0, 1, true)
	}
}

// Primitive returns the root primitive for this view.
func (v *MessagesView) Primitive() tview.Primitive {
	return v.layout
}

func (v *MessagesView) Refresh() {
	v.Load()
}

func (v *MessagesView) render() {
	var data [][]string
	var meta []RowMeta

	// Parse search query if filter is set
	var searchQuery *SearchQuery
	if v.filter != "" {
		searchQuery = ParseSearchQuery(v.filter)
	}

	for _, thread := range v.threads {
		// Apply search filter
		if searchQuery != nil && !searchQuery.MatchesThread(&thread) {
			continue
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
	// If folder panel is showing, delegate to it
	if v.showingFolders {
		switch event.Key() {
		case tcell.KeyEscape:
			v.showingFolders = false
			v.updateLayout()
			v.app.SetFocus(v.table)
			return nil
		case tcell.KeyTab:
			// Tab switches between folder panel and message list
			v.app.SetFocus(v.table)
			return nil
		default:
			// Let folder panel handle other keys
			return v.folderPanel.handleInput(event)
		}
	}

	switch event.Key() {
	case tcell.KeyEscape:
		// If showing detail, close it and return nil to indicate we handled it
		if v.showingDetail {
			v.closeDetail()
			return nil
		}
		// Otherwise, let app handle the Escape
		return event

	case tcell.KeyTab:
		// Tab toggles folder panel when not showing detail
		if !v.showingDetail && v.showingFolders {
			v.app.SetFocus(v.folderPanel)
			return nil
		}
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
		case 'F':
			// Toggle folder panel
			v.showingFolders = !v.showingFolders
			v.updateLayout()
			if v.showingFolders {
				v.app.SetFocus(v.folderPanel)
			} else {
				v.app.SetFocus(v.table)
			}
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
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
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
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
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

			// Clear attachments list
			v.attachments = nil

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
					fmt.Fprintf(detail, "[%s]Date:[-] [%s]%s[-]\n", key, value, msg.Date.Format("Mon, Jan 2, 2006 3:04 PM"))

					// Display attachments if any
					if len(msg.Attachments) > 0 {
						fmt.Fprintf(detail, "[%s]Attachments:[-]", key)
						for _, att := range msg.Attachments {
							if att.IsInline {
								continue // Skip inline attachments (images in HTML)
							}
							// Track attachment with its message ID
							attachmentIdx := len(v.attachments)
							v.attachments = append(v.attachments, AttachmentInfo{
								MessageID:  msg.ID,
								Attachment: att,
							})
							sizeStr := formatFileSize(att.Size)
							fmt.Fprintf(detail, " [%s][%d] %s (%s)[-]", hint, attachmentIdx+1, att.Filename, sizeStr)
						}
						fmt.Fprintln(detail)
					}
					fmt.Fprintln(detail)

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

			// Build help line based on available actions
			helpLine := fmt.Sprintf("[%s]R[-][%s::d]=reply  [-::-][%s]A[-][%s::d]=reply all  [-::-]", hint, muted, hint, muted)
			if len(v.attachments) > 0 {
				helpLine += fmt.Sprintf("[%s]D[-][%s::d]=download  [-::-]", hint, muted)
			}
			helpLine += fmt.Sprintf("[%s]Esc[-][%s::d]=back[-::-]", hint, muted)
			fmt.Fprint(detail, helpLine)
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
			case 'D':
				if len(v.attachments) > 0 {
					v.showDownloadDialog()
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
	v.attachments = nil
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

func (v *MessagesView) showDownloadDialog() {
	if len(v.attachments) == 0 {
		return
	}

	styles := v.app.styles

	// Create list for attachment selection
	list := tview.NewList()
	list.SetBackgroundColor(styles.BgColor)
	list.SetMainTextColor(styles.FgColor)
	list.SetSecondaryTextColor(styles.InfoColor)
	list.SetSelectedBackgroundColor(styles.FocusColor)
	list.SetSelectedTextColor(styles.BgColor)
	list.SetBorder(true)
	list.SetBorderColor(styles.FocusColor)
	list.SetTitle(" Download Attachment ")
	list.SetTitleColor(styles.TitleFg)

	// Add attachments to list
	for i, attInfo := range v.attachments {
		idx := i
		att := attInfo.Attachment
		msgID := attInfo.MessageID
		sizeStr := formatFileSize(att.Size)
		list.AddItem(
			fmt.Sprintf("%s (%s)", att.Filename, sizeStr),
			att.ContentType,
			rune('1'+i),
			func() {
				v.downloadAttachment(msgID, att.ID, att.Filename, idx+1)
			},
		)
	}

	// Handle Escape to close
	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			v.app.PopDetail()
			return nil
		}
		return event
	})

	// Center the list
	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(nil, 0, 1, false).
			AddItem(list, 60, 0, true).
			AddItem(nil, 0, 1, false), 0, 2, true).
		AddItem(nil, 0, 1, false)

	v.app.PushDetail("download-dialog", flex)
	v.app.SetFocus(list)
}

func (v *MessagesView) downloadAttachment(messageID, attachmentID, filename string, displayNum int) {
	v.app.Flash(FlashInfo, "Downloading %s...", filename)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		reader, err := v.app.config.Client.DownloadAttachment(ctx, v.app.config.GrantID, messageID, attachmentID)
		if err != nil {
			v.app.QueueUpdateDraw(func() {
				v.app.Flash(FlashError, "Download failed: %v", err)
			})
			return
		}
		defer reader.Close()

		// Get Downloads directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			v.app.QueueUpdateDraw(func() {
				v.app.Flash(FlashError, "Cannot find home directory: %v", err)
			})
			return
		}
		downloadDir := filepath.Join(homeDir, "Downloads")

		// Ensure download directory exists
		if err := os.MkdirAll(downloadDir, 0750); err != nil {
			v.app.QueueUpdateDraw(func() {
				v.app.Flash(FlashError, "Cannot create Downloads directory: %v", err)
			})
			return
		}

		// Create file with unique name if exists
		destPath := filepath.Join(downloadDir, filename)
		destPath = v.getUniqueFilename(destPath)

		file, err := os.Create(destPath)
		if err != nil {
			v.app.QueueUpdateDraw(func() {
				v.app.Flash(FlashError, "Cannot create file: %v", err)
			})
			return
		}
		defer file.Close()

		// Copy content
		written, err := io.Copy(file, reader)
		if err != nil {
			v.app.QueueUpdateDraw(func() {
				v.app.Flash(FlashError, "Download failed: %v", err)
			})
			return
		}

		v.app.QueueUpdateDraw(func() {
			v.app.PopDetail() // Close download dialog
			v.app.Flash(FlashInfo, "Downloaded %s (%s) to %s", filename, formatFileSize(written), destPath)
		})
	}()
}

func (v *MessagesView) getUniqueFilename(path string) string {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path
	}

	dir := filepath.Dir(path)
	ext := filepath.Ext(path)
	name := strings.TrimSuffix(filepath.Base(path), ext)

	for i := 1; i < 1000; i++ {
		newPath := filepath.Join(dir, fmt.Sprintf("%s (%d)%s", name, i, ext))
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}
	}

	// Fallback: append timestamp
	return filepath.Join(dir, fmt.Sprintf("%s_%d%s", name, time.Now().Unix(), ext))
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
		{Key: "n", Desc: "new event"},
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

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

		// Title with recurring indicator
		title := evt.Title
		if isRecurringEvent(&evt) {
			title = "🔁 " + title
		}
		fmt.Fprintf(v.eventsList, "[%s::b]%s[-::-]\n", eventColor, title)

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
		case 'n': // New event
			v.createNewEvent()
			return nil
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

		// Add description if available
		if cal.Description != "" {
			desc := cal.Description
			if len(desc) > 30 {
				desc = desc[:27] + "..."
			}
			secondary = desc + " | " + secondary
		}

		// Add color indicator
		if cal.HexColor != "" {
			name = "■ " + name // Color square (will be colored in custom draw)
		}

		// Mark primary and current
		if cal.IsPrimary {
			name = "★ " + name
		}
		if currentCal != nil && cal.ID == currentCal.ID {
			name = "● " + name
		}

		// Add read-only indicator
		if cal.ReadOnly {
			name = name + " [RO]"
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

func (v *EventsView) createNewEvent() {
	calendarID := v.calendar.GetCurrentCalendarID()
	if calendarID == "" {
		v.app.Flash(FlashWarn, "No calendar selected")
		return
	}

	v.app.ShowEventForm(calendarID, nil, func(event *domain.Event) {
		// Refresh events after creation
		go func() {
			v.loadEventsForCalendar(calendarID)
			v.app.QueueUpdateDraw(func() {})
		}()
	})
}

func (v *EventsView) showDayDetail() {
	date := v.calendar.GetSelectedDate()
	events := v.calendar.GetEventsForDate(date)

	if len(events) == 0 {
		v.app.Flash(FlashInfo, "No events on %s - press 'n' to create one", date.Format("Jan 2"))
		return
	}

	// Create a list for event selection (supports edit/delete)
	list := tview.NewList()
	list.SetBackgroundColor(v.app.styles.BgColor)
	list.SetBorder(true)
	list.SetBorderColor(v.app.styles.FocusColor)
	list.SetTitle(fmt.Sprintf(" %s (%d events) ", date.Format("Jan 2, 2006"), len(events)))
	list.SetTitleColor(v.app.styles.TitleFg)
	list.ShowSecondaryText(true)
	list.SetHighlightFullLine(true)
	list.SetSelectedBackgroundColor(v.app.styles.TableSelectBg)
	list.SetSelectedTextColor(v.app.styles.TableSelectFg)
	list.SetMainTextColor(v.app.styles.FgColor)
	list.SetSecondaryTextColor(v.app.styles.BorderColor)

	calendarID := v.calendar.GetCurrentCalendarID()

	for i, evt := range events {
		// Build main text
		title := evt.Title
		if evt.When.IsAllDay() {
			title = "📅 " + title
		}
		// Add recurring indicator
		if isRecurringEvent(&evt) {
			title = "🔁 " + title
		}

		// Build secondary text with time
		timeStr := "All day"
		if !evt.When.IsAllDay() {
			start := evt.When.StartDateTime()
			end := evt.When.EndDateTime()
			timeStr = fmt.Sprintf("%s - %s", start.Format("3:04 PM"), end.Format("3:04 PM"))
		}
		secondary := timeStr
		if evt.Location != "" {
			secondary += " | 📍 " + evt.Location
		}

		// Capture event for closure
		eventCopy := events[i]

		list.AddItem(title, secondary, 0, func() {
			// Show event detail on Enter
			v.showEventDetail(&eventCopy)
		})
	}

	// Handle keyboard events
	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			v.app.PopDetail()
			v.app.SetFocus(v.calendar)
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'e': // Edit selected event
				idx := list.GetCurrentItem()
				if idx >= 0 && idx < len(events) {
					evt := events[idx]
					if isRecurringEvent(&evt) {
						// Show dialog for recurring event
						v.showRecurringEventEditDialog(calendarID, &evt, list)
					} else {
						v.app.PopDetail()
						v.app.ShowEventForm(calendarID, &evt, func(updatedEvent *domain.Event) {
							go func() {
								v.loadEventsForCalendar(calendarID)
								v.app.QueueUpdateDraw(func() {})
							}()
						})
					}
				}
				return nil
			case 'd': // Delete selected event
				idx := list.GetCurrentItem()
				if idx >= 0 && idx < len(events) {
					evt := events[idx]
					if isRecurringEvent(&evt) {
						// Show dialog for recurring event
						v.showRecurringEventDeleteDialog(calendarID, &evt, list)
					} else {
						v.app.PopDetail()
						v.app.DeleteEvent(calendarID, &evt, func() {
							go func() {
								v.loadEventsForCalendar(calendarID)
								v.app.QueueUpdateDraw(func() {})
							}()
						})
					}
				}
				return nil
			case 'n': // New event
				v.app.PopDetail()
				v.createNewEvent()
				return nil
			}
		}
		return event
	})

	// Push list onto page stack
	v.app.PushDetail("day-detail", list)
	v.app.SetFocus(list)
}

func (v *EventsView) showEventDetail(evt *domain.Event) {
	// Create detailed view of a single event
	detail := tview.NewTextView()
	detail.SetDynamicColors(true)
	detail.SetBackgroundColor(v.app.styles.BgColor)
	detail.SetBorder(true)
	detail.SetBorderColor(v.app.styles.FocusColor)
	detail.SetTitle(fmt.Sprintf(" %s ", evt.Title))
	detail.SetTitleColor(v.app.styles.TitleFg)
	detail.SetBorderPadding(1, 1, 2, 2)
	detail.SetScrollable(true)

	info := colorToHex(v.app.styles.InfoColor)
	key := colorToHex(v.app.styles.FgColor)
	value := colorToHex(v.app.styles.InfoSectionFg)
	muted := colorToHex(v.app.styles.BorderColor)

	// Time
	var timeStr string
	if !evt.When.IsAllDay() {
		start := evt.When.StartDateTime()
		end := evt.When.EndDateTime()
		dateStr := start.Format("Monday, January 2, 2006")
		timeStr = fmt.Sprintf("%s\n%s - %s", dateStr, start.Format("3:04 PM"), end.Format("3:04 PM"))
	} else {
		timeStr = evt.When.StartDateTime().Format("Monday, January 2, 2006") + " (All day)"
	}
	fmt.Fprintf(detail, "[%s::b]When[-::-]\n[%s]%s[-]\n\n", info, value, timeStr)

	// Location
	if evt.Location != "" {
		fmt.Fprintf(detail, "[%s::b]Location[-::-]\n[%s]%s[-]\n\n", info, value, evt.Location)
	}

	// Description
	if evt.Description != "" {
		fmt.Fprintf(detail, "[%s::b]Description[-::-]\n[%s]%s[-]\n\n", info, value, evt.Description)
	}

	// Participants
	if len(evt.Participants) > 0 {
		fmt.Fprintf(detail, "[%s::b]Participants[-::-]\n", info)
		for _, p := range evt.Participants {
			name := p.Name
			if name == "" {
				name = p.Email
			}
			status := p.Status
			if status == "" {
				status = "pending"
			}
			statusIcon := "⏳"
			switch status {
			case "yes":
				statusIcon = "✓"
			case "no":
				statusIcon = "✗"
			case "maybe":
				statusIcon = "?"
			}
			fmt.Fprintf(detail, "[%s]  %s %s[-]\n", value, statusIcon, name)
		}
		fmt.Fprintln(detail)
	}

	// Conferencing
	if evt.Conferencing != nil && evt.Conferencing.Details != nil && evt.Conferencing.Details.URL != "" {
		fmt.Fprintf(detail, "[%s::b]Meeting Link[-::-]\n[%s]%s[-]\n\n", info, value, evt.Conferencing.Details.URL)
	}

	// Recurrence
	if isRecurringEvent(evt) {
		fmt.Fprintf(detail, "[%s::b]Recurrence[-::-]\n", info)
		if len(evt.Recurrence) > 0 {
			recurrenceStr := formatRecurrenceRule(evt.Recurrence)
			if recurrenceStr != "" {
				fmt.Fprintf(detail, "[%s]🔁 %s[-]\n\n", value, recurrenceStr)
			} else {
				fmt.Fprintf(detail, "[%s]🔁 Recurring event[-]\n\n", value)
			}
		} else if evt.MasterEventID != "" {
			fmt.Fprintf(detail, "[%s]🔁 Instance of recurring event[-]\n\n", value)
		}
	}

	// Status
	fmt.Fprintf(detail, "[%s]Status:[-] [%s]%s[-]\n", key, value, evt.Status)
	if evt.Busy {
		fmt.Fprintf(detail, "[%s]Availability:[-] [%s]Busy[-]\n", key, value)
	} else {
		fmt.Fprintf(detail, "[%s]Availability:[-] [%s]Free[-]\n", key, value)
	}

	fmt.Fprintf(detail, "\n\n[%s::d]Press Esc to go back[-::-]", muted)

	// Handle escape
	detail.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			v.app.PopDetail()
			return nil
		}
		return event
	})

	v.app.PushDetail("event-detail", detail)
	v.app.SetFocus(detail)
}

func (v *EventsView) showRecurringEventEditDialog(calendarID string, evt *domain.Event, parentList *tview.List) {
	// Create a simple list for the options
	optionsList := tview.NewList()
	optionsList.SetBackgroundColor(v.app.styles.BgColor)
	optionsList.SetBorder(true)
	optionsList.SetBorderColor(v.app.styles.FocusColor)
	optionsList.SetTitle(" Edit Recurring Event ")
	optionsList.SetTitleColor(v.app.styles.TitleFg)
	optionsList.ShowSecondaryText(true)
	optionsList.SetHighlightFullLine(true)
	optionsList.SetSelectedBackgroundColor(v.app.styles.TableSelectBg)
	optionsList.SetSelectedTextColor(v.app.styles.TableSelectFg)
	optionsList.SetMainTextColor(v.app.styles.FgColor)
	optionsList.SetSecondaryTextColor(v.app.styles.BorderColor)

	eventCopy := *evt

	// Add options
	optionsList.AddItem("Edit this occurrence", "Only modify this instance", '1', func() {
		v.app.PopDetail() // Close options dialog
		v.app.PopDetail() // Close day detail
		// For editing a single occurrence, we pass the event as-is
		// The API will handle creating an exception
		v.app.ShowEventForm(calendarID, &eventCopy, func(updatedEvent *domain.Event) {
			go func() {
				v.loadEventsForCalendar(calendarID)
				v.app.QueueUpdateDraw(func() {})
			}()
		})
	})

	optionsList.AddItem("Edit all occurrences", "Modify the entire series", '2', func() {
		v.app.PopDetail() // Close options dialog
		v.app.PopDetail() // Close day detail
		// For editing the series, we need to use the master event ID if available
		editEvt := &eventCopy
		if eventCopy.MasterEventID != "" {
			// This is an instance - we'd need to fetch the master event
			// For now, just edit the current event which will prompt the API behavior
			v.app.Flash(FlashInfo, "Editing series from instance...")
		}
		v.app.ShowEventForm(calendarID, editEvt, func(updatedEvent *domain.Event) {
			go func() {
				v.loadEventsForCalendar(calendarID)
				v.app.QueueUpdateDraw(func() {})
			}()
		})
	})

	optionsList.AddItem("Cancel", "Go back", 'c', func() {
		v.app.PopDetail()
	})

	// Handle escape
	optionsList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			v.app.PopDetail()
			return nil
		}
		return event
	})

	v.app.PushDetail("recurring-edit-options", optionsList)
	v.app.SetFocus(optionsList)
}

func (v *EventsView) showRecurringEventDeleteDialog(calendarID string, evt *domain.Event, parentList *tview.List) {
	// Create a simple list for the options
	optionsList := tview.NewList()
	optionsList.SetBackgroundColor(v.app.styles.BgColor)
	optionsList.SetBorder(true)
	optionsList.SetBorderColor(v.app.styles.FocusColor)
	optionsList.SetTitle(" Delete Recurring Event ")
	optionsList.SetTitleColor(v.app.styles.TitleFg)
	optionsList.ShowSecondaryText(true)
	optionsList.SetHighlightFullLine(true)
	optionsList.SetSelectedBackgroundColor(v.app.styles.TableSelectBg)
	optionsList.SetSelectedTextColor(v.app.styles.TableSelectFg)
	optionsList.SetMainTextColor(v.app.styles.FgColor)
	optionsList.SetSecondaryTextColor(v.app.styles.BorderColor)

	eventCopy := *evt

	// Add options
	optionsList.AddItem("Delete this occurrence", "Only remove this instance", '1', func() {
		v.app.PopDetail() // Close options dialog
		v.app.PopDetail() // Close day detail
		v.app.ShowConfirmDialog("Delete Occurrence",
			fmt.Sprintf("Delete this occurrence of '%s'?", eventCopy.Title),
			func() {
				v.app.DeleteEvent(calendarID, &eventCopy, func() {
					go func() {
						v.loadEventsForCalendar(calendarID)
						v.app.QueueUpdateDraw(func() {})
					}()
				})
			})
	})

	optionsList.AddItem("Delete all occurrences", "Remove the entire series", '2', func() {
		v.app.PopDetail() // Close options dialog
		v.app.PopDetail() // Close day detail
		v.app.ShowConfirmDialog("Delete Series",
			fmt.Sprintf("Delete all occurrences of '%s'? This cannot be undone.", eventCopy.Title),
			func() {
				// For deleting the series, we use the master event ID if available
				deleteEvt := &eventCopy
				if eventCopy.MasterEventID != "" {
					deleteEvt = &domain.Event{ID: eventCopy.MasterEventID}
				}
				v.app.DeleteEvent(calendarID, deleteEvt, func() {
					go func() {
						v.loadEventsForCalendar(calendarID)
						v.app.QueueUpdateDraw(func() {})
					}()
				})
			})
	})

	optionsList.AddItem("Cancel", "Go back", 'c', func() {
		v.app.PopDetail()
	})

	// Handle escape
	optionsList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			v.app.PopDetail()
			return nil
		}
		return event
	})

	v.app.PushDetail("recurring-delete-options", optionsList)
	v.app.SetFocus(optionsList)
}

// isRecurringEvent returns true if the event is recurring.
func isRecurringEvent(evt *domain.Event) bool {
	return len(evt.Recurrence) > 0 || evt.MasterEventID != ""
}

// formatRecurrenceRule formats an RRULE string into a human-readable format.
func formatRecurrenceRule(rules []string) string {
	if len(rules) == 0 {
		return ""
	}

	// Find the first RRULE
	var rule string
	for _, r := range rules {
		if len(r) >= 6 && r[:6] == "RRULE:" {
			rule = r[6:]
			break
		}
		if len(r) > 0 && r[0] != 'E' { // Not EXDATE
			rule = r
			break
		}
	}

	if rule == "" {
		return ""
	}

	// Parse the RRULE
	parts := make(map[string]string)
	for _, part := range splitRRuleParts(rule) {
		if idx := indexByte(part, '='); idx > 0 {
			parts[part[:idx]] = part[idx+1:]
		}
	}

	freq := parts["FREQ"]
	interval := parts["INTERVAL"]
	if interval == "" {
		interval = "1"
	}
	byday := parts["BYDAY"]
	count := parts["COUNT"]
	until := parts["UNTIL"]

	// Build human-readable string
	var result string
	switch freq {
	case "DAILY":
		if interval == "1" {
			result = "Every day"
		} else {
			result = "Every " + interval + " days"
		}
	case "WEEKLY":
		if interval == "1" {
			result = "Every week"
		} else {
			result = "Every " + interval + " weeks"
		}
		if byday != "" {
			result += " on " + formatDays(byday)
		}
	case "MONTHLY":
		if interval == "1" {
			result = "Every month"
		} else {
			result = "Every " + interval + " months"
		}
	case "YEARLY":
		if interval == "1" {
			result = "Every year"
		} else {
			result = "Every " + interval + " years"
		}
	default:
		result = rule
	}

	// Add end condition
	if count != "" {
		result += " (" + count + " times)"
	} else if until != "" {
		// Parse UNTIL date (format: YYYYMMDD or YYYYMMDDTHHmmssZ)
		if len(until) >= 8 {
			result += " until " + until[:4] + "-" + until[4:6] + "-" + until[6:8]
		}
	}

	return result
}

// splitRRuleParts splits an RRULE into its component parts.
func splitRRuleParts(rule string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(rule); i++ {
		if rule[i] == ';' {
			parts = append(parts, rule[start:i])
			start = i + 1
		}
	}
	if start < len(rule) {
		parts = append(parts, rule[start:])
	}
	return parts
}

// indexByte returns the index of the first occurrence of c in s, or -1 if not found.
func indexByte(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

// formatDays formats BYDAY values into human-readable day names.
func formatDays(byday string) string {
	dayMap := map[string]string{
		"SU": "Sun", "MO": "Mon", "TU": "Tue", "WE": "Wed",
		"TH": "Thu", "FR": "Fri", "SA": "Sat",
	}

	var days []string
	for _, part := range splitByComma(byday) {
		// Handle numeric prefix (e.g., "1MO" for first Monday)
		day := part
		if len(part) > 2 {
			day = part[len(part)-2:]
		}
		if name, ok := dayMap[day]; ok {
			days = append(days, name)
		}
	}

	if len(days) == 0 {
		return byday
	}
	return joinStrings(days, ", ")
}

// splitByComma splits a string by comma.
func splitByComma(s string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			parts = append(parts, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		parts = append(parts, s[start:])
	}
	return parts
}

// joinStrings joins strings with a separator.
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
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
		{Key: "n", Desc: "new"},
		{Key: "e", Desc: "edit"},
		{Key: "d", Desc: "delete"},
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
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

// HandleKey handles keyboard input for contacts view.
func (v *ContactsView) HandleKey(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEnter:
		// View contact detail
		if idx, _ := v.table.GetSelection(); idx > 0 && idx-1 < len(v.contacts) {
			v.showContactDetail(&v.contacts[idx-1])
		}
		return nil

	case tcell.KeyRune:
		switch event.Rune() {
		case 'n': // New contact
			v.app.ShowContactForm(nil, func(contact *domain.Contact) {
				v.Refresh()
			})
			return nil

		case 'e': // Edit selected contact
			if idx, _ := v.table.GetSelection(); idx > 0 && idx-1 < len(v.contacts) {
				contact := v.contacts[idx-1]
				v.app.ShowContactForm(&contact, func(updatedContact *domain.Contact) {
					v.Refresh()
				})
			}
			return nil

		case 'd': // Delete selected contact
			if idx, _ := v.table.GetSelection(); idx > 0 && idx-1 < len(v.contacts) {
				contact := v.contacts[idx-1]
				v.app.DeleteContact(&contact, func() {
					v.Refresh()
				})
			}
			return nil
		}
	}

	return event
}

func (v *ContactsView) showContactDetail(contact *domain.Contact) {
	detail := tview.NewTextView()
	detail.SetDynamicColors(true)
	detail.SetBackgroundColor(v.app.styles.BgColor)
	detail.SetBorder(true)
	detail.SetBorderColor(v.app.styles.FocusColor)
	detail.SetTitle(fmt.Sprintf(" %s ", contact.DisplayName()))
	detail.SetTitleColor(v.app.styles.TitleFg)
	detail.SetBorderPadding(1, 1, 2, 2)
	detail.SetScrollable(true)

	info := colorToHex(v.app.styles.InfoColor)
	value := colorToHex(v.app.styles.InfoSectionFg)
	muted := colorToHex(v.app.styles.BorderColor)

	// Name
	if contact.GivenName != "" || contact.Surname != "" {
		fmt.Fprintf(detail, "[%s::b]Name[-::-]\n", info)
		if contact.GivenName != "" {
			fmt.Fprintf(detail, "[%s]%s[-]", value, contact.GivenName)
		}
		if contact.Surname != "" {
			if contact.GivenName != "" {
				fmt.Fprintf(detail, "[%s] %s[-]", value, contact.Surname)
			} else {
				fmt.Fprintf(detail, "[%s]%s[-]", value, contact.Surname)
			}
		}
		fmt.Fprintln(detail)
	}

	// Emails
	if len(contact.Emails) > 0 {
		fmt.Fprintf(detail, "[%s::b]Email[-::-]\n", info)
		for _, e := range contact.Emails {
			typeStr := e.Type
			if typeStr == "" {
				typeStr = "other"
			}
			fmt.Fprintf(detail, "[%s]%s[-] [%s](%s)[-]\n", value, e.Email, muted, typeStr)
		}
		fmt.Fprintln(detail)
	}

	// Phone numbers
	if len(contact.PhoneNumbers) > 0 {
		fmt.Fprintf(detail, "[%s::b]Phone[-::-]\n", info)
		for _, p := range contact.PhoneNumbers {
			typeStr := p.Type
			if typeStr == "" {
				typeStr = "other"
			}
			fmt.Fprintf(detail, "[%s]%s[-] [%s](%s)[-]\n", value, p.Number, muted, typeStr)
		}
		fmt.Fprintln(detail)
	}

	// Company
	if contact.CompanyName != "" || contact.JobTitle != "" {
		fmt.Fprintf(detail, "[%s::b]Work[-::-]\n", info)
		if contact.JobTitle != "" {
			fmt.Fprintf(detail, "[%s]%s[-]\n", value, contact.JobTitle)
		}
		if contact.CompanyName != "" {
			fmt.Fprintf(detail, "[%s]%s[-]\n", value, contact.CompanyName)
		}
		fmt.Fprintln(detail)
	}

	// Notes
	if contact.Notes != "" {
		fmt.Fprintf(detail, "[%s::b]Notes[-::-]\n", info)
		fmt.Fprintf(detail, "[%s]%s[-]\n\n", value, contact.Notes)
	}

	fmt.Fprintf(detail, "\n[%s::d]Press Esc to go back, 'e' to edit, 'd' to delete[-::-]", muted)

	// Handle keyboard
	contactCopy := contact
	detail.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			v.app.PopDetail()
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'e':
				v.app.PopDetail()
				v.app.ShowContactForm(contactCopy, func(updatedContact *domain.Contact) {
					v.Refresh()
				})
				return nil
			case 'd':
				v.app.PopDetail()
				v.app.DeleteContact(contactCopy, func() {
					v.Refresh()
				})
				return nil
			}
		}
		return event
	})

	v.app.PushDetail("contact-detail", detail)
	v.app.SetFocus(detail)
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
		{Key: "n", Desc: "new"},
		{Key: "e", Desc: "edit"},
		{Key: "d", Desc: "delete"},
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
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

// HandleKey handles keyboard input for webhooks view.
func (v *WebhooksView) HandleKey(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEnter:
		// View webhook detail
		if idx, _ := v.table.GetSelection(); idx > 0 && idx-1 < len(v.webhooks) {
			v.showWebhookDetail(&v.webhooks[idx-1])
		}
		return nil

	case tcell.KeyRune:
		switch event.Rune() {
		case 'n': // New webhook
			v.app.ShowWebhookForm(nil, func(webhook *domain.Webhook) {
				v.Refresh()
			})
			return nil

		case 'e': // Edit selected webhook
			if idx, _ := v.table.GetSelection(); idx > 0 && idx-1 < len(v.webhooks) {
				webhook := v.webhooks[idx-1]
				v.app.ShowWebhookForm(&webhook, func(updatedWebhook *domain.Webhook) {
					v.Refresh()
				})
			}
			return nil

		case 'd': // Delete selected webhook
			if idx, _ := v.table.GetSelection(); idx > 0 && idx-1 < len(v.webhooks) {
				webhook := v.webhooks[idx-1]
				v.app.DeleteWebhook(&webhook, func() {
					v.Refresh()
				})
			}
			return nil
		}
	}

	return event
}

func (v *WebhooksView) showWebhookDetail(webhook *domain.Webhook) {
	detail := tview.NewTextView()
	detail.SetDynamicColors(true)
	detail.SetBackgroundColor(v.app.styles.BgColor)
	detail.SetBorder(true)
	detail.SetBorderColor(v.app.styles.FocusColor)

	titleStr := webhook.Description
	if titleStr == "" {
		titleStr = webhook.ID
	}
	detail.SetTitle(fmt.Sprintf(" Webhook: %s ", titleStr))
	detail.SetTitleColor(v.app.styles.TitleFg)
	detail.SetBorderPadding(1, 1, 2, 2)
	detail.SetScrollable(true)

	info := colorToHex(v.app.styles.InfoColor)
	value := colorToHex(v.app.styles.InfoSectionFg)
	muted := colorToHex(v.app.styles.BorderColor)
	success := colorToHex(v.app.styles.SuccessColor)
	errColor := colorToHex(v.app.styles.ErrorColor)

	// URL
	fmt.Fprintf(detail, "[%s::b]Webhook URL[-::-]\n", info)
	fmt.Fprintf(detail, "[%s]%s[-]\n\n", value, webhook.WebhookURL)

	// Status
	fmt.Fprintf(detail, "[%s::b]Status[-::-]\n", info)
	statusColor := success
	if webhook.Status != "active" {
		statusColor = errColor
	}
	fmt.Fprintf(detail, "[%s]%s[-]\n\n", statusColor, webhook.Status)

	// Description
	if webhook.Description != "" {
		fmt.Fprintf(detail, "[%s::b]Description[-::-]\n", info)
		fmt.Fprintf(detail, "[%s]%s[-]\n\n", value, webhook.Description)
	}

	// Trigger types
	fmt.Fprintf(detail, "[%s::b]Trigger Types[-::-]\n", info)
	for _, trigger := range webhook.TriggerTypes {
		fmt.Fprintf(detail, "[%s]  • %s[-]\n", value, trigger)
	}
	fmt.Fprintln(detail)

	// Notification emails
	if len(webhook.NotificationEmailAddresses) > 0 {
		fmt.Fprintf(detail, "[%s::b]Notification Emails[-::-]\n", info)
		for _, email := range webhook.NotificationEmailAddresses {
			fmt.Fprintf(detail, "[%s]  • %s[-]\n", value, email)
		}
		fmt.Fprintln(detail)
	}

	// Dates
	if !webhook.CreatedAt.IsZero() {
		fmt.Fprintf(detail, "[%s]Created:[-] [%s]%s[-]\n", muted, value, webhook.CreatedAt.Format("Jan 2, 2006 3:04 PM"))
	}
	if !webhook.UpdatedAt.IsZero() {
		fmt.Fprintf(detail, "[%s]Updated:[-] [%s]%s[-]\n", muted, value, webhook.UpdatedAt.Format("Jan 2, 2006 3:04 PM"))
	}

	fmt.Fprintf(detail, "\n\n[%s::d]Press Esc to go back, 'e' to edit, 'd' to delete[-::-]", muted)

	// Handle keyboard
	webhookCopy := webhook
	detail.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			v.app.PopDetail()
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'e':
				v.app.PopDetail()
				v.app.ShowWebhookForm(webhookCopy, func(updatedWebhook *domain.Webhook) {
					v.Refresh()
				})
				return nil
			case 'd':
				v.app.PopDetail()
				v.app.DeleteWebhook(webhookCopy, func() {
					v.Refresh()
				})
				return nil
			}
		}
		return event
	})

	v.app.PushDetail("webhook-detail", detail)
	v.app.SetFocus(detail)
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
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
