package admin

import (
	"encoding/json"
	"fmt"

	"github.com/fatih/color"
	"github.com/mqasimca/nylas/internal/cli/common"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/spf13/cobra"
)

func newApplicationsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "applications",
		Aliases: []string{"app", "apps"},
		Short:   "Manage Nylas applications",
		Long:    "Manage Nylas applications in your organization.",
	}

	cmd.AddCommand(newAppListCmd())
	cmd.AddCommand(newAppShowCmd())
	cmd.AddCommand(newAppCreateCmd())
	cmd.AddCommand(newAppUpdateCmd())
	cmd.AddCommand(newAppDeleteCmd())

	return cmd
}

func newAppListCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List applications",
		Long:    "List all applications in your organization.",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}

			ctx, cancel := createContext()
			defer cancel()

			apps, err := client.ListApplications(ctx)
			if err != nil {
				return fmt.Errorf("failed to list applications: %w", err)
			}

			if jsonOutput {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(apps)
			}

			if len(apps) == 0 {
				fmt.Println("No applications found.")
				return nil
			}

			cyan := color.New(color.FgCyan)
			green := color.New(color.FgGreen)

			fmt.Printf("Found %d application(s):\n\n", len(apps))

			table := common.NewTable("APP ID", "REGION", "ENVIRONMENT")
			for _, app := range apps {
				region := app.Region
				if region == "" {
					region = "-"
				}
				env := app.Environment
				if env == "" {
					env = "-"
				}
				table.AddRow(cyan.Sprint(app.ApplicationID), green.Sprint(region), env)
			}
			table.Render()

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}

func newAppShowCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "show <app-id>",
		Short: "Show application details",
		Long:  "Show detailed information about a specific application.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}

			ctx, cancel := createContext()
			defer cancel()

			app, err := client.GetApplication(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to get application: %w", err)
			}

			if jsonOutput {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(app)
			}

			cyan := color.New(color.FgCyan)
			green := color.New(color.FgGreen)
			bold := color.New(color.Bold)

			bold.Println("Application Details")
			fmt.Printf("  ID: %s\n", cyan.Sprint(app.ID))
			fmt.Printf("  Application ID: %s\n", app.ApplicationID)
			fmt.Printf("  Organization ID: %s\n", app.OrganizationID)
			fmt.Printf("  Region: %s\n", green.Sprint(app.Region))
			fmt.Printf("  Environment: %s\n", app.Environment)

			if app.BrandingSettings != nil {
				fmt.Printf("\nBranding:\n")
				if app.BrandingSettings.Name != "" {
					fmt.Printf("  Name: %s\n", app.BrandingSettings.Name)
				}
				if app.BrandingSettings.WebsiteURL != "" {
					fmt.Printf("  Website: %s\n", cyan.Sprint(app.BrandingSettings.WebsiteURL))
				}
			}

			if len(app.CallbackURIs) > 0 {
				fmt.Printf("\nCallback URIs (%d):\n", len(app.CallbackURIs))
				for i, uri := range app.CallbackURIs {
					fmt.Printf("  %d. %s\n", i+1, uri)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}

func newAppCreateCmd() *cobra.Command {
	var (
		name         string
		region       string
		brandingName string
		websiteURL   string
		callbackURIs []string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an application",
		Long:  "Create a new Nylas application.",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}

			req := &domain.CreateApplicationRequest{
				Name:   name,
				Region: region,
			}

			if brandingName != "" || websiteURL != "" {
				req.BrandingSettings = &domain.BrandingSettings{
					Name:       brandingName,
					WebsiteURL: websiteURL,
				}
			}

			if len(callbackURIs) > 0 {
				req.CallbackURIs = callbackURIs
			}

			ctx, cancel := createContext()
			defer cancel()

			app, err := client.CreateApplication(ctx, req)
			if err != nil {
				return fmt.Errorf("failed to create application: %w", err)
			}

			green := color.New(color.FgGreen)
			cyan := color.New(color.FgCyan)

			green.Printf("✓ Created application\n")
			fmt.Printf("  ID: %s\n", cyan.Sprint(app.ID))
			fmt.Printf("  Application ID: %s\n", app.ApplicationID)
			fmt.Printf("  Region: %s\n", app.Region)

			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Application name (required)")
	cmd.Flags().StringVar(&region, "region", "us", "Region (us, eu)")
	cmd.Flags().StringVar(&brandingName, "branding-name", "", "Branding name")
	cmd.Flags().StringVar(&websiteURL, "website-url", "", "Website URL")
	cmd.Flags().StringSliceVar(&callbackURIs, "callback-uris", []string{}, "Callback URIs (comma-separated)")

	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func newAppUpdateCmd() *cobra.Command {
	var (
		name         string
		brandingName string
		websiteURL   string
	)

	cmd := &cobra.Command{
		Use:   "update <app-id>",
		Short: "Update an application",
		Long:  "Update an existing application.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}

			req := &domain.UpdateApplicationRequest{}

			if name != "" {
				req.Name = &name
			}

			if brandingName != "" || websiteURL != "" {
				req.BrandingSettings = &domain.BrandingSettings{
					Name:       brandingName,
					WebsiteURL: websiteURL,
				}
			}

			ctx, cancel := createContext()
			defer cancel()

			app, err := client.UpdateApplication(ctx, args[0], req)
			if err != nil {
				return fmt.Errorf("failed to update application: %w", err)
			}

			green := color.New(color.FgGreen)
			green.Printf("✓ Updated application: %s\n", app.ID)

			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Application name")
	cmd.Flags().StringVar(&brandingName, "branding-name", "", "Branding name")
	cmd.Flags().StringVar(&websiteURL, "website-url", "", "Website URL")

	return cmd
}

func newAppDeleteCmd() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete <app-id>",
		Short: "Delete an application",
		Long:  "Delete an application permanently.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !yes {
				fmt.Printf("Are you sure you want to delete application %s? (y/N): ", args[0])
				var confirm string
				_, _ = fmt.Scanln(&confirm)
				if confirm != "y" && confirm != "Y" {
					fmt.Println("Cancelled.")
					return nil
				}
			}

			client, err := getClient()
			if err != nil {
				return err
			}

			ctx, cancel := createContext()
			defer cancel()

			if err := client.DeleteApplication(ctx, args[0]); err != nil {
				return fmt.Errorf("failed to delete application: %w", err)
			}

			green := color.New(color.FgGreen)
			green.Printf("✓ Deleted application: %s\n", args[0])

			return nil
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}
