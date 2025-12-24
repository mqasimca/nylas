package demo

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/mqasimca/nylas/internal/adapters/nylas"
	"github.com/mqasimca/nylas/internal/tui"
)

// newDemoTUICmd creates the demo tui command.
func newDemoTUICmd() *cobra.Command {
	var refreshInterval int
	var theme string

	cmd := &cobra.Command{
		Use:   "tui [resource]",
		Short: "Launch interactive TUI with sample data",
		Long: `Launch the k9s-style terminal interface with demo data.

Explore the full TUI experience without any credentials:
  - Browse sample emails and threads
  - View demo calendar events
  - Explore sample contacts
  - Test keyboard navigation and commands

Navigation:
  ↑/k, ↓/j    Move up/down
  g/G         Go to top/bottom
  enter       Open/select
  esc         Go back
  :           Command mode
  /           Filter
  ?           Help
  Ctrl+C      Quit

Themes:
  k9s, amber, green, apple2, vintage, ibm, futuristic, matrix, norton`,
		Example: `  # Launch demo TUI at dashboard
  nylas demo tui

  # Launch demo TUI directly to messages
  nylas demo tui messages

  # Launch demo TUI with retro amber theme
  nylas demo tui --theme amber

  # Launch demo TUI to calendar view
  nylas demo tui events`,
		RunE: func(cmd *cobra.Command, args []string) error {
			initialView := ""
			if len(args) > 0 {
				initialView = args[0]
			}
			return runDemoTUI(time.Duration(refreshInterval)*time.Second, initialView, tui.ThemeName(theme))
		},
	}

	cmd.Flags().IntVar(&refreshInterval, "refresh", 3, "Refresh interval in seconds")
	cmd.Flags().StringVar(&theme, "theme", "k9s", "Color theme (k9s, amber, green, apple2, vintage, ibm, futuristic, matrix, norton)")

	return cmd
}

func runDemoTUI(refreshInterval time.Duration, initialView string, theme tui.ThemeName) error {
	client := nylas.NewDemoClient()

	app := tui.NewApp(tui.Config{
		Client:          client,
		GrantID:         "demo-grant-001",
		Email:           "demo@example.com",
		Provider:        "google",
		RefreshInterval: refreshInterval,
		InitialView:     initialView,
		Theme:           theme,
	})

	return app.Run()
}
