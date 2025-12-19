# Go Modernize

Check Go specification and latest features before implementing code changes.

## Instructions

**CRITICAL: Before making ANY Go code changes, you MUST:**

1. **Check Current Go Version**
```bash
go version
cat go.mod | grep "^go "
```

2. **Review Go Specification & Docs**

Use WebSearch to check for:
- Latest Go release notes: https://go.dev/doc/devel/release
- Go specification: https://go.dev/ref/spec
- Standard library changes: https://pkg.go.dev/std

Search queries to run:
```
"golang $VERSION release notes"
"golang $VERSION new features"
"golang standard library $VERSION updates"
```

3. **Check for Modern Alternatives**

For each code change, verify:

| Old Pattern | Check For | Go Version |
|-------------|-----------|------------|
| `ioutil.ReadFile` | `os.ReadFile` | Go 1.16+ |
| `ioutil.WriteFile` | `os.WriteFile` | Go 1.16+ |
| `io/ioutil` | Direct `os` or `io` usage | Go 1.16+ |
| Manual slice operations | `slices` package | Go 1.21+ |
| Manual map operations | `maps` package | Go 1.21+ |
| `interface{}` | `any` | Go 1.18+ |
| Manual generics workarounds | Type parameters `[T any]` | Go 1.18+ |
| `append(x[:0], y...)` | `clear()` | Go 1.21+ |
| Manual min/max | `min()`, `max()` | Go 1.21+ |
| Custom comparison | `cmp` package | Go 1.21+ |
| `errors.New` + `fmt.Sprintf` | `fmt.Errorf` with `%w` | Go 1.13+ |

4. **Research Before Implementing**

When implementing new features:

a) **Search for official guidance:**
   - "golang how to $TASK official documentation"
   - "golang $PACKAGE best practices"
   - "golang effective go $TOPIC"

b) **Check package documentation:**
   - https://pkg.go.dev/$PACKAGE
   - Look for examples and usage patterns

c) **Review Go blog for context:**
   - https://go.dev/blog/
   - Search for relevant articles

5. **Apply Modern Go Idioms**

Ensure code follows:
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)

## Workflow

```
1. User requests code change
   ↓
2. Check go.mod for Go version
   ↓
3. WebSearch for latest Go features/docs
   ↓
4. Review Go spec for relevant sections
   ↓
5. Apply modern patterns and idioms
   ↓
6. Implement changes
   ↓
7. Run: go fmt, go vet, golangci-lint
   ↓
8. Document any new patterns used
```

## Examples

### Example 1: File Operations

**User Request:** "Read a file"

**Process:**
1. Check Go version: 1.24.0
2. Search: "golang 1.24 read file best practice"
3. Find: `os.ReadFile` is standard (since Go 1.16)
4. Implement:
```go
// Modern (Go 1.16+)
data, err := os.ReadFile("file.txt")

// NOT this (deprecated):
// data, err := ioutil.ReadFile("file.txt")
```

### Example 2: Slice Operations

**User Request:** "Sort a slice of structs"

**Process:**
1. Check Go version: 1.24.0
2. Search: "golang 1.21 slices package"
3. Find: `slices.SortFunc` available
4. Implement:
```go
import "slices"

// Modern (Go 1.21+)
slices.SortFunc(users, func(a, b User) int {
    return cmp.Compare(a.Name, b.Name)
})

// NOT this (verbose):
// sort.Slice(users, func(i, j int) bool {
//     return users[i].Name < users[j].Name
// })
```

### Example 3: Generics vs Interfaces

**User Request:** "Create a utility function"

**Process:**
1. Check Go version: 1.24.0
2. Search: "golang generics when to use"
3. Decide: Use generics for type-safe collections
4. Implement:
```go
// Modern (Go 1.18+) - Type safe
func Map[T, U any](slice []T, fn func(T) U) []U {
    result := make([]U, len(slice))
    for i, v := range slice {
        result[i] = fn(v)
    }
    return result
}

// NOT this (loses type information):
// func Map(slice []interface{}, fn func(interface{}) interface{}) []interface{}
```

## Research Checklist

Before implementing, verify:

- [ ] Checked current Go version from go.mod
- [ ] Searched Go release notes for version-specific features
- [ ] Reviewed relevant package documentation on pkg.go.dev
- [ ] Checked Go specification for language features
- [ ] Applied modern idioms and patterns
- [ ] Avoided deprecated functions/packages
- [ ] Used standard library when available
- [ ] Considered performance implications
- [ ] Followed Go naming conventions
- [ ] Added appropriate error handling

## Key Resources

### Must Check (in order):

1. **Go Version**
   ```bash
   go version
   grep "^go " go.mod
   ```

2. **Go Specification**
   - https://go.dev/ref/spec
   - Check for language features available in current version

3. **Standard Library Docs**
   - https://pkg.go.dev/std
   - Look for existing solutions before writing custom code

4. **Release Notes**
   - https://go.dev/doc/devel/release
   - Check what's new in current Go version

5. **Effective Go**
   - https://go.dev/doc/effective_go
   - Follow established patterns

6. **Go Blog**
   - https://go.dev/blog/
   - Read articles about relevant topics

## Anti-Patterns to Avoid

### ❌ Don't Do This

```go
// Using deprecated packages
import "io/ioutil"

// Using interface{} when any is available (Go 1.18+)
func Process(data interface{}) error

// Manual slice operations when slices package available (Go 1.21+)
func Contains(slice []string, item string) bool {
    for _, s := range slice {
        if s == item {
            return true
        }
    }
    return false
}

// Not using clear() (Go 1.21+)
myMap = make(map[string]int)  // Recreating instead of clearing

// Not using min/max (Go 1.21+)
func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}
```

### ✅ Do This Instead

```go
// Use modern packages
import "os"

// Use any (Go 1.18+)
func Process(data any) error

// Use slices package (Go 1.21+)
import "slices"
func Contains(slice []string, item string) bool {
    return slices.Contains(slice, item)
}

// Use clear() (Go 1.21+)
clear(myMap)

// Use built-in min/max (Go 1.21+)
result := min(a, b)
```

## Version-Specific Features

### Go 1.24 (Latest)
- Check release notes: https://go.dev/doc/go1.24

### Go 1.23
- `unique` package for interning
- `iter` package for iterators
- Timer channel behavior improvements

### Go 1.22
- Enhanced `for` loop variable semantics
- `math/rand/v2` with better random number generation
- Improved HTTP routing patterns

### Go 1.21
- `clear()` built-in function
- `min()`, `max()` built-in functions
- `maps` package
- `slices` package
- `cmp` package

### Go 1.18
- Generics (type parameters)
- `any` type alias for `interface{}`
- Fuzzing support

## Output Format

When suggesting code changes, always include:

1. **Research Summary:**
   ```
   Checked: Go 1.24.0
   Feature: slices.SortFunc (available since Go 1.21)
   Reference: https://pkg.go.dev/slices#SortFunc
   ```

2. **Code Changes:**
   ```go
   // Old pattern (if replacing)
   // New pattern with explanation
   ```

3. **Why This Approach:**
   - Version compatibility
   - Performance benefits
   - Idiomatic Go
   - Standard library usage

## Post-Implementation

After making changes:

```bash
# Format code
go fmt ./...

# Vet code
go vet ./...

# Run linter
golangci-lint run

# Run tests
go test ./...

# Update dependencies if needed
go mod tidy
```

## Notes

- **Always prefer standard library** over third-party packages
- **Check Go version compatibility** before using new features
- **Document why** modern patterns are used (for learning)
- **Be conservative** with experimental features
- **Test thoroughly** after applying modern patterns

---

**Remember: Modern Go code is:**
- Simple and readable
- Uses standard library
- Leverages language features appropriately
- Follows established conventions
- Well-documented and tested
