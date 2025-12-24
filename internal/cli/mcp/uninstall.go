package mcp

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func newUninstallCmd() *cobra.Command {
	var (
		assistantID  string
		uninstallAll bool
	)

	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Remove MCP configuration from AI assistants",
		Long: `Remove the Nylas MCP server configuration from AI assistants.

This command removes the nylas entry from the mcpServers section of
the AI assistant's configuration file.`,
		Example: `  # Uninstall from specific assistant
  nylas mcp uninstall --assistant claude-desktop

  # Uninstall from all configured assistants
  nylas mcp uninstall --all`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUninstall(assistantID, uninstallAll)
		},
	}

	cmd.Flags().StringVarP(&assistantID, "assistant", "a", "", "Target assistant (claude-desktop, claude-code, cursor, windsurf, vscode)")
	cmd.Flags().BoolVar(&uninstallAll, "all", false, "Uninstall from all configured assistants")

	return cmd
}

func runUninstall(assistantID string, uninstallAll bool) error {
	var assistants []Assistant

	if uninstallAll {
		assistants = Assistants
	} else if assistantID != "" {
		a := GetAssistantByID(assistantID)
		if a == nil {
			return fmt.Errorf("unknown assistant: %s\n\nSupported: claude-desktop, claude-code, cursor, windsurf, vscode", assistantID)
		}
		assistants = []Assistant{*a}
	} else {
		return fmt.Errorf("please specify --assistant or --all")
	}

	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)
	successCount := 0

	for _, a := range assistants {
		configPath := a.GetConfigPath()
		if configPath == "" {
			continue
		}

		// Check if config exists
		if !a.IsConfigured() {
			if !uninstallAll {
				yellow.Printf("  ! %s: not configured\n", a.Name)
			}
			continue
		}

		// Check if nylas is in the config
		hasNylas, _ := checkNylasInConfig(configPath)
		if !hasNylas {
			if !uninstallAll {
				yellow.Printf("  ! %s: nylas not found in config\n", a.Name)
			}
			continue
		}

		err := uninstallFromAssistant(a)
		if err != nil {
			yellow.Printf("  ! %s: %v\n", a.Name, err)
			continue
		}

		green.Printf("  ✓ %s: removed from %s\n", a.Name, configPath)
		successCount++
	}

	if successCount > 0 {
		fmt.Println()
		fmt.Println("Restart your AI assistant to apply the changes.")
	}

	return nil
}

func uninstallFromAssistant(a Assistant) error {
	configPath := a.GetConfigPath()

	// Read existing config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("reading config: %w", err)
	}

	var config map[string]any
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("parsing config: %w", err)
	}

	// Get mcpServers section
	mcpServers, ok := config["mcpServers"].(map[string]any)
	if !ok {
		return nil // No mcpServers section
	}

	// Remove nylas entry
	delete(mcpServers, "nylas")

	// If mcpServers is now empty, remove it entirely
	if len(mcpServers) == 0 {
		delete(config, "mcpServers")
	} else {
		config["mcpServers"] = mcpServers
	}

	// Write config back
	data, err = json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	return nil
}
