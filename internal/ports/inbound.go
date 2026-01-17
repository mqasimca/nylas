package ports

import (
	"context"

	"github.com/mqasimca/nylas/internal/domain"
)

// InboundClient defines the interface for inbound inbox operations.
type InboundClient interface {
	// ListInboundInboxes retrieves all inbound inboxes.
	ListInboundInboxes(ctx context.Context) ([]domain.InboundInbox, error)

	// GetInboundInbox retrieves a specific inbound inbox.
	GetInboundInbox(ctx context.Context, grantID string) (*domain.InboundInbox, error)

	// CreateInboundInbox creates a new inbound inbox.
	CreateInboundInbox(ctx context.Context, email string) (*domain.InboundInbox, error)

	// DeleteInboundInbox deletes an inbound inbox.
	DeleteInboundInbox(ctx context.Context, grantID string) error

	// GetInboundMessages retrieves inbound messages with query parameters.
	GetInboundMessages(ctx context.Context, grantID string, params *domain.MessageQueryParams) ([]domain.InboundMessage, error)
}
