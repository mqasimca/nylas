// channel_info.go provides the channel info subcommand for viewing channel details.

package slack

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/mqasimca/nylas/internal/cli/common"
)

// newChannelInfoCmd creates the info subcommand for getting channel details.
func newChannelInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info [channel-id]",
		Short: "Get detailed info about a channel",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getSlackClientFromKeyring()
			if err != nil {
				return common.NewUserError(
					"not authenticated with Slack",
					"Run: nylas slack auth set --token YOUR_TOKEN",
				)
			}

			ctx, cancel := common.CreateContext()
			defer cancel()

			channelID := args[0]
			ch, err := client.GetChannel(ctx, channelID)
			if err != nil {
				return fmt.Errorf("failed to get channel: %w", err)
			}

			cyan := color.New(color.FgCyan)
			dim := color.New(color.Faint)

			_, _ = cyan.Printf("Channel: #%s\n", ch.Name)
			fmt.Printf("  ID:           %s\n", ch.ID)
			fmt.Printf("  Is Channel:   %v\n", ch.IsChannel)
			fmt.Printf("  Is Private:   %v\n", ch.IsPrivate)
			fmt.Printf("  Is Archived:  %v\n", ch.IsArchived)
			fmt.Printf("  Is Member:    %v\n", ch.IsMember)
			fmt.Printf("  Is Shared:    %v\n", ch.IsShared)
			fmt.Printf("  Is OrgShared: %v\n", ch.IsOrgShared)
			fmt.Printf("  Is ExtShared: %v\n", ch.IsExtShared)
			fmt.Printf("  Is IM:        %v\n", ch.IsIM)
			fmt.Printf("  Is MPIM:      %v\n", ch.IsMPIM)
			fmt.Printf("  Is Group:     %v\n", ch.IsGroup)
			fmt.Printf("  Members:      %d\n", ch.MemberCount)
			if ch.Purpose != "" {
				_, _ = dim.Printf("  Purpose:      %s\n", ch.Purpose)
			}
			if ch.Topic != "" {
				_, _ = dim.Printf("  Topic:        %s\n", ch.Topic)
			}

			return nil
		},
	}

	return cmd
}
