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

  # List your channels
  nylas slack channels list

  # Get channel info
  nylas slack channels info C1234567890

  # Read messages from a channel
  nylas slack messages list --channel general

  # List workspace users
  nylas slack users list

  # Send a message
  nylas slack send --channel general --text "Hello!"

  # Reply to a thread
  nylas slack reply --channel general --thread 1234567890.123456 --text "Got it!"

  # Search messages
  nylas slack search --query "important"

  # List files in a channel
  nylas slack files list --channel general

  # Download a file
  nylas slack files download F1234567890`,
	}

	cmd.AddCommand(newAuthCmd())
	cmd.AddCommand(newChannelsCmd())
	cmd.AddCommand(newMessagesCmd())
	cmd.AddCommand(newFilesCmd())
	cmd.AddCommand(newSendCmd())
	cmd.AddCommand(newReplyCmd())
	cmd.AddCommand(newUsersCmd())
	cmd.AddCommand(newSearchCmd())

	return cmd
}
