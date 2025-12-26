package calendar

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mqasimca/nylas/internal/adapters/ai"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/spf13/cobra"
)

// newAnalyzeThreadCmd creates the analyze-thread AI command.
func newAnalyzeThreadCmd() *cobra.Command {
	var (
		threadID      string
		includeAgenda bool
		includeTime   bool
		createMeeting bool
		provider      string
		jsonOutput    bool
	)

	cmd := &cobra.Command{
		Use:   "analyze-thread --thread <thread-id>",
		Short: "Analyze email thread to extract meeting context",
		Long: `Analyze an email thread using AI to extract meeting context and generate insights.

This command uses AI to:
- Extract the primary purpose of the discussion
- Identify key topics and action items
- Detect priority level based on email tone
- Identify required vs optional participants
- Suggest optimal meeting duration
- Auto-generate a structured meeting agenda
- Recommend best meeting times`,
		Example: `  # Analyze an email thread
  nylas calendar ai analyze-thread --thread thread_abc123

  # Include auto-generated agenda
  nylas calendar ai analyze-thread --thread thread_abc123 --agenda

  # Include meeting time suggestions
  nylas calendar ai analyze-thread --thread thread_abc123 --time

  # Analyze and create meeting directly
  nylas calendar ai analyze-thread --thread thread_abc123 --create-meeting

  # Use specific AI provider
  nylas calendar ai analyze-thread --thread thread_abc123 --provider claude

  # Output as JSON
  nylas calendar ai analyze-thread --thread thread_abc123 --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get thread ID from args if not provided via flag
			if threadID == "" && len(args) > 0 {
				threadID = args[0]
			}

			if threadID == "" {
				return fmt.Errorf("thread ID is required (use --thread flag or provide as argument)")
			}

			// Get Nylas client
			client, err := getClient()
			if err != nil {
				return fmt.Errorf("create client: %w", err)
			}

			// Get grant ID
			grantID, err := getGrantID(args)
			if err != nil {
				return fmt.Errorf("get grant ID: %w", err)
			}

			// Get LLM router
			llmRouter, err := getLLMRouter()
			if err != nil {
				return fmt.Errorf("create LLM router: %w", err)
			}

			// Create email analyzer
			analyzer := ai.NewEmailAnalyzer(client, llmRouter)

			// Create analysis request
			req := &domain.EmailAnalysisRequest{
				ThreadID:      threadID,
				IncludeAgenda: includeAgenda,
				IncludeTime:   includeTime,
			}

			// Show progress
			fmt.Printf("🤖 AI Email Thread Analysis\n\n")
			if provider != "" {
				fmt.Printf("Using provider: %s\n", provider)
			}
			fmt.Printf("Analyzing email thread...\n")

			// Analyze thread
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			analysis, err := analyzer.AnalyzeThread(ctx, grantID, threadID, req)
			if err != nil {
				return fmt.Errorf("analyze thread: %w", err)
			}

			// Output results
			if jsonOutput {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(analysis)
			}

			displayThreadAnalysis(analysis)

			// Create meeting if requested
			if createMeeting {
				fmt.Println("\n---")
				return createMeetingFromAnalysis(ctx, client, grantID, analysis)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&threadID, "thread", "", "Email thread ID to analyze (required)")
	cmd.Flags().BoolVar(&includeAgenda, "agenda", false, "Generate meeting agenda")
	cmd.Flags().BoolVar(&includeTime, "time", false, "Suggest best meeting time")
	cmd.Flags().BoolVar(&createMeeting, "create-meeting", false, "Create meeting event after analysis")
	cmd.Flags().StringVar(&provider, "provider", "", "AI provider to use (ollama, claude, openai, groq)")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}

// displayThreadAnalysis displays the thread analysis in a formatted way.
func displayThreadAnalysis(analysis *domain.EmailThreadAnalysis) {
	fmt.Printf("\n📧 Thread Analysis Results\n\n")

	// Subject and basic info
	if analysis.Subject != "" {
		fmt.Printf("Subject: %s\n", analysis.Subject)
	}
	fmt.Printf("Messages: %d\n", analysis.MessageCount)
	fmt.Printf("Participants: %d\n", analysis.ParticipantCount)
	fmt.Println()

	// Purpose
	if analysis.Purpose != "" {
		fmt.Printf("📋 Meeting Purpose:\n")
		fmt.Printf("  %s\n\n", analysis.Purpose)
	}

	// Topics
	if len(analysis.Topics) > 0 {
		fmt.Printf("🎯 Key Topics:\n")
		for _, topic := range analysis.Topics {
			fmt.Printf("  • %s\n", topic)
		}
		fmt.Println()
	}

	// Priority
	fmt.Printf("⏱️  Priority: %s", strings.ToUpper(string(analysis.Priority)))
	if len(analysis.UrgencyIndicators) > 0 {
		fmt.Printf("\n  Urgency Indicators:\n")
		for _, indicator := range analysis.UrgencyIndicators {
			fmt.Printf("  • %s\n", indicator)
		}
	}
	fmt.Println()

	// Duration
	fmt.Printf("⌚ Suggested Duration: %d minutes\n\n", analysis.SuggestedDuration)

	// Participants
	if len(analysis.Participants) > 0 {
		fmt.Printf("👥 Participant Analysis:\n\n")

		// Group by involvement
		required := []domain.ParticipantInfo{}
		optional := []domain.ParticipantInfo{}

		for _, p := range analysis.Participants {
			if p.Required {
				required = append(required, p)
			} else {
				optional = append(optional, p)
			}
		}

		if len(required) > 0 {
			fmt.Printf("  Required Attendees:\n")
			for _, p := range required {
				displayParticipant(p)
			}
			fmt.Println()
		}

		if len(optional) > 0 {
			fmt.Printf("  Optional Attendees:\n")
			for _, p := range optional {
				displayParticipant(p)
			}
			fmt.Println()
		}
	}

	// Agenda
	if analysis.Agenda != nil {
		displayAgenda(analysis.Agenda)
	}

	// Meeting time suggestion
	if analysis.BestMeetingTime != nil {
		fmt.Printf("🌍 Recommended Meeting Time:\n")
		fmt.Printf("  Time: %s %s\n", analysis.BestMeetingTime.Time, analysis.BestMeetingTime.Timezone)
		fmt.Printf("  Score: %d/100\n", analysis.BestMeetingTime.Score)
		if analysis.BestMeetingTime.Reasoning != "" {
			fmt.Printf("  Reasoning: %s\n", analysis.BestMeetingTime.Reasoning)
		}
		fmt.Println()
	}
}

// displayParticipant displays participant information.
func displayParticipant(p domain.ParticipantInfo) {
	name := p.Email
	if p.Name != "" {
		name = fmt.Sprintf("%s <%s>", p.Name, p.Email)
	}

	involvement := ""
	switch p.Involvement {
	case domain.InvolvementHigh:
		involvement = "⭐ High"
	case domain.InvolvementMedium:
		involvement = "🔸 Medium"
	case domain.InvolvementLow:
		involvement = "🔹 Low"
	}

	fmt.Printf("  • %s\n", name)
	fmt.Printf("    Involvement: %s (%d messages, %d mentions)\n", involvement, p.MessageCount, p.MentionCount)
}

// displayAgenda displays the meeting agenda.
func displayAgenda(agenda *domain.MeetingAgenda) {
	fmt.Printf("📝 Auto-Generated Meeting Agenda\n\n")

	if agenda.Title != "" {
		fmt.Printf("## %s\n", agenda.Title)
	} else {
		fmt.Printf("## Meeting Agenda\n")
	}

	if agenda.Duration > 0 {
		fmt.Printf("Duration: %d minutes\n", agenda.Duration)
	}
	fmt.Println()

	for i, item := range agenda.Items {
		fmt.Printf("%d. %s", i+1, item.Title)
		if item.Duration > 0 {
			fmt.Printf(" (%d min)", item.Duration)
		}
		fmt.Println()

		if item.Description != "" {
			fmt.Printf("   %s\n", item.Description)
		}

		if item.Source != "" {
			fmt.Printf("   [From email: \"%s\"]\n", item.Source)
		}

		if item.Owner != "" {
			fmt.Printf("   Owner: %s\n", item.Owner)
		}

		if item.Decision {
			fmt.Printf("   ⚠️  Decision required\n")
		}

		fmt.Println()
	}

	if len(agenda.Notes) > 0 {
		fmt.Printf("Notes:\n")
		for _, note := range agenda.Notes {
			fmt.Printf("  • %s\n", note)
		}
		fmt.Println()
	}
}

// createMeetingFromAnalysis creates a calendar event based on the analysis.
func createMeetingFromAnalysis(ctx context.Context, client any, grantID string, analysis *domain.EmailThreadAnalysis) error {
	fmt.Printf("Creating meeting from analysis...\n\n")

	// Extract required participants
	participants := []domain.EmailParticipant{}
	for _, p := range analysis.Participants {
		if p.Required {
			participants = append(participants, domain.EmailParticipant{
				Email: p.Email,
				Name:  p.Name,
			})
		}
	}

	// Build event title
	title := analysis.Subject
	if title == "" && analysis.Agenda != nil {
		title = analysis.Agenda.Title
	}
	if title == "" {
		title = "Meeting"
	}

	// Build event description from agenda
	description := analysis.Purpose
	if analysis.Agenda != nil {
		description += "\n\nAgenda:\n"
		for i, item := range analysis.Agenda.Items {
			description += fmt.Sprintf("%d. %s", i+1, item.Title)
			if item.Duration > 0 {
				description += fmt.Sprintf(" (%d min)", item.Duration)
			}
			description += "\n"
			if item.Description != "" {
				description += fmt.Sprintf("   %s\n", item.Description)
			}
		}
	}

	fmt.Printf("📅 Meeting Details:\n")
	fmt.Printf("  Title: %s\n", title)
	fmt.Printf("  Duration: %d minutes\n", analysis.SuggestedDuration)
	fmt.Printf("  Participants: %d required attendees\n", len(participants))
	fmt.Println()

	fmt.Printf("💡 To create this meeting, use:\n")
	fmt.Printf("  nylas calendar events create \\\n")
	fmt.Printf("    --title \"%s\" \\\n", title)
	fmt.Printf("    --duration %dm \\\n", analysis.SuggestedDuration)

	if len(participants) > 0 {
		emails := make([]string, len(participants))
		for i, p := range participants {
			emails[i] = p.Email
		}
		fmt.Printf("    --participants \"%s\" \\\n", strings.Join(emails, ","))
	}

	if analysis.BestMeetingTime != nil {
		fmt.Printf("    --when \"%s\" \\\n", analysis.BestMeetingTime.Time)
		fmt.Printf("    --timezone \"%s\"\n", analysis.BestMeetingTime.Timezone)
	} else {
		fmt.Printf("    --when \"<time>\"\n")
	}

	return nil
}
