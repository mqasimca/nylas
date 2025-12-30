# Generate CRUD Command

Auto-generate a complete CRUD CLI command package with list, show, create, update, and delete operations.

Resource: $ARGUMENTS

---

## Quick Start

1. Gather: resource name, parent command, operations needed, API endpoint, key fields
2. Use patterns from `../add-command/references/`
3. Follow `templates/crud-checklist.md` for file generation
4. Run `make ci` to verify

---

## Reference Files

| File | Purpose |
|------|---------|
| `../add-command/references/domain-patterns.md` | Domain type templates |
| `../add-command/references/adapter-patterns.md` | API implementation |
| `../add-command/references/cli-patterns.md` | CLI command structure |
| `templates/crud-checklist.md` | Complete file checklist |

---

## Generation Steps

### 1. Domain (`internal/domain/{resource}.go`)
```go
type {Resource} struct {
    ID string `json:"id"`
    // Fields from API spec
}

type Create{Resource}Request struct { /* required fields */ }
type Update{Resource}Request struct { /* optional fields */ }
type {Resource}QueryParams struct { Limit int; PageToken string }
```

### 2. Port (`internal/ports/nylas.go`)
Add: List, Get, Create, Update, Delete methods

### 3. Adapter (`internal/adapters/nylas/`)
- `{resource}s.go` - Implementation
- `mock_{resource}.go` - Mock functions
- `demo_{resource}.go` - Demo data

### 4. CLI (`internal/cli/{resource}/`)
- Root command + list, show, create, update, delete subcommands
- helpers.go with getClient(), getGrantID(), createContext()
- Tests

### 5. Register (`cmd/nylas/main.go`)
```go
rootCmd.AddCommand({resource}.New{Resource}Cmd())
```

---

## Verification

```bash
make ci-full
./bin/nylas {resource} --help
```

---

## Checklist

See `templates/crud-checklist.md` for complete file list.
