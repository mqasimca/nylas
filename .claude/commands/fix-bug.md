# Fix Bug

Systematically fix a bug in the Nylas CLI.

Bug description: $ARGUMENTS

## Instructions

1. **Understand the bug**
   - Identify the expected behavior
   - Identify the actual behavior
   - Determine reproduction steps

2. **Locate the relevant code**
   - Search for related keywords using Grep
   - Check these locations based on bug type:
     - CLI behavior: `internal/cli/<feature>/`
     - API issues: `internal/adapters/nylas/`
     - Data issues: `internal/domain/`
     - Auth issues: `internal/adapters/keyring/`, `internal/cli/auth/`

3. **Write a failing test** (if possible)
   - Add test to appropriate `*_test.go` file
   - Test should fail with current code
   - Test should pass after fix

4. **Fix the bug**
   - Make minimal changes needed
   - Follow existing code patterns
   - Don't introduce new dependencies unless necessary

5. **Verify the fix**
   - Run the specific test: `go test ./path/to/package -run TestName -v`
   - Run full unit test suite: `go test ./... -short`
   - Run integration tests (if API-related): `go test ./... -tags=integration`
   - Manual testing if needed: `./bin/nylas <command>`

6. **Check for side effects**
   - Run all unit tests: `make test`
   - Run integration tests: `go test ./... -tags=integration`
   - Build succeeds: `make build`
   - Lint passes: `make lint` (if available)

7. **Document if needed**
   - Update docs if behavior changed
   - Add code comments if fix is non-obvious

## Common Bug Locations

| Bug Type | Check These Files |
|----------|-------------------|
| Wrong API response | `internal/adapters/nylas/*.go` |
| Missing field | `internal/domain/*.go` |
| Auth failure | `internal/adapters/keyring/`, `internal/cli/auth/helpers.go` |
| CLI flag issue | `internal/cli/<feature>/*.go` |
| Display issue | `internal/cli/<feature>/list.go`, `show.go` |
| TUI issue | `internal/tui/` |

## Debugging Tips

```bash
# Build and test locally
make build && ./bin/nylas <command> --verbose

# Run specific test with verbose output
go test ./internal/cli/email/... -run TestSendCommand -v

# Check API response
curl -H "Authorization: Bearer $NYLAS_API_KEY" https://api.us.nylas.com/v3/grants/$GRANT_ID/<endpoint>
```

## Checklist
- [ ] Bug understood and reproducible
- [ ] Relevant code located
- [ ] **Failing test written** that reproduces the bug
- [ ] Bug fixed
- [ ] **Test now passes**
- [ ] **All unit tests pass**: `go test ./... -short`
- [ ] **Integration tests pass** (if API-related): `go test ./... -tags=integration`
- [ ] **Linting passes**: `golangci-lint run`
- [ ] **Security scan passes**: `make security`
- [ ] **Documentation updated** (if behavior changes):
  - [ ] `docs/COMMANDS.md` - If command behavior changed
  - [ ] `README.md` - If user-visible behavior changed
- [ ] Build succeeds: `make build`
- [ ] Full check passes: `make check`
- [ ] No regressions introduced

## ⛔ MANDATORY - Before Committing:
```bash
# Run full verification
make check

# Verify no secrets
git diff --cached | grep -iE "(api_key|password|secret|token|nyk_v0)" && echo "STOP!" || echo "✓ OK"

# ⛔ NEVER run git push - only local commits
```
