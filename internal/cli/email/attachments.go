package email

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// newAttachmentsCmd creates the attachments command group.
func newAttachmentsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attachments",
		Short: "Manage email attachments",
		Long:  "Commands to list, view, and download email attachments.",
	}

	cmd.AddCommand(newAttachmentsListCmd())
	cmd.AddCommand(newAttachmentsShowCmd())
	cmd.AddCommand(newAttachmentsDownloadCmd())

	return cmd
}

// newAttachmentsListCmd creates the attachments list command.
func newAttachmentsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <message-id> [grant-id]",
		Short: "List attachments in a message",
		Long:  "List all attachments in a specific email message.",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			messageID := args[0]

			client, err := getClient()
			if err != nil {
				return err
			}

			grantID, err := getGrantID(args[1:])
			if err != nil {
				return err
			}

			ctx, cancel := createContext()
			defer cancel()

			attachments, err := client.ListAttachments(ctx, grantID, messageID)
			if err != nil {
				return fmt.Errorf("failed to list attachments: %w", err)
			}

			if len(attachments) == 0 {
				fmt.Println("No attachments found in this message.")
				return nil
			}

			fmt.Printf("Found %d attachment(s):\n\n", len(attachments))
			fmt.Println(strings.Repeat("─", 70))

			for i, a := range attachments {
				fmt.Printf("%d. %s\n", i+1, boldWhite.Sprint(a.Filename))
				fmt.Printf("   ID:   %s\n", a.ID)
				fmt.Printf("   Type: %s\n", a.ContentType)
				fmt.Printf("   Size: %s\n", formatSize(a.Size))
				if a.IsInline {
					fmt.Printf("   Inline: yes\n")
				}
				if i < len(attachments)-1 {
					fmt.Println()
				}
			}

			fmt.Println(strings.Repeat("─", 70))
			fmt.Printf("\nUse 'nylas email attachments download <attachment-id> <message-id>' to download.\n")

			return nil
		},
	}

	return cmd
}

// newAttachmentsShowCmd creates the attachments show command.
func newAttachmentsShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <attachment-id> <message-id> [grant-id]",
		Short: "Show attachment metadata",
		Long:  "Display detailed metadata for a specific attachment.",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			attachmentID := args[0]
			messageID := args[1]

			client, err := getClient()
			if err != nil {
				return err
			}

			grantID, err := getGrantID(args[2:])
			if err != nil {
				return err
			}

			ctx, cancel := createContext()
			defer cancel()

			attachment, err := client.GetAttachment(ctx, grantID, messageID, attachmentID)
			if err != nil {
				return fmt.Errorf("failed to get attachment: %w", err)
			}

			fmt.Println(strings.Repeat("─", 60))
			boldWhite.Printf("Filename:     %s\n", attachment.Filename)
			fmt.Printf("ID:           %s\n", attachment.ID)
			fmt.Printf("Content Type: %s\n", attachment.ContentType)
			fmt.Printf("Size:         %s (%d bytes)\n", formatSize(attachment.Size), attachment.Size)
			if attachment.ContentID != "" {
				fmt.Printf("Content ID:   %s\n", attachment.ContentID)
			}
			fmt.Printf("Inline:       %v\n", attachment.IsInline)
			fmt.Println(strings.Repeat("─", 60))

			return nil
		},
	}

	return cmd
}

// newAttachmentsDownloadCmd creates the attachments download command.
func newAttachmentsDownloadCmd() *cobra.Command {
	var outputPath string

	cmd := &cobra.Command{
		Use:   "download <attachment-id> <message-id> [grant-id]",
		Short: "Download an attachment",
		Long:  "Download an attachment to a local file.",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			attachmentID := args[0]
			messageID := args[1]

			client, err := getClient()
			if err != nil {
				return err
			}

			grantID, err := getGrantID(args[2:])
			if err != nil {
				return err
			}

			ctx, cancel := createContext()
			defer cancel()

			// Get attachment metadata first to get filename
			attachment, err := client.GetAttachment(ctx, grantID, messageID, attachmentID)
			if err != nil {
				return fmt.Errorf("failed to get attachment metadata: %w", err)
			}

			// Determine output path
			if outputPath == "" {
				outputPath = attachment.Filename
			}

			// If outputPath is a directory, append filename
			if info, err := os.Stat(outputPath); err == nil && info.IsDir() {
				outputPath = filepath.Join(outputPath, attachment.Filename)
			}

			// Download the attachment
			reader, err := client.DownloadAttachment(ctx, grantID, messageID, attachmentID)
			if err != nil {
				return fmt.Errorf("failed to download attachment: %w", err)
			}
			defer reader.Close()

			// Create output file
			file, err := os.Create(outputPath)
			if err != nil {
				return fmt.Errorf("failed to create output file: %w", err)
			}
			defer file.Close()

			// Copy content
			written, err := io.Copy(file, reader)
			if err != nil {
				return fmt.Errorf("failed to write file: %w", err)
			}

			printSuccess("Downloaded %s (%s) to %s", attachment.Filename, formatSize(written), outputPath)
			return nil
		},
	}

	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output file path (default: original filename)")

	return cmd
}
