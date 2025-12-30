# VHS Tape: Test Dashboard Screen
# This tape tests the main dashboard rendering and navigation

Output dashboard.gif

Set Shell bash
Set FontFamily "MesloLGS Nerd Font Mono"
Set FontSize 14
Set Width 1200
Set Height 800
Set Theme "Catppuccin Mocha"

# Launch the TUI and skip splash
Type "./bin/nylas tui --engine bubbletea"
Enter
Sleep 2s

# Skip splash screen
Type " "
Sleep 1.5s

# Capture clean dashboard
Screenshot dashboard-main.png

# Test theme cycling (press 't')
Type "t"
Sleep 1s
Screenshot dashboard-theme-changed.png

# Cycle through a few more themes
Type "t"
Sleep 500ms
Type "t"
Sleep 1s
Screenshot dashboard-theme-cycled.png

# Return to default theme (cycle back to k9s)
Type "t"
Sleep 1s

# Capture final state
Screenshot dashboard-final.png

# Quit
Sleep 500ms
Ctrl+C
Sleep 500ms
