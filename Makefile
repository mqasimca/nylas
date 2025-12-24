.PHONY: build test test-short test-integration test-integration-clean test-cleanup test-coverage clean install lint deps check security check-context

VERSION ?= dev
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-s -w -X github.com/mqasimca/nylas/internal/cli.Version=$(VERSION) -X github.com/mqasimca/nylas/internal/cli.Commit=$(COMMIT) -X github.com/mqasimca/nylas/internal/cli.BuildDate=$(BUILD_DATE)"

build:
	@mkdir -p bin
	@go clean -cache
	go build $(LDFLAGS) -o bin/nylas ./cmd/nylas

test:
	@go clean -testcache
	go test ./... -v

test-coverage:
	@go clean -testcache
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
	@go clean -testcache
	go test ./... -short

# Integration tests (requires NYLAS_API_KEY and NYLAS_GRANT_ID env vars)
# Uses 10 minute timeout to prevent hanging on slow LLM calls
test-integration:
	@go clean -testcache
	go test ./... -tags=integration -v -timeout 10m

# Integration tests excluding slow LLM-dependent tests (for when Ollama is slow/unavailable)
# Runs: Admin, Timezone, AIConfig, CalendarAI (Basic, Adapt, Analyze working hours)
test-integration-fast:
	@go clean -testcache
	NYLAS_TEST_BINARY=$(CURDIR)/bin/nylas go test ./internal/cli/integration/... -tags=integration -v -timeout 2m \
		-run "TestCLI_Admin|TestCLI_Timezone|TestCLI_AIConfig|TestCLI_AIProvider|TestCLI_CalendarAI_Basic|TestCLI_CalendarAI_Adapt|TestCLI_CalendarAI_Analyze_Respects|TestCLI_CalendarAI_Analyze_Default|TestCLI_CalendarAI_Analyze_Disabled|TestCLI_CalendarAI_Analyze_Focus|TestCLI_CalendarAI_Analyze_With"

# Integration tests with extended timeout and cleanup
test-integration-clean: test-integration test-cleanup

# Clean up test resources (virtual calendars, test grants, test events, test emails, etc.)
test-cleanup:
	@echo "=== Cleaning up test resources ==="
	@echo ""
	@echo "1. Cleaning test emails (messages and drafts)..."
	@./bin/nylas email list --limit 100 --id 2>/dev/null | \
		grep -E "(Test|Integration|Draft|AI|Metadata)" -A1 | \
		grep "ID:" | \
		awk '{print $$2}' | \
		while read msg_id; do \
			if [ ! -z "$$msg_id" ]; then \
				echo "  Deleting test message: $$msg_id"; \
				./bin/nylas email delete $$msg_id --force 2>/dev/null && \
				echo "    ✓ Deleted message $$msg_id" || echo "    ⚠ Could not delete $$msg_id"; \
			fi \
		done
	@echo ""
	@echo "2. Cleaning test events from calendars..."
	@./bin/nylas calendar events list --limit 100 2>/dev/null | \
		awk '/AI Test|Test Meeting|Integration Test|test-event/ { \
			getline; getline; getline; getline; \
			if ($$0 ~ /ID:/) { split($$0, arr, " "); print arr[2] } \
		}' | \
		while read event_id; do \
			if [ ! -z "$$event_id" ]; then \
				echo "  Deleting test event: $$event_id"; \
				./bin/nylas calendar events delete $$event_id --force 2>/dev/null && \
				echo "    ✓ Deleted event $$event_id" || echo "    ⚠ Could not delete $$event_id"; \
			fi \
		done
	@echo ""
	@echo "3. Cleaning test virtual calendar grants..."
	@./bin/nylas admin grants list | grep -E "^(test-|integration-)" | awk '{print $$2}' | while read grant_id; do \
		if [ ! -z "$$grant_id" ] && [ "$$grant_id" != "ID" ]; then \
			echo "  Deleting test grant: $$grant_id"; \
			curl -s -X DELETE "https://api.us.nylas.com/v3/grants/$$grant_id" \
				-H "Authorization: Bearer $$NYLAS_API_KEY" > /dev/null && \
			echo "    ✓ Deleted grant $$grant_id" || echo "    ✗ Failed to delete $$grant_id"; \
		fi \
	done
	@echo ""
	@echo "✓ Test cleanup complete"

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

# Check context size for Claude Code
check-context:
	@echo "📊 Context Size Report"
	@echo "======================"
	@echo ""
	@echo "Auto-loaded files (excluding FAQ, EXAMPLES, TROUBLESHOOTING, INDEX per .claudeignore):"
	@ls -lh CLAUDE.md .claude/rules/*.md docs/AI.md docs/ARCHITECTURE.md docs/COMMANDS.md docs/DEVELOPMENT.md docs/SECURITY.md docs/TIMEZONE.md docs/TUI.md docs/WEBHOOKS.md 2>/dev/null | awk '{print $$5, $$9}'
	@echo ""
	@TOTAL=$$(ls -l CLAUDE.md .claude/rules/*.md docs/AI.md docs/ARCHITECTURE.md docs/COMMANDS.md docs/DEVELOPMENT.md docs/SECURITY.md docs/TIMEZONE.md docs/TUI.md docs/WEBHOOKS.md 2>/dev/null | awk '{sum+=$$5} END {print int(sum/1024)}'); \
	TIMEZONE=$$(ls -l docs/TIMEZONE.md 2>/dev/null | awk '{print int($$5/1024)}'); \
	echo "Total auto-loaded context: $${TOTAL}KB (~$$((TOTAL / 4)) tokens)"; \
	echo "TIMEZONE.md: $${TIMEZONE}KB"; \
	echo ""; \
	if [ $$TOTAL -gt 60 ]; then \
		echo "⚠️  Context exceeds 60KB budget (currently $${TOTAL}KB)"; \
	else \
		echo "✅ Context within 60KB budget ($${TOTAL}KB)"; \
	fi; \
	if [ $$TIMEZONE -gt 5 ]; then \
		echo "⚠️  TIMEZONE.md exceeds 5KB target (currently $${TIMEZONE}KB)"; \
	else \
		echo "✅ TIMEZONE.md within 5KB target ($${TIMEZONE}KB)"; \
	fi

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
	@echo "  build                - Build the CLI binary"
	@echo "  test                 - Run all tests with verbose output"
	@echo "  test-short           - Run tests (skip slow ones)"
	@echo "  test-integration     - Run integration tests"
	@echo "  test-integration-clean - Run integration tests + cleanup"
	@echo "  test-cleanup         - Clean up test resources (grants, calendars)"
	@echo "  test-coverage        - Run tests with coverage report"
	@echo "  test-pkg PKG=x       - Run tests for specific package"
	@echo "  lint                 - Run golangci-lint"
	@echo "  security             - Run security scan for credentials"
	@echo "  check                - Run lint, test, security, build (pre-commit)"
	@echo "  check-context        - Check Claude Code context size"
	@echo "  clean                - Remove build artifacts"
	@echo "  install              - Install binary to GOPATH/bin"
	@echo "  deps                 - Tidy and download dependencies"
	@echo "  run ARGS='...'       - Build and run with arguments"
	@echo "  help                 - Show this help"
