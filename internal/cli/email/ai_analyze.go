package email

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mqasimca/nylas/internal/adapters/ai"
	"github.com/mqasimca/nylas/internal/cli/common"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/spf13/cobra"
)

func newAnalyzeCmd() *cobra.Command {
	var (
		limit    int
		provider string
		unread   bool
		folder   string
	)

	cmd := &cobra.Command{
		Use:   "analyze [grant-id]",
		Short: "Analyze recent emails with AI",
		Long: `Analyze recent emails using AI to get a summary, categorization, and action items.

This command fetches your recent emails and uses AI to provide:
- A brief summary of your inbox
- Categorization of emails (Work, Personal, Newsletters, etc.)
- Action items that need your attention
- Key highlights from your emails`,
		Example: `  # Analyze last 10 emails
  nylas email ai analyze

  # Analyze last 25 emails
  nylas email ai analyze --limit 25

  # Use specific AI provider
  nylas email ai analyze --provider claude

  # Only analyze unread emails
  nylas email ai analyze --unread

  # Analyze specific folder
  nylas email ai analyze --folder SENT`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return fmt.Errorf("failed to get client: %w", err)
			}

			grantID, err := common.GetGrantID(args)
			if err != nil {
				return fmt.Errorf("failed to get grant ID: %w", err)
			}

			// AI analysis can take time - use longer timeout
			ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
			defer cancel()

			// Load config for AI settings
			configStore := common.GetConfigStore(cmd)
			cfg, err := configStore.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			if cfg.AI == nil || !cfg.AI.IsConfigured() {
				return fmt.Errorf("AI is not configucommon.Red. Run 'nylas config ai setup' to configure AI providers")
			}

			// Fetch emails
			fmt.Printf("📧 Fetching %d emails...\n", limit)

			params := &domain.MessageQueryParams{
				Limit: limit,
			}

			if unread {
				params.Unread = &unread
			}

			if folder != "" {
				params.In = []string{folder}
			} else {
				params.In = []string{"INBOX"}
			}

			messages, err := client.GetMessagesWithParams(ctx, grantID, params)
			if err != nil {
				return fmt.Errorf("failed to fetch emails: %w", err)
			}

			if len(messages) == 0 {
				fmt.Println("No emails found to analyze.")
				return nil
			}

			fmt.Printf("🔍 Analyzing %d emails with AI...\n\n", len(messages))

			// Create AI router and analyzer
			router := ai.NewRouter(cfg.AI)
			analyzer := ai.NewEmailAnalyzer(client, router)

			// Analyze emails
			req := &ai.InboxSummaryRequest{
				Messages:     messages,
				ProviderName: provider,
			}

			result, err := analyzer.AnalyzeInbox(ctx, req)
			if err != nil {
				return fmt.Errorf("AI analysis failed: %w", err)
			}

			// Display results
			displayInboxAnalysis(result, len(messages))

			return nil
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "l", 10, "Number of emails to analyze")
	cmd.Flags().StringVarP(&provider, "provider", "p", "", "AI provider to use (ollama, claude, openai, groq)")
	cmd.Flags().BoolVar(&unread, "unread", false, "Only analyze unread emails")
	cmd.Flags().StringVar(&folder, "folder", "", "Folder to analyze (default: INBOX)")

	return cmd
}

func displayInboxAnalysis(result *ai.InboxSummaryResponse, emailCount int) {
	fmt.Printf("📧 Email Analysis (%d emails)\n", emailCount)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Summary
	if result.Summary != "" {
		fmt.Println("📋 Summary")
		fmt.Println(result.Summary)
		fmt.Println()
	}

	// Categories
	if len(result.Categories) > 0 {
		fmt.Println("📁 Categories")
		for _, cat := range result.Categories {
			fmt.Printf("  %s (%d)\n", cat.Name, cat.Count)
			for _, subject := range cat.Subjects {
				fmt.Printf("    • %s\n", common.Truncate(subject, 50))
			}
		}
		fmt.Println()
	}

	// Action Items
	if len(result.ActionItems) > 0 {
		fmt.Println("⚡ Action Items")
		for _, item := range result.ActionItems {
			urgencyIcon := "🔵"
			switch item.Urgency {
			case "high":
				urgencyIcon = "🔴"
			case "medium":
				urgencyIcon = "🟡"
			case "low":
				urgencyIcon = "🔵"
			}
			fmt.Printf("  %s %s: \"%s\" from %s\n", urgencyIcon, strings.ToUpper(item.Urgency), common.Truncate(item.Subject, 40), item.From)
			if item.Reason != "" {
				fmt.Printf("     → %s\n", item.Reason)
			}
		}
		fmt.Println()
	}

	// Highlights
	if len(result.Highlights) > 0 {
		fmt.Println("✨ Highlights")
		for _, highlight := range result.Highlights {
			fmt.Printf("  • %s\n", highlight)
		}
		fmt.Println()
	}

	// Footer
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("Provider: %s", result.ProviderUsed)
	if result.TokensUsed > 0 {
		fmt.Printf(" | Tokens: %d", result.TokensUsed)
	}
	fmt.Println()
}
