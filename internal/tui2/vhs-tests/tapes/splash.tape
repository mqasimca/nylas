# VHS Tape: Test Splash Screen
# This tape tests the initial splash screen rendering

Output ../output/splash.png

Set Shell bash
Set FontSize 14
Set Width 1200
Set Height 800
Set Theme "Catppuccin Mocha"

# Launch the TUI
Type "./bin/nylas tui --engine bubbletea"
Sleep 500ms

# Capture splash screen
Screenshot ../output/splash-initial.png

# Wait for progress
Sleep 1s
Screenshot ../output/splash-progress.png

# Skip splash by pressing a key
Enter
Sleep 500ms

# Capture dashboard after splash
Screenshot ../output/dashboard-after-splash.png

# Quit
Ctrl+C
Sleep 200ms
