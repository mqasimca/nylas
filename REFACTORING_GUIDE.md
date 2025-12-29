# Refactoring Guide - Files Over 500 Lines

**Status:** 48 production files need refactoring
**Goal:** No file should exceed 500 lines
**Progress:** Test coverage improvements completed (4 packages: 0% → 95% average)

---

## Priority 1: Critical Files (>1,000 lines)

### 1. internal/tui/views.go (2,619 lines) 🔴 CRITICAL

**Current Structure:**
- Lines 1-30: Package, imports, ResourceView interface
- Lines 31-63: BaseTableView (33 lines)
- Lines 64-151: DashboardView (88 lines)
- Lines 152-790: MessagesView (638 lines) ⚠️
- Lines 791-1656: EventsView (865 lines) ⚠️
- Lines 1657-1879: ContactsView (222 lines)
- Lines 1880-2086: WebhooksView (206 lines)
- Lines 2087-2206: GrantsView (119 lines)
- Lines 2207-2619: InboundView (412 lines)

**Refactoring Plan:**
```bash
# Split into 9 files:
internal/tui/views_base.go        # ResourceView interface + BaseTableView (50 lines)
internal/tui/views_dashboard.go   # DashboardView (100 lines)
internal/tui/views_messages.go    # MessagesView (650 lines) - May need further split
internal/tui/views_events.go      # EventsView (870 lines) - May need further split
internal/tui/views_contacts.go    # ContactsView (230 lines)
internal/tui/views_webhooks.go    # WebhooksView (210 lines)
internal/tui/views_grants.go      # GrantsView (125 lines)
internal/tui/views_inbound.go     # InboundView (420 lines)
```

**Further Splits Needed:**
- **MessagesView** (650 lines) → Split into:
  - `views_messages.go` - Main view struct and Load() (300 lines)
  - `views_messages_detail.go` - Detail view rendering (200 lines)
  - `views_messages_actions.go` - Star, unread, compose actions (150 lines)

- **EventsView** (870 lines) → Split into:
  - `views_events.go` - Main view struct and Load() (400 lines)
  - `views_events_detail.go` - Event detail dialog (250 lines)
  - `views_events_recurring.go` - Recurring event handling (220 lines)

**Commands:**
```bash
# Extract BaseTableView
sed -n '1,63p' internal/tui/views.go > internal/tui/views_base.go

# Extract DashboardView
echo 'package tui

import (
	"context"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/rivo/tview"
)' > internal/tui/views_dashboard.go
sed -n '59,151p' internal/tui/views.go >> internal/tui/views_dashboard.go

# Continue for each view...
```

---

### 2. internal/adapters/nylas/demo.go (1,623 lines) 🔴 CRITICAL

**Current Structure:**
- DemoClient with 125 methods
- Mock data generation for all Nylas resources
- Realistic demo data (messages, events, contacts, calendars, etc.)

**Refactoring Plan:**
```bash
# Split by resource type:
internal/adapters/nylas/demo/client.go      # Main DemoClient struct (50 lines)
internal/adapters/nylas/demo/grants.go      # Grant methods (100 lines)
internal/adapters/nylas/demo/messages.go    # Message methods + data (300 lines)
internal/adapters/nylas/demo/threads.go     # Thread methods + data (200 lines)
internal/adapters/nylas/demo/drafts.go      # Draft methods (100 lines)
internal/adapters/nylas/demo/folders.go     # Folder methods (80 lines)
internal/adapters/nylas/demo/calendars.go   # Calendar methods + data (250 lines)
internal/adapters/nylas/demo/events.go      # Event methods + data (300 lines)
internal/adapters/nylas/demo/contacts.go    # Contact methods + data (150 lines)
internal/adapters/nylas/demo/webhooks.go    # Webhook methods (100 lines)
```

**Note:** demo/base.go already exists with minimal structure. Migrate full implementation there.

---

### 3. internal/adapters/nylas/mock.go (1,459 lines) 🔴 CRITICAL

**Similar structure to demo.go - split by resource type**

```bash
# Split into:
internal/adapters/nylas/mock/client.go
internal/adapters/nylas/mock/messages.go
internal/adapters/nylas/mock/calendars.go
internal/adapters/nylas/mock/contacts.go
# etc...
```

---

### 4. internal/cli/ui/server.go (1,286 lines) 🔴 CRITICAL

**Already partially refactored!** ✅

Current structure shows good modular design:
- `server.go` (51 lines) - Server struct
- `server_lifecycle.go` (315 lines) - Init, routing, lifecycle
- `server_stores.go` (67 lines) - Cache accessors
- `server_sync.go` (187 lines) - Background sync
- `server_offline.go` (98 lines) - Offline queue
- `server_converters.go` (116 lines) - Domain conversions
- `server_template.go` (163 lines) - Template handling

**Remaining work:**
```bash
# The original server.go (1,286 lines) has already been split!
# Just verify all pieces are < 500 lines ✅

# Largest remaining: server_lifecycle.go (315 lines) - OK
```

---

## Priority 2: Large Files (700-1,000 lines)

### 5. internal/tui2/models/compose.go (1,162 lines)

**Split into:**
```bash
internal/tui2/models/compose.go           # Main Model struct (300 lines)
internal/tui2/models/compose_view.go      # View rendering (400 lines)
internal/tui2/models/compose_update.go    # Update handlers (300 lines)
internal/tui2/models/compose_commands.go  # Tea.Cmd functions (162 lines)
```

### 6. internal/cli/calendar/events.go (1,125 lines)

**Split into:**
```bash
internal/cli/calendar/events.go           # Main command (200 lines)
internal/cli/calendar/events_create.go    # Create event (300 lines)
internal/cli/calendar/events_update.go    # Update event (300 lines)
internal/cli/calendar/events_list.go      # List events (200 lines)
internal/cli/calendar/events_helpers.go   # Helpers (125 lines)
```

### 7. internal/tui2/components/calendar_grid.go (1,060 lines)

**Split into:**
```bash
internal/tui2/components/calendar_grid.go         # Main component (300 lines)
internal/tui2/components/calendar_grid_view.go    # Rendering (400 lines)
internal/tui2/components/calendar_grid_update.go  # Update logic (360 lines)
```

### 8. internal/tui2/models/calendar.go (1,058 lines)

**Split into:**
```bash
internal/tui2/models/calendar.go        # Main Model (300 lines)
internal/tui2/models/calendar_view.go   # View rendering (400 lines)
internal/tui2/models/calendar_update.go # Update handlers (358 lines)
```

### 9. internal/cli/demo/email.go (1,004 lines)

**Split into:**
```bash
internal/cli/demo/email.go       # Main command (200 lines)
internal/cli/demo/email_data.go  # Demo data generation (500 lines)
internal/cli/demo/email_send.go  # Send operations (304 lines)
```

---

## Priority 3: Medium Files (500-700 lines)

**44 files total** - Each needs splitting into 2-3 files

### Example: internal/tui2/models/messages.go (953 lines)

```bash
# Split into:
internal/tui2/models/messages.go        # Main Model (350 lines)
internal/tui2/models/messages_view.go   # View rendering (350 lines)
internal/tui2/models/messages_update.go # Update handlers (253 lines)
```

---

## Refactoring Methodology

### Step-by-Step Process:

1. **Analyze Structure**
   ```bash
   # Find type definitions and functions
   grep -n "^type\|^func" file.go

   # Count lines per section
   wc -l file.go
   ```

2. **Create Split Plan**
   - Group by responsibility (CRUD operations, data models, handlers, etc.)
   - Aim for 300-400 lines per file
   - Keep related code together

3. **Extract Files**
   ```bash
   # Method 1: Manual extraction (safest)
   # - Copy package declaration and imports
   # - Copy relevant types and functions
   # - Add necessary imports

   # Method 2: sed extraction
   sed -n 'START,ENDp' original.go > new.go
   ```

4. **Update Imports**
   ```bash
   # Auto-fix imports
   goimports -w internal/path/*.go
   ```

5. **Run Tests**
   ```bash
   # After each file split
   go test ./internal/path/...

   # Full CI
   make ci
   ```

6. **Delete Original**
   ```bash
   # Only after all tests pass
   git rm original.go
   ```

---

## Tools to Help

### 1. AST-Based Refactoring (Recommended)
```bash
# Use gomvpkg or gofmt-based tools
# These understand Go syntax and won't break code
```

### 2. Manual Splitting
```bash
# Safest approach:
1. Create new file
2. Copy package + imports
3. Copy relevant code
4. Run goimports -w
5. Run tests
6. Repeat
```

### 3. Automated Script
```bash
#!/bin/bash
# split_views.sh - Example for views.go

cat > internal/tui/views_base.go << 'EOF'
package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ResourceView interface for all views.
type ResourceView interface {
	Name() string
	Title() string
	Primitive() tview.Primitive
	Hints() []Hint
	Load()
	Refresh()
	Filter(string)
	HandleKey(*tcell.EventKey) *tcell.EventKey
}

// BaseTableView provides common table view functionality.
type BaseTableView struct {
	app    *App
	table  *Table
	name   string
	title  string
	hints  []Hint
	filter string
}
// ... rest of BaseTableView methods
EOF

# Run imports fixer
goimports -w internal/tui/views_base.go

# Test
go test ./internal/tui
```

---

## Testing Strategy

After each refactoring:

```bash
# 1. Build check
go build ./internal/tui

# 2. Unit tests
go test ./internal/tui -v

# 3. Integration tests (if applicable)
go test ./internal/tui -tags=integration

# 4. Full CI
make ci
```

---

## Progress Tracking

| File | Lines | Status | Split Into | Test Status |
|------|-------|--------|------------|-------------|
| internal/tui/views.go | 2,619 | ⏳ Planned | 9 files | ⏳ |
| internal/adapters/nylas/demo.go | 1,623 | ⏳ Planned | 10 files | ⏳ |
| internal/adapters/nylas/mock.go | 1,459 | ⏳ Planned | 10 files | ⏳ |
| internal/cli/ui/server.go | 1,286 | ✅ DONE | Already split | ✅ |
| internal/tui2/models/compose.go | 1,162 | ⏳ Planned | 4 files | ⏳ |
| ... 43 more files | 500-1,000 | ⏳ Planned | 2-3 each | ⏳ |

**Total:** 48 files → ~150 files (estimated)

---

## Completion Checklist

For each file refactoring:

- [ ] Analyze structure and create split plan
- [ ] Create new files with proper package/imports
- [ ] Move code to new files
- [ ] Run `goimports -w` on all new files
- [ ] Run unit tests
- [ ] Run integration tests (if applicable)
- [ ] Run `make ci`
- [ ] Delete original file (after tests pass)
- [ ] Update REFACTORING_GUIDE.md progress
- [ ] Commit changes

---

## Estimated Effort

| Priority | Files | Estimated Time |
|----------|-------|----------------|
| Priority 1 (>1,000 lines) | 9 files | 8-12 hours |
| Priority 2 (700-1,000 lines) | 5 files | 4-6 hours |
| Priority 3 (500-700 lines) | 34 files | 12-16 hours |
| **Total** | **48 files** | **24-34 hours** |

**Recommendation:** Tackle files in priority order, running full tests after each file.

---

## Safety Notes

⚠️ **Important:**
1. **Never skip tests** - Run tests after each file split
2. **Use version control** - Commit working state frequently
3. **One file at a time** - Don't refactor multiple files simultaneously
4. **Keep backups** - Don't delete original until tests pass
5. **Verify imports** - Use `goimports` to fix import issues automatically

---

**Last Updated:** December 29, 2024
**Next Action:** Start with `internal/tui/views.go` (highest priority)
