package ports

import (
	"context"

	"github.com/mqasimca/nylas/internal/domain"
)

// NotetakerClient defines the interface for notetaker operations.
type NotetakerClient interface {
	// ListNotetakers retrieves notetakers with query parameters.
	ListNotetakers(ctx context.Context, grantID string, params *domain.NotetakerQueryParams) ([]domain.Notetaker, error)

	// GetNotetaker retrieves a specific notetaker.
	GetNotetaker(ctx context.Context, grantID, notetakerID string) (*domain.Notetaker, error)

	// CreateNotetaker creates a new notetaker.
	CreateNotetaker(ctx context.Context, grantID string, req *domain.CreateNotetakerRequest) (*domain.Notetaker, error)

	// DeleteNotetaker deletes a notetaker.
	DeleteNotetaker(ctx context.Context, grantID, notetakerID string) error

	// GetNotetakerMedia retrieves media data for a notetaker.
	GetNotetakerMedia(ctx context.Context, grantID, notetakerID string) (*domain.MediaData, error)
}
