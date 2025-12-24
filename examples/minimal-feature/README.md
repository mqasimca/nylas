# Minimal Feature Example

A complete minimal example demonstrating the hexagonal architecture pattern used in the Nylas CLI.

This example implements a simple "Widget" feature that demonstrates:
- Domain models
- Port definitions (interfaces)
- Adapter implementations
- CLI commands
- Testing patterns

---

## File Structure

```
minimal-feature/
├── README.md              # This file
├── domain/
│   └── widget.go          # Domain model
├── ports/
│   └── widget_port.go     # Interface contract
├── adapters/
│   ├── widget_adapter.go  # Implementation
│   └── mock_widget.go     # Mock for testing
└── cli/
    ├── widget.go          # Root command
    ├── list.go            # List subcommand
    ├── create.go          # Create subcommand
    └── helpers.go         # Shared helpers
```

---

## Layer Responsibilities

### 1. Domain (domain/widget.go)

**Purpose:** Define business entities and logic

**Characteristics:**
- No external dependencies
- Pure Go types
- Business rules and validations
- Serialization tags

```go
// Domain model - what IS a Widget?
type Widget struct {
    ID          string
    Name        string
    Description string
    CreatedAt   time.Time
}
```

---

### 2. Ports (ports/widget_port.go)

**Purpose:** Define interface contracts

**Characteristics:**
- Interface definitions only
- No implementation
- Context-first parameters
- Clear method signatures

```go
// Port - what CAN you do with Widgets?
type WidgetService interface {
    ListWidgets(ctx context.Context) ([]*domain.Widget, error)
    CreateWidget(ctx context.Context, widget *domain.Widget) (*domain.Widget, error)
    GetWidget(ctx context.Context, id string) (*domain.Widget, error)
}
```

---

### 3. Adapters (adapters/widget_adapter.go)

**Purpose:** Implement port interfaces

**Characteristics:**
- Implements port interfaces
- Handles external communication (API, DB, etc.)
- Error handling
- Data transformation

```go
// Adapter - HOW do you actually do it?
type WidgetAdapter struct {
    client *http.Client
    apiURL string
}

func (a *WidgetAdapter) ListWidgets(ctx context.Context) ([]*domain.Widget, error) {
    // Implementation details
}
```

---

### 4. CLI (cli/widget.go)

**Purpose:** User interface layer

**Characteristics:**
- Cobra command definitions
- Flag parsing
- Output formatting
- User interaction

```go
// CLI - what does the USER type?
func NewWidgetCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "widget",
        Short: "Manage widgets",
    }

    cmd.AddCommand(newListCmd())
    cmd.AddCommand(newCreateCmd())

    return cmd
}
```

---

## Data Flow Example

### User runs: `nylas widget list`

```
┌─────────────┐
│ CLI Layer   │  widget list command
│             │  - Parse flags
│             │  - Get service instance
└─────┬───────┘
      │
      ▼
┌─────────────┐
│ Port Layer  │  WidgetService.ListWidgets()
│             │  - Interface contract
└─────┬───────┘
      │
      ▼
┌─────────────┐
│ Adapter     │  HTTP call to API
│ Layer       │  - GET /widgets
│             │  - Parse response
└─────┬───────┘
      │
      ▼
┌─────────────┐
│ Domain      │  []*Widget
│ Layer       │  - Pure business objects
└─────────────┘
```

---

## Testing Strategy

### Unit Tests
```go
// Test domain logic
func TestWidget_Validate(t *testing.T) {
    // Test business rules
}
```

### Integration Tests
```go
// Test with mock adapter
func TestListCommand(t *testing.T) {
    mockAdapter := &MockWidgetAdapter{
        ListFunc: func(ctx) ([]*Widget, error) {
            return mockWidgets, nil
        },
    }
    // Test command with mock
}
```

---

## How to Add This to Your Project

### Step 1: Domain Model
1. Create `internal/domain/widget.go`
2. Define your business entity
3. Add validation methods if needed

### Step 2: Port Definition
1. Add methods to `internal/ports/nylas.go`:
   ```go
   ListWidgets(ctx context.Context) ([]*domain.Widget, error)
   CreateWidget(ctx context.Context, widget *domain.Widget) (*domain.Widget, error)
   ```

### Step 3: Adapter Implementation
1. Create `internal/adapters/nylas/widgets.go`
2. Implement the port methods
3. Handle API communication

### Step 4: CLI Commands
1. Create `internal/cli/widget/`
2. Add `widget.go` (root command)
3. Add `list.go`, `create.go` (subcommands)
4. Add `helpers.go` (shared utilities)

### Step 5: Register Command
1. In `cmd/nylas/main.go`:
   ```go
   rootCmd.AddCommand(widget.NewWidgetCmd())
   ```

### Step 6: Tests
1. Create `internal/cli/widget/widget_test.go` (unit tests)
2. Create `internal/cli/integration/widget_test.go` (integration tests)

### Step 7: Documentation
1. Update `docs/COMMANDS.md`
2. Update `CLAUDE.md` file structure table
3. Add examples to `docs/examples/`

---

## Key Principles

1. **Dependency Direction:** Always point inward
   - CLI → Ports → Domain
   - Adapters → Ports → Domain
   - Never: Domain → Anything

2. **Single Responsibility:**
   - Domain: Business logic
   - Ports: Contracts
   - Adapters: External communication
   - CLI: User interaction

3. **Testability:**
   - Mock adapters for CLI tests
   - Pure functions in domain
   - Interface-based dependencies

4. **Separation of Concerns:**
   - Don't mix layers
   - Each layer has one job
   - Clear boundaries

---

## Benefits of This Pattern

✅ **Testable:** Mock external dependencies easily
✅ **Maintainable:** Change one layer without affecting others
✅ **Clear:** Each file has a single, obvious purpose
✅ **Scalable:** Add features following the same pattern
✅ **AI-Friendly:** Predictable structure makes AI assistance easier

---

## Common Mistakes to Avoid

❌ **Don't:** Put business logic in CLI commands
✅ **Do:** Keep CLI layer thin, delegate to domain/adapter

❌ **Don't:** Import adapters in domain
✅ **Do:** Use ports (interfaces) in domain

❌ **Don't:** Mix concerns (e.g., HTTP calls in CLI)
✅ **Do:** Separate external communication (adapters) from UI (CLI)

❌ **Don't:** Make large files (>500 lines)
✅ **Do:** Split by operation (list.go, create.go, etc.)

---

## Related Documentation

- [ARCHITECTURE.md](../../docs/ARCHITECTURE.md) - Full architecture guide
- [add-feature skill](../../.claude/commands/add-feature.md) - Automated workflow
- [CLAUDE.md](../../CLAUDE.md) - Quick reference

---

**Last Updated:** December 23, 2024
