# Development

Build and test the Nylas CLI.

> **Quick Links:** [README](../README.md) | [Commands](COMMANDS.md) | [Architecture](ARCHITECTURE.md)

---

## Prerequisites

- Go 1.21+
- Make (optional)

---

## Build

```bash
make build                   # Build binary
make clean                   # Clean artifacts
```

---

## Test

```bash
make test                    # Run all tests
make lint                    # Run linter
make check                   # Run lint + test + security + build
```

---

## Integration Tests

```bash
export NYLAS_API_KEY="your-api-key"
export NYLAS_GRANT_ID="your-grant-id"

make test-integration
```

---

## Project Structure

```
cmd/nylas/main.go           # Entry point
internal/
  ├── domain/               # Domain models
  ├── ports/                # Interfaces
  ├── adapters/             # Implementations
  └── cli/                  # Commands
```

---

## Detailed Guides

For contributors, comprehensive guides are available:

- **[Adding Commands](development/adding-command.md)** - Step-by-step guide for new CLI commands
- **[Adding Adapters](development/adding-adapter.md)** - Implementing API adapters
- **[Testing Guide](development/testing-guide.md)** - Unit and integration testing
- **[Debugging](development/debugging.md)** - Debugging tips and techniques

---

**Quick reference:** See `CLAUDE.md` for project overview and AI assistant guidelines.
