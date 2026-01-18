package ports

import (
	"context"
	"time"

	"github.com/mqasimca/nylas/internal/domain"
)

// UtilityServices defines interfaces for non-Nylas utility features.
// These services provide offline-capable tools that don't require Nylas API access.
type UtilityServices interface {
	TimeZoneService
}

// TimeZoneService provides time zone conversion and meeting finder utilities.
// Addresses the pain point where 83% of professionals struggle with time zone scheduling.
type TimeZoneService interface {
	// ConvertTime converts a time from one zone to another
	ConvertTime(ctx context.Context, fromZone, toZone string, t time.Time) (time.Time, error)

	// FindMeetingTime finds overlapping working hours across multiple time zones
	FindMeetingTime(ctx context.Context, req *domain.MeetingFinderRequest) (*domain.MeetingTimeSlots, error)

	// GetDSTTransitions returns DST transition dates for a zone in a given year
	GetDSTTransitions(ctx context.Context, zone string, year int) ([]domain.DSTTransition, error)

	// ListTimeZones returns all available IANA time zones
	ListTimeZones(ctx context.Context) ([]string, error)

	// GetTimeZoneInfo returns detailed information about a time zone
	GetTimeZoneInfo(ctx context.Context, zone string, at time.Time) (*domain.TimeZoneInfo, error)
}
