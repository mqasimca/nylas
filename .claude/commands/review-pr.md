# Review Pull Request

Review code changes following nylas CLI standards and best practices.

## Instructions

1. First, get the diff to review:
```bash
git diff main...HEAD
```

Or for a specific PR:
```bash
gh pr diff <pr-number>
```

2. Review checklist:

### Architecture
- [ ] Changes follow hexagonal architecture (domain → ports → adapters → CLI)
- [ ] No direct dependencies on concrete implementations (use interfaces)
- [ ] New code is in the correct layer/package

### Code Quality
- [ ] Functions are appropriately sized (<50 lines ideal)
- [ ] Error messages are user-friendly with suggestions
- [ ] No hardcoded credentials or secrets
- [ ] Context is passed to all API calls

### CLI Standards
- [ ] Commands follow naming conventions (newXxxCmd)
- [ ] Flags have descriptions and appropriate defaults
- [ ] Help text includes examples
- [ ] Output supports --format flag where appropriate

### Testing
- [ ] Unit tests added/updated for new functionality
- [ ] Mock implementations updated if interface changed
- [ ] Integration tests added for user-facing features
- [ ] Tests pass: `go test ./...`

### Documentation
- [ ] README.md updated if user-facing changes
- [ ] Code comments for non-obvious logic
- [ ] Examples in command help text

3. Run verification:
```bash
# Build
go build ./...

# Lint (if available)
golangci-lint run

# Tests
go test ./...

# Integration tests (if credentials available)
go test -tags=integration ./internal/cli/...
```

4. Provide feedback with:
- Specific file:line references
- Suggested fixes with code examples
- Priority (must fix, should fix, nice to have)
