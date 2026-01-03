package ai

import (
	"fmt"
	"strings"

	"github.com/mqasimca/nylas/internal/cli/common"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/spf13/cobra"
)

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
  - privacy.allow_cloud_ai (true, false)
  - privacy.data_retention (number of days, 0 to disable)
  - privacy.local_storage_only (true, false)
  - features.natural_language_scheduling (true, false)
  - features.predictive_scheduling (true, false)
  - features.focus_time_protection (true, false)
  - features.conflict_resolution (true, false)
  - features.email_context_analysis (true, false)

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
  nylas ai config set fallback.providers ollama,claude,openai

  # Configure privacy settings
  nylas ai config set privacy.allow_cloud_ai false
  nylas ai config set privacy.data_retention 90
  nylas ai config set privacy.local_storage_only true

  # Configure feature toggles
  nylas ai config set features.natural_language_scheduling true
  nylas ai config set features.focus_time_protection true`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			value := args[1]

			store := common.GetConfigStore(cmd)
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

// maskAPIKey masks an API key for display, showing first 8 and last 4 characters.
// Example: "sk-proj-abcdefghijklmnop" -> "sk-proj-***...***mnop"
func maskAPIKey(key string) string {
	if len(key) <= 12 {
		// Too short to mask meaningfully
		return "***"
	}
	return key[:8] + "***...***" + key[len(key)-4:]
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

	case "privacy":
		if ai.Privacy == nil {
			return "", fmt.Errorf("privacy not configured")
		}
		if len(parts) < 2 {
			return "", fmt.Errorf("invalid key: %s", key)
		}
		switch parts[1] {
		case "allow_cloud_ai":
			return fmt.Sprintf("%v", ai.Privacy.AllowCloudAI), nil
		case "data_retention":
			return fmt.Sprintf("%d", ai.Privacy.DataRetention), nil
		case "local_storage_only":
			return fmt.Sprintf("%v", ai.Privacy.LocalStorageOnly), nil
		default:
			return "", fmt.Errorf("unknown privacy key: %s", parts[1])
		}

	case "features":
		if ai.Features == nil {
			return "", fmt.Errorf("features not configured")
		}
		if len(parts) < 2 {
			return "", fmt.Errorf("invalid key: %s", key)
		}
		switch parts[1] {
		case "natural_language_scheduling":
			return fmt.Sprintf("%v", ai.Features.NaturalLanguageScheduling), nil
		case "predictive_scheduling":
			return fmt.Sprintf("%v", ai.Features.PredictiveScheduling), nil
		case "focus_time_protection":
			return fmt.Sprintf("%v", ai.Features.FocusTimeProtection), nil
		case "conflict_resolution":
			return fmt.Sprintf("%v", ai.Features.ConflictResolution), nil
		case "email_context_analysis":
			return fmt.Sprintf("%v", ai.Features.EmailContextAnalysis), nil
		default:
			return "", fmt.Errorf("unknown features key: %s", parts[1])
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

	case "privacy":
		if ai.Privacy == nil {
			ai.Privacy = &domain.PrivacyConfig{}
		}
		if len(parts) < 2 {
			return fmt.Errorf("invalid key: %s", key)
		}
		switch parts[1] {
		case "allow_cloud_ai":
			ai.Privacy.AllowCloudAI = value == "true"
		case "data_retention":
			var retention int
			_, err := fmt.Sscanf(value, "%d", &retention)
			if err != nil {
				return fmt.Errorf("invalid data_retention value: %s (must be integer)", value)
			}
			ai.Privacy.DataRetention = retention
		case "local_storage_only":
			ai.Privacy.LocalStorageOnly = value == "true"
		default:
			return fmt.Errorf("unknown privacy key: %s", parts[1])
		}

	case "features":
		if ai.Features == nil {
			ai.Features = &domain.FeaturesConfig{}
		}
		if len(parts) < 2 {
			return fmt.Errorf("invalid key: %s", key)
		}
		switch parts[1] {
		case "natural_language_scheduling":
			ai.Features.NaturalLanguageScheduling = value == "true"
		case "predictive_scheduling":
			ai.Features.PredictiveScheduling = value == "true"
		case "focus_time_protection":
			ai.Features.FocusTimeProtection = value == "true"
		case "conflict_resolution":
			ai.Features.ConflictResolution = value == "true"
		case "email_context_analysis":
			ai.Features.EmailContextAnalysis = value == "true"
		default:
			return fmt.Errorf("unknown features key: %s", parts[1])
		}

	default:
		return fmt.Errorf("unknown config key: %s", key)
	}

	return nil
}
