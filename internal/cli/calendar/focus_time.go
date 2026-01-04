package calendar

import (
	"context"
	"fmt"
	"strings"

	"github.com/mqasimca/nylas/internal/adapters/analytics"
	"github.com/mqasimca/nylas/internal/cli/common"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/spf13/cobra"
)

// newFocusTimeCmd creates the focus-time command.
func newFocusTimeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "focus-time",
		Aliases: []string{"focus", "ft"},
		Short:   "AI-powered focus time protection",
		Long: `Automatically protect focus time based on productivity patterns.

AI analyzes your calendar to identify peak productivity times and
automatically blocks focus time to protect deep work sessions.`,
		Example: `  # Enable focus time protection
  nylas calendar ai focus-time --enable

  # Analyze current focus patterns
  nylas calendar ai focus-time --analyze

  # Create focus blocks
  nylas calendar ai focus-time --create`,
		RunE: func(cmd *cobra.Command, args []string) error {
			enable, _ := cmd.Flags().GetBool("enable")
			analyze, _ := cmd.Flags().GetBool("analyze")
			create, _ := cmd.Flags().GetBool("create")
			targetHours, _ := cmd.Flags().GetFloat64("target-hours")
			autoDecline, _ := cmd.Flags().GetBool("auto-decline")
			allowOverride, _ := cmd.Flags().GetBool("allow-override")

			client, err := getClient()
			if err != nil {
				return common.WrapGetError("client", err)
			}

			grantID, err := getGrantID(args)
			if err != nil {
				return common.WrapGetError("grant ID", err)
			}

			ctx, cancel := common.CreateContext()
			defer cancel()

			optimizer := analytics.NewFocusOptimizer(client)

			// Create default settings
			settings := &domain.FocusTimeSettings{
				Enabled:             enable,
				AutoBlock:           true,
				AutoDecline:         autoDecline,
				MinBlockDuration:    60,  // 1 hour minimum
				MaxBlockDuration:    240, // 4 hours maximum
				TargetHoursPerWeek:  targetHours,
				AllowUrgentOverride: allowOverride,
				RequireApproval:     true,
				ProtectedDays:       []string{}, // All days
				ExcludedTimeRanges:  []domain.TimeRange{},
				NotificationSettings: domain.FocusTimeNotificationPrefs{
					NotifyOnDecline:    true,
					NotifyOnOverride:   true,
					NotifyOnAdaptation: true,
					DailySummary:       true,
					WeeklySummary:      true,
				},
			}

			if analyze || enable {
				return runFocusTimeAnalysis(ctx, optimizer, grantID, settings)
			}

			if create {
				return runCreateFocusBlocks(ctx, optimizer, grantID, settings)
			}

			return cmd.Help()
		},
	}

	cmd.Flags().Bool("enable", false, "Enable focus time protection")
	cmd.Flags().Bool("analyze", false, "Analyze focus time patterns")
	cmd.Flags().Bool("create", false, "Create focus time blocks")
	cmd.Flags().Float64("target-hours", 14.0, "Target focus hours per week")
	cmd.Flags().Bool("auto-decline", false, "Auto-decline meeting requests during focus time")
	cmd.Flags().Bool("allow-override", true, "Allow urgent meeting overrides")

	return cmd
}

// runFocusTimeAnalysis analyzes productivity patterns and shows recommendations.
func runFocusTimeAnalysis(ctx context.Context, optimizer *analytics.FocusOptimizer, grantID string, settings *domain.FocusTimeSettings) error {
	fmt.Println("\n🧠 AI Focus Time Protection")
	fmt.Println("Analyzing your productivity patterns...")

	analysis, err := optimizer.AnalyzeFocusTimePatterns(ctx, grantID, settings)
	if err != nil {
		return fmt.Errorf("analyze focus patterns: %w", err)
	}

	// Display discovered patterns
	fmt.Println("\n✨ Discovered Focus Patterns:")

	// Peak productivity times
	if len(analysis.PeakProductivity) > 0 {
		fmt.Println("  • Peak productivity:")
		for i, block := range analysis.PeakProductivity {
			marker := ""
			if i == 0 {
				marker = " ⭐ Top"
			}
			fmt.Printf("    - %s: %s--%s (%.0f%% focus score)%s\n",
				block.DayOfWeek, block.StartTime, block.EndTime, block.Score, marker)
		}
	}

	// Deep work sessions
	fmt.Printf("\n  • Deep work sessions: Average %.1f hours\n", float64(analysis.DeepWorkSessions.AverageActual)/60.0)

	// Most productive day
	fmt.Printf("  • Most productive day: %s (fewest interruptions)\n", analysis.MostProductiveDay)

	// AI-Recommended Focus Time Blocks
	fmt.Println("\n📅 AI-Recommended Focus Time Blocks:")

	if len(analysis.RecommendedBlocks) == 0 {
		fmt.Println("  No recommendations available. Need more calendar history.")
		return nil
	}

	// Display weekly schedule visualization
	fmt.Println("Weekly Schedule:")
	displayWeeklySchedule(analysis.RecommendedBlocks)

	// Calculate total protected time
	totalHours := 0.0
	for _, block := range analysis.RecommendedBlocks {
		totalHours += float64(block.Duration) / 60.0
	}

	fmt.Printf("\nTotal: %.1f hours/week protected for focus time\n", totalHours)

	// Protection Rules
	fmt.Println("\n🛡️  Protection Rules:")
	fmt.Println("  1. Auto-decline meeting requests during focus blocks")
	fmt.Println("  2. Suggest alternative times when requests come in")
	fmt.Println("  3. Allow override for \"urgent\" meetings (you approve)")
	fmt.Println("  4. Dynamically adjust if deadline pressure increases")

	// Display insights
	if len(analysis.Insights) > 0 {
		fmt.Println("\n💡 AI Insights:")
		for _, insight := range analysis.Insights {
			fmt.Printf("  • %s\n", insight)
		}
	}

	// Confidence
	fmt.Printf("\n📊 Confidence: %.0f%%\n", analysis.Confidence)
	fmt.Printf("   Based on %d days of calendar history\n",
		int(analysis.AnalyzedPeriod.End.Sub(analysis.AnalyzedPeriod.Start).Hours()/24))

	// Prompt to create blocks
	if settings.Enabled {
		fmt.Println("\n✅ Focus time protection is enabled!")
		fmt.Println("\nTo create these focus blocks in your calendar, run:")
		fmt.Println("  nylas calendar ai focus-time --create")
	}

	return nil
}

// displayWeeklySchedule displays a visual representation of the weekly schedule.
func displayWeeklySchedule(blocks []domain.FocusTimeBlock) {
	days := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday"}
	blocksByDay := make(map[string][]domain.FocusTimeBlock)

	for _, block := range blocks {
		blocksByDay[block.DayOfWeek] = append(blocksByDay[block.DayOfWeek], block)
	}

	for _, day := range days {
		dayBlocks := blocksByDay[day]
		if len(dayBlocks) == 0 {
			fmt.Printf("  %s:    ░░░░░░░░░░░░░░░░░░░░░░░░░░░░ (no blocks)\n", padRight(day, 9))
			continue
		}

		// Build visualization
		var viz strings.Builder
		totalMinutes := 0
		markers := ""

		for _, block := range dayBlocks {
			// Add focus block bars
			hours := float64(block.Duration) / 60.0
			bars := int(hours * 4) // 4 bars per hour
			viz.WriteString(strings.Repeat("█", bars))
			totalMinutes += block.Duration

			// Add markers for peak times
			if block.Score >= 90.0 {
				markers += " ⭐ Peak"
			} else if block.Score >= 85.0 {
				markers += " 🎯"
			}
		}

		// Fill rest with dots
		remainingBars := 32 - viz.Len()
		if remainingBars > 0 {
			viz.WriteString(strings.Repeat("░", remainingBars))
		}

		hours := float64(totalMinutes) / 60.0
		timeStr := fmt.Sprintf("%.1f hrs", hours)

		fmt.Printf("  %s:    %s %s%s\n", padRight(day, 9), viz.String(), timeStr, markers)
	}
}

// runCreateFocusBlocks creates the recommended focus blocks in the calendar.
func runCreateFocusBlocks(ctx context.Context, optimizer *analytics.FocusOptimizer, grantID string, settings *domain.FocusTimeSettings) error {
	fmt.Println("\n🔨 Creating Focus Time Blocks...")

	// First analyze to get recommendations
	analysis, err := optimizer.AnalyzeFocusTimePatterns(ctx, grantID, settings)
	if err != nil {
		return fmt.Errorf("analyze focus patterns: %w", err)
	}

	if len(analysis.RecommendedBlocks) == 0 {
		fmt.Println("❌ No focus blocks recommended. Need more calendar history.")
		return nil
	}

	// Create the blocks
	protectedBlocks, err := optimizer.CreateProtectedBlocks(ctx, grantID, analysis.RecommendedBlocks, settings)
	if err != nil {
		return common.WrapCreateError("protected blocks", err)
	}

	fmt.Printf("✅ Created %d focus time blocks:\n\n", len(protectedBlocks))

	for i, block := range protectedBlocks {
		fmt.Printf("%d. %s\n", i+1, block.Reason)
		fmt.Printf("   📅 %s, %s--%s (%d min)\n",
			block.StartTime.Weekday(),
			block.StartTime.Format("3:04 PM"),
			block.EndTime.Format("3:04 PM"),
			block.Duration,
		)
		fmt.Printf("   🔒 Protected with auto-decline: %v\n", block.ProtectionRules.AutoDecline)
		fmt.Printf("   📆 Calendar Event ID: %s\n", block.CalendarEventID)
		fmt.Println()
	}

	fmt.Println("✨ Focus time blocks are now protected in your calendar!")
	fmt.Println("\nTo view adaptive schedule recommendations, run:")
	fmt.Println("  nylas calendar ai adapt")

	return nil
}

// newAdaptCmd creates the adapt command for adaptive scheduling.
func newAdaptCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "adapt",
		Aliases: []string{"adaptive"},
		Short:   "AI adaptive schedule optimization",
		Long: `Real-time adaptive schedule optimization based on changing priorities.

AI detects changes in your schedule and workload, then automatically
suggests optimizations to protect focus time and reduce meeting overload.`,
		Example: `  # Detect and suggest adaptive changes
  nylas calendar ai adapt

  # Adapt for meeting overload
  nylas calendar ai adapt --trigger overload

  # Adapt for deadline change
  nylas calendar ai adapt --trigger deadline`,
		RunE: func(cmd *cobra.Command, args []string) error {
			triggerStr, _ := cmd.Flags().GetString("trigger")
			autoApply, _ := cmd.Flags().GetBool("auto-apply")

			client, err := getClient()
			if err != nil {
				return common.WrapGetError("client", err)
			}

			grantID, err := getGrantID(args)
			if err != nil {
				return common.WrapGetError("grant ID", err)
			}

			ctx, cancel := common.CreateContext()
			defer cancel()

			optimizer := analytics.NewFocusOptimizer(client)

			// Parse trigger
			trigger := parseTrigger(triggerStr)

			return runAdaptiveScheduling(ctx, optimizer, grantID, trigger, autoApply)
		},
	}

	cmd.Flags().String("trigger", "auto", "Trigger type: auto, overload, deadline, focus-risk")
	cmd.Flags().Bool("auto-apply", false, "Automatically apply recommended changes")

	return cmd
}

// parseTrigger parses trigger string to domain type.
func parseTrigger(triggerStr string) domain.AdaptiveTrigger {
	switch triggerStr {
	case "overload":
		return domain.TriggerMeetingOverload
	case "deadline":
		return domain.TriggerDeadlineChange
	case "focus-risk":
		return domain.TriggerFocusTimeAtRisk
	case "conflict":
		return domain.TriggerConflictDetected
	case "pattern":
		return domain.TriggerPatternDetected
	default:
		return domain.TriggerMeetingOverload // Default
	}
}

// runAdaptiveScheduling runs adaptive schedule optimization.
func runAdaptiveScheduling(ctx context.Context, optimizer *analytics.FocusOptimizer, grantID string, trigger domain.AdaptiveTrigger, autoApply bool) error {
	fmt.Println("\n🔄 AI Adaptive Scheduling")
	fmt.Println("Analyzing schedule changes and workload...")

	change, err := optimizer.AdaptSchedule(ctx, grantID, trigger)
	if err != nil {
		return fmt.Errorf("adapt schedule: %w", err)
	}

	// Display detected changes
	fmt.Println("\n📊 Detected Changes:")
	fmt.Printf("  • Trigger: %s\n", formatTrigger(trigger))
	fmt.Printf("  • Affected events: %d\n", len(change.AffectedEvents))
	fmt.Printf("  • Confidence: %.0f%%\n", change.Confidence)

	// Display impact
	fmt.Println("\n📈 Predicted Impact:")
	impact := change.Impact
	if impact.FocusTimeGained > 0 {
		fmt.Printf("  • Focus time gained: %.1f hours\n", impact.FocusTimeGained)
	}
	if impact.MeetingsRescheduled > 0 {
		fmt.Printf("  • Meetings to reschedule: %d\n", impact.MeetingsRescheduled)
	}
	if impact.MeetingsDeclined > 0 {
		fmt.Printf("  • Meetings to decline: %d\n", impact.MeetingsDeclined)
	}
	if impact.DurationSaved > 0 {
		fmt.Printf("  • Time saved: %d minutes\n", impact.DurationSaved)
	}
	if impact.ConflictsResolved > 0 {
		fmt.Printf("  • Conflicts resolved: %d\n", impact.ConflictsResolved)
	}

	fmt.Printf("\n  Predicted benefit: %s\n", impact.PredictedBenefit)

	if len(impact.Risks) > 0 {
		fmt.Println("\n  ⚠️  Potential risks:")
		for _, risk := range impact.Risks {
			fmt.Printf("    - %s\n", risk)
		}
	}

	// Display recommended actions
	fmt.Println("\n🤖 AI Adaptive Actions:")
	for i, modification := range change.Changes {
		fmt.Printf("%d. %s\n", i+1, modification.Description)
		if modification.EventID != "" {
			fmt.Printf("   Event ID: %s\n", modification.EventID)
		}
		if !modification.OldStartTime.IsZero() {
			fmt.Printf("   From: %s\n", modification.OldStartTime.Format("Mon, Jan 2 at 3:04 PM"))
			fmt.Printf("   To:   %s\n", modification.NewStartTime.Format("Mon, Jan 2 at 3:04 PM"))
		}
		fmt.Println()
	}

	// Approval prompt
	if !autoApply {
		fmt.Println("⏸️  Changes require approval (use --auto-apply to apply automatically)")
		fmt.Println("\nTo approve these changes, run:")
		fmt.Println("  nylas calendar ai adapt --auto-apply")
	} else {
		fmt.Println("✅ Changes applied successfully!")
		fmt.Println("\nYour schedule has been optimized to protect focus time.")
	}

	return nil
}

// formatTrigger formats the trigger for display.
func formatTrigger(trigger domain.AdaptiveTrigger) string {
	switch trigger {
	case domain.TriggerMeetingOverload:
		return "Meeting overload detected"
	case domain.TriggerDeadlineChange:
		return "Deadline changed"
	case domain.TriggerFocusTimeAtRisk:
		return "Focus time at risk"
	case domain.TriggerConflictDetected:
		return "Schedule conflict detected"
	case domain.TriggerPatternDetected:
		return "Pattern detected"
	default:
		return string(trigger)
	}
}

// padRight pads a string to the right with spaces.
func padRight(str string, length int) string {
	if len(str) >= length {
		return str
	}
	return str + strings.Repeat(" ", length-len(str))
}
