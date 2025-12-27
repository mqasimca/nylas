package calendar

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mqasimca/nylas/internal/adapters/analytics"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/ports"
	"github.com/spf13/cobra"
)

func newAnalyzeCmd() *cobra.Command {
	var (
		days         int
		applyRecs    bool
		scoreTime    string
		participants []string
		duration     int
	)

	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "Analyze meeting patterns and get AI recommendations",
		Long: `Analyze historical meeting data to learn patterns and provide AI-powered recommendations.

This command analyzes your calendar history to discover:
- Meeting acceptance patterns (by day/time)
- Actual vs scheduled meeting durations
- Timezone preferences for cross-TZ meetings
- Productivity insights (peak focus times)
- Per-participant preferences

It provides actionable AI recommendations for optimizing your calendar.`,
		Example: `  # Analyze last 90 days
  nylas calendar analyze

  # Analyze last 60 days
  nylas calendar analyze --days 60

  # Score a specific meeting time
  nylas calendar analyze --score-time "2025-01-15T14:00:00Z" --participants user@example.com --duration 30

  # Apply top recommendations automatically
  nylas calendar analyze --apply`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return fmt.Errorf("failed to get client: %w", err)
			}

			grantID, err := getGrantID(args)
			if err != nil {
				return fmt.Errorf("failed to get grant ID: %w", err)
			}

			// AI analysis can take time - use longer timeout
			ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
			defer cancel()

			// Load config to get working hours - respect --config flag
			configStore := getConfigStore(cmd)
			cfg, _ := configStore.Load()

			// Get working hours from config (use default if not set)
			var workingHours *domain.DaySchedule
			if cfg != nil && cfg.WorkingHours != nil && cfg.WorkingHours.Default != nil {
				workingHours = cfg.WorkingHours.Default
			}

			// Create pattern learner with working hours
			learner := analytics.NewPatternLearnerWithWorkingHours(client, workingHours)

			// If scoring a specific time
			if scoreTime != "" {
				return scoreSpecificTime(ctx, learner, client, grantID, scoreTime, participants, duration)
			}

			// Analyze historical patterns
			fmt.Printf("🔍 Analyzing %d days of meeting history...\n\n", days)

			analysis, err := learner.AnalyzeHistory(ctx, grantID, days)
			if err != nil {
				return fmt.Errorf("failed to analyze history: %w", err)
			}

			// Display results
			displayAnalysis(analysis, workingHours)

			// Apply recommendations if requested
			if applyRecs {
				return applyRecommendations(ctx, client, grantID, analysis)
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&days, "days", 90, "Number of days to analyze")
	cmd.Flags().BoolVar(&applyRecs, "apply", false, "Apply top recommendations automatically")
	cmd.Flags().StringVar(&scoreTime, "score-time", "", "Score a specific meeting time (RFC3339 format)")
	cmd.Flags().StringSliceVar(&participants, "participants", nil, "Participants for scoring (email addresses)")
	cmd.Flags().IntVar(&duration, "duration", 30, "Meeting duration in minutes for scoring")

	return cmd
}

func displayAnalysis(analysis *domain.MeetingAnalysis, workingHours *domain.DaySchedule) {
	fmt.Printf("📊 Analysis Period: %s to %s\n",
		analysis.Period.Start.Format("2006-01-02"),
		analysis.Period.End.Format("2006-01-02"))
	fmt.Printf("📅 Total Meetings Analyzed: %d\n\n", analysis.TotalMeetings)

	if analysis.Patterns == nil {
		fmt.Println("⚠️  Insufficient data for pattern analysis")
		return
	}

	patterns := analysis.Patterns

	// Get working hours range (default 9-17)
	startHour, endHour := 9, 17
	if workingHours != nil && workingHours.Enabled {
		if workingHours.Start != "" {
			var h, m int
			if _, err := fmt.Sscanf(workingHours.Start, "%d:%d", &h, &m); err == nil {
				startHour = h
				if m > 0 {
					startHour = h + 1
				}
			}
		}
		if workingHours.End != "" {
			var h, m int
			if _, err := fmt.Sscanf(workingHours.End, "%d:%d", &h, &m); err == nil {
				endHour = h
			}
		}
	}

	// Acceptance Patterns
	fmt.Println("✅ Meeting Acceptance Patterns")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("Overall Acceptance Rate: %.1f%%\n\n", patterns.Acceptance.Overall*100)

	if len(patterns.Acceptance.ByDayOfWeek) > 0 {
		fmt.Println("By Day of Week:")
		days := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
		for _, day := range days {
			if rate, exists := patterns.Acceptance.ByDayOfWeek[day]; exists {
				bar := strings.Repeat("█", int(rate*20))
				fmt.Printf("  %9s: %.1f%% %s\n", day, rate*100, bar)
			}
		}
		fmt.Println()
	}

	if len(patterns.Acceptance.ByTimeOfDay) > 0 {
		fmt.Println("By Time of Day (working hours):")
		for hour := startHour; hour <= endHour; hour++ {
			hourStr := fmt.Sprintf("%02d:00", hour)
			if rate, exists := patterns.Acceptance.ByTimeOfDay[hourStr]; exists {
				bar := strings.Repeat("█", int(rate*20))
				fmt.Printf("  %s: %.1f%% %s\n", hourStr, rate*100, bar)
			}
		}
		fmt.Println()
	}

	// Duration Patterns
	if patterns.Duration.Overall.AverageScheduled > 0 {
		fmt.Println("⏱️  Meeting Duration Patterns")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("Average Scheduled: %d minutes\n", patterns.Duration.Overall.AverageScheduled)
		fmt.Printf("Average Actual: %d minutes\n", patterns.Duration.Overall.AverageActual)
		fmt.Printf("Overrun Rate: %.1f%%\n\n", patterns.Duration.Overall.OverrunRate*100)
	}

	// Timezone Patterns
	if len(patterns.Timezone.Distribution) > 0 {
		fmt.Println("🌍 Timezone Distribution")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		for tz, count := range patterns.Timezone.Distribution {
			fmt.Printf("  %s: %d meetings\n", tz, count)
		}
		fmt.Println()
	}

	// Productivity Patterns
	if len(patterns.Productivity.PeakFocus) > 0 {
		fmt.Println("🎯 Productivity Insights")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println("Peak Focus Times (recommended for deep work):")
		for i, block := range patterns.Productivity.FocusBlocks {
			if i >= 5 {
				break
			}
			fmt.Printf("  %d. %s %s-%s (score: %.0f/100)\n",
				i+1, block.DayOfWeek, block.StartTime, block.EndTime, block.Score)
		}
		fmt.Println()

		if len(patterns.Productivity.MeetingDensity) > 0 {
			fmt.Println("Meeting Density by Day:")
			days := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday"}
			for _, day := range days {
				if density, exists := patterns.Productivity.MeetingDensity[day]; exists {
					fmt.Printf("  %9s: %.1f meetings/day\n", day, density)
				}
			}
			fmt.Println()
		}
	}

	// Recommendations
	if len(analysis.Recommendations) > 0 {
		fmt.Println("💡 AI Recommendations")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		for i, rec := range analysis.Recommendations {
			priorityIcon := "🔵"
			switch rec.Priority {
			case "high":
				priorityIcon = "🔴"
			case "medium":
				priorityIcon = "🟡"
			}

			fmt.Printf("%d. %s %s [%s]\n", i+1, priorityIcon, rec.Title, rec.Type)
			fmt.Printf("   %s\n", rec.Description)
			fmt.Printf("   📌 Action: %s\n", rec.Action)
			fmt.Printf("   📈 Impact: %s\n", rec.Impact)
			fmt.Printf("   🎯 Confidence: %.0f%%\n\n", rec.Confidence)
		}
	}

	// Insights
	if len(analysis.Insights) > 0 {
		fmt.Println("📝 Key Insights")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		for i, insight := range analysis.Insights {
			fmt.Printf("%d. %s\n", i+1, insight)
		}
		fmt.Println()
	}
}

func scoreSpecificTime(ctx context.Context, learner *analytics.PatternLearner, client ports.NylasClient, grantID, timeStr string, participants []string, duration int) error {
	// Parse the time
	proposedTime, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return fmt.Errorf("invalid time format (use RFC3339): %w", err)
	}

	// Analyze history to get patterns
	fmt.Println("🔍 Analyzing historical patterns...")
	analysis, err := learner.AnalyzeHistory(ctx, grantID, 90)
	if err != nil {
		return fmt.Errorf("failed to analyze history: %w", err)
	}

	if analysis.Patterns == nil {
		return fmt.Errorf("insufficient historical data for scoring")
	}

	// Create scorer and score the time
	scorer := analytics.NewMeetingScorer(analysis.Patterns)
	score := scorer.ScoreMeetingTime(proposedTime, participants, duration)

	// Display score
	fmt.Printf("\n🎯 Meeting Score for %s\n", proposedTime.Format("Monday, Jan 2, 2006 at 3:04 PM MST"))
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Score with color coding
	scoreColor := "🟢"
	if score.Score < 50 {
		scoreColor = "🔴"
	} else if score.Score < 70 {
		scoreColor = "🟡"
	}

	scoreBar := strings.Repeat("█", score.Score/5)
	fmt.Printf("\n%s Overall Score: %d/100\n", scoreColor, score.Score)
	fmt.Printf("   %s\n\n", scoreBar)

	fmt.Printf("🎯 Confidence: %.0f%%\n", score.Confidence)
	fmt.Printf("📊 Historical Success Rate: %.0f%%\n\n", score.SuccessRate*100)

	// Factors
	if len(score.Factors) > 0 {
		fmt.Println("Contributing Factors:")
		for _, factor := range score.Factors {
			impactIcon := "➕"
			if factor.Impact < 0 {
				impactIcon = "➖"
			} else if factor.Impact == 0 {
				impactIcon = "⚪"
			}

			fmt.Printf("  %s %s: %+d\n", impactIcon, factor.Name, factor.Impact)
			fmt.Printf("     %s\n", factor.Description)
		}
		fmt.Println()
	}

	// Recommendation
	fmt.Printf("💡 %s\n\n", score.Recommendation)

	// Alternative times
	if len(score.AlternativeTimes) > 0 {
		fmt.Println("🔄 Suggested Alternative Times:")
		for i, altTime := range score.AlternativeTimes {
			fmt.Printf("  %d. %s\n", i+1, altTime.Format("Monday, Jan 2, 2006 at 3:04 PM MST"))
		}
		fmt.Println()
	}

	return nil
}

func applyRecommendations(ctx context.Context, client ports.NylasClient, grantID string, analysis *domain.MeetingAnalysis) error {
	if len(analysis.Recommendations) == 0 {
		fmt.Println("⚠️  No recommendations to apply")
		return nil
	}

	fmt.Println("\n🚀 Applying Top Recommendations")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	applied := 0
	for _, rec := range analysis.Recommendations {
		// Only apply high priority focus time recommendations
		if rec.Priority == "high" && rec.Type == "focus_time" {
			fmt.Printf("📌 %s\n", rec.Title)

			// Parse the recommendation to extract time block
			// This is a simplified implementation - production would parse the title more robustly
			if strings.Contains(rec.Title, "Block") {
				fmt.Printf("   ℹ️  To apply this, create a recurring event manually:\n")
				fmt.Printf("      nylas calendar events create --title \"Focus Time\" --description \"%s\"\n\n", rec.Description)
				applied++
			}
		}

		if applied >= 3 {
			break
		}
	}

	if applied == 0 {
		fmt.Println("⚠️  No auto-applicable recommendations found")
		fmt.Println("   Review recommendations above and apply them manually")
	} else {
		fmt.Printf("✅ Provided instructions for %d recommendations\n", applied)
	}

	return nil
}
