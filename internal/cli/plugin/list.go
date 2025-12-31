package plugin

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

// newListCmd creates the plugin list command.
func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all available plugins",
		Long: `List all available plugins discovered in PATH and ~/.nylas/plugins/.

Plugins are standalone executables named "nylas-{name}" that extend the
Nylas CLI with additional functionality.

Examples:
  # List all plugins
  nylas plugin list
`,
		RunE: runList,
	}
}

func runList(cmd *cobra.Command, args []string) error {
	// Discover all plugins
	plugins, err := Discover()
	if err != nil {
		return fmt.Errorf("failed to discover plugins: %w", err)
	}

	if len(plugins) == 0 {
		fmt.Println("No plugins found.")
		fmt.Println()
		fmt.Println("To install a plugin:")
		fmt.Println("  nylas plugin install <name>")
		fmt.Println()
		fmt.Println("Available official plugins:")
		fmt.Println("  air     - Web UI for Nylas CLI")
		fmt.Println("  tui     - Terminal UI with Bubble Tea")
		fmt.Println("  slack   - Slack integration")
		fmt.Println("  mcp     - MCP server for AI assistants")
		return nil
	}

	// Display plugins in a table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tPATH")
	fmt.Fprintln(w, "----\t----")

	for _, plugin := range plugins {
		fmt.Fprintf(w, "%s\t%s\n", plugin.Name, plugin.Path)
	}

	return w.Flush()
}
