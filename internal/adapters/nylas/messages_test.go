//go:build !integration
// +build !integration

package nylas_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mqasimca/nylas/internal/adapters/nylas"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPClient_GetMessages(t *testing.T) {
	tests := []struct {
		name           string
		grantID        string
		limit          int
		serverResponse map[string]interface{}
		statusCode     int
		wantErr        bool
		wantCount      int
	}{
		{
			name:    "returns messages successfully",
			grantID: "grant-123",
			limit:   10,
			serverResponse: map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"id":        "msg-1",
						"grant_id":  "grant-123",
						"thread_id": "thread-1",
						"subject":   "Test Subject",
						"from":      []map[string]string{{"name": "Alice", "email": "alice@example.com"}},
						"to":        []map[string]string{{"name": "Bob", "email": "bob@example.com"}},
						"body":      "Test body content",
						"snippet":   "Test body...",
						"date":      1704067200,
						"unread":    true,
						"starred":   false,
						"folders":   []string{"INBOX"},
					},
					{
						"id":        "msg-2",
						"grant_id":  "grant-123",
						"thread_id": "thread-2",
						"subject":   "Another Subject",
						"from":      []map[string]string{{"name": "Charlie", "email": "charlie@example.com"}},
						"to":        []map[string]string{{"name": "Bob", "email": "bob@example.com"}},
						"body":      "Another body",
						"date":      1704153600,
						"unread":    false,
					},
				},
			},
			statusCode: http.StatusOK,
			wantErr:    false,
			wantCount:  2,
		},
		{
			name:    "returns empty list when no messages",
			grantID: "grant-456",
			limit:   10,
			serverResponse: map[string]interface{}{
				"data": []interface{}{},
			},
			statusCode: http.StatusOK,
			wantErr:    false,
			wantCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Contains(t, r.URL.Path, "/v3/grants/"+tt.grantID+"/messages")
				assert.Contains(t, r.Header.Get("Authorization"), "Bearer")

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				_ = json.NewEncoder(w).Encode(tt.serverResponse)
			}))
			defer server.Close()

			client := nylas.NewHTTPClient()
			client.SetCredentials("client-id", "secret", "api-key")
			client.SetBaseURL(server.URL)

			ctx := context.Background()
			messages, err := client.GetMessages(ctx, tt.grantID, tt.limit)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, messages, tt.wantCount)
		})
	}
}

func TestHTTPClient_GetMessagesWithParams(t *testing.T) {
	tests := []struct {
		name         string
		params       *domain.MessageQueryParams
		wantQuery    map[string]string
		notWantQuery []string
	}{
		{
			name: "includes all filter params",
			params: &domain.MessageQueryParams{
				Limit:     25,
				Subject:   "important",
				From:      "sender@example.com",
				To:        "recipient@example.com",
				ThreadID:  "thread-123",
				PageToken: "next-page",
			},
			wantQuery: map[string]string{
				"limit":      "25",
				"subject":    "important",
				"from":       "sender@example.com",
				"to":         "recipient@example.com",
				"thread_id":  "thread-123",
				"page_token": "next-page",
			},
		},
		{
			name: "includes boolean filters",
			params: func() *domain.MessageQueryParams {
				unread := true
				starred := false
				hasAttachment := true
				return &domain.MessageQueryParams{
					Limit:         10,
					Unread:        &unread,
					Starred:       &starred,
					HasAttachment: &hasAttachment,
				}
			}(),
			wantQuery: map[string]string{
				"unread":         "true",
				"starred":        "false",
				"has_attachment": "true",
			},
		},
		{
			name: "includes date range params",
			params: &domain.MessageQueryParams{
				Limit:          10,
				ReceivedBefore: 1704153600,
				ReceivedAfter:  1704067200,
			},
			wantQuery: map[string]string{
				"received_before": "1704153600",
				"received_after":  "1704067200",
			},
		},
		{
			name: "includes search query",
			params: &domain.MessageQueryParams{
				Limit:       10,
				SearchQuery: "meeting notes",
			},
			wantQuery: map[string]string{
				"q": "meeting notes",
			},
		},
		{
			name: "includes folder filter",
			params: &domain.MessageQueryParams{
				Limit: 10,
				In:    []string{"INBOX", "SENT"},
			},
			wantQuery: map[string]string{
				"in": "INBOX",
			},
		},
		{
			name:   "uses default limit for nil params",
			params: nil,
			wantQuery: map[string]string{
				"limit": "10",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				for key, expectedValue := range tt.wantQuery {
					actualValue := r.URL.Query().Get(key)
					if actualValue == "" {
						values := r.URL.Query()[key]
						if len(values) > 0 {
							actualValue = values[0]
						}
					}
					assert.Equal(t, expectedValue, actualValue, "Query param %s mismatch", key)
				}

				for _, key := range tt.notWantQuery {
					assert.Empty(t, r.URL.Query().Get(key), "Query param %s should not be present", key)
				}

				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"data": []interface{}{},
				})
			}))
			defer server.Close()

			client := nylas.NewHTTPClient()
			client.SetCredentials("client-id", "secret", "api-key")
			client.SetBaseURL(server.URL)

			ctx := context.Background()
			_, _ = client.GetMessagesWithParams(ctx, "grant-123", tt.params)
		})
	}
}

func TestHTTPClient_GetMessagesWithCursor(t *testing.T) {
	t.Run("returns pagination info", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := map[string]interface{}{
				"data": []map[string]interface{}{
					{"id": "msg-1", "subject": "First", "date": 1704067200},
					{"id": "msg-2", "subject": "Second", "date": 1704153600},
				},
				"next_cursor": "eyJsYXN0X2lkIjoibXNnLTIifQ==",
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := nylas.NewHTTPClient()
		client.SetCredentials("client-id", "secret", "api-key")
		client.SetBaseURL(server.URL)

		ctx := context.Background()
		result, err := client.GetMessagesWithCursor(ctx, "grant-123", &domain.MessageQueryParams{Limit: 2})

		require.NoError(t, err)
		assert.Len(t, result.Data, 2)
		assert.Equal(t, "eyJsYXN0X2lkIjoibXNnLTIifQ==", result.Pagination.NextCursor)
		assert.True(t, result.Pagination.HasMore)
	})

	t.Run("handles last page without cursor", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := map[string]interface{}{
				"data": []map[string]interface{}{
					{"id": "msg-1", "subject": "Last", "date": 1704067200},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := nylas.NewHTTPClient()
		client.SetCredentials("client-id", "secret", "api-key")
		client.SetBaseURL(server.URL)

		ctx := context.Background()
		result, err := client.GetMessagesWithCursor(ctx, "grant-123", nil)

		require.NoError(t, err)
		assert.Empty(t, result.Pagination.NextCursor)
		assert.False(t, result.Pagination.HasMore)
	})
}

func TestHTTPClient_GetMessage(t *testing.T) {
	tests := []struct {
		name           string
		grantID        string
		messageID      string
		serverResponse map[string]interface{}
		statusCode     int
		wantErr        bool
		errContains    string
	}{
		{
			name:      "returns message successfully",
			grantID:   "grant-123",
			messageID: "msg-456",
			serverResponse: map[string]interface{}{
				"data": map[string]interface{}{
					"id":        "msg-456",
					"grant_id":  "grant-123",
					"thread_id": "thread-789",
					"subject":   "Test Email",
					"from":      []map[string]string{{"name": "Sender", "email": "sender@example.com"}},
					"to":        []map[string]string{{"name": "Receiver", "email": "receiver@example.com"}},
					"cc":        []map[string]string{{"name": "CC Person", "email": "cc@example.com"}},
					"bcc":       []map[string]string{{"name": "BCC Person", "email": "bcc@example.com"}},
					"reply_to":  []map[string]string{{"name": "Reply", "email": "reply@example.com"}},
					"body":      "<p>Email body content</p>",
					"snippet":   "Email body content",
					"date":      1704067200,
					"unread":    true,
					"starred":   true,
					"folders":   []string{"INBOX"},
					"attachments": []map[string]interface{}{
						{
							"id":           "attach-1",
							"filename":     "report.pdf",
							"content_type": "application/pdf",
							"size":         12345,
							"is_inline":    false,
						},
					},
					"metadata":   map[string]string{"custom_key": "custom_value"},
					"created_at": 1704067200,
				},
			},
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:      "returns error for not found",
			grantID:   "grant-123",
			messageID: "nonexistent",
			serverResponse: map[string]interface{}{
				"error": map[string]string{"message": "message not found"},
			},
			statusCode:  http.StatusNotFound,
			wantErr:     true,
			errContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				expectedPath := "/v3/grants/" + tt.grantID + "/messages/" + tt.messageID
				assert.Equal(t, expectedPath, r.URL.Path)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				_ = json.NewEncoder(w).Encode(tt.serverResponse)
			}))
			defer server.Close()

			client := nylas.NewHTTPClient()
			client.SetCredentials("client-id", "secret", "api-key")
			client.SetBaseURL(server.URL)

			ctx := context.Background()
			message, err := client.GetMessage(ctx, tt.grantID, tt.messageID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.messageID, message.ID)
			assert.Equal(t, tt.grantID, message.GrantID)
			assert.Equal(t, "Test Email", message.Subject)
			assert.Len(t, message.From, 1)
			assert.Equal(t, "Sender", message.From[0].Name)
			assert.Len(t, message.Attachments, 1)
			assert.True(t, message.Unread)
			assert.True(t, message.Starred)
		})
	}
}

func TestHTTPClient_UpdateMessage(t *testing.T) {
	tests := []struct {
		name       string
		request    *domain.UpdateMessageRequest
		wantFields map[string]interface{}
	}{
		{
			name: "marks as read",
			request: func() *domain.UpdateMessageRequest {
				unread := false
				return &domain.UpdateMessageRequest{Unread: &unread}
			}(),
			wantFields: map[string]interface{}{"unread": false},
		},
		{
			name: "marks as starred",
			request: func() *domain.UpdateMessageRequest {
				starred := true
				return &domain.UpdateMessageRequest{Starred: &starred}
			}(),
			wantFields: map[string]interface{}{"starred": true},
		},
		{
			name: "moves to folders",
			request: &domain.UpdateMessageRequest{
				Folders: []string{"Archive", "Important"},
			},
			wantFields: map[string]interface{}{"folders": []string{"Archive", "Important"}},
		},
		{
			name: "updates multiple fields",
			request: func() *domain.UpdateMessageRequest {
				unread := true
				starred := true
				return &domain.UpdateMessageRequest{
					Unread:  &unread,
					Starred: &starred,
					Folders: []string{"INBOX"},
				}
			}(),
			wantFields: map[string]interface{}{
				"unread":  true,
				"starred": true,
				"folders": []string{"INBOX"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "PUT", r.Method)
				assert.Equal(t, "/v3/grants/grant-123/messages/msg-456", r.URL.Path)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				var body map[string]interface{}
				_ = json.NewDecoder(r.Body).Decode(&body)

				for key := range tt.wantFields {
					assert.Contains(t, body, key, "Missing field: %s", key)
				}

				response := map[string]interface{}{
					"data": map[string]interface{}{
						"id":       "msg-456",
						"grant_id": "grant-123",
						"subject":  "Updated",
						"date":     1704067200,
					},
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			client := nylas.NewHTTPClient()
			client.SetCredentials("client-id", "secret", "api-key")
			client.SetBaseURL(server.URL)

			ctx := context.Background()
			message, err := client.UpdateMessage(ctx, "grant-123", "msg-456", tt.request)

			require.NoError(t, err)
			assert.Equal(t, "msg-456", message.ID)
		})
	}
}

func TestHTTPClient_DeleteMessage(t *testing.T) {
	tests := []struct {
		name       string
		grantID    string
		messageID  string
		statusCode int
		wantErr    bool
	}{
		{
			name:       "deletes successfully with 200",
			grantID:    "grant-123",
			messageID:  "msg-456",
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "deletes successfully with 204",
			grantID:    "grant-123",
			messageID:  "msg-789",
			statusCode: http.StatusNoContent,
			wantErr:    false,
		},
		{
			name:       "returns error for not found",
			grantID:    "grant-123",
			messageID:  "nonexistent",
			statusCode: http.StatusNotFound,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "DELETE", r.Method)
				expectedPath := "/v3/grants/" + tt.grantID + "/messages/" + tt.messageID
				assert.Equal(t, expectedPath, r.URL.Path)

				w.WriteHeader(tt.statusCode)
				if tt.statusCode >= 400 {
					_ = json.NewEncoder(w).Encode(map[string]interface{}{
						"error": map[string]string{"message": "not found"},
					})
				}
			}))
			defer server.Close()

			client := nylas.NewHTTPClient()
			client.SetCredentials("client-id", "secret", "api-key")
			client.SetBaseURL(server.URL)

			ctx := context.Background()
			err := client.DeleteMessage(ctx, tt.grantID, tt.messageID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHTTPClient_GetMessages_ErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		statusCode  int
		response    map[string]interface{}
		errContains string
	}{
		{
			name:       "handles 401 unauthorized",
			statusCode: http.StatusUnauthorized,
			response: map[string]interface{}{
				"error": map[string]string{"message": "Invalid API key"},
			},
			errContains: "Invalid API key",
		},
		{
			name:       "handles 403 forbidden",
			statusCode: http.StatusForbidden,
			response: map[string]interface{}{
				"error": map[string]string{"message": "Access denied"},
			},
			errContains: "Access denied",
		},
		{
			name:       "handles 429 rate limited",
			statusCode: http.StatusTooManyRequests,
			response: map[string]interface{}{
				"error": map[string]string{"message": "Rate limit exceeded"},
			},
			errContains: "Rate limit exceeded",
		},
		{
			name:       "handles 500 server error",
			statusCode: http.StatusInternalServerError,
			response: map[string]interface{}{
				"error": map[string]string{"message": "Internal server error"},
			},
			errContains: "Internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				_ = json.NewEncoder(w).Encode(tt.response)
			}))
			defer server.Close()

			client := nylas.NewHTTPClient()
			client.SetCredentials("client-id", "secret", "api-key")
			client.SetBaseURL(server.URL)

			ctx := context.Background()
			_, err := client.GetMessages(ctx, "grant-123", 10)

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errContains)
		})
	}
}

func TestHTTPClient_GetMessage_FullConversion(t *testing.T) {
	timestamp := time.Now().Unix()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"id":        "msg-full",
				"grant_id":  "grant-full",
				"thread_id": "thread-full",
				"subject":   "Complete Message",
				"from": []map[string]string{
					{"name": "Alice Smith", "email": "alice@example.com"},
				},
				"to": []map[string]string{
					{"name": "Bob Jones", "email": "bob@example.com"},
					{"name": "Carol White", "email": "carol@example.com"},
				},
				"cc": []map[string]string{
					{"name": "Dave Brown", "email": "dave@example.com"},
				},
				"bcc": []map[string]string{
					{"name": "Eve Black", "email": "eve@example.com"},
				},
				"reply_to": []map[string]string{
					{"name": "Reply Handler", "email": "reply@example.com"},
				},
				"body":    "<html><body><p>Full body content</p></body></html>",
				"snippet": "Full body content",
				"date":    timestamp,
				"unread":  true,
				"starred": false,
				"folders": []string{"INBOX", "Important"},
				"attachments": []map[string]interface{}{
					{
						"id":           "attach-1",
						"filename":     "document.pdf",
						"content_type": "application/pdf",
						"size":         50000,
						"content_id":   "",
						"is_inline":    false,
					},
					{
						"id":           "attach-2",
						"filename":     "image.png",
						"content_type": "image/png",
						"size":         25000,
						"content_id":   "cid:123",
						"is_inline":    true,
					},
				},
				"metadata":   map[string]string{"key1": "value1", "key2": "value2"},
				"created_at": timestamp,
				"object":     "message",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := nylas.NewHTTPClient()
	client.SetCredentials("client-id", "secret", "api-key")
	client.SetBaseURL(server.URL)

	ctx := context.Background()
	msg, err := client.GetMessage(ctx, "grant-full", "msg-full")

	require.NoError(t, err)

	// Verify all fields
	assert.Equal(t, "msg-full", msg.ID)
	assert.Equal(t, "grant-full", msg.GrantID)
	assert.Equal(t, "thread-full", msg.ThreadID)
	assert.Equal(t, "Complete Message", msg.Subject)

	// From
	require.Len(t, msg.From, 1)
	assert.Equal(t, "Alice Smith", msg.From[0].Name)
	assert.Equal(t, "alice@example.com", msg.From[0].Email)

	// To
	require.Len(t, msg.To, 2)
	assert.Equal(t, "Bob Jones", msg.To[0].Name)
	assert.Equal(t, "Carol White", msg.To[1].Name)

	// CC
	require.Len(t, msg.Cc, 1)
	assert.Equal(t, "Dave Brown", msg.Cc[0].Name)

	// BCC
	require.Len(t, msg.Bcc, 1)
	assert.Equal(t, "Eve Black", msg.Bcc[0].Name)

	// Reply-To
	require.Len(t, msg.ReplyTo, 1)
	assert.Equal(t, "Reply Handler", msg.ReplyTo[0].Name)

	// Body and snippet
	assert.Contains(t, msg.Body, "Full body content")
	assert.Equal(t, "Full body content", msg.Snippet)

	// Flags
	assert.True(t, msg.Unread)
	assert.False(t, msg.Starred)

	// Folders
	assert.Contains(t, msg.Folders, "INBOX")
	assert.Contains(t, msg.Folders, "Important")

	// Attachments
	require.Len(t, msg.Attachments, 2)
	assert.Equal(t, "document.pdf", msg.Attachments[0].Filename)
	assert.Equal(t, "application/pdf", msg.Attachments[0].ContentType)
	assert.False(t, msg.Attachments[0].IsInline)
	assert.Equal(t, "image.png", msg.Attachments[1].Filename)
	assert.True(t, msg.Attachments[1].IsInline)
	assert.Equal(t, "cid:123", msg.Attachments[1].ContentID)

	// Metadata
	assert.Equal(t, "value1", msg.Metadata["key1"])
	assert.Equal(t, "value2", msg.Metadata["key2"])

	// Object type
	assert.Equal(t, "message", msg.Object)

	// Timestamps
	assert.Equal(t, time.Unix(timestamp, 0), msg.Date)
	assert.Equal(t, time.Unix(timestamp, 0), msg.CreatedAt)
}
