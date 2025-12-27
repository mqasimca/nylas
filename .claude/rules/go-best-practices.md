# Go Best Practices Rules

Auto-applied to all Go code changes in this repository.

---

## 🔒 MANDATORY WORKFLOW

### Before Writing Go Code:

1. **Check Go version:** `go version && grep "^go " go.mod` (Currently: **Go 1.24.0**)
2. **Research official docs:** Use WebSearch for `go.dev/ref/spec`, `pkg.go.dev`
3. **Verify standard library first** before writing custom code

---

## Modern Go Patterns (Go 1.24+)

| Instead of | Use | Since |
|------------|-----|-------|
| `io/ioutil.ReadFile` | `os.ReadFile` | Go 1.16+ |
| `interface{}` | `any` | Go 1.18+ |
| Manual slice helpers | `slices` package | Go 1.21+ |
| Manual map helpers | `maps` package | Go 1.21+ |
| Recreate to clear | `clear()` | Go 1.21+ |
| Custom min/max | `min()`, `max()` | Go 1.21+ |
| Manual comparison | `cmp.Compare()` | Go 1.21+ |
| `sort.Slice` | `slices.SortFunc` | Go 1.21+ |

**Examples at:** `pkg.go.dev/slices`, `pkg.go.dev/maps`, `pkg.go.dev/cmp`

---

## Error Handling (CRITICAL)

**Every error MUST be handled or explicitly ignored.**

```go
// ✅ Explicit ignore (with comment)
_ = json.Encode(data)  // Test helper, error not actionable

// ✅ Proper handling
if err := json.Encode(data); err != nil {
    return fmt.Errorf("encode failed: %w", err)
}

// ✅ Error checking
if errors.Is(err, os.ErrNotExist) { /* handle */ }

// ✅ Nil check before dereference (especially in tests)
if obj == nil {
    t.Error("object is nil")
    return
}
// Now safe to use obj.Field
```

**When to use `_ =`:**
- Test code where error is irrelevant
- Error cannot occur (hardcoded values)
- Best-effort operations

**When to handle:**
- Production code
- Error affects behavior
- User needs to know

---

## Forbidden Patterns

**Do NOT use:**
- ❌ `io/ioutil` - Deprecated
- ❌ `interface{}` - Use `any`
- ❌ Custom slice/map helpers - Use stdlib
- ❌ Manual min/max - Use built-ins
- ❌ `math/rand` for new code - Use `crypto/rand` or `math/rand/v2`

---

## Quality Checks (REQUIRED)

```bash
make ci        # Runs: fmt → vet → lint → test-unit → test-race → security → vuln → build
make ci-full   # Complete CI: all quality checks + integration tests + cleanup
```

**Never skip linting.** Fix ALL issues in code you wrote.

---

## Code Conventions

Follow [Effective Go](https://go.dev/doc/effective_go) and [Code Review Comments](https://go.dev/wiki/CodeReviewComments):

- Use `MixedCaps` for names (not underscores)
- Short names in short scopes (`i`, `r`, `w`)
- Descriptive names for package-level
- Comment exported identifiers
- Handle errors explicitly

---

## Resources

- **Go Spec:** https://go.dev/ref/spec
- **Stdlib:** https://pkg.go.dev/std
- **Effective Go:** https://go.dev/doc/effective_go
- **Release Notes:** https://go.dev/doc/devel/release
- **Uber Guide:** https://github.com/uber-go/guide/blob/master/style.md

---

**Remember:** Always check official docs first. If unsure, search `golang [feature] best practice`.
