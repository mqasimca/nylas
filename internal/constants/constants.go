// Package constants provides commonly used constant values across the Nylas CLI.
// These constants centralize default ports, URLs, timeouts, and service names
// to ensure consistency and make configuration changes easier.
package constants

import (
	"time"
)

// Default port numbers for various services.
const (
	// DefaultInboundPort is the default port for inbound/webhook server.
	// Used by: nylas inbound monitor
	DefaultInboundPort = 3000

	// DefaultCallbackPort is the default port for OAuth callback server.
	// Used during OAuth authentication flows.
	DefaultCallbackPort = 8080

	// DefaultOllamaPort is the default port for Ollama AI service.
	// Used by AI features for local LLM inference.
	DefaultOllamaPort = 11434

	// DefaultAirUIPort is the default port for Air web email client.
	// Used by: nylas air
	DefaultAirUIPort = 7365

	// DefaultAirPort is the deprecated port for Air API.
	// Kept for backward compatibility.
	// Deprecated: Use DefaultAirUIPort instead.
	DefaultAirPort = 7363

	// DefaultUIPort is the default port for the web-based CLI admin interface.
	// Used by: nylas ui
	DefaultUIPort = 7363
)

// Default URLs and endpoints.
const (
	// DefaultNylasAPIBaseURL is the default Nylas API base URL (US region).
	// Used as the default value in config when no region is specified.
	DefaultNylasAPIBaseURL = "https://api.us.nylas.com"

	// DefaultSchedulerBaseURL is the base URL for Nylas Scheduler pages.
	// Used to construct booking URLs for scheduler configurations.
	DefaultSchedulerBaseURL = "https://schedule.nylas.com/"

	// DefaultOllamaHost is the default Ollama server URL.
	// Used by AI features when no custom Ollama host is configured.
	DefaultOllamaHost = "http://localhost:11434"
)

// Service names and identifiers.
const (
	// ServiceName is the canonical name of the Nylas CLI service.
	// Used for logging, configuration paths, and service identification.
	ServiceName = "nylas"
)

// Default timeout values.
// Note: These reference the authoritative timeout values from internal/domain/config.go.
const (
	// DefaultAPITimeout is the default timeout for Nylas API calls.
	// Reference: domain.TimeoutAPI (90 seconds)
	DefaultAPITimeout = 90 * time.Second

	// DefaultOAuthTimeout is the default timeout for OAuth authentication flows.
	// OAuth requires user interaction in browser, so needs longer timeout.
	// Reference: domain.TimeoutOAuth (5 minutes)
	DefaultOAuthTimeout = 5 * time.Minute

	// DefaultAITimeout is the default timeout for AI/LLM operations.
	// AI providers may take longer due to model inference time.
	// Reference: domain.TimeoutAI (120 seconds)
	DefaultAITimeout = 120 * time.Second

	// DefaultHealthCheckTimeout is the default timeout for health/connectivity checks.
	// Reference: domain.TimeoutHealthCheck (10 seconds)
	DefaultHealthCheckTimeout = 10 * time.Second

	// DefaultQuickCheckTimeout is the default timeout for quick checks like version checking.
	// Reference: domain.TimeoutQuickCheck (5 seconds)
	DefaultQuickCheckTimeout = 5 * time.Second
)

// HTTP Server timeouts.
// Note: These reference the authoritative timeout values from internal/domain/config.go.
const (
	// DefaultHTTPReadHeaderTimeout is the time to read request headers.
	// Reference: domain.HTTPReadHeaderTimeout (10 seconds)
	DefaultHTTPReadHeaderTimeout = 10 * time.Second

	// DefaultHTTPReadTimeout is the time to read entire request.
	// Reference: domain.HTTPReadTimeout (30 seconds)
	DefaultHTTPReadTimeout = 30 * time.Second

	// DefaultHTTPWriteTimeout is the time to write response.
	// Reference: domain.HTTPWriteTimeout (30 seconds)
	DefaultHTTPWriteTimeout = 30 * time.Second

	// DefaultHTTPIdleTimeout is the keep-alive connection idle timeout.
	// Reference: domain.HTTPIdleTimeout (120 seconds)
	DefaultHTTPIdleTimeout = 120 * time.Second
)
