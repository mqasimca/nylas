package email

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mqasimca/nylas/internal/cli/common"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/spf13/cobra"
)

func newSmartComposeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "smart-compose",
		Short: "Generate AI-powered email drafts",
		Long: `Generate AI-powered email drafts using Nylas Smart Compose.

Smart Compose uses AI to generate well-written email drafts based on your prompts.
This feature requires a Nylas Plus package subscription.

Examples:
  # Generate a new email draft
  nylas email smart-compose --prompt "Draft a thank you email for the meeting"

  # Generate a reply to a specific message
  nylas email smart-compose --message-id <id> --prompt "Reply accepting the invitation"

  # Output as JSON
  nylas email smart-compose --prompt "Write a follow-up email" --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			prompt, _ := cmd.Flags().GetString("prompt")
			messageID, _ := cmd.Flags().GetString("message-id")
			jsonOutput, _ := cmd.Flags().GetBool("json")

			if prompt == "" {
				return common.NewUserError("prompt is required", "Use --prompt to describe the email you want to compose")
			}

			client, err := getClient()
			if err != nil {
				return err
			}

			grantID, err := common.GetGrantID(args)
			if err != nil {
				return err
			}

			ctx, cancel := common.CreateContext()
			defer cancel()

			req := &domain.SmartComposeRequest{
				Prompt: prompt,
			}

			var suggestion *domain.SmartComposeSuggestion
			if messageID != "" {
				suggestion, err = client.SmartComposeReply(ctx, grantID, messageID, req)
			} else {
				suggestion, err = client.SmartCompose(ctx, grantID, req)
			}

			if err != nil {
				return common.WrapGenerateError("email", err)
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(suggestion)
			}

			// Pretty print the suggestion
			fmt.Println("AI-Generated Email:")
			fmt.Println("==================")
			fmt.Println()
			fmt.Println(suggestion.Suggestion)
			fmt.Println()
			fmt.Println("Note: This is an AI-generated suggestion. Please review and edit as needed.")

			return nil
		},
	}

	cmd.Flags().String("prompt", "", "AI instruction for generating the email (required, max 1000 tokens)")
	cmd.Flags().String("message-id", "", "Message ID to reply to (optional, generates reply if provided)")
	_ = cmd.MarkFlagRequired("prompt")

	return cmd
}

// newTrackingInfoCmd creates a command to explain email tracking.
func newTrackingInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tracking-info",
		Short: "Information about email tracking",
		Long: `Information about email tracking features in Nylas.

Nylas supports tracking email opens, link clicks, and thread replies through webhooks.
Tracking is enabled when sending messages using the --track-opens and --track-links flags.

TRACKING FEATURES:
  • Opens: Track when recipients open your emails
  • Clicks: Track when recipients click links in your emails
  • Replies: Track when recipients reply to your messages

ENABLING TRACKING:
When sending an email, use these flags to enable tracking:

  nylas email send --to user@example.com \\
    --subject "Meeting Request" \\
    --body "Let's schedule a meeting" \\
    --track-opens \\
    --track-links

RECEIVING TRACKING DATA:
Tracking data is delivered via webhooks. You need to:

1. Create a webhook for tracking events:
   nylas webhook create --url https://your-server.com/webhooks \\
     --triggers message.opened,message.link_clicked,thread.replied

2. Your webhook endpoint will receive POST requests with tracking data:

   For opens:
   {
     "type": "message.opened",
     "data": {
       "message_id": "abc123",
       "recents": [{
         "opened_id": "open_xyz",
         "timestamp": 1234567890,
         "ip": "192.168.1.1",
         "user_agent": "Mozilla/5.0..."
       }]
     }
   }

   For clicks:
   {
     "type": "message.link_clicked",
     "data": {
       "message_id": "abc123",
       "link_data": [{"url": "https://example.com", "count": 3}],
       "recents": [{
         "click_id": "click_xyz",
         "timestamp": 1234567890,
         "url": "https://example.com",
         "ip": "192.168.1.1"
       }]
     }
   }

WEBHOOK SETUP:
See webhook documentation for more details:
  nylas webhook --help
  nylas webhook create --help

For more information:
  https://developer.nylas.com/docs/v3/email/message-tracking/`,
		Run: func(cmd *cobra.Command, args []string) {
			// The help text is shown automatically
		},
	}

	return cmd
}
