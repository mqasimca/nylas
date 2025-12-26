// Package ui provides a web-based user interface for the Nylas CLI.
package ui

import (
	"fmt"

	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

// NewUICmd creates the ui command.
func NewUICmd() *cobra.Command {
	var (
		port      int
		noBrowser bool
	)

	cmd := &cobra.Command{
		Use:   "ui",
		Short: "Start the web-based user interface",
		Long: `Start a local web server providing a graphical interface for Nylas CLI.

The UI allows you to:
  - Configure API credentials
  - View authenticated accounts
  - Manage email and calendar (coming soon)

The server runs on localhost only for security.`,
		Example: `  # Start UI on default port (7363)
  nylas ui

  # Start UI on custom port
  nylas ui --port 8080

  # Start without opening browser
  nylas ui --no-browser`,
		RunE: func(cmd *cobra.Command, args []string) error {
			addr := fmt.Sprintf("localhost:%d", port)
			url := fmt.Sprintf("http://%s", addr)

			fmt.Printf("Starting Nylas UI at %s\n", url)
			fmt.Println("Press Ctrl+C to stop")
			fmt.Println()

			// Open browser unless disabled
			if !noBrowser {
				if err := browser.OpenURL(url); err != nil {
					fmt.Printf("Could not open browser: %v\n", err)
					fmt.Printf("Please open %s manually\n", url)
				}
			}

			// Start the server (blocks until interrupted)
			server := NewServer(addr)
			return server.Start()
		},
	}

	cmd.Flags().IntVarP(&port, "port", "p", 7363, "Port to run the server on")
	cmd.Flags().BoolVar(&noBrowser, "no-browser", false, "Don't open browser automatically")

	return cmd
}
