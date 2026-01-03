package air

import (
	"net/http"

	"github.com/mqasimca/nylas/internal/domain"
)

// handleListCalendars returns all calendars for the current account.
func (s *Server) handleListCalendars(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Demo mode: return mock calendars
	if s.demoMode {
		writeJSON(w, http.StatusOK, CalendarsResponse{
			Calendars: demoCalendars(),
		})
		return
	}

	// Check if configured
	if !s.requireConfig(w) {
		return
	}

	grantID, ok := s.requireDefaultGrant(w)
	if !ok {
		return
	}

	// Fetch calendars from Nylas API
	ctx, cancel := s.withTimeout(r)
	defer cancel()

	calendars, err := s.nylasClient.GetCalendars(ctx, grantID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch calendars: " + err.Error(),
		})
		return
	}

	// Convert to response format
	resp := CalendarsResponse{
		Calendars: make([]CalendarResponse, 0, len(calendars)),
	}
	for _, c := range calendars {
		resp.Calendars = append(resp.Calendars, calendarToResponse(c))
	}

	writeJSON(w, http.StatusOK, resp)
}

// calendarToResponse converts a domain calendar to an API response.
func calendarToResponse(c domain.Calendar) CalendarResponse {
	return CalendarResponse{
		ID:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		Timezone:    c.Timezone,
		IsPrimary:   c.IsPrimary,
		ReadOnly:    c.ReadOnly,
		HexColor:    c.HexColor,
	}
}

// demoCalendars returns demo calendar data.
func demoCalendars() []CalendarResponse {
	return []CalendarResponse{
		{
			ID:        "primary",
			Name:      "Personal Calendar",
			Timezone:  "America/New_York",
			IsPrimary: true,
			ReadOnly:  false,
			HexColor:  "#4285f4",
		},
		{
			ID:        "work",
			Name:      "Work Calendar",
			Timezone:  "America/New_York",
			IsPrimary: false,
			ReadOnly:  false,
			HexColor:  "#0b8043",
		},
		{
			ID:          "holidays",
			Name:        "US Holidays",
			Description: "Public holidays in the United States",
			Timezone:    "America/New_York",
			IsPrimary:   false,
			ReadOnly:    true,
			HexColor:    "#f6bf26",
		},
	}
}
