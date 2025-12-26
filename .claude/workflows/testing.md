# Testing Workflows

Comprehensive guide for running integration tests and analyzing test coverage.

## Quick Reference

```bash
# Run integration tests
go test ./... -tags=integration -v

# Analyze coverage
go test ./... -short -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run specific test
go test -tags=integration -run=TestCLI_Email ./internal/cli/integration/...
```

---

## Integration Testing

### Prerequisites

**Required environment variables:**
```bash
export NYLAS_API_KEY="your-api-key"
export NYLAS_GRANT_ID="your-grant-id"
```

Get credentials from [Nylas Dashboard](https://dashboardv3.nylas.com).

### Running Integration Tests

```bash
# All integration tests
go test ./... -tags=integration -v

# With race detection (recommended)
go test ./... -tags=integration -race -v

# With timeout
go test ./... -tags=integration -timeout=10m -v

# Specific test
go test ./internal/cli/integration -tags=integration -run=TestCLI_Email -v
```

### Common Issues & Solutions

**Rate limit exceeded (429):**
```bash
sleep 60
go test ./... -tags=integration -v
```

**Invalid grant:**
```bash
# Verify grant ID
echo $NYLAS_GRANT_ID

# Get fresh grant from dashboard
export NYLAS_GRANT_ID="new-grant-id"
```

**Network timeout:**
```bash
# Increase timeout
go test ./... -tags=integration -timeout=20m -v
```

### Integration Test Patterns

**Create-Verify-Delete:**
```go
func TestResource_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    client, grantID := setupIntegrationTest(t)
    ctx := context.Background()

    // Create
    resource, err := client.CreateResource(ctx, grantID, req)
    if err != nil {
        t.Fatalf("CreateResource() error = %v", err)
    }

    // Cleanup
    t.Cleanup(func() {
        _ = client.DeleteResource(ctx, grantID, resource.ID)
    })

    // Verify
    if resource.ID == "" {
        t.Error("Created resource has empty ID")
    }
}
```

**Skip on missing features:**
```go
if err != nil {
    if strings.Contains(err.Error(), "not found") ||
       strings.Contains(err.Error(), "forbidden") {
        t.Skip("Feature not available for this account")
    }
    t.Fatalf("CreateResource() error = %v", err)
}
```

### Retry Script

Create `scripts/run_integration_tests.sh`:

```bash
#!/bin/bash
set -e

# Check environment
if [ -z "$NYLAS_API_KEY" ] || [ -z "$NYLAS_GRANT_ID" ]; then
    echo "❌ Set NYLAS_API_KEY and NYLAS_GRANT_ID"
    exit 1
fi

# Retry logic
for i in {1..3}; do
    echo "📊 Test Run $i of 3"
    if go test ./... -tags=integration -timeout=15m -v; then
        echo "✅ All tests passed!"
        exit 0
    fi
    [ $i -lt 3 ] && sleep 30
done

echo "❌ Tests failed after 3 attempts"
exit 1
```

---

## Coverage Analysis

### Coverage Targets

| Package Type | Minimum | Target |
|--------------|---------|--------|
| Core Adapters | 70% | 85%+ |
| Business Logic | 60% | 80%+ |
| CLI Commands | 50% | 70%+ |
| Utilities | 90% | 100% |

### Analyzing Coverage

```bash
# Generate coverage report
go test ./... -short -coverprofile=coverage.out

# View overall coverage
go tool cover -func=coverage.out | grep "total:"

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html
open coverage.html
```

### Finding Low Coverage Packages

```bash
# Check specific package
go test ./internal/adapters/browser -coverprofile=browser_cov.out
go tool cover -func=browser_cov.out
```

### HTML Report Interpretation

- **Green lines** - Covered by tests
- **Red lines** - Not covered (focus here!)
- **Gray lines** - Not executable

### Coverage Analysis Script

Create `scripts/coverage_analysis.sh`:

```bash
#!/bin/bash

echo "🔍 Analyzing test coverage..."

# Generate coverage
go test ./... -short -coverprofile=coverage.out

# Overall coverage
echo "📊 Overall Coverage:"
go tool cover -func=coverage.out | grep "total:"

echo
echo "📈 View detailed report:"
echo "   go tool cover -html=coverage.out -o coverage.html"
echo "   open coverage.html"
```

### Tips for Improving Coverage

1. **Start with zero coverage packages** - High priority
2. **Use table-driven tests** - Cover multiple scenarios quickly
3. **Focus on public APIs** - Internal helpers get covered indirectly
4. **Don't chase 100%** - Target 70-85% for meaningful code
5. **Use coverage to find gaps** - Untested error paths, edge cases

### What Not to Test

- Simple getters/setters
- Trivial constructors
- Generated code
- Main functions

---

## Best Practices

**Integration Tests:**
- ✅ Set environment variables
- ✅ Use `t.Cleanup()` for resource cleanup
- ✅ Skip gracefully if features unavailable
- ✅ Handle rate limits with delays
- ✅ Use timeouts to prevent hanging
- ❌ Don't commit credentials
- ❌ Don't assume test account state

**Coverage:**
- ✅ Test behavior, not just increase numbers
- ✅ Focus on critical paths and error handling
- ✅ Use coverage to find missing test scenarios
- ❌ Don't write tests just to hit 100%
- ❌ Don't test trivial code

---

## CI/CD Integration

**GitHub Actions:**

```yaml
name: Tests

on: [push, pull_request]

jobs:
  unit:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'

      - name: Run unit tests
        run: go test ./... -short -v

      - name: Check coverage
        run: |
          go test ./... -short -coverprofile=coverage.out
          go tool cover -func=coverage.out | grep "total:" | awk '{if ($3 < "70.0%") exit 1}'

  integration:
    runs-on: ubuntu-latest
    if: github.event_name == 'schedule' || github.event_name == 'workflow_dispatch'
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'

      - name: Run integration tests
        env:
          NYLAS_API_KEY: ${{ secrets.NYLAS_API_KEY }}
          NYLAS_GRANT_ID: ${{ secrets.NYLAS_GRANT_ID }}
        run: go test ./... -tags=integration -timeout=15m -v
```

---

## Checklists

**Before Running Integration Tests:**
- [ ] Set `NYLAS_API_KEY` environment variable
- [ ] Set `NYLAS_GRANT_ID` environment variable
- [ ] Verify credentials are valid
- [ ] Ensure internet connection is stable
- [ ] Clear test cache if needed (`go clean -testcache`)

**Coverage Analysis:**
- [ ] Generated coverage report (`coverage.out`)
- [ ] Viewed overall coverage percentage
- [ ] Identified packages below target coverage
- [ ] Created HTML report for detailed view
- [ ] Listed uncovered functions/methods
- [ ] Created action items with priorities

---

## Summary

**Integration tests verify:**
- Real API interactions
- Error handling
- Authentication flows
- Data models match API responses
- Cleanup happens properly

**Coverage analysis helps:**
- Identify untested code
- Find missing test scenarios
- Prioritize testing efforts
- Track quality metrics

**Remember:** Integration tests are slow and fragile—use them for critical end-to-end flows. Coverage is a metric, not a goal—write meaningful tests that verify behavior.
