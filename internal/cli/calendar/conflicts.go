package calendar

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/mqasimca/nylas/internal/adapters/analytics"
	"github.com/mqasimca/nylas/internal/cli/common"
	"github.com/mqasimca/nylas/internal/domain"
)

func newConflictsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "conflicts",
		Short: "Detect and resolve scheduling conflicts",
		Long: `Analyze proposed meetings for scheduling conflicts and get AI-powered resolution suggestions.

Detects:
- Hard conflicts (overlapping meetings)
- Soft conflicts (back-to-back, focus time interruption, meeting overload)
- Suggests alternative times with scoring`,
	}

	cmd.AddCommand(newCheckConflictsCmd())

	return cmd
}

func newCheckConflictsCmd() *cobra.Command {
	var (
		title        string
		startTime    string
		endTime      string
		duration     int
		participants []string
		autoResolve  bool
	)

	cmd := &cobra.Command{
		Use:   "check",
		Short: "Check for conflicts with a proposed meeting",
		Example: `  # Check conflicts for a new meeting
  nylas calendar conflicts check \
    --title "Product Review" \
    --start "2025-01-22T14:00:00Z" \
    --duration 60 \
    --participants team@company.com

  # Check and auto-select best alternative
  nylas calendar conflicts check \
    --title "Team Sync" \
    --start "2025-01-23T10:00:00Z" \
    --duration 30 \
    --auto-resolve`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return common.WrapGetError("client", err)
			}

			grantID, err := getGrantID(args)
			if err != nil {
				return common.WrapGetError("grant ID", err)
			}

			// Parse start time
			start, err := time.Parse(time.RFC3339, startTime)
			if err != nil {
				return fmt.Errorf("invalid start time (use RFC3339 format): %w", err)
			}

			// Calculate end time
			var end time.Time
			if endTime != "" {
				end, err = time.Parse(time.RFC3339, endTime)
				if err != nil {
					return fmt.Errorf("invalid end time (use RFC3339 format): %w", err)
				}
			} else {
				end = start.Add(time.Duration(duration) * time.Minute)
			}

			// Create proposed event
			proposedEvent := &domain.Event{
				Title: title,
				When: domain.EventWhen{
					StartTime: start.Unix(),
					EndTime:   end.Unix(),
				},
			}

			// Add participants
			for _, email := range participants {
				proposedEvent.Participants = append(proposedEvent.Participants, domain.Participant{
					Person: domain.Person{Email: email},
				})
			}

			// Pattern analysis can take time - use longer timeout
			ctx, cancel := common.CreateContextWithTimeout(domain.TimeoutAI)
			defer cancel()

			// Analyze patterns first
			fmt.Println("🔍 Analyzing your calendar patterns...")
			learner := analytics.NewPatternLearner(client)
			analysis, err := learner.AnalyzeHistory(ctx, grantID, 90)
			if err != nil {
				fmt.Printf("⚠️  Could not analyze patterns: %v\n", err)
			}

			// Create conflict resolver
			var patterns *domain.MeetingPattern
			if analysis != nil && analysis.Patterns != nil {
				patterns = analysis.Patterns
			}

			resolver := analytics.NewConflictResolver(client, patterns)

			// Detect conflicts
			fmt.Println("\n⚙️  Detecting conflicts...")
			conflicts, err := resolver.DetectConflicts(ctx, grantID, proposedEvent, patterns)
			if err != nil {
				return common.WrapGetError("conflicts", err)
			}

			// Display results
			displayConflicts(conflicts)

			// Handle auto-resolve
			if autoResolve && len(conflicts.AlternativeTimes) > 0 {
				fmt.Println("\n🤖 Auto-selecting best alternative...")
				best := conflicts.AlternativeTimes[0]
				fmt.Printf("✓ Selected: %s (Score: %d/100)\n",
					best.ProposedTime.Format("Mon, Jan 2, 2006 at 3:04 PM MST"),
					best.Score)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "Meeting title (required)")
	cmd.Flags().StringVar(&startTime, "start", "", "Start time in RFC3339 format (required)")
	cmd.Flags().StringVar(&endTime, "end", "", "End time in RFC3339 format (optional, uses --duration if not set)")
	cmd.Flags().IntVar(&duration, "duration", 60, "Duration in minutes")
	cmd.Flags().StringSliceVar(&participants, "participants", nil, "Participant email addresses")
	cmd.Flags().BoolVar(&autoResolve, "auto-resolve", false, "Automatically select best alternative")

	_ = cmd.MarkFlagRequired("title")
	_ = cmd.MarkFlagRequired("start")

	return cmd
}

func displayConflicts(analysis *domain.ConflictAnalysis) {
	fmt.Println("\n📊 Conflict Analysis")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Hard conflicts
	if len(analysis.HardConflicts) > 0 {
		fmt.Printf("\n🔴 Hard Conflicts (%d)\n", len(analysis.HardConflicts))
		for i, conflict := range analysis.HardConflicts {
			fmt.Printf("\n%d. %s\n", i+1, conflict.Description)
			if conflict.ConflictingEvent != nil {
				eventTime := time.Unix(conflict.ConflictingEvent.When.StartTime, 0)
				fmt.Printf("   Event: %s\n", conflict.ConflictingEvent.Title)
				fmt.Printf("   Time: %s\n", eventTime.Format("Mon, Jan 2 at 3:04 PM MST"))
				fmt.Printf("   Status: %s\n", conflict.ConflictingEvent.Status)
			}
			fmt.Printf("   Impact: %s\n", conflict.Impact)
			fmt.Printf("   Suggestion: %s\n", conflict.Suggestion)
		}
	}

	// Soft conflicts
	if len(analysis.SoftConflicts) > 0 {
		fmt.Printf("\n🟡 Soft Conflicts (%d)\n", len(analysis.SoftConflicts))
		for i, conflict := range analysis.SoftConflicts {
			icon := "•"
			switch conflict.Type {
			case domain.ConflictTypeSoftFocusTime:
				icon = "🎯"
			case domain.ConflictTypeSoftBackToBack:
				icon = "⏱️"
			case domain.ConflictTypeSoftOverload:
				icon = "📈"
			}

			fmt.Printf("\n%d. %s %s\n", i+1, icon, conflict.Description)
			fmt.Printf("   Severity: %s\n", conflict.Severity)
			fmt.Printf("   Impact: %s\n", conflict.Impact)
			if conflict.CanAutoResolve {
				fmt.Printf("   ✓ Can auto-resolve\n")
			}
		}
	}

	// No conflicts
	if len(analysis.HardConflicts) == 0 && len(analysis.SoftConflicts) == 0 {
		fmt.Println("\n✅ No conflicts detected!")
		fmt.Println("   This is a good time for the meeting.")
	}

	// Recommendations
	if len(analysis.Recommendations) > 0 {
		fmt.Println("\n💡 Recommendations")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		for _, rec := range analysis.Recommendations {
			fmt.Printf("  %s\n", rec)
		}
	}

	// Alternative times
	if len(analysis.AlternativeTimes) > 0 {
		fmt.Println("\n🔄 Suggested Alternative Times")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		for i, alt := range analysis.AlternativeTimes {
			// Score color coding
			scoreIcon := "🟢"
			if alt.Score < 50 {
				scoreIcon = "🔴"
			} else if alt.Score < 75 {
				scoreIcon = "🟡"
			}

			fmt.Printf("\n%d. %s %s (Score: %d/100)\n",
				i+1,
				scoreIcon,
				alt.ProposedTime.Format("Mon, Jan 2, 2006 at 3:04 PM MST"),
				alt.Score)

			if len(alt.Pros) > 0 {
				fmt.Println("\n   Pros:")
				for _, pro := range alt.Pros {
					fmt.Printf("   • %s\n", pro)
				}
			}

			if len(alt.Cons) > 0 {
				fmt.Println("\n   Cons:")
				for _, con := range alt.Cons {
					fmt.Printf("   • %s\n", con)
				}
			}

			if alt.AIInsight != "" {
				fmt.Printf("\n   💡 %s\n", alt.AIInsight)
			}

			if len(alt.Conflicts) > 0 {
				fmt.Printf("\n   ⚠️  %d soft conflict(s) remain\n", len(alt.Conflicts))
			}
		}
	}

	// AI Recommendation
	if analysis.AIRecommendation != "" {
		fmt.Println("\n🤖 AI Recommendation")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("%s\n", analysis.AIRecommendation)
	}

	// Status
	fmt.Println("\n" + strings.Repeat("━", 40))
	if analysis.CanProceed {
		if len(analysis.SoftConflicts) > 0 {
			fmt.Println("⚠️  Status: Can proceed (with soft conflicts)")
		} else {
			fmt.Println("✅ Status: Clear to proceed")
		}
	} else {
		fmt.Println("❌ Status: Cannot proceed (hard conflicts)")
	}
	fmt.Println()
}
