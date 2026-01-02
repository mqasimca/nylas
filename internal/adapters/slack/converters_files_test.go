//go:build !integration
// +build !integration

package slack

import (
	"testing"

	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
)

func TestConvertFiles(t *testing.T) {
	tests := []struct {
		name    string
		files   []slack.File
		wantLen int
	}{
		{
			name:    "nil files returns nil",
			files:   nil,
			wantLen: 0,
		},
		{
			name:    "empty slice returns nil",
			files:   []slack.File{},
			wantLen: 0,
		},
		{
			name: "single file",
			files: []slack.File{
				{
					ID:       "F12345",
					Name:     "document.pdf",
					Title:    "Important Document",
					Mimetype: "application/pdf",
					Filetype: "pdf",
					Size:     1024,
					User:     "U12345",
				},
			},
			wantLen: 1,
		},
		{
			name: "multiple files",
			files: []slack.File{
				{ID: "F1", Name: "image.png", Mimetype: "image/png", Size: 512},
				{ID: "F2", Name: "doc.pdf", Mimetype: "application/pdf", Size: 1024},
				{ID: "F3", Name: "video.mp4", Mimetype: "video/mp4", Size: 2048},
			},
			wantLen: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertFiles(tt.files)

			if tt.wantLen == 0 {
				assert.Nil(t, got)
			} else {
				assert.Len(t, got, tt.wantLen)
				for i, f := range tt.files {
					assert.Equal(t, f.ID, got[i].ID)
					assert.Equal(t, f.Name, got[i].Name)
					assert.Equal(t, f.Mimetype, got[i].MimeType)
				}
			}
		})
	}
}

func TestConvertFile(t *testing.T) {
	tests := []struct {
		name string
		file slack.File
		want struct {
			id          string
			name        string
			title       string
			mimeType    string
			fileType    string
			size        int64
			downloadURL string
			permalink   string
			userID      string
			imageWidth  int
			imageHeight int
			thumb360    string
			thumb480    string
		}
	}{
		{
			name: "basic document file",
			file: slack.File{
				ID:                 "F12345",
				Name:               "report.pdf",
				Title:              "Monthly Report",
				Mimetype:           "application/pdf",
				Filetype:           "pdf",
				Size:               2048,
				URLPrivateDownload: "https://files.slack.com/download/report.pdf",
				Permalink:          "https://workspace.slack.com/files/report.pdf",
				User:               "U12345",
			},
			want: struct {
				id          string
				name        string
				title       string
				mimeType    string
				fileType    string
				size        int64
				downloadURL string
				permalink   string
				userID      string
				imageWidth  int
				imageHeight int
				thumb360    string
				thumb480    string
			}{
				id:          "F12345",
				name:        "report.pdf",
				title:       "Monthly Report",
				mimeType:    "application/pdf",
				fileType:    "pdf",
				size:        2048,
				downloadURL: "https://files.slack.com/download/report.pdf",
				permalink:   "https://workspace.slack.com/files/report.pdf",
				userID:      "U12345",
				imageWidth:  0,
				imageHeight: 0,
				thumb360:    "",
				thumb480:    "",
			},
		},
		{
			name: "image file with thumbnails",
			file: slack.File{
				ID:                 "F67890",
				Name:               "screenshot.png",
				Title:              "App Screenshot",
				Mimetype:           "image/png",
				Filetype:           "png",
				Size:               512000,
				URLPrivateDownload: "https://files.slack.com/download/screenshot.png",
				Permalink:          "https://workspace.slack.com/files/screenshot.png",
				User:               "U99999",
				OriginalW:          1920,
				OriginalH:          1080,
				Thumb360:           "https://files.slack.com/thumb360/screenshot.png",
				Thumb480:           "https://files.slack.com/thumb480/screenshot.png",
			},
			want: struct {
				id          string
				name        string
				title       string
				mimeType    string
				fileType    string
				size        int64
				downloadURL string
				permalink   string
				userID      string
				imageWidth  int
				imageHeight int
				thumb360    string
				thumb480    string
			}{
				id:          "F67890",
				name:        "screenshot.png",
				title:       "App Screenshot",
				mimeType:    "image/png",
				fileType:    "png",
				size:        512000,
				downloadURL: "https://files.slack.com/download/screenshot.png",
				permalink:   "https://workspace.slack.com/files/screenshot.png",
				userID:      "U99999",
				imageWidth:  1920,
				imageHeight: 1080,
				thumb360:    "https://files.slack.com/thumb360/screenshot.png",
				thumb480:    "https://files.slack.com/thumb480/screenshot.png",
			},
		},
		{
			name: "empty file fields",
			file: slack.File{
				ID: "F00000",
			},
			want: struct {
				id          string
				name        string
				title       string
				mimeType    string
				fileType    string
				size        int64
				downloadURL string
				permalink   string
				userID      string
				imageWidth  int
				imageHeight int
				thumb360    string
				thumb480    string
			}{
				id:          "F00000",
				name:        "",
				title:       "",
				mimeType:    "",
				fileType:    "",
				size:        0,
				downloadURL: "",
				permalink:   "",
				userID:      "",
				imageWidth:  0,
				imageHeight: 0,
				thumb360:    "",
				thumb480:    "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertFile(tt.file)

			assert.Equal(t, tt.want.id, got.ID)
			assert.Equal(t, tt.want.name, got.Name)
			assert.Equal(t, tt.want.title, got.Title)
			assert.Equal(t, tt.want.mimeType, got.MimeType)
			assert.Equal(t, tt.want.fileType, got.FileType)
			assert.Equal(t, tt.want.size, got.Size)
			assert.Equal(t, tt.want.downloadURL, got.DownloadURL)
			assert.Equal(t, tt.want.permalink, got.Permalink)
			assert.Equal(t, tt.want.userID, got.UserID)
			assert.Equal(t, tt.want.imageWidth, got.ImageWidth)
			assert.Equal(t, tt.want.imageHeight, got.ImageHeight)
			assert.Equal(t, tt.want.thumb360, got.Thumb360)
			assert.Equal(t, tt.want.thumb480, got.Thumb480)
		})
	}
}

func TestConvertMessage_WithFiles(t *testing.T) {
	msg := slack.Message{
		Msg: slack.Msg{
			Timestamp: "1234567890.123456",
			User:      "U12345",
			Text:      "Check out this file",
			Files: []slack.File{
				{
					ID:       "F12345",
					Name:     "document.pdf",
					Mimetype: "application/pdf",
					Size:     1024,
				},
			},
		},
	}

	got := convertMessage(msg, "C12345")

	assert.Len(t, got.Attachments, 1)
	assert.Equal(t, "F12345", got.Attachments[0].ID)
	assert.Equal(t, "document.pdf", got.Attachments[0].Name)
}

func TestConvertMessage_WithMultipleFiles(t *testing.T) {
	msg := slack.Message{
		Msg: slack.Msg{
			Timestamp: "1234567890.123456",
			User:      "U12345",
			Text:      "Multiple files attached",
			Files: []slack.File{
				{ID: "F1", Name: "file1.txt"},
				{ID: "F2", Name: "file2.pdf"},
				{ID: "F3", Name: "file3.png"},
			},
		},
	}

	got := convertMessage(msg, "C12345")

	assert.Len(t, got.Attachments, 3)
	assert.Equal(t, "F1", got.Attachments[0].ID)
	assert.Equal(t, "F2", got.Attachments[1].ID)
	assert.Equal(t, "F3", got.Attachments[2].ID)
}

func TestConvertMessage_EmptyFiles(t *testing.T) {
	msg := slack.Message{
		Msg: slack.Msg{
			Timestamp: "1234567890.123456",
			User:      "U12345",
			Text:      "No files here",
			Files:     nil,
		},
	}

	got := convertMessage(msg, "C12345")

	assert.Nil(t, got.Attachments)
}

func TestConvertChannel_WithAllFields(t *testing.T) {
	ch := slack.Channel{
		GroupConversation: slack.GroupConversation{
			Conversation: slack.Conversation{
				ID:          "C12345",
				IsShared:    true,
				IsOrgShared: true,
				IsExtShared: false,
				IsGroup:     false,
				NumMembers:  150,
			},
			Name:       "shared-channel",
			Topic:      slack.Topic{Value: "Cross-team collaboration"},
			Purpose:    slack.Purpose{Value: "Shared workspace channel"},
			IsArchived: false,
		},
		IsChannel: true,
		IsMember:  true,
	}

	got := convertChannel(ch)

	assert.Equal(t, "C12345", got.ID)
	assert.Equal(t, "shared-channel", got.Name)
	assert.True(t, got.IsChannel)
	assert.True(t, got.IsMember)
	assert.True(t, got.IsShared)
	assert.True(t, got.IsOrgShared)
	assert.False(t, got.IsExtShared)
	assert.Equal(t, "Cross-team collaboration", got.Topic)
	assert.Equal(t, "Shared workspace channel", got.Purpose)
	assert.Equal(t, 150, got.MemberCount)
	assert.False(t, got.IsArchived)
}
