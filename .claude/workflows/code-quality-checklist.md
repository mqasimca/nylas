# Code Quality Checklist

This checklist applies to **ALL** code changes, whether you're adding features, fixing bugs, or making any other modifications.

## ⚠️ MANDATORY Steps After Writing Code

**Never skip these steps. Treat them like compilation - if code doesn't lint, it's not done.**

### Step 1: Format Code

```bash
go fmt ./...
```

### Step 2: Run Linting

```bash
golangci-lint run --timeout=5m
```

### Step 3: Fix ALL Linting Issues in Your Code

**You MUST fix linting issues in any file you created or modified.**

Common issues and fixes:

#### errcheck - Unchecked Errors
```go
// ❌ Before
json.NewEncoder(w).Encode(data)
cmd.MarkFlagRequired("name")

// ✅ After
_ = json.NewEncoder(w).Encode(data)  // Test helper
_ = cmd.MarkFlagRequired("name")      // Hardcoded value
```

#### unused - Unused Code
```go
// ❌ Before
func helperFunc() {}  // Never called
import "os"           // Never used

// ✅ After
// Delete unused functions and imports
```

#### staticcheck SA5011 - Nil Pointer
```go
// ❌ Before
flag := cmd.Flags().Lookup("name")
if flag.DefValue != "x" {  // Can panic!

// ✅ After
flag := cmd.Flags().Lookup("name")
if flag == nil {
    t.Error("flag not found")
    return  // Early return
}
if flag.DefValue != "x" {
```

**See `.claude/rules/linting.md` for comprehensive guide.**

### Step 4: Run Tests

```bash
# Run all unit tests
go test ./... -short

# Or run specific package tests
go test ./internal/cli/email/... -v
```

### Step 5: Verify Build

```bash
go build -o ./bin/nylas ./cmd/nylas
```

### Step 6: Run Full Verification (Optional but Recommended)

```bash
make check
```

This runs: lint → test → security → build

## 🚫 What NOT to Do

- ❌ Skip linting ("I'll fix it later")
- ❌ Proceed with failing lints in your code
- ❌ Ask user "Should I fix the linting errors?" (just fix them)
- ❌ Fix pre-existing linting issues in files you didn't touch (out of scope)
- ❌ Leave unused code "in case we need it later"
- ❌ Ignore errcheck warnings

## ✅ What TO Do

- ✅ Run linting after every code change
- ✅ Fix ALL linting issues in code you wrote
- ✅ Treat linting like compilation (must pass)
- ✅ Use `_ =` for explicitly ignored errors with comment
- ✅ Remove unused code immediately
- ✅ Add nil checks in tests before dereferencing

## Quality Gate

**Your code is NOT complete until:**

1. ✅ Code compiles
2. ✅ Linting passes (no errors in your code)
3. ✅ Tests pass
4. ✅ Build succeeds

All four are required. Linting is not optional.

## When You See Linting Errors

### Scenario 1: Errors in Files You Modified

**Action:** Fix immediately. Don't proceed until fixed.

```bash
# Example output
internal/cli/email/send.go:42: Error return value not checked (errcheck)

# Fix before continuing
```

### Scenario 2: Pre-existing Errors in Files You Didn't Touch

**Action:** Ignore them (out of scope). Focus on your code.

```bash
# Example output
internal/tui/app.go:123: unused function (unused)  # You didn't touch this

# Ignore - not your responsibility
```

### Scenario 3: Overwhelming Pre-existing Issues (50+)

**Action:**
1. Fix errors in your new code only
2. Note the pre-existing issues
3. Suggest a cleanup task to user
4. Don't let pre-existing issues block your work

## Integration with Workflows

### For Bug Fixes

```
1. Understand bug
2. Locate code
3. Write failing test
4. Fix bug
5. Lint and fix issues ← MANDATORY
6. Verify tests pass
7. Done
```

### For New Features

```
1. Plan feature
2. Implement code
3. Write tests
4. Lint and fix issues ← MANDATORY
5. Update docs
6. Done
```

### For Refactoring

```
1. Understand current code
2. Plan refactoring
3. Make changes
4. Lint and fix issues ← MANDATORY
5. Verify tests still pass
6. Done
```

## Examples from This Codebase

### Example 1: After Adding Metadata Feature

```bash
# 1. Implemented feature
# 2. Ran linting
golangci-lint run --timeout=5m

# 3. Found issues:
#    - Unused imports in metadata.go
#    - Unchecked errors in test files
#    - Nil pointer issues in tests

# 4. Fixed all issues in new code
#    - Removed unused imports
#    - Added _ = for test errors
#    - Added nil checks in tests

# 5. Re-ran linting - clean!
# 6. Tests passed
# 7. Feature complete
```

### Example 2: Quick Bug Fix

```bash
# 1. Fixed nil pointer bug
# 2. Ran linting
golangci-lint run --timeout=5m

# 3. Found: unused variable from old debugging code
# 4. Removed unused variable
# 5. Re-ran linting - clean!
# 6. Done
```

## Checklist Template

Copy this for every code change:

```
Before marking task complete:

[ ] Ran go fmt ./...
[ ] Ran golangci-lint run --timeout=5m
[ ] Fixed all errcheck issues (added _ = or proper handling)
[ ] Removed all unused code (functions, vars, constants, imports)
[ ] Fixed all staticcheck issues (nil checks, deprecated usage)
[ ] Verified no new linting errors in my code
[ ] Ran go test ./... -short
[ ] Tests pass
[ ] Ran go build -o ./bin/nylas ./cmd/nylas
[ ] Build succeeds
[ ] Ready to mark complete
```

## Quick Reference

| Issue Type | Quick Fix |
|------------|-----------|
| errcheck | Add `_ =` with comment |
| unused function | Delete it |
| unused import | Remove the import line |
| unused variable | Delete it |
| SA5011 nil pointer | Add nil check + early return |
| SA1019 deprecated | Use modern alternative |
| S1009 redundant nil | Remove nil check |

## Summary

**Linting is mandatory. It's not optional. It's part of "done".**

Write → Format → Lint → Fix → Test → Done
          ↑                      |
          └────── If errors ─────┘

Make this your automatic workflow. Never skip linting.
