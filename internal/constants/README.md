# Constants Package

Centralized constant values for the Nylas CLI.

## Purpose

This package provides a single source of truth for commonly used default values across the codebase:
- Default port numbers for various services
- Default URLs and endpoints
- Service names and identifiers
- Timeout values for different operations

## Usage

Import and use constants instead of hardcoding values:

```go
import "github.com/mqasimca/nylas/internal/constants"

// Use port constants
server := NewServer(constants.DefaultAirUIPort)

// Use URL constants
client := NewClient(constants.DefaultNylasAPIBaseURL)

// Use timeout constants
ctx, cancel := context.WithTimeout(context.Background(), constants.DefaultAPITimeout)
defer cancel()
```

## Constants Reference

### Ports

| Constant | Value | Usage |
|----------|-------|-------|
| `DefaultInboundPort` | 3000 | Inbound/webhook server (`nylas inbound monitor`) |
| `DefaultCallbackPort` | 8080 | OAuth callback server |
| `DefaultOllamaPort` | 11434 | Ollama AI service |
| `DefaultAirUIPort` | 7365 | Air web email client (`nylas air`) |
| `DefaultUIPort` | 7363 | Web-based CLI admin interface (`nylas ui`) |
| `DefaultAirPort` | 7363 | Deprecated: Use `DefaultAirUIPort` instead |

### URLs

| Constant | Value | Usage |
|----------|-------|-------|
| `DefaultNylasAPIBaseURL` | `https://api.us.nylas.com` | Nylas API base URL (US region) |
| `DefaultSchedulerBaseURL` | `https://schedule.nylas.com/` | Scheduler booking URL base |
| `DefaultOllamaHost` | `http://localhost:11434` | Ollama AI server URL |

### Service Names

| Constant | Value | Usage |
|----------|-------|-------|
| `ServiceName` | `nylas` | Canonical service name |

### Timeouts

| Constant | Value | Usage |
|----------|-------|-------|
| `DefaultAPITimeout` | 90s | Nylas API calls |
| `DefaultOAuthTimeout` | 5m | OAuth flows (requires user interaction) |
| `DefaultAITimeout` | 120s | AI/LLM operations |
| `DefaultHealthCheckTimeout` | 10s | Health/connectivity checks |
| `DefaultQuickCheckTimeout` | 5s | Quick checks (version, etc.) |
| `DefaultHTTPReadHeaderTimeout` | 10s | HTTP server read headers |
| `DefaultHTTPReadTimeout` | 30s | HTTP server read request |
| `DefaultHTTPWriteTimeout` | 30s | HTTP server write response |
| `DefaultHTTPIdleTimeout` | 120s | HTTP keep-alive idle |

## Authoritative Sources

**Timeout values** in this package mirror the authoritative definitions in `internal/domain/config.go`.

If you need to change timeout behavior, update the values in **both** locations:
1. `internal/domain/config.go` - Authoritative source
2. `internal/constants/constants.go` - Convenience constants

## Testing

The package includes comprehensive tests:
- Port value validation
- URL format validation
- Timeout value validation
- Logical relationship checks (e.g., quick timeouts < longer timeouts)

Run tests:
```bash
go test ./internal/constants/...
```

## Guidelines

- **DO** use these constants instead of hardcoding values
- **DO** add new constants when a value is used in 2+ places
- **DO** document the usage/purpose of each constant
- **DO NOT** modify timeout values without updating `internal/domain/config.go`
- **DO NOT** add constants for values used in only one place
