package ports

import (
	"context"

	"github.com/mqasimca/nylas/internal/domain"
)

// SchedulerClient defines the interface for scheduler operations.
type SchedulerClient interface {
	// ================================
	// CONFIGURATION OPERATIONS
	// ================================

	// ListSchedulerConfigurations retrieves all scheduler configurations.
	ListSchedulerConfigurations(ctx context.Context) ([]domain.SchedulerConfiguration, error)

	// GetSchedulerConfiguration retrieves a specific scheduler configuration.
	GetSchedulerConfiguration(ctx context.Context, configID string) (*domain.SchedulerConfiguration, error)

	// CreateSchedulerConfiguration creates a new scheduler configuration.
	CreateSchedulerConfiguration(ctx context.Context, req *domain.CreateSchedulerConfigurationRequest) (*domain.SchedulerConfiguration, error)

	// UpdateSchedulerConfiguration updates an existing scheduler configuration.
	UpdateSchedulerConfiguration(ctx context.Context, configID string, req *domain.UpdateSchedulerConfigurationRequest) (*domain.SchedulerConfiguration, error)

	// DeleteSchedulerConfiguration deletes a scheduler configuration.
	DeleteSchedulerConfiguration(ctx context.Context, configID string) error

	// ================================
	// SESSION OPERATIONS
	// ================================

	// CreateSchedulerSession creates a new scheduler session.
	CreateSchedulerSession(ctx context.Context, req *domain.CreateSchedulerSessionRequest) (*domain.SchedulerSession, error)

	// GetSchedulerSession retrieves a specific scheduler session.
	GetSchedulerSession(ctx context.Context, sessionID string) (*domain.SchedulerSession, error)

	// ================================
	// BOOKING OPERATIONS
	// ================================

	// ListBookings retrieves all bookings for a configuration.
	ListBookings(ctx context.Context, configID string) ([]domain.Booking, error)

	// GetBooking retrieves a specific booking.
	GetBooking(ctx context.Context, bookingID string) (*domain.Booking, error)

	// ConfirmBooking confirms a booking.
	ConfirmBooking(ctx context.Context, bookingID string, req *domain.ConfirmBookingRequest) (*domain.Booking, error)

	// RescheduleBooking reschedules an existing booking.
	RescheduleBooking(ctx context.Context, bookingID string, req *domain.RescheduleBookingRequest) (*domain.Booking, error)

	// CancelBooking cancels a booking.
	CancelBooking(ctx context.Context, bookingID string, reason string) error

	// ================================
	// SCHEDULER PAGE OPERATIONS
	// ================================

	// ListSchedulerPages retrieves all scheduler pages.
	ListSchedulerPages(ctx context.Context) ([]domain.SchedulerPage, error)

	// GetSchedulerPage retrieves a specific scheduler page.
	GetSchedulerPage(ctx context.Context, pageID string) (*domain.SchedulerPage, error)

	// CreateSchedulerPage creates a new scheduler page.
	CreateSchedulerPage(ctx context.Context, req *domain.CreateSchedulerPageRequest) (*domain.SchedulerPage, error)

	// UpdateSchedulerPage updates an existing scheduler page.
	UpdateSchedulerPage(ctx context.Context, pageID string, req *domain.UpdateSchedulerPageRequest) (*domain.SchedulerPage, error)

	// DeleteSchedulerPage deletes a scheduler page.
	DeleteSchedulerPage(ctx context.Context, pageID string) error
}
