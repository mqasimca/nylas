package email

import (
	"fmt"
	"html"
	"strings"

	"github.com/mqasimca/nylas/internal/adapters/config"
	"github.com/mqasimca/nylas/internal/adapters/keyring"
	"github.com/mqasimca/nylas/internal/cli/common"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/ports"
)

// getClient creates and configures a Nylas client.
// Supports credentials from keyring/file store or environment variables.
func getClient() (ports.NylasClient, error) {
	return common.GetNylasClient()
}

// getGrantID gets the grant ID from args or default.
// If the argument contains '@', it's treated as an email and looked up.
func getGrantID(args []string) (string, error) {
	secretStore, err := keyring.NewSecretStore(config.DefaultConfigDir())
	if err != nil {
		return "", fmt.Errorf("couldn't access secret store: %w", err)
	}
	grantStore := keyring.NewGrantStore(secretStore)

	if len(args) > 0 {
		identifier := args[0]

		// If it looks like an email, try to find by email
		if strings.Contains(identifier, "@") {
			grant, err := grantStore.GetGrantByEmail(identifier)
			if err != nil {
				return "", fmt.Errorf("no grant found for email: %s", identifier)
			}
			return grant.ID, nil
		}

		// Otherwise treat as grant ID
		return identifier, nil
	}

	// Try to get default grant
	defaultGrant, err := grantStore.GetDefaultGrant()
	if err != nil {
		return "", fmt.Errorf("no grant ID provided and no default grant set. Use 'nylas auth list' to see available grants")
	}

	return defaultGrant, nil
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

// printMessage prints a message in a formatted way.
func printMessage(msg domain.Message, showBody bool) {
	// Status indicators
	status := ""
	if msg.Unread {
		status += common.Cyan.Sprint("●") + " "
	}
	if msg.Starred {
		status += common.Yellow.Sprint("★") + " "
	}

	// Print header
	fmt.Println(strings.Repeat("─", 60))
	_, _ = common.BoldWhite.Printf("Subject: %s\n", msg.Subject)
	fmt.Printf("From:    %s\n", formatContacts(msg.From))
	if len(msg.To) > 0 {
		fmt.Printf("To:      %s\n", formatContacts(msg.To))
	}
	fmt.Printf("Date:    %s (%s)\n", msg.Date.Format("Jan 2, 2006 3:04 PM"), common.FormatTimeAgo(msg.Date))
	if status != "" {
		fmt.Printf("Status:  %s\n", status)
	}
	if len(msg.Attachments) > 0 {
		fmt.Printf("Attachments: %d files\n", len(msg.Attachments))
		for _, a := range msg.Attachments {
			_, _ = common.Dim.Printf("  - %s (%s)\n", a.Filename, formatSize(a.Size))
		}
	}

	if showBody {
		fmt.Println(strings.Repeat("─", 60))
		body := msg.Body
		if body == "" {
			body = msg.Snippet
		}
		// Strip HTML tags for terminal display
		body = stripHTML(body)
		fmt.Println(body)
	}
	fmt.Println()
}

// printMessageRaw prints a message with raw body (no HTML processing).
func printMessageRaw(msg domain.Message) {
	// Print header
	fmt.Println(strings.Repeat("─", 60))
	_, _ = common.BoldWhite.Printf("Subject: %s\n", msg.Subject)
	fmt.Printf("From:    %s\n", formatContacts(msg.From))
	if len(msg.To) > 0 {
		fmt.Printf("To:      %s\n", formatContacts(msg.To))
	}
	fmt.Printf("Date:    %s (%s)\n", msg.Date.Format("Jan 2, 2006 3:04 PM"), common.FormatTimeAgo(msg.Date))
	fmt.Printf("ID:      %s\n", msg.ID)
	fmt.Println(strings.Repeat("─", 60))

	// Print raw body without any processing
	body := msg.Body
	if body == "" {
		body = msg.Snippet
	}
	fmt.Println(body)
	fmt.Println()
}

// printMessageSummary prints a single-line message summary.
func printMessageSummary(msg domain.Message, index int) {
	printMessageSummaryWithID(msg, index, false)
}

// printMessageSummaryWithID prints a single-line message summary, optionally with ID.
func printMessageSummaryWithID(msg domain.Message, index int, showID bool) {
	status := " "
	if msg.Unread {
		status = common.Cyan.Sprint("●")
	}

	star := " "
	if msg.Starred {
		star = common.Yellow.Sprint("★")
	}

	from := formatContacts(msg.From)
	if len(from) > 20 {
		from = from[:17] + "..."
	}

	subject := msg.Subject
	if len(subject) > 40 {
		subject = subject[:37] + "..."
	}

	dateStr := common.FormatTimeAgo(msg.Date)
	if len(dateStr) > 12 {
		dateStr = msg.Date.Format("Jan 2")
	}

	if showID {
		// Show full ID on its own line for easy copying
		fmt.Printf("%s %s %-20s %-40s %s\n", status, star, from, subject, common.Dim.Sprint(dateStr))
		_, _ = common.Dim.Printf("      ID: %s\n", msg.ID)
	} else {
		fmt.Printf("%s %s %-20s %-40s %s\n", status, star, from, subject, common.Dim.Sprint(dateStr))
	}
}

// formatSize formats a file size in bytes to a human-readable string.
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// stripHTML removes HTML tags from a string and decodes HTML entities.
func stripHTML(s string) string {
	// Remove style and script tags and their contents
	s = removeTagWithContent(s, "style")
	s = removeTagWithContent(s, "script")
	s = removeTagWithContent(s, "head")

	// Replace block-level elements with newlines before stripping tags
	blockTags := []string{"br", "p", "div", "tr", "li", "h1", "h2", "h3", "h4", "h5", "h6"}
	for _, tag := range blockTags {
		// Handle <br>, <br/>, <br />
		s = strings.ReplaceAll(s, "<"+tag+">", "\n")
		s = strings.ReplaceAll(s, "<"+tag+"/>", "\n")
		s = strings.ReplaceAll(s, "<"+tag+" />", "\n")
		s = strings.ReplaceAll(s, "</"+tag+">", "\n")
		// Case insensitive
		s = strings.ReplaceAll(s, "<"+strings.ToUpper(tag)+">", "\n")
		s = strings.ReplaceAll(s, "</"+strings.ToUpper(tag)+">", "\n")
	}

	// Strip remaining HTML tags
	var result strings.Builder
	inTag := false

	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			result.WriteRune(r)
		}
	}

	// Decode HTML entities (&nbsp;, &lt;, &gt;, etc.)
	text := html.UnescapeString(result.String())

	// Clean up whitespace
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	// Collapse multiple spaces on the same line
	for strings.Contains(text, "  ") {
		text = strings.ReplaceAll(text, "  ", " ")
	}

	// Collapse multiple newlines
	for strings.Contains(text, "\n\n\n") {
		text = strings.ReplaceAll(text, "\n\n\n", "\n\n")
	}

	// Trim spaces from each line
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}
	text = strings.Join(lines, "\n")

	// Remove leading/trailing empty lines
	return strings.TrimSpace(text)
}

// removeTagWithContent removes a tag and all its content.
func removeTagWithContent(s, tag string) string {
	result := s
	for {
		lower := strings.ToLower(result)
		startIdx := strings.Index(lower, "<"+tag)
		if startIdx == -1 {
			break
		}
		endTag := "</" + tag + ">"
		endIdx := strings.Index(lower[startIdx:], endTag)
		if endIdx == -1 {
			// No closing tag, just remove opening tag
			closeIdx := strings.Index(result[startIdx:], ">")
			if closeIdx == -1 {
				break
			}
			result = result[:startIdx] + result[startIdx+closeIdx+1:]
		} else {
			result = result[:startIdx] + result[startIdx+endIdx+len(endTag):]
		}
	}
	return result
}

// printSuccess prints a success message in green.
func printSuccess(format string, args ...any) {
	_, _ = common.Green.Printf(format+"\n", args...)
}
