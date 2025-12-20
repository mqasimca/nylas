# Run Integration Tests

This workflow runs integration tests with proper environment setup, error handling, and retry logic for flaky tests.

## What It Does

1. Verifies required environment variables are set
2. Runs integration tests with proper build tags
3. Handles API rate limits gracefully
4. Retries flaky tests automatically
5. Reports failures with actionable context
6. Suggests fixes for common failure patterns

## When to Use

- After implementing API-dependent features
- Before releases (end-to-end verification)
- When debugging API integration issues
- As part of pre-deployment checks
- When validating credentials or grants

## Prerequisites

### Required Environment Variables

```bash
export NYLAS_API_KEY="your-api-key-here"
export NYLAS_GRANT_ID="your-grant-id-here"
```

**Where to get credentials:**
1. Go to [Nylas Dashboard](https://dashboard.nylas.com)
2. Navigate to API Keys section
3. Copy your API key
4. Get a grant ID from the Grants section

### Optional Environment Variables

```bash
# For testing specific regions
export NYLAS_REGION="us"  # or "eu", "ireland", etc.

# For rate limit testing
export NYLAS_RATE_LIMIT="100"  # requests per minute

# For retry configuration
export TEST_RETRY_COUNT="3"
export TEST_RETRY_DELAY="5s"
```

## Step-by-Step Workflow

### Step 1: Verify Environment Setup

```bash
# Check if required env vars are set
if [ -z "$NYLAS_API_KEY" ] || [ -z "$NYLAS_GRANT_ID" ]; then
    echo "❌ Missing required environment variables"
    echo "   Set NYLAS_API_KEY and NYLAS_GRANT_ID"
    exit 1
fi

echo "✅ Environment variables set"
echo "   API Key: ${NYLAS_API_KEY:0:10}..."
echo "   Grant ID: $NYLAS_GRANT_ID"
```

### Step 2: Run All Integration Tests

```bash
# Run all integration tests with verbose output
go test ./... -tags=integration -v

# Run with race detection (recommended)
go test ./... -tags=integration -race -v

# Run with timeout (prevent hanging tests)
go test ./... -tags=integration -timeout=10m -v
```

### Step 3: Run Specific Integration Tests

```bash
# Run specific test file
go test ./internal/cli/integration_email_test.go -tags=integration -v

# Run specific test function
go test ./internal/cli/... -tags=integration -run=TestEmail_Integration -v

# Run specific subtest
go test ./internal/cli/... -tags=integration -run=TestEmail_Integration/SendEmail -v
```

### Step 4: Handle Test Failures

When tests fail, analyze the error:

#### Common Failure: Rate Limit Exceeded

```
Error: rate limit exceeded (429)
```

**Fix:**
```bash
# Add delay between test runs
sleep 60
go test ./internal/cli/... -tags=integration -v
```

#### Common Failure: Invalid Grant

```
Error: grant not found or invalid
```

**Fix:**
```bash
# Verify grant ID is valid
echo $NYLAS_GRANT_ID

# Get fresh grant ID from Nylas Dashboard
export NYLAS_GRANT_ID="new-grant-id"
```

#### Common Failure: Network Timeout

```
Error: context deadline exceeded
```

**Fix:**
```bash
# Increase timeout
go test ./... -tags=integration -timeout=20m -v

# Or fix slow tests to clean up faster
```

### Step 5: Retry Flaky Tests

Some tests may be flaky due to network issues or API timing. Use retry logic:

```bash
# Retry failed tests up to 3 times
for i in {1..3}; do
    echo "Attempt $i of 3"
    if go test ./... -tags=integration -v; then
        echo "✅ Tests passed on attempt $i"
        exit 0
    fi
    echo "❌ Tests failed, retrying in 30s..."
    sleep 30
done

echo "❌ Tests failed after 3 attempts"
exit 1
```

## Integration Test Patterns

### Pattern 1: Create-Verify-Delete

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

    // Cleanup (runs even if test fails)
    t.Cleanup(func() {
        _ = client.DeleteResource(ctx, grantID, resource.ID)
    })

    // Verify
    retrieved, err := client.GetResource(ctx, grantID, resource.ID)
    if err != nil {
        t.Errorf("GetResource() error = %v", err)
    }

    // Assert
    if retrieved.ID != resource.ID {
        t.Errorf("ID = %q, want %q", retrieved.ID, resource.ID)
    }
}
```

### Pattern 2: Skip on Missing Features

```go
func TestNotetaker_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    client, grantID := setupIntegrationTest(t)
    ctx := context.Background()

    notetaker, err := client.CreateNotetaker(ctx, grantID, req)
    if err != nil {
        // Some test accounts may not have this feature
        if strings.Contains(err.Error(), "not found") ||
           strings.Contains(err.Error(), "forbidden") ||
           strings.Contains(err.Error(), "not available") {
            t.Skip("Notetaker not available for this account")
        }
        t.Fatalf("CreateNotetaker() error = %v", err)
    }

    // Continue with test...
}
```

### Pattern 3: Rate Limit Handling

```go
func TestBulkOperation_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    client, grantID := setupIntegrationTest(t)
    ctx := context.Background()

    const batchSize = 10
    const delayBetweenBatches = 1 * time.Second

    for i := 0; i < 100; i++ {
        // Create resource
        _, err := client.CreateResource(ctx, grantID, req)
        if err != nil {
            if strings.Contains(err.Error(), "rate limit") {
                t.Logf("Rate limit hit at iteration %d, waiting...", i)
                time.Sleep(60 * time.Second)
                continue
            }
            t.Fatalf("CreateResource() error = %v", err)
        }

        // Add delay every N items
        if (i+1)%batchSize == 0 {
            time.Sleep(delayBetweenBatches)
        }
    }
}
```

## Test Automation Script

Create `scripts/run_integration_tests.sh`:

```bash
#!/bin/bash
# Integration test runner with error handling

set -e

echo "🧪 Running Integration Tests"
echo "============================"
echo

# Check environment variables
if [ -z "$NYLAS_API_KEY" ]; then
    echo "❌ Error: NYLAS_API_KEY not set"
    echo "   export NYLAS_API_KEY='your-api-key'"
    exit 1
fi

if [ -z "$NYLAS_GRANT_ID" ]; then
    echo "❌ Error: NYLAS_GRANT_ID not set"
    echo "   export NYLAS_GRANT_ID='your-grant-id'"
    exit 1
fi

echo "✅ Environment configured"
echo "   API Key: ${NYLAS_API_KEY:0:10}..."
echo "   Grant ID: $NYLAS_GRANT_ID"
echo

# Run tests with retry logic
MAX_RETRIES=3
RETRY_DELAY=30

for i in $(seq 1 $MAX_RETRIES); do
    echo "📊 Test Run $i of $MAX_RETRIES"
    echo

    if go test ./... -tags=integration -timeout=15m -v; then
        echo
        echo "✅ All integration tests passed!"
        exit 0
    fi

    if [ $i -lt $MAX_RETRIES ]; then
        echo
        echo "⚠️  Tests failed, retrying in ${RETRY_DELAY}s..."
        sleep $RETRY_DELAY
    fi
done

echo
echo "❌ Tests failed after $MAX_RETRIES attempts"
exit 1
```

Make it executable:

```bash
chmod +x scripts/run_integration_tests.sh
./scripts/run_integration_tests.sh
```

## Debugging Failed Tests

### Enable Verbose Logging

```bash
# Set log level
export LOG_LEVEL=debug

# Run tests with verbose output
go test ./... -tags=integration -v -count=1

# Disable test cache to ensure fresh run
go clean -testcache
go test ./... -tags=integration -v
```

### Inspect API Requests

Add debug logging to client:

```go
// In internal/adapters/nylas/client.go (temporarily)
func (c *HTTPClient) do(ctx context.Context, req *http.Request) (*http.Response, error) {
    // Debug: print request
    fmt.Printf("→ %s %s\n", req.Method, req.URL)

    resp, err := c.httpClient.Do(req.WithContext(ctx))

    // Debug: print response
    if resp != nil {
        fmt.Printf("← %d\n", resp.StatusCode)
    }

    return resp, err
}
```

### Run Single Test in Isolation

```bash
# Run one specific test to debug
go test -tags=integration -v -run=TestEmail_Integration/SendEmail ./internal/cli/...

# With race detection
go test -tags=integration -race -v -run=TestEmail_Integration/SendEmail ./internal/cli/...
```

## Common Issues and Solutions

### Issue 1: Tests Skip Due to testing.Short()

**Problem:**
```
--- SKIP: TestFeature_Integration (0.00s)
    integration_test.go:15: Skipping integration test
```

**Solution:**
Don't use `-short` flag when running integration tests:
```bash
# ❌ Wrong - tests will skip
go test ./... -short -tags=integration

# ✅ Correct
go test ./... -tags=integration
```

### Issue 2: Build Tags Not Recognized

**Problem:**
Integration tests don't run or are included in unit tests.

**Solution:**
Ensure proper build tags at top of test file:
```go
//go:build integration
// +build integration

package cli
```

### Issue 3: Tests Hang Indefinitely

**Problem:**
Tests never complete, process hangs.

**Solution:**
```bash
# Add timeout
go test ./... -tags=integration -timeout=10m -v

# Find hanging test
go test ./... -tags=integration -v | tee test.log
# Check last test that ran
```

### Issue 4: Credentials Expired

**Problem:**
```
Error: unauthorized (401)
```

**Solution:**
```bash
# Get fresh credentials from Nylas Dashboard
export NYLAS_API_KEY="new-key"
export NYLAS_GRANT_ID="new-grant"

# Verify credentials work
curl -H "Authorization: Bearer $NYLAS_API_KEY" \
     https://api.us.nylas.com/v3/grants/$NYLAS_GRANT_ID
```

## Integration Test Checklist

Before running integration tests:

- [ ] Set `NYLAS_API_KEY` environment variable
- [ ] Set `NYLAS_GRANT_ID` environment variable
- [ ] Verify credentials are valid (curl test)
- [ ] Ensure internet connection is stable
- [ ] Check API status (status.nylas.com)
- [ ] Clear test cache if needed (`go clean -testcache`)

During test execution:

- [ ] Monitor for rate limit errors
- [ ] Watch for flaky tests (inconsistent failures)
- [ ] Note any skipped tests (missing features)
- [ ] Check cleanup runs even on failures

After test completion:

- [ ] Review any skipped tests
- [ ] Verify all created resources were cleaned up
- [ ] Document any new failures or flaky tests
- [ ] Update credentials if they expire

## CI/CD Integration

For GitHub Actions:

```yaml
name: Integration Tests

on:
  schedule:
    - cron: '0 0 * * *'  # Run daily
  workflow_dispatch:     # Manual trigger

jobs:
  integration:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'

      - name: Run integration tests
        env:
          NYLAS_API_KEY: ${{ secrets.NYLAS_API_KEY }}
          NYLAS_GRANT_ID: ${{ secrets.NYLAS_GRANT_ID }}
        run: |
          go test ./... -tags=integration -timeout=15m -v
```

**Important:** Store credentials as GitHub secrets, never in code.

## Summary

**Integration tests verify:**
- Real API interactions work
- Error handling is correct
- Authentication flows succeed
- Data models match API responses
- Cleanup happens properly

**Best practices:**
1. ✅ Always set environment variables
2. ✅ Use `t.Cleanup()` for resource cleanup
3. ✅ Skip gracefully if features unavailable
4. ✅ Handle rate limits with delays
5. ✅ Use timeouts to prevent hanging
6. ✅ Run in isolation (not in CI on every commit)
7. ❌ Don't commit credentials
8. ❌ Don't assume test account state
9. ❌ Don't run in parallel (API rate limits)

**Remember:** Integration tests are slower and more fragile than unit tests. Use them to verify critical end-to-end flows, not every code path.
