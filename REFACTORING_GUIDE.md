# Refactoring Guide: File Splitting Patterns

**Purpose**: Document proven patterns for splitting large files into maintainable modules.

---

## Refactoring Philosophy

### The 500-Line Rule

**Target**: All files ≤500 lines (ideal), ≤600 lines (acceptable)

**Why**:
- Easier to understand and navigate
- Reduces cognitive load for AI assistants
- Improves Git diff readability
- Encourages better separation of concerns
- Faster Claude Code context loading

### When to Split

✅ **SPLIT when**:
- File exceeds 600 lines
- File has multiple distinct responsibilities
- Functions can be grouped by purpose
- Tests cover different feature areas

❌ **DON'T SPLIT when**:
- File is under 500 lines
- Functions are tightly coupled
- Split would create circular dependencies
- File represents a single cohesive unit

---

## Proven Split Patterns

### Pattern 1: Handler Split by Type

**Use Case**: Large HTTP handler files with CRUD + helpers

**Before** (985 lines):
\`\`\`
handlers_types_test.go
  ├── Helper function tests
  ├── Contact helper tests
  ├── Conflict detection tests
  ├── Time rounding tests
  ├── CSS styling tests
  └── Response converter tests
\`\`\`

**After** (5 files, avg 197 lines):
\`\`\`
handlers_types_base_test.go (221 lines)
  └── writeJSON, participantsToEmail tests

handlers_types_utilities_test.go (73 lines)
  └── containsEmail, matchesContactQuery tests

handlers_types_conflicts_test.go (236 lines)
  └── findConflicts, roundUpTo5Min tests

handlers_types_css_test.go (47 lines)
  └── EmailBodyCSS tests

handlers_types_responses_test.go (408 lines)
  └── draftToResponse, calendarToResponse, etc.
\`\`\`

**Key Principles**:
- Group by test category (helpers, conflicts, CSS, responses)
- Keep related tests together
- Split at function boundaries, never mid-function

### Pattern 2: Command Split by Action

**Use Case**: CLI commands with list/create/update/delete

**Before** (775 lines):
\`\`\`
contacts.go
  ├── Main command setup
  ├── List contacts
  ├── Show contact
  ├── Create contact
  ├── Update contact
  ├── Delete contact
  ├── Groups management
  └── Photo sync
\`\`\`

**After** (3 files):
\`\`\`
contacts_main.go (366 lines)
  ├── Main command
  ├── List, Show
  └── Create, Update, Delete

contacts_groups.go (180 lines)
  └── Group operations

contacts_photo_sync.go (256 lines)
  └── Photo sync functionality
\`\`\`

**Key Principles**:
- Keep main CRUD in one file if cohesive
- Extract distinct features (groups, photos)
- Maintain command hierarchy clarity

### Pattern 3: Test Split by Complexity

**Use Case**: Large test files with basic + advanced tests

**Before** (709 lines):
\`\`\`
messages_test.go
  ├── Basic message tests
  ├── Advanced scenarios
  └── Edge cases
\`\`\`

**After** (2 files):
\`\`\`
messages_test_basic.go (313 lines)
  └── Core functionality tests

messages_test_advanced.go (405 lines)
  └── Complex scenarios, edge cases
\`\`\`

**Key Principles**:
- Split between test functions, never within
- Group basic vs advanced
- Keep test helpers in basic file

### Pattern 4: Feature Split by Module

**Use Case**: Large adapters with multiple responsibilities

**Before** (776 lines):
\`\`\`
focus_optimizer.go
  ├── Analysis logic
  ├── Block detection
  └── Adaptive optimization
\`\`\`

**After** (3 files):
\`\`\`
focus_optimizer_analysis.go (395 lines)
  └── Core analysis algorithms

focus_optimizer_blocks.go (223 lines)
  └── Focus block detection

focus_optimizer_adaptive.go (182 lines)
  └── Adaptive optimization logic
\`\`\`

**Key Principles**:
- Split by responsibility (SRP)
- Keep types in analysis file
- Functions grouped by purpose

### Pattern 5: Template Split by Section

**Use Case**: Large HTML templates with distinct modals

**Before** (733 lines):
\`\`\`
modals.gohtml
  ├── Command palette
  ├── Search overlay
  ├── Compose modal
  ├── Event modal
  ├── Contact modal
  ├── Settings modal
  └── Snooze picker
\`\`\`

**After** (4 files):
\`\`\`
modals.gohtml (5 lines)
  {{template "modals_navigation" .}}
  {{template "modals_calendar" .}}
  {{template "modals_settings" .}}

modals_navigation.gohtml (290 lines)
  └── Command palette, search, shortcuts

modals_calendar.gohtml (165 lines)
  └── Event modal, snooze picker

modals_settings.gohtml (281 lines)
  └── Settings, notetaker config
\`\`\`

**Key Principles**:
- Main file includes sub-templates
- Split by UI section (navigation, calendar, settings)
- Each section is self-contained

---

## Naming Conventions

### Source Files

\`\`\`
# Types and core functionality
<feature>_<module>.go          # e.g., handlers_email_crud.go

# Specific functionality
<feature>_<action>.go          # e.g., server_lifecycle.go

# Helper functions
<feature>_helpers.go           # e.g., calendar_helpers.go
\`\`\`

### Test Files

\`\`\`
# Basic tests (MUST end with _test.go)
<feature>_test_basic.go        # e.g., email_test_basic.go
<feature>_basic_test.go        # Alternative: basic_test.go

# Advanced tests
<feature>_test_advanced.go     # e.g., email_test_advanced.go
<feature>_advanced_test.go     # Alternative

# Integration tests
integration_<feature>_test.go  # e.g., integration_email_test.go
<feature>_integration_test.go  # Alternative
\`\`\`

**CRITICAL**: Test files MUST end with \`_test.go\` for Go to recognize them!

---

## Step-by-Step Refactoring Process

### Phase 1: Analysis

\`\`\`bash
# 1. Check current size
wc -l <file>.go

# 2. List all functions
grep -n "^func" <file>.go

# 3. Identify logical groups
# Group functions by:
# - Responsibility (CRUD, helpers, types)
# - Feature area (search, validation, conversion)
# - Test category (basic, advanced, integration)
\`\`\`

### Phase 2: Planning

\`\`\`
# Create split plan
<file>.go (794 lines)
  ├── Group 1: Types + Core (lines 1-250)
  ├── Group 2: CRUD Ops (lines 251-500)
  ├── Group 3: Helpers (lines 501-650)
  └── Group 4: Advanced (lines 651-794)

→ 4 files, ~200 lines each
\`\`\`

### Phase 3: Execution

\`\`\`bash
# 1. Create first split file
sed -n '1,250p' <file>.go > <file>_types.go

# 2. Create subsequent files
sed -n '1,50p' <file>.go > <file>_crud.go   # Copy header
sed -n '251,500p' <file>.go >> <file>_crud.go  # Add content

# 3. Repeat for all groups

# 4. Fix imports
goimports -w <file>_*.go

# 5. Verify build
make build

# 6. Remove original
git rm <file>.go
\`\`\`

### Phase 4: Verification

\`\`\`bash
# 1. Build check
make build

# 2. Test check
make test-unit

# 3. Lint check
golangci-lint run

# 4. Integration test
make test-integration
\`\`\`

---

## Common Pitfalls & Solutions

### Pitfall 1: Split Mid-Function

**Problem**: Split at arbitrary line number, breaks function

\`\`\`go
// ❌ BAD: Split here causes EOF error
func MyFunction() {
    // ... 100 lines
    if condition {
        // <-- Split point causes "expected '}', found 'EOF'"
\`\`\`

**Solution**: Always split at function boundaries

\`\`\`bash
# Find safe split points
grep -n "^func" <file>.go

# Choose line AFTER complete function
sed -n '1,420p'  # Line 420 is AFTER TestFunction ends
\`\`\`

### Pitfall 2: Duplicate Helper Functions

**Problem**: Copy helper to all split files

**Solution**: Keep helpers in ONE file (usually the base/core file)

\`\`\`go
// ✅ GOOD: Helper in base file only
// contacts_test_crud.go
func executeCommand(...) { ... }  # <-- Keep here

// contacts_test_groups.go
// (No helper function - uses one from crud file)
\`\`\`

### Pitfall 3: Wrong Package Name

**Problem**: Test file uses wrong package

\`\`\`go
// ❌ BAD: External test but wrong package
package scheduler_test  // File: scheduler_config_test.go

// ❌ BAD: Internal test but missing _test suffix
package scheduler  // File: scheduler_test.go (should be scheduler_test)
\`\`\`

**Solution**: Match package convention

\`\`\`go
// ✅ GOOD: External test file
package nylas_test  // File: scheduler_config_test.go

// ✅ GOOD: Internal test file
package scheduler  // File: scheduler_config_test.go
\`\`\`

### Pitfall 4: Missing Test File Suffix

**Problem**: Test file doesn't end with \`_test.go\`

\`\`\`bash
# ❌ BAD: Won't be recognized by Go
scheduler_test_config.go

# ✅ GOOD: Proper test file name
scheduler_config_test.go
\`\`\`

**Solution**: Always use \`*_test.go\` suffix!

---

## Real Examples from This Repo

### Example 1: calendar_test.go → 4 files

**Original**: 794 lines, all tests in one file

**Result**: 4 focused files
- \`calendar_cmd_test.go\` (223) - Command tests
- \`calendar_events_test.go\` (264) - Event tests
- \`calendar_availability_test.go\` (141) - Availability tests
- \`calendar_helpers_test.go\` (241) - Helper functions

**Benefit**: Each file has single responsibility, easier to find tests

### Example 2: handlers_types_test.go → 5 files

**Original**: 985 lines, 50+ test functions

**Result**: 5 categorized files
- \`handlers_types_base_test.go\` (221) - Basic helpers
- \`handlers_types_utilities_test.go\` (73) - Utility functions
- \`handlers_types_conflicts_test.go\` (236) - Conflict detection
- \`handlers_types_css_test.go\` (47) - CSS tests
- \`handlers_types_responses_test.go\` (408) - Response converters

**Benefit**: Tests grouped by feature, clearer organization

### Example 3: app.go → 5 files

**Original**: 790 lines, complex TUI application

**Result**: 5 responsibility-based files
- \`app_base.go\` (80) - Types and constructor
- \`app_init.go\` (209) - Initialization logic
- \`app_ui.go\` (292) - UI and navigation
- \`app_control.go\` (108) - App control methods
- \`app_commands.go\` (149) - Page movement commands

**Benefit**: Clear separation, easier to maintain

---

## Metrics & Validation

### Success Criteria

✅ **Successful Split**:
- All files ≤500 lines (ideal) or ≤600 lines (acceptable)
- Build succeeds: \`make build\`
- Tests pass: \`make test-unit\`
- Linting clean: \`golangci-lint run\`
- No duplicate functions
- Clear file naming

❌ **Failed Split (Needs Revision)**:
- Build errors (EOF, redeclaration)
- Test failures
- Files still >600 lines
- Circular dependencies
- Unclear file purposes

### Historical Stats

**Refactoring Wave 1-5 (Dec 2024-Jan 2025)**:
- Files refactored: 35
- Original total: 26,142 lines
- Result: 89 files, same functionality
- Average file size: 294 lines (was 747)
- Build success rate: 100%

---

## Quick Reference Card

\`\`\`
┌─────────────────────────────────────────────────────────┐
│  REFACTORING QUICK STEPS                                │
├─────────────────────────────────────────────────────────┤
│  1. Check size: wc -l <file>.go                         │
│  2. List funcs: grep -n "^func" <file>.go               │
│  3. Group by responsibility                             │
│  4. Split at function boundaries                        │
│  5. Create files with proper naming                     │
│  6. Fix imports: goimports -w *.go                      │
│  7. Verify: make build && make test-unit                │
│  8. Commit with descriptive message                     │
└─────────────────────────────────────────────────────────┘

REMEMBER:
- Split at function boundaries, NEVER mid-function
- Test files MUST end with _test.go
- Use goimports to fix imports
- Verify with make build before committing
- Keep related functionality together
\`\`\`

---

**Last Updated**: 2025-12-29
**Patterns Validated**: 35 file refactorings, 100% success rate
**Next Target**: 20 files still 590-626 lines
