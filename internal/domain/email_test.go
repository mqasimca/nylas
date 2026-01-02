package domain

import (
	"testing"
	"time"
)

// =============================================================================
// Thread Tests
// =============================================================================

func TestThread_Creation(t *testing.T) {
	now := time.Now()
	thread := Thread{
		ID:                    "thread-123",
		GrantID:               "grant-456",
		HasAttachments:        true,
		HasDrafts:             false,
		Starred:               true,
		Unread:                true,
		EarliestMessageDate:   now.AddDate(0, 0, -7),
		LatestMessageRecvDate: now.AddDate(0, 0, -1),
		LatestMessageSentDate: now.AddDate(0, 0, -2),
		Participants: []EmailParticipant{
			{Name: "John Doe", Email: "john@example.com"},
			{Name: "Jane Smith", Email: "jane@example.com"},
		},
		MessageIDs: []string{"msg-1", "msg-2", "msg-3"},
		DraftIDs:   []string{},
		FolderIDs:  []string{"inbox"},
		Snippet:    "Here's the latest update on the project...",
		Subject:    "Project Update",
	}

	if thread.ID != "thread-123" {
		t.Errorf("Thread.ID = %q, want %q", thread.ID, "thread-123")
	}
	if !thread.HasAttachments {
		t.Error("Thread.HasAttachments should be true")
	}
	if !thread.Starred {
		t.Error("Thread.Starred should be true")
	}
	if len(thread.Participants) != 2 {
		t.Errorf("Thread.Participants length = %d, want 2", len(thread.Participants))
	}
	if len(thread.MessageIDs) != 3 {
		t.Errorf("Thread.MessageIDs length = %d, want 3", len(thread.MessageIDs))
	}
}

// =============================================================================
// Draft Tests
// =============================================================================

func TestDraft_Creation(t *testing.T) {
	now := time.Now()
	draft := Draft{
		ID:      "draft-123",
		GrantID: "grant-456",
		Subject: "Meeting Follow-up",
		Body:    "<p>Thanks for the meeting today!</p>",
		From: []EmailParticipant{
			{Name: "Sender", Email: "sender@example.com"},
		},
		To: []EmailParticipant{
			{Name: "Recipient", Email: "recipient@example.com"},
		},
		Cc: []EmailParticipant{
			{Name: "CC Person", Email: "cc@example.com"},
		},
		ReplyToMsgID: "original-msg-123",
		ThreadID:     "thread-456",
		Attachments: []Attachment{
			{Filename: "report.pdf", ContentType: "application/pdf", Size: 1024},
		},
		CreatedAt: now.Add(-1 * time.Hour),
		UpdatedAt: now,
	}

	if draft.Subject != "Meeting Follow-up" {
		t.Errorf("Draft.Subject = %q, want %q", draft.Subject, "Meeting Follow-up")
	}
	if len(draft.To) != 1 {
		t.Errorf("Draft.To length = %d, want 1", len(draft.To))
	}
	if len(draft.Cc) != 1 {
		t.Errorf("Draft.Cc length = %d, want 1", len(draft.Cc))
	}
	if len(draft.Attachments) != 1 {
		t.Errorf("Draft.Attachments length = %d, want 1", len(draft.Attachments))
	}
}

// =============================================================================
// Folder Tests
// =============================================================================

func TestFolder_Creation(t *testing.T) {
	folder := Folder{
		ID:              "folder-123",
		GrantID:         "grant-456",
		Name:            "Important",
		SystemFolder:    "",
		ParentID:        "parent-folder",
		BackgroundColor: "#ff0000",
		TextColor:       "#ffffff",
		TotalCount:      150,
		UnreadCount:     12,
		ChildIDs:        []string{"child-1", "child-2"},
		Attributes:      []string{"user_created"},
	}

	if folder.Name != "Important" {
		t.Errorf("Folder.Name = %q, want %q", folder.Name, "Important")
	}
	if folder.TotalCount != 150 {
		t.Errorf("Folder.TotalCount = %d, want 150", folder.TotalCount)
	}
	if folder.UnreadCount != 12 {
		t.Errorf("Folder.UnreadCount = %d, want 12", folder.UnreadCount)
	}
	if len(folder.ChildIDs) != 2 {
		t.Errorf("Folder.ChildIDs length = %d, want 2", len(folder.ChildIDs))
	}
}

func TestSystemFolderConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		want     string
	}{
		{"inbox", FolderInbox, "inbox"},
		{"sent", FolderSent, "sent"},
		{"drafts", FolderDrafts, "drafts"},
		{"trash", FolderTrash, "trash"},
		{"spam", FolderSpam, "spam"},
		{"archive", FolderArchive, "archive"},
		{"all", FolderAll, "all"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.want {
				t.Errorf("Folder constant = %q, want %q", tt.constant, tt.want)
			}
		})
	}
}

// =============================================================================
// Attachment Tests
// =============================================================================

func TestAttachment_Creation(t *testing.T) {
	attachment := Attachment{
		ID:          "attach-123",
		GrantID:     "grant-456",
		Filename:    "document.pdf",
		ContentType: "application/pdf",
		Size:        2048,
		ContentID:   "cid:image001",
		IsInline:    true,
		Content:     []byte{0x25, 0x50, 0x44, 0x46}, // PDF magic bytes
	}

	if attachment.Filename != "document.pdf" {
		t.Errorf("Attachment.Filename = %q, want %q", attachment.Filename, "document.pdf")
	}
	if attachment.ContentType != "application/pdf" {
		t.Errorf("Attachment.ContentType = %q, want %q", attachment.ContentType, "application/pdf")
	}
	if attachment.Size != 2048 {
		t.Errorf("Attachment.Size = %d, want 2048", attachment.Size)
	}
	if !attachment.IsInline {
		t.Error("Attachment.IsInline should be true")
	}
	if len(attachment.Content) != 4 {
		t.Errorf("Attachment.Content length = %d, want 4", len(attachment.Content))
	}
}

// =============================================================================
// SendMessageRequest Tests
// =============================================================================

func TestSendMessageRequest_Creation(t *testing.T) {
	req := SendMessageRequest{
		Subject: "Test Email",
		Body:    "<p>Hello World</p>",
		From: []EmailParticipant{
			{Name: "Sender", Email: "sender@example.com"},
		},
		To: []EmailParticipant{
			{Name: "Recipient", Email: "recipient@example.com"},
		},
		Cc: []EmailParticipant{
			{Name: "CC", Email: "cc@example.com"},
		},
		Bcc: []EmailParticipant{
			{Name: "BCC", Email: "bcc@example.com"},
		},
		ReplyTo: []EmailParticipant{
			{Name: "Reply To", Email: "reply@example.com"},
		},
		ReplyToMsgID: "msg-123",
		TrackingOpts: &TrackingOptions{
			Opens: true,
			Links: true,
			Label: "campaign-123",
		},
		Attachments: []Attachment{
			{Filename: "file.txt", ContentType: "text/plain", Size: 100},
		},
		SendAt: 1704067200,
		Metadata: map[string]string{
			"campaign_id": "camp-123",
		},
	}

	if req.Subject != "Test Email" {
		t.Errorf("SendMessageRequest.Subject = %q, want %q", req.Subject, "Test Email")
	}
	if len(req.To) != 1 {
		t.Errorf("SendMessageRequest.To length = %d, want 1", len(req.To))
	}
	if req.TrackingOpts == nil {
		t.Fatal("SendMessageRequest.TrackingOpts should not be nil")
	}
	if !req.TrackingOpts.Opens {
		t.Error("TrackingOptions.Opens should be true")
	}
	if req.SendAt != 1704067200 {
		t.Errorf("SendMessageRequest.SendAt = %d, want 1704067200", req.SendAt)
	}
}

// =============================================================================
// ScheduledMessage Tests
// =============================================================================

func TestScheduledMessage_Creation(t *testing.T) {
	tests := []struct {
		name   string
		msg    ScheduledMessage
		status string
	}{
		{
			name: "pending scheduled message",
			msg: ScheduledMessage{
				ScheduleID: "sched-123",
				Status:     "pending",
				CloseTime:  1704067200,
			},
			status: "pending",
		},
		{
			name: "sent scheduled message",
			msg: ScheduledMessage{
				ScheduleID: "sched-456",
				Status:     "sent",
				CloseTime:  1704060000,
			},
			status: "sent",
		},
		{
			name: "cancelled scheduled message",
			msg: ScheduledMessage{
				ScheduleID: "sched-789",
				Status:     "cancelled",
				CloseTime:  1704070000,
			},
			status: "cancelled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.msg.Status != tt.status {
				t.Errorf("ScheduledMessage.Status = %q, want %q", tt.msg.Status, tt.status)
			}
		})
	}
}

// =============================================================================
// TrackingOptions Tests
// =============================================================================

func TestTrackingOptions_Creation(t *testing.T) {
	tests := []struct {
		name string
		opts TrackingOptions
	}{
		{
			name: "full tracking enabled",
			opts: TrackingOptions{
				Opens: true,
				Links: true,
				Label: "newsletter-2024-01",
			},
		},
		{
			name: "opens only",
			opts: TrackingOptions{
				Opens: true,
				Links: false,
			},
		},
		{
			name: "links only",
			opts: TrackingOptions{
				Opens: false,
				Links: true,
			},
		},
		{
			name: "no tracking",
			opts: TrackingOptions{
				Opens: false,
				Links: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify the struct can be created
			_ = tt.opts
		})
	}
}

// =============================================================================
// MessageQueryParams Tests
// =============================================================================

func TestMessageQueryParams_Creation(t *testing.T) {
	unread := true
	starred := false
	hasAttachment := true

	params := MessageQueryParams{
		Limit:          100,
		Offset:         50,
		PageToken:      "token-123",
		Subject:        "important",
		From:           "boss@example.com",
		To:             "me@example.com",
		Cc:             "team@example.com",
		In:             []string{"inbox", "important"},
		Unread:         &unread,
		Starred:        &starred,
		ThreadID:       "thread-123",
		ReceivedBefore: 1704153600,
		ReceivedAfter:  1704067200,
		HasAttachment:  &hasAttachment,
		SearchQuery:    "project update",
		Fields:         "include_headers",
		MetadataPair:   "key1:value1",
	}

	if params.Limit != 100 {
		t.Errorf("MessageQueryParams.Limit = %d, want 100", params.Limit)
	}
	if params.Subject != "important" {
		t.Errorf("MessageQueryParams.Subject = %q, want %q", params.Subject, "important")
	}
	if params.Unread == nil || !*params.Unread {
		t.Error("MessageQueryParams.Unread should be true")
	}
	if len(params.In) != 2 {
		t.Errorf("MessageQueryParams.In length = %d, want 2", len(params.In))
	}
}

// =============================================================================
// ThreadQueryParams Tests
// =============================================================================

func TestThreadQueryParams_Creation(t *testing.T) {
	unread := true
	hasAttachment := true

	params := ThreadQueryParams{
		Limit:           50,
		PageToken:       "cursor-abc",
		Subject:         "meeting",
		From:            "sender@example.com",
		To:              "recipient@example.com",
		In:              []string{"inbox"},
		Unread:          &unread,
		LatestMsgBefore: 1704153600,
		LatestMsgAfter:  1704067200,
		HasAttachment:   &hasAttachment,
		SearchQuery:     "quarterly review",
	}

	if params.Limit != 50 {
		t.Errorf("ThreadQueryParams.Limit = %d, want 50", params.Limit)
	}
	if params.Subject != "meeting" {
		t.Errorf("ThreadQueryParams.Subject = %q, want %q", params.Subject, "meeting")
	}
	if params.HasAttachment == nil || !*params.HasAttachment {
		t.Error("ThreadQueryParams.HasAttachment should be true")
	}
}

// =============================================================================
// UpdateMessageRequest Tests
// =============================================================================

func TestUpdateMessageRequest_Creation(t *testing.T) {
	unread := false
	starred := true

	req := UpdateMessageRequest{
		Unread:  &unread,
		Starred: &starred,
		Folders: []string{"archive", "important"},
	}

	if req.Unread == nil || *req.Unread {
		t.Error("UpdateMessageRequest.Unread should be false")
	}
	if req.Starred == nil || !*req.Starred {
		t.Error("UpdateMessageRequest.Starred should be true")
	}
	if len(req.Folders) != 2 {
		t.Errorf("UpdateMessageRequest.Folders length = %d, want 2", len(req.Folders))
	}
}

// =============================================================================
// CreateDraftRequest Tests
// =============================================================================

func TestCreateDraftRequest_Creation(t *testing.T) {
	req := CreateDraftRequest{
		Subject: "Draft Email",
		Body:    "<p>Draft content</p>",
		To: []EmailParticipant{
			{Email: "to@example.com"},
		},
		Cc: []EmailParticipant{
			{Email: "cc@example.com"},
		},
		ReplyToMsgID: "orig-msg-123",
		Attachments: []Attachment{
			{Filename: "draft-attachment.pdf", Size: 500},
		},
		Metadata: map[string]string{
			"draft_type": "follow_up",
		},
	}

	if req.Subject != "Draft Email" {
		t.Errorf("CreateDraftRequest.Subject = %q, want %q", req.Subject, "Draft Email")
	}
	if len(req.To) != 1 {
		t.Errorf("CreateDraftRequest.To length = %d, want 1", len(req.To))
	}
	if req.ReplyToMsgID != "orig-msg-123" {
		t.Errorf("CreateDraftRequest.ReplyToMsgID = %q, want %q", req.ReplyToMsgID, "orig-msg-123")
	}
}

// =============================================================================
// CreateFolderRequest Tests
// =============================================================================

func TestCreateFolderRequest_Creation(t *testing.T) {
	req := CreateFolderRequest{
		Name:            "Projects",
		ParentID:        "parent-123",
		BackgroundColor: "#0000ff",
		TextColor:       "#ffffff",
	}

	if req.Name != "Projects" {
		t.Errorf("CreateFolderRequest.Name = %q, want %q", req.Name, "Projects")
	}
	if req.ParentID != "parent-123" {
		t.Errorf("CreateFolderRequest.ParentID = %q, want %q", req.ParentID, "parent-123")
	}
}

// =============================================================================
// Pagination Tests
// =============================================================================

func TestPagination_Creation(t *testing.T) {
	tests := []struct {
		name       string
		pagination Pagination
		hasMore    bool
	}{
		{
			name: "has more pages",
			pagination: Pagination{
				NextCursor: "next-page-cursor",
				HasMore:    true,
			},
			hasMore: true,
		},
		{
			name: "last page",
			pagination: Pagination{
				NextCursor: "",
				HasMore:    false,
			},
			hasMore: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.pagination.HasMore != tt.hasMore {
				t.Errorf("Pagination.HasMore = %v, want %v", tt.pagination.HasMore, tt.hasMore)
			}
		})
	}
}

// =============================================================================
// SmartComposeRequest Tests
// =============================================================================

func TestSmartComposeRequest_Creation(t *testing.T) {
	req := SmartComposeRequest{
		Prompt: "Write a polite follow-up email to a client about a delayed project",
	}

	if req.Prompt == "" {
		t.Error("SmartComposeRequest.Prompt should not be empty")
	}
}

// =============================================================================
// SmartComposeSuggestion Tests
// =============================================================================

func TestSmartComposeSuggestion_Creation(t *testing.T) {
	suggestion := SmartComposeSuggestion{
		Suggestion: "Dear Client,\n\nI hope this email finds you well...",
	}

	if suggestion.Suggestion == "" {
		t.Error("SmartComposeSuggestion.Suggestion should not be empty")
	}
}

// =============================================================================
// TrackingData Tests
// =============================================================================

func TestTrackingData_Creation(t *testing.T) {
	now := time.Now()
	data := TrackingData{
		MessageID: "msg-123",
		Opens: []OpenEvent{
			{
				OpenedID:  "open-1",
				Timestamp: now.Add(-1 * time.Hour),
				IPAddress: "192.168.1.1",
				UserAgent: "Mozilla/5.0",
			},
		},
		Clicks: []ClickEvent{
			{
				ClickID:   "click-1",
				Timestamp: now.Add(-30 * time.Minute),
				URL:       "https://example.com/link",
				IPAddress: "192.168.1.1",
				UserAgent: "Mozilla/5.0",
				LinkIndex: 0,
			},
		},
		Replies: []ReplyEvent{
			{
				MessageID:     "reply-msg-1",
				Timestamp:     now,
				ThreadID:      "thread-123",
				RootMessageID: "msg-123",
			},
		},
	}

	if len(data.Opens) != 1 {
		t.Errorf("TrackingData.Opens length = %d, want 1", len(data.Opens))
	}
	if len(data.Clicks) != 1 {
		t.Errorf("TrackingData.Clicks length = %d, want 1", len(data.Clicks))
	}
	if len(data.Replies) != 1 {
		t.Errorf("TrackingData.Replies length = %d, want 1", len(data.Replies))
	}
	if data.Clicks[0].URL != "https://example.com/link" {
		t.Errorf("ClickEvent.URL = %q, want expected URL", data.Clicks[0].URL)
	}
}
