package models

import (
	"fmt"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/mqasimca/nylas/internal/tui2/components"
	"github.com/mqasimca/nylas/internal/tui2/state"
	"github.com/mqasimca/nylas/internal/tui2/styles"
)

func TestCalendarAndSchedulePanelAlignment(t *testing.T) {
	// Create a mock global state
	global := &state.GlobalState{
		Theme: "default",
	}
	global.SetWindowSize(120, 40)

	// Create calendar screen
	screen := NewCalendarScreen(global)
	screen.width = 120
	screen.height = 40

	// Set up calendar grid
	theme := styles.GetTheme("default")
	grid := components.NewCalendarGrid(theme)

	// Test with different sizes
	testCases := []struct {
		name       string
		width      int
		height     int
		gridWidth  int
		gridHeight int
	}{
		{"small", 100, 30, 65, 24},
		{"medium", 120, 40, 85, 34},
		{"large", 160, 50, 125, 44},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			screen.width = tc.width
			screen.height = tc.height

			// Set grid size
			grid.SetSize(tc.gridWidth, tc.gridHeight)

			// Render calendar
			calendarView := grid.View()
			calendarLines := len(strings.Split(calendarView, "\n"))

			t.Logf("Calendar grid size: %dx%d, rendered lines: %d", tc.gridWidth, tc.gridHeight, calendarLines)

			// Render schedule panel - now auto-sized based on content
			scheduleWidth := 35
			schedulePanel := screen.renderTodaySchedule(scheduleWidth, calendarLines)
			scheduleLines := len(strings.Split(schedulePanel, "\n"))

			t.Logf("Schedule panel max height: %d, rendered lines: %d", calendarLines, scheduleLines)

			// Schedule panel should be smaller or equal (auto-sized, not forced to match)
			if scheduleLines > calendarLines {
				t.Errorf("Schedule panel too tall: calendar=%d lines, schedule=%d lines",
					calendarLines, scheduleLines)
			}
		})
	}
}

func TestSchedulePanelAutoSize(t *testing.T) {
	global := &state.GlobalState{
		Theme: "default",
	}
	global.SetWindowSize(120, 40)

	screen := NewCalendarScreen(global)
	screen.width = 120
	screen.height = 40

	testCases := []struct {
		maxHeight int
		width     int
	}{
		{20, 35},
		{30, 35},
		{40, 35},
		{50, 35},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("max_%d", tc.maxHeight), func(t *testing.T) {
			panel := screen.renderTodaySchedule(tc.width, tc.maxHeight)
			actualLines := len(strings.Split(panel, "\n"))

			t.Logf("Max height: %d, actual lines: %d", tc.maxHeight, actualLines)

			// Panel should auto-size and not exceed max height
			if actualLines > tc.maxHeight {
				t.Errorf("Schedule panel exceeds max height: max=%d, got=%d",
					tc.maxHeight, actualLines)
			}
			// Panel should have at least the header (minimum ~7 lines with border)
			if actualLines < 7 {
				t.Errorf("Schedule panel too short: got %d lines", actualLines)
			}
		})
	}
}

func TestCountLines(t *testing.T) {
	// Test our line counting method
	testCases := []struct {
		input    string
		expected int
	}{
		{"line1", 1},
		{"line1\n", 2},
		{"line1\nline2", 2},
		{"line1\nline2\n", 3},
		{"line1\nline2\nline3", 3},
		{"\n\n\n", 4},
		{"", 1},
	}

	for _, tc := range testCases {
		lines := len(strings.Split(tc.input, "\n"))
		if lines != tc.expected {
			t.Errorf("countLines(%q) = %d, want %d", tc.input, lines, tc.expected)
		}
	}
}

func TestViewLayoutAlignment(t *testing.T) {
	// This test mimics what View() actually does
	global := &state.GlobalState{
		Theme: "default",
	}

	// Simulate different terminal sizes
	testCases := []struct {
		name   string
		width  int
		height int
	}{
		{"standard", 120, 40},
		{"wide", 160, 40},
		{"tall", 120, 50},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			global.SetWindowSize(tc.width, tc.height)
			screen := NewCalendarScreen(global)
			screen.width = tc.width
			screen.height = tc.height

			// Replicate View() logic
			scheduleWidth := 35
			if tc.width < 100 {
				scheduleWidth = 0
			}
			gridWidth := tc.width - scheduleWidth - 2
			if gridWidth < 50 {
				gridWidth = tc.width
				scheduleWidth = 0
			}

			gridHeight := tc.height - 6
			if gridHeight < 10 {
				gridHeight = 10
			}

			screen.calendarGrid.SetSize(gridWidth, gridHeight)

			// Render calendar
			calendarView := screen.calendarGrid.View()
			calendarLines := len(strings.Split(calendarView, "\n"))

			// Render schedule panel
			if scheduleWidth > 0 {
				schedulePanel := screen.renderTodaySchedule(scheduleWidth, calendarLines)
				scheduleLines := len(strings.Split(schedulePanel, "\n"))

				t.Logf("Window: %dx%d, Grid: %dx%d", tc.width, tc.height, gridWidth, gridHeight)
				t.Logf("Calendar lines: %d, Schedule lines: %d", calendarLines, scheduleLines)

				// Schedule panel should auto-size and be <= calendar lines
				// (lipgloss.Place will center it)
				if scheduleLines > calendarLines {
					t.Errorf("Schedule too tall: calendar=%d, schedule=%d",
						calendarLines, scheduleLines)
				}

				t.Logf("Grid height target: %d, Calendar actual: %d", gridHeight, calendarLines)
			}
		})
	}
}

func TestJoinHorizontalAlignment(t *testing.T) {
	// Test that JoinHorizontal produces expected results with lipgloss.Place
	global := &state.GlobalState{
		Theme: "default",
	}
	global.SetWindowSize(120, 40)

	screen := NewCalendarScreen(global)
	screen.width = 120
	screen.height = 40

	scheduleWidth := 35
	gridWidth := 120 - scheduleWidth - 2
	gridHeight := 40 - 6

	screen.calendarGrid.SetSize(gridWidth, gridHeight)

	calendarView := screen.calendarGrid.View()
	calendarLines := len(strings.Split(calendarView, "\n"))

	schedulePanel := screen.renderTodaySchedule(scheduleWidth, calendarLines)
	scheduleLines := len(strings.Split(schedulePanel, "\n"))

	t.Logf("Before join - Calendar: %d lines, Schedule: %d lines", calendarLines, scheduleLines)

	// Center schedule panel vertically (like View() does)
	centeredSchedule := lipgloss.Place(
		scheduleWidth,
		calendarLines,
		lipgloss.Center,
		lipgloss.Center,
		schedulePanel,
	)
	centeredLines := len(strings.Split(centeredSchedule, "\n"))
	t.Logf("After Place - Centered schedule: %d lines", centeredLines)

	// Join them
	joined := lipgloss.JoinHorizontal(lipgloss.Top, calendarView, centeredSchedule)
	joinedLines := len(strings.Split(joined, "\n"))

	t.Logf("After join - Joined: %d lines", joinedLines)

	// The joined result should have the same number of lines as the calendar
	// (since we used Place to match heights)
	if joinedLines != calendarLines {
		t.Errorf("JoinHorizontal result has %d lines, expected %d (calendar lines)", joinedLines, calendarLines)
	}
}

func TestDebugActualRendering(t *testing.T) {
	// Debug test to see what's actually being rendered
	global := &state.GlobalState{
		Theme: "default",
	}
	// Simulate a larger screen like in the screenshot
	global.SetWindowSize(200, 48)

	screen := NewCalendarScreen(global)
	screen.width = 200
	screen.height = 48

	scheduleWidth := 35
	gridWidth := 200 - scheduleWidth - 2
	gridHeight := 48 - 6 // = 42

	t.Logf("Grid dimensions: %dx%d", gridWidth, gridHeight)

	screen.calendarGrid.SetSize(gridWidth, gridHeight)

	calendarView := screen.calendarGrid.View()
	calendarLinesList := strings.Split(calendarView, "\n")
	t.Logf("Calendar rendered to %d lines", len(calendarLinesList))

	// Show first and last few lines of calendar
	if len(calendarLinesList) > 0 {
		t.Logf("Calendar first line: %q", truncateForLog(calendarLinesList[0], 60))
	}
	if len(calendarLinesList) > 1 {
		t.Logf("Calendar second line: %q", truncateForLog(calendarLinesList[1], 60))
	}
	if len(calendarLinesList) > 2 {
		lastIdx := len(calendarLinesList) - 1
		t.Logf("Calendar last line (%d): %q", lastIdx, truncateForLog(calendarLinesList[lastIdx], 60))
		if lastIdx > 0 {
			t.Logf("Calendar second-to-last line (%d): %q", lastIdx-1, truncateForLog(calendarLinesList[lastIdx-1], 60))
		}
	}

	schedulePanel := screen.renderTodaySchedule(scheduleWidth, len(calendarLinesList))
	scheduleLinesList := strings.Split(schedulePanel, "\n")
	t.Logf("Schedule panel rendered to %d lines (auto-sized)", len(scheduleLinesList))

	// Show first and last few lines of schedule
	if len(scheduleLinesList) > 0 {
		t.Logf("Schedule first line: %q", truncateForLog(scheduleLinesList[0], 60))
	}
	if len(scheduleLinesList) > 1 {
		t.Logf("Schedule second line: %q", truncateForLog(scheduleLinesList[1], 60))
	}
	if len(scheduleLinesList) > 2 {
		lastIdx := len(scheduleLinesList) - 1
		t.Logf("Schedule last line (%d): %q", lastIdx, truncateForLog(scheduleLinesList[lastIdx], 60))
		if lastIdx > 0 {
			t.Logf("Schedule second-to-last line (%d): %q", lastIdx-1, truncateForLog(scheduleLinesList[lastIdx-1], 60))
		}
	}

	// Center the schedule panel (like View() does)
	centeredSchedule := lipgloss.Place(
		scheduleWidth,
		len(calendarLinesList),
		lipgloss.Center,
		lipgloss.Center,
		schedulePanel,
	)
	centeredLinesList := strings.Split(centeredSchedule, "\n")
	t.Logf("Centered schedule has %d lines", len(centeredLinesList))

	// Join and check
	joined := lipgloss.JoinHorizontal(lipgloss.Top, calendarView, centeredSchedule)
	joinedLinesList := strings.Split(joined, "\n")
	t.Logf("Joined content has %d lines", len(joinedLinesList))

	// Check last line of joined content
	if len(joinedLinesList) > 0 {
		lastIdx := len(joinedLinesList) - 1
		t.Logf("Joined last line (%d): %q", lastIdx, truncateForLog(joinedLinesList[lastIdx], 80))
	}

	// Verify centered schedule matches calendar height
	if len(calendarLinesList) != len(centeredLinesList) {
		t.Errorf("MISMATCH: calendar=%d lines, centered schedule=%d lines",
			len(calendarLinesList), len(centeredLinesList))
	}
}

func truncateForLog(s string, maxLen int) string {
	// Strip ANSI codes for cleaner output
	if len(s) > maxLen {
		return s[:maxLen] + "..."
	}
	return s
}

func TestBorderedPanelWithEmptyLines(t *testing.T) {
	// Test that empty lines inside a bordered panel are preserved
	content := "Line 1\nLine 2\n\n\n\nLine 6"
	contentLines := len(strings.Split(content, "\n"))
	t.Logf("Content has %d lines", contentLines) // Should be 6

	style := lipgloss.NewStyle().
		Width(20).
		BorderStyle(lipgloss.RoundedBorder())

	rendered := style.Render(content)
	renderedLines := len(strings.Split(rendered, "\n"))
	t.Logf("Rendered has %d lines", renderedLines) // Should be 8 (6 content + 2 border)

	expectedLines := contentLines + 2 // +2 for top and bottom border
	if renderedLines != expectedLines {
		t.Errorf("Border added wrong number of lines: got %d, want %d", renderedLines, expectedLines)
	}

	// Check that the bordered output contains empty lines in the middle
	lines := strings.Split(rendered, "\n")
	for i, line := range lines {
		t.Logf("Line %d: %q", i, truncateForLog(line, 40))
	}
}

func TestEmptyLinePreservation(t *testing.T) {
	// Create content with trailing empty lines
	lines := []string{"Header", "Content", "", "", "", "", ""}
	content := strings.Join(lines, "\n")

	t.Logf("Input lines: %d", len(lines))
	t.Logf("Joined content lines: %d", len(strings.Split(content, "\n")))

	style := lipgloss.NewStyle().
		Width(20).
		BorderStyle(lipgloss.RoundedBorder())

	rendered := style.Render(content)
	renderedLines := strings.Split(rendered, "\n")
	t.Logf("Rendered lines: %d", len(renderedLines))

	// Print each line
	for i, line := range renderedLines {
		isEmpty := len(strings.TrimSpace(line)) == 0 ||
			strings.TrimSpace(stripANSI(line)) == "│" ||
			strings.TrimSpace(stripANSI(line)) == "│ │"
		t.Logf("Line %d (empty=%v): %q", i, isEmpty, truncateForLog(line, 50))
	}
}

func stripANSI(s string) string {
	// Simple ANSI stripper - removes escape sequences
	result := ""
	inEscape := false
	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		result += string(r)
	}
	return result
}
