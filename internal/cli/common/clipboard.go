package common

import (
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/atotto/clipboard"
)

// CopyToClipboard copies text to the system clipboard.
// On Linux, it supports both X11 (xclip/xsel) and Wayland (wl-copy).
// On macOS and Windows, it uses the native clipboard.
func CopyToClipboard(text string) error {
	// On Linux, check for Wayland first
	if runtime.GOOS == "linux" {
		// Check if running on Wayland
		if os.Getenv("WAYLAND_DISPLAY") != "" || os.Getenv("XDG_SESSION_TYPE") == "wayland" {
			// Try wl-copy (Wayland)
			if path, err := exec.LookPath("wl-copy"); err == nil {
				// #nosec G204 -- path is from exec.LookPath("wl-copy"), not user input
				cmd := exec.Command(path)
				cmd.Stdin = strings.NewReader(text)
				if err := cmd.Run(); err == nil {
					return nil
				}
			}
		}

		// Try xclip (X11)
		if path, err := exec.LookPath("xclip"); err == nil {
			// #nosec G204 -- path is from exec.LookPath("xclip"), not user input
			cmd := exec.Command(path, "-selection", "clipboard")
			cmd.Stdin = strings.NewReader(text)
			if err := cmd.Run(); err == nil {
				return nil
			}
		}

		// Try xsel (X11)
		if path, err := exec.LookPath("xsel"); err == nil {
			// #nosec G204 -- path is from exec.LookPath("xsel"), not user input
			cmd := exec.Command(path, "--clipboard", "--input")
			cmd.Stdin = strings.NewReader(text)
			if err := cmd.Run(); err == nil {
				return nil
			}
		}
	}

	// Fall back to atotto/clipboard (works on macOS, Windows, and X11 with xclip/xsel)
	return clipboard.WriteAll(text)
}
