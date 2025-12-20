package notetaker

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func newListCmd() *cobra.Command {
	var (
		limit      int
		state      string
		outputJSON bool
	)

	cmd := &cobra.Command{
		Use:     "list [grant-id]",
		Aliases: []string{"ls"},
		Short:   "List notetakers",
		Long:    `List all notetakers for a grant. Filter by state using --state flag.`,
		Example: `  # List all notetakers
  nylas notetaker list

  # List only scheduled notetakers
  nylas notetaker list --state scheduled

  # List completed notetakers
  nylas notetaker list --state complete

  # Output as JSON
  nylas notetaker list --json`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}

			grantID, err := getGrantID(args)
			if err != nil {
				return err
			}

			ctx, cancel := createContext()
			defer cancel()

			params := &domain.NotetakerQueryParams{
				Limit: limit,
				State: state,
			}

			notetakers, err := client.ListNotetakers(ctx, grantID, params)
			if err != nil {
				return fmt.Errorf("failed to list notetakers: %w", err)
			}

			if outputJSON {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(notetakers)
			}

			if len(notetakers) == 0 {
				fmt.Println("No notetakers found.")
				return nil
			}

			cyan := color.New(color.FgCyan)
			green := color.New(color.FgGreen)
			yellow := color.New(color.FgYellow)
			dim := color.New(color.Faint)

			fmt.Printf("Found %d notetaker(s):\n\n", len(notetakers))

			for _, n := range notetakers {
				cyan.Printf("ID: %s\n", n.ID)
				fmt.Printf("  State:   %s\n", formatState(n.State))
				if n.MeetingTitle != "" {
					fmt.Printf("  Title:   %s\n", n.MeetingTitle)
				}
				if n.MeetingLink != "" {
					fmt.Printf("  Link:    %s\n", truncate(n.MeetingLink, 60))
				}
				if n.MeetingInfo != nil && n.MeetingInfo.Provider != "" {
					caser := cases.Title(language.English)
					green.Printf("  Provider: %s\n", caser.String(n.MeetingInfo.Provider))
				}
				if !n.JoinTime.IsZero() {
					yellow.Printf("  Join:    %s\n", n.JoinTime.Local().Format("Mon Jan 2, 2006 3:04 PM"))
				}
				if !n.CreatedAt.IsZero() {
					dim.Printf("  Created: %s\n", formatTimeAgo(n.CreatedAt))
				}
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "l", 20, "Maximum number of notetakers to return")
	cmd.Flags().StringVar(&state, "state", "", "Filter by state (scheduled, connecting, attending, complete, cancelled, failed)")
	cmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")

	return cmd
}

func formatState(state string) string {
	switch state {
	case domain.NotetakerStateScheduled:
		return color.YellowString("scheduled")
	case domain.NotetakerStateConnecting:
		return color.CyanString("connecting")
	case domain.NotetakerStateWaitingForEntry:
		return color.CyanString("waiting")
	case domain.NotetakerStateAttending:
		return color.GreenString("attending")
	case domain.NotetakerStateMediaProcessing:
		return color.CyanString("processing")
	case domain.NotetakerStateComplete:
		return color.GreenString("complete")
	case domain.NotetakerStateCancelled:
		return color.New(color.Faint).Sprint("cancelled")
	case domain.NotetakerStateFailed:
		return color.RedString("failed")
	default:
		return state
	}
}
