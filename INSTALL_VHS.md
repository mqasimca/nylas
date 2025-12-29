# Installing VHS for Visual Testing

Quick guide to install VHS and start testing the TUI.

## 🚀 Installation

### Arch Linux (CachyOS)
```bash
sudo pacman -S vhs
```

### macOS
```bash
brew install vhs
```

### Other Linux Distributions

**Ubuntu/Debian:**
```bash
# Install dependencies
sudo apt install ffmpeg

# Install VHS via go
go install github.com/charmbracelet/vhs@latest
```

**Fedora:**
```bash
sudo dnf install vhs
```

**Manual Installation:**
```bash
# Download latest release from GitHub
wget https://github.com/charmbracelet/vhs/releases/latest/download/vhs_Linux_x86_64.tar.gz
tar -xzf vhs_Linux_x86_64.tar.gz
sudo mv vhs /usr/local/bin/
```

## ✅ Verify Installation

```bash
vhs --version
# Should output: vhs version 0.10.0 or similar
```

## 🧪 Run Your First Test

```bash
# 1. Build the binary
make build

# 2. Run a quick dashboard test
make test-vhs

# 3. View the output
ls -lh internal/tui2/vhs-tests/output/
```

You should see PNG screenshot files generated!

## 📸 What Gets Generated

After running `make test-vhs-all`, you'll have:

```
internal/tui2/vhs-tests/output/
├── splash-initial.png          # Splash screen on load
├── splash-progress.png         # Splash with progress bar
├── dashboard-after-splash.png  # Dashboard after skip
├── dashboard-main.png          # Clean dashboard
├── dashboard-theme-changed.png # After pressing 't'
├── dashboard-final.png         # Final state
├── navigation.gif              # Animated navigation demo
├── nav-*.png                   # Navigation screenshots
└── regression-*.png            # Regression test baselines
```

## 🎯 Next Steps

1. **Test the current UI**:
   ```bash
   make test-vhs-all
   ```

2. **Make your UI changes**:
   - Edit files in `internal/tui2/models/`
   - Run `make build`

3. **Compare before/after**:
   ```bash
   make test-vhs-all
   # Check the new screenshots vs old ones
   ```

4. **Commit baselines** (if changes look good):
   ```bash
   git add internal/tui2/vhs-tests/output/regression-*.png
   git commit -m "Update visual regression baselines"
   ```

## 📚 Documentation

- **VHS Testing Guide**: `internal/tui2/vhs-tests/README.md`
- **VHS Project**: https://github.com/charmbracelet/vhs
- **Tape File Examples**: `internal/tui2/vhs-tests/tapes/`

## 🐛 Troubleshooting

**Error: "vhs: command not found"**
- Make sure VHS is in your PATH
- Try: `which vhs`

**Screenshots are blank**
- Increase sleep times in tape files
- Ensure terminal has time to render

**ffmpeg not found**
- Install ffmpeg: `sudo pacman -S ffmpeg`

---

**Ready to test!** Run `make test-vhs` to generate your first screenshot.
