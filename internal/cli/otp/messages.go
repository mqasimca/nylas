package otp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/mqasimca/nylas/internal/adapters/nylas"
	"github.com/mqasimca/nylas/internal/domain"
)

func newMessagesCmd() *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "messages [email]",
		Short: "Show recent messages (debug)",
		Long: `Show recent messages from your email.

If no email is specified, uses the default account.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			otpSvc, err := createOTPService()
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			var messages []domain.Message
			if len(args) > 0 {
				messages, err = otpSvc.GetMessages(ctx, args[0], limit)
			} else {
				messages, err = otpSvc.GetMessagesDefault(ctx, limit)
			}
			if err != nil {
				return err
			}

			if len(messages) == 0 {
				fmt.Println("No messages found")
				return nil
			}

			jsonOutput, _ := cmd.Root().PersistentFlags().GetBool("json")
			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(messages)
			}

			cyan := color.New(color.FgCyan, color.Bold)
			bold := color.New(color.Bold)
			green := color.New(color.FgGreen)

			_, _ = cyan.Printf("Recent Messages (%d)\n", len(messages))
			fmt.Println()

			// Print header
			_, _ = bold.Printf("  %-3s  %-24s  %-24s  %-14s  %-5s\n", "#", "FROM", "SUBJECT", "DATE", "OTP?")

			for i, msg := range messages {
				from := ""
				if len(msg.From) > 0 {
					from = msg.From[0].Email
				}
				if len(from) > 22 {
					from = from[:22] + "…"
				}

				subject := msg.Subject
				if len(subject) > 22 {
					subject = subject[:22] + "…"
				}

				date := msg.Date.Format("Jan 02 15:04")

				// Check for OTP (try body first, then snippet)
				otp := nylas.ExtractOTP(msg.Subject, msg.Body)
				if otp == "" {
					otp = nylas.ExtractOTP(msg.Subject, msg.Snippet)
				}
				otpIndicator := "—"
				if otp != "" {
					otpIndicator = green.Sprint("✓ " + otp)
				}

				fmt.Printf("  %-3d  %-24s  %-24s  %-14s  %-5s\n",
					i+1, from, subject, date, otpIndicator)
			}

			return nil
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "l", 10, "Number of messages to fetch")

	return cmd
}
