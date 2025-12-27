# Testing Guidelines

Consolidated testing rules for the Nylas CLI project.

---

## Test Organization

### Unit Tests
- **Location:** Alongside source (`*_test.go`)
- **Function:** `TestFunctionName_Scenario`
- **Pattern:** Table-driven with `t.Run(tt.name, ...)`

### Integration Tests
- **Location:** `internal/cli/integration/`
- **Build tags:** `//go:build integration` and `// +build integration`
- **Package:** `package integration`
- **Function:** `TestCLI_CommandName`

### Air Integration Tests (Web UI)
- **Location:** `internal/air/integration_*.go`
- **Build tags:** `//go:build integration` and `// +build integration`
- **Package:** `package air`
- **Function:** `TestIntegration_FeatureName`
- **Files:**
  - `integration_base_test.go` - Shared helpers (`testServer()`, utilities)
  - `integration_core_test.go` - Config, Grants, Folders, Index
  - `integration_email_test.go` - Email and draft operations
  - `integration_calendar_test.go` - Calendar, events, availability
  - `integration_contacts_test.go` - Contact operations
  - `integration_cache_test.go` - Cache operations
  - `integration_ai_test.go` - AI features
  - `integration_middleware_test.go` - Middleware tests

**⚠️ CRITICAL: Always run Air tests with cleanup**

Air tests create real resources (drafts, events, contacts) in the connected account. **Use `make ci-full` which includes automatic cleanup**, or run cleanup manually:

```bash
make ci-full                     # RECOMMENDED: Complete CI with automatic cleanup
make test-air-integration        # Air integration tests only
make test-cleanup                # Manual cleanup if needed
```

---

## Parallel Testing

```go
func TestExample(t *testing.T) {
    skipIfMissingCreds(t)
    t.Parallel()  // Enable parallel execution

    // For API calls - use rate-limited wrapper
    stdout, stderr, err := runCLIWithRateLimit(t, "command", "subcommand")

    // For offline commands - no rate limiting
    stdout, stderr, err := runCLI("timezone", "list")
}
```

**Rate limiting config:**
```bash
export NYLAS_TEST_RATE_LIMIT_RPS="2.0"    # Requests per second
export NYLAS_TEST_RATE_LIMIT_BURST="5"    # Burst size
```

**✅ Use rate limiting for:** API commands (calendar, email, contacts, webhooks)
**❌ Skip rate limiting for:** Offline commands (timezone, version, help)

---

## Test Patterns

### Table-Driven Tests
```go
tests := []struct {
    name    string
    input   string
    wantErr bool
}{
    {"valid", "test@example.com", false},
    {"invalid", "not-email", true},
}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // test logic
    })
}
```

### Mock Pattern
```go
type MockClient struct {
    GetFunc func(ctx context.Context, id string) (*domain.Object, error)
}

func (m *MockClient) Get(ctx context.Context, id string) (*domain.Object, error) {
    if m.GetFunc != nil {
        return m.GetFunc(ctx, id)
    }
    return nil, nil
}
```

### Cleanup
```go
tmpDir := t.TempDir()  // Auto-cleaned

t.Cleanup(func() {
    acquireRateLimit(t)
    _ = client.Delete(ctx, resourceID)
})
```

---

## Test Coverage

| Package Type | Minimum | Target |
|--------------|---------|--------|
| Core Adapters | 70% | 85%+ |
| Business Logic | 60% | 80%+ |
| CLI Commands | 50% | 70%+ |
| Utilities | 90% | 100% |

**Check coverage:**
```bash
make test-coverage  # Generates coverage.html and opens in browser
```

---

## Quick Reference

### Run Tests
```bash
make ci-full                               # Complete CI pipeline (RECOMMENDED)
make test-unit                             # Unit tests only
make test-integration                      # CLI integration tests
make test-air-integration                  # Air web UI integration tests
make test-cleanup                          # Clean up test resources
```

### CLI Integration Test Template
```go
//go:build integration
// +build integration

package integration

func TestFeature(t *testing.T) {
    skipIfMissingCreds(t)
    t.Parallel()

    acquireRateLimit(t)
    resource, err := createResource(t)

    t.Cleanup(func() {
        acquireRateLimit(t)
        _ = deleteResource(t, resource.ID)
    })

    // Test logic here
}
```

### Air Integration Test Template
```go
//go:build integration
// +build integration

package air

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestIntegration_Feature(t *testing.T) {
    server := testServer(t)  // Uses shared helper from integration_base_test.go

    req := httptest.NewRequest(http.MethodGet, "/api/endpoint", nil)
    w := httptest.NewRecorder()

    server.handleEndpoint(w, req)

    if w.Code != http.StatusOK {
        t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
    }

    var resp ResponseType
    if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
        t.Fatalf("failed to decode response: %v", err)
    }

    // Assertions here
}
```

**Note:** Air tests use `testServer(t)` helper from `integration_base_test.go` and test HTTP handlers directly using `httptest`.

---

**Key principles:**
1. Test behavior, not implementation
2. Use table-driven tests for multiple scenarios
3. Mock external dependencies
4. Clean up resources in `t.Cleanup()`
5. Enable `t.Parallel()` for independent tests
6. Rate limit API calls in parallel tests
