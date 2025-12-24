package domain

import (
	"fmt"
	"time"
)

// Widget represents a simple business entity in the domain layer.
//
// Domain models are:
// - Pure Go types with no external dependencies
// - Contain business logic and validation
// - Serializable (JSON tags for API communication)
type Widget struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Validate checks if the widget meets business rules.
//
// Domain layer contains business logic:
// - Validation rules
// - Business constraints
// - Data integrity checks
func (w *Widget) Validate() error {
	if w.Name == "" {
		return fmt.Errorf("widget name cannot be empty")
	}

	if len(w.Name) > 100 {
		return fmt.Errorf("widget name must be 100 characters or less")
	}

	if len(w.Description) > 500 {
		return fmt.Errorf("widget description must be 500 characters or less")
	}

	return nil
}

// IsNew returns true if this is a new widget (no ID assigned yet).
func (w *Widget) IsNew() bool {
	return w.ID == ""
}

// String returns a human-readable representation.
func (w *Widget) String() string {
	return fmt.Sprintf("Widget{ID: %s, Name: %s}", w.ID, w.Name)
}
