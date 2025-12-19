# Generate CRUD Command

Auto-generate a complete CRUD CLI command package with list, show, create, update, and delete operations.

Resource: $ARGUMENTS

## Instructions

1. **Gather requirements**
   - Resource name (singular, e.g., "widget")
   - Parent command if nested (e.g., "email" for "email drafts")
   - Which operations: list, show, create, update, delete
   - API endpoint path (e.g., "/grants/{grantID}/widgets")
   - Key fields for the resource

2. **Generate domain model** (`internal/domain/{resource}.go`):

```go
package domain

type {Resource} struct {
    ID        string `json:"id"`
    // Add fields based on API spec
    CreatedAt int64  `json:"created_at,omitempty"`
    UpdatedAt int64  `json:"updated_at,omitempty"`
}

type Create{Resource}Request struct {
    // Required fields for creation
}

type Update{Resource}Request struct {
    // Optional fields for update
}

type {Resource}QueryParams struct {
    Limit      int    `url:"limit,omitempty"`
    PageToken  string `url:"page_token,omitempty"`
}
```

3. **Add port interface methods** (`internal/ports/nylas.go`):

```go
// {Resource} operations
List{Resource}s(ctx context.Context, grantID string, params *domain.{Resource}QueryParams) ([]domain.{Resource}, string, error)
Get{Resource}(ctx context.Context, grantID, {resource}ID string) (*domain.{Resource}, error)
Create{Resource}(ctx context.Context, grantID string, req *domain.Create{Resource}Request) (*domain.{Resource}, error)
Update{Resource}(ctx context.Context, grantID, {resource}ID string, req *domain.Update{Resource}Request) (*domain.{Resource}, error)
Delete{Resource}(ctx context.Context, grantID, {resource}ID string) error
```

4. **Implement adapter** (`internal/adapters/nylas/{resource}s.go`):

```go
package nylas

import (
    "context"
    "fmt"
    "github.com/mqasimca/nylas/internal/domain"
)

func (c *HTTPClient) List{Resource}s(ctx context.Context, grantID string, params *domain.{Resource}QueryParams) ([]domain.{Resource}, string, error) {
    var resp struct {
        Data      []domain.{Resource} `json:"data"`
        NextCursor string             `json:"next_cursor,omitempty"`
    }
    path := fmt.Sprintf("/grants/%s/{resource}s", grantID)
    if err := c.get(ctx, path, &resp, params); err != nil {
        return nil, "", err
    }
    return resp.Data, resp.NextCursor, nil
}

func (c *HTTPClient) Get{Resource}(ctx context.Context, grantID, {resource}ID string) (*domain.{Resource}, error) {
    var resp struct {
        Data domain.{Resource} `json:"data"`
    }
    path := fmt.Sprintf("/grants/%s/{resource}s/%s", grantID, {resource}ID)
    if err := c.get(ctx, path, &resp, nil); err != nil {
        return nil, err
    }
    return &resp.Data, nil
}

func (c *HTTPClient) Create{Resource}(ctx context.Context, grantID string, req *domain.Create{Resource}Request) (*domain.{Resource}, error) {
    var resp struct {
        Data domain.{Resource} `json:"data"`
    }
    path := fmt.Sprintf("/grants/%s/{resource}s", grantID)
    if err := c.post(ctx, path, req, &resp); err != nil {
        return nil, err
    }
    return &resp.Data, nil
}

func (c *HTTPClient) Update{Resource}(ctx context.Context, grantID, {resource}ID string, req *domain.Update{Resource}Request) (*domain.{Resource}, error) {
    var resp struct {
        Data domain.{Resource} `json:"data"`
    }
    path := fmt.Sprintf("/grants/%s/{resource}s/%s", grantID, {resource}ID)
    if err := c.put(ctx, path, req, &resp); err != nil {
        return nil, err
    }
    return &resp.Data, nil
}

func (c *HTTPClient) Delete{Resource}(ctx context.Context, grantID, {resource}ID string) error {
    path := fmt.Sprintf("/grants/%s/{resource}s/%s", grantID, {resource}ID)
    return c.delete(ctx, path)
}
```

5. **Add mock implementation** (`internal/adapters/nylas/mock.go`):

```go
func (m *MockClient) List{Resource}s(ctx context.Context, grantID string, params *domain.{Resource}QueryParams) ([]domain.{Resource}, string, error) {
    return []domain.{Resource}{{ID: "mock-{resource}-1"}}, "", nil
}

func (m *MockClient) Get{Resource}(ctx context.Context, grantID, {resource}ID string) (*domain.{Resource}, error) {
    return &domain.{Resource}{ID: {resource}ID}, nil
}

func (m *MockClient) Create{Resource}(ctx context.Context, grantID string, req *domain.Create{Resource}Request) (*domain.{Resource}, error) {
    return &domain.{Resource}{ID: "new-{resource}-id"}, nil
}

func (m *MockClient) Update{Resource}(ctx context.Context, grantID, {resource}ID string, req *domain.Update{Resource}Request) (*domain.{Resource}, error) {
    return &domain.{Resource}{ID: {resource}ID}, nil
}

func (m *MockClient) Delete{Resource}(ctx context.Context, grantID, {resource}ID string) error {
    return nil
}
```

6. **Add demo implementation** (`internal/adapters/nylas/demo.go`):

```go
func (d *DemoClient) List{Resource}s(ctx context.Context, grantID string, params *domain.{Resource}QueryParams) ([]domain.{Resource}, string, error) {
    return []domain.{Resource}{
        {ID: "demo-{resource}-1", /* realistic fields */},
        {ID: "demo-{resource}-2", /* realistic fields */},
    }, "", nil
}
// ... implement other methods with realistic demo data
```

7. **Create CLI package** (`internal/cli/{resource}/`):

**{resource}.go** - Root command:
```go
package {resource}

import "github.com/spf13/cobra"

func New{Resource}Cmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "{resource}",
        Short: "Manage {resource}s",
        Long:  "List, view, create, update, and delete {resource}s.",
    }
    cmd.AddCommand(newListCmd())
    cmd.AddCommand(newShowCmd())
    cmd.AddCommand(newCreateCmd())
    cmd.AddCommand(newUpdateCmd())
    cmd.AddCommand(newDeleteCmd())
    return cmd
}
```

**list.go**, **show.go**, **create.go**, **update.go**, **delete.go** - Follow existing patterns in `internal/cli/` for each subcommand.

**helpers.go** - Standard helpers:
```go
package {resource}

import (
    "context"
    "time"
    // ... imports
)

func getClient() (ports.NylasClient, error) {
    // Use existing pattern from other CLI packages
}

func getGrantID(args []string) (string, error) {
    // Use existing pattern
}

func createContext() (context.Context, context.CancelFunc) {
    return context.WithTimeout(context.Background(), 30*time.Second)
}
```

8. **Register command** (`cmd/nylas/main.go`):

```go
import "{resource}pkg" "github.com/mqasimca/nylas/internal/cli/{resource}"

rootCmd.AddCommand({resource}pkg.New{Resource}Cmd())
```

9. **Add tests** (`internal/cli/{resource}/{resource}_test.go`):

```go
package {resource}

import "testing"

func TestNew{Resource}Cmd(t *testing.T) {
    cmd := New{Resource}Cmd()
    if cmd.Use != "{resource}" {
        t.Errorf("expected Use to be '{resource}', got %s", cmd.Use)
    }
    // Test subcommands exist
    subcommands := []string{"list", "show", "create", "update", "delete"}
    for _, name := range subcommands {
        found := false
        for _, sub := range cmd.Commands() {
            if sub.Name() == name {
                found = true
                break
            }
        }
        if !found {
            t.Errorf("expected subcommand %s not found", name)
        }
    }
}
```

## Generated Files Checklist

- [ ] `internal/domain/{resource}.go` - Domain types
- [ ] `internal/ports/nylas.go` - Interface methods added
- [ ] `internal/adapters/nylas/{resource}s.go` - API implementation
- [ ] `internal/adapters/nylas/mock.go` - Mock methods added
- [ ] `internal/adapters/nylas/demo.go` - Demo methods added
- [ ] `internal/cli/{resource}/{resource}.go` - Root command
- [ ] `internal/cli/{resource}/list.go` - List subcommand
- [ ] `internal/cli/{resource}/show.go` - Show subcommand
- [ ] `internal/cli/{resource}/create.go` - Create subcommand
- [ ] `internal/cli/{resource}/update.go` - Update subcommand
- [ ] `internal/cli/{resource}/delete.go` - Delete subcommand
- [ ] `internal/cli/{resource}/helpers.go` - Helper functions
- [ ] `internal/cli/{resource}/{resource}_test.go` - Unit tests
- [ ] `cmd/nylas/main.go` - Command registered

## Verification

```bash
# Build and test
go build ./...
go test ./internal/cli/{resource}/... -v
go test ./... -short

# Verify command works
./bin/nylas {resource} --help
./bin/nylas {resource} list --help
```

## ⛔ MANDATORY - Before Committing:
```bash
make check
git diff --cached | grep -iE "(api_key|password|secret|token|nyk_v0)" && echo "STOP!" || echo "✓ OK"
# ⛔ NEVER run git push
```
