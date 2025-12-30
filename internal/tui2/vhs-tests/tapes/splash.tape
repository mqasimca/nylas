# VHS Tape: Test Splash Screen
# This tape tests the initial splash screen rendering

Output splash.gif

Set Shell bash
Set FontFamily "MesloLGS Nerd Font Mono"
Set FontSize 14
Set Width 1200
Set Height 800
Set Theme "Catppuccin Mocha"

# Launch the TUI
Type "./bin/nylas tui --engine bubbletea"
Enter
Sleep 2s

# Capture splash screen (should be visible now)
Screenshot splash-initial.png

# Wait for progress animation
Sleep 1.5s
Screenshot splash-progress.png

# Skip splash by pressing a key
Type " "
Sleep 1.5s

# Capture dashboard after splash
Screenshot dashboard-after-splash.png

# Quit
Sleep 500ms
Ctrl+C
Sleep 500ms
