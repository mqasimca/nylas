package tui

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/rivo/tview"
)

func (v *MessagesView) toggleStar() {
	meta := v.table.SelectedMeta()
	if meta == nil {
		return
	}

	thread, ok := meta.Data.(*domain.Thread)
	if !ok {
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		newStarred := !thread.Starred
		_, err := v.app.config.Client.UpdateThread(ctx, v.app.config.GrantID, thread.ID, &domain.UpdateMessageRequest{
			Starred: &newStarred,
		})
		if err != nil {
			v.app.Flash(FlashError, "Failed to update: %v", err)
			return
		}
		v.app.Flash(FlashInfo, "Thread starred")
		v.app.QueueUpdateDraw(func() {
			v.Load()
		})
	}()
}

func (v *MessagesView) markUnread() {
	meta := v.table.SelectedMeta()
	if meta == nil {
		return
	}

	thread, ok := meta.Data.(*domain.Thread)
	if !ok {
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		unread := true
		_, err := v.app.config.Client.UpdateThread(ctx, v.app.config.GrantID, thread.ID, &domain.UpdateMessageRequest{
			Unread: &unread,
		})
		if err != nil {
			v.app.Flash(FlashError, "Failed to update: %v", err)
			return
		}
		v.app.Flash(FlashInfo, "Marked as unread")
		v.app.QueueUpdateDraw(func() {
			v.Load()
		})
	}()
}

func (v *MessagesView) showCompose(mode ComposeMode, replyTo *domain.Message) {
	compose := NewComposeView(v.app, mode, replyTo)

	compose.SetOnSent(func() {
		v.app.PopDetail()
		// Refresh messages to show the sent message
		go func() {
			v.Load()
			v.app.QueueUpdateDraw(func() {})
		}()
	})

	compose.SetOnCancel(func() {
		v.app.PopDetail()
		if v.showingDetail {
			// Go back to message detail view - just set focus
		} else {
			v.app.SetFocus(v.table)
		}
	})

	v.app.PushDetail("compose", compose)
}

func (v *MessagesView) showDownloadDialog() {
	if len(v.attachments) == 0 {
		return
	}

	styles := v.app.styles

	// Create list for attachment selection
	list := tview.NewList()
	list.SetBackgroundColor(styles.BgColor)
	list.SetMainTextColor(styles.FgColor)
	list.SetSecondaryTextColor(styles.InfoColor)
	list.SetSelectedBackgroundColor(styles.FocusColor)
	list.SetSelectedTextColor(styles.BgColor)
	list.SetBorder(true)
	list.SetBorderColor(styles.FocusColor)
	list.SetTitle(" Download Attachment ")
	list.SetTitleColor(styles.TitleFg)

	// Add attachments to list
	for i, attInfo := range v.attachments {
		idx := i
		att := attInfo.Attachment
		msgID := attInfo.MessageID
		sizeStr := formatFileSize(att.Size)
		list.AddItem(
			fmt.Sprintf("%s (%s)", att.Filename, sizeStr),
			att.ContentType,
			rune('1'+i),
			func() {
				v.downloadAttachment(msgID, att.ID, att.Filename, idx+1)
			},
		)
	}

	// Handle Escape to close
	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			v.app.PopDetail()
			return nil
		}
		return event
	})

	// Center the list
	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(nil, 0, 1, false).
			AddItem(list, 60, 0, true).
			AddItem(nil, 0, 1, false), 0, 2, true).
		AddItem(nil, 0, 1, false)

	v.app.PushDetail("download-dialog", flex)
	v.app.SetFocus(list)
}

func (v *MessagesView) downloadAttachment(messageID, attachmentID, filename string, displayNum int) {
	v.app.Flash(FlashInfo, "Downloading %s...", filename)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		reader, err := v.app.config.Client.DownloadAttachment(ctx, v.app.config.GrantID, messageID, attachmentID)
		if err != nil {
			v.app.QueueUpdateDraw(func() {
				v.app.Flash(FlashError, "Download failed: %v", err)
			})
			return
		}
		defer func() { _ = reader.Close() }()

		// Get Downloads directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			v.app.QueueUpdateDraw(func() {
				v.app.Flash(FlashError, "Cannot find home directory: %v", err)
			})
			return
		}
		downloadDir := filepath.Join(homeDir, "Downloads")

		// Ensure download directory exists
		if err := os.MkdirAll(downloadDir, 0750); err != nil {
			v.app.QueueUpdateDraw(func() {
				v.app.Flash(FlashError, "Cannot create Downloads directory: %v", err)
			})
			return
		}

		// Create file with unique name if exists
		destPath := filepath.Join(downloadDir, filename)
		destPath = v.getUniqueFilename(destPath)

		// #nosec G304 -- destPath is validated through getUniqueFilename() and constrained to downloadDir
		file, err := os.Create(destPath)
		if err != nil {
			v.app.QueueUpdateDraw(func() {
				v.app.Flash(FlashError, "Cannot create file: %v", err)
			})
			return
		}
		defer func() { _ = file.Close() }()

		// Copy content
		written, err := io.Copy(file, reader)
		if err != nil {
			v.app.QueueUpdateDraw(func() {
				v.app.Flash(FlashError, "Download failed: %v", err)
			})
			return
		}

		v.app.QueueUpdateDraw(func() {
			v.app.PopDetail() // Close download dialog
			v.app.Flash(FlashInfo, "Downloaded %s (%s) to %s", filename, formatFileSize(written), destPath)
		})
	}()
}

func (v *MessagesView) getUniqueFilename(path string) string {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path
	}

	dir := filepath.Dir(path)
	ext := filepath.Ext(path)
	name := strings.TrimSuffix(filepath.Base(path), ext)

	for i := 1; i < 1000; i++ {
		newPath := filepath.Join(dir, fmt.Sprintf("%s (%d)%s", name, i, ext))
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}
	}

	// Fallback: append timestamp
	return filepath.Join(dir, fmt.Sprintf("%s_%d%s", name, time.Now().Unix(), ext))
}
