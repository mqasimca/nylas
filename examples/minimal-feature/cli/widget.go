package cli

import (
	"github.com/spf13/cobra"
)

// NewWidgetCmd creates the root widget command.
//
// CLI commands are:
// - Entry points for user interaction
// - Organized hierarchically (parent → children)
// - Thin layer that delegates to domain/adapter
func NewWidgetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "widget",
		Short: "Manage widgets",
		Long: `Manage widgets in your Nylas account.

Examples:
  # List all widgets
  nylas widget list

  # Create a new widget
  nylas widget create --name "My Widget" --description "A great widget"

  # Get widget details
  nylas widget show <widget-id>`,
	}

	// Add subcommands
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newCreateCmd())

	return cmd
}
