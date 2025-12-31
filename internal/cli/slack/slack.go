// Package slack provides CLI commands for Slack integration.
package slack

import (
	"github.com/spf13/cobra"
)

// NewSlackCmd creates the root slack command.
func NewSlackCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "slack",
		Aliases: []string{"sl"},
		Short:   "Interact with Slack workspaces",
		Long: `Interact with Slack workspaces - read and send messages as yourself.

Examples:
  # Authenticate with Slack
  nylas slack auth set --token xoxp-YOUR-TOKEN

  # Check authentication status
  nylas slack auth status

  # List channels
  nylas slack channels

  # Read messages from a channel
  nylas slack messages --channel general

  # Send a message
  nylas slack send --channel general --text "Hello!"

  # Reply to a thread
  nylas slack reply --channel general --thread 1234567890.123456 --text "Got it!"

  # Search messages
  nylas slack search --query "important"`,
	}

	cmd.AddCommand(newAuthCmd())
	cmd.AddCommand(newChannelsCmd())
	cmd.AddCommand(newMessagesCmd())
	cmd.AddCommand(newSendCmd())
	cmd.AddCommand(newReplyCmd())
	cmd.AddCommand(newUsersCmd())
	cmd.AddCommand(newSearchCmd())

	return cmd
}
