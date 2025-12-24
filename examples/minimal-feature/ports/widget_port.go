package ports

import (
	"context"

	"github.com/mqasimca/nylas/examples/minimal-feature/domain"
)

// WidgetService defines the contract for widget operations.
//
// Ports (interfaces) are:
// - Pure interface definitions
// - Define WHAT can be done, not HOW
// - Context-first parameter pattern
// - No implementation details
//
// Benefits:
// - Enables dependency injection
// - Makes testing easy (mock implementations)
// - Decouples business logic from infrastructure
type WidgetService interface {
	// ListWidgets retrieves all widgets.
	// Returns empty slice if no widgets exist.
	ListWidgets(ctx context.Context) ([]*domain.Widget, error)

	// GetWidget retrieves a widget by ID.
	// Returns error if widget not found.
	GetWidget(ctx context.Context, id string) (*domain.Widget, error)

	// CreateWidget creates a new widget.
	// The widget's ID will be generated and returned.
	CreateWidget(ctx context.Context, widget *domain.Widget) (*domain.Widget, error)

	// UpdateWidget updates an existing widget.
	// Returns error if widget doesn't exist.
	UpdateWidget(ctx context.Context, widget *domain.Widget) (*domain.Widget, error)

	// DeleteWidget deletes a widget by ID.
	// Returns error if widget doesn't exist.
	DeleteWidget(ctx context.Context, id string) error
}
