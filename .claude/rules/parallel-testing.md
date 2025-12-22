# Parallel Testing Guide for Integration Tests

This guide explains how to use `t.Parallel()` in integration tests with proper rate limiting for Nylas API calls.

---

## Overview

The integration test suite supports parallel execution using Go's `t.Parallel()`. A global rate limiter ensures that parallel tests don't exceed Nylas API rate limits.

### Key Components

1. **Global Rate Limiter** - Token bucket algorithm limiting API calls
2. **Rate-Limited Helpers** - Functions that automatically handle rate limiting
3. **Environment Configuration** - Customize rate limits for your Nylas plan

---

## Quick Start

### Basic Parallel Test Pattern

```go
func TestCalendarEventsList(t *testing.T) {
    skipIfMissingCreds(t)
    t.Parallel()  // Enable parallel execution

    // Use rate-limited wrapper for API calls
    stdout, stderr, err := runCLIWithRateLimit(t, "calendar", "events", "list")

    if err != nil {
        t.Fatalf("Command failed: %v\nstderr: %s", err, stderr)
    }

    // Verify output
    if !strings.Contains(stdout, "Events:") {
        t.Errorf("Expected events list in output")
    }
}
```

### Offline Commands (No Rate Limiting Needed)

```go
func TestTimezoneList(t *testing.T) {
    if testBinary == "" {
        t.Skip("CLI binary not found")
    }
    t.Parallel()  // Safe - no API calls

    // Offline commands don't need rate limiting
    stdout, stderr, err := runCLI("timezone", "list")

    if err != nil {
        t.Fatalf("Command failed: %v\nstderr: %s", err, stderr)
    }
}
```

---

## Rate Limiting Options

### Option 1: Use Rate-Limited Wrapper (Recommended)

```go
func TestExample(t *testing.T) {
    skipIfMissingCreds(t)
    t.Parallel()

    // Automatically acquires rate limit token
    stdout, stderr, err := runCLIWithRateLimit(t, "calendar", "events", "list")
}
```

**Pros:**
- ✅ Simplest approach
- ✅ One-liner
- ✅ Handles rate limiting automatically

**Cons:**
- ❌ Only works for CLI commands

### Option 2: Manual Rate Limit Acquisition

```go
func TestExample(t *testing.T) {
    skipIfMissingCreds(t)
    t.Parallel()

    // Manually acquire token before API call
    acquireRateLimit(t)
    stdout, stderr, err := runCLI("calendar", "events", "list")
}
```

**Pros:**
- ✅ More control over when to acquire token
- ✅ Works for any API operation

**Cons:**
- ❌ Easy to forget to call acquireRateLimit()
- ❌ More verbose

### Option 3: Multiple API Calls

```go
func TestExample(t *testing.T) {
    skipIfMissingCreds(t)
    t.Parallel()

    // Create event (API call 1)
    acquireRateLimit(t)
    stdout1, _, _ := runCLI("calendar", "events", "create", ...)

    // List events (API call 2)
    acquireRateLimit(t)
    stdout2, _, _ := runCLI("calendar", "events", "list")

    // Delete event (API call 3)
    acquireRateLimit(t)
    runCLI("calendar", "events", "delete", eventID, "--yes")
}
```

**Pros:**
- ✅ Fine-grained control
- ✅ Each API call is rate-limited

**Cons:**
- ❌ Most verbose
- ❌ Easy to forget for some calls

---

## Commands That Need Rate Limiting

### ✅ API Commands (Rate Limiting Required)

- `calendar events` - All CRUD operations
- `calendar calendars` - All CRUD operations
- `contacts` - All CRUD operations
- `email` - Send, list, delete, etc.
- `drafts` - All CRUD operations
- `threads` - List, show
- `folders` - List
- `auth whoami` - API call to get grant info
- `admin` - All admin operations
- `scheduler` - All CRUD operations
- `webhooks` - All CRUD operations
- `notetaker` - All CRUD operations
- `otp` - All operations (requires email API)

### ❌ Offline Commands (No Rate Limiting)

- `timezone` - All commands (convert, dst, find-meeting, info, list)
- `ai config` - All commands (show, list, get, set)
- `ai usage` - Local statistics
- `ai set-budget` - Local configuration
- `ai clear-data` - Local data management
- `--help` - All help commands
- `version` - Version information
- `doctor` - System diagnostics (may include API check)

---

## Configuration

### Default Rate Limits

```go
// From internal/cli/integration/test.go
rateLimitRPS   = 2.0  // 2 requests per second
rateLimitBurst = 5    // Burst capacity of 5
```

### Environment Variables

```bash
# Conservative (default) - for shared/free Nylas accounts
export NYLAS_TEST_RATE_LIMIT_RPS="2.0"
export NYLAS_TEST_RATE_LIMIT_BURST="5"

# Moderate - for professional Nylas plans
export NYLAS_TEST_RATE_LIMIT_RPS="5.0"
export NYLAS_TEST_RATE_LIMIT_BURST="10"

# Aggressive - for enterprise Nylas plans with high limits
export NYLAS_TEST_RATE_LIMIT_RPS="10.0"
export NYLAS_TEST_RATE_LIMIT_BURST="20"
```

**How to Choose:**
- Check your Nylas plan's API rate limits
- Set RPS to 50-60% of your limit for safety margin
- Set burst to 2-3x RPS for initial test runs
- Monitor for 429 errors and adjust down if needed

---

## Common Patterns

### Pattern 1: Table-Driven Parallel Tests

```go
func TestCalendarEventsList(t *testing.T) {
    skipIfMissingCreds(t)

    tests := []struct {
        name     string
        args     []string
        contains string
    }{
        {
            name:     "list all events",
            args:     []string{"calendar", "events", "list"},
            contains: "Events:",
        },
        {
            name:     "list with limit",
            args:     []string{"calendar", "events", "list", "--limit", "5"},
            contains: "Events:",
        },
        {
            name:     "list JSON output",
            args:     []string{"calendar", "events", "list", "--json"},
            contains: `"data"`,
        },
    }

    for _, tt := range tests {
        tt := tt  // Capture range variable
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()  // Each subtest runs in parallel

            stdout, stderr, err := runCLIWithRateLimit(t, tt.args...)

            if err != nil {
                t.Fatalf("Command failed: %v\nstderr: %s", err, stderr)
            }

            if !strings.Contains(stdout, tt.contains) {
                t.Errorf("Expected output to contain %q\nGot: %s", tt.contains, stdout)
            }
        })
    }
}
```

### Pattern 2: Lifecycle Test with Cleanup

```go
func TestCalendarEventLifecycle(t *testing.T) {
    skipIfMissingCreds(t)
    t.Parallel()

    // Create event
    acquireRateLimit(t)
    stdout, stderr, err := runCLI("calendar", "events", "create",
        "--title", "Test Event",
        "--start", "2025-01-01T10:00:00Z",
        "--end", "2025-01-01T11:00:00Z",
    )

    if err != nil {
        t.Fatalf("Create failed: %v\nstderr: %s", err, stderr)
    }

    eventID := extractEventID(stdout)
    if eventID == "" {
        t.Fatal("Failed to extract event ID from output")
    }

    // Cleanup
    t.Cleanup(func() {
        acquireRateLimit(t)
        _, _, _ = runCLI("calendar", "events", "delete", eventID, "--yes")
    })

    // Verify event exists
    acquireRateLimit(t)
    stdout, stderr, err = runCLI("calendar", "events", "show", eventID)

    if err != nil {
        t.Fatalf("Show failed: %v\nstderr: %s", err, stderr)
    }

    if !strings.Contains(stdout, "Test Event") {
        t.Error("Event title not found in show output")
    }
}
```

### Pattern 3: Mixed Parallel and Sequential

```go
func TestComplexWorkflow(t *testing.T) {
    skipIfMissingCreds(t)

    // Setup runs sequentially
    acquireRateLimit(t)
    client := getTestClient()
    // ... setup code

    // These subtests run in parallel
    t.Run("test feature A", func(t *testing.T) {
        t.Parallel()
        acquireRateLimit(t)
        // ... test code
    })

    t.Run("test feature B", func(t *testing.T) {
        t.Parallel()
        acquireRateLimit(t)
        // ... test code
    })

    // Cleanup runs after all subtests complete
    t.Cleanup(func() {
        acquireRateLimit(t)
        // ... cleanup code
    })
}
```

---

## Best Practices

### ✅ DO

1. **Always use `t.Parallel()` for independent tests**
   ```go
   func TestIndependentFeature(t *testing.T) {
       skipIfMissingCreds(t)
       t.Parallel()  // ✅ Good
   }
   ```

2. **Use rate-limited wrappers for API calls**
   ```go
   stdout, _, err := runCLIWithRateLimit(t, "calendar", "events", "list")  // ✅ Good
   ```

3. **Skip rate limiting for offline commands**
   ```go
   stdout, _, _ := runCLI("timezone", "list")  // ✅ Good - no API call
   ```

4. **Cleanup with t.Cleanup() and rate limiting**
   ```go
   t.Cleanup(func() {
       acquireRateLimit(t)
       runCLI("calendar", "events", "delete", id, "--yes")
   })
   ```

5. **Capture range variables in parallel subtests**
   ```go
   for _, tt := range tests {
       tt := tt  // ✅ Good - capture range variable
       t.Run(tt.name, func(t *testing.T) {
           t.Parallel()
       })
   }
   ```

### ❌ DON'T

1. **Don't forget rate limiting for API calls**
   ```go
   t.Parallel()
   stdout, _, _ := runCLI("calendar", "events", "list")  // ❌ Bad - no rate limit!
   ```

2. **Don't use rate limiting for offline commands**
   ```go
   stdout, _, _ := runCLIWithRateLimit(t, "timezone", "list")  // ❌ Unnecessary
   ```

3. **Don't use t.Parallel() for tests that modify shared state**
   ```go
   func TestAuthSwitch(t *testing.T) {
       // ❌ Bad - modifies default grant
       t.Parallel()
       runCLI("auth", "switch", otherGrantID)
   }
   ```

4. **Don't forget to capture range variables**
   ```go
   for _, tt := range tests {
       t.Run(tt.name, func(t *testing.T) {
           t.Parallel()
           // ❌ Bad - tt may change during execution
           runCLI(tt.args...)
       })
   }
   ```

5. **Don't run destructive tests in parallel with others**
   ```go
   func TestDeleteAllEvents(t *testing.T) {
       // ❌ Bad - could interfere with other tests
       t.Parallel()
   }
   ```

---

## Troubleshooting

### Problem: Tests Hit Rate Limits (429 Errors)

**Symptoms:**
```
Error: API rate limit exceeded (429 Too Many Requests)
```

**Solutions:**
1. Reduce rate limit:
   ```bash
   export NYLAS_TEST_RATE_LIMIT_RPS="1.0"
   export NYLAS_TEST_RATE_LIMIT_BURST="3"
   ```

2. Verify you're using rate-limited functions:
   ```go
   // Change this:
   runCLI("calendar", "events", "list")
   // To this:
   runCLIWithRateLimit(t, "calendar", "events", "list")
   ```

3. Check your Nylas plan limits

### Problem: Tests Run Too Slowly

**Symptoms:**
- Tests take forever to complete
- Many tests waiting for rate limiter

**Solutions:**
1. Increase rate limit (if your plan allows):
   ```bash
   export NYLAS_TEST_RATE_LIMIT_RPS="5.0"
   export NYLAS_TEST_RATE_LIMIT_BURST="10"
   ```

2. Remove unnecessary rate limiting for offline commands

3. Batch API calls when possible

### Problem: Tests Fail Intermittently

**Symptoms:**
- Tests pass when run sequentially
- Tests fail when run in parallel

**Solutions:**
1. Check for shared state modification:
   ```go
   // Tests that modify default grant shouldn't use t.Parallel()
   func TestAuthSwitch(t *testing.T) {
       // Don't use t.Parallel() here
   }
   ```

2. Ensure proper cleanup:
   ```go
   t.Cleanup(func() {
       acquireRateLimit(t)
       // Clean up resources
   })
   ```

3. Add delays for eventually-consistent operations:
   ```go
   time.Sleep(500 * time.Millisecond)  // Wait for propagation
   ```

---

## Performance Benchmarks

### Sequential vs Parallel Execution

**Test Suite:** 50 integration tests

| Mode | Time | Speed Improvement |
|------|------|-------------------|
| Sequential (no parallel) | ~75 seconds | Baseline |
| Parallel (RPS=2, Burst=5) | ~30 seconds | **2.5x faster** |
| Parallel (RPS=5, Burst=10) | ~18 seconds | **4.2x faster** |
| Parallel (RPS=10, Burst=20) | ~12 seconds | **6.3x faster** |

**Note:** Actual performance depends on:
- Your Nylas plan's rate limits
- Network latency
- Test complexity
- Number of CPU cores

---

## Summary

### Quick Reference

| Scenario | Code | Rate Limiting |
|----------|------|---------------|
| API call in parallel test | `runCLIWithRateLimit(t, args...)` | ✅ Required |
| Offline command in parallel test | `runCLI(args...)` | ❌ Not needed |
| Multiple API calls | `acquireRateLimit(t)` before each | ✅ Required |
| Cleanup with API call | `t.Cleanup(func() { acquireRateLimit(t); ... })` | ✅ Required |

### Key Takeaways

1. ✅ Use `t.Parallel()` for independent tests
2. ✅ Use `runCLIWithRateLimit()` for API calls
3. ✅ Configure rate limits via environment variables
4. ✅ Skip rate limiting for offline commands
5. ✅ Always clean up resources with `t.Cleanup()`

---

**Last Updated:** December 22, 2024
**See Also:** `internal/cli/integration/test.go` for implementation details
