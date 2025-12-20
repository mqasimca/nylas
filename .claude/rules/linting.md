# Linting Rules - Always Run Before Completion

## MANDATORY: Lint After Every Code Change

**After writing ANY Go code, you MUST:**

1. **Format the code:**
   ```bash
   go fmt ./...
   ```

2. **Run the linter:**
   ```bash
   golangci-lint run --timeout=5m
   ```

3. **Fix ALL linting issues in code you wrote/modified**
   - Do NOT leave linting errors for later
   - Do NOT ignore linting warnings in your new code
   - Pre-existing issues in untouched files can be ignored

## Common Linting Fixes

### 1. errcheck - Unchecked Errors

**Problem:** Function returns error but it's not checked

**Fix:**
```go
// ❌ BEFORE (BAD)
json.NewEncoder(w).Encode(data)
cmd.MarkFlagRequired("name")
fmt.Scanln(&input)
w.Write([]byte("data"))

// ✅ AFTER (GOOD - Explicit ignore)
_ = json.NewEncoder(w).Encode(data)  // Test helper, error not actionable
_ = cmd.MarkFlagRequired("name")      // Hardcoded flag name, won't fail
_, _ = fmt.Scanln(&input)             // User input, validation handled later
_, _ = w.Write([]byte("data"))        // Test response writer, error not relevant

// ✅ AFTER (GOOD - Proper handling)
if err := json.NewEncoder(w).Encode(data); err != nil {
    return fmt.Errorf("encode response: %w", err)
}
```

**When to use `_ =` vs proper error handling:**
- **Use `_ =`** in tests, when error truly cannot occur, or when error is not actionable
- **Use proper handling** in production code, when error affects program behavior

### 2. unused - Unused Code

**Problem:** Code defined but never used

**Fix:** Delete it completely
```go
// ❌ BEFORE
func helperFunction() string {  // Never called
    return "unused"
}

const maxSize = 1024  // Never referenced

var tempData []string  // Never used

import "os"  // Never used

// ✅ AFTER
// All removed - clean code!
```

**What to remove:**
- ✅ Unused functions
- ✅ Unused variables
- ✅ Unused constants
- ✅ Unused imports
- ✅ Unused struct fields
- ✅ Unused type definitions

### 3. staticcheck SA5011 - Nil Pointer Dereference

**Problem:** Possible nil pointer dereference in tests

**Fix:**
```go
// ❌ BEFORE (BAD - Can panic)
flag := cmd.Flags().Lookup("name")
if flag.DefValue != "expected" {  // Panic if flag is nil!
    t.Error("wrong default")
}

// ✅ AFTER (GOOD - Safe)
flag := cmd.Flags().Lookup("name")
if flag == nil {
    t.Error("flag not found")
    return  // Early return prevents nil dereference
}
if flag.DefValue != "expected" {
    t.Error("wrong default")
}
```

### 4. staticcheck SA9003 - Empty Branch

**Problem:** Empty if/else branch

**Fix:** Either implement the branch or remove it
```go
// ❌ BEFORE
if err := server.Serve(); err != http.ErrServerClosed {
    // Empty - what should happen here?
}

// ✅ AFTER (Option 1 - Handle error)
if err := server.Serve(); err != http.ErrServerClosed {
    log.Printf("Server error: %v", err)
}

// ✅ AFTER (Option 2 - Remove if not needed)
_ = server.Serve()  // Error handled elsewhere
```

### 5. staticcheck SA1019 - Deprecated Function

**Problem:** Using deprecated stdlib function

**Fix:**
```go
// ❌ BEFORE (Deprecated since Go 1.18)
title := strings.Title(name)

// ✅ AFTER (Use cases package)
import "golang.org/x/text/cases"
import "golang.org/x/text/language"

caser := cases.Title(language.English)
title := caser.String(name)
```

### 6. gosimple S1009 - Unnecessary Nil Check

**Problem:** Checking if slice is nil before checking length

**Fix:**
```go
// ❌ BEFORE (Redundant)
if slice == nil || len(slice) == 0 {
    return
}

// ✅ AFTER (len() handles nil)
if len(slice) == 0 {
    return
}
```

### 7. ineffassign - Ineffectual Assignment

**Problem:** Variable assigned but never used

**Fix:**
```go
// ❌ BEFORE
status := "unknown"
if config != nil {
    status = config.Status  // Original value never used
}

// ✅ AFTER
status := "unknown"
if config != nil {
    status = config.Status
}
// OR just initialize with the final value
var status string
if config != nil {
    status = config.Status
} else {
    status = "unknown"
}
```

## Workflow Integration

### When to Run Linting

```
Write Code → Format → Lint → Fix Issues → Test → Complete
     ↑                              |
     └──────── Back if errors ──────┘
```

Run linting:
- ✅ **After** writing new code
- ✅ **After** modifying existing code
- ✅ **Before** running tests
- ✅ **Before** marking task as complete
- ✅ **Before** every commit (automated via make check)

### What to Fix vs What to Ignore

**MUST FIX:**
- ✅ All linting errors in files you created
- ✅ All linting errors in files you modified
- ✅ All errcheck issues in your code
- ✅ All unused code you introduced

**CAN IGNORE:**
- ⚠️ Pre-existing errors in files you didn't touch
- ⚠️ Warnings in vendored/generated code
- ⚠️ Style preferences that don't affect functionality (if pre-existing)

### Quick Linting Commands

```bash
# Format all code
go fmt ./...

# Lint everything
golangci-lint run --timeout=5m

# Lint only changed files (faster)
golangci-lint run --timeout=5m --new-from-rev=HEAD~1

# Show only errors (not warnings)
golangci-lint run --timeout=5m --issues-exit-code=1

# Fix auto-fixable issues
golangci-lint run --fix
```

## Quality Gate: Zero Linting Errors in New Code

### The Rule

**Your code changes should NEVER introduce new linting errors.**

### What This Means

If `golangci-lint run` shows errors in files you modified:
1. ✅ Fix them immediately (don't wait)
2. ✅ Don't proceed to next task
3. ✅ Don't mark current task as complete
4. ✅ Don't ask user if you should fix them (just fix)
5. ✅ Treat linting errors like compilation errors

### Exception

If there are **overwhelming pre-existing** linting issues (50+ errors):
1. Focus on fixing errors in your new code only
2. Note the pre-existing issues
3. Suggest a separate cleanup task to the user
4. Don't let pre-existing issues block your progress

## Linting Checklist

Before marking any task complete:

- [ ] Ran `go fmt ./...`
- [ ] Ran `golangci-lint run --timeout=5m`
- [ ] Fixed all errcheck issues (added `_ =` or proper handling)
- [ ] Removed all unused code (functions, vars, constants, imports)
- [ ] Fixed all nil pointer checks in tests
- [ ] Verified no new linting errors introduced
- [ ] All tests still pass after linting fixes
- [ ] Build still succeeds after linting fixes

## Integration with make check

The `make check` command runs:
1. `golangci-lint run` - Linting
2. `go test ./...` - Tests
3. `make security` - Security scan
4. `make build` - Build verification

**Always run `make check` before considering a task complete.**

## Examples from This Codebase

### Example 1: Test Helper Error

```go
// ❌ Found in internal/adapters/nylas/admin_test.go
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(response)

// ✅ Fixed
w.Header().Set("Content-Type", "application/json")
_ = json.NewEncoder(w).Encode(response)  // Test helper, encode error not relevant
```

### Example 2: Unused Test Helper

```go
// ❌ Found in internal/cli/admin/admin_test.go
func executeCommand(root *cobra.Command, args ...string) (string, string, error) {
    // ... implementation never called
}

// ✅ Fixed - Removed entirely
```

### Example 3: Nil Pointer in Test

```go
// ❌ Found in internal/cli/otp/otp_test.go
flag := cmd.Flags().Lookup("interval")
if flag == nil {
    t.Error("Expected --interval flag")
}
if flag.DefValue != "10" {  // Could panic if flag is nil!

// ✅ Fixed
flag := cmd.Flags().Lookup("interval")
if flag == nil {
    t.Error("Expected --interval flag")
    return  // Early return prevents nil dereference
}
if flag.DefValue != "10" {
    t.Errorf("--interval default = %q, want %q", flag.DefValue, "10")
}
```

### Example 4: Unused Import

```go
// ❌ Found in internal/cli/email/helpers.go
import (
    "context"
    "fmt"
    "html"
    "os"  // Not used anywhere
    "strings"
)

// ✅ Fixed - Removed unused import
import (
    "context"
    "fmt"
    "html"
    "strings"
)
```

## Summary

**Linting is not optional. It's a mandatory quality check.**

Treat linting like compilation:
- If code doesn't compile → fix it
- If code doesn't lint → fix it
- If tests don't pass → fix it

All three are required for code to be considered "done".
