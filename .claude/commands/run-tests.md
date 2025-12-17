# Run Tests

Run unit tests and/or integration tests for the nylas CLI.

## Instructions

### Unit Tests

Run all unit tests:
```bash
go test ./...
```

Run tests for specific package:
```bash
go test ./internal/cli/email/...
go test ./internal/cli/webhook/...
go test ./internal/domain/...
go test ./internal/adapters/nylas/...
```

Run with verbose output:
```bash
go test -v ./...
```

Run with coverage:
```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Integration Tests

Integration tests require credentials. Set environment variables:
```bash
export NYLAS_API_KEY="your-api-key"
export NYLAS_GRANT_ID="your-grant-id"
export NYLAS_TEST_BINARY="$(pwd)/bin/nylas"
```

Build binary first:
```bash
go build -o bin/nylas ./cmd/nylas
```

Run integration tests:
```bash
go test -tags=integration ./internal/cli/...
```

Run specific integration test:
```bash
go test -tags=integration -v ./internal/cli/... -run "TestCLI_EmailList"
```

### Common Test Patterns

If tests fail, check:
1. **Build errors**: Run `go build ./...` first
2. **Missing mocks**: Update `mock.go` if interface changed
3. **API changes**: Update expected values in tests
4. **Credentials**: Ensure env vars are set for integration tests

After fixing, always run full test suite:
```bash
go build ./... && go test ./...
```
