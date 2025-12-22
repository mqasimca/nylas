.PHONY: build test test-short test-integration test-coverage clean install lint deps check security

VERSION ?= dev
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-s -w -X github.com/mqasimca/nylas/internal/cli.Version=$(VERSION) -X github.com/mqasimca/nylas/internal/cli.Commit=$(COMMIT) -X github.com/mqasimca/nylas/internal/cli.BuildDate=$(BUILD_DATE)"

build:
	@mkdir -p bin
	go build $(LDFLAGS) -o bin/nylas ./cmd/nylas

test:
	go test ./... -v

test-coverage:
	go test ./... -cover -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

install: build
	cp bin/nylas $(GOPATH)/bin/nylas

lint:
	golangci-lint run

deps:
	go mod tidy
	go mod download

# Quick test (skip slow tests)
test-short:
	go test ./... -short

# Integration tests (requires NYLAS_API_KEY and NYLAS_GRANT_ID env vars)
test-integration:
	go test ./... -tags=integration -v

# Security scan for credentials and secrets
security:
	@echo "=== Security Scan ==="
	@echo "Checking for hardcoded API keys..."
	@grep -rE "nyk_v0[a-zA-Z0-9_]{20,}" --include="*.go" . | grep -v "_test.go" && echo "WARNING: Possible API key found!" || echo "✓ No API keys found"
	@echo ""
	@echo "Checking for credential patterns..."
	@grep -rE "(api_key|password|secret)\s*=\s*\"[^\"]+\"" --include="*.go" . | grep -v "_test.go" | grep -v "mock.go" && echo "WARNING: Possible credentials found!" || echo "✓ No hardcoded credentials"
	@echo ""
	@echo "Checking for full credential logging..."
	@grep -rE "fmt\.(Print|Fprint|Sprint).*[Aa]pi[Kk]ey[^:\[]" --include="*.go" . | grep -v "token.go" | grep -v "doctor.go" && echo "WARNING: Possible credential logging!" || echo "✓ No credential logging"
	@echo ""
	@echo "Checking staged files..."
	@git diff --cached --name-only | grep -E "\.(env|key|pem|json)$$" && echo "WARNING: Sensitive file staged!" || echo "✓ No sensitive files staged"
	@echo ""
	@echo "=== Security scan complete ==="

# Full check before commit
check: lint test-short security build
	@echo "All checks passed!"

# Run a specific package's tests
# Usage: make test-pkg PKG=email
test-pkg:
	go test ./internal/cli/$(PKG)/... -v

# Quick build and run
run: build
	./bin/nylas $(ARGS)

# Show help
help:
	@echo "Available targets:"
	@echo "  build           - Build the CLI binary"
	@echo "  test            - Run all tests with verbose output"
	@echo "  test-short      - Run tests (skip slow ones)"
	@echo "  test-integration- Run integration tests"
	@echo "  test-coverage   - Run tests with coverage report"
	@echo "  test-pkg PKG=x  - Run tests for specific package"
	@echo "  lint            - Run golangci-lint"
	@echo "  security        - Run security scan for credentials"
	@echo "  check           - Run lint, test, security, build (pre-commit)"
	@echo "  clean           - Remove build artifacts"
	@echo "  install         - Install binary to GOPATH/bin"
	@echo "  deps            - Tidy and download dependencies"
	@echo "  run ARGS='...'  - Build and run with arguments"
	@echo "  help            - Show this help"
