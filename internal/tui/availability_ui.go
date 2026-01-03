package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

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
		// Re-validate bounds since participants may have changed before callback
		if idx < 0 || idx >= len(v.participants) {
			return
		}
		// Safe slice removal: handle last element case explicitly
		if idx == len(v.participants)-1 {
			v.participants = v.participants[:idx]
		} else {
			v.participants = append(v.participants[:idx], v.participants[idx+1:]...)
		}
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
