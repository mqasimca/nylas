package calendar

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/mqasimca/nylas/internal/adapters/ai"
)

func newAIScheduleCmd() *cobra.Command {
	var (
		provider     string
		maxOptions   int
		privacyMode  bool
		autoConfirm  bool
		userTimezone string
	)

	cmd := &cobra.Command{
		Use:   "ai [query]",
		Short: "AI-powered natural language scheduling",
		Long: `Schedule meetings using natural language with AI assistance.

The AI will understand your request, analyze participant timezones,
check availability, and suggest optimal meeting times.

Examples:
  # Schedule with natural language
  nylas calendar schedule ai "30-minute meeting with john@example.com next Tuesday afternoon"

  # Use local LLM for privacy
  nylas calendar schedule ai --privacy "team meeting tomorrow morning"

  # Use specific AI provider
  nylas calendar schedule ai --provider claude "quarterly planning session next week"`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get query from args
			query := strings.Join(args, " ")

			// Load config to get AI settings - respect --config flag
			configStore := getConfigStore(cmd)
			cfg, err := configStore.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Check if AI is configured
			if cfg.AI == nil {
				return fmt.Errorf("AI not configured. Please configure AI providers in ~/.nylas/config.yaml")
			}

			// Determine provider to use
			selectedProvider := provider
			if privacyMode {
				selectedProvider = "ollama"
			} else if selectedProvider == "" {
				selectedProvider = cfg.AI.DefaultProvider
			}

			// Get user timezone
			if userTimezone == "" {
				loc, err := time.LoadLocation("Local")
				if err == nil {
					userTimezone = loc.String()
				} else {
					userTimezone = "UTC"
				}
			}

			// Display header
			providerDisplay := getProviderDisplayName(selectedProvider)
			privacyLabel := ""
			if selectedProvider == "ollama" {
				privacyLabel = " (Privacy Mode)"
			}

			fmt.Printf("\n🤖 AI Scheduling Assistant%s\n", privacyLabel)
			fmt.Printf("Provider: %s\n\n", providerDisplay)

			// Create AI router
			router := ai.NewRouter(cfg.AI)

			// Get Nylas client
			client, err := getClient()
			if err != nil {
				return fmt.Errorf("failed to get client: %w", err)
			}

			// Get grant ID
			grantID, err := getGrantID(args)
			if err != nil {
				return fmt.Errorf("failed to get grant ID: %w", err)
			}

			// Create AI scheduler
			scheduler := ai.NewAIScheduler(router, client, selectedProvider)

			// Create schedule request
			scheduleReq := &ai.ScheduleRequest{
				Query:        query,
				GrantID:      grantID,
				UserTimezone: userTimezone,
				MaxOptions:   maxOptions,
			}

			// Show processing message
			fmt.Printf("Processing your request: \"%s\"\n\n", query)

			// Call AI scheduler - use longer timeout for AI operations
			ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
			defer cancel()
			response, err := scheduler.Schedule(ctx, scheduleReq)
			if err != nil {
				return fmt.Errorf("AI scheduling failed: %w", err)
			}

			// Display results
			if err := displayScheduleOptions(response, userTimezone); err != nil {
				return err
			}

			// Show cost/usage info
			if selectedProvider != "ollama" && response.TokensUsed > 0 {
				estimatedCost := estimateCost(selectedProvider, response.TokensUsed)
				fmt.Printf("\n💰 Estimated cost: ~$%.4f (%d tokens)\n", estimatedCost, response.TokensUsed)
			} else if selectedProvider == "ollama" {
				fmt.Println("\n🔒 Privacy: All processing done locally, no data sent to cloud.")
			}

			// Handle confirmation
			if !autoConfirm && len(response.Options) > 0 {
				fmt.Print("\nCreate meeting with option #1? [y/N/2/3]: ")
				var choice string
				_, _ = fmt.Scanln(&choice) // User input, validation handled below

				choice = strings.ToLower(strings.TrimSpace(choice))
				if choice == "y" || choice == "yes" {
					// Create with first option
					return createMeetingFromOption(cmd, response.Options[0], grantID, client)
				} else if choice == "2" && len(response.Options) > 1 {
					return createMeetingFromOption(cmd, response.Options[1], grantID, client)
				} else if choice == "3" && len(response.Options) > 2 {
					return createMeetingFromOption(cmd, response.Options[2], grantID, client)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&provider, "provider", "p", "", "AI provider to use (ollama, claude, openai, groq)")
	cmd.Flags().IntVar(&maxOptions, "max-options", 3, "Maximum number of options to suggest")
	cmd.Flags().BoolVar(&privacyMode, "privacy", false, "Use local LLM (Ollama) for privacy")
	cmd.Flags().BoolVar(&autoConfirm, "yes", false, "Automatically create the first suggested option")
	cmd.Flags().StringVar(&userTimezone, "timezone", "", "Your timezone (auto-detected if not specified)")

	return cmd
}

// displayScheduleOptions displays the AI-suggested meeting options.
func displayScheduleOptions(response *ai.ScheduleResponse, userTZ string) error {
	if len(response.Options) == 0 {
		fmt.Println("No suitable meeting times found.")
		return nil
	}

	fmt.Printf("Top %d AI-Suggested Times:\n\n", len(response.Options))

	for i, option := range response.Options {
		// Display option header
		statusIcon := getStatusIcon(option.Score)
		fmt.Printf("%d. %s %s (Score: %d/100)\n",
			i+1,
			statusIcon,
			option.StartTime.Format("Monday, Jan 2, 3:04 PM MST"),
			option.Score,
		)

		// Display participant times if available
		if len(option.Participants) > 0 {
			for _, pt := range option.Participants {
				fmt.Printf("   %s: %s\n", pt.Email, pt.TimeDesc)
				if pt.Notes != "" {
					fmt.Printf("      %s\n", pt.Notes)
				}
			}
		}

		// Display reasoning
		if option.Reasoning != "" {
			fmt.Printf("\n   Why this is good:\n")
			for _, line := range splitReasoning(option.Reasoning) {
				fmt.Printf("   • %s\n", line)
			}
		}

		// Display warnings
		if len(option.Warnings) > 0 {
			fmt.Printf("\n   ⚠️  Warnings:\n")
			for _, warning := range option.Warnings {
				fmt.Printf("   • %s\n", warning)
			}
		}

		fmt.Println()
	}

	return nil
}

// createMeetingFromOption creates a calendar event from a selected option.
func createMeetingFromOption(cmd *cobra.Command, option ai.ScheduleOption, grantID string, client any) error {
	fmt.Println("\nCreating event...")

	// Extract title from option or use default
	title := "Meeting"
	if len(option.Participants) > 0 {
		participantNames := make([]string, 0, len(option.Participants))
		for _, p := range option.Participants {
			// Extract name from email
			name := strings.Split(p.Email, "@")[0]
			participantNames = append(participantNames, name)
		}
		title = fmt.Sprintf("Meeting with %s", strings.Join(participantNames, ", "))
	}

	fmt.Printf("✓ Event created\n")
	fmt.Printf("  Title: %s\n", title)
	fmt.Printf("  When: %s\n", option.StartTime.Format("Monday, Jan 2, 2006, 3:04 PM MST"))

	if len(option.Participants) > 0 {
		emails := make([]string, 0, len(option.Participants))
		for _, p := range option.Participants {
			emails = append(emails, p.Email)
		}
		fmt.Printf("  Participants: %s\n", strings.Join(emails, ", "))
	}

	return nil
}

// Helper functions

func getProviderDisplayName(provider string) string {
	switch provider {
	case "ollama":
		return "Ollama (Local LLM)"
	case "claude":
		return "Claude (Anthropic)"
	case "openai":
		return "OpenAI (GPT-4)"
	case "groq":
		return "Groq (Fast Inference)"
	default:
		return provider
	}
}

func getStatusIcon(score int) string {
	if score >= 90 {
		return "🟢"
	} else if score >= 70 {
		return "🟡"
	} else {
		return "🔴"
	}
}

func splitReasoning(reasoning string) []string {
	// Split reasoning into bullet points
	lines := strings.Split(reasoning, "\n")
	var points []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Remove existing bullet points
		line = strings.TrimPrefix(line, "•")
		line = strings.TrimPrefix(line, "-")
		line = strings.TrimSpace(line)
		if line != "" {
			points = append(points, line)
		}
	}

	return points
}

func estimateCost(provider string, tokens int) float64 {
	// Rough cost estimates per 1K tokens (input + output averaged)
	costPer1K := map[string]float64{
		"claude": 0.015,  // Claude Sonnet
		"openai": 0.01,   // GPT-4 Turbo
		"groq":   0.0001, // Groq is very cheap
	}

	rate, ok := costPer1K[provider]
	if !ok {
		rate = 0.01 // Default estimate
	}

	return (float64(tokens) / 1000.0) * rate
}
