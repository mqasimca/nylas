package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// CalendarViewMode represents the calendar display mode.

func (c *CalendarView) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return c.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		switch event.Key() {
		case tcell.KeyLeft:
			c.selectedDate = c.selectedDate.AddDate(0, 0, -1)
			c.updateCurrentMonth()
		case tcell.KeyRight:
			c.selectedDate = c.selectedDate.AddDate(0, 0, 1)
			c.updateCurrentMonth()
		case tcell.KeyUp:
			c.selectedDate = c.selectedDate.AddDate(0, 0, -7)
			c.updateCurrentMonth()
		case tcell.KeyDown:
			c.selectedDate = c.selectedDate.AddDate(0, 0, 7)
			c.updateCurrentMonth()
		case tcell.KeyPgUp:
			c.PrevMonth()
		case tcell.KeyPgDn:
			c.NextMonth()
		case tcell.KeyEnter:
			if c.onDateSelect != nil {
				c.onDateSelect(c.selectedDate)
			}
		case tcell.KeyRune:
			switch event.Rune() {
			case 'h':
				c.selectedDate = c.selectedDate.AddDate(0, 0, -1)
				c.updateCurrentMonth()
			case 'l':
				c.selectedDate = c.selectedDate.AddDate(0, 0, 1)
				c.updateCurrentMonth()
			case 'k':
				c.selectedDate = c.selectedDate.AddDate(0, 0, -7)
				c.updateCurrentMonth()
			case 'j':
				c.selectedDate = c.selectedDate.AddDate(0, 0, 7)
				c.updateCurrentMonth()
			case 'H':
				c.PrevMonth()
			case 'L':
				c.NextMonth()
			case 't':
				c.GoToToday()
			case 'm':
				c.SetViewMode(CalendarMonthView)
			case 'w':
				c.SetViewMode(CalendarWeekView)
			case 'a':
				c.SetViewMode(CalendarAgendaView)
			case 'c':
				c.NextCalendar()
			case 'C':
				c.PrevCalendar()
			}
		}
	})
}

// MouseHandler returns the mouse handler for the calendar.
func (c *CalendarView) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return c.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
		if !c.InRect(event.Position()) {
			return false, nil
		}

		switch action {
		case tview.MouseLeftClick:
			setFocus(c)
			x, y, _, _ := c.GetInnerRect()
			mouseX, mouseY := event.Position()

			if c.viewMode == CalendarMonthView {
				// Calculate which day was clicked
				headerOffset := 4 // Header + day names + separator
				if mouseY < y+headerOffset {
					// Check if clicked on navigation
					return true, nil
				}

				// Calculate clicked cell
				col := (mouseX - x) / c.cellWidth
				row := (mouseY - y - headerOffset) / c.cellHeight

				if col >= 0 && col < 7 && row >= 0 && row < 6 {
					// Calculate the date
					firstDay := c.currentMonth
					firstWeekday := int(firstDay.Weekday())
					startDate := firstDay.AddDate(0, 0, -firstWeekday)
					clickedDate := startDate.AddDate(0, 0, row*7+col)

					c.selectedDate = clickedDate
					c.updateCurrentMonth()

					if c.onDateSelect != nil {
						c.onDateSelect(c.selectedDate)
					}
				}
			}
			return true, nil

		case tview.MouseLeftDoubleClick:
			if c.onDateSelect != nil {
				c.onDateSelect(c.selectedDate)
			}
			return true, nil

		case tview.MouseScrollUp:
			c.PrevMonth()
			return true, nil

		case tview.MouseScrollDown:
			c.NextMonth()
			return true, nil
		}

		return false, nil
	})
}

// Focus is called when this primitive receives focus.
func (c *CalendarView) Focus(delegate func(p tview.Primitive)) {
	c.Box.Focus(delegate)
}

// HasFocus returns whether or not this primitive has focus.
func (c *CalendarView) HasFocus() bool {
	return c.Box.HasFocus()
}
