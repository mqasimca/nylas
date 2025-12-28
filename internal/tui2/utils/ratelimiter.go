// Package utils provides utility functions for the TUI.
package utils

import (
	"sync"
	"time"
)

// RateLimiter ensures API calls are spaced out to avoid rate limiting.
type RateLimiter struct {
	lastCall time.Time
	minDelay time.Duration
	mu       sync.Mutex
}

// NewRateLimiter creates a new rate limiter.
// minDelay is the minimum time between API calls.
func NewRateLimiter(minDelay time.Duration) *RateLimiter {
	return &RateLimiter{
		minDelay: minDelay,
	}
}

// Wait blocks until enough time has passed since the last API call.
func (r *RateLimiter) Wait() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.lastCall.IsZero() {
		r.lastCall = time.Now()
		return
	}

	elapsed := time.Since(r.lastCall)
	if elapsed < r.minDelay {
		time.Sleep(r.minDelay - elapsed)
	}

	r.lastCall = time.Now()
}

// TryWait returns true if enough time has passed, false otherwise.
// Non-blocking variant of Wait.
func (r *RateLimiter) TryWait() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.lastCall.IsZero() {
		r.lastCall = time.Now()
		return true
	}

	elapsed := time.Since(r.lastCall)
	if elapsed < r.minDelay {
		return false
	}

	r.lastCall = time.Now()
	return true
}
