package nylas

import (
	"time"

	"github.com/mqasimca/nylas/internal/domain"
)

func convertCalendars(cals []calendarResponse) []domain.Calendar {
	result := make([]domain.Calendar, len(cals))
	for i, c := range cals {
		result[i] = convertCalendar(c)
	}
	return result
}

// convertCalendar converts an API calendar response to domain model.
func convertCalendar(c calendarResponse) domain.Calendar {
	return domain.Calendar{
		ID:          c.ID,
		GrantID:     c.GrantID,
		Name:        c.Name,
		Description: c.Description,
		Location:    c.Location,
		Timezone:    c.Timezone,
		ReadOnly:    c.ReadOnly,
		IsPrimary:   c.IsPrimary,
		IsOwner:     c.IsOwner,
		HexColor:    c.HexColor,
		Object:      c.Object,
	}
}

// convertEvents converts API event responses to domain models.
func convertEvents(events []eventResponse) []domain.Event {
	result := make([]domain.Event, len(events))
	for i, e := range events {
		result[i] = convertEvent(e)
	}
	return result
}

// convertEvent converts an API event response to domain model.
func convertEvent(e eventResponse) domain.Event {
	participants := make([]domain.Participant, len(e.Participants))
	for j, p := range e.Participants {
		participants[j] = domain.Participant{
			Name:    p.Name,
			Email:   p.Email,
			Status:  p.Status,
			Comment: p.Comment,
		}
	}

	var organizer *domain.Participant
	if e.Organizer != nil {
		organizer = &domain.Participant{
			Name:    e.Organizer.Name,
			Email:   e.Organizer.Email,
			Status:  e.Organizer.Status,
			Comment: e.Organizer.Comment,
		}
	}

	var conferencing *domain.Conferencing
	if e.Conferencing != nil {
		conferencing = &domain.Conferencing{
			Provider: e.Conferencing.Provider,
		}
		if e.Conferencing.Details != nil {
			conferencing.Details = &domain.ConferencingDetails{
				URL:         e.Conferencing.Details.URL,
				MeetingCode: e.Conferencing.Details.MeetingCode,
				Password:    e.Conferencing.Details.Password,
				Phone:       e.Conferencing.Details.Phone,
			}
		}
	}

	var reminders *domain.Reminders
	if e.Reminders != nil {
		overrides := make([]domain.Reminder, len(e.Reminders.Overrides))
		for j, o := range e.Reminders.Overrides {
			overrides[j] = domain.Reminder{
				ReminderMinutes: o.ReminderMinutes,
				ReminderMethod:  o.ReminderMethod,
			}
		}
		reminders = &domain.Reminders{
			UseDefault: e.Reminders.UseDefault,
			Overrides:  overrides,
		}
	}

	return domain.Event{
		ID:          e.ID,
		GrantID:     e.GrantID,
		CalendarID:  e.CalendarID,
		Title:       e.Title,
		Description: e.Description,
		Location:    e.Location,
		When: domain.EventWhen{
			StartTime:     e.When.StartTime,
			EndTime:       e.When.EndTime,
			StartTimezone: e.When.StartTimezone,
			EndTimezone:   e.When.EndTimezone,
			Date:          e.When.Date,
			EndDate:       e.When.EndDate,
			StartDate:     e.When.StartDate,
			Object:        e.When.Object,
		},
		Participants:  participants,
		Organizer:     organizer,
		Status:        e.Status,
		Busy:          e.Busy,
		ReadOnly:      e.ReadOnly,
		Visibility:    e.Visibility,
		Recurrence:    e.Recurrence,
		Conferencing:  conferencing,
		Reminders:     reminders,
		MasterEventID: e.MasterEventID,
		ICalUID:       e.ICalUID,
		HtmlLink:      e.HtmlLink,
		CreatedAt:     time.Unix(e.CreatedAt, 0),
		UpdatedAt:     time.Unix(e.UpdatedAt, 0),
		Object:        e.Object,
	}
}
