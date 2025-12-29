# VHS Tape: Test Dashboard Screen
# This tape tests the main dashboard rendering and navigation

Output ../output/dashboard.png

Set Shell bash
Set FontSize 14
Set Width 1200
Set Height 800
Set Theme "Catppuccin Mocha"

# Launch the TUI and skip splash
Type "./bin/nylas tui --engine bubbletea"
Sleep 500ms
Enter
Sleep 500ms

# Capture clean dashboard
Screenshot ../output/dashboard-main.png

# Test theme cycling (press 't')
Type "t"
Sleep 300ms
Screenshot ../output/dashboard-theme-changed.png

# Cycle through a few more themes
Type "t"
Sleep 300ms
Type "t"
Sleep 300ms
Screenshot ../output/dashboard-theme-cycled.png

# Return to default theme (cycle back to k9s)
Type "t"
Sleep 300ms

# Capture final state
Screenshot ../output/dashboard-final.png

# Quit
Ctrl+C
Sleep 200ms
