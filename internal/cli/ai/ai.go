package ai

import (
	"github.com/spf13/cobra"
)

// NewAICmd creates the AI command.
func NewAICmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ai",
		Short: "AI configuration and management",
		Long: `Manage AI/LLM settings and configuration for the Nylas CLI.

Configure AI providers (Ollama, Claude, OpenAI, Groq), manage settings,
and control AI features for calendar intelligence and scheduling.

Examples:
  # Show current AI configuration
  nylas ai config show

  # Set default AI provider
  nylas ai config set default_provider ollama

  # Configure Ollama settings
  nylas ai config set ollama.host http://localhost:11434
  nylas ai config set ollama.model llama3.1:8b

  # Configure Claude (Anthropic) settings
  nylas ai config set claude.model claude-3-5-sonnet-20241022

  # Get a specific config value
  nylas ai config get default_provider

  # List all AI configuration
  nylas ai config list`,
	}

	// Add subcommands
	cmd.AddCommand(newConfigCmd())
	cmd.AddCommand(newClearDataCmd())
	cmd.AddCommand(newUsageCmd())
	cmd.AddCommand(newSetBudgetCmd())
	cmd.AddCommand(newShowBudgetCmd())

	return cmd
}
