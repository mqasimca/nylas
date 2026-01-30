// Package ports defines the interfaces for external dependencies.
package ports

import (
	"context"

	"github.com/mqasimca/nylas/internal/domain"
)

// TemplateStore defines the interface for email template storage.
// Templates are stored locally (not in Nylas API) for use with email sending.
type TemplateStore interface {
	// List returns all templates, optionally filtered by category.
	// If category is empty, all templates are returned.
	List(ctx context.Context, category string) ([]domain.EmailTemplate, error)

	// Get retrieves a template by its ID.
	Get(ctx context.Context, id string) (*domain.EmailTemplate, error)

	// Create creates a new template and returns it with generated ID.
	Create(ctx context.Context, t *domain.EmailTemplate) (*domain.EmailTemplate, error)

	// Update updates an existing template.
	Update(ctx context.Context, t *domain.EmailTemplate) (*domain.EmailTemplate, error)

	// Delete removes a template by its ID.
	Delete(ctx context.Context, id string) error

	// IncrementUsage increments the usage count for a template.
	IncrementUsage(ctx context.Context, id string) error

	// Path returns the path to the templates file.
	Path() string
}
