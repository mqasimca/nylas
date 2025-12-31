package slack

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/mqasimca/nylas/internal/cli/common"
	"github.com/mqasimca/nylas/internal/domain"
)

func newChannelsCmd() *cobra.Command {
	var (
		channelTypes    []string
		excludeArchived bool
		limit           int
		showID          bool
	)

	cmd := &cobra.Command{
		Use:     "channels",
		Aliases: []string{"ch", "channel"},
		Short:   "List Slack channels",
		Long: `List accessible Slack channels, including public, private, DMs, and group DMs.

Examples:
  # List all channels
  nylas slack channels

  # List only public channels
  nylas slack channels --type public_channel

  # List channels with IDs
  nylas slack channels --id

  # Exclude archived channels
  nylas slack channels --exclude-archived`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getSlackClientFromKeyring()
			if err != nil {
				return common.NewUserError(
					"not authenticated with Slack",
					"Run: nylas slack auth set --token YOUR_TOKEN",
				)
			}

			ctx, cancel := createContext()
			defer cancel()

			params := &domain.SlackChannelQueryParams{
				Types:           channelTypes,
				ExcludeArchived: excludeArchived,
				Limit:           limit,
			}

			resp, err := client.ListChannels(ctx, params)
			if err != nil {
				return fmt.Errorf("failed to list channels: %w", err)
			}

			if len(resp.Channels) == 0 {
				fmt.Println("No channels found")
				return nil
			}

			printChannels(resp.Channels, showID)

			if resp.NextCursor != "" {
				dim := color.New(color.Faint)
				_, _ = dim.Printf("\n(more channels available)\n")
			}

			return nil
		},
	}

	cmd.Flags().StringSliceVar(&channelTypes, "type", nil, "Channel types: public_channel, private_channel, mpim, im")
	cmd.Flags().BoolVar(&excludeArchived, "exclude-archived", false, "Exclude archived channels")
	cmd.Flags().IntVarP(&limit, "limit", "l", 100, "Maximum number of channels to return")
	cmd.Flags().BoolVar(&showID, "id", false, "Show channel IDs")

	return cmd
}

func printChannels(channels []domain.SlackChannel, showID bool) {
	cyan := color.New(color.FgCyan)
	dim := color.New(color.Faint)
	yellow := color.New(color.FgYellow)

	for _, ch := range channels {
		name := ch.ChannelDisplayName()

		if ch.IsPrivate && !ch.IsIM && !ch.IsMPIM {
			_, _ = yellow.Print("🔒 ")
		}

		_, _ = cyan.Print(name)

		if showID {
			_, _ = dim.Printf(" [%s]", ch.ID)
		}

		if ch.MemberCount > 0 {
			_, _ = dim.Printf(" (%d members)", ch.MemberCount)
		}

		typeLabel := ch.ChannelType()
		if typeLabel != "public" {
			_, _ = dim.Printf(" [%s]", typeLabel)
		}

		if ch.IsArchived {
			_, _ = dim.Print(" (archived)")
		}

		fmt.Println()

		if ch.Purpose != "" {
			_, _ = dim.Printf("  %s\n", truncateString(ch.Purpose, 60))
		}
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
