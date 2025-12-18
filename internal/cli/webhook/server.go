package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/fatih/color"
	"github.com/mqasimca/nylas/internal/adapters/tunnel"
	"github.com/mqasimca/nylas/internal/adapters/webhookserver"
	"github.com/mqasimca/nylas/internal/cli/common"
	"github.com/mqasimca/nylas/internal/ports"
	"github.com/spf13/cobra"
)

// Color helpers
var (
	cyan   = color.New(color.FgCyan)
	green  = color.New(color.FgGreen)
	yellow = color.New(color.FgYellow)
	red    = color.New(color.FgRed)
	blue   = color.New(color.FgBlue)
	bold   = color.New(color.Bold)
	dim    = color.New(color.Faint)
)

func newServerCmd() *cobra.Command {
	var (
		port          int
		path          string
		tunnelType    string
		webhookSecret string
		jsonOutput    bool
		quiet         bool
	)

	cmd := &cobra.Command{
		Use:   "server",
		Short: "Start a local webhook receiver server",
		Long: `Start a local HTTP server to receive and display webhook events.

The server can optionally expose itself via a tunnel (cloudflared) for
receiving webhooks from the internet when developing locally.

Examples:
  # Start server on default port 3000
  nylas webhooks server

  # Start server with cloudflared tunnel
  nylas webhooks server --tunnel cloudflared

  # Start server on custom port with tunnel
  nylas webhooks server --port 8080 --tunnel cloudflared

  # Start server with webhook signature verification
  nylas webhooks server --tunnel cloudflared --secret your-webhook-secret

Press Ctrl+C to stop the server.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServer(port, path, tunnelType, webhookSecret, jsonOutput, quiet)
		},
	}

	cmd.Flags().IntVarP(&port, "port", "p", 3000, "Port to listen on")
	cmd.Flags().StringVar(&path, "path", "/webhook", "Webhook endpoint path")
	cmd.Flags().StringVarP(&tunnelType, "tunnel", "t", "", "Tunnel provider (cloudflared)")
	cmd.Flags().StringVarP(&webhookSecret, "secret", "s", "", "Webhook secret for signature verification")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output events as JSON")
	cmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Suppress startup messages, only show events")

	return cmd
}

func runServer(port int, path, tunnelType, webhookSecret string, jsonOutput, quiet bool) error {
	// Create server config
	config := ports.WebhookServerConfig{
		Port:           port,
		Path:           path,
		WebhookSecret:  webhookSecret,
		TunnelProvider: tunnelType,
	}

	// Create webhook server
	server := webhookserver.NewServer(config)

	// Set up tunnel if requested
	if tunnelType != "" {
		switch strings.ToLower(tunnelType) {
		case "cloudflared", "cloudflare", "cf":
			if !tunnel.IsCloudflaredInstalled() {
				return common.NewUserError(
					"cloudflared is not installed",
					"Install it with: brew install cloudflared (macOS) or see https://developers.cloudflare.com/cloudflare-one/connections/connect-apps/install-and-setup/installation/",
				)
			}
			localURL := fmt.Sprintf("http://localhost:%d", port)
			t := tunnel.NewCloudflaredTunnel(localURL)
			server.SetTunnel(t)
		default:
			return common.NewUserError(
				fmt.Sprintf("unsupported tunnel provider: %s", tunnelType),
				"Supported providers: cloudflared",
			)
		}
	}

	// Set up context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Print startup message
	if !quiet {
		printStartupBanner()
	}

	// Start spinner while starting tunnel
	var spinner *common.Spinner
	if tunnelType != "" && !quiet {
		spinner = common.NewSpinner("Starting tunnel...")
		spinner.Start()
	}

	// Start the server
	if err := server.Start(ctx); err != nil {
		if spinner != nil {
			spinner.Stop()
		}
		return fmt.Errorf("failed to start server: %w", err)
	}

	if spinner != nil {
		spinner.Stop()
	}

	// Print server info
	stats := server.GetStats()
	if !quiet {
		printServerInfo(stats, tunnelType)
	}

	// Event display loop
	go func() {
		for event := range server.Events() {
			if jsonOutput {
				printEventJSON(event)
			} else {
				printEventFormatted(event, quiet)
			}
		}
	}()

	// Wait for interrupt
	<-sigChan

	if !quiet {
		fmt.Println("\n\nShutting down server...")
	}

	// Stop the server
	if err := server.Stop(); err != nil {
		return fmt.Errorf("error during shutdown: %w", err)
	}

	if !quiet {
		finalStats := server.GetStats()
		fmt.Printf("Server stopped. Total events received: %d\n", finalStats.EventsReceived)
	}

	return nil
}

func printStartupBanner() {
	fmt.Println()
	cyan.Println("╔══════════════════════════════════════════════════════════════╗")
	cyan.Print("║")
	fmt.Print("              ")
	bold.Print("Nylas Webhook Server")
	fmt.Print("                         ")
	cyan.Println("║")
	cyan.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()
}

func printServerInfo(stats ports.WebhookServerStats, tunnelType string) {
	green.Println("✓ Server started successfully")
	fmt.Println()

	bold.Print("  Local URL:    ")
	fmt.Println(stats.LocalURL)

	if stats.PublicURL != "" {
		bold.Print("  Public URL:   ")
		green.Println(stats.PublicURL)
		fmt.Println()
		bold.Print("  Tunnel:       ")
		fmt.Printf("%s (%s)\n", tunnelType, stats.TunnelStatus)
	}

	fmt.Println()
	yellow.Println("Register this URL with Nylas:")
	webhookURL := stats.LocalURL
	if stats.PublicURL != "" {
		webhookURL = stats.PublicURL
	}
	fmt.Printf("  nylas webhooks create --url %s --triggers message.created\n", webhookURL)
	fmt.Println()
	dim.Println("Press Ctrl+C to stop")
	fmt.Println()
	cyan.Println("─────────────────────────────────────────────────────────────────")
	bold.Println("Incoming Webhooks:")
	fmt.Println()
}

func printEventJSON(event *ports.WebhookEvent) {
	data, _ := json.Marshal(event)
	fmt.Println(string(data))
}

func printEventFormatted(event *ports.WebhookEvent, quiet bool) {
	timestamp := event.ReceivedAt.Format("15:04:05")

	// Determine verification status
	verifyIcon := ""
	if event.Signature != "" {
		if event.Verified {
			verifyIcon = green.Sprint(" ✓")
		} else {
			verifyIcon = red.Sprint(" ✗")
		}
	}

	// Event type coloring
	var typeColorFn func(a ...interface{}) string
	switch {
	case strings.Contains(event.Type, "created"):
		typeColorFn = green.Sprint
	case strings.Contains(event.Type, "deleted"):
		typeColorFn = red.Sprint
	case strings.Contains(event.Type, "updated"):
		typeColorFn = blue.Sprint
	default:
		typeColorFn = yellow.Sprint
	}

	fmt.Printf("%s %s%s\n",
		dim.Sprintf("[%s]", timestamp),
		typeColorFn(event.Type),
		verifyIcon,
	)

	if !quiet {
		// Print additional details
		if event.ID != "" {
			fmt.Printf("  %s %s\n", dim.Sprint("ID:"), event.ID)
		}
		if event.GrantID != "" {
			fmt.Printf("  %s %s\n", dim.Sprint("Grant:"), event.GrantID)
		}

		// Print a summary of the payload
		if event.Body != nil {
			if data, ok := event.Body["data"].(map[string]interface{}); ok {
				if obj, ok := data["object"].(map[string]interface{}); ok {
					// Print key fields based on event type
					if subject, ok := obj["subject"].(string); ok {
						fmt.Printf("  %s %s\n", dim.Sprint("Subject:"), truncate(subject, 60))
					}
					if title, ok := obj["title"].(string); ok {
						fmt.Printf("  %s %s\n", dim.Sprint("Title:"), truncate(title, 60))
					}
					if email, ok := obj["email"].(string); ok {
						fmt.Printf("  %s %s\n", dim.Sprint("Email:"), email)
					}
				}
			}
		}
		fmt.Println()
	}
}
