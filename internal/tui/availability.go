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

// AvailabilityView displays free/busy information and helps find meeting times.
type AvailabilityView struct {
	app          *App
	layout       *tview.Flex
	name         string
	title        string
	participants []string
	startDate    time.Time
	endDate      time.Time
	duration     int // in minutes
	slots        []domain.AvailableSlot
	freeBusy     []domain.FreeBusyCalendar

	// Calendar selection for creating events
	calendars          []domain.Calendar
	selectedCalendarID string

	// UI components
	participantsList *tview.List
	slotsList        *tview.List
	timeline         *tview.TextView
	infoPanel        *tview.TextView
	focusedPanel     int // 0=participants, 1=slots, 2=timeline
}

// NewAvailabilityView creates a new availability view.
func NewAvailabilityView(app *App) *AvailabilityView {
	v := &AvailabilityView{
		app:       app,
		name:      "availability",
		title:     "Availability",
		startDate: time.Now(),
		endDate:   time.Now().AddDate(0, 0, 7), // Default 1 week
		duration:  30,                          // Default 30 minutes
	}

	v.init()
	return v
}

func (v *AvailabilityView) init() {
	styles := v.app.styles

	// Participants list
	v.participantsList = tview.NewList()
	v.participantsList.SetBackgroundColor(styles.BgColor)
	v.participantsList.SetMainTextColor(styles.FgColor)
	v.participantsList.SetSecondaryTextColor(styles.InfoColor)
	v.participantsList.SetSelectedBackgroundColor(styles.FocusColor)
	v.participantsList.SetSelectedTextColor(styles.BgColor)
	v.participantsList.SetBorder(true)
	v.participantsList.SetBorderColor(styles.BorderColor)
	v.participantsList.SetTitle(" Participants (a=add, d=delete) ")
	v.participantsList.SetTitleColor(styles.TitleFg)
	v.participantsList.ShowSecondaryText(false)

	// Available slots list
	v.slotsList = tview.NewList()
	v.slotsList.SetBackgroundColor(styles.BgColor)
	v.slotsList.SetMainTextColor(styles.FgColor)
	v.slotsList.SetSecondaryTextColor(styles.InfoColor)
	v.slotsList.SetSelectedBackgroundColor(styles.FocusColor)
	v.slotsList.SetSelectedTextColor(styles.BgColor)
	v.slotsList.SetBorder(true)
	v.slotsList.SetBorderColor(styles.BorderColor)
	v.slotsList.SetTitle(" Available Slots ")
	v.slotsList.SetTitleColor(styles.TitleFg)
	v.slotsList.ShowSecondaryText(true)

	// Handle slot selection to create event
	v.slotsList.SetSelectedFunc(func(index int, _, _ string, _ rune) {
		if index < len(v.slots) {
			v.createEventFromSlot(v.slots[index])
		}
	})

	// Timeline visualization
	v.timeline = tview.NewTextView()
	v.timeline.SetDynamicColors(true)
	v.timeline.SetBackgroundColor(styles.BgColor)
	v.timeline.SetBorder(true)
	v.timeline.SetBorderColor(styles.BorderColor)
	v.timeline.SetTitle(" Timeline (Free/Busy) ")
	v.timeline.SetTitleColor(styles.TitleFg)
	v.timeline.SetBorderPadding(0, 0, 1, 1)

	// Info panel
	v.infoPanel = tview.NewTextView()
	v.infoPanel.SetDynamicColors(true)
	v.infoPanel.SetBackgroundColor(styles.BgColor)
	v.infoPanel.SetBorder(true)
	v.infoPanel.SetBorderColor(styles.BorderColor)
	v.infoPanel.SetTitle(" Settings ")
	v.infoPanel.SetTitleColor(styles.TitleFg)
	v.infoPanel.SetBorderPadding(0, 0, 1, 1)
	v.updateInfoPanel()

	// Layout:
	// Left column: Participants | Info
	// Right column: Timeline | Available Slots
	leftCol := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(v.participantsList, 0, 1, true).
		AddItem(v.infoPanel, 7, 0, false)

	rightCol := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(v.timeline, 0, 1, false).
		AddItem(v.slotsList, 0, 1, false)

	v.layout = tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(leftCol, 35, 0, true).
		AddItem(rightCol, 0, 1, false)

	// Set up input handling
	v.participantsList.SetInputCapture(v.handleParticipantsInput)
	v.slotsList.SetInputCapture(v.handleSlotsInput)
	v.timeline.SetInputCapture(v.handleTimelineInput)
}

func (v *AvailabilityView) Name() string               { return v.name }
func (v *AvailabilityView) Title() string              { return v.title }
func (v *AvailabilityView) Primitive() tview.Primitive { return v.layout }
func (v *AvailabilityView) Filter(string)              {}

func (v *AvailabilityView) Hints() []Hint {
	return []Hint{
		{Key: "a", Desc: "add participant"},
		{Key: "d", Desc: "remove"},
		{Key: "enter", Desc: "create event"},
		{Key: "D", Desc: "set duration"},
		{Key: "S", Desc: "set date range"},
		{Key: "Tab", Desc: "switch panel"},
		{Key: "r", Desc: "refresh"},
	}
}

func (v *AvailabilityView) Load() {
	// Add current user as first participant if empty
	if len(v.participants) == 0 {
		// Try to get current user's email from config
		if v.app.config.GrantID != "" {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			grant, err := v.app.config.Client.GetGrant(ctx, v.app.config.GrantID)
			if err == nil && grant.Email != "" {
				v.participants = append(v.participants, grant.Email)
			}
		}
	}

	// Load calendars for event creation
	v.loadCalendars()

	v.renderParticipants()
	v.updateInfoPanel()
	v.fetchAvailability()
}

func (v *AvailabilityView) loadCalendars() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	calendars, err := v.app.config.Client.GetCalendars(ctx, v.app.config.GrantID)
	if err != nil {
		v.app.Flash(FlashError, "Failed to load calendars: %v", err)
		return
	}

	v.calendars = calendars

	// Select primary calendar by default
	for _, cal := range calendars {
		if cal.IsPrimary {
			v.selectedCalendarID = cal.ID
			return
		}
	}

	// Fall back to first calendar
	if len(calendars) > 0 {
		v.selectedCalendarID = calendars[0].ID
	}
}

func (v *AvailabilityView) Refresh() {
	v.fetchAvailability()
}

func (v *AvailabilityView) updateInfoPanel() {
	styles := v.app.styles
	info := fmt.Sprintf("[%s]Duration:[-] %d min\n", colorToHex(styles.InfoColor), v.duration)
	info += fmt.Sprintf("[%s]Start:[-] %s\n", colorToHex(styles.InfoColor), v.startDate.Format("Jan 2, 2006"))
	info += fmt.Sprintf("[%s]End:[-] %s\n", colorToHex(styles.InfoColor), v.endDate.Format("Jan 2, 2006"))
	info += fmt.Sprintf("[%s]Participants:[-] %d", colorToHex(styles.InfoColor), len(v.participants))
	v.infoPanel.SetText(info)
}

func (v *AvailabilityView) renderParticipants() {
	v.participantsList.Clear()
	for i, email := range v.participants {
		idx := i
		v.participantsList.AddItem(email, "", rune('1'+i), func() {
			// Could show participant details or remove
			_ = idx
		})
	}
}

func (v *AvailabilityView) handleParticipantsInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyTab:
		v.focusedPanel = 1
		v.app.SetFocus(v.slotsList)
		return nil
	case tcell.KeyRune:
		switch event.Rune() {
		case 'a':
			v.showAddParticipantDialog()
			return nil
		case 'd':
			v.removeSelectedParticipant()
			return nil
		case 'D':
			v.showDurationDialog()
			return nil
		case 'S':
			v.showDateRangeDialog()
			return nil
		case 'r':
			v.fetchAvailability()
			return nil
		}
	}
	return event
}

func (v *AvailabilityView) handleSlotsInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyTab:
		v.focusedPanel = 2
		v.app.SetFocus(v.timeline)
		return nil
	case tcell.KeyBacktab:
		v.focusedPanel = 0
		v.app.SetFocus(v.participantsList)
		return nil
	case tcell.KeyRune:
		switch event.Rune() {
		case 'r':
			v.fetchAvailability()
			return nil
		}
	}
	return event
}

func (v *AvailabilityView) handleTimelineInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyTab:
		v.focusedPanel = 0
		v.app.SetFocus(v.participantsList)
		return nil
	case tcell.KeyBacktab:
		v.focusedPanel = 1
		v.app.SetFocus(v.slotsList)
		return nil
	case tcell.KeyRune:
		switch event.Rune() {
		case 'r':
			v.fetchAvailability()
			return nil
		}
	}
	return event
}

func (v *AvailabilityView) showAddParticipantDialog() {
	styles := v.app.styles

	form := tview.NewForm()
	form.SetBackgroundColor(styles.BgColor)
	form.SetFieldBackgroundColor(styles.BgColor)
	form.SetFieldTextColor(styles.FgColor)
	form.SetLabelColor(styles.TitleFg)
	form.SetButtonBackgroundColor(styles.FocusColor)
	form.SetButtonTextColor(styles.BgColor)
	form.SetBorder(true)
	form.SetBorderColor(styles.FocusColor)
	form.SetTitle(" Add Participant ")
	form.SetTitleColor(styles.TitleFg)

	var email string

	form.AddInputField("Email", "", 40, nil, func(text string) {
		email = text
	})

	onClose := func() {
		v.app.content.Pop()
		v.app.SetFocus(v.participantsList)
	}

	form.AddButton("Add", func() {
		email = strings.TrimSpace(email)
		if email == "" {
			v.app.Flash(FlashError, "Email is required")
			return
		}
		if !strings.Contains(email, "@") {
			v.app.Flash(FlashError, "Invalid email address")
			return
		}

		// Check for duplicate
		for _, p := range v.participants {
			if strings.EqualFold(p, email) {
				v.app.Flash(FlashWarn, "Participant already added")
				return
			}
		}

		v.participants = append(v.participants, email)
		v.renderParticipants()
		v.updateInfoPanel()
		onClose()
		v.fetchAvailability()
	})

	form.AddButton("Cancel", onClose)

	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			onClose()
			return nil
		}
		return event
	})

	// Center the form
	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(nil, 0, 1, false).
			AddItem(form, 50, 0, true).
			AddItem(nil, 0, 1, false), 0, 2, true).
		AddItem(nil, 0, 1, false)

	v.app.content.Push("add-participant", flex)
	v.app.SetFocus(form)
}

func (v *AvailabilityView) removeSelectedParticipant() {
	idx := v.participantsList.GetCurrentItem()
	if idx < 0 || idx >= len(v.participants) {
		return
	}

	email := v.participants[idx]
	v.app.ShowConfirmDialog("Remove Participant", fmt.Sprintf("Remove %s?", email), func() {
		v.participants = append(v.participants[:idx], v.participants[idx+1:]...)
		v.renderParticipants()
		v.updateInfoPanel()
		v.fetchAvailability()
	})
}

func (v *AvailabilityView) showDurationDialog() {
	styles := v.app.styles

	form := tview.NewForm()
	form.SetBackgroundColor(styles.BgColor)
	form.SetFieldBackgroundColor(styles.BgColor)
	form.SetFieldTextColor(styles.FgColor)
	form.SetLabelColor(styles.TitleFg)
	form.SetButtonBackgroundColor(styles.FocusColor)
	form.SetButtonTextColor(styles.BgColor)
	form.SetBorder(true)
	form.SetBorderColor(styles.FocusColor)
	form.SetTitle(" Meeting Duration ")
	form.SetTitleColor(styles.TitleFg)

	durations := []string{"15 minutes", "30 minutes", "45 minutes", "60 minutes", "90 minutes", "120 minutes"}
	durationValues := []int{15, 30, 45, 60, 90, 120}

	// Find current selection
	currentIdx := 1 // default to 30 min
	for i, d := range durationValues {
		if d == v.duration {
			currentIdx = i
			break
		}
	}

	var selectedDuration int

	form.AddDropDown("Duration", durations, currentIdx, func(option string, index int) {
		if index < len(durationValues) {
			selectedDuration = durationValues[index]
		}
	})

	onClose := func() {
		v.app.content.Pop()
		v.app.SetFocus(v.participantsList)
	}

	form.AddButton("Set", func() {
		if selectedDuration > 0 {
			v.duration = selectedDuration
			v.updateInfoPanel()
			onClose()
			v.fetchAvailability()
		}
	})

	form.AddButton("Cancel", onClose)

	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			onClose()
			return nil
		}
		return event
	})

	// Center the form
	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(nil, 0, 1, false).
			AddItem(form, 40, 0, true).
			AddItem(nil, 0, 1, false), 0, 2, true).
		AddItem(nil, 0, 1, false)

	v.app.content.Push("duration-dialog", flex)
	v.app.SetFocus(form)
}

func (v *AvailabilityView) showDateRangeDialog() {
	styles := v.app.styles

	form := tview.NewForm()
	form.SetBackgroundColor(styles.BgColor)
	form.SetFieldBackgroundColor(styles.BgColor)
	form.SetFieldTextColor(styles.FgColor)
	form.SetLabelColor(styles.TitleFg)
	form.SetButtonBackgroundColor(styles.FocusColor)
	form.SetButtonTextColor(styles.BgColor)
	form.SetBorder(true)
	form.SetBorderColor(styles.FocusColor)
	form.SetTitle(" Date Range ")
	form.SetTitleColor(styles.TitleFg)

	var startStr, endStr string

	form.AddInputField("Start Date (YYYY-MM-DD)", v.startDate.Format("2006-01-02"), 15, nil, func(text string) {
		startStr = text
	})

	form.AddInputField("End Date (YYYY-MM-DD)", v.endDate.Format("2006-01-02"), 15, nil, func(text string) {
		endStr = text
	})

	onClose := func() {
		v.app.content.Pop()
		v.app.SetFocus(v.participantsList)
	}

	form.AddButton("Set", func() {
		start, err := time.Parse("2006-01-02", startStr)
		if err != nil {
			v.app.Flash(FlashError, "Invalid start date format")
			return
		}

		end, err := time.Parse("2006-01-02", endStr)
		if err != nil {
			v.app.Flash(FlashError, "Invalid end date format")
			return
		}

		if end.Before(start) {
			v.app.Flash(FlashError, "End date must be after start date")
			return
		}

		v.startDate = start
		v.endDate = end
		v.updateInfoPanel()
		onClose()
		v.fetchAvailability()
	})

	form.AddButton("Cancel", onClose)

	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			onClose()
			return nil
		}
		return event
	})

	// Center the form
	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(nil, 0, 1, false).
			AddItem(form, 45, 0, true).
			AddItem(nil, 0, 1, false), 0, 2, true).
		AddItem(nil, 0, 1, false)

	v.app.content.Push("date-range-dialog", flex)
	v.app.SetFocus(form)
}

func (v *AvailabilityView) fetchAvailability() {
	if len(v.participants) == 0 {
		v.timeline.SetText("[gray]Add participants to check availability[-]")
		v.slotsList.Clear()
		v.slots = nil
		return
	}

	v.timeline.SetText("[gray]Loading availability...[-]")
	v.slotsList.Clear()

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Build participants list
		participants := make([]domain.AvailabilityParticipant, len(v.participants))
		for i, email := range v.participants {
			participants[i] = domain.AvailabilityParticipant{
				Email: email,
			}
		}

		req := &domain.AvailabilityRequest{
			StartTime:       v.startDate.Unix(),
			EndTime:         v.endDate.Add(24 * time.Hour).Unix(), // Include full end day
			DurationMinutes: v.duration,
			Participants:    participants,
			IntervalMinutes: 15,
		}

		resp, err := v.app.config.Client.GetAvailability(ctx, req)
		if err != nil {
			v.app.QueueUpdateDraw(func() {
				v.timeline.SetText(fmt.Sprintf("[red]Failed to load availability: %v[-]", err))
			})
			return
		}

		// Also fetch free/busy for timeline visualization
		freeBusyReq := &domain.FreeBusyRequest{
			StartTime: v.startDate.Unix(),
			EndTime:   v.endDate.Add(24 * time.Hour).Unix(),
			Emails:    v.participants,
		}

		freeBusyResp, _ := v.app.config.Client.GetFreeBusy(ctx, v.app.config.GrantID, freeBusyReq)

		v.app.QueueUpdateDraw(func() {
			v.slots = resp.Data.TimeSlots
			if freeBusyResp != nil {
				v.freeBusy = freeBusyResp.Data
			}
			v.renderTimeline()
			v.renderSlots()
		})
	}()
}

func (v *AvailabilityView) renderTimeline() {
	styles := v.app.styles
	var content strings.Builder

	if len(v.freeBusy) == 0 {
		content.WriteString("[gray]No free/busy data available[-]\n")
		v.timeline.SetText(content.String())
		return
	}

	// Display busy times for each participant
	for _, fb := range v.freeBusy {
		fmt.Fprintf(&content, "[%s]%s[-]\n", colorToHex(styles.TitleFg), fb.Email)

		if len(fb.TimeSlots) == 0 {
			content.WriteString("  [green]All free[-]\n")
		} else {
			// Group by day
			slotsByDay := make(map[string][]domain.TimeSlot)
			for _, slot := range fb.TimeSlots {
				day := time.Unix(slot.StartTime, 0).Format("Jan 2")
				slotsByDay[day] = append(slotsByDay[day], slot)
			}

			for day, slots := range slotsByDay {
				fmt.Fprintf(&content, "  [%s]%s:[-]", colorToHex(styles.InfoColor), day)
				for _, slot := range slots {
					start := time.Unix(slot.StartTime, 0).Local()
					end := time.Unix(slot.EndTime, 0).Local()
					fmt.Fprintf(&content, " [red]%s-%s[-]", start.Format("3:04PM"), end.Format("3:04PM"))
				}
				content.WriteString("\n")
			}
		}
		content.WriteString("\n")
	}

	// Add legend
	fmt.Fprintf(&content, "[%s]Legend: [-][red]Busy[-] [green]Free[-]\n", colorToHex(styles.BorderColor))

	v.timeline.SetText(content.String())
}

func (v *AvailabilityView) renderSlots() {
	v.slotsList.Clear()

	if len(v.slots) == 0 {
		v.slotsList.AddItem("[No available slots found]", "", 0, nil)
		return
	}

	// Group slots by day
	slotsByDay := make(map[string][]domain.AvailableSlot)
	for _, slot := range v.slots {
		day := time.Unix(slot.StartTime, 0).Local().Format("Jan 2, 2006")
		slotsByDay[day] = append(slotsByDay[day], slot)
	}

	// Sort days and display
	count := 0
	for _, slot := range v.slots {
		if count >= 20 {
			// Limit displayed slots
			v.slotsList.AddItem(fmt.Sprintf("... and %d more slots", len(v.slots)-20), "", 0, nil)
			break
		}

		start := time.Unix(slot.StartTime, 0).Local()
		end := time.Unix(slot.EndTime, 0).Local()

		mainText := fmt.Sprintf("%s %s - %s",
			start.Format("Mon, Jan 2"),
			start.Format("3:04 PM"),
			end.Format("3:04 PM"))

		secondaryText := fmt.Sprintf("%d min", v.duration)
		if len(slot.Emails) > 0 {
			secondaryText += " | " + strings.Join(slot.Emails, ", ")
		}

		idx := count
		v.slotsList.AddItem(mainText, secondaryText, 0, func() {
			if idx < len(v.slots) {
				v.createEventFromSlot(v.slots[idx])
			}
		})
		count++
	}
}

func (v *AvailabilityView) createEventFromSlot(slot domain.AvailableSlot) {
	start := time.Unix(slot.StartTime, 0).Local()
	end := time.Unix(slot.EndTime, 0).Local()

	// Create a new event with the selected time slot
	event := &domain.Event{
		When: domain.EventWhen{
			StartTime: slot.StartTime,
			EndTime:   slot.EndTime,
		},
	}

	// Add participants
	for _, email := range v.participants {
		event.Participants = append(event.Participants, domain.Participant{
			Email:  email,
			Status: "noreply",
		})
	}

	v.app.ShowConfirmDialog("Create Event",
		fmt.Sprintf("Create meeting on %s %s - %s?", start.Format("Mon, Jan 2"), start.Format("3:04 PM"), end.Format("3:04 PM")),
		func() {
			// Show event form with pre-filled data
			if v.selectedCalendarID == "" {
				v.app.Flash(FlashError, "No calendar selected")
				return
			}

			form := NewEventForm(v.app, v.selectedCalendarID, event,
				func(e *domain.Event) {
					v.app.content.Pop()
					v.app.Flash(FlashInfo, "Event created successfully")
				},
				func() {
					v.app.content.Pop()
				})
			v.app.content.Push("event-form", form)
			v.app.SetFocus(form)
		})
}

func (v *AvailabilityView) HandleKey(event *tcell.EventKey) *tcell.EventKey {
	// Global key handling
	switch event.Key() {
	case tcell.KeyRune:
		switch event.Rune() {
		case 'r':
			v.fetchAvailability()
			return nil
		}
	}
	return event
}
