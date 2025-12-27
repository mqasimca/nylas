package common

import (
	"errors"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/mqasimca/nylas/internal/domain"
)

// CLIError wraps an error with additional context for CLI display.
type CLIError struct {
	Err        error
	Message    string
	Suggestion string
	Code       string
}

func (e *CLIError) Error() string {
	return e.Message
}

func (e *CLIError) Unwrap() error {
	return e.Err
}

// ErrorCode constants for common errors.
const (
	ErrCodeNotConfigured    = "E001"
	ErrCodeAuthFailed       = "E002"
	ErrCodeNetworkError     = "E003"
	ErrCodeNotFound         = "E004"
	ErrCodePermissionDenied = "E005"
	ErrCodeInvalidInput     = "E006"
	ErrCodeRateLimited      = "E007"
	ErrCodeServerError      = "E008"
)

// WrapError wraps an error with CLI-friendly context.
func WrapError(err error) *CLIError {
	if err == nil {
		return nil
	}

	// Check for existing CLIError
	var cliErr *CLIError
	if errors.As(err, &cliErr) {
		return cliErr
	}

	// Map domain errors to CLI errors
	switch {
	case errors.Is(err, domain.ErrNotConfigured):
		return &CLIError{
			Err:        err,
			Message:    "Nylas CLI is not configured",
			Suggestion: "Run 'nylas auth config' to set up your API credentials",
			Code:       ErrCodeNotConfigured,
		}

	case errors.Is(err, domain.ErrAuthFailed):
		return &CLIError{
			Err:        err,
			Message:    "Authentication failed",
			Suggestion: "Check your API key with 'nylas auth status' or reconfigure with 'nylas auth config'",
			Code:       ErrCodeAuthFailed,
		}

	case errors.Is(err, domain.ErrGrantNotFound):
		return &CLIError{
			Err:        err,
			Message:    "Grant not found",
			Suggestion: "Run 'nylas auth list' to see available grants, or 'nylas auth login' to add a new one",
			Code:       ErrCodeNotFound,
		}

	case errors.Is(err, domain.ErrNoDefaultGrant):
		return &CLIError{
			Err:        err,
			Message:    "No default grant set",
			Suggestion: "Run 'nylas auth list' to see grants, then 'nylas auth switch <grant-id>' to set a default",
			Code:       ErrCodeNotConfigured,
		}

	case errors.Is(err, domain.ErrSecretNotFound):
		return &CLIError{
			Err:        err,
			Message:    "Credentials not found in secret store",
			Suggestion: "Run 'nylas auth config' to set up your API credentials",
			Code:       ErrCodeNotConfigured,
		}

	case errors.Is(err, domain.ErrSecretStoreFailed):
		return &CLIError{
			Err:        err,
			Message:    "Failed to access secret store",
			Suggestion: "Check that your system keyring is accessible, or try running with elevated permissions",
			Code:       ErrCodePermissionDenied,
		}

	case errors.Is(err, domain.ErrNetworkError):
		return &CLIError{
			Err:        err,
			Message:    "Network error",
			Suggestion: "Check your internet connection and try again",
			Code:       ErrCodeNetworkError,
		}

	case errors.Is(err, domain.ErrTokenExpired):
		return &CLIError{
			Err:        err,
			Message:    "Authentication token has expired",
			Suggestion: "Run 'nylas auth login' to re-authenticate",
			Code:       ErrCodeAuthFailed,
		}

	case errors.Is(err, domain.ErrOTPNotFound):
		return &CLIError{
			Err:        err,
			Message:    "No OTP code found",
			Suggestion: "Check that you have OTP emails in your inbox, or try 'nylas otp watch' to wait for new codes",
			Code:       ErrCodeNotFound,
		}

	case errors.Is(err, domain.ErrInvalidProvider):
		return &CLIError{
			Err:        err,
			Message:    "Invalid email provider",
			Suggestion: "Supported providers are 'google' and 'microsoft'",
			Code:       ErrCodeInvalidInput,
		}
	}

	// Check for common error patterns in the error message
	errMsg := err.Error()

	if strings.Contains(errMsg, "Invalid API Key") {
		return &CLIError{
			Err:        err,
			Message:    "Invalid API key",
			Suggestion: "Run 'nylas auth config' to update your API key",
			Code:       ErrCodeAuthFailed,
		}
	}

	if strings.Contains(errMsg, "rate limit") || strings.Contains(errMsg, "429") {
		return &CLIError{
			Err:        err,
			Message:    "Rate limit exceeded",
			Suggestion: "Wait a moment and try again, or reduce the frequency of requests",
			Code:       ErrCodeRateLimited,
		}
	}

	if strings.Contains(errMsg, "connection refused") || strings.Contains(errMsg, "no such host") {
		return &CLIError{
			Err:        err,
			Message:    "Unable to connect to Nylas API",
			Suggestion: "Check your internet connection and firewall settings",
			Code:       ErrCodeNetworkError,
		}
	}

	if strings.Contains(errMsg, "timeout") {
		return &CLIError{
			Err:        err,
			Message:    "Request timed out",
			Suggestion: "The server is taking too long to respond. Try again later",
			Code:       ErrCodeNetworkError,
		}
	}

	if strings.Contains(errMsg, "500") || strings.Contains(errMsg, "502") || strings.Contains(errMsg, "503") {
		return &CLIError{
			Err:        err,
			Message:    "Nylas API server error",
			Suggestion: "This is a temporary issue. Please try again in a few minutes",
			Code:       ErrCodeServerError,
		}
	}

	// Default wrapper
	return &CLIError{
		Err:     err,
		Message: err.Error(),
	}
}

// FormatError formats an error for CLI display.
func FormatError(err error) string {
	cliErr := WrapError(err)
	if cliErr == nil {
		return ""
	}

	var sb strings.Builder
	red := color.New(color.FgRed)
	yellow := color.New(color.FgYellow)
	dim := color.New(color.Faint)

	// Error message
	_, _ = red.Fprintf(&sb, "Error: %s\n", cliErr.Message)

	// Error code (if available)
	if cliErr.Code != "" {
		_, _ = dim.Fprintf(&sb, "  Code: %s\n", cliErr.Code)
	}

	// Suggestion (if available)
	if cliErr.Suggestion != "" {
		_, _ = yellow.Fprintf(&sb, "  Hint: %s\n", cliErr.Suggestion)
	}

	// Original error in debug mode
	if IsDebug() && cliErr.Err != nil && cliErr.Err.Error() != cliErr.Message {
		_, _ = dim.Fprintf(&sb, "  Details: %s\n", cliErr.Err.Error())
	}

	return sb.String()
}

// PrintFormattedError prints a formatted error to stderr.
func PrintFormattedError(err error) {
	_, _ = fmt.Fprint(color.Error, FormatError(err))
}

// NewUserError creates a user-facing error with a suggestion.
func NewUserError(message, suggestion string) error {
	return &CLIError{
		Message:    message,
		Suggestion: suggestion,
	}
}

// NewInputError creates an input validation error.
func NewInputError(message string) error {
	return &CLIError{
		Message: message,
		Code:    ErrCodeInvalidInput,
	}
}
