package config

import (
	"fmt"

	"github.com/mqasimca/nylas/internal/cli/common"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize configuration with defaults",
		Long: `Create a new configuration file with default values.

This will NOT overwrite an existing configuration file unless --force is used.`,
		Example: `  # Initialize config with defaults
  nylas config init

  # Force overwrite existing config
  nylas config init --force`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if configStore.Exists() && !force {
				return fmt.Errorf("configuration file already exists at %s\nUse --force to overwrite", configStore.Path())
			}

			cfg := domain.DefaultConfig()
			if err := configStore.Save(cfg); err != nil {
				return common.WrapSaveError("configuration", err)
			}

			fmt.Printf("%s Configuration initialized with defaults\n", common.Green.Sprint("✓"))
			fmt.Printf("Config file: %s\n", configStore.Path())
			fmt.Println("\nEdit with: nylas config list")
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing configuration")

	return cmd
}
