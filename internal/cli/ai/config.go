package ai

import (
	"fmt"
	"strings"

	"github.com/mqasimca/nylas/internal/adapters/config"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/ports"
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
					fmt.Printf("    api_key: %s\n", cfg.AI.Claude.APIKey)
				}
				fmt.Printf("    model: %s\n", cfg.AI.Claude.Model)
			}

			// OpenAI
			if cfg.AI.OpenAI != nil {
				fmt.Printf("\n  OpenAI:\n")
				if cfg.AI.OpenAI.APIKey != "" {
					fmt.Printf("    api_key: %s\n", cfg.AI.OpenAI.APIKey)
				}
				fmt.Printf("    model: %s\n", cfg.AI.OpenAI.Model)
			}

			// Groq
			if cfg.AI.Groq != nil {
				fmt.Printf("\n  Groq:\n")
				if cfg.AI.Groq.APIKey != "" {
					fmt.Printf("    api_key: %s\n", cfg.AI.Groq.APIKey)
				}
				fmt.Printf("    model: %s\n", cfg.AI.Groq.Model)
			}

			// OpenRouter
			if cfg.AI.OpenRouter != nil {
				fmt.Printf("\n  OpenRouter:\n")
				if cfg.AI.OpenRouter.APIKey != "" {
					fmt.Printf("    api_key: %s\n", cfg.AI.OpenRouter.APIKey)
				}
				fmt.Printf("    model: %s\n", cfg.AI.OpenRouter.Model)
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

Examples:
  nylas ai config get default_provider
  nylas ai config get ollama.model
  nylas ai config get claude.model`,
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

func newConfigSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set an AI configuration value",
		Long: `Set a specific AI configuration value.

Supported keys:
  - default_provider (ollama, claude, openai, groq, openrouter)
  - fallback.enabled (true, false)
  - fallback.providers (comma-separated list: ollama,claude,openai)
  - ollama.host (e.g., http://localhost:11434)
  - ollama.model (e.g., llama3.1:8b, mistral:latest)
  - claude.api_key (e.g., ${ANTHROPIC_API_KEY} or actual key)
  - claude.model (e.g., claude-3-5-sonnet-20241022)
  - openai.api_key (e.g., ${OPENAI_API_KEY} or actual key)
  - openai.model (e.g., gpt-4-turbo, gpt-4o)
  - groq.api_key (e.g., ${GROQ_API_KEY} or actual key)
  - groq.model (e.g., mixtral-8x7b-32768)
  - openrouter.api_key (e.g., ${OPENROUTER_API_KEY} or actual key)
  - openrouter.model (e.g., anthropic/claude-3.5-sonnet)

Examples:
  # Set Ollama as default provider
  nylas ai config set default_provider ollama

  # Configure Ollama
  nylas ai config set ollama.host http://localhost:11434
  nylas ai config set ollama.model llama3.1:8b

  # Configure Claude
  nylas ai config set claude.model claude-3-5-sonnet-20241022
  nylas ai config set claude.api_key '${ANTHROPIC_API_KEY}'

  # Enable fallback with multiple providers
  nylas ai config set fallback.enabled true
  nylas ai config set fallback.providers ollama,claude,openai`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			value := args[1]

			store := getConfigStore(cmd)
			cfg, err := store.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Initialize AI config if it doesn't exist
			if cfg.AI == nil {
				cfg.AI = &domain.AIConfig{}
			}

			if err := setConfigValue(cfg.AI, key, value); err != nil {
				return err
			}

			if err := store.Save(cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Printf("✓ Set %s = %s\n", key, value)
			fmt.Printf("\nConfiguration saved to: %s\n", store.Path())
			return nil
		},
	}
}

// getConfigValue retrieves a configuration value by key path.
func getConfigValue(ai *domain.AIConfig, key string) (string, error) {
	parts := strings.Split(key, ".")

	switch parts[0] {
	case "default_provider":
		return ai.DefaultProvider, nil

	case "fallback":
		if ai.Fallback == nil {
			return "", fmt.Errorf("fallback not configured")
		}
		if len(parts) < 2 {
			return "", fmt.Errorf("invalid key: %s", key)
		}
		switch parts[1] {
		case "enabled":
			return fmt.Sprintf("%v", ai.Fallback.Enabled), nil
		case "providers":
			return strings.Join(ai.Fallback.Providers, ","), nil
		default:
			return "", fmt.Errorf("unknown fallback key: %s", parts[1])
		}

	case "ollama":
		if ai.Ollama == nil {
			return "", fmt.Errorf("ollama not configured")
		}
		if len(parts) < 2 {
			return "", fmt.Errorf("invalid key: %s", key)
		}
		switch parts[1] {
		case "host":
			return ai.Ollama.Host, nil
		case "model":
			return ai.Ollama.Model, nil
		default:
			return "", fmt.Errorf("unknown ollama key: %s", parts[1])
		}

	case "claude":
		if ai.Claude == nil {
			return "", fmt.Errorf("claude not configured")
		}
		if len(parts) < 2 {
			return "", fmt.Errorf("invalid key: %s", key)
		}
		switch parts[1] {
		case "api_key":
			return ai.Claude.APIKey, nil
		case "model":
			return ai.Claude.Model, nil
		default:
			return "", fmt.Errorf("unknown claude key: %s", parts[1])
		}

	case "openai":
		if ai.OpenAI == nil {
			return "", fmt.Errorf("openai not configured")
		}
		if len(parts) < 2 {
			return "", fmt.Errorf("invalid key: %s", key)
		}
		switch parts[1] {
		case "api_key":
			return ai.OpenAI.APIKey, nil
		case "model":
			return ai.OpenAI.Model, nil
		default:
			return "", fmt.Errorf("unknown openai key: %s", parts[1])
		}

	case "groq":
		if ai.Groq == nil {
			return "", fmt.Errorf("groq not configured")
		}
		if len(parts) < 2 {
			return "", fmt.Errorf("invalid key: %s", key)
		}
		switch parts[1] {
		case "api_key":
			return ai.Groq.APIKey, nil
		case "model":
			return ai.Groq.Model, nil
		default:
			return "", fmt.Errorf("unknown groq key: %s", parts[1])
		}

	case "openrouter":
		if ai.OpenRouter == nil {
			return "", fmt.Errorf("openrouter not configured")
		}
		if len(parts) < 2 {
			return "", fmt.Errorf("invalid key: %s", key)
		}
		switch parts[1] {
		case "api_key":
			return ai.OpenRouter.APIKey, nil
		case "model":
			return ai.OpenRouter.Model, nil
		default:
			return "", fmt.Errorf("unknown openrouter key: %s", parts[1])
		}

	default:
		return "", fmt.Errorf("unknown config key: %s", key)
	}
}

// setConfigValue sets a configuration value by key path.
func setConfigValue(ai *domain.AIConfig, key, value string) error {
	parts := strings.Split(key, ".")

	switch parts[0] {
	case "default_provider":
		// Validate provider
		validProviders := []string{"ollama", "claude", "openai", "groq", "openrouter"}
		valid := false
		for _, p := range validProviders {
			if value == p {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid provider: %s (must be one of: %s)", value, strings.Join(validProviders, ", "))
		}
		ai.DefaultProvider = value

	case "fallback":
		if ai.Fallback == nil {
			ai.Fallback = &domain.AIFallbackConfig{}
		}
		if len(parts) < 2 {
			return fmt.Errorf("invalid key: %s", key)
		}
		switch parts[1] {
		case "enabled":
			ai.Fallback.Enabled = value == "true"
		case "providers":
			ai.Fallback.Providers = strings.Split(value, ",")
		default:
			return fmt.Errorf("unknown fallback key: %s", parts[1])
		}

	case "ollama":
		if ai.Ollama == nil {
			ai.Ollama = &domain.OllamaConfig{}
		}
		if len(parts) < 2 {
			return fmt.Errorf("invalid key: %s", key)
		}
		switch parts[1] {
		case "host":
			ai.Ollama.Host = value
		case "model":
			ai.Ollama.Model = value
		default:
			return fmt.Errorf("unknown ollama key: %s", parts[1])
		}

	case "claude":
		if ai.Claude == nil {
			ai.Claude = &domain.ClaudeConfig{}
		}
		if len(parts) < 2 {
			return fmt.Errorf("invalid key: %s", key)
		}
		switch parts[1] {
		case "api_key":
			ai.Claude.APIKey = value
		case "model":
			ai.Claude.Model = value
		default:
			return fmt.Errorf("unknown claude key: %s", parts[1])
		}

	case "openai":
		if ai.OpenAI == nil {
			ai.OpenAI = &domain.OpenAIConfig{}
		}
		if len(parts) < 2 {
			return fmt.Errorf("invalid key: %s", key)
		}
		switch parts[1] {
		case "api_key":
			ai.OpenAI.APIKey = value
		case "model":
			ai.OpenAI.Model = value
		default:
			return fmt.Errorf("unknown openai key: %s", parts[1])
		}

	case "groq":
		if ai.Groq == nil {
			ai.Groq = &domain.GroqConfig{}
		}
		if len(parts) < 2 {
			return fmt.Errorf("invalid key: %s", key)
		}
		switch parts[1] {
		case "api_key":
			ai.Groq.APIKey = value
		case "model":
			ai.Groq.Model = value
		default:
			return fmt.Errorf("unknown groq key: %s", parts[1])
		}

	case "openrouter":
		if ai.OpenRouter == nil {
			ai.OpenRouter = &domain.OpenRouterConfig{}
		}
		if len(parts) < 2 {
			return fmt.Errorf("invalid key: %s", key)
		}
		switch parts[1] {
		case "api_key":
			ai.OpenRouter.APIKey = value
		case "model":
			ai.OpenRouter.Model = value
		default:
			return fmt.Errorf("unknown openrouter key: %s", parts[1])
		}

	default:
		return fmt.Errorf("unknown config key: %s", key)
	}

	return nil
}

// getConfigStore returns the appropriate config store based on the --config flag
func getConfigStore(cmd *cobra.Command) ports.ConfigStore {
	// Try to get custom config path from flag
	configPath, _ := cmd.Flags().GetString("config")
	if configPath == "" {
		// Try to get from parent (persistent flag)
		if cmd.Parent() != nil {
			configPath, _ = cmd.Parent().Flags().GetString("config")
		}
	}

	if configPath != "" {
		return config.NewFileStore(configPath)
	}
	return config.NewDefaultFileStore()
}
