# Generate Tests

Generate comprehensive unit and integration tests for Go code.

## Instructions

1. Ask me for:
   - Which file/function to test
   - Test type: unit test or integration test
   - Specific scenarios to cover (optional)

2. Analyze the code and generate tests following project patterns.

## Test Patterns

### Unit Tests (Table-Driven)

Place in same directory as source file with `_test.go` suffix.

```go
func TestFunctionName(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name    string
        input   InputType
        want    OutputType
        wantErr bool
    }{
        {
            name:  "valid input returns expected output",
            input: validInput,
            want:  expectedOutput,
        },
        {
            name:    "empty input returns error",
            input:   "",
            wantErr: true,
        },
        {
            name:  "edge case with special characters",
            input: "unicode: 日本語",
            want:  expectedOutput,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := FunctionUnderTest(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("got = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Integration Tests (CLI)

Place in `internal/cli/integration/<feature>_test.go` with build tags.

```go
//go:build integration
// +build integration

package integration

func TestCLI_CommandName(t *testing.T) {
    skipIfMissingCreds(t)
    t.Parallel()

    stdout, stderr, err := runCLIWithRateLimit(t, "command", "subcommand", "--flag", "value")
    skipIfProviderNotSupported(t, stderr)

    if err != nil {
        t.Fatalf("command failed: %v\nstderr: %s", err, stderr)
    }

    if !strings.Contains(stdout, "expected text") {
        t.Errorf("Expected 'expected text' in output, got: %s", stdout)
    }
}
```

### HTTP Handler Tests (Air)

Place in `internal/air/handlers_<feature>_test.go`.

```go
func TestHandleFeature_Success(t *testing.T) {
    t.Parallel()

    server := newTestDemoServer()

    reqBody := RequestType{Field: "value"}
    body, _ := json.Marshal(reqBody)
    req := httptest.NewRequest(http.MethodPost, "/api/endpoint", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()

    server.handleFeature(w, req)

    if w.Code != http.StatusOK {
        t.Errorf("expected status 200, got %d", w.Code)
    }

    var resp ResponseType
    if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
        t.Fatalf("failed to decode response: %v", err)
    }

    if !resp.Success {
        t.Error("expected Success to be true")
    }
}
```

## Test Categories to Cover

| Category | Description | Examples |
|----------|-------------|----------|
| Happy path | Normal inputs, success cases | Valid email, correct credentials |
| Error cases | Invalid inputs, failures | Empty fields, bad format |
| Edge cases | Boundary conditions | Empty slices, nil values, unicode |
| Method guards | Wrong HTTP methods | GET instead of POST |
| JSON handling | Marshaling/unmarshaling | Invalid JSON, missing fields |

## Mock Pattern

Use mocks from `internal/adapters/nylas/mock.go`:

```go
mockClient := &nylas.MockClient{
    ListMessagesFunc: func(ctx context.Context, grantID string, params *domain.MessageQueryParams) ([]domain.Message, error) {
        return []domain.Message{{ID: "test-id"}}, nil
    },
}
```

## Test Naming Convention

| Type | Pattern | Example |
|------|---------|---------|
| Unit test | `TestFunctionName_Scenario` | `TestParseEmail_ValidInput` |
| CLI integration | `TestCLI_CommandName` | `TestCLI_EmailSend` |
| HTTP handler | `TestHandleFeature_Scenario` | `TestHandleAISummarize_EmptyBody` |

## Run Tests

```bash
# Unit tests
go test ./internal/cli/email/... -v

# Integration tests
make test-integration

# Specific test
go test -tags=integration -v ./internal/cli/integration/... -run "TestCLI_EmailSend"

# With coverage
make test-coverage
```

3. After generating tests, verify:
   - Tests pass: `go test ./path/to/package/...`
   - Linting passes: `golangci-lint run`
   - Coverage improved: `make test-coverage`
