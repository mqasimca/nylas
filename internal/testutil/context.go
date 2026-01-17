// Package testutil provides common test utilities and helpers for the Nylas CLI.
package testutil

import (
	"context"
	"testing"
	"time"
)

// TestContext creates a 30-second timeout context with automatic cleanup.
// This is the standard context for most tests that interact with the API.
func TestContext(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	return ctx
}

// LongTestContext creates a 120-second timeout context with automatic cleanup.
// Use this for integration tests that may take longer (e.g., multiple API calls).
func LongTestContext(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	t.Cleanup(cancel)
	return ctx
}

// QuickTestContext creates a 5-second timeout context with automatic cleanup.
// Use this for unit tests that should complete quickly (e.g., no network calls).
func QuickTestContext(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	return ctx
}
