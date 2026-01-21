#!/bin/bash
# Quality gate - runs when Claude tries to stop
# Ensures code changes pass basic quality checks

set -euo pipefail

# Check if any Go files were modified
MODIFIED_GO=$(git diff --name-only 2>/dev/null | grep -E '\.go$' | head -5 || true)

if [ -n "$MODIFIED_GO" ]; then
    echo "Running quality checks on modified Go files..."

    # Run go fmt
    if ! go fmt ./... > /dev/null 2>&1; then
        echo '{"decision": "block", "reason": "go fmt failed - please fix formatting issues"}' >&2
        exit 2
    fi

    # Run go vet
    if ! go vet ./... > /dev/null 2>&1; then
        echo '{"decision": "block", "reason": "go vet found issues - please fix before completing"}' >&2
        exit 2
    fi

    # Run linter (quick check with timeout) - only on changed files for speed
    LINT_OUTPUT=$(timeout 30 golangci-lint run --new-from-rev=HEAD~1 --timeout=30s 2>&1 || true)
    LINT_ERRORS=$(echo "$LINT_OUTPUT" | grep -c "error" || true)
    if [ "$LINT_ERRORS" -gt 0 ]; then
        echo '{"decision": "block", "reason": "golangci-lint found '"$LINT_ERRORS"' errors - run: golangci-lint run --timeout=5m"}' >&2
        exit 2
    fi

    # NOTE: Unit tests removed from Stop hook for performance (~73s savings)
    # Run 'make ci' manually for full validation before committing

    echo "Quality checks passed for Go files (run 'make ci' for full tests)"
fi

# Check if any JavaScript files were modified
MODIFIED_JS=$(git diff --name-only 2>/dev/null | grep -E '\.js$' | head -5 || true)

if [ -n "$MODIFIED_JS" ]; then
    echo "Running quality checks on modified JavaScript files..."

    # Check for syntax errors using node --check
    for file in $MODIFIED_JS; do
        if [ -f "$file" ]; then
            if ! node --check "$file" > /dev/null 2>&1; then
                echo '{"decision": "block", "reason": "JavaScript syntax error in '"$file"' - please fix before completing"}' >&2
                exit 2
            fi
        fi
    done

    echo "Quality checks passed for JavaScript files"
fi

# All checks passed
echo "All quality checks passed"
exit 0
