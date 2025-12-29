// Package components provides reusable Bubble Tea components.
package components

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/mqasimca/nylas/internal/domain"
)

func (c *CalendarGrid) View() string {
	switch c.viewMode {
	case CalendarWeekView:
		return c.renderWeekView()
	case CalendarAgendaView:
		return c.renderAgendaView()
	default:
		return c.renderMonthView()
	}
}

// renderMonthView renders the month grid with clean cells like Google Calendar.
func (c *CalendarGrid) renderMonthView() string {
	var b strings.Builder

	// Calculate dimensions
	numCols := 7
	if !c.showWeekends {
		numCols = 5
	}

	// Cell dimensions - use full available width
	cellWidth := c.width / numCols
	if cellWidth < 10 {
		cellWidth = 10
	}

	// Calculate cell height based on available space
	// Reserve: header (2), day names (1), bottom margin (1)
	reservedLines := 4
	availableHeight := c.height - reservedLines
	days := c.getMonthDays()
	numWeeks := (len(days) + numCols - 1) / numCols
	if numWeeks == 0 {
		numWeeks = 1
	}
	cellHeight := availableHeight / numWeeks
	if cellHeight < 4 {
		cellHeight = 4
	}
	if cellHeight > 7 {
		cellHeight = 7
	}

	// Header with month/year and navigation hints
	header := fmt.Sprintf("← %s →", c.currentMonth.Format("January 2006"))
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(c.theme.Primary).
		Width(c.width).
		Align(lipgloss.Center)
	b.WriteString(headerStyle.Render(header))
	b.WriteString("\n\n")

	// Day names header
	dayNames := c.getDayNamesShort()
	dayHeaderStyle := lipgloss.NewStyle().
		Foreground(c.theme.Secondary).
		Width(cellWidth).
		Align(lipgloss.Center).
		Bold(true)

	for _, name := range dayNames {
		b.WriteString(dayHeaderStyle.Render(name))
	}
	b.WriteString("\n")

	// Generate calendar grid
	for week := 0; week < numWeeks; week++ {
		// Build each line of the week row
		weekCells := make([]string, numCols)
		for day := 0; day < numCols; day++ {
			idx := week*numCols + day
			if idx < len(days) {
				weekCells[day] = c.renderCleanCell(days[idx], cellWidth, cellHeight)
			} else {
				weekCells[day] = c.renderEmptyCellClean(cellWidth, cellHeight)
			}
		}
		// Join cells horizontally
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, weekCells...))
		// Only add newline between rows, not after the last row
		if week < numWeeks-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

// getDayNamesShort returns short day names for the header.
func (c *CalendarGrid) getDayNamesShort() []string {
	if c.firstDayMon {
		if c.showWeekends {
			return []string{"MON", "TUE", "WED", "THU", "FRI", "SAT", "SUN"}
		}
		return []string{"MON", "TUE", "WED", "THU", "FRI"}
	}
	if c.showWeekends {
		return []string{"SUN", "MON", "TUE", "WED", "THU", "FRI", "SAT"}
	}
	return []string{"MON", "TUE", "WED", "THU", "FRI"}
}

// renderEmptyCellClean renders an empty cell without content.
func (c *CalendarGrid) renderEmptyCellClean(width, height int) string {
	cellStyle := lipgloss.NewStyle().
		Width(width).
		Height(height).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(c.theme.Dimmed.GetForeground())

	return cellStyle.Render("")
}

// renderCleanCell renders a single day cell with clean formatting.
func (c *CalendarGrid) renderCleanCell(date time.Time, width, height int) string {
	isSelected := c.IsSelected(date)
	isToday := c.IsToday(date)
	isCurrentMonth := c.IsCurrentMonth(date)
	events := c.GetEventsForDate(date)

	// Cell style with border
	cellStyle := lipgloss.NewStyle().
		Width(width).
		Height(height).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(c.theme.Dimmed.GetForeground()).
		Padding(0, 1)

	// Selected cell gets highlighted background
	if isSelected {
		cellStyle = cellStyle.
			Background(c.theme.Primary).
			BorderForeground(c.theme.Primary)
	}

	var content strings.Builder
	contentWidth := width - 4 // Account for border and padding

	// Line 1: Day number
	dayNum := fmt.Sprintf("%d", date.Day())
	dayStyle := lipgloss.NewStyle()

	if isToday {
		dayStyle = dayStyle.Bold(true).Foreground(c.theme.Success)
		if isSelected {
			dayStyle = dayStyle.Background(c.theme.Success).Foreground(lipgloss.Color("#000000"))
		}
	} else if !isCurrentMonth {
		dayStyle = dayStyle.Foreground(c.theme.Dimmed.GetForeground())
	} else if isSelected {
		dayStyle = dayStyle.Foreground(lipgloss.Color("#000000")).Bold(true)
	}

	content.WriteString(dayStyle.Render(dayNum))
	content.WriteString("\n")

	// Line 2: Event dots
	if len(events) > 0 {
		dots := c.renderEventDotsClean(events, contentWidth, isSelected)
		content.WriteString(dots)
		content.WriteString("\n")

		// Line 3+: Event titles (up to 2)
		maxTitles := height - 3
		if maxTitles > 2 {
			maxTitles = 2
		}
		for i := 0; i < maxTitles && i < len(events); i++ {
			title := events[i].Title
			if title == "" {
				title = "(No title)"
			}
			// Truncate title to fit
			if len(title) > contentWidth-1 {
				title = title[:contentWidth-2] + "…"
			}

			titleStyle := lipgloss.NewStyle()
			if isSelected {
				titleStyle = titleStyle.Foreground(lipgloss.Color("#000000"))
			} else {
				titleStyle = titleStyle.Foreground(c.theme.Secondary)
			}
			content.WriteString(titleStyle.Render(title))
			if i < maxTitles-1 && i < len(events)-1 {
				content.WriteString("\n")
			}
		}
	}

	return cellStyle.Render(content.String())
}

// renderEventDotsClean renders colored dots for events.
func (c *CalendarGrid) renderEventDotsClean(events []domain.Event, maxWidth int, isSelected bool) string {
	var dots strings.Builder

	maxDots := min(len(events), maxWidth/2)
	if maxDots > 5 {
		maxDots = 5
	}

	for i := 0; i < maxDots; i++ {
		evt := events[i]
		dotColor := c.theme.Primary
		if evt.Status == "cancelled" {
			dotColor = c.theme.Error
		} else if !evt.Busy {
			dotColor = c.theme.Success
		}

		dotStyle := lipgloss.NewStyle().Foreground(dotColor)
		if isSelected {
			dotStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
		}
		dots.WriteString(dotStyle.Render("●"))
		if i < maxDots-1 {
			dots.WriteString(" ")
		}
	}

	// Show "+N" if more events
	if len(events) > maxDots {
		moreStyle := lipgloss.NewStyle().Foreground(c.theme.Secondary)
		if isSelected {
			moreStyle = moreStyle.Foreground(lipgloss.Color("#000000"))
		}
		dots.WriteString(moreStyle.Render(fmt.Sprintf("+%d", len(events)-maxDots)))
	}

	return dots.String()
}

// renderWeekView renders a Google Calendar-style week view with time slots.
