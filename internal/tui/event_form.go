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

// EventFormMode indicates if we're creating or editing.
type EventFormMode int

const (
	EventFormCreate EventFormMode = iota
	EventFormEdit
)

// EventForm provides a form for creating/editing events.
type EventForm struct {
	*tview.Flex
	app        *App
	form       *tview.Form
	mode       EventFormMode
	event      *domain.Event // nil for create, populated for edit
	calendarID string
	onSubmit   func(*domain.Event)
	onCancel   func()

	// Form field values
	title       string
	description string
	location    string
	startDate   string
	startTime   string
	endDate     string
	endTime     string
	allDay      bool
	busy        bool
}

// NewEventForm creates a new event form.
func NewEventForm(app *App, calendarID string, event *domain.Event, onSubmit func(*domain.Event), onCancel func()) *EventForm {
	mode := EventFormCreate
	if event != nil {
		mode = EventFormEdit
	}

	f := &EventForm{
		Flex:       tview.NewFlex(),
		app:        app,
		mode:       mode,
		event:      event,
		calendarID: calendarID,
		onSubmit:   onSubmit,
		onCancel:   onCancel,
		busy:       true, // Default to busy
	}

	// Populate from existing event if editing
	if event != nil {
		f.title = event.Title
		f.description = event.Description
		f.location = event.Location
		f.allDay = event.When.IsAllDay()
		f.busy = event.Busy

		if f.allDay {
			if event.When.Date != "" {
				f.startDate = event.When.Date
				f.endDate = event.When.Date
			} else if event.When.StartDate != "" {
				f.startDate = event.When.StartDate
				f.endDate = event.When.EndDate
			}
		} else {
			start := event.When.StartDateTime()
			end := event.When.EndDateTime()
			f.startDate = start.Format("2006-01-02")
			f.startTime = start.Format("15:04")
			f.endDate = end.Format("2006-01-02")
			f.endTime = end.Format("15:04")
		}
	} else {
		// Default to today for new events
		now := time.Now()
		f.startDate = now.Format("2006-01-02")
		f.endDate = now.Format("2006-01-02")
		f.startTime = now.Add(time.Hour).Format("15:04")
		f.endTime = now.Add(2 * time.Hour).Format("15:04")
	}

	f.init()
	return f
}

func (f *EventForm) init() {
	styles := f.app.styles

	f.form = tview.NewForm()
	f.form.SetBackgroundColor(styles.BgColor)
	f.form.SetFieldBackgroundColor(styles.BgColor)
	f.form.SetFieldTextColor(styles.FgColor)
	f.form.SetLabelColor(styles.TitleFg)
	f.form.SetButtonBackgroundColor(styles.FocusColor)
	f.form.SetButtonTextColor(styles.BgColor)
	f.form.SetBorder(true)
	f.form.SetBorderColor(styles.FocusColor)

	title := "New Event"
	if f.mode == EventFormEdit {
		title = "Edit Event"
	}
	f.form.SetTitle(fmt.Sprintf(" %s ", title))
	f.form.SetTitleColor(styles.TitleFg)

	// Add form fields
	f.form.AddInputField("Title", f.title, 40, nil, func(text string) {
		f.title = text
	})

	f.form.AddTextArea("Description", f.description, 40, 3, 0, func(text string) {
		f.description = text
	})

	f.form.AddInputField("Location", f.location, 40, nil, func(text string) {
		f.location = text
	})

	f.form.AddCheckbox("All Day", f.allDay, func(checked bool) {
		f.allDay = checked
		f.updateTimeFields()
	})

	f.form.AddInputField("Start Date (YYYY-MM-DD)", f.startDate, 15, nil, func(text string) {
		f.startDate = text
	})

	f.form.AddInputField("Start Time (HH:MM)", f.startTime, 10, nil, func(text string) {
		f.startTime = text
	})

	f.form.AddInputField("End Date (YYYY-MM-DD)", f.endDate, 15, nil, func(text string) {
		f.endDate = text
	})

	f.form.AddInputField("End Time (HH:MM)", f.endTime, 10, nil, func(text string) {
		f.endTime = text
	})

	f.form.AddCheckbox("Busy", f.busy, func(checked bool) {
		f.busy = checked
	})

	// Add buttons
	f.form.AddButton("Save", f.submit)
	f.form.AddButton("Cancel", f.cancel)

	// Set up key capture
	f.form.SetInputCapture(f.handleInput)

	// Update time field visibility based on all-day
	f.updateTimeFields()

	// Center the form
	f.SetDirection(tview.FlexRow)
	f.AddItem(nil, 0, 1, false)
	f.AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(nil, 0, 1, false).
		AddItem(f.form, 60, 0, true).
		AddItem(nil, 0, 1, false), 0, 3, true)
	f.AddItem(nil, 0, 1, false)
}

func (f *EventForm) updateTimeFields() {
	// Note: tview doesn't support dynamic show/hide of form items easily
	// We'll validate based on allDay flag during submit
}

func (f *EventForm) handleInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEscape:
		f.cancel()
		return nil
	case tcell.KeyCtrlS:
		f.submit()
		return nil
	}
	return event
}

func (f *EventForm) validate() []string {
	var errors []string

	if strings.TrimSpace(f.title) == "" {
		errors = append(errors, "Title is required")
	}

	if strings.TrimSpace(f.startDate) == "" {
		errors = append(errors, "Start date is required")
	} else if _, err := time.Parse("2006-01-02", f.startDate); err != nil {
		errors = append(errors, "Start date must be YYYY-MM-DD format")
	}

	if !f.allDay {
		if strings.TrimSpace(f.startTime) == "" {
			errors = append(errors, "Start time is required for non all-day events")
		} else if _, err := time.Parse("15:04", f.startTime); err != nil {
			errors = append(errors, "Start time must be HH:MM format")
		}

		if strings.TrimSpace(f.endTime) == "" {
			errors = append(errors, "End time is required for non all-day events")
		} else if _, err := time.Parse("15:04", f.endTime); err != nil {
			errors = append(errors, "End time must be HH:MM format")
		}
	}

	if strings.TrimSpace(f.endDate) == "" {
		errors = append(errors, "End date is required")
	} else if _, err := time.Parse("2006-01-02", f.endDate); err != nil {
		errors = append(errors, "End date must be YYYY-MM-DD format")
	}

	return errors
}

func (f *EventForm) submit() {
	// Validate
	errors := f.validate()
	if len(errors) > 0 {
		f.app.Flash(FlashError, "%s", strings.Join(errors, "; "))
		return
	}

	// Build the event when
	var when domain.EventWhen
	if f.allDay {
		if f.startDate == f.endDate {
			when = domain.EventWhen{
				Object: "date",
				Date:   f.startDate,
			}
		} else {
			when = domain.EventWhen{
				Object:    "datespan",
				StartDate: f.startDate,
				EndDate:   f.endDate,
			}
		}
	} else {
		// Parse times
		startDateTime, _ := time.ParseInLocation("2006-01-02 15:04", f.startDate+" "+f.startTime, time.Local)
		endDateTime, _ := time.ParseInLocation("2006-01-02 15:04", f.endDate+" "+f.endTime, time.Local)

		when = domain.EventWhen{
			Object:    "timespan",
			StartTime: startDateTime.Unix(),
			EndTime:   endDateTime.Unix(),
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var resultEvent *domain.Event
	var err error

	if f.mode == EventFormCreate {
		// Create new event
		req := &domain.CreateEventRequest{
			Title:       strings.TrimSpace(f.title),
			Description: strings.TrimSpace(f.description),
			Location:    strings.TrimSpace(f.location),
			When:        when,
			Busy:        f.busy,
		}
		resultEvent, err = f.app.config.Client.CreateEvent(ctx, f.app.config.GrantID, f.calendarID, req)
	} else {
		// Update existing event
		title := strings.TrimSpace(f.title)
		desc := strings.TrimSpace(f.description)
		loc := strings.TrimSpace(f.location)

		req := &domain.UpdateEventRequest{
			Title:       &title,
			Description: &desc,
			Location:    &loc,
			When:        &when,
			Busy:        &f.busy,
		}
		resultEvent, err = f.app.config.Client.UpdateEvent(ctx, f.app.config.GrantID, f.calendarID, f.event.ID, req)
	}

	if err != nil {
		f.app.Flash(FlashError, "Failed to save event: %v", err)
		return
	}

	if f.mode == EventFormCreate {
		f.app.Flash(FlashInfo, "Event created: %s", resultEvent.Title)
	} else {
		f.app.Flash(FlashInfo, "Event updated: %s", resultEvent.Title)
	}

	if f.onSubmit != nil {
		f.onSubmit(resultEvent)
	}
}

func (f *EventForm) cancel() {
	if f.onCancel != nil {
		f.onCancel()
	}
}

// Focus sets focus to the form.
func (f *EventForm) Focus(delegate func(p tview.Primitive)) {
	delegate(f.form)
}

// ShowEventForm displays an event form for create/edit.
func (a *App) ShowEventForm(calendarID string, event *domain.Event, onSave func(*domain.Event)) {
	onClose := func() {
		a.content.Pop()
		if view := a.getCurrentView(); view != nil {
			a.SetFocus(view.Primitive())
		}
	}

	form := NewEventForm(a, calendarID, event, func(savedEvent *domain.Event) {
		onClose()
		if onSave != nil {
			onSave(savedEvent)
		}
	}, onClose)

	a.content.Push("event-form", form)
	a.SetFocus(form)
}

// DeleteEvent shows a confirmation dialog and deletes an event.
func (a *App) DeleteEvent(calendarID string, event *domain.Event, onDelete func()) {
	a.ShowConfirmDialog("Delete Event", fmt.Sprintf("Delete event '%s'?", event.Title), func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err := a.config.Client.DeleteEvent(ctx, a.config.GrantID, calendarID, event.ID)
		if err != nil {
			a.Flash(FlashError, "Failed to delete event: %v", err)
			return
		}

		a.Flash(FlashInfo, "Event deleted: %s", event.Title)
		if onDelete != nil {
			onDelete()
		}
	})
}
