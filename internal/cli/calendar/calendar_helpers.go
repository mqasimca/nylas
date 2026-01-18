package calendar

import (
	"context"

	"github.com/mqasimca/nylas/internal/cli/common"
	"github.com/mqasimca/nylas/internal/ports"
)

// GetDefaultCalendarID retrieves the default calendar ID for a given grant.
// If calendarID is already provided, returns it unchanged.
// Otherwise, fetches calendars and returns:
//  1. The primary writable calendar (for writable=true)
//  2. The primary calendar (for writable=false)
//  3. The first writable calendar (fallback for writable=true)
//  4. The first calendar (fallback for writable=false)
//
// Returns an error if no suitable calendar is found.
func GetDefaultCalendarID(
	ctx context.Context,
	client ports.NylasClient,
	grantID string,
	calendarID string,
	writable bool,
) (string, error) {
	// If calendar ID already specified, use it
	if calendarID != "" {
		return calendarID, nil
	}

	// Fetch all calendars for the grant
	calendars, err := client.GetCalendars(ctx, grantID)
	if err != nil {
		return "", common.WrapListError("calendars", err)
	}

	if len(calendars) == 0 {
		return "", common.NewUserError(
			"no calendars found",
			"Connect a calendar account with: nylas auth login",
		)
	}

	// If writable required, find primary writable calendar
	if writable {
		for _, cal := range calendars {
			if cal.IsPrimary && !cal.ReadOnly {
				return cal.ID, nil
			}
		}
		// Fallback: any writable calendar
		for _, cal := range calendars {
			if !cal.ReadOnly {
				return cal.ID, nil
			}
		}
		return "", common.NewUserError(
			"no writable calendar found",
			"Specify a calendar with --calendar",
		)
	}

	// For read-only access, prefer primary calendar
	for _, cal := range calendars {
		if cal.IsPrimary {
			return cal.ID, nil
		}
	}

	// Fallback: first calendar
	return calendars[0].ID, nil
}
