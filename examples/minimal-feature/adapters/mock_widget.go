package adapters

import (
	"context"

	"github.com/mqasimca/nylas/examples/minimal-feature/domain"
	"github.com/mqasimca/nylas/examples/minimal-feature/ports"
)

// MockWidgetService is a mock implementation for testing.
//
// Mocks are:
// - Test doubles for adapter implementations
// - Allow controlled behavior in tests
// - Enable testing without external dependencies
type MockWidgetService struct {
	// Function fields allow test-specific behavior
	ListFunc   func(ctx context.Context) ([]*domain.Widget, error)
	GetFunc    func(ctx context.Context, id string) (*domain.Widget, error)
	CreateFunc func(ctx context.Context, widget *domain.Widget) (*domain.Widget, error)
	UpdateFunc func(ctx context.Context, widget *domain.Widget) (*domain.Widget, error)
	DeleteFunc func(ctx context.Context, id string) error
}

// Ensure MockWidgetService implements the interface at compile time
var _ ports.WidgetService = (*MockWidgetService)(nil)

// ListWidgets calls the mock function if set.
func (m *MockWidgetService) ListWidgets(ctx context.Context) ([]*domain.Widget, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx)
	}
	return []*domain.Widget{}, nil
}

// GetWidget calls the mock function if set.
func (m *MockWidgetService) GetWidget(ctx context.Context, id string) (*domain.Widget, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, id)
	}
	return nil, nil
}

// CreateWidget calls the mock function if set.
func (m *MockWidgetService) CreateWidget(ctx context.Context, widget *domain.Widget) (*domain.Widget, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, widget)
	}
	return widget, nil
}

// UpdateWidget calls the mock function if set.
func (m *MockWidgetService) UpdateWidget(ctx context.Context, widget *domain.Widget) (*domain.Widget, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, widget)
	}
	return widget, nil
}

// DeleteWidget calls the mock function if set.
func (m *MockWidgetService) DeleteWidget(ctx context.Context, id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}
