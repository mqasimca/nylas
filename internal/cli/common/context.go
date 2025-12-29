package common

import (
	"context"
	"time"
)

// CreateContext creates a context with a 30-second timeout.
// Returns the context and a cancel function that should be deferred.
func CreateContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 30*time.Second)
}

// CreateContextWithTimeout creates a context with a custom timeout.
func CreateContextWithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}
