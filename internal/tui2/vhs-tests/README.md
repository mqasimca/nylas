# VHS Visual Testing for Nylas CLI TUI

This directory contains VHS (Video Home System) tape files for visual testing of the Bubble Tea TUI.

## 📼 What is VHS?

[VHS](https://github.com/charmbracelet/vhs) is a tool from Charmbracelet (creators of Bubble Tea) that allows you to:
- Record terminal sessions as GIFs, MP4s, or PNGs
- Write terminal interactions as code (`.tape` files)
- Generate visual regression tests
- Create documentation screenshots

## 🚀 Quick Start

### 1. Install VHS

```bash
# Arch Linux (recommended)
sudo pacman -S vhs

# macOS
brew install vhs

# Other systems - see https://github.com/charmbracelet/vhs
```

### 2. Run Tests

```bash
# Run dashboard test only
make test-vhs

# Run all visual tests
make test-vhs-all

# Clean test outputs
make test-vhs-clean
```

### 3. View Results

Screenshots and GIFs are saved to `internal/tui2/vhs-tests/output/`

## 📁 Directory Structure

```
internal/tui2/vhs-tests/
├── README.md              # This file
├── tapes/                 # VHS tape files (.tape)
│   ├── splash.tape        # Test splash screen rendering
│   ├── dashboard.tape     # Test dashboard screen
│   ├── navigation.tape    # Test navigation flow (generates GIF)
│   └── visual-regression.tape  # Reference screenshots
└── output/                # Generated screenshots and GIFs
    ├── *.png              # Screenshot files
    ├── *.gif              # Animation files
    └── *.txt              # Text output for golden file testing
```

## 🎬 Available Tape Files

### `splash.tape`
Tests the initial splash screen:
- Captures splash screen rendering
- Tests progress bar animation
- Verifies skip functionality
- **Output**: `splash-*.png`

### `dashboard.tape`
Tests the main dashboard:
- Captures clean dashboard render
- Tests theme cycling ('t' key)
- Verifies all UI elements
- **Output**: `dashboard-*.png`

### `navigation.tape`
Tests navigation between screens:
- Dashboard → Help → Settings → Debug
- Generates animated GIF showing full flow
- **Output**: `navigation.gif`, `nav-*.png`

### `visual-regression.tape`
Creates reference screenshots for regression testing:
- Fixed window size (1200x800)
- Text output for golden file comparison
- **Output**: `regression-*.png`, `regression-output.txt`

## ✍️ Writing Custom Tape Files

Create a new `.tape` file in the `tapes/` directory:

```tape
# My Custom Test
Output ../output/my-test.png

Set Shell bash
Set FontSize 14
Set Width 1200
Set Height 800
Set Theme "Catppuccin Mocha"

# Launch TUI
Type "./bin/nylas tui --engine bubbletea"
Sleep 500ms

# Skip splash
Enter
Sleep 500ms

# Take screenshot
Screenshot ../output/my-test.png

# Interact with TUI
Type "t"  # Cycle theme
Sleep 300ms

# Quit
Ctrl+C
```

### Available VHS Commands

- `Type "text"` - Type text
- `Enter` - Press Enter key
- `Escape` - Press Escape
- `Ctrl+C` - Send Ctrl+C
- `Sleep 500ms` - Wait
- `Screenshot file.png` - Capture screenshot
- `Set Width 1200` - Set terminal width
- `Set Height 800` - Set terminal height
- `Set Theme "name"` - Set color theme

See [VHS documentation](https://github.com/charmbracelet/vhs) for more commands.

## 🔍 Visual Regression Testing

### Generating Golden Files

1. Run tests to generate reference screenshots:
   ```bash
   make test-vhs-all
   ```

2. Commit the reference images to git:
   ```bash
   git add internal/tui2/vhs-tests/output/regression-*.png
   git commit -m "Add visual regression baseline"
   ```

### Checking for Regressions

After making UI changes:

1. Run tests again:
   ```bash
   make test-vhs-all
   ```

2. Compare new screenshots with committed baselines:
   ```bash
   # Use any image diff tool
   compare baseline.png new.png diff.png  # ImageMagick

   # Or visual inspection
   open internal/tui2/vhs-tests/output/regression-1200x800.png
   ```

3. If changes are intentional, update baselines:
   ```bash
   git add internal/tui2/vhs-tests/output/regression-*.png
   git commit -m "Update visual regression baselines"
   ```

## 🎨 Testing Different Themes

Modify the `Set Theme` line in tape files:

```tape
Set Theme "Catppuccin Mocha"   # Dark purple theme
Set Theme "Dracula"            # Dark theme
Set Theme "Nord"               # Cool blue theme
Set Theme "Tokyo Night"        # Purple-blue theme
```

## 🐛 Debugging Tips

### Tests Fail Silently
- Ensure binary is built: `make build`
- Check working directory in tape: should run from project root
- Verify output paths: `../output/` is relative to `tapes/` directory

### Screenshots Are Blank
- Increase `Sleep` durations for slow systems
- Check terminal size matches your screen
- Ensure TUI has time to fully render

### GIF Generation Fails
- Install ffmpeg: `sudo pacman -S ffmpeg`
- Check VHS has all dependencies: `vhs --version`

## 📊 CI/CD Integration

### GitHub Actions Example

```yaml
name: Visual Tests

on: [pull_request]

jobs:
  visual-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Install Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.24'

      - name: Install VHS
        run: |
          sudo apt-get update
          sudo apt-get install -y ttyd chromium-browser
          go install github.com/charmbracelet/vhs@latest

      - name: Run Visual Tests
        run: make test-vhs-all

      - name: Upload Screenshots
        uses: actions/upload-artifact@v3
        with:
          name: screenshots
          path: internal/tui2/vhs-tests/output/*.png
```

## 📝 Best Practices

1. **Keep tapes fast**: Use minimum sleep times needed
2. **Fixed dimensions**: Always set Width and Height for consistency
3. **Clean state**: Start each test from dashboard
4. **Descriptive names**: Use clear output file names
5. **Small outputs**: Prefer PNG for screenshots, GIF only for demos
6. **Version control**: Commit reference screenshots for regression testing

## 🔗 Resources

- [VHS GitHub Repository](https://github.com/charmbracelet/vhs)
- [VHS Documentation](https://github.com/charmbracelet/vhs/blob/main/README.md)
- [Bubble Tea Framework](https://github.com/charmbracelet/bubbletea)
- [Lip Gloss Styling](https://github.com/charmbracelet/lipgloss)

## 🆘 Troubleshooting

**Q: VHS says "command not found"**
A: Install VHS: `sudo pacman -S vhs` or see [installation guide](https://github.com/charmbracelet/vhs#installation)

**Q: Screenshots don't match my terminal**
A: VHS uses its own rendering. Use `Set Theme` to match your terminal theme.

**Q: Can I test in headless CI?**
A: Yes! VHS works in headless environments. Install ttyd and chromium dependencies.

**Q: How do I test different window sizes?**
A: Create separate tape files with different `Set Width` and `Set Height` values.

---

**Happy Testing! 🎉**
