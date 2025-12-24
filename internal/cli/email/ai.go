package email

import (
	"github.com/spf13/cobra"
)

// newAICmd creates the AI command group for email intelligence features.
func newAICmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ai",
		Aliases: []string{"intelligence"},
		Short:   "AI-powered email intelligence",
		Long: `AI-powered email intelligence features.

Use AI to analyze your emails and get actionable insights:
- Inbox analysis: Summarize recent emails with categories and action items
- Smart prioritization: Identify emails that need attention`,
		Example: `  # Analyze last 10 emails
  nylas email ai analyze

  # Analyze last 25 emails
  nylas email ai analyze --limit 25

  # Use specific AI provider
  nylas email ai analyze --provider claude

  # Only analyze unread emails
  nylas email ai analyze --unread`,
	}

	// Add AI subcommands
	cmd.AddCommand(newAnalyzeCmd())

	return cmd
}
