package scheduler

import (
	"encoding/json"
	"fmt"

	"github.com/fatih/color"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/spf13/cobra"
)

func newSessionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "sessions",
		Aliases: []string{"session"},
		Short:   "Manage scheduler sessions",
		Long:    "Manage scheduler sessions for booking workflows.",
	}

	cmd.AddCommand(newSessionCreateCmd())
	cmd.AddCommand(newSessionShowCmd())

	return cmd
}

func newSessionCreateCmd() *cobra.Command {
	var (
		configID string
		ttl      int
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a scheduler session",
		Long:  "Create a new scheduler session for a configuration.",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}

			req := &domain.CreateSchedulerSessionRequest{
				ConfigurationID: configID,
				TimeToLive:      ttl,
			}

			ctx, cancel := createContext()
			defer cancel()

			session, err := client.CreateSchedulerSession(ctx, req)
			if err != nil {
				return fmt.Errorf("failed to create session: %w", err)
			}

			green := color.New(color.FgGreen)
			cyan := color.New(color.FgCyan)

			green.Println("✓ Created scheduler session")
			fmt.Printf("  Session ID: %s\n", cyan.Sprint(session.SessionID))
			fmt.Printf("  Configuration ID: %s\n", session.ConfigurationID)

			return nil
		},
	}

	cmd.Flags().StringVar(&configID, "config-id", "", "Configuration ID (required)")
	cmd.Flags().IntVar(&ttl, "ttl", 30, "Time to live in minutes")

	_ = cmd.MarkFlagRequired("config-id")

	return cmd
}

func newSessionShowCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "show <session-id>",
		Short: "Show scheduler session details",
		Long:  "Show detailed information about a scheduler session.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}

			ctx, cancel := createContext()
			defer cancel()

			session, err := client.GetSchedulerSession(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to get session: %w", err)
			}

			if jsonOutput {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(session)
			}

			cyan := color.New(color.FgCyan)
			bold := color.New(color.Bold)

			bold.Println("Scheduler Session")
			fmt.Printf("  Session ID: %s\n", cyan.Sprint(session.SessionID))
			fmt.Printf("  Configuration ID: %s\n", session.ConfigurationID)

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}
