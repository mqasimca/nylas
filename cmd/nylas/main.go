// Package main is the entry point for the nylas CLI.
package main

import (
	"fmt"
	"os"

	"github.com/mqasimca/nylas/internal/cli"
	"github.com/mqasimca/nylas/internal/cli/auth"
	"github.com/mqasimca/nylas/internal/cli/calendar"
	"github.com/mqasimca/nylas/internal/cli/contacts"
	"github.com/mqasimca/nylas/internal/cli/email"
	"github.com/mqasimca/nylas/internal/cli/otp"
	"github.com/mqasimca/nylas/internal/cli/webhook"
)

func main() {
	// Add subcommands
	rootCmd := cli.GetRootCmd()
	rootCmd.AddCommand(auth.NewAuthCmd())
	rootCmd.AddCommand(otp.NewOTPCmd())
	rootCmd.AddCommand(email.NewEmailCmd())
	rootCmd.AddCommand(calendar.NewCalendarCmd())
	rootCmd.AddCommand(contacts.NewContactsCmd())
	rootCmd.AddCommand(webhook.NewWebhookCmd())

	if err := cli.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
