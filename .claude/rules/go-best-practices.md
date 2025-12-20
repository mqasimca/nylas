# Go Best Practices Rules

These rules are automatically applied to all Go code changes in this repository.

---

## 🔒 MANDATORY RULES - MUST FOLLOW

### Rule 1: Research Before Coding

**BEFORE making ANY Go code changes, you MUST:**

1. Check the current Go version:
   ```bash
   go version
   grep "^go " go.mod
   ```

2. Search for official Go documentation:
   - Use WebSearch to check https://go.dev/ref/spec
   - Check https://pkg.go.dev for standard library
   - Review https://go.dev/doc/devel/release for version features

3. Verify modern alternatives exist:
   - Search: "golang [version] [feature] best practice"
   - Search: "golang [package] official documentation"
   - Check if standard library has the solution

### Rule 2: Use Modern Go Features

Based on Go version in `go.mod`, you MUST use modern features:

| Feature | Minimum Version | Required |
|---------|----------------|----------|
| `os.ReadFile` instead of `ioutil.ReadFile` | Go 1.16+ | ✅ REQUIRED |
| `any` instead of `interface{}` | Go 1.18+ | ✅ REQUIRED |
| `slices` package for slice operations | Go 1.21+ | ✅ REQUIRED |
| `maps` package for map operations | Go 1.21+ | ✅ REQUIRED |
| `clear()` for clearing slices/maps | Go 1.21+ | ✅ REQUIRED |
| `min()`, `max()` built-ins | Go 1.21+ | ✅ REQUIRED |
| `cmp.Compare` for comparisons | Go 1.21+ | ✅ REQUIRED |
| Generics for type-safe utilities | Go 1.18+ | ⚠️ WHEN APPROPRIATE |

### Rule 3: Never Use Deprecated Packages

**FORBIDDEN - Do NOT use:**

- ❌ `io/ioutil` - Use `os` and `io` directly
- ❌ `interface{}` - Use `any` (Go 1.18+)
- ❌ Custom slice helpers - Use `slices` package (Go 1.21+)
- ❌ Custom map helpers - Use `maps` package (Go 1.21+)
- ❌ Manual min/max - Use built-in `min()`/`max()` (Go 1.21+)
- ❌ `math/rand` for new code - Use `crypto/rand` or `math/rand/v2`

### Rule 4: Standard Library First

**ALWAYS check if standard library provides the solution:**

Before writing custom code, search:
```
"golang standard library [functionality]"
"pkg.go.dev [related package]"
```

Examples:
- Need to sort? → Use `slices.Sort` or `slices.SortFunc`
- Need to compare? → Use `cmp.Compare`
- Need to filter/map? → Check if `slices` package has it, or use generics
- Need to read files? → Use `os.ReadFile`
- Need HTTP client? → Use `net/http`

### Rule 5: Follow Effective Go

All code MUST follow:
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)

Key conventions:
- Use `MixedCaps` or `mixedCaps` for names (not underscores)
- Use short variable names in short scopes (`i`, `r`, `w`)
- Use descriptive names for package-level declarations
- Group related declarations
- Comment exported identifiers
- Handle errors explicitly

### Rule 6: Error Handling (CRITICAL)

**MANDATORY: Every function that returns an error MUST have it handled or explicitly ignored.**

#### Always Handle or Explicitly Ignore Errors

```go
// ❌ WRONG - Silently ignoring errors (will fail linting)
json.NewEncoder(w).Encode(response)
cmd.MarkFlagRequired("field")
fmt.Scanln(&input)
w.Write([]byte("data"))

// ✅ CORRECT - Explicitly ignoring with _ =
_ = json.NewEncoder(w).Encode(response)  // Test helper, error not actionable
_ = cmd.MarkFlagRequired("field")        // Hardcoded field name, won't fail
_, _ = fmt.Scanln(&input)                // User input, validation happens later
_, _ = w.Write([]byte("data"))           // Test response, error not relevant

// ✅ CORRECT - Proper error handling
if err := json.NewEncoder(w).Encode(response); err != nil {
    return fmt.Errorf("encode response: %w", err)
}
```

#### When to Use `_ =` vs Proper Handling

**Use `_ =` (explicit ignore) when:**
- In test code where error is not relevant
- Error cannot occur (e.g., hardcoded values)
- Error is not actionable (e.g., best-effort operations)
- Always add a comment explaining why

**Use proper handling when:**
- In production code
- Error affects program behavior
- Error indicates a bug or invalid state
- User needs to know about the failure

#### Error Wrapping and Checking

```go
// ✅ Use fmt.Errorf with %w for error wrapping (Go 1.13+)
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}

// ✅ Use errors.Is for error checking
if errors.Is(err, os.ErrNotExist) {
    // handle specific error
}

// ✅ Use errors.As for error type assertion
var pathErr *fs.PathError
if errors.As(err, &pathErr) {
    // handle path error with details
    fmt.Printf("Path error: %s\n", pathErr.Path)
}
```

#### Deferred Function Error Handling

```go
// ❌ WRONG - Error ignored in defer
defer file.Close()

// ✅ CORRECT - Document why ignored
defer file.Close()  // Read-only file, close error not actionable

// ✅ CORRECT - Capture in named return
func processFile(path string) (err error) {
    f, err := os.Open(path)
    if err != nil {
        return err
    }
    defer func() {
        if closeErr := f.Close(); closeErr != nil && err == nil {
            err = closeErr
        }
    }()
    // ... process file
}

// ✅ CORRECT - Wrap in anonymous function for complex cleanup
defer func() { _ = server.Stop() }()
```

#### Nil Checks in Tests

**ALWAYS check for nil before dereferencing in tests:**

```go
// ❌ WRONG - Can panic if nil
flag := cmd.Flags().Lookup("name")
if flag.DefValue != "expected" {  // Panic if flag is nil!
    t.Error("wrong default")
}

// ✅ CORRECT - Check nil first
flag := cmd.Flags().Lookup("name")
if flag == nil {
    t.Error("flag not found")
    return  // Early return prevents nil dereference
}
if flag.DefValue != "expected" {
    t.Errorf("wrong default: got %q, want %q", flag.DefValue, "expected")
}
```

---

## 📋 Code Change Workflow

When making code changes, follow this exact sequence:

```
1. Read the code change request
   ↓
2. Check go.mod for Go version
   ↓
3. WebSearch for:
   - "golang [version] [feature] specification"
   - "golang [package] pkg.go.dev"
   - "golang [feature] best practices"
   ↓
4. Review search results:
   - Go specification
   - Standard library docs
   - Official Go blog articles
   ↓
5. Plan implementation using modern patterns
   ↓
6. Make code changes
   ↓
7. Run: go fmt ./...
   ↓
8. Run: golangci-lint run --timeout=5m
   ↓
9. Fix ALL linting issues in code you wrote
   ↓
10. Run: go test ./...
   ↓
11. Verify: make build
```

**CRITICAL: Do not skip step 8-9. Linting is MANDATORY.**

---

## 🎯 Pattern Replacements

### File Operations

```go
// ❌ OLD (Deprecated)
import "io/ioutil"
data, err := ioutil.ReadFile("file.txt")
err = ioutil.WriteFile("file.txt", data, 0644)

// ✅ NEW (Go 1.16+)
import "os"
data, err := os.ReadFile("file.txt")
err = os.WriteFile("file.txt", data, 0644)
```

### Type Parameters

```go
// ❌ OLD
func Contains(slice []string, item string) bool {
    for _, v := range slice {
        if v == item {
            return true
        }
    }
    return false
}

// ✅ NEW (Go 1.21+)
import "slices"
found := slices.Contains(slice, item)
```

### Any Type

```go
// ❌ OLD
func Process(data interface{}) error

// ✅ NEW (Go 1.18+)
func Process(data any) error
```

### Map/Slice Operations

```go
// ❌ OLD
for k := range myMap {
    delete(myMap, k)
}

// ✅ NEW (Go 1.21+)
clear(myMap)
```

```go
// ❌ OLD
func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}

// ✅ NEW (Go 1.21+)
result := min(a, b)
```

### Sorting

```go
// ❌ OLD
sort.Slice(users, func(i, j int) bool {
    return users[i].Name < users[j].Name
})

// ✅ NEW (Go 1.21+)
import (
    "slices"
    "cmp"
)

slices.SortFunc(users, func(a, b User) int {
    return cmp.Compare(a.Name, b.Name)
})
```

### Generics

```go
// ❌ OLD (Type assertions everywhere)
func Map(items []interface{}, fn func(interface{}) interface{}) []interface{} {
    result := make([]interface{}, len(items))
    for i, item := range items {
        result[i] = fn(item)
    }
    return result
}

// ✅ NEW (Go 1.18+)
func Map[T, U any](items []T, fn func(T) U) []U {
    result := make([]U, len(items))
    for i, item := range items {
        result[i] = fn(item)
    }
    return result
}
```

---

## 🚫 Anti-Patterns to Reject

If you see these patterns, ALWAYS suggest modern alternatives:

### 1. Using io/ioutil

```go
// ❌ REJECT THIS
import "io/ioutil"
data, _ := ioutil.ReadFile("file")

// ✅ SUGGEST THIS
import "os"
data, err := os.ReadFile("file")
if err != nil {
    return fmt.Errorf("read file: %w", err)
}
```

### 2. Using interface{} instead of any

```go
// ❌ REJECT THIS (Go 1.18+)
func Process(data interface{}) error

// ✅ SUGGEST THIS
func Process(data any) error
```

### 3. Manual slice operations

```go
// ❌ REJECT THIS (Go 1.21+)
func Contains(slice []string, item string) bool {
    for _, s := range slice {
        if s == item {
            return true
        }
    }
    return false
}

// ✅ SUGGEST THIS
import "slices"
// Use directly: slices.Contains(slice, item)
```

### 4. Ignoring errors silently

```go
// ❌ REJECT THIS
file.Close()
json.Unmarshal(data, &v)

// ✅ SUGGEST THIS
_ = file.Close()  // Document why ignored
if err := json.Unmarshal(data, &v); err != nil {
    return fmt.Errorf("unmarshal: %w", err)
}
```

---

## 📚 Required Research Sources

Before making code changes, you MUST check these sources:

### Primary Sources (Required)

1. **Go Specification**
   - URL: https://go.dev/ref/spec
   - Search: "golang spec [feature]"

2. **Standard Library Docs**
   - URL: https://pkg.go.dev/std
   - Search: "pkg.go.dev [package]"

3. **Release Notes**
   - URL: https://go.dev/doc/devel/release
   - Search: "golang [version] release notes"

4. **Effective Go**
   - URL: https://go.dev/doc/effective_go
   - Required reading for all Go code

### Secondary Sources (Recommended)

5. **Go Blog**
   - URL: https://go.dev/blog/
   - Search: "golang blog [topic]"

6. **Code Review Comments**
   - URL: https://go.dev/wiki/CodeReviewComments
   - Required for PR reviews

---

## ✅ Quality Checks

After making code changes, you MUST run:

```bash
# 1. Format code (REQUIRED)
go fmt ./...

# 2. Lint code (REQUIRED - not optional!)
golangci-lint run --timeout=5m

# 3. Fix ALL linting issues in your code (REQUIRED)
#    - errcheck: Add _ = or proper error handling
#    - unused: Remove unused code
#    - staticcheck: Fix nil checks, deprecated usage
#    See .claude/rules/linting.md for common fixes

# 4. Run tests (REQUIRED)
go test ./... -short

# 5. Build verification (REQUIRED)
go build -o ./bin/nylas ./cmd/nylas

# 6. Update dependencies (if you added/removed packages)
go mod tidy
```

**CRITICAL: If linting shows errors in code you wrote, fix them immediately.**
**Do NOT:**
- ❌ Skip linting
- ❌ Proceed with failing lint
- ❌ Ask user if you should fix linting errors (just fix them)
- ❌ Leave linting fixes for later

---

## 📖 Documentation Requirements

When introducing new patterns or modern features, you MUST:

1. **Add code comments explaining:**
   - Why this pattern is used
   - What Go version introduced it
   - Link to relevant documentation

Example:
```go
// Use slices.SortFunc (Go 1.21+) for type-safe sorting with custom comparison.
// See: https://pkg.go.dev/slices#SortFunc
slices.SortFunc(users, func(a, b User) int {
    return cmp.Compare(a.Name, b.Name)
})
```

2. **Include in commit message:**
   - What modern feature was used
   - Why it's better than the old approach

Example:
```
refactor: use slices.SortFunc for type-safe sorting

Replace sort.Slice with slices.SortFunc (Go 1.21+) for improved
type safety and clearer comparison logic using cmp.Compare.

See: https://pkg.go.dev/slices#SortFunc
```

---

## 🎓 Learning Resources

Keep these bookmarked for quick reference:

- **Go Specification**: https://go.dev/ref/spec
- **Standard Library**: https://pkg.go.dev/std
- **Effective Go**: https://go.dev/doc/effective_go
- **Code Review Comments**: https://go.dev/wiki/CodeReviewComments
- **Go Blog**: https://go.dev/blog/
- **Release Notes**: https://go.dev/doc/devel/release
- **Uber Go Style Guide**: https://github.com/uber-go/guide/blob/master/style.md

---

## 🚀 Summary

**Every time you write Go code:**

1. ✅ Check Go version
2. ✅ Research official docs
3. ✅ Use modern features
4. ✅ Follow Effective Go
5. ✅ Use standard library
6. ✅ Handle errors properly
7. ✅ Format and vet code
8. ✅ Run tests
9. ✅ Document new patterns

**Remember: If you're not sure, search the official Go docs first!**
