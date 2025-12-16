package email

import (
	"fmt"

	"github.com/mqasimca/nylas/internal/domain"
	"github.com/spf13/cobra"
)

func newFoldersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "folders",
		Short: "Manage email folders/labels",
		Long:  "List, create, update, and delete email folders or labels.",
	}

	cmd.AddCommand(newFoldersListCmd())
	cmd.AddCommand(newFoldersCreateCmd())
	cmd.AddCommand(newFoldersDeleteCmd())

	return cmd
}

func newFoldersListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list [grant-id]",
		Short: "List all folders",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}

			grantID, err := getGrantID(args)
			if err != nil {
				return err
			}

			ctx, cancel := createContext()
			defer cancel()

			folders, err := client.GetFolders(ctx, grantID)
			if err != nil {
				return fmt.Errorf("failed to get folders: %w", err)
			}

			if len(folders) == 0 {
				fmt.Println("No folders found.")
				return nil
			}

			fmt.Println("Folders:")
			fmt.Println()
			fmt.Printf("%-30s %-12s %8s %8s\n", "NAME", "TYPE", "TOTAL", "UNREAD")
			fmt.Println("------------------------------------------------------------")

			for _, f := range folders {
				folderType := f.SystemFolder
				if folderType == "" {
					folderType = "custom"
				}

				name := f.Name
				if len(name) > 28 {
					name = name[:25] + "..."
				}

				unreadStr := fmt.Sprintf("%d", f.UnreadCount)
				if f.UnreadCount > 0 {
					unreadStr = cyan.Sprintf("%d", f.UnreadCount)
				}

				fmt.Printf("%-30s %-12s %8d %8s\n",
					name, folderType, f.TotalCount, unreadStr)
			}

			fmt.Println()
			dim.Printf("Folder IDs can be used with --folder flag in email list/search\n")

			return nil
		},
	}
}

func newFoldersCreateCmd() *cobra.Command {
	var parentID string
	var bgColor string
	var textColor string

	cmd := &cobra.Command{
		Use:   "create <name> [grant-id]",
		Short: "Create a new folder",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			client, err := getClient()
			if err != nil {
				return err
			}

			var grantID string
			if len(args) > 1 {
				grantID = args[1]
			} else {
				grantID, err = getGrantID(nil)
				if err != nil {
					return err
				}
			}

			ctx, cancel := createContext()
			defer cancel()

			req := &domain.CreateFolderRequest{
				Name:            name,
				ParentID:        parentID,
				BackgroundColor: bgColor,
				TextColor:       textColor,
			}

			folder, err := client.CreateFolder(ctx, grantID, req)
			if err != nil {
				return fmt.Errorf("failed to create folder: %w", err)
			}

			printSuccess("Created folder '%s' (ID: %s)", folder.Name, folder.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&parentID, "parent", "", "Parent folder ID")
	cmd.Flags().StringVar(&bgColor, "bg-color", "", "Background color (hex)")
	cmd.Flags().StringVar(&textColor, "text-color", "", "Text color (hex)")

	return cmd
}

func newFoldersDeleteCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <folder-id> [grant-id]",
		Short: "Delete a folder",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			folderID := args[0]

			client, err := getClient()
			if err != nil {
				return err
			}

			var grantID string
			if len(args) > 1 {
				grantID = args[1]
			} else {
				grantID, err = getGrantID(nil)
				if err != nil {
					return err
				}
			}

			if !force {
				fmt.Printf("Delete folder %s? [y/N]: ", folderID)
				var confirm string
				fmt.Scanln(&confirm)
				if confirm != "y" && confirm != "Y" && confirm != "yes" {
					fmt.Println("Cancelled.")
					return nil
				}
			}

			ctx, cancel := createContext()
			defer cancel()

			err = client.DeleteFolder(ctx, grantID, folderID)
			if err != nil {
				return fmt.Errorf("failed to delete folder: %w", err)
			}

			printSuccess("Folder deleted")
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation")

	return cmd
}
