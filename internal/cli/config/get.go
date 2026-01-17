package config

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/mqasimca/nylas/internal/cli/common"
	"github.com/spf13/cobra"
)

func newGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Get a configuration value",
		Long: `Get a specific configuration value using dot notation.

Examples of keys:
  region
  default_grant
  api.timeout
  api.base_url
  output.format
  output.color`,
		Example: `  # Get API timeout
  nylas config get api.timeout

  # Get default grant ID
  nylas config get default_grant

  # Get output format
  nylas config get output.format`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := configStore.Load()
			if err != nil {
				return common.WrapLoadError("configuration", err)
			}

			key := args[0]
			value, err := getConfigValue(cfg, key)
			if err != nil {
				return err
			}

			fmt.Println(value)
			return nil
		},
	}
}

func getConfigValue(cfg any, key string) (string, error) {
	parts := strings.Split(key, ".")

	v := reflect.ValueOf(cfg)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	for _, part := range parts {
		if v.Kind() != reflect.Struct {
			return "", fmt.Errorf("cannot access field %s", part)
		}

		// Convert snake_case to PascalCase for struct fields
		fieldName := snakeToPascal(part)
		field := v.FieldByName(fieldName)

		if !field.IsValid() {
			return "", fmt.Errorf("unknown config key: %s", key)
		}

		// If field is a pointer, dereference it
		if field.Kind() == reflect.Ptr {
			if field.IsNil() {
				return "", nil
			}
			field = field.Elem()
		}

		v = field
	}

	return fmt.Sprintf("%v", v.Interface()), nil
}

func snakeToPascal(s string) string {
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, "")
}
