package notetaker

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func newMediaCmd() *cobra.Command {
	var outputJSON bool

	cmd := &cobra.Command{
		Use:   "media <notetaker-id> [grant-id]",
		Short: "Get notetaker media (recording and transcript)",
		Long: `Retrieve media files from a completed notetaker session.

Returns URLs to download:
- Recording: Video/audio recording of the meeting
- Transcript: Text transcript of the meeting

Note: Media URLs have an expiration time. Download them promptly.`,
		Example: `  # Get media URLs
  nylas notetaker media abc123

  # Output as JSON
  nylas notetaker media abc123 --json`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}

			notetakerID := args[0]
			grantID, err := getGrantID(args[1:])
			if err != nil {
				return err
			}

			ctx, cancel := createContext()
			defer cancel()

			media, err := client.GetNotetakerMedia(ctx, grantID, notetakerID)
			if err != nil {
				return fmt.Errorf("failed to get notetaker media: %w", err)
			}

			if outputJSON {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(media)
			}

			cyan := color.New(color.FgCyan, color.Bold)
			green := color.New(color.FgGreen)
			dim := color.New(color.Faint)

			if media.Recording == nil && media.Transcript == nil {
				fmt.Println("No media available yet.")
				fmt.Println("Media is generated after the meeting ends and processing completes.")
				return nil
			}

			_, _ = cyan.Println("Notetaker Media:")
			fmt.Println()

			if media.Recording != nil {
				_, _ = green.Println("Recording:")
				fmt.Printf("  URL:  %s\n", media.Recording.URL)
				if media.Recording.ContentType != "" {
					_, _ = dim.Printf("  Type: %s\n", media.Recording.ContentType)
				}
				if media.Recording.Size > 0 {
					_, _ = dim.Printf("  Size: %s\n", formatBytes(media.Recording.Size))
				}
				if media.Recording.ExpiresAt > 0 {
					expires := time.Unix(media.Recording.ExpiresAt, 0)
					_, _ = dim.Printf("  Expires: %s\n", expires.Local().Format("Mon Jan 2, 2006 3:04 PM MST"))
				}
				fmt.Println()
			}

			if media.Transcript != nil {
				_, _ = green.Println("Transcript:")
				fmt.Printf("  URL:  %s\n", media.Transcript.URL)
				if media.Transcript.ContentType != "" {
					_, _ = dim.Printf("  Type: %s\n", media.Transcript.ContentType)
				}
				if media.Transcript.Size > 0 {
					_, _ = dim.Printf("  Size: %s\n", formatBytes(media.Transcript.Size))
				}
				if media.Transcript.ExpiresAt > 0 {
					expires := time.Unix(media.Transcript.ExpiresAt, 0)
					_, _ = dim.Printf("  Expires: %s\n", expires.Local().Format("Mon Jan 2, 2006 3:04 PM MST"))
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")

	return cmd
}

// formatBytes formats bytes into a human-readable string.
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
