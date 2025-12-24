package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// newListCmd creates the list subcommand.
func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all widgets",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get service instance (in real app, from dependency injection)
			service := getWidgetService()

			// Create context
			ctx := context.Background()

			// Call service method (business logic in adapter/domain)
			widgets, err := service.ListWidgets(ctx)
			if err != nil {
				return fmt.Errorf("failed to list widgets: %w", err)
			}

			// Format output (presentation logic in CLI)
			if len(widgets) == 0 {
				fmt.Println("No widgets found")
				return nil
			}

			fmt.Printf("Found %d widgets:\n\n", len(widgets))
			for _, w := range widgets {
				fmt.Printf("  %s - %s\n", w.ID, w.Name)
				if w.Description != "" {
					fmt.Printf("    %s\n", w.Description)
				}
				fmt.Println()
			}

			return nil
		},
	}

	return cmd
}
