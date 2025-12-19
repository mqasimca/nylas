package inbound

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/mqasimca/nylas/internal/adapters/config"
	"github.com/mqasimca/nylas/internal/adapters/keyring"
	"github.com/mqasimca/nylas/internal/adapters/nylas"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/ports"
)

var (
	boldWhite = color.New(color.FgWhite, color.Bold)
	cyan      = color.New(color.FgCyan)
	green     = color.New(color.FgGreen)
	yellow    = color.New(color.FgYellow)
	red       = color.New(color.FgRed)
	dim       = color.New(color.Faint)
)

// getClient creates and configures a Nylas client.
func getClient() (ports.NylasClient, error) {
	configStore := config.NewDefaultFileStore()
	cfg, _ := configStore.Load()

	secretStore, err := keyring.NewSecretStore(config.DefaultConfigDir())
	if err != nil {
		return nil, fmt.Errorf("failed to initialize secret store: %w", err)
	}

	apiKey, err := secretStore.Get(ports.KeyAPIKey)
	if err != nil {
		return nil, fmt.Errorf("API key not configured. Run 'nylas auth config' first")
	}

	clientID, _ := secretStore.Get(ports.KeyClientID)
	clientSecret, _ := secretStore.Get(ports.KeyClientSecret)

	client := nylas.NewHTTPClient()
	client.SetRegion(cfg.Region)
	client.SetCredentials(clientID, clientSecret, apiKey)

	return client, nil
}

// getInboxID gets the inbox ID from args or environment variable.
func getInboxID(args []string) (string, error) {
	if len(args) > 0 {
		return args[0], nil
	}

	// Try to get from environment variable
	if envID := os.Getenv("NYLAS_INBOUND_GRANT_ID"); envID != "" {
		return envID, nil
	}

	return "", fmt.Errorf("inbox ID required. Provide as argument or set NYLAS_INBOUND_GRANT_ID environment variable")
}

// createContext creates a context with timeout.
func createContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 30*time.Second)
}

// printError prints an error message in red.
func printError(format string, args ...interface{}) {
	red.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
}

// printSuccess prints a success message in green.
func printSuccess(format string, args ...interface{}) {
	green.Printf(format+"\n", args...)
}

// formatTimeAgo formats a time as a relative string.
func formatTimeAgo(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "just now"
	} else if diff < time.Hour {
		mins := int(diff.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	} else if diff < 24*time.Hour {
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	} else if diff < 48*time.Hour {
		return "yesterday"
	} else {
		days := int(diff.Hours() / 24)
		return fmt.Sprintf("%d days ago", days)
	}
}

// truncate truncates a string to the given length.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// formatContact formats a contact for display.
func formatContact(c domain.EmailParticipant) string {
	if c.Name != "" {
		return c.Name
	}
	return c.Email
}

// formatContacts formats multiple contacts for display.
func formatContacts(contacts []domain.EmailParticipant) string {
	names := make([]string, len(contacts))
	for i, c := range contacts {
		names[i] = formatContact(c)
	}
	return strings.Join(names, ", ")
}

// printInboxSummary prints a single-line inbox summary.
func printInboxSummary(inbox domain.InboundInbox, index int) {
	status := green.Sprint("active")
	if inbox.GrantStatus != "valid" {
		status = yellow.Sprint(inbox.GrantStatus)
	}

	createdStr := formatTimeAgo(inbox.CreatedAt.Time)

	fmt.Printf("%d. %-40s %s  %s\n",
		index+1,
		cyan.Sprint(inbox.Email),
		dim.Sprint(createdStr),
		status,
	)
	dim.Printf("   ID: %s\n", inbox.ID)
}

// printInboxDetails prints detailed inbox information.
func printInboxDetails(inbox domain.InboundInbox) {
	fmt.Println(strings.Repeat("─", 60))
	boldWhite.Printf("Inbox: %s\n", inbox.Email)
	fmt.Println(strings.Repeat("─", 60))
	fmt.Printf("ID:          %s\n", inbox.ID)
	fmt.Printf("Email:       %s\n", inbox.Email)
	fmt.Printf("Status:      %s\n", formatStatus(inbox.GrantStatus))
	fmt.Printf("Created:     %s (%s)\n", inbox.CreatedAt.Time.Format("Jan 2, 2006 3:04 PM"), formatTimeAgo(inbox.CreatedAt.Time))
	if !inbox.UpdatedAt.Time.IsZero() {
		fmt.Printf("Updated:     %s (%s)\n", inbox.UpdatedAt.Time.Format("Jan 2, 2006 3:04 PM"), formatTimeAgo(inbox.UpdatedAt.Time))
	}
	fmt.Println()
}

// formatStatus formats the grant status with color.
func formatStatus(status string) string {
	switch status {
	case "valid":
		return green.Sprint("active")
	case "invalid":
		return red.Sprint("invalid")
	default:
		return yellow.Sprint(status)
	}
}

// printInboundMessageSummary prints an inbound message summary.
func printInboundMessageSummary(msg domain.InboundMessage, _ int) {
	status := " "
	if msg.Unread {
		status = cyan.Sprint("●")
	}

	star := " "
	if msg.Starred {
		star = yellow.Sprint("★")
	}

	from := formatContacts(msg.From)
	if len(from) > 20 {
		from = from[:17] + "..."
	}

	subject := msg.Subject
	if len(subject) > 40 {
		subject = subject[:37] + "..."
	}

	dateStr := formatTimeAgo(msg.Date)
	if len(dateStr) > 12 {
		dateStr = msg.Date.Format("Jan 2")
	}

	fmt.Printf("%s %s %-20s %-40s %s\n", status, star, from, subject, dim.Sprint(dateStr))
	dim.Printf("      ID: %s\n", msg.ID)
}
