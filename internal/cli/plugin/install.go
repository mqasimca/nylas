package plugin

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// newInstallCmd creates the plugin install command.
func newInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install <name-or-url>",
		Short: "Install a plugin",
		Long: `Install a plugin from a URL or the official registry.

You can install plugins in two ways:

1. From a direct URL:
   nylas plugin install https://example.com/nylas-air

2. From the official registry (by name):
   nylas plugin install air

Plugins are installed to ~/.nylas/plugins/ and must be named "nylas-{name}".

Examples:
  # Install from registry
  nylas plugin install air

  # Install from URL
  nylas plugin install https://github.com/nylas/cli-plugin-air/releases/download/v1.0.0/nylas-air

  # Install from local file
  nylas plugin install ./nylas-air
`,
		Args: cobra.ExactArgs(1),
		RunE: runInstall,
	}
}

func runInstall(cmd *cobra.Command, args []string) error {
	nameOrURL := args[0]

	// Determine installation source
	var sourceURL string
	var pluginName string

	if strings.HasPrefix(nameOrURL, "http://") || strings.HasPrefix(nameOrURL, "https://") {
		// Direct URL
		sourceURL = nameOrURL
		// Extract name from URL (last path component)
		parts := strings.Split(nameOrURL, "/")
		pluginName = parts[len(parts)-1]
	} else if strings.HasPrefix(nameOrURL, "./") || strings.HasPrefix(nameOrURL, "/") {
		// Local file
		return installFromLocal(nameOrURL)
	} else {
		// Official registry
		return installFromRegistry(nameOrURL)
	}

	// Validate plugin name format
	if !strings.HasPrefix(pluginName, pluginPrefix) {
		return fmt.Errorf("plugin must be named %s{name}, got: %s", pluginPrefix, pluginName)
	}

	// Download and install
	fmt.Printf("Installing plugin %s...\n", pluginName)
	if err := downloadAndInstall(sourceURL, pluginName); err != nil {
		return err
	}

	fmt.Printf("✓ Plugin %s installed successfully\n", strings.TrimPrefix(pluginName, pluginPrefix))
	return nil
}

// downloadAndInstall downloads a plugin from a URL and installs it.
func downloadAndInstall(url, filename string) error {
	// Create plugins directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	pluginDir := filepath.Join(homeDir, ".nylas", "plugins")
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return err
	}

	// Download file
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download plugin: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}

	// Write to file
	destPath := filepath.Join(pluginDir, filename)
	f, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return err
	}

	// Make executable
	if err := os.Chmod(destPath, 0755); err != nil {
		return err
	}

	return nil
}

// installFromLocal installs a plugin from a local file.
func installFromLocal(localPath string) error {
	// Get absolute path
	absPath, err := filepath.Abs(localPath)
	if err != nil {
		return err
	}

	// Validate file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", localPath)
	}

	// Get filename
	filename := filepath.Base(absPath)
	if !strings.HasPrefix(filename, pluginPrefix) {
		return fmt.Errorf("plugin must be named %s{name}, got: %s", pluginPrefix, filename)
	}

	// Create plugins directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	pluginDir := filepath.Join(homeDir, ".nylas", "plugins")
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return err
	}

	// Copy file
	destPath := filepath.Join(pluginDir, filename)
	if err := copyFile(absPath, destPath); err != nil {
		return err
	}

	// Make executable
	if err := os.Chmod(destPath, 0755); err != nil {
		return err
	}

	fmt.Printf("✓ Plugin %s installed successfully\n", strings.TrimPrefix(filename, pluginPrefix))
	return nil
}

// installFromRegistry installs a plugin from the official registry.
func installFromRegistry(name string) error {
	// TODO: Implement registry support
	// For now, return a helpful error message
	return fmt.Errorf(`registry installation not yet implemented

To install plugins, use one of these methods:

1. Direct URL:
   nylas plugin install https://example.com/nylas-%s

2. Local file:
   nylas plugin install ./nylas-%s

3. Build from source:
   git clone https://github.com/nylas/cli-plugin-%s
   cd cli-plugin-%s
   go build -o nylas-%s
   nylas plugin install ./nylas-%s`, name, name, name, name, name, name)
}

// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	return nil
}
