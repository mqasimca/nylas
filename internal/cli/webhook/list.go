package webhook

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/mqasimca/nylas/internal/cli/common"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func newListCmd() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all webhooks",
		Long: `List all webhooks configured for your Nylas application.

Shows webhook ID, description, URL, status, and trigger types.`,
		Example: `  # List all webhooks
  nylas webhook list

  # List in JSON format
  nylas webhook list --format json

  # List in YAML format
  nylas webhook list --format yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := getClient()
			if err != nil {
				return common.NewUserError("Failed to initialize client: "+err.Error(),
					"Run 'nylas auth login' to authenticate")
			}

			ctx, cancel := createContext()
			defer cancel()

			spinner := common.NewSpinner("Fetching webhooks...")
			spinner.Start()

			webhooks, err := c.ListWebhooks(ctx)
			spinner.Stop()

			if err != nil {
				return common.NewUserError("Failed to list webhooks: "+err.Error(),
					"Check your API key has webhook management permissions")
			}

			if len(webhooks) == 0 {
				fmt.Println("No webhooks configured.")
				fmt.Println("\nCreate a webhook with: nylas webhook create --url <URL> --triggers <triggers>")
				return nil
			}

			switch format {
			case "json":
				return outputJSON(webhooks)
			case "yaml":
				return outputYAML(webhooks)
			case "csv":
				return outputCSV(webhooks)
			default:
				return outputTable(webhooks)
			}
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "table", "Output format (table, json, yaml, csv)")

	return cmd
}

func outputJSON(webhooks interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(webhooks)
}

func outputYAML(webhooks interface{}) error {
	return yaml.NewEncoder(os.Stdout).Encode(webhooks)
}

func outputCSV(webhooks interface{}) error {
	w := csv.NewWriter(os.Stdout)
	defer w.Flush()

	// Write header
	w.Write([]string{"ID", "Description", "URL", "Status", "Triggers"})

	// Get webhooks as slice
	data, _ := json.Marshal(webhooks)
	var items []map[string]interface{}
	json.Unmarshal(data, &items)

	for _, item := range items {
		id, _ := item["id"].(string)
		desc, _ := item["description"].(string)
		url, _ := item["webhook_url"].(string)
		status, _ := item["status"].(string)

		var triggers []string
		if triggerList, ok := item["trigger_types"].([]interface{}); ok {
			for _, t := range triggerList {
				triggers = append(triggers, fmt.Sprintf("%v", t))
			}
		}

		w.Write([]string{id, desc, url, status, strings.Join(triggers, ";")})
	}

	return nil
}

func outputTable(webhooks interface{}) error {
	data, _ := json.Marshal(webhooks)
	var items []map[string]interface{}
	json.Unmarshal(data, &items)

	// Calculate column widths
	headers := []string{"ID", "Description", "URL", "Status", "Triggers"}
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}

	type row struct {
		id, desc, url, status, triggers string
	}
	var rows []row

	for _, item := range items {
		r := row{
			id:     truncate(fmt.Sprintf("%v", item["id"]), 20),
			desc:   truncate(fmt.Sprintf("%v", item["description"]), 25),
			url:    truncate(fmt.Sprintf("%v", item["webhook_url"]), 35),
			status: fmt.Sprintf("%v", item["status"]),
		}

		var triggers []string
		if triggerList, ok := item["trigger_types"].([]interface{}); ok {
			for _, t := range triggerList {
				triggers = append(triggers, fmt.Sprintf("%v", t))
			}
		}
		r.triggers = truncate(strings.Join(triggers, ", "), 30)

		rows = append(rows, r)

		if len(r.id) > widths[0] {
			widths[0] = len(r.id)
		}
		if len(r.desc) > widths[1] {
			widths[1] = len(r.desc)
		}
		if len(r.url) > widths[2] {
			widths[2] = len(r.url)
		}
		if len(r.status) > widths[3] {
			widths[3] = len(r.status)
		}
		if len(r.triggers) > widths[4] {
			widths[4] = len(r.triggers)
		}
	}

	// Print header
	fmt.Printf("%-*s  %-*s  %-*s  %-*s  %s\n",
		widths[0], headers[0],
		widths[1], headers[1],
		widths[2], headers[2],
		widths[3], headers[3],
		headers[4])

	// Print separator
	for i, w := range widths {
		if i > 0 {
			fmt.Print("  ")
		}
		fmt.Print(strings.Repeat("-", w))
	}
	fmt.Println()

	// Print rows
	for _, r := range rows {
		statusIcon := getStatusIcon(r.status)
		fmt.Printf("%-*s  %-*s  %-*s  %s %-*s  %s\n",
			widths[0], r.id,
			widths[1], r.desc,
			widths[2], r.url,
			statusIcon, widths[3]-2, r.status,
			r.triggers)
	}

	fmt.Printf("\nTotal: %d webhooks\n", len(rows))
	return nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func getStatusIcon(status string) string {
	switch status {
	case "active":
		return "\033[32m●\033[0m" // Green
	case "inactive":
		return "\033[33m●\033[0m" // Yellow
	case "failing":
		return "\033[31m●\033[0m" // Red
	default:
		return "○"
	}
}
