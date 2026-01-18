package utilities

import (
	"context"
	"time"

	"github.com/mqasimca/nylas/internal/domain"
)

// MockUtilityServices implements ports.UtilityServices for testing.
type MockUtilityServices struct {
	// TimeZoneService
	ConvertTimeFunc       func(ctx context.Context, fromZone, toZone string, t time.Time) (time.Time, error)
	FindMeetingTimeFunc   func(ctx context.Context, req *domain.MeetingFinderRequest) (*domain.MeetingTimeSlots, error)
	GetDSTTransitionsFunc func(ctx context.Context, zone string, year int) ([]domain.DSTTransition, error)
	ListTimeZonesFunc     func(ctx context.Context) ([]string, error)
	GetTimeZoneInfoFunc   func(ctx context.Context, zone string, at time.Time) (*domain.TimeZoneInfo, error)
}

// NewMockUtilityServices creates a new mock utility services with sensible defaults.
func NewMockUtilityServices() *MockUtilityServices {
	return &MockUtilityServices{
		// TimeZoneService defaults
		ConvertTimeFunc: func(ctx context.Context, fromZone, toZone string, t time.Time) (time.Time, error) {
			return t, nil
		},
		FindMeetingTimeFunc: func(ctx context.Context, req *domain.MeetingFinderRequest) (*domain.MeetingTimeSlots, error) {
			return &domain.MeetingTimeSlots{Slots: []domain.MeetingSlot{}, TimeZones: req.TimeZones}, nil
		},
		GetDSTTransitionsFunc: func(ctx context.Context, zone string, year int) ([]domain.DSTTransition, error) {
			return []domain.DSTTransition{}, nil
		},
		ListTimeZonesFunc: func(ctx context.Context) ([]string, error) {
			return []string{"UTC", "America/New_York", "Europe/London"}, nil
		},
		GetTimeZoneInfoFunc: func(ctx context.Context, zone string, at time.Time) (*domain.TimeZoneInfo, error) {
			return &domain.TimeZoneInfo{Name: zone, Abbreviation: "UTC", Offset: 0}, nil
		},
	}
}

// ============================================================================
// TimeZoneService implementation
// ============================================================================

func (m *MockUtilityServices) ConvertTime(ctx context.Context, fromZone, toZone string, t time.Time) (time.Time, error) {
	if m.ConvertTimeFunc != nil {
		return m.ConvertTimeFunc(ctx, fromZone, toZone, t)
	}
	return t, nil
}

func (m *MockUtilityServices) FindMeetingTime(ctx context.Context, req *domain.MeetingFinderRequest) (*domain.MeetingTimeSlots, error) {
	if m.FindMeetingTimeFunc != nil {
		return m.FindMeetingTimeFunc(ctx, req)
	}
	return &domain.MeetingTimeSlots{}, nil
}

func (m *MockUtilityServices) GetDSTTransitions(ctx context.Context, zone string, year int) ([]domain.DSTTransition, error) {
	if m.GetDSTTransitionsFunc != nil {
		return m.GetDSTTransitionsFunc(ctx, zone, year)
	}
	return []domain.DSTTransition{}, nil
}

func (m *MockUtilityServices) ListTimeZones(ctx context.Context) ([]string, error) {
	if m.ListTimeZonesFunc != nil {
		return m.ListTimeZonesFunc(ctx)
	}
	return []string{}, nil
}

func (m *MockUtilityServices) GetTimeZoneInfo(ctx context.Context, zone string, at time.Time) (*domain.TimeZoneInfo, error) {
	if m.GetTimeZoneInfoFunc != nil {
		return m.GetTimeZoneInfoFunc(ctx, zone, at)
	}
	return &domain.TimeZoneInfo{}, nil
}
