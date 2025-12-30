# VHS Tape: Test Navigation Flow
# This tape tests navigating through different screens

Output navigation.gif

Set Shell bash
Set FontFamily "MesloLGS Nerd Font Mono"
Set FontSize 14
Set Width 1200
Set Height 800
Set Theme "Catppuccin Mocha"

# Launch and skip splash
Type "./bin/nylas tui --engine bubbletea"
Enter
Sleep 2s

# Skip splash
Type " "
Sleep 1.5s

# Start from dashboard
Screenshot nav-dashboard.png
Sleep 500ms

# Navigate to Help (press '?')
Type "?"
Sleep 1s
Screenshot nav-help.png
Sleep 500ms

# Go back to dashboard (press ESC)
Escape
Sleep 1s
Screenshot nav-back-to-dashboard.png

# Navigate to Settings (press 's')
Type "s"
Sleep 1s
Screenshot nav-settings.png
Sleep 500ms

# Back to dashboard
Escape
Sleep 1s

# Navigate to Debug (press 'd')
Type "d"
Sleep 1s
Screenshot nav-debug.png
Sleep 500ms

# Back to dashboard
Escape
Sleep 1s

# Final dashboard view
Screenshot nav-final.png

# Quit
Sleep 500ms
Ctrl+C
Sleep 500ms
