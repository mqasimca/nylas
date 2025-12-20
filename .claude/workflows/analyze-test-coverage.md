# Analyze Test Coverage

This workflow analyzes test coverage across the codebase and suggests improvements for packages with low coverage.

## What It Does

1. Generates comprehensive test coverage report
2. Identifies packages below target coverage thresholds
3. Lists specific uncovered functions and methods
4. Suggests concrete tests to write
5. Calculates coverage deltas if baseline exists

## When to Use

- After adding new features (verify coverage)
- Before releases (quality check)
- When improving code quality
- As part of regular maintenance
- When reviewing pull requests

## Coverage Targets

| Package Type | Minimum Coverage | Target Coverage |
|--------------|------------------|-----------------|
| Core Adapters | 70% | 85%+ |
| Business Logic | 60% | 80%+ |
| CLI Commands | 50% | 70%+ |
| Utilities | 90% | 100% |

## Step-by-Step Workflow

### Step 1: Generate Coverage Report

```bash
# Generate coverage for all packages (unit tests only)
go test ./... -short -coverprofile=coverage.out

# If you want to include integration tests:
# go test ./... -tags=integration -coverprofile=coverage_full.out
```

### Step 2: Analyze Coverage by Package

```bash
# View coverage summary
go tool cover -func=coverage.out | grep "total:"

# View all packages with coverage
go tool cover -func=coverage.out
```

### Step 3: Identify Low Coverage Packages

```bash
# Find packages with < 70% coverage
go tool cover -func=coverage.out | grep -E "^.*\.go:[0-9]+" | awk '{print $1, $3}' | sort -t: -k1,1 | awk -F: '{pkg=$1; sub(/\/[^\/]+$/, "", pkg); if (pkg != last) {if (last != "") print last, total; last=pkg; total=0; count=0} total+=$2; count++} END {print last, total/count}' | awk '$2 < 70 {print $1, $2"%"}'

# Or use a simpler approach - check coverage by specific package:
go test ./internal/adapters/browser -coverprofile=browser_cov.out
go tool cover -func=browser_cov.out
```

### Step 4: Generate HTML Coverage Report

```bash
# Generate interactive HTML report
go tool cover -html=coverage.out -o coverage.html

# Open in browser (macOS)
open coverage.html

# Open in browser (Linux)
xdg-open coverage.html
```

### Step 5: Analyze Uncovered Code

The HTML report shows:
- **Green lines**: Covered by tests
- **Red lines**: Not covered by tests
- **Gray lines**: Not executable (comments, declarations)

Focus on red lines in critical paths:
- Error handling
- Edge cases
- Public APIs
- Business logic

### Step 6: Suggest Tests to Write

Based on uncovered code, suggest:

1. **Missing unit tests**: Test files that don't exist
2. **Incomplete tests**: Existing tests missing scenarios
3. **Edge cases**: Error paths, boundary conditions
4. **Integration gaps**: Missing integration tests

## Example Analysis Output

```
Coverage Analysis Results
=========================

Overall Coverage: 68.4% (target: 70%+)

Packages Below Target:
----------------------

❌ internal/adapters/browser (0.0% → target: 70%)
   Missing tests:
   - internal/adapters/browser/browser_test.go

⚠️  internal/cli/email (45.2% → target: 70%)
   Uncovered functions:
   - sendWithRetry() - retry logic not tested
   - parseAttachments() - parsing edge cases missing
   Suggested tests:
   - TestSendEmail_WithRetry_Success
   - TestSendEmail_WithRetry_MaxRetriesExceeded
   - TestParseAttachments_InvalidFormat

✅ internal/util (100% → target: 90%)
   Excellent coverage!

Action Items:
-------------
1. Add browser_test.go (priority: HIGH)
2. Add retry tests to email package (priority: HIGH)
3. Add attachment parsing tests (priority: MEDIUM)

Estimated effort: 2-3 hours
Expected coverage increase: 68.4% → 78.2%
```

## Common Patterns to Test

### 1. Success Path

```go
func TestFunction_Success(t *testing.T) {
    // Happy path test
}
```

### 2. Error Handling

```go
func TestFunction_ErrorCases(t *testing.T) {
    tests := []struct {
        name        string
        input       Input
        wantErr     bool
        errContains string
    }{
        {"nil input", nil, true, "input cannot be nil"},
        {"invalid format", invalidInput, true, "invalid format"},
    }
    // Test all error scenarios
}
```

### 3. Edge Cases

```go
func TestFunction_EdgeCases(t *testing.T) {
    // Empty input
    // Maximum values
    // Minimum values
    // Boundary conditions
}
```

### 4. Integration Paths

```go
//go:build integration

func TestFeature_Integration(t *testing.T) {
    // Test with real dependencies
}
```

## Automation Script

Create `scripts/coverage_analysis.sh`:

```bash
#!/bin/bash
# Coverage analysis script

echo "🔍 Analyzing test coverage..."
echo

# Generate coverage
go test ./... -short -coverprofile=coverage.out 2>&1 | grep -v "no test files"

# Overall coverage
echo "📊 Overall Coverage:"
go tool cover -func=coverage.out | grep "total:" | awk '{print "   " $3}'
echo

# Packages below 70%
echo "❌ Packages Below 70% Coverage:"
go tool cover -func=coverage.out | awk '
    /^total:/ { next }
    {
        file = $1
        coverage = $3
        gsub(/%/, "", coverage)

        # Extract package from file path
        split(file, parts, "/")
        pkg = ""
        for (i = 1; i < length(parts); i++) {
            if (i > 1) pkg = pkg "/"
            pkg = pkg parts[i]
        }

        # Track package coverage
        if (pkg != last_pkg) {
            if (last_pkg != "" && pkg_total/pkg_count < 70) {
                printf "   %s: %.1f%%\n", last_pkg, pkg_total/pkg_count
            }
            last_pkg = pkg
            pkg_total = 0
            pkg_count = 0
        }

        pkg_total += coverage
        pkg_count++
    }
    END {
        if (last_pkg != "" && pkg_count > 0 && pkg_total/pkg_count < 70) {
            printf "   %s: %.1f%%\n", last_pkg, pkg_total/pkg_count
        }
    }
'

echo
echo "📈 To view detailed report:"
echo "   go tool cover -html=coverage.out -o coverage.html"
echo "   open coverage.html"
```

Make it executable:

```bash
chmod +x scripts/coverage_analysis.sh
./scripts/coverage_analysis.sh
```

## Tips for Improving Coverage

### 1. Start with Zero Coverage Packages

Priority order:
1. Core adapters (0% coverage)
2. Business logic (0% coverage)
3. CLI commands (0% coverage)
4. Utilities (0% coverage)

### 2. Use Table-Driven Tests

Quickly cover multiple scenarios:

```go
tests := []struct {
    name    string
    input   string
    want    string
    wantErr bool
}{
    {"valid input", "test", "TEST", false},
    {"empty input", "", "", true},
    {"special chars", "a@b", "A@B", false},
}
```

### 3. Focus on Public APIs

Test public functions first (exported names). Internal helpers will get covered indirectly.

### 4. Don't Chase 100% Coverage

Some code is not worth testing:
- Simple getters/setters
- Trivial constructors
- Generated code
- Main functions

Target 70-85% for meaningful code.

### 5. Use Coverage to Find Missing Tests

Coverage tools show:
- Untested error paths
- Missed edge cases
- Dead code (remove it!)

## Integration with CI/CD

Add to `.github/workflows/test.yml`:

```yaml
- name: Check coverage
  run: |
    go test ./... -short -coverprofile=coverage.out
    go tool cover -func=coverage.out | grep "total:" | awk '{if ($3 < "70.0%") exit 1}'
```

This fails CI if coverage drops below 70%.

## Checklist

Before marking coverage analysis complete:

- [ ] Generated coverage report (`coverage.out`)
- [ ] Viewed overall coverage percentage
- [ ] Identified packages below 70% coverage
- [ ] Created HTML report for detailed view
- [ ] Listed uncovered functions/methods
- [ ] Suggested specific tests to write
- [ ] Estimated effort to reach target coverage
- [ ] Created action items with priorities

## Summary

**Coverage analysis helps you:**
- Identify untested code
- Find missing test scenarios
- Prioritize testing efforts
- Track quality metrics
- Ensure critical paths are tested

**Remember:** Coverage is a metric, not a goal. Write meaningful tests that verify behavior, not just increase numbers.
