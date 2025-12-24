package cli

import (
	"context"
	"fmt"

	"github.com/mqasimca/nylas/examples/minimal-feature/domain"
	"github.com/spf13/cobra"
)

// newCreateCmd creates the create subcommand.
func newCreateCmd() *cobra.Command {
	var name, description string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new widget",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate flags (presentation validation)
			if name == "" {
				return fmt.Errorf("--name is required")
			}

			// Create domain object
			widget := &domain.Widget{
				Name:        name,
				Description: description,
			}

			// Validate business rules (domain validation)
			if err := widget.Validate(); err != nil {
				return err
			}

			// Get service
			service := getWidgetService()
			ctx := context.Background()

			// Create via service
			created, err := service.CreateWidget(ctx, widget)
			if err != nil {
				return fmt.Errorf("failed to create widget: %w", err)
			}

			// Display success
			fmt.Printf("Widget created successfully!\n")
			fmt.Printf("  ID: %s\n", created.ID)
			fmt.Printf("  Name: %s\n", created.Name)

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringVar(&name, "name", "", "Widget name (required)")
	cmd.Flags().StringVar(&description, "description", "", "Widget description")

	return cmd
}
