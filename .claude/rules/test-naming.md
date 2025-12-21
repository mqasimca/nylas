# Test Naming Conventions

## Unit Tests
- File: `*_test.go` (in same package)
- Function: `TestFunctionName_Scenario`
- Table tests: use `t.Run(tt.name, ...)`

## Integration Tests
- Location: `internal/cli/integration/`
- File: `*_test.go` (e.g., `email_test.go`, `auth_test.go`)
- Build tag: `//go:build integration` and `// +build integration`
- Package: `package integration`
- Function: `TestCLI_CommandName` (e.g., `TestCLI_EmailList`)
- Skip tag: `if testAPIKey == "" { t.Skip() }`

## Examples

**Good**:
```go
func TestUserService_CreateUser_Success(t *testing.T)
func TestUserService_CreateUser_DuplicateEmail(t *testing.T)
func TestAuth_Integration(t *testing.T)
```

**Bad**:
```go
func TestCreateUser(t *testing.T)  // Not specific enough
func Test1(t *testing.T)            // No context
```

## Table Tests

Prefer table-driven tests for multiple scenarios:

```go
func TestValidateEmail(t *testing.T) {
    tests := []struct {
        name    string
        email   string
        wantErr bool
    }{
        {"valid email", "user@example.com", false},
        {"missing @", "userexample.com", true},
        {"empty", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateEmail(tt.email)
            if (err != nil) != tt.wantErr {
                t.Errorf("got error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

## Test Helper Functions

Use `t.Helper()` to mark test helper functions:

```go
func assertNoError(t *testing.T, err error, msg string) {
    t.Helper()
    if err != nil {
        t.Fatalf("%s: %v", msg, err)
    }
}

func assertEqual[T comparable](t *testing.T, got, want T, msg string) {
    t.Helper()
    if got != want {
        t.Errorf("%s: got %v, want %v", msg, got, want)
    }
}
```

## Best Practices

### 1. Descriptive Test Names
Test names should describe:
- What is being tested
- What scenario is being tested
- What the expected outcome is

```go
// ✅ GOOD - Clear what's being tested
func TestEmailSender_SendEmail_WithInvalidAddress_ReturnsError(t *testing.T)

// ❌ BAD - Unclear what's being tested
func TestSendEmail(t *testing.T)
```

### 2. Use Subtests for Related Tests

```go
func TestEmailValidation(t *testing.T) {
    t.Run("valid email", func(t *testing.T) {
        // Test valid email
    })

    t.Run("missing @ symbol", func(t *testing.T) {
        // Test invalid email
    })

    t.Run("empty string", func(t *testing.T) {
        // Test empty email
    })
}
```

### 3. Cleanup with t.Cleanup()

```go
func TestFileOperation(t *testing.T) {
    tmpFile, err := os.CreateTemp("", "test")
    assertNoError(t, err, "create temp file")

    t.Cleanup(func() {
        os.Remove(tmpFile.Name())
    })

    // Test file operations...
}
```

### 4. Use t.TempDir() for Temporary Directories

```go
func TestConfigFile(t *testing.T) {
    tmpDir := t.TempDir()  // Automatically cleaned up
    configPath := filepath.Join(tmpDir, "config.yaml")

    // Test config file operations...
}
```

### 5. Parallel Tests When Possible

```go
func TestIndependentFunction(t *testing.T) {
    t.Parallel()  // Run this test in parallel with other parallel tests

    // Test independent functionality...
}
```

## Integration Test Pattern

```go
//go:build integration
// +build integration

package cli

import (
    "context"
    "testing"
)

func TestFeature_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    client, grantID := setupIntegrationTest(t)
    ctx := context.Background()

    t.Run("CreateAndDelete", func(t *testing.T) {
        // Create resource
        resource, err := client.CreateResource(ctx, grantID, req)
        if err != nil {
            t.Fatalf("CreateResource() error = %v", err)
        }

        // Cleanup
        t.Cleanup(func() {
            _ = client.DeleteResource(ctx, grantID, resource.ID)
        })

        // Verify resource was created
        if resource.ID == "" {
            t.Error("Created resource has empty ID")
        }
    })
}
```

## Benchmark Tests

For performance-critical code:

```go
func BenchmarkEmailParsing(b *testing.B) {
    email := "user@example.com"

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        ParseEmail(email)
    }
}
```

## Coverage Best Practices

### 1. Focus on Behavior, Not Lines

Don't write tests just to increase coverage percentage. Write tests that verify:
- Expected behavior
- Error handling
- Edge cases
- Integration points

### 2. Test Public APIs

Focus on testing public APIs (exported functions/methods). Internal implementation details can change without breaking tests.

### 3. Mock External Dependencies

Use mocks for:
- API calls
- Database operations
- File system operations
- Time-dependent code

```go
type MockClient struct {
    GetUserFunc func(ctx context.Context, id string) (*User, error)
}

func (m *MockClient) GetUser(ctx context.Context, id string) (*User, error) {
    if m.GetUserFunc != nil {
        return m.GetUserFunc(ctx, id)
    }
    return nil, nil
}
```

## Summary

**Key Principles:**
1. Test names should be descriptive and self-documenting
2. Use table-driven tests for multiple scenarios
3. Use subtests to organize related tests
4. Mark integration tests with build tags
5. Clean up resources with t.Cleanup()
6. Use t.TempDir() for temporary directories
7. Focus on testing behavior, not implementation details
8. Mock external dependencies appropriately
