# Linting Rules - Always Run Before Completion

## MANDATORY: Lint After Every Code Change

**After writing ANY Go code:**

```bash
go fmt ./...                    # Format code
golangci-lint run --timeout=5m  # Lint code
```

**Fix ALL linting issues in code you wrote/modified.**

---

## Common Linting Fixes

| Error | Fix | Example |
|-------|-----|---------|
| **errcheck** | Add `_ =` or handle error | `_ = json.Encode(data)  // Test helper` |
| **unused** | Delete unused code | Remove unused funcs/vars/imports |
| **SA5011** | Add nil check before deref | `if x == nil { return }` then use `x.Field` |
| **SA9003** | Implement or remove empty branch | Add logic or delete empty if/else |
| **SA1019** | Use non-deprecated function | `cases.Title()` instead of `strings.Title()` |
| **S1009** | Remove unnecessary nil check | `len(slice) == 0` handles nil |
| **ineffassign** | Remove/use variable | Delete unused assignments |

---

## Workflow Integration

```
Write Code → Format → Lint → Fix Issues → Test → Complete
     ↑                              |
     └──────── Back if errors ──────┘
```

**Run linting:**
- ✅ After writing new code
- ✅ After modifying existing code
- ✅ Before running tests
- ✅ Before marking task complete

---

## What to Fix vs Ignore

**MUST FIX:**
- ✅ All errors in files you created
- ✅ All errors in files you modified
- ✅ All errcheck issues in your code
- ✅ All unused code you introduced

**CAN IGNORE:**
- ⚠️ Pre-existing errors in untouched files
- ⚠️ Warnings in vendored/generated code

---

## Quick Commands

```bash
go fmt ./...                                # Format all code
golangci-lint run --timeout=5m              # Lint everything
golangci-lint run --timeout=5m --fix        # Auto-fix issues
golangci-lint run --new-from-rev=HEAD~1     # Lint only changed files
```

---

## Quality Gate: Zero Errors in New Code

**Your code changes should NEVER introduce new linting errors.**

If `golangci-lint run` shows errors in files you modified:
1. ✅ Fix them immediately
2. ✅ Don't proceed to next task
3. ✅ Don't mark current task complete
4. ✅ Treat linting errors like compilation errors

**Exception:** If 50+ pre-existing errors, focus on fixing errors in your new code only.

---

## Linting Checklist

Before marking task complete:

- [ ] Ran `go fmt ./...`
- [ ] Ran `golangci-lint run --timeout=5m`
- [ ] Fixed all errcheck issues
- [ ] Removed all unused code
- [ ] Fixed all nil pointer checks
- [ ] Verified no new linting errors
- [ ] All tests still pass
- [ ] Build still succeeds

---

## Integration with make check

```bash
make check   # Runs: lint → test → security → build
```

**Always run before completing a task.**

---

**Summary:** Linting is mandatory. Treat it like compilation - if it doesn't lint, fix it.
