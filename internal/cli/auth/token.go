package auth

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/atotto/clipboard"
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/mqasimca/nylas/internal/ports"
)

func newTokenCmd() *cobra.Command {
	var copyToClipboard bool

	cmd := &cobra.Command{
		Use:   "token",
		Short: "Show or copy API key",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, secretStore, _, err := createDependencies()
			if err != nil {
				return err
			}

			apiKey, err := secretStore.Get(ports.KeyAPIKey)
			if err != nil {
				return fmt.Errorf("API key not found - run 'nylas auth config' first")
			}

			jsonOutput, _ := cmd.Root().PersistentFlags().GetBool("json")
			if jsonOutput {
				output := map[string]string{"api_key": apiKey}
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(output)
			}

			if copyToClipboard {
				if err := clipboard.WriteAll(apiKey); err != nil {
					return fmt.Errorf("failed to copy to clipboard: %w", err)
				}
				green := color.New(color.FgGreen)
				_, _ = green.Println("✓ API key copied to clipboard")
			} else {
				fmt.Println(apiKey)
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&copyToClipboard, "copy", "c", false, "Copy to clipboard")

	return cmd
}
