package ai

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage AI configuration",
		Long: `View and update AI configuration in ~/.config/nylas/config.yaml

The AI configuration includes:
  - default_provider: Which AI provider to use (ollama, claude, openai, groq)
  - ollama: Ollama-specific settings (host, model)
  - claude: Claude/Anthropic settings (model)
  - openai: OpenAI settings (model)
  - groq: Groq settings (model)
  - fallback: Fallback provider configuration

Examples:
  # Show full AI configuration
  nylas ai config show

  # List all AI settings
  nylas ai config list

  # Get a specific value
  nylas ai config get default_provider
  nylas ai config get ollama.model

  # Set a value
  nylas ai config set default_provider ollama
  nylas ai config set ollama.host http://localhost:11434
  nylas ai config set ollama.model llama3.1:8b
  nylas ai config set claude.model claude-3-5-sonnet-20241022`,
	}

	cmd.AddCommand(newConfigShowCmd())
	cmd.AddCommand(newConfigListCmd())
	cmd.AddCommand(newConfigGetCmd())
	cmd.AddCommand(newConfigSetCmd())

	return cmd
}

func newConfigShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show full AI configuration",
		Long:  "Display the complete AI configuration in YAML format",
		RunE: func(cmd *cobra.Command, args []string) error {
			store := getConfigStore(cmd)
			cfg, err := store.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			if cfg.AI == nil {
				fmt.Println("No AI configuration found")
				fmt.Println("\nTo configure AI, use:")
				fmt.Println("  nylas ai config set default_provider ollama")
				return nil
			}

			data, err := yaml.Marshal(cfg.AI)
			if err != nil {
				return fmt.Errorf("failed to format config: %w", err)
			}

			fmt.Printf("AI Configuration:\n\n")
			fmt.Print(string(data))
			return nil
		},
	}
}

func newConfigListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all AI configuration keys and values",
		Long:  "Display all AI configuration settings as key-value pairs",
		RunE: func(cmd *cobra.Command, args []string) error {
			store := getConfigStore(cmd)
			cfg, err := store.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			if cfg.AI == nil {
				fmt.Println("No AI configuration found")
				return nil
			}

			fmt.Println("AI Configuration:")
			fmt.Println()

			// Default provider
			fmt.Printf("  default_provider: %s\n", cfg.AI.DefaultProvider)

			// Fallback
			if cfg.AI.Fallback != nil {
				fmt.Printf("\n  Fallback:\n")
				fmt.Printf("    enabled: %v\n", cfg.AI.Fallback.Enabled)
				if len(cfg.AI.Fallback.Providers) > 0 {
					fmt.Printf("    providers: [%s]\n", strings.Join(cfg.AI.Fallback.Providers, ", "))
				}
			}

			// Ollama
			if cfg.AI.Ollama != nil {
				fmt.Printf("\n  Ollama:\n")
				fmt.Printf("    host: %s\n", cfg.AI.Ollama.Host)
				fmt.Printf("    model: %s\n", cfg.AI.Ollama.Model)
			}

			// Claude
			if cfg.AI.Claude != nil {
				fmt.Printf("\n  Claude:\n")
				if cfg.AI.Claude.APIKey != "" {
					fmt.Printf("    api_key: %s\n", maskAPIKey(cfg.AI.Claude.APIKey))
				}
				fmt.Printf("    model: %s\n", cfg.AI.Claude.Model)
			}

			// OpenAI
			if cfg.AI.OpenAI != nil {
				fmt.Printf("\n  OpenAI:\n")
				if cfg.AI.OpenAI.APIKey != "" {
					fmt.Printf("    api_key: %s\n", maskAPIKey(cfg.AI.OpenAI.APIKey))
				}
				fmt.Printf("    model: %s\n", cfg.AI.OpenAI.Model)
			}

			// Groq
			if cfg.AI.Groq != nil {
				fmt.Printf("\n  Groq:\n")
				if cfg.AI.Groq.APIKey != "" {
					fmt.Printf("    api_key: %s\n", maskAPIKey(cfg.AI.Groq.APIKey))
				}
				fmt.Printf("    model: %s\n", cfg.AI.Groq.Model)
			}

			// OpenRouter
			if cfg.AI.OpenRouter != nil {
				fmt.Printf("\n  OpenRouter:\n")
				if cfg.AI.OpenRouter.APIKey != "" {
					fmt.Printf("    api_key: %s\n", maskAPIKey(cfg.AI.OpenRouter.APIKey))
				}
				fmt.Printf("    model: %s\n", cfg.AI.OpenRouter.Model)
			}

			// Privacy
			if cfg.AI.Privacy != nil {
				fmt.Printf("\n  Privacy:\n")
				fmt.Printf("    allow_cloud_ai: %v\n", cfg.AI.Privacy.AllowCloudAI)
				fmt.Printf("    data_retention: %d\n", cfg.AI.Privacy.DataRetention)
				fmt.Printf("    local_storage_only: %v\n", cfg.AI.Privacy.LocalStorageOnly)
			}

			// Features
			if cfg.AI.Features != nil {
				fmt.Printf("\n  Features:\n")
				fmt.Printf("    natural_language_scheduling: %v\n", cfg.AI.Features.NaturalLanguageScheduling)
				fmt.Printf("    predictive_scheduling: %v\n", cfg.AI.Features.PredictiveScheduling)
				fmt.Printf("    focus_time_protection: %v\n", cfg.AI.Features.FocusTimeProtection)
				fmt.Printf("    conflict_resolution: %v\n", cfg.AI.Features.ConflictResolution)
				fmt.Printf("    email_context_analysis: %v\n", cfg.AI.Features.EmailContextAnalysis)
			}

			return nil
		},
	}
}

func newConfigGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Get an AI configuration value",
		Long: `Get a specific AI configuration value.

Supported keys:
  - default_provider
  - fallback.enabled
  - fallback.providers
  - ollama.host
  - ollama.model
  - claude.api_key
  - claude.model
  - openai.api_key
  - openai.model
  - groq.api_key
  - groq.model
  - openrouter.api_key
  - openrouter.model
  - privacy.allow_cloud_ai
  - privacy.data_retention
  - privacy.local_storage_only
  - features.natural_language_scheduling
  - features.predictive_scheduling
  - features.focus_time_protection
  - features.conflict_resolution
  - features.email_context_analysis

Examples:
  nylas ai config get default_provider
  nylas ai config get ollama.model
  nylas ai config get privacy.allow_cloud_ai
  nylas ai config get features.natural_language_scheduling`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			store := getConfigStore(cmd)
			cfg, err := store.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			if cfg.AI == nil {
				return fmt.Errorf("no AI configuration found")
			}

			value, err := getConfigValue(cfg.AI, key)
			if err != nil {
				return err
			}

			fmt.Println(value)
			return nil
		},
	}
}
