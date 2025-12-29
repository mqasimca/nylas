# VHS Tape: Test Navigation Flow
# This tape tests navigating through different screens

Output ../output/navigation.gif

Set Shell bash
Set FontSize 14
Set Width 1200
Set Height 800
Set Theme "Catppuccin Mocha"

# Launch and skip splash
Type "./bin/nylas tui --engine bubbletea"
Sleep 500ms
Enter
Sleep 500ms

# Start from dashboard
Screenshot ../output/nav-dashboard.png
Sleep 500ms

# Navigate to Help (press '?')
Type "?"
Sleep 500ms
Screenshot ../output/nav-help.png
Sleep 1s

# Go back to dashboard (press ESC or 'q')
Escape
Sleep 500ms
Screenshot ../output/nav-back-to-dashboard.png

# Navigate to Settings (press 's')
Type "s"
Sleep 500ms
Screenshot ../output/nav-settings.png
Sleep 1s

# Back to dashboard
Escape
Sleep 500ms

# Navigate to Debug (press 'd')
Type "d"
Sleep 500ms
Screenshot ../output/nav-debug.png
Sleep 1s

# Back to dashboard
Escape
Sleep 500ms

# Final dashboard view
Screenshot ../output/nav-final.png

# Quit
Ctrl+C
Sleep 200ms
