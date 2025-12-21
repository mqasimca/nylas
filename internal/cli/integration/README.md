# Integration Tests

This directory contains all integration tests for Nylas CLI commands.

## Directory Structure

```
internal/cli/integration/
├── test.go                      # Common setup and test helpers
├── admin_test.go                # Admin command tests
├── auth_test.go                 # Auth command tests
├── auth_enhancements_test.go    # Auth enhancements tests
├── calendar_test.go             # Calendar command tests
├── contacts_test.go             # Contact command tests
├── contact_enhancements_test.go # Contact enhancements tests
├── email_test.go                # Email command tests
├── drafts_test.go               # Draft command tests
├── threads_test.go              # Thread command tests
├── attachments_test.go          # Attachment tests
├── folders_test.go              # Folder command tests
├── inbound_test.go              # Inbound command tests
├── webhooks_test.go             # Webhook command tests
├── scheduler_test.go            # Scheduler command tests
├── notetaker_test.go            # Notetaker command tests
├── otp_test.go                  # OTP command tests
├── metadata_test.go             # Metadata tests
├── recurring_events_test.go     # Recurring event tests
├── scheduled_messages_test.go   # Scheduled message tests
├── smart_compose_test.go        # Smart compose tests
├── virtual_calendar_test.go     # Virtual calendar tests
├── misc_test.go                 # Miscellaneous tests
└── README.md                    # This file
```

## Running Integration Tests

### Prerequisites

Integration tests require valid Nylas API credentials:

```bash
export NYLAS_API_KEY="your-api-key"
export NYLAS_GRANT_ID="your-grant-id"
export NYLAS_CLIENT_ID="your-client-id"  # Optional
```

Optional environment variables for specific tests:

```bash
export NYLAS_TEST_EMAIL="test@example.com"  # Email for send tests
export NYLAS_TEST_SEND_EMAIL="true"         # Enable email send tests
export NYLAS_TEST_DELETE="true"              # Enable delete tests
```

### Run All Integration Tests

```bash
# Run all integration tests
go test -tags=integration -v ./internal/cli/integration/...

# Run with timeout
go test -tags=integration -v -timeout 30m ./internal/cli/integration/...

# Run specific test file
go test -tags=integration -v ./internal/cli/integration/ -run TestAuth
```

### Run Specific Test

```bash
# Run specific test function
go test -tags=integration -v ./internal/cli/integration/ -run TestCLI_EmailList

# Run tests matching pattern
go test -tags=integration -v ./internal/cli/integration/ -run "TestCLI_Email.*"
```

### Build Integration Tests (Without Running)

```bash
go test -tags=integration -c ./internal/cli/integration -o integration.test
```

## Test Organization

### Common Setup (`test.go`)

The `test.go` file contains:
- Test configuration loading from environment variables
- CLI binary building and execution helpers
- Shared test utilities (runCLI, cleanupResource, etc.)
- Common test data setup

### Test Files

Each `*_test.go` file focuses on a specific CLI command group:

- **auth_test.go**: Tests for `nylas auth` commands (login, logout, list, etc.)
- **email_test.go**: Tests for `nylas email` commands (list, show, send, etc.)
- **calendar_test.go**: Tests for `nylas calendar` commands (list, create, update, etc.)
- And so on...

### Test Naming Convention

All test functions follow the pattern:
```go
func TestCLI_<Command>_<Action>(t *testing.T)
```

Examples:
- `TestCLI_EmailList` - Tests email list command
- `TestCLI_CalendarCreate` - Tests calendar creation
- `TestCLI_AuthLogin` - Tests auth login flow

## Best Practices

### 1. Resource Cleanup

Always clean up resources created during tests:

```go
func TestSomething(t *testing.T) {
    // Create resource
    id := createResource(t)

    // Clean up at end
    t.Cleanup(func() {
        deleteResource(t, id)
    })

    // Test logic here
}
```

### 2. Skipping Tests

Skip tests when credentials are missing:

```go
func TestSomething(t *testing.T) {
    if testAPIKey == "" {
        t.Skip("NYLAS_API_KEY not set")
    }
    // Test logic
}
```

### 3. Test Independence

Each test should be independent and not rely on other tests:
- Create your own test data
- Don't assume specific account state
- Clean up after yourself

### 4. Timeouts

Use reasonable timeouts for API calls:
- List operations: 30s
- Create operations: 30s
- Send operations: 60s
- Webhook tests: 120s

## Debugging

### Verbose Output

```bash
# Run with verbose output
go test -tags=integration -v ./internal/cli/integration/...

# Show all log output
go test -tags=integration -v ./internal/cli/integration/... -test.v
```

### Skip Cleanup for Debugging

Temporarily comment out `t.Cleanup()` calls to inspect created resources.

### Run Single Test

```bash
go test -tags=integration -v ./internal/cli/integration/ -run TestCLI_EmailList
```

## CI/CD Integration

### GitHub Actions

Example workflow:

```yaml
name: Integration Tests

on: [push, pull_request]

jobs:
  integration:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.24'

      - name: Run Integration Tests
        env:
          NYLAS_API_KEY: ${{ secrets.NYLAS_API_KEY }}
          NYLAS_GRANT_ID: ${{ secrets.NYLAS_GRANT_ID }}
        run: |
          go test -tags=integration -v -timeout 30m ./internal/cli/integration/...
```

## Coverage

Generate coverage report for integration tests:

```bash
go test -tags=integration -coverprofile=integration-coverage.out ./internal/cli/integration/...
go tool cover -html=integration-coverage.out
```

## Troubleshooting

### "CLI binary not found"

The test framework builds the CLI binary automatically. If you see this error:
1. Check that `go build` works: `make build`
2. Verify the binary path is correct
3. Check file permissions

### "API key not set"

Export your Nylas API credentials:
```bash
export NYLAS_API_KEY="nyk_v0_..."
export NYLAS_GRANT_ID="your-grant-id"
```

### Rate Limiting

If tests fail due to rate limiting:
1. Reduce concurrent test execution
2. Add delays between tests
3. Use `-p 1` flag to run packages sequentially

### Flaky Tests

If tests are flaky:
1. Check for race conditions
2. Increase timeouts
3. Add retry logic for API calls
4. Check test independence

## Contributing

When adding new integration tests:

1. Create tests in the appropriate `*_test.go` file
2. Follow the existing naming convention
3. Add proper cleanup using `t.Cleanup()`
4. Document any new environment variables needed
5. Ensure tests are independent and can run in any order
6. Add comments explaining non-obvious test logic

## References

- [Go Testing Package](https://pkg.go.dev/testing)
- [Nylas API Documentation](https://developer.nylas.com/docs/api/v3/)
- [Main CLI README](../../../README.md)
