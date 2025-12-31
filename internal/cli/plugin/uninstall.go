package plugin

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// newUninstallCmd creates the plugin uninstall command.
func newUninstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall <name>",
		Short: "Uninstall a plugin",
		Long: `Uninstall a plugin from ~/.nylas/plugins/.

This command only removes plugins installed to ~/.nylas/plugins/.
Plugins in system PATH must be removed manually.

Examples:
  # Uninstall a plugin
  nylas plugin uninstall air

  # Uninstall multiple plugins
  nylas plugin uninstall air tui slack
`,
		Args: cobra.MinimumNArgs(1),
		RunE: runUninstall,
	}
}

func runUninstall(cmd *cobra.Command, args []string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	pluginDir := filepath.Join(homeDir, ".nylas", "plugins")

	for _, name := range args {
		pluginName := pluginPrefix + name
		pluginPath := filepath.Join(pluginDir, pluginName)

		// Check if plugin exists
		if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
			fmt.Printf("⚠ Plugin %s not found in ~/.nylas/plugins/\n", name)
			fmt.Printf("  (It may be installed in PATH - remove manually)\n")
			continue
		}

		// Remove plugin
		if err := os.Remove(pluginPath); err != nil {
			fmt.Printf("✗ Failed to uninstall %s: %v\n", name, err)
			continue
		}

		fmt.Printf("✓ Plugin %s uninstalled successfully\n", name)
	}

	return nil
}
