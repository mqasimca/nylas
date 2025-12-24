package mcp

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show MCP installation status",
		Long: `Show the MCP configuration status for all supported AI assistants.

This command checks which AI assistants have Nylas MCP configured and
displays the configuration path for each.`,
		Example: `  nylas mcp status`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus()
		},
	}

	return cmd
}

func runStatus() error {
	fmt.Println("MCP Installation Status:")
	fmt.Println()

	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)
	gray := color.New(color.FgHiBlack)

	for _, a := range Assistants {
		configPath := a.GetConfigPath()
		if configPath == "" {
			gray.Printf("  - %-16s  unsupported on this platform\n", a.Name)
			continue
		}

		// Check if app is installed
		if !a.IsProjectConfig() && !a.IsInstalled() {
			gray.Printf("  - %-16s  application not installed\n", a.Name)
			continue
		}

		// Check if config file exists
		if !a.IsConfigured() {
			yellow.Printf("  ○ %-16s  ", a.Name)
			fmt.Printf("not configured  %s\n", configPath)
			continue
		}

		// Check if nylas is in the config
		hasNylas, binaryPath := checkNylasInConfig(configPath)
		if !hasNylas {
			yellow.Printf("  ○ %-16s  ", a.Name)
			fmt.Printf("config exists, nylas not added  %s\n", configPath)
			continue
		}

		green.Printf("  ✓ %-16s  ", a.Name)
		fmt.Printf("configured  %s\n", configPath)
		if binaryPath != "" {
			gray.Printf("                       binary: %s\n", binaryPath)
		}
	}

	fmt.Println()
	fmt.Println("Legend:")
	green.Print("  ✓")
	fmt.Println(" Nylas MCP configured")
	yellow.Print("  ○")
	fmt.Println(" Available but not configured")
	gray.Print("  -")
	fmt.Println(" Not available")

	return nil
}

// checkNylasInConfig checks if nylas is configured in the MCP config file.
func checkNylasInConfig(configPath string) (bool, string) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return false, ""
	}

	var config map[string]any
	if err := json.Unmarshal(data, &config); err != nil {
		return false, ""
	}

	mcpServers, ok := config["mcpServers"].(map[string]any)
	if !ok {
		return false, ""
	}

	nylas, ok := mcpServers["nylas"].(map[string]any)
	if !ok {
		return false, ""
	}

	binaryPath, _ := nylas["command"].(string)
	return true, binaryPath
}
