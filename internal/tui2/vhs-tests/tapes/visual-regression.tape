# VHS Tape: Visual Regression Test
# This tape captures reference screenshots for visual regression testing

Output visual-regression.txt

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

# Capture reference screenshots at different window sizes
Screenshot regression-1200x800.png

# Note: VHS doesn't support runtime window resizing in tape files
# You would need separate tape files for different sizes

# Capture text output for golden file testing
Output regression-output.txt

# Wait to ensure full render
Sleep 1s

# Quit
Ctrl+C
Sleep 500ms
