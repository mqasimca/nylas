# Integration Test Patterns

Shared patterns for Go integration tests in the Nylas CLI project.

> **This is the authoritative source for rate limiting patterns.** Other files reference this document.

---

## CLI Integration Tests

### Location & Build Tags

```go
//go:build integration
// +build integration

package integration
```

Location: `internal/cli/integration/{feature}_test.go`

---

## CLI Integration Test Template

```go
func TestCLI_FeatureName(t *testing.T) {
    skipIfMissingCreds(t)
    t.Parallel()

    // Rate limit for API calls
    acquireRateLimit(t)

    // Create test resource
    resource, err := createTestResource(t)
    require.NoError(t, err)

    // Cleanup after test
    t.Cleanup(func() {
        acquireRateLimit(t)
        _ = deleteTestResource(t, resource.ID)
    })

    // Test the CLI command
    stdout, stderr, err := runCLIWithRateLimit(t, "command", "subcommand", "--flag", "value")
    require.NoError(t, err)
    assert.Empty(t, stderr)
    assert.Contains(t, stdout, "expected output")
}
```

---

## Air Integration Tests (Web UI)

### Location & Build Tags

```go
//go:build integration
// +build integration

package air
```

Location: `internal/air/integration_*.go`

### Files (10 total):
- `integration_base_test.go` - Shared helpers (`testServer()`, utilities)
- `integration_core_test.go` - Config, Grants, Folders, Index
- `integration_email_test.go` - Email and draft operations
- `integration_calendar_test.go` - Calendar, events, availability
- `integration_contacts_test.go` - Contact operations
- `integration_cache_test.go` - Cache operations
- `integration_ai_test.go` - AI features (summarize, smart compose, thread analysis)
- `integration_middleware_test.go` - Middleware tests
- `integration_bundles_test.go` - Email bundles, categorization
- `integration_productivity_test.go` - Scheduled send, undo send, snooze

---

## Air Integration Test Template

```go
func TestIntegration_Feature(t *testing.T) {
    server := testServer(t)  // Uses shared helper

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

---

## Rate Limiting (CRITICAL)

ALWAYS use rate limiting for API calls in parallel tests:

```go
acquireRateLimit(t)  // Call before each API operation
```

**Rate limiting config:**
```bash
export NYLAS_TEST_RATE_LIMIT_RPS="2.0"    # Requests per second
export NYLAS_TEST_RATE_LIMIT_BURST="5"    # Burst size
```

| Command Type | Rate Limit |
|--------------|------------|
| API commands (calendar, email, contacts) | ✅ Required |
| Offline commands (timezone, version, help) | ❌ Not needed |

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

---

## Cleanup Pattern

```go
t.Cleanup(func() {
    acquireRateLimit(t)
    _ = client.Delete(ctx, resourceID)
})
```

**CRITICAL:** Air tests create real resources. Always use:
```bash
make ci-full         # RECOMMENDED: Complete CI with automatic cleanup
make test-cleanup    # Manual cleanup if needed
```

---

## Commands

**See:** `.claude/commands/run-tests.md` for full command details.

```bash
make ci-full              # Complete CI pipeline (RECOMMENDED)
make test-integration     # CLI integration tests
make test-air-integration # Air web UI integration tests
```
