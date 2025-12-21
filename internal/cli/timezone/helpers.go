package timezone

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mqasimca/nylas/internal/adapters/utilities/timezone"
)

// getService creates a new timezone service.
func getService() *timezone.Service {
	return timezone.NewService()
}

// createContext creates a context with timeout.
func createContext() context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	_ = cancel // Will be called when function returns
	return ctx
}

// formatTime formats a time with time zone information.
func formatTime(t time.Time, showZone bool) string {
	if showZone {
		return fmt.Sprintf("%s (%s)", t.Format("2006-01-02 15:04:05"), t.Format("MST"))
	}
	return t.Format("2006-01-02 15:04:05")
}

// printJSON prints data as formatted JSON.
func printJSON(data any) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// parseTimeZones parses a comma-separated list of time zones.
func parseTimeZones(zonesStr string) []string {
	if zonesStr == "" {
		return []string{}
	}

	zones := strings.Split(zonesStr, ",")
	result := make([]string, 0, len(zones))

	for _, zone := range zones {
		trimmed := strings.TrimSpace(zone)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

// parseDuration parses a duration string (e.g., "1h", "30m", "1h30m").
func parseDuration(durationStr string) (time.Duration, error) {
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return 0, fmt.Errorf("invalid duration format (use 1h, 30m, 1h30m, etc.): %w", err)
	}
	return duration, nil
}

// parseWorkingHours parses working hours in HH:MM format.
func parseWorkingHours(start, end string) (string, string, error) {
	// Validate format
	if _, err := time.Parse("15:04", start); err != nil {
		return "", "", fmt.Errorf("invalid start time format (use HH:MM): %w", err)
	}

	if _, err := time.Parse("15:04", end); err != nil {
		return "", "", fmt.Errorf("invalid end time format (use HH:MM): %w", err)
	}

	return start, end, nil
}

// printTable prints a simple table.
func printTable(headers []string, rows [][]string) {
	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}

	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Print header
	printRow(headers, widths)
	printSeparator(widths)

	// Print rows
	for _, row := range rows {
		printRow(row, widths)
	}
}

// printRow prints a single row with proper padding.
func printRow(cells []string, widths []int) {
	for i, cell := range cells {
		if i < len(widths) {
			fmt.Printf("%-*s  ", widths[i], cell)
		}
	}
	fmt.Println()
}

// printSeparator prints a separator line.
func printSeparator(widths []int) {
	for i, width := range widths {
		fmt.Print(strings.Repeat("-", width))
		if i < len(widths)-1 {
			fmt.Print("  ")
		}
	}
	fmt.Println()
}

// formatOffset formats a UTC offset in seconds to a readable string.
func formatOffset(seconds int) string {
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60

	// Handle negative offsets properly
	if minutes < 0 {
		minutes = -minutes
	}

	if minutes == 0 {
		return fmt.Sprintf("UTC%+d", hours)
	}
	return fmt.Sprintf("UTC%+d:%02d", hours, minutes)
}

// normalizeTimeZone attempts to normalize common time zone abbreviations to IANA names.
func normalizeTimeZone(zone string) string {
	// Map of common abbreviations to IANA names
	abbrevMap := map[string]string{
		"PST":  "America/Los_Angeles",
		"PDT":  "America/Los_Angeles",
		"EST":  "America/New_York",
		"EDT":  "America/New_York",
		"CST":  "America/Chicago",
		"CDT":  "America/Chicago",
		"MST":  "America/Denver",
		"MDT":  "America/Denver",
		"GMT":  "Europe/London",
		"BST":  "Europe/London",
		"IST":  "Asia/Kolkata",
		"JST":  "Asia/Tokyo",
		"AEST": "Australia/Sydney",
		"AEDT": "Australia/Sydney",
	}

	if iana, ok := abbrevMap[strings.ToUpper(zone)]; ok {
		return iana
	}

	return zone
}
