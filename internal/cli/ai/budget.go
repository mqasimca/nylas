package ai

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// BudgetConfig represents AI budget configuration
type BudgetConfig struct {
	MonthlyLimit float64 `json:"monthly_limit"` // Monthly spending limit in USD
	AlertAt      float64 `json:"alert_at"`      // Alert when spending reaches this percentage (0-100)
	Enabled      bool    `json:"enabled"`       // Whether budget enforcement is enabled
}

func newSetBudgetCmd() *cobra.Command {
	var monthly float64
	var alertAt float64
	var disable bool

	cmd := &cobra.Command{
		Use:   "set-budget",
		Short: "Set monthly AI usage budget",
		Long: `Configure monthly spending limits for AI provider usage.

The budget applies to cloud AI providers (Claude, OpenAI, Groq, OpenRouter).
Ollama (local) usage is free and not counted toward the budget.

Examples:
  # Set monthly budget to $50
  nylas ai set-budget --monthly 50

  # Set budget with 80% alert threshold
  nylas ai set-budget --monthly 50 --alert-at 80

  # Disable budget enforcement
  nylas ai set-budget --disable`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if disable {
				if err := disableBudget(); err != nil {
					return err
				}
				fmt.Println("✓ Budget enforcement disabled")
				return nil
			}

			if monthly <= 0 {
				return fmt.Errorf("monthly budget must be greater than 0")
			}

			if alertAt < 0 || alertAt > 100 {
				return fmt.Errorf("alert threshold must be between 0 and 100")
			}

			// Default alert at 80%
			if alertAt == 0 {
				alertAt = 80
			}

			budget := BudgetConfig{
				MonthlyLimit: monthly,
				AlertAt:      alertAt,
				Enabled:      true,
			}

			if err := saveBudgetConfig(budget); err != nil {
				return err
			}

			fmt.Printf("✓ Monthly budget set to $%.2f\n", monthly)
			fmt.Printf("  Alert threshold: %.0f%%\n", alertAt)
			fmt.Println()
			fmt.Println("Budget applies to:")
			fmt.Println("  - Claude (Anthropic)")
			fmt.Println("  - OpenAI")
			fmt.Println("  - Groq")
			fmt.Println("  - OpenRouter")
			fmt.Println()
			fmt.Println("Ollama (local) usage is free and not counted.")

			return nil
		},
	}

	cmd.Flags().Float64Var(&monthly, "monthly", 0, "Monthly budget in USD")
	cmd.Flags().Float64Var(&alertAt, "alert-at", 80, "Alert when spending reaches this percentage (0-100)")
	cmd.Flags().BoolVar(&disable, "disable", false, "Disable budget enforcement")

	return cmd
}

func newShowBudgetCmd() *cobra.Command {
	var jsonOutput bool

	return &cobra.Command{
		Use:   "show-budget",
		Short: "Show current AI budget configuration",
		Long:  "Display the current monthly budget and spending limits for AI usage.",
		RunE: func(cmd *cobra.Command, args []string) error {
			budget, err := loadBudgetConfig()
			if err != nil {
				fmt.Println("No budget configured")
				fmt.Println()
				fmt.Println("To set a budget:")
				fmt.Println("  nylas ai set-budget --monthly 50")
				return nil
			}

			if jsonOutput {
				data, err := json.MarshalIndent(budget, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to format JSON: %w", err)
				}
				fmt.Println(string(data))
				return nil
			}

			fmt.Println("AI Budget Configuration")
			fmt.Println()
			if !budget.Enabled {
				fmt.Println("  Status:              Disabled")
			} else {
				fmt.Println("  Status:              Enabled")
			}
			fmt.Printf("  Monthly Limit:       $%.2f\n", budget.MonthlyLimit)
			fmt.Printf("  Alert Threshold:     %.0f%%\n", budget.AlertAt)

			return nil
		},
	}
}

// saveBudgetConfig saves budget configuration to disk
func saveBudgetConfig(budget BudgetConfig) error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get config directory: %w", err)
	}

	budgetDir := filepath.Join(configDir, "nylas", "ai-data")
	if err := os.MkdirAll(budgetDir, 0750); err != nil {
		return fmt.Errorf("failed to create budget directory: %w", err)
	}

	budgetFile := filepath.Join(budgetDir, "budget.json")

	data, err := json.MarshalIndent(budget, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to format budget config: %w", err)
	}

	if err := os.WriteFile(budgetFile, data, 0600); err != nil {
		return fmt.Errorf("failed to save budget config: %w", err)
	}

	return nil
}

// loadBudgetConfig loads budget configuration from disk
func loadBudgetConfig() (*BudgetConfig, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get config directory: %w", err)
	}

	budgetFile := filepath.Join(configDir, "nylas", "ai-data", "budget.json")

	// #nosec G304 -- budgetFile constructed from UserConfigDir + "nylas/ai-data/budget.json"
	data, err := os.ReadFile(budgetFile)
	if err != nil {
		return nil, err
	}

	var budget BudgetConfig
	if err := json.Unmarshal(data, &budget); err != nil {
		return nil, fmt.Errorf("failed to parse budget config: %w", err)
	}

	return &budget, nil
}

// disableBudget disables budget enforcement
func disableBudget() error {
	budget, err := loadBudgetConfig()
	if err != nil {
		// No budget exists, nothing to disable
		return nil
	}

	budget.Enabled = false
	return saveBudgetConfig(*budget)
}
