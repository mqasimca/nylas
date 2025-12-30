# Testing Guidelines

Consolidated testing rules for the Nylas CLI project.

**Detailed Patterns:**
- Go unit tests: `.claude/shared/patterns/go-test-patterns.md`
- Integration tests: `.claude/shared/patterns/integration-test-patterns.md`
- Playwright E2E: `.claude/shared/patterns/playwright-patterns.md`

---

## Test Organization

### Unit Tests
- **Location:** Alongside source (`*_test.go`)
- **Function:** `TestFunctionName_Scenario`
- **Pattern:** Table-driven with `t.Run(tt.name, ...)`

### CLI Integration Tests
- **Location:** `internal/cli/integration/`
- **Build tags:** `//go:build integration` and `// +build integration`
- **Function:** `TestCLI_CommandName`

### Air Integration Tests
- **Location:** `internal/air/integration_*.go`
- **Build tags:** `//go:build integration` and `// +build integration`
- **Function:** `TestIntegration_FeatureName`

---

## Rate Limiting (CRITICAL)

```go
acquireRateLimit(t)  // Call before each API operation
```

| Command Type | Rate Limit |
|--------------|------------|
| API commands (calendar, email, contacts) | ✅ Required |
| Offline commands (timezone, version, help) | ❌ Not needed |

---

## Test Coverage

| Package Type | Minimum | Target |
|--------------|---------|--------|
| Core Adapters | 70% | 85%+ |
| Business Logic | 60% | 80%+ |
| CLI Commands | 50% | 70%+ |
| Utilities | 90% | 100% |

```bash
make test-coverage  # Generates coverage.html and opens in browser
```

---

## Quick Reference

### Run Tests
```bash
make ci-full                     # Complete CI pipeline (RECOMMENDED)
make test-unit                   # Unit tests only
make test-integration            # CLI integration tests
make test-air-integration        # Air web UI integration tests
make test-cleanup                # Clean up test resources
```

**CRITICAL:** Air tests create real resources. Always use `make ci-full` for automatic cleanup.

---

## Key Principles

1. Test behavior, not implementation
2. Use table-driven tests for multiple scenarios
3. Mock external dependencies
4. Clean up resources in `t.Cleanup()`
5. Enable `t.Parallel()` for independent tests
6. Rate limit API calls in parallel tests
