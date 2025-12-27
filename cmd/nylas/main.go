// Package main is the entry point for the nylas CLI.
package main

import (
	"fmt"
	"os"

	"github.com/mqasimca/nylas/internal/air"
	"github.com/mqasimca/nylas/internal/cli"
	"github.com/mqasimca/nylas/internal/cli/admin"
	"github.com/mqasimca/nylas/internal/cli/ai"
	"github.com/mqasimca/nylas/internal/cli/auth"
	"github.com/mqasimca/nylas/internal/cli/calendar"
	"github.com/mqasimca/nylas/internal/cli/contacts"
	"github.com/mqasimca/nylas/internal/cli/demo"
	"github.com/mqasimca/nylas/internal/cli/email"
	"github.com/mqasimca/nylas/internal/cli/inbound"
	"github.com/mqasimca/nylas/internal/cli/mcp"
	"github.com/mqasimca/nylas/internal/cli/notetaker"
	"github.com/mqasimca/nylas/internal/cli/otp"
	"github.com/mqasimca/nylas/internal/cli/scheduler"
	"github.com/mqasimca/nylas/internal/cli/timezone"
	"github.com/mqasimca/nylas/internal/cli/ui"
	"github.com/mqasimca/nylas/internal/cli/update"
	"github.com/mqasimca/nylas/internal/cli/webhook"
)

func main() {
	// Add subcommands
	rootCmd := cli.GetRootCmd()
	rootCmd.AddCommand(ai.NewAICmd())
	rootCmd.AddCommand(auth.NewAuthCmd())
	rootCmd.AddCommand(otp.NewOTPCmd())
	rootCmd.AddCommand(email.NewEmailCmd())
	rootCmd.AddCommand(calendar.NewCalendarCmd())
	rootCmd.AddCommand(contacts.NewContactsCmd())
	rootCmd.AddCommand(scheduler.NewSchedulerCmd())
	rootCmd.AddCommand(admin.NewAdminCmd())
	rootCmd.AddCommand(webhook.NewWebhookCmd())
	rootCmd.AddCommand(notetaker.NewNotetakerCmd())
	rootCmd.AddCommand(inbound.NewInboundCmd())
	rootCmd.AddCommand(timezone.NewTimezoneCmd())
	rootCmd.AddCommand(mcp.NewMCPCmd())
	rootCmd.AddCommand(demo.NewDemoCmd())
	rootCmd.AddCommand(cli.NewTUICmd())
	rootCmd.AddCommand(ui.NewUICmd())
	rootCmd.AddCommand(air.NewAirCmd())
	rootCmd.AddCommand(update.NewUpdateCmd())

	if err := cli.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
