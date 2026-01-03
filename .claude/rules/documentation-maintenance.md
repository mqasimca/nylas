# Documentation Maintenance Rule

**CRITICAL**: Always update documentation when making code changes.

---

## Documentation Update Matrix

| Change Type | Update Files | Priority |
|-------------|--------------|----------|
| **New CLI command** | CLAUDE.md, docs/COMMANDS.md, cmd/nylas/main.go | 🔴 CRITICAL |
| **New integration test** | CLAUDE.md, docs/DEVELOPMENT.md | 🔴 CRITICAL |
| **New adapter/API method** | CLAUDE.md, docs/ARCHITECTURE.md (if new file) | 🟡 IF NEEDED |
| **New domain model** | CLAUDE.md, docs/ARCHITECTURE.md (if major) | 🟡 IF NEEDED |
| **Test structure change** | CLAUDE.md, docs/DEVELOPMENT.md, .claude/rules/testing.md | 🔴 CRITICAL |
| **New skill/workflow** | CLAUDE.md (if user-facing) | 🟡 IF NEEDED |
| **Security change** | docs/security/overview.md | 🔴 CRITICAL |
| **Architecture change** | docs/ARCHITECTURE.md, CLAUDE.md | 🔴 CRITICAL |
| **Utility feature** | CLAUDE.md, docs/COMMANDS.md | 🔴 CRITICAL |
| **Timezone change** | docs/commands/timezone.md, docs/COMMANDS.md, CLAUDE.md | 🔴 CRITICAL |
| **Working hours/breaks** | docs/commands/timezone.md, docs/ARCHITECTURE.md, CLAUDE.md | 🔴 CRITICAL |

---

## Timezone & Working Hours Changes ⚠️ CRITICAL

**Always update `docs/commands/timezone.md` when modifying:**
- `internal/cli/calendar/helpers.go` (timezone conversion)
- `internal/cli/calendar/events.go` (--timezone, --show-tz flags)
- `internal/adapters/utilities/timezone/service.go` (timezone service)
- `internal/domain/config.go` (WorkingHoursConfig, DaySchedule, BreakBlock)
- DST detection, natural language parsing, timezone validation
- Working hours validation, break block enforcement

**Update must include:**
- New features/flags with examples
- Changed behavior with before/after
- Best practices if applicable
- Troubleshooting for common issues

**Reason:** Timezone handling is complex. Users need clear, accurate docs.

---

## Quick Reference Checklist

**Before marking task complete:**

### For New Features:
- [ ] Updated CLAUDE.md file structure table
- [ ] Updated docs/COMMANDS.md with examples
- [ ] Updated README.md (if major feature)

### For New Tests:
- [ ] Updated CLAUDE.md test paths
- [ ] Updated docs/DEVELOPMENT.md test list

### For Structural Changes:
- [ ] Updated ALL affected docs
- [ ] Verified no old references remain
- [ ] Updated .claude/ rules if needed

### For Timezone/Calendar:
- [ ] Updated docs/commands/timezone.md
- [ ] Updated docs/COMMANDS.md calendar section
- [ ] Verified examples work

---

## Golden Rule

**If you changed code → Update docs**

No exceptions.

---

**Files to Never Reference:**
- ❌ `local/*.md` - Temporary/historical docs (excluded from context)
- ❌ `local/suggestions.md` - Feature proposals only
- ❌ `local/SECURITY_REPORT.md` - Historical report

**Quick verification:**
```bash
# After structural changes, verify no stale references:
grep -r "old-pattern" docs/ .claude/ *.md
```

---

**Last Updated:** January 3, 2025
