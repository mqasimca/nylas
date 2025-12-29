#!/bin/bash
# Automated keyboard test verification script for TUI2

set -e

echo "================================"
echo "TUI2 Keyboard Tests"
echo "================================"
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counter
PASSED=0
FAILED=0

run_test() {
    local test_name="$1"
    local test_pattern="$2"

    echo -n "Testing $test_name... "

    if go test ./internal/tui2/models/... -run "$test_pattern" -v > /dev/null 2>&1; then
        echo -e "${GREEN}✓ PASS${NC}"
        ((PASSED++))
    else
        echo -e "${RED}✗ FAIL${NC}"
        ((FAILED++))
    fi
}

echo -e "${BLUE}Running automated keyboard tests...${NC}"
echo ""

# Dashboard tests
run_test "Dashboard: Press 'a' navigates to messages" "TestDashboard_KeyboardShortcuts/press_'a'"
run_test "Dashboard: Press 'c' navigates to calendar" "TestDashboard_KeyboardShortcuts/press_'c'"
run_test "Dashboard: Press 'p' navigates to contacts" "TestDashboard_KeyboardShortcuts/press_'p'"
run_test "Dashboard: Press 'd' navigates to debug" "TestDashboard_KeyboardShortcuts/press_'d'"
run_test "Dashboard: Press 's' navigates to settings" "TestDashboard_KeyboardShortcuts/press_'s'"
run_test "Dashboard: Press '?' navigates to help" "TestDashboard_KeyboardShortcuts/press_'\\?'"
run_test "Dashboard: Press 't' cycles theme" "TestDashboard_ThemeCycling"

# Help screen tests
run_test "Help: Press 'esc' navigates back" "TestHelp_KeyboardShortcuts/esc"
run_test "Help: Press 'q' navigates back" "TestHelp_KeyboardShortcuts/q"
run_test "Help: Press 'ctrl+c' quits" "TestHelp_KeyboardShortcuts/ctrl"

# Splash screen tests
run_test "Splash: Press any key skips" "TestSplash_SkipWithKeyPress"

# Calendar tests
run_test "Calendar: View mode navigation" "TestCalendarScreen_Update_KeyNavigation"
run_test "Calendar: Press 'esc' goes back" "TestCalendarScreen_Update_EscapeGoesBack"
run_test "Calendar: Press 'ctrl+c' quits" "TestCalendarScreen_Update_CtrlCQuits"
run_test "Calendar: Press 't' goes to today" "TestCalendarScreen_Update_TodayKey"
run_test "Calendar: Press 'r' refreshes" "TestCalendarScreen_Update_RefreshKey"

# Message detail tests
run_test "MessageDetail: Press 'esc' goes back" "TestMessageDetail_UpdateWithKeyPress/esc"
run_test "MessageDetail: Press 'r' replies" "TestMessageDetail_UpdateWithKeyPress/r_key"
run_test "MessageDetail: Press 'a' replies all" "TestMessageDetail_UpdateWithKeyPress/a_key"
run_test "MessageDetail: Press 'f' forwards" "TestMessageDetail_UpdateWithKeyPress/f_key"

# Settings tests
run_test "Settings: Theme cycling" "TestSettingsCycleTheme"
run_test "Settings: Toggle settings" "TestSettingsToggleSetting"

echo ""
echo "================================"
echo -e "Results: ${GREEN}$PASSED passed${NC}, ${RED}$FAILED failed${NC}"
echo "================================"

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}All keyboard tests PASSED! ✓${NC}"
    echo ""
    echo "Manual test: Run ./bin/nylas tui --engine bubbletea"
    echo "See internal/tui2/TESTING.md for complete test checklist"
    exit 0
else
    echo -e "${RED}Some tests FAILED! ✗${NC}"
    exit 1
fi
