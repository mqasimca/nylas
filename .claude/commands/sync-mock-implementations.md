# Sync Mock Implementations

Detect interface methods missing from mock implementations and generate stubs.

Focus: $ARGUMENTS

## Instructions

1. **Extract interface methods from ports**

   Read the NylasClient interface:
   ```bash
   # View the interface definition
   cat internal/ports/nylas.go
   ```

   Extract all method signatures from `NylasClient` interface.

2. **Extract mock implementation methods**

   ```bash
   # View mock implementation
   cat internal/adapters/nylas/mock.go

   # List all methods on MockClient
   grep -E "^func \(m \*MockClient\)" internal/adapters/nylas/mock.go
   ```

3. **Compare and find missing methods**

   Create a list of:
   - Methods in interface but not in mock
   - Methods with mismatched signatures
   - Methods in mock but not in interface (stale)

4. **Generate mock stubs for missing methods**

   For each missing method, generate:

   ```go
   func (m *MockClient) {MethodName}(ctx context.Context, {params}) ({returnTypes}, error) {
       // Default mock implementation
       return {defaultReturn}, nil
   }
   ```

   Patterns by operation type:

   **List operations:**
   ```go
   func (m *MockClient) List{Resource}s(ctx context.Context, grantID string, params *domain.{Resource}QueryParams) ([]domain.{Resource}, string, error) {
       if m.ListError != nil {
           return nil, "", m.ListError
       }
       return []domain.{Resource}{
           {ID: "mock-{resource}-1"},
           {ID: "mock-{resource}-2"},
       }, "", nil
   }
   ```

   **Get operations:**
   ```go
   func (m *MockClient) Get{Resource}(ctx context.Context, grantID, {resource}ID string) (*domain.{Resource}, error) {
       if m.GetError != nil {
           return nil, m.GetError
       }
       return &domain.{Resource}{ID: {resource}ID}, nil
   }
   ```

   **Create operations:**
   ```go
   func (m *MockClient) Create{Resource}(ctx context.Context, grantID string, req *domain.Create{Resource}Request) (*domain.{Resource}, error) {
       if m.CreateError != nil {
           return nil, m.CreateError
       }
       return &domain.{Resource}{ID: "new-{resource}-id"}, nil
   }
   ```

   **Update operations:**
   ```go
   func (m *MockClient) Update{Resource}(ctx context.Context, grantID, {resource}ID string, req *domain.Update{Resource}Request) (*domain.{Resource}, error) {
       if m.UpdateError != nil {
           return nil, m.UpdateError
       }
       return &domain.{Resource}{ID: {resource}ID}, nil
   }
   ```

   **Delete operations:**
   ```go
   func (m *MockClient) Delete{Resource}(ctx context.Context, grantID, {resource}ID string) error {
       return m.DeleteError
   }
   ```

5. **Also check demo.go implementation**

   ```bash
   # View demo implementation
   grep -E "^func \(d \*DemoClient\)" internal/adapters/nylas/demo.go
   ```

   Generate demo stubs with realistic data.

6. **Verify interface compliance**

   After adding stubs:
   ```bash
   # This will fail if implementations don't satisfy interface
   go build ./...

   # Run tests
   go test ./internal/adapters/nylas/... -v
   ```

## Automated Detection Script

```bash
# Extract interface methods
grep -E "^\s+[A-Z][a-zA-Z]+\(" internal/ports/nylas.go | sed 's/(.*//' | sort > /tmp/interface_methods.txt

# Extract mock methods
grep -E "^func \(m \*MockClient\)" internal/adapters/nylas/mock.go | sed 's/.*\) //' | sed 's/(.*//' | sort > /tmp/mock_methods.txt

# Find missing in mock
comm -23 /tmp/interface_methods.txt /tmp/mock_methods.txt

# Clean up
rm /tmp/interface_methods.txt /tmp/mock_methods.txt
```

## Report Format

```markdown
# Mock Sync Report

## Interface: NylasClient
Total methods: {N}

## MockClient Status
- Implemented: {X}
- Missing: {Y}
- Signature mismatch: {Z}

### Missing Methods
1. `{MethodName}({params}) ({returns})`
2. `{MethodName}({params}) ({returns})`

### Signature Mismatches
1. `{MethodName}`:
   - Interface: `({params}) ({returns})`
   - Mock: `({params}) ({returns})`

## DemoClient Status
- Implemented: {X}
- Missing: {Y}

### Missing Methods
1. `{MethodName}({params}) ({returns})`
```

## Generated Stubs Location

Add generated stubs to:
- `internal/adapters/nylas/mock.go` - For MockClient
- `internal/adapters/nylas/demo.go` - For DemoClient

## Verification

```bash
# Compile check (catches missing interface methods)
go build ./...

# Run mock tests
go test ./internal/adapters/nylas/... -v

# Verify no compilation errors
go vet ./...
```

## Checklist

- [ ] Extracted all interface methods from `internal/ports/nylas.go`
- [ ] Extracted all mock methods from `internal/adapters/nylas/mock.go`
- [ ] Identified missing methods
- [ ] Generated mock stubs for missing methods
- [ ] Checked demo.go for missing methods
- [ ] Generated demo stubs if needed
- [ ] Verified compilation: `go build ./...`
- [ ] Ran tests: `go test ./internal/adapters/nylas/...`
- [ ] No interface compliance errors
