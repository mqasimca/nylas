package nylas_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mqasimca/nylas/internal/adapters/nylas"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *nylas.HTTPClient) {
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	client := nylas.NewHTTPClient()
	client.SetCredentials("test-client-id", "test-client-secret", "test-api-key")
	// Override the base URL to use our test server
	// We need to use reflection or a setter method; for now we'll test via the mock
	return server, client
}

// Test mock client implements interface
func TestMockClientImplementsInterface(t *testing.T) {
	var _ interface {
		SetRegion(region string)
		SetCredentials(clientID, clientSecret, apiKey string)
		BuildAuthURL(provider domain.Provider, redirectURI string) string
	} = nylas.NewMockClient()
}

func TestNewHTTPClient(t *testing.T) {
	client := nylas.NewHTTPClient()
	assert.NotNil(t, client)
}

func TestHTTPClient_SetRegion(t *testing.T) {
	client := nylas.NewHTTPClient()

	t.Run("sets US region by default", func(t *testing.T) {
		client.SetRegion("us")
		url := client.BuildAuthURL(domain.ProviderGoogle, "http://localhost")
		assert.Contains(t, url, "api.us.nylas.com")
	})

	t.Run("sets EU region", func(t *testing.T) {
		client.SetRegion("eu")
		url := client.BuildAuthURL(domain.ProviderGoogle, "http://localhost")
		assert.Contains(t, url, "api.eu.nylas.com")
	})
}

func TestHTTPClient_SetCredentials(t *testing.T) {
	client := nylas.NewHTTPClient()
	client.SetCredentials("my-client-id", "my-secret", "my-api-key")

	url := client.BuildAuthURL(domain.ProviderGoogle, "http://localhost")
	assert.Contains(t, url, "client_id=my-client-id")
}

func TestHTTPClient_BuildAuthURL(t *testing.T) {
	client := nylas.NewHTTPClient()
	client.SetCredentials("test-client-id", "", "")

	tests := []struct {
		name        string
		provider    domain.Provider
		redirectURI string
		wantContain []string
	}{
		{
			name:        "Google provider",
			provider:    domain.ProviderGoogle,
			redirectURI: "http://localhost:8080/callback",
			wantContain: []string{
				"provider=google",
				"redirect_uri=http",
				"client_id=test-client-id",
				"response_type=code",
			},
		},
		{
			name:        "Microsoft provider",
			provider:    domain.ProviderMicrosoft,
			redirectURI: "http://localhost:8080/callback",
			wantContain: []string{
				"provider=microsoft",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := client.BuildAuthURL(tt.provider, tt.redirectURI)
			for _, want := range tt.wantContain {
				assert.Contains(t, url, want)
			}
		})
	}
}

func TestMockClient_Messages(t *testing.T) {
	ctx := context.Background()
	mock := nylas.NewMockClient()

	t.Run("GetMessages", func(t *testing.T) {
		mock.GetMessagesFunc = func(ctx context.Context, grantID string, limit int) ([]domain.Message, error) {
			return []domain.Message{
				{ID: "msg-1", Subject: "Test 1"},
				{ID: "msg-2", Subject: "Test 2"},
			}, nil
		}

		messages, err := mock.GetMessages(ctx, "grant-123", 10)
		require.NoError(t, err)
		assert.Len(t, messages, 2)
		assert.True(t, mock.GetMessagesCalled)
		assert.Equal(t, "grant-123", mock.LastGrantID)
	})

	t.Run("GetMessagesWithParams", func(t *testing.T) {
		unread := true
		params := &domain.MessageQueryParams{
			Limit:  5,
			Unread: &unread,
			From:   "sender@example.com",
		}

		mock.GetMessagesWithParamsFunc = func(ctx context.Context, grantID string, p *domain.MessageQueryParams) ([]domain.Message, error) {
			assert.Equal(t, params, p)
			return []domain.Message{{ID: "msg-1"}}, nil
		}

		messages, err := mock.GetMessagesWithParams(ctx, "grant-123", params)
		require.NoError(t, err)
		assert.Len(t, messages, 1)
		assert.True(t, mock.GetMessagesWithParamsCalled)
	})

	t.Run("GetMessage", func(t *testing.T) {
		msg, err := mock.GetMessage(ctx, "grant-123", "msg-456")
		require.NoError(t, err)
		assert.Equal(t, "msg-456", msg.ID)
		assert.True(t, mock.GetMessageCalled)
		assert.Equal(t, "msg-456", mock.LastMessageID)
	})

	t.Run("SendMessage", func(t *testing.T) {
		req := &domain.SendMessageRequest{
			Subject: "Test Subject",
			Body:    "Test Body",
			To:      []domain.EmailParticipant{{Email: "recipient@example.com"}},
		}

		msg, err := mock.SendMessage(ctx, "grant-123", req)
		require.NoError(t, err)
		assert.Equal(t, "Test Subject", msg.Subject)
		assert.True(t, mock.SendMessageCalled)
	})

	t.Run("UpdateMessage", func(t *testing.T) {
		unread := false
		starred := true
		req := &domain.UpdateMessageRequest{
			Unread:  &unread,
			Starred: &starred,
		}

		msg, err := mock.UpdateMessage(ctx, "grant-123", "msg-456", req)
		require.NoError(t, err)
		assert.False(t, msg.Unread)
		assert.True(t, msg.Starred)
		assert.True(t, mock.UpdateMessageCalled)
	})

	t.Run("DeleteMessage", func(t *testing.T) {
		err := mock.DeleteMessage(ctx, "grant-123", "msg-456")
		require.NoError(t, err)
		assert.True(t, mock.DeleteMessageCalled)
	})
}

func TestMockClient_Threads(t *testing.T) {
	ctx := context.Background()
	mock := nylas.NewMockClient()

	t.Run("GetThreads", func(t *testing.T) {
		mock.GetThreadsFunc = func(ctx context.Context, grantID string, params *domain.ThreadQueryParams) ([]domain.Thread, error) {
			return []domain.Thread{
				{ID: "thread-1", Subject: "Thread 1"},
				{ID: "thread-2", Subject: "Thread 2"},
			}, nil
		}

		threads, err := mock.GetThreads(ctx, "grant-123", nil)
		require.NoError(t, err)
		assert.Len(t, threads, 2)
		assert.True(t, mock.GetThreadsCalled)
	})

	t.Run("GetThread", func(t *testing.T) {
		thread, err := mock.GetThread(ctx, "grant-123", "thread-456")
		require.NoError(t, err)
		assert.Equal(t, "thread-456", thread.ID)
		assert.True(t, mock.GetThreadCalled)
		assert.Equal(t, "thread-456", mock.LastThreadID)
	})

	t.Run("UpdateThread", func(t *testing.T) {
		unread := false
		req := &domain.UpdateMessageRequest{
			Unread: &unread,
		}

		thread, err := mock.UpdateThread(ctx, "grant-123", "thread-456", req)
		require.NoError(t, err)
		assert.False(t, thread.Unread)
		assert.True(t, mock.UpdateThreadCalled)
	})

	t.Run("DeleteThread", func(t *testing.T) {
		err := mock.DeleteThread(ctx, "grant-123", "thread-456")
		require.NoError(t, err)
		assert.True(t, mock.DeleteThreadCalled)
	})
}

func TestMockClient_Drafts(t *testing.T) {
	ctx := context.Background()
	mock := nylas.NewMockClient()

	t.Run("GetDrafts", func(t *testing.T) {
		mock.GetDraftsFunc = func(ctx context.Context, grantID string, limit int) ([]domain.Draft, error) {
			return []domain.Draft{
				{ID: "draft-1", Subject: "Draft 1"},
			}, nil
		}

		drafts, err := mock.GetDrafts(ctx, "grant-123", 10)
		require.NoError(t, err)
		assert.Len(t, drafts, 1)
		assert.True(t, mock.GetDraftsCalled)
	})

	t.Run("GetDraft", func(t *testing.T) {
		draft, err := mock.GetDraft(ctx, "grant-123", "draft-456")
		require.NoError(t, err)
		assert.Equal(t, "draft-456", draft.ID)
		assert.True(t, mock.GetDraftCalled)
	})

	t.Run("CreateDraft", func(t *testing.T) {
		req := &domain.CreateDraftRequest{
			Subject: "New Draft",
			Body:    "Draft body",
			To:      []domain.EmailParticipant{{Email: "to@example.com"}},
		}

		draft, err := mock.CreateDraft(ctx, "grant-123", req)
		require.NoError(t, err)
		assert.Equal(t, "New Draft", draft.Subject)
		assert.True(t, mock.CreateDraftCalled)
	})

	t.Run("UpdateDraft", func(t *testing.T) {
		req := &domain.CreateDraftRequest{
			Subject: "Updated Draft",
			Body:    "Updated body",
		}

		draft, err := mock.UpdateDraft(ctx, "grant-123", "draft-456", req)
		require.NoError(t, err)
		assert.Equal(t, "Updated Draft", draft.Subject)
		assert.True(t, mock.UpdateDraftCalled)
	})

	t.Run("DeleteDraft", func(t *testing.T) {
		err := mock.DeleteDraft(ctx, "grant-123", "draft-456")
		require.NoError(t, err)
		assert.True(t, mock.DeleteDraftCalled)
	})

	t.Run("SendDraft", func(t *testing.T) {
		msg, err := mock.SendDraft(ctx, "grant-123", "draft-456")
		require.NoError(t, err)
		assert.NotEmpty(t, msg.ID)
		assert.True(t, mock.SendDraftCalled)
	})
}

func TestMockClient_Folders(t *testing.T) {
	ctx := context.Background()
	mock := nylas.NewMockClient()

	t.Run("GetFolders", func(t *testing.T) {
		folders, err := mock.GetFolders(ctx, "grant-123")
		require.NoError(t, err)
		assert.Len(t, folders, 3) // Default mock returns inbox, sent, drafts
		assert.True(t, mock.GetFoldersCalled)
	})

	t.Run("GetFolder", func(t *testing.T) {
		folder, err := mock.GetFolder(ctx, "grant-123", "folder-456")
		require.NoError(t, err)
		assert.Equal(t, "folder-456", folder.ID)
		assert.True(t, mock.GetFolderCalled)
	})

	t.Run("CreateFolder", func(t *testing.T) {
		req := &domain.CreateFolderRequest{
			Name: "New Folder",
		}

		folder, err := mock.CreateFolder(ctx, "grant-123", req)
		require.NoError(t, err)
		assert.Equal(t, "New Folder", folder.Name)
		assert.True(t, mock.CreateFolderCalled)
	})

	t.Run("UpdateFolder", func(t *testing.T) {
		req := &domain.UpdateFolderRequest{
			Name: "Renamed Folder",
		}

		folder, err := mock.UpdateFolder(ctx, "grant-123", "folder-456", req)
		require.NoError(t, err)
		assert.Equal(t, "Renamed Folder", folder.Name)
		assert.True(t, mock.UpdateFolderCalled)
	})

	t.Run("DeleteFolder", func(t *testing.T) {
		err := mock.DeleteFolder(ctx, "grant-123", "folder-456")
		require.NoError(t, err)
		assert.True(t, mock.DeleteFolderCalled)
	})
}

func TestMockClient_Attachments(t *testing.T) {
	ctx := context.Background()
	mock := nylas.NewMockClient()

	t.Run("GetAttachment", func(t *testing.T) {
		attachment, err := mock.GetAttachment(ctx, "grant-123", "msg-789", "attach-456")
		require.NoError(t, err)
		assert.Equal(t, "attach-456", attachment.ID)
		assert.Equal(t, "test.pdf", attachment.Filename)
		assert.True(t, mock.GetAttachmentCalled)
	})

	t.Run("DownloadAttachment", func(t *testing.T) {
		reader, err := mock.DownloadAttachment(ctx, "grant-123", "msg-789", "attach-456")
		require.NoError(t, err)
		defer reader.Close()

		content, err := io.ReadAll(reader)
		require.NoError(t, err)
		assert.Equal(t, "mock attachment content", string(content))
		assert.True(t, mock.DownloadAttachmentCalled)
	})
}

func TestMockClient_Grants(t *testing.T) {
	ctx := context.Background()
	mock := nylas.NewMockClient()

	t.Run("ExchangeCode", func(t *testing.T) {
		grant, err := mock.ExchangeCode(ctx, "auth-code", "http://localhost")
		require.NoError(t, err)
		assert.Equal(t, "mock-grant-id", grant.ID)
		assert.True(t, mock.ExchangeCodeCalled)
	})

	t.Run("ListGrants", func(t *testing.T) {
		mock.ListGrantsFunc = func(ctx context.Context) ([]domain.Grant, error) {
			return []domain.Grant{
				{ID: "grant-1", Email: "user1@example.com"},
				{ID: "grant-2", Email: "user2@example.com"},
			}, nil
		}

		grants, err := mock.ListGrants(ctx)
		require.NoError(t, err)
		assert.Len(t, grants, 2)
		assert.True(t, mock.ListGrantsCalled)
	})

	t.Run("GetGrant", func(t *testing.T) {
		grant, err := mock.GetGrant(ctx, "grant-123")
		require.NoError(t, err)
		assert.Equal(t, "grant-123", grant.ID)
		assert.True(t, mock.GetGrantCalled)
	})

	t.Run("RevokeGrant", func(t *testing.T) {
		err := mock.RevokeGrant(ctx, "grant-123")
		require.NoError(t, err)
		assert.True(t, mock.RevokeGrantCalled)
	})
}

func TestDomainTypes(t *testing.T) {
	t.Run("Contact String method", func(t *testing.T) {
		tests := []struct {
			contact domain.EmailParticipant
			want    string
		}{
			{domain.EmailParticipant{Name: "John Doe", Email: "john@example.com"}, "John Doe <john@example.com>"},
			{domain.EmailParticipant{Name: "", Email: "john@example.com"}, "john@example.com"},
		}

		for _, tt := range tests {
			assert.Equal(t, tt.want, tt.contact.String())
		}
	})

	t.Run("Provider validation", func(t *testing.T) {
		assert.True(t, domain.ProviderGoogle.IsValid())
		assert.True(t, domain.ProviderMicrosoft.IsValid())
		assert.True(t, domain.ProviderIMAP.IsValid())
		assert.False(t, domain.Provider("invalid").IsValid())
	})

	t.Run("Provider display name", func(t *testing.T) {
		assert.Equal(t, "Google", domain.ProviderGoogle.DisplayName())
		assert.Equal(t, "Microsoft", domain.ProviderMicrosoft.DisplayName())
	})

	t.Run("ParseProvider", func(t *testing.T) {
		provider, err := domain.ParseProvider("google")
		require.NoError(t, err)
		assert.Equal(t, domain.ProviderGoogle, provider)

		_, err = domain.ParseProvider("invalid")
		assert.Error(t, err)
	})

	t.Run("Grant IsValid", func(t *testing.T) {
		validGrant := domain.Grant{GrantStatus: "valid"}
		invalidGrant := domain.Grant{GrantStatus: "invalid"}

		assert.True(t, validGrant.IsValid())
		assert.False(t, invalidGrant.IsValid())
	})
}

func TestMessageQueryParams(t *testing.T) {
	t.Run("creates params with defaults", func(t *testing.T) {
		params := &domain.MessageQueryParams{}
		assert.Equal(t, 0, params.Limit)
		assert.Nil(t, params.Unread)
	})

	t.Run("creates params with values", func(t *testing.T) {
		unread := true
		starred := false
		params := &domain.MessageQueryParams{
			Limit:         20,
			Unread:        &unread,
			Starred:       &starred,
			From:          "sender@example.com",
			SearchQuery:   "important",
			ReceivedAfter: time.Now().Unix(),
		}

		assert.Equal(t, 20, params.Limit)
		assert.True(t, *params.Unread)
		assert.False(t, *params.Starred)
		assert.Equal(t, "sender@example.com", params.From)
	})
}

func TestThreadQueryParams(t *testing.T) {
	t.Run("creates params with defaults", func(t *testing.T) {
		params := &domain.ThreadQueryParams{}
		assert.Equal(t, 0, params.Limit)
	})

	t.Run("creates params with values", func(t *testing.T) {
		unread := true
		params := &domain.ThreadQueryParams{
			Limit:       50,
			Unread:      &unread,
			Subject:     "test subject",
			SearchQuery: "keyword",
		}

		assert.Equal(t, 50, params.Limit)
		assert.True(t, *params.Unread)
	})
}

func TestSendMessageRequest(t *testing.T) {
	req := &domain.SendMessageRequest{
		Subject: "Test Email",
		Body:    "<html><body>Hello World</body></html>",
		To:      []domain.EmailParticipant{{Name: "Recipient", Email: "to@example.com"}},
		Cc:      []domain.EmailParticipant{{Email: "cc@example.com"}},
		Bcc:     []domain.EmailParticipant{{Email: "bcc@example.com"}},
		TrackingOpts: &domain.TrackingOptions{
			Opens: true,
			Links: true,
		},
	}

	assert.Equal(t, "Test Email", req.Subject)
	assert.Len(t, req.To, 1)
	assert.Len(t, req.Cc, 1)
	assert.Len(t, req.Bcc, 1)
	assert.True(t, req.TrackingOpts.Opens)
}

func TestCreateDraftRequest(t *testing.T) {
	req := &domain.CreateDraftRequest{
		Subject:      "Draft Subject",
		Body:         "Draft body content",
		To:           []domain.EmailParticipant{{Email: "to@example.com"}},
		ReplyToMsgID: "original-msg-id",
	}

	assert.Equal(t, "Draft Subject", req.Subject)
	assert.Equal(t, "original-msg-id", req.ReplyToMsgID)
}

func TestCreateFolderRequest(t *testing.T) {
	req := &domain.CreateFolderRequest{
		Name:            "My Folder",
		ParentID:        "parent-folder-id",
		BackgroundColor: "#FF0000",
		TextColor:       "#FFFFFF",
	}

	assert.Equal(t, "My Folder", req.Name)
	assert.Equal(t, "parent-folder-id", req.ParentID)
}

func TestUpdateMessageRequest(t *testing.T) {
	unread := false
	starred := true
	req := &domain.UpdateMessageRequest{
		Unread:  &unread,
		Starred: &starred,
		Folders: []string{"folder-1", "folder-2"},
	}

	assert.False(t, *req.Unread)
	assert.True(t, *req.Starred)
	assert.Len(t, req.Folders, 2)
}

func TestFolderSystemConstants(t *testing.T) {
	assert.Equal(t, "inbox", domain.FolderInbox)
	assert.Equal(t, "sent", domain.FolderSent)
	assert.Equal(t, "drafts", domain.FolderDrafts)
	assert.Equal(t, "trash", domain.FolderTrash)
	assert.Equal(t, "spam", domain.FolderSpam)
	assert.Equal(t, "archive", domain.FolderArchive)
	assert.Equal(t, "all", domain.FolderAll)
}

func TestUnixTimeUnmarshal(t *testing.T) {
	t.Run("unmarshals unix timestamp", func(t *testing.T) {
		jsonData := `{"created_at": 1703001600}`
		var result struct {
			CreatedAt domain.UnixTime `json:"created_at"`
		}
		err := json.Unmarshal([]byte(jsonData), &result)
		require.NoError(t, err)
		assert.Equal(t, int64(1703001600), result.CreatedAt.Unix())
	})

	t.Run("unmarshals RFC3339 string", func(t *testing.T) {
		jsonData := `{"created_at": "2023-12-19T12:00:00Z"}`
		var result struct {
			CreatedAt domain.UnixTime `json:"created_at"`
		}
		err := json.Unmarshal([]byte(jsonData), &result)
		require.NoError(t, err)
		assert.Equal(t, 2023, result.CreatedAt.Year())
	})
}

func TestAttachmentModel(t *testing.T) {
	attachment := domain.Attachment{
		ID:          "attach-123",
		GrantID:     "grant-456",
		Filename:    "document.pdf",
		ContentType: "application/pdf",
		Size:        1024000,
		ContentID:   "cid-789",
		IsInline:    false,
	}

	assert.Equal(t, "attach-123", attachment.ID)
	assert.Equal(t, "document.pdf", attachment.Filename)
	assert.Equal(t, int64(1024000), attachment.Size)
	assert.False(t, attachment.IsInline)
}

func TestThreadModel(t *testing.T) {
	now := time.Now()
	thread := domain.Thread{
		ID:                    "thread-123",
		GrantID:               "grant-456",
		Subject:               "Test Thread Subject",
		Snippet:               "This is a preview...",
		HasAttachments:        true,
		HasDrafts:             false,
		Starred:               true,
		Unread:                true,
		EarliestMessageDate:   now.Add(-24 * time.Hour),
		LatestMessageRecvDate: now,
		Participants: []domain.EmailParticipant{
			{Name: "Alice", Email: "alice@example.com"},
			{Name: "Bob", Email: "bob@example.com"},
		},
		MessageIDs: []string{"msg-1", "msg-2"},
		FolderIDs:  []string{"inbox"},
	}

	assert.Equal(t, "thread-123", thread.ID)
	assert.Equal(t, "Test Thread Subject", thread.Subject)
	assert.True(t, thread.HasAttachments)
	assert.Len(t, thread.Participants, 2)
	assert.Len(t, thread.MessageIDs, 2)
}

func TestDraftModel(t *testing.T) {
	draft := domain.Draft{
		ID:           "draft-123",
		GrantID:      "grant-456",
		Subject:      "Draft Email",
		Body:         "Draft content here",
		From:         []domain.EmailParticipant{{Email: "me@example.com"}},
		To:           []domain.EmailParticipant{{Email: "recipient@example.com"}},
		ReplyToMsgID: "original-msg-id",
		ThreadID:     "thread-789",
	}

	assert.Equal(t, "draft-123", draft.ID)
	assert.Equal(t, "Draft Email", draft.Subject)
	assert.Len(t, draft.To, 1)
	assert.Equal(t, "original-msg-id", draft.ReplyToMsgID)
}

func TestFolderModel(t *testing.T) {
	folder := domain.Folder{
		ID:              "folder-123",
		GrantID:         "grant-456",
		Name:            "Important",
		SystemFolder:    "",
		ParentID:        "parent-folder",
		BackgroundColor: "#FF0000",
		TextColor:       "#FFFFFF",
		TotalCount:      100,
		UnreadCount:     25,
		ChildIDs:        []string{"child-1", "child-2"},
		Attributes:      []string{"\\HasNoChildren"},
	}

	assert.Equal(t, "folder-123", folder.ID)
	assert.Equal(t, "Important", folder.Name)
	assert.Equal(t, 100, folder.TotalCount)
	assert.Equal(t, 25, folder.UnreadCount)
	assert.Len(t, folder.ChildIDs, 2)
}

func TestMessageModel(t *testing.T) {
	now := time.Now()
	message := domain.Message{
		ID:       "msg-123",
		GrantID:  "grant-456",
		ThreadID: "thread-789",
		Subject:  "Test Subject",
		From:     []domain.EmailParticipant{{Name: "Sender", Email: "sender@example.com"}},
		To:       []domain.EmailParticipant{{Email: "to@example.com"}},
		Cc:       []domain.EmailParticipant{{Email: "cc@example.com"}},
		Body:     "<html>Hello</html>",
		Snippet:  "Hello...",
		Date:     now,
		Unread:   true,
		Starred:  false,
		Folders:  []string{"INBOX"},
		Attachments: []domain.Attachment{
			{ID: "attach-1", Filename: "file.pdf"},
		},
		Headers: []domain.Header{
			{Name: "X-Custom", Value: "custom-value"},
		},
	}

	assert.Equal(t, "msg-123", message.ID)
	assert.Equal(t, "Test Subject", message.Subject)
	assert.Len(t, message.From, 1)
	assert.Len(t, message.Attachments, 1)
	assert.Len(t, message.Headers, 1)
	assert.True(t, message.Unread)
}

func TestPaginationModel(t *testing.T) {
	pagination := domain.Pagination{
		NextCursor: "cursor-abc123",
		HasMore:    true,
	}

	assert.Equal(t, "cursor-abc123", pagination.NextCursor)
	assert.True(t, pagination.HasMore)
}

func TestTrackingOptions(t *testing.T) {
	opts := domain.TrackingOptions{
		Opens: true,
		Links: true,
		Label: "campaign-2024",
	}

	assert.True(t, opts.Opens)
	assert.True(t, opts.Links)
	assert.Equal(t, "campaign-2024", opts.Label)
}
