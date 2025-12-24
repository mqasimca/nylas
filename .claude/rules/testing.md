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
go test ./... -short -coverprofile=coverage.out
go tool cover -html=coverage.out
```

---

## Quick Reference

### Run Tests
```bash
go test ./... -short                     # Unit tests
make test-integration                     # Integration tests
```

### Integration Test Template
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

---

**Key principles:**
1. Test behavior, not implementation
2. Use table-driven tests for multiple scenarios
3. Mock external dependencies
4. Clean up resources in `t.Cleanup()`
5. Enable `t.Parallel()` for independent tests
6. Rate limit API calls in parallel tests
