# Mocking Guidelines

## When to Mock

### ✅ DO Mock These:
- External APIs (Nylas API, third-party services)
- File system operations
- Network calls
- Database operations
- Time-dependent operations (`time.Now()`, timers)
- Random number generation
- OS-specific operations

### ❌ DON'T Mock These:
- Simple value objects
- Pure functions (no side effects)
- Standard library types (use test doubles instead)
- Internal business logic (test directly)

## Mock Patterns

### 1. Interface-Based Mocks

Prefer interface-based mocks over concrete implementations:

```go
// Good - testable
type UserService struct {
    client ports.NylasClient  // Interface
}

// Test
func TestUserService(t *testing.T) {
    mockClient := nylas.NewMockClient()
    service := NewUserService(mockClient)
    // ... test service
}
```

**Why?** Interfaces allow dependency injection and make code testable without complex mocking frameworks.

### 2. Function Injection

For simple functions, use function injection:

```go
// Production code
var timeNow = time.Now

func ProcessExpiredItems() {
    now := timeNow()
    // ... process items based on current time
}

// Test
func TestProcessExpiredItems(t *testing.T) {
    oldTimeNow := timeNow
    timeNow = func() time.Time {
        return time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
    }
    defer func() { timeNow = oldTimeNow }()

    ProcessExpiredItems()
    // ... verify behavior
}
```

**Why?** Simple, no frameworks needed, easy to understand.

### 3. Callback-Based Mocks

MockClient pattern for flexible test scenarios:

```go
type MockClient struct {
    GetMessagesFunc func(ctx context.Context, grantID string, limit int) ([]domain.Message, error)
    CreateMessageFunc func(ctx context.Context, grantID string, req *domain.CreateMessageRequest) (*domain.Message, error)
}

func (m *MockClient) GetMessages(ctx context.Context, grantID string, limit int) ([]domain.Message, error) {
    if m.GetMessagesFunc != nil {
        return m.GetMessagesFunc(ctx, grantID, limit)
    }
    return nil, nil
}

// Test with custom behavior
func TestWithError(t *testing.T) {
    client := nylas.NewMockClient()
    client.GetMessagesFunc = func(ctx context.Context, grantID string, limit int) ([]domain.Message, error) {
        return nil, errors.New("API error")
    }
    // Test error handling
}
```

**Why?** Allows precise control over mock behavior per test case.

### 4. Table-Driven Mock Configuration

```go
func TestUserService_GetUser(t *testing.T) {
    tests := []struct {
        name        string
        userID      string
        setupMock   func(*MockClient)
        wantErr     bool
        errContains string
    }{
        {
            name:   "successful retrieval",
            userID: "user-123",
            setupMock: func(m *MockClient) {
                m.GetUserFunc = func(ctx context.Context, id string) (*User, error) {
                    return &User{ID: id, Name: "Test User"}, nil
                }
            },
            wantErr: false,
        },
        {
            name:   "user not found",
            userID: "invalid-id",
            setupMock: func(m *MockClient) {
                m.GetUserFunc = func(ctx context.Context, id string) (*User, error) {
                    return nil, domain.ErrUserNotFound
                }
            },
            wantErr:     true,
            errContains: "not found",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            client := NewMockClient()
            tt.setupMock(client)

            service := NewUserService(client)
            user, err := service.GetUser(context.Background(), tt.userID)

            if tt.wantErr {
                if err == nil {
                    t.Error("expected error, got nil")
                }
                return
            }

            if err != nil {
                t.Errorf("unexpected error: %v", err)
            }
            // ... verify user
        })
    }
}
```

## Mock Implementation Best Practices

### 1. Implement All Interface Methods

```go
// ✅ GOOD - All methods implemented
type MockGrantStore struct {
    grants []domain.GrantInfo
}

func (m *MockGrantStore) GetGrant(grantID string) (*domain.GrantInfo, error) {
    for _, grant := range m.grants {
        if grant.ID == grantID {
            return &grant, nil
        }
    }
    return nil, domain.ErrGrantNotFound
}

func (m *MockGrantStore) ListGrants() ([]domain.GrantInfo, error) {
    return m.grants, nil
}

func (m *MockGrantStore) SaveGrant(grant domain.GrantInfo) error {
    m.grants = append(m.grants, grant)
    return nil
}

// ... implement all other interface methods
```

### 2. Provide Sensible Defaults

```go
type MockClient struct {
    GetMessagesFunc func(ctx context.Context, grantID string, limit int) ([]domain.Message, error)
}

func NewMockClient() *MockClient {
    return &MockClient{
        // Default implementation returns empty list
        GetMessagesFunc: func(ctx context.Context, grantID string, limit int) ([]domain.Message, error) {
            return []domain.Message{}, nil
        },
    }
}

func (m *MockClient) GetMessages(ctx context.Context, grantID string, limit int) ([]domain.Message, error) {
    if m.GetMessagesFunc != nil {
        return m.GetMessagesFunc(ctx, grantID, limit)
    }
    return nil, nil
}
```

### 3. Allow Assertion on Mock Calls

```go
type MockBrowser struct {
    OpenCalled bool
    LastURL    string
    OpenFunc   func(url string) error
}

func (m *MockBrowser) Open(url string) error {
    m.OpenCalled = true
    m.LastURL = url
    if m.OpenFunc != nil {
        return m.OpenFunc(url)
    }
    return nil
}

// Test
func TestOpenBrowser(t *testing.T) {
    browser := &MockBrowser{}

    OpenAuthURL(browser, "https://auth.example.com")

    if !browser.OpenCalled {
        t.Error("Open() was not called")
    }
    if browser.LastURL != "https://auth.example.com" {
        t.Errorf("URL = %q, want %q", browser.LastURL, "https://auth.example.com")
    }
}
```

## Anti-Patterns

### ❌ Don't Mock Everything

```go
// BAD - over-mocking
type MockString struct {
    Value string
}

func (m *MockString) String() string {
    return m.Value
}

// GOOD - just use string
value := "test"
```

### ❌ Don't Create Mocks for Standard Library

```go
// BAD
type MockHTTPClient struct {
    DoFunc func(*http.Request) (*http.Response, error)
}

// GOOD - use httptest
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(response)
}))
defer server.Close()

client := &http.Client{}
resp, err := client.Get(server.URL)
```

### ❌ Don't Mock Internal Implementation Details

```go
// BAD - mocking internal helper
type MockStringBuilder struct {
    BuildFunc func(parts []string) string
}

// GOOD - test the public API that uses the helper
func TestFormatMessage(t *testing.T) {
    result := FormatMessage("Hello", "World")
    if result != "Hello World" {
        t.Errorf("got %q, want %q", result, "Hello World")
    }
}
```

### ❌ Don't Use Global State in Mocks

```go
// BAD - global state makes tests non-deterministic
var mockResponses []string

type BadMock struct{}

func (m *BadMock) GetData() string {
    if len(mockResponses) == 0 {
        return ""
    }
    resp := mockResponses[0]
    mockResponses = mockResponses[1:]
    return resp
}

// GOOD - use instance state
type GoodMock struct {
    responses []string
    callIndex int
}

func (m *GoodMock) GetData() string {
    if m.callIndex >= len(m.responses) {
        return ""
    }
    resp := m.responses[m.callIndex]
    m.callIndex++
    return resp
}
```

## Testing HTTP Handlers

Use `httptest` for HTTP handler testing:

```go
func TestHTTPHandler(t *testing.T) {
    handler := NewMyHandler()

    req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
    w := httptest.NewRecorder()

    handler.ServeHTTP(w, req)

    if w.Code != http.StatusOK {
        t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
    }

    var users []User
    if err := json.NewDecoder(w.Body).Decode(&users); err != nil {
        t.Fatalf("Decode error: %v", err)
    }

    // Verify users...
}
```

## Mocking Time

### Pattern 1: Time Provider Interface

```go
type TimeProvider interface {
    Now() time.Time
}

type RealTime struct{}

func (RealTime) Now() time.Time {
    return time.Now()
}

type MockTime struct {
    CurrentTime time.Time
}

func (m *MockTime) Now() time.Time {
    return m.CurrentTime
}

// Test
func TestExpirationCheck(t *testing.T) {
    mockTime := &MockTime{
        CurrentTime: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
    }

    checker := NewExpirationChecker(mockTime)
    // ... test expiration logic
}
```

### Pattern 2: Function Variable

```go
var timeNow = time.Now

func IsExpired(expiryTime time.Time) bool {
    return timeNow().After(expiryTime)
}

// Test
func TestIsExpired(t *testing.T) {
    oldTimeNow := timeNow
    defer func() { timeNow = oldTimeNow }()

    timeNow = func() time.Time {
        return time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
    }

    if !IsExpired(time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)) {
        t.Error("Should be expired")
    }
}
```

## Mock Organization

### Project Structure

```
internal/
  adapters/
    nylas/
      client.go      # Real implementation
      mock.go        # Mock implementation
      demo.go        # Demo data for TUI
  ports/
    nylas.go         # Interface definition
```

### Mock File Convention

```go
// File: internal/adapters/nylas/mock.go
package nylas

type MockClient struct {
    // Function fields for all interface methods
    GetMessagesFunc    func(ctx context.Context, grantID string, limit int) ([]domain.Message, error)
    CreateMessageFunc  func(ctx context.Context, grantID string, req *domain.CreateMessageRequest) (*domain.Message, error)
    // ... all other methods
}

func NewMockClient() *MockClient {
    return &MockClient{
        // Sensible defaults
    }
}

// Implement interface methods
func (m *MockClient) GetMessages(ctx context.Context, grantID string, limit int) ([]domain.Message, error) {
    if m.GetMessagesFunc != nil {
        return m.GetMessagesFunc(ctx, grantID, limit)
    }
    return []domain.Message{}, nil
}

// ... implement all other methods
```

## Summary

**Key Principles:**
1. ✅ Use interfaces for dependency injection
2. ✅ Mock external dependencies (APIs, file system, network)
3. ✅ Use callback-based mocks for flexibility
4. ✅ Provide sensible defaults in mock constructors
5. ✅ Use `httptest` for HTTP testing
6. ✅ Mock time using function variables or time providers
7. ❌ Don't mock everything (value objects, pure functions)
8. ❌ Don't mock standard library (use test doubles)
9. ❌ Don't mock internal implementation details
10. ❌ Don't use global state in mocks

**Remember:** Mocks are tools to make testing easier. Use them wisely to test behavior, not implementation.
