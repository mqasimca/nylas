// Package domain contains the core business logic and domain models.
package domain

import "errors"

// Sentinel errors for the application.
var (
	// Auth errors
	ErrNotConfigured   = errors.New("nylas not configured")
	ErrAuthFailed      = errors.New("authentication failed")
	ErrAuthTimeout     = errors.New("authentication timed out")
	ErrInvalidProvider = errors.New("invalid provider")
	ErrGrantNotFound   = errors.New("grant not found")
	ErrNoDefaultGrant  = errors.New("no default grant set")
	ErrInvalidGrant    = errors.New("invalid or expired grant")
	ErrTokenExpired    = errors.New("token expired")
	ErrAPIError        = errors.New("nylas API error")
	ErrNetworkError    = errors.New("network error")

	// Secret store errors
	ErrSecretNotFound    = errors.New("secret not found")
	ErrSecretStoreFailed = errors.New("secret store operation failed")

	// Config errors
	ErrConfigNotFound = errors.New("config not found")
	ErrConfigInvalid  = errors.New("config invalid")

	// OTP errors
	ErrOTPNotFound     = errors.New("no OTP found in recent messages")
	ErrAccountNotFound = errors.New("account not found")
	ErrNoMessages      = errors.New("no messages found")
)
