# Test Coverage Report - December 29, 2024

## Summary

**Overall Coverage: 37.6%**

**Goal: Close to 100% unit test coverage**

## Files Over 500 Lines

**Total: 95 files exceed 500 lines**

### Top 20 Largest Files (Need Refactoring):

| Lines | File | Status |
|-------|------|--------|
| 2,619 | `internal/tui/views.go` | 🔴 CRITICAL - Needs major refactoring |
| 2,050 | `internal/adapters/nylas/integration_test.go` | ⚠️ Test file - OK to be large |
| 1,855 | `internal/cli/ui/server_test.go` | ⚠️ Test file - OK to be large |
| 1,623 | `internal/adapters/nylas/demo.go` | 🔴 CRITICAL - Mock data, refactor |
| 1,544 | `internal/cli/calendar/helpers_test.go` | ⚠️ Test file - OK to be large |
| 1,459 | `internal/adapters/nylas/mock.go` | 🔴 CRITICAL - Mock implementation, refactor |
| 1,286 | `internal/cli/ui/server.go` | 🔴 CRITICAL - Needs major refactoring |
| 1,202 | `internal/tui/app_test.go` | ⚠️ Test file - OK to be large |
| 1,162 | `internal/tui2/models/compose.go` | 🔴 Needs refactoring |
| 1,125 | `internal/cli/calendar/events.go` | 🔴 Needs refactoring |
| 1,069 | `internal/cli/integration/ai_config_test.go` | ⚠️ Test file - OK to be large |
| 1,060 | `internal/tui2/components/calendar_grid.go` | 🔴 Needs refactoring |
| 1,058 | `internal/tui2/models/calendar.go` | 🔴 Needs refactoring |
| 1,004 | `internal/cli/demo/email.go` | 🔴 Demo data, refactor |
| 1,003 | `internal/adapters/nylas/client_test.go` | ⚠️ Test file - OK to be large |
| 995 | `internal/tui/calendar.go` | 🔴 Needs refactoring |
| 994 | `internal/adapters/mcp/proxy_test.go` | ⚠️ Test file - OK to be large |
| 984 | `internal/air/handlers_types_test.go` | ⚠️ Test file - OK to be large |
| 968 | `internal/air/cache/cache_search_extended_test.go` | ⚠️ Test file - OK to be large |
| 956 | `internal/cli/integration/email_test.go` | ⚠️ Test file - OK to be large |

**🔴 Production files over 500 lines: 48 files**
**⚠️ Test files over 500 lines: 47 files (acceptable)**

## Test Coverage by Package

### 🔴 CRITICAL - Zero Coverage (11 packages):

| Package | Coverage | Priority | Action Needed |
|---------|----------|----------|---------------|
| `cmd/nylas` | 0.0% | 🔴 HIGH | Add main.go integration tests |
| `internal/adapters/nylas/demo` | 0.0% | 🔴 HIGH | Add unit tests for demo client |
| `internal/adapters/utilities` | 0.0% | 🔴 HIGH | Add utility tests |
| `internal/adapters/utilities/webhook` | 0.0% | 🔴 HIGH | Add webhook utility tests |
| `internal/cli/demo` | 0.0% | 🟡 MEDIUM | Add demo command tests |
| `internal/ports` | 0.0% | 🟡 MEDIUM | Interface package - tests optional |
| `internal/tui2/state` | 0.0% | 🔴 HIGH | Add state management tests |
| `internal/tui2/utils` | 0.0% | 🔴 HIGH | Add utility function tests |
| `examples/minimal-feature/*` | 0.0% | 🟢 LOW | Example code - tests optional |

### ⚠️ LOW Coverage (< 25%):

| Package | Coverage | Needs Improvement |
|---------|----------|-------------------|
| `internal/app/auth` | 2.8% | ✅ Add auth flow tests |
| `internal/cli/mcp` | 4.3% | ✅ Add MCP command tests |
| `internal/cli/update` | 8.8% | ✅ Add update command tests |
| `internal/adapters/utilities/email` | 13.2% | ✅ Add email utility tests |
| `internal/cli/inbound` | 13.9% | ✅ Add inbound tests |
| `internal/cli/webhook` | 14.7% | ✅ Add webhook tests |
| `internal/cli/contacts` | 17.6% | ✅ Add contact tests |
| `internal/cli/otp` | 18.1% | ✅ Add OTP tests |
| `internal/cli/notetaker` | 19.8% | ✅ Add notetaker tests |
| `internal/adapters/utilities/contacts` | 20.7% | ✅ Add contact utility tests |
| `internal/tui2` | 21.0% | ✅ Add TUI2 tests |
| `internal/adapters/ai` | 22.0% | ✅ Add AI adapter tests |
| `internal/cli/email` | 22.6% | ✅ Add email command tests |
| `internal/cli/calendar` | 22.9% | ✅ Add calendar tests |
| `internal/cli` | 23.1% | ✅ Add CLI tests |
| `internal/cli/scheduler` | 24.3% | ✅ Add scheduler tests |
| `internal/cli/ai` | 24.4% | ✅ Add AI command tests |

### 🟢 Good Coverage (> 75%):

| Package | Coverage | Status |
|---------|----------|--------|
| `internal/tui2/styles` | 100.0% | ✅ Perfect |
| `internal/util` | 100.0% | ✅ Perfect |
| `internal/cli/timezone` | 91.8% | ✅ Excellent |
| `internal/tui2/components` | 79.0% | ✅ Good |

## Action Plan to Reach 100% Coverage

### Phase 1: Zero Coverage Packages (Priority 1)

**Estimated Impact: +15% total coverage**

1. ✅ **internal/adapters/nylas/demo** (0% → 80%)
   - Add tests for demo client methods
   - Test mock data generation
   - Files: `demo/base.go`, `demo/calendars.go`, `demo/contacts.go`, etc.

2. ✅ **internal/tui2/state** (0% → 90%)
   - Add state management tests
   - Test state transitions
   - File: `state/state.go`

3. ✅ **internal/tui2/utils** (0% → 90%)
   - Add utility function tests
   - Files: `utils/*.go`

4. ✅ **internal/adapters/utilities** (0% → 80%)
   - Add base utility tests
   - Files: `utilities/*.go`

5. ✅ **internal/adapters/utilities/webhook** (0% → 80%)
   - Add webhook utility tests
   - Files: `utilities/webhook/*.go`

### Phase 2: Low Coverage CLI Commands (Priority 2)

**Estimated Impact: +10% total coverage**

6. ✅ **internal/cli/mcp** (4.3% → 70%)
   - Add MCP command tests
   - Test install, serve, status, uninstall

7. ✅ **internal/cli/update** (8.8% → 70%)
   - Add update command tests
   - Test version checking, downloading

8. ✅ **internal/cli/webhook** (14.7% → 70%)
   - Add webhook command tests
   - Test list, create, update, delete

9. ✅ **internal/cli/contacts** (17.6% → 70%)
   - Add contact command tests
   - Test CRUD operations

10. ✅ **internal/cli/otp** (18.1% → 70%)
    - Add OTP tests
    - Test code generation/verification

### Phase 3: Medium Coverage Packages (Priority 3)

**Estimated Impact: +8% total coverage**

11. ✅ **internal/cli/email** (22.6% → 70%)
    - Add more email command tests
    - Test edge cases, error handling

12. ✅ **internal/cli/calendar** (22.9% → 70%)
    - Add more calendar tests
    - Test timezone handling, conflicts

13. ✅ **internal/adapters/ai** (22.0% → 70%)
    - Add AI adapter tests
    - Test all providers (OpenAI, Claude, Groq, Ollama)

### Phase 4: Refactor Large Files (Priority 4)

**Files to split (500+ lines):**

1. 🔴 **internal/tui/views.go** (2,619 lines → max 500 per file)
   - Split into: `views_list.go`, `views_detail.go`, `views_compose.go`, `views_calendar.go`, `views_helpers.go`

2. 🔴 **internal/adapters/nylas/demo.go** (1,623 lines → max 500 per file)
   - Split into: `demo/messages.go`, `demo/calendars.go`, `demo/contacts.go`, `demo/helpers.go`

3. 🔴 **internal/adapters/nylas/mock.go** (1,459 lines → max 500 per file)
   - Split into: `mock/messages.go`, `mock/calendars.go`, `mock/contacts.go`, `mock/helpers.go`

4. 🔴 **internal/cli/ui/server.go** (1,286 lines → max 500 per file)
   - Split into: `ui/server_handlers.go`, `ui/server_middleware.go`, `ui/server_sync.go`, `ui/server_offline.go`

5. 🔴 **internal/tui2/models/compose.go** (1,162 lines → max 500 per file)
   - Split into: `models/compose_state.go`, `models/compose_view.go`, `models/compose_handlers.go`

## Recommendations

### Immediate Actions (This Week):

1. ✅ Create unit tests for all 0% coverage packages
2. ✅ Focus on `internal/adapters/nylas/demo` (1,623 lines, 0% coverage)
3. ✅ Focus on `internal/tui2/state` and `internal/tui2/utils` (0% coverage)
4. ✅ Add tests for `internal/cli/mcp`, `internal/cli/update`, `internal/cli/webhook`

### Medium Term (Next 2 Weeks):

5. ✅ Improve coverage for all CLI commands to 70%+
6. ✅ Improve coverage for all adapters to 70%+
7. ✅ Refactor `views.go` (2,619 lines) into smaller modules

### Long Term (Next Month):

8. ✅ Refactor all 48 production files over 500 lines
9. ✅ Achieve 80%+ total coverage
10. ✅ Implement table-driven tests for all new code

## Testing Strategy

### For Each Package:

1. **Create `<package>_test.go` file**
2. **Write table-driven tests** for all public functions
3. **Test edge cases**: nil inputs, empty strings, zero values
4. **Test error paths**: invalid inputs, network errors, timeouts
5. **Use mocks** for external dependencies (API calls, file system)
6. **Verify behavior**, not implementation

### Example Test Template:

```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name    string
        input   InputType
        want    OutputType
        wantErr bool
    }{
        {"valid input", validInput, expectedOutput, false},
        {"invalid input", invalidInput, nil, true},
        {"edge case", edgeInput, edgeOutput, false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := FunctionName(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("got %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Current Status

✅ **Completed (December 29, 2024):**
- ✅ Base client utilities (100% coverage)
- ✅ Common utilities (63.5% coverage) - time, string, context, path helpers
- ✅ Security modules (100% coverage) - validation, path security
- ✅ Code deduplication across AI clients (4.5% code reduction)
- ✅ **Zero-coverage packages eliminated (4 packages):**
  - internal/adapters/nylas/demo: 0% → 100% (+100%)
  - internal/tui2/state: 0% → 100% (+100%)
  - internal/tui2/utils: 0% → 92.9% (+92.9%)
  - internal/adapters/utilities/webhook: 0% → 89.7% (+89.7%)
- ✅ **61 new test functions created** across 6 test files
- ✅ **Overall coverage: 37.6% → 37.9%** (+0.3%)

📋 **Documentation Created:**
- ✅ COVERAGE_REPORT.md - Comprehensive test coverage analysis
- ✅ REFACTORING_GUIDE.md - Step-by-step refactoring guide for 48 files

⏳ **In Progress:**
- File size refactoring (48 files identified, guide created)

❌ **Not Started:**
- CLI command tests (8 packages at <25% coverage)
- Adapter tests (3 packages at ~22% coverage)
- Executing refactoring plan (guide ready, implementation pending)

## Next Steps

**Run these commands to start:**

```bash
# 1. Create tests for demo package
touch internal/adapters/nylas/demo/base_test.go

# 2. Create tests for tui2/state
touch internal/tui2/state/state_test.go

# 3. Create tests for tui2/utils
touch internal/tui2/utils/utils_test.go

# 4. Create tests for CLI commands
touch internal/cli/mcp/mcp_test.go
touch internal/cli/update/update_test.go
touch internal/cli/webhook/webhook_test.go

# 5. After creating tests, verify coverage
make test-coverage
```

---

**Target: 80%+ coverage by January 15, 2025**
**Ultimate Goal: 95%+ coverage (industry best practice)**
