package plugin

import (
	"github.com/spf13/cobra"
)

// NewPluginCmd creates the plugin management command.
func NewPluginCmd(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugin",
		Short: "Manage Nylas CLI plugins",
		Long: `Manage Nylas CLI plugins.

Plugins extend the Nylas CLI with additional functionality. Plugins are
standalone executables named "nylas-{name}" that can be discovered in your
system PATH or installed to ~/.nylas/plugins/.

Examples:
  # List all available plugins
  nylas plugin list

  # Install a plugin from a URL
  nylas plugin install https://github.com/nylas/cli-plugin-air/releases/download/v1.0.0/nylas-air

  # Install a plugin from the official registry
  nylas plugin install air

  # Uninstall a plugin
  nylas plugin uninstall air

  # Use a plugin
  nylas air serve
  nylas tui --engine bubbletea
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	// Add subcommands
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newInstallCmd())
	cmd.AddCommand(newUninstallCmd())

	return cmd
}
