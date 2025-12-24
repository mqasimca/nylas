package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mqasimca/nylas/examples/minimal-feature/domain"
	"github.com/mqasimca/nylas/examples/minimal-feature/ports"
)

// WidgetAdapter implements the WidgetService interface using HTTP API.
//
// Adapters are:
// - Concrete implementations of ports (interfaces)
// - Handle external communication (API, DB, etc.)
// - Transform between domain models and external formats
// - Isolated from business logic
type WidgetAdapter struct {
	client  *http.Client
	baseURL string
	apiKey  string
}

// NewWidgetAdapter creates a new widget adapter.
func NewWidgetAdapter(baseURL, apiKey string) ports.WidgetService {
	return &WidgetAdapter{
		client:  &http.Client{Timeout: 30 * time.Second},
		baseURL: baseURL,
		apiKey:  apiKey,
	}
}

// ListWidgets retrieves all widgets from the API.
func (a *WidgetAdapter) ListWidgets(ctx context.Context) ([]*domain.Widget, error) {
	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", a.baseURL+"/widgets", nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Add authentication
	req.Header.Set("Authorization", "Bearer "+a.apiKey)

	// Execute request
	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	// Check status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	// Parse response
	var widgets []*domain.Widget
	if err := json.NewDecoder(resp.Body).Decode(&widgets); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return widgets, nil
}

// GetWidget retrieves a single widget by ID.
func (a *WidgetAdapter) GetWidget(ctx context.Context, id string) (*domain.Widget, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", a.baseURL+"/widgets/"+id, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+a.apiKey)

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("widget not found: %s", id)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var widget domain.Widget
	if err := json.NewDecoder(resp.Body).Decode(&widget); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &widget, nil
}

// CreateWidget creates a new widget via API.
func (a *WidgetAdapter) CreateWidget(ctx context.Context, widget *domain.Widget) (*domain.Widget, error) {
	// Validate before sending to API
	if err := widget.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Implementation would make POST request...
	// For example purposes, simplified
	return widget, nil
}

// UpdateWidget updates an existing widget.
func (a *WidgetAdapter) UpdateWidget(ctx context.Context, widget *domain.Widget) (*domain.Widget, error) {
	if err := widget.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Implementation would make PUT request...
	return widget, nil
}

// DeleteWidget deletes a widget by ID.
func (a *WidgetAdapter) DeleteWidget(ctx context.Context, id string) error {
	// Implementation would make DELETE request...
	return nil
}
