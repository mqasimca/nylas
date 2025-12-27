package timezone

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/mqasimca/nylas/internal/domain"
)

// Service implements ports.TimeZoneService.
// Provides time zone conversion, meeting finder, and DST transition utilities.
type Service struct{}

// NewService creates a new time zone service.
func NewService() *Service {
	return &Service{}
}

// ConvertTime converts a time from one zone to another.
func (s *Service) ConvertTime(ctx context.Context, fromZone, toZone string, t time.Time) (time.Time, error) {
	fromLoc, err := time.LoadLocation(fromZone)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid from zone %q: %w", fromZone, err)
	}

	toLoc, err := time.LoadLocation(toZone)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid to zone %q: %w", toZone, err)
	}

	// Convert time to source zone, then to target zone
	timeInFrom := t.In(fromLoc)
	timeInTo := timeInFrom.In(toLoc)

	return timeInTo, nil
}

// FindMeetingTime finds overlapping working hours across multiple time zones.
// Analyzes each time zone's working hours and finds slots where all zones overlap.
func (s *Service) FindMeetingTime(ctx context.Context, req *domain.MeetingFinderRequest) (*domain.MeetingTimeSlots, error) {
	if len(req.TimeZones) == 0 {
		return nil, fmt.Errorf("at least one time zone required")
	}

	// Parse working hours
	startHour, err := parseTime(req.WorkingHoursStart)
	if err != nil {
		return nil, fmt.Errorf("invalid working hours start: %w", err)
	}

	endHour, err := parseTime(req.WorkingHoursEnd)
	if err != nil {
		return nil, fmt.Errorf("invalid working hours end: %w", err)
	}

	// Load all time zones
	locations := make([]*time.Location, len(req.TimeZones))
	for i, tz := range req.TimeZones {
		loc, err := time.LoadLocation(tz)
		if err != nil {
			return nil, fmt.Errorf("invalid time zone %q: %w", tz, err)
		}
		locations[i] = loc
	}

	// TODO: Implement actual meeting time finding logic
	// This would iterate through the date range, find working hours in each zone,
	// and identify overlapping slots across all zones

	result := &domain.MeetingTimeSlots{
		Slots:      []domain.MeetingSlot{},
		TimeZones:  req.TimeZones,
		TotalSlots: 0,
	}

	// Placeholder: Return empty result for now
	// Real implementation would:
	// 1. Iterate through each day in DateRange
	// 2. For each day, find working hours in each time zone
	// 3. Calculate overlapping time slots
	// 4. Score each slot based on quality (middle of day = higher score)
	// 5. Filter by duration requirement

	_ = startHour
	_ = endHour
	_ = locations

	return result, nil
}

// GetDSTTransitions returns DST transition dates for a zone in a given year.
func (s *Service) GetDSTTransitions(ctx context.Context, zone string, year int) ([]domain.DSTTransition, error) {
	loc, err := time.LoadLocation(zone)
	if err != nil {
		return nil, fmt.Errorf("invalid zone %q: %w", zone, err)
	}

	var transitions []domain.DSTTransition

	// Check each hour of the year for DST transitions
	// DST transitions typically occur at 2 AM local time
	start := time.Date(year, 1, 1, 0, 0, 0, 0, loc)
	end := time.Date(year+1, 1, 1, 0, 0, 0, 0, loc)

	current := start
	lastOffset := getOffset(current)

	for current.Before(end) {
		currentOffset := getOffset(current)
		currentName := current.Format("MST")

		// Detect offset change (DST transition)
		if currentOffset != lastOffset {
			direction := "forward"
			if currentOffset < lastOffset {
				direction = "backward"
			}

			transitions = append(transitions, domain.DSTTransition{
				Date:      current,
				Offset:    currentOffset,
				Name:      currentName,
				IsDST:     isDST(current),
				Direction: direction,
			})

			lastOffset = currentOffset
		}

		current = current.Add(1 * time.Hour) // Check hourly for precise transition detection
	}

	return transitions, nil
}

// ListTimeZones returns all available IANA time zones.
func (s *Service) ListTimeZones(ctx context.Context) ([]string, error) {
	// Common IANA time zones
	// In a full implementation, this would read from the system's timezone database
	zones := []string{
		"America/New_York",
		"America/Chicago",
		"America/Denver",
		"America/Los_Angeles",
		"America/Phoenix",
		"America/Anchorage",
		"Pacific/Honolulu",
		"Europe/London",
		"Europe/Paris",
		"Europe/Berlin",
		"Europe/Rome",
		"Asia/Dubai",
		"Asia/Kolkata",
		"Asia/Singapore",
		"Asia/Tokyo",
		"Asia/Shanghai",
		"Australia/Sydney",
		"Pacific/Auckland",
		"UTC",
	}

	sort.Strings(zones)
	return zones, nil
}

// GetTimeZoneInfo returns detailed information about a time zone at a specific time.
func (s *Service) GetTimeZoneInfo(ctx context.Context, zone string, at time.Time) (*domain.TimeZoneInfo, error) {
	loc, err := time.LoadLocation(zone)
	if err != nil {
		return nil, fmt.Errorf("invalid zone %q: %w", zone, err)
	}

	timeInZone := at.In(loc)
	abbreviation := timeInZone.Format("MST")
	offset := getOffset(timeInZone)
	isDSTNow := isDST(timeInZone)

	// Find next DST transition (simplified - checks next 365 days)
	var nextDST *time.Time
	current := timeInZone
	currentOffset := offset
	for i := 0; i < 365; i++ {
		current = current.AddDate(0, 0, 1)
		if getOffset(current) != currentOffset {
			nextDST = &current
			break
		}
	}

	return &domain.TimeZoneInfo{
		Name:         zone,
		Abbreviation: abbreviation,
		Offset:       offset,
		IsDST:        isDSTNow,
		NextDST:      nextDST,
	}, nil
}

// ============================================================================
// Helper functions
// ============================================================================

// parseTime parses a time string in "HH:MM" format.
func parseTime(s string) (time.Time, error) {
	t, err := time.Parse("15:04", s)
	if err != nil {
		return time.Time{}, fmt.Errorf("expected format HH:MM: %w", err)
	}
	return t, nil
}

// getOffset returns the UTC offset in seconds for the given time.
func getOffset(t time.Time) int {
	_, offset := t.Zone()
	return offset
}

// isDST determines if the given time is during daylight saving time.
// This is a heuristic based on the zone name containing "DT" (Daylight Time).
func isDST(t time.Time) bool {
	name := t.Format("MST")
	// Common DST abbreviations contain "D" (PDT, EDT, CDT, etc.)
	// This is a simple heuristic and may not work for all zones
	return len(name) > 0 && (name[len(name)-2] == 'D' || name == "BST" || name == "CEST")
}

// CheckDSTWarning checks if a time is near or during a DST transition.
// Returns a DSTWarning if the time is within warningDays of a transition.
func (s *Service) CheckDSTWarning(ctx context.Context, t time.Time, zone string, warningDays int) (*domain.DSTWarning, error) {
	loc, err := time.LoadLocation(zone)
	if err != nil {
		return nil, fmt.Errorf("invalid zone %q: %w", zone, err)
	}

	timeInZone := t.In(loc)
	year := timeInZone.Year()

	// Get DST transitions for this year
	transitions, err := s.GetDSTTransitions(ctx, zone, year)
	if err != nil {
		return nil, err
	}

	// If no transitions, zone doesn't observe DST
	if len(transitions) == 0 {
		return nil, nil
	}

	// Check if time is near any transition
	warningWindow := time.Duration(warningDays) * 24 * time.Hour

	for _, transition := range transitions {
		timeDiff := transition.Date.Sub(timeInZone)
		absTimeDiff := timeDiff
		if absTimeDiff < 0 {
			absTimeDiff = -absTimeDiff
		}

		// Special handling for warningDays=0: check if on same day and in gap/duplicate hour
		onSameDay := timeInZone.Year() == transition.Date.Year() &&
			timeInZone.Month() == transition.Date.Month() &&
			timeInZone.Day() == transition.Date.Day()

		// Check if within warning window OR if warningDays=0 and on same day
		if absTimeDiff <= warningWindow || (warningDays == 0 && onSameDay) {
			warning := &domain.DSTWarning{
				IsNearTransition: true,
				TransitionDate:   transition.Date,
				Direction:        transition.Direction,
				DaysUntil:        int(timeDiff.Hours() / 24),
				TransitionName:   transition.Name,
			}

			// Check if time falls during transition (spring forward gap or fall back duplicate)
			switch transition.Direction {
			case "forward":
				// Spring forward: 2:00 AM -> 3:00 AM (2:00-2:59 doesn't exist)
				warning.InTransitionGap = s.isInSpringForwardGap(timeInZone, transition.Date)
				if warning.InTransitionGap {
					warning.Warning = "This time will not exist due to Daylight Saving Time (clocks spring forward)"
					warning.Severity = "error"
				} else if timeDiff > 0 && timeDiff <= 7*24*time.Hour {
					warning.Warning = fmt.Sprintf("Daylight Saving Time begins in %d days (clocks spring forward 1 hour)", warning.DaysUntil)
					warning.Severity = "warning"
				}
			case "backward":
				// Fall back: 2:00 AM -> 1:00 AM (1:00-1:59 occurs twice)
				warning.InDuplicateHour = s.isInFallBackDuplicate(timeInZone, transition.Date)
				if warning.InDuplicateHour {
					warning.Warning = "This time occurs twice due to Daylight Saving Time (clocks fall back)"
					warning.Severity = "warning"
				} else if timeDiff > 0 && timeDiff <= 7*24*time.Hour {
					warning.Warning = fmt.Sprintf("Daylight Saving Time ends in %d days (clocks fall back 1 hour)", warning.DaysUntil)
					warning.Severity = "info"
				}
			}

			// If warningDays=0, only return warning if actually in gap/duplicate hour
			if warningDays == 0 && !warning.InTransitionGap && !warning.InDuplicateHour {
				continue
			}

			return warning, nil
		}
	}

	return nil, nil
}

// isInSpringForwardGap checks if a time falls in the "spring forward" gap.
// During spring forward, times like 2:30 AM don't exist (clock jumps 2:00 -> 3:00).
// Go's time.Date normalizes such times backwards, so we detect this normalization.
func (s *Service) isInSpringForwardGap(t time.Time, transitionDate time.Time) bool {
	// The transition typically happens at 2 AM or 3 AM
	// We need to find the gap hour by detecting which hour gets normalized

	// Check if this time is on the transition day
	if t.Year() != transitionDate.Year() ||
		t.Month() != transitionDate.Month() ||
		t.Day() != transitionDate.Day() {
		return false
	}

	loc := t.Location()
	transitionHour := transitionDate.Hour()

	// The gap hour is typically one hour before the transition hour
	// (e.g., transition at 3:00 means gap is 2:00-2:59)
	gapHour := transitionHour - 1
	if gapHour < 0 {
		gapHour = 23
	}

	// Test if creating a time at the gap hour results in normalization
	testGapTime := time.Date(t.Year(), t.Month(), t.Day(), gapHour, 30, 0, 0, loc)

	// If the hour changed, this is the gap hour
	if testGapTime.Hour() != gapHour {
		// Check if the input time 't' has an hour that indicates it was normalized
		// Times in the gap get normalized backward (e.g., 2:30 -> 1:30)
		normalizedHour := testGapTime.Hour()

		// If t's hour matches the normalized hour, it might have been in the gap
		if t.Hour() == normalizedHour {
			return true
		}
	}

	return false
}

// isInFallBackDuplicate checks if a time falls in the "fall back" duplicate hour.
// During fall back, times like 1:30 AM occur twice (clock goes 2:00 -> 1:00).
func (s *Service) isInFallBackDuplicate(t time.Time, transitionDate time.Time) bool {
	// Get the hour when transition occurs (usually 2 AM, which becomes 1 AM)
	transitionHour := transitionDate.Hour()
	duplicateHour := transitionHour - 1

	// Check if time is on the same day and in the duplicate hour
	if t.Year() == transitionDate.Year() &&
		t.Month() == transitionDate.Month() &&
		t.Day() == transitionDate.Day() &&
		t.Hour() == duplicateHour {
		return true
	}

	return false
}

// SuggestAlternativeTimes suggests alternative times when a time falls in a DST conflict.
// For spring forward gaps, suggests times before and after the gap.
// For fall back duplicates, suggests clarification.
func (s *Service) SuggestAlternativeTimes(ctx context.Context, t time.Time, zone string, duration time.Duration) ([]time.Time, error) {
	loc, err := time.LoadLocation(zone)
	if err != nil {
		return nil, fmt.Errorf("invalid zone %q: %w", zone, err)
	}

	timeInZone := t.In(loc)

	// Check for DST warning
	warning, err := s.CheckDSTWarning(ctx, timeInZone, zone, 0)
	if err != nil {
		return nil, err
	}

	// If no DST conflict, no alternatives needed
	if warning == nil || (!warning.InTransitionGap && !warning.InDuplicateHour) {
		return nil, nil
	}

	var alternatives []time.Time

	if warning.InTransitionGap {
		// Spring forward: suggest time before the gap and after the gap
		// E.g., if trying to schedule at 2:30 AM (doesn't exist)
		// Suggest: 1:30 AM (before) and 3:00 AM (after)

		// Alternative 1: Same time one hour earlier (before DST)
		beforeGap := timeInZone.Add(-1 * time.Hour)
		alternatives = append(alternatives, beforeGap)

		// Alternative 2: Adjusted time after DST (accounting for the hour jump)
		afterGap := timeInZone.Add(1 * time.Hour)
		alternatives = append(alternatives, afterGap)
	} else if warning.InDuplicateHour {
		// Fall back: time occurs twice
		// Suggest the first occurrence (before fall back) and second occurrence (after fall back)

		// Alternative 1: First occurrence (DST time)
		firstOccurrence := timeInZone
		alternatives = append(alternatives, firstOccurrence)

		// Alternative 2: Second occurrence (standard time, one hour later in UTC)
		secondOccurrence := timeInZone.Add(1 * time.Hour)
		alternatives = append(alternatives, secondOccurrence)
	}

	return alternatives, nil
}
