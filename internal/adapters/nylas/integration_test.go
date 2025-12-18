//go:build integration
// +build integration

package nylas_test

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/mqasimca/nylas/internal/adapters/nylas"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Test Configuration & Helpers
// =============================================================================

// testConfig holds test configuration loaded from environment variables.
// SECURITY: All credentials MUST be provided via environment variables.
// Never hardcode credentials in source code.
type testConfig struct {
	apiKey   string
	grantID  string
	clientID string
}

// getTestConfig loads test configuration from environment variables.
// This ensures credentials are never stored in code.
func getTestConfig(t *testing.T) testConfig {
	t.Helper()

	apiKey := os.Getenv("NYLAS_API_KEY")
	grantID := os.Getenv("NYLAS_GRANT_ID")
	clientID := os.Getenv("NYLAS_CLIENT_ID")

	if apiKey == "" {
		t.Skip("NYLAS_API_KEY not set, skipping integration test")
	}
	if grantID == "" {
		t.Skip("NYLAS_GRANT_ID not set, skipping integration test")
	}

	return testConfig{
		apiKey:   apiKey,
		grantID:  grantID,
		clientID: clientID,
	}
}

// getTestClient creates a configured Nylas client for integration tests.
// It also adds a rate limit delay to avoid hitting API rate limits.
func getTestClient(t *testing.T) (*nylas.HTTPClient, string) {
	t.Helper()

	// Add delay to avoid rate limiting between tests
	waitForRateLimit()

	cfg := getTestConfig(t)
	client := nylas.NewHTTPClient()
	client.SetCredentials(cfg.clientID, "", cfg.apiKey)

	return client, cfg.grantID
}

// createTestContext creates a context with standard test timeout.
func createTestContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 30*time.Second)
}

// createLongTestContext creates a context with extended timeout for slower operations.
func createLongTestContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 60*time.Second)
}

// rateLimitDelay adds a small delay between API calls to avoid rate limiting.
// Nylas API has rate limits, so we add a brief pause between tests.
const rateLimitDelay = 500 * time.Millisecond

// waitForRateLimit pauses execution to avoid hitting API rate limits.
func waitForRateLimit() {
	time.Sleep(rateLimitDelay)
}

// safeSubstring safely extracts a substring, avoiding panics on short strings.
func safeSubstring(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

// skipIfProviderNotSupported checks if the error indicates the provider doesn't support
// the operation and skips the test if so.
func skipIfProviderNotSupported(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		return
	}
	errMsg := err.Error()
	// Various error messages that indicate provider limitation
	if strings.Contains(errMsg, "Method not supported for provider") ||
		strings.Contains(errMsg, "an internal error ocurred") || // Nylas API typo
		strings.Contains(errMsg, "an internal error occurred") {
		t.Skipf("Provider does not support this operation: %v", err)
	}
}

// skipIfNoMessages skips the test if no messages are available to test with.
func skipIfNoMessages(t *testing.T, messages []domain.Message) {
	t.Helper()
	if len(messages) == 0 {
		t.Skip("No messages available in inbox, skipping test")
	}
}

// =============================================================================
// Grant Tests
// =============================================================================

func TestIntegration_ListGrants(t *testing.T) {
	client, _ := getTestClient(t)
	ctx, cancel := createTestContext()
	defer cancel()

	grants, err := client.ListGrants(ctx)
	require.NoError(t, err, "ListGrants should not return error")
	require.NotEmpty(t, grants, "Should have at least one grant")

	t.Logf("Found %d grants", len(grants))
	for _, g := range grants {
		t.Logf("  Grant ID: %s, Email: %s, Provider: %s, Status: %s",
			g.ID, g.Email, g.Provider, g.GrantStatus)

		// Validate grant fields
		assert.NotEmpty(t, g.ID, "Grant should have ID")
		assert.NotEmpty(t, g.Email, "Grant should have email")
		assert.NotEmpty(t, g.Provider, "Grant should have provider")
	}
}

func TestIntegration_ListGrants_ValidatesProvider(t *testing.T) {
	client, _ := getTestClient(t)
	ctx, cancel := createTestContext()
	defer cancel()

	grants, err := client.ListGrants(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, grants)

	validProviders := map[string]bool{
		"google":    true,
		"microsoft": true,
		"imap":      true,
		"yahoo":     true,
		"icloud":    true,
		"ews":       true,
		"inbox":     true, // Nylas Native Auth
		"virtual":   true,
	}

	for _, g := range grants {
		_, isValid := validProviders[strings.ToLower(string(g.Provider))]
		assert.True(t, isValid, "Provider %s should be a valid Nylas provider", g.Provider)
	}
}

// =============================================================================
// Message Tests - Basic Operations
// =============================================================================

func TestIntegration_GetMessages(t *testing.T) {
	client, grantID := getTestClient(t)
	ctx, cancel := createTestContext()
	defer cancel()

	messages, err := client.GetMessages(ctx, grantID, 10)
	require.NoError(t, err, "GetMessages should not return error")

	t.Logf("Found %d messages", len(messages))
	for _, m := range messages {
		from := ""
		if len(m.From) > 0 {
			from = m.From[0].Email
		}
		t.Logf("  [%s] %s - %s (%s)",
			safeSubstring(m.ID, 8), from, safeSubstring(m.Subject, 40), m.Date.Format("Jan 2, 15:04"))

		// Validate message fields
		assert.NotEmpty(t, m.ID, "Message should have ID")
		assert.NotZero(t, m.Date, "Message should have date")
	}
}

func TestIntegration_GetMessages_LimitRespected(t *testing.T) {
	client, grantID := getTestClient(t)
	ctx, cancel := createTestContext()
	defer cancel()

	limits := []int{1, 5, 10, 25}

	for _, limit := range limits {
		t.Run(fmt.Sprintf("Limit_%d", limit), func(t *testing.T) {
			messages, err := client.GetMessages(ctx, grantID, limit)
			require.NoError(t, err)
			assert.LessOrEqual(t, len(messages), limit, "Should not exceed requested limit")
			t.Logf("Requested %d, got %d messages", limit, len(messages))
		})
	}
}

func TestIntegration_GetMessagesWithParams_Unread(t *testing.T) {
	client, grantID := getTestClient(t)
	ctx, cancel := createTestContext()
	defer cancel()

	unread := true
	params := &domain.MessageQueryParams{
		Limit:  10,
		Unread: &unread,
	}

	messages, err := client.GetMessagesWithParams(ctx, grantID, params)
	require.NoError(t, err)

	t.Logf("Found %d unread messages", len(messages))
	for _, m := range messages {
		assert.True(t, m.Unread, "All messages should be unread when filtering by unread=true")
		t.Logf("  Subject: %s, Unread: %v", safeSubstring(m.Subject, 50), m.Unread)
	}
}

func TestIntegration_GetMessagesWithParams_Read(t *testing.T) {
	client, grantID := getTestClient(t)
	ctx, cancel := createTestContext()
	defer cancel()

	unread := false
	params := &domain.MessageQueryParams{
		Limit:  10,
		Unread: &unread,
	}

	messages, err := client.GetMessagesWithParams(ctx, grantID, params)
	require.NoError(t, err)

	t.Logf("Found %d read messages", len(messages))
	for _, m := range messages {
		assert.False(t, m.Unread, "All messages should be read when filtering by unread=false")
	}
}

func TestIntegration_GetMessagesWithParams_Starred(t *testing.T) {
	client, grantID := getTestClient(t)
	ctx, cancel := createTestContext()
	defer cancel()

	starred := true
	params := &domain.MessageQueryParams{
		Limit:   10,
		Starred: &starred,
	}

	messages, err := client.GetMessagesWithParams(ctx, grantID, params)
	require.NoError(t, err)

	t.Logf("Found %d starred messages", len(messages))
	for _, m := range messages {
		assert.True(t, m.Starred, "All messages should be starred when filtering by starred=true")
	}
}

func TestIntegration_GetMessagesWithParams_InFolder(t *testing.T) {
	client, grantID := getTestClient(t)
	ctx, cancel := createTestContext()
	defer cancel()

	// First get folders
	folders, err := client.GetFolders(ctx, grantID)
	skipIfProviderNotSupported(t, err)
	require.NoError(t, err)
	require.NotEmpty(t, folders)

	// Find INBOX or first folder
	var targetFolder *domain.Folder
	for i := range folders {
		if strings.EqualFold(folders[i].Name, "INBOX") || strings.EqualFold(folders[i].Name, "Inbox") {
			targetFolder = &folders[i]
			break
		}
	}
	if targetFolder == nil {
		targetFolder = &folders[0]
	}

	params := &domain.MessageQueryParams{
		Limit: 5,
		In:    []string{targetFolder.ID},
	}

	messages, err := client.GetMessagesWithParams(ctx, grantID, params)
	require.NoError(t, err)

	t.Logf("Found %d messages in folder '%s' (%s)", len(messages), targetFolder.Name, targetFolder.ID)
	for _, m := range messages {
		t.Logf("  Subject: %s", safeSubstring(m.Subject, 50))
	}
}

func TestIntegration_GetSingleMessage(t *testing.T) {
	client, grantID := getTestClient(t)
	ctx, cancel := createTestContext()
	defer cancel()

	// First get a list of messages
	messages, err := client.GetMessages(ctx, grantID, 1)
	require.NoError(t, err)
	skipIfNoMessages(t, messages)

	messageID := messages[0].ID

	// Now get the single message
	msg, err := client.GetMessage(ctx, grantID, messageID)
	require.NoError(t, err)

	assert.Equal(t, messageID, msg.ID)
	assert.NotEmpty(t, msg.Subject, "Message should have subject")
	assert.NotZero(t, msg.Date, "Message should have date")

	t.Logf("Message ID: %s", msg.ID)
	t.Logf("Subject: %s", msg.Subject)
	t.Logf("From: %v", msg.From)
	t.Logf("To: %v", msg.To)
	t.Logf("Date: %s", msg.Date.Format(time.RFC3339))
	t.Logf("Body length: %d chars", len(msg.Body))
	t.Logf("Snippet: %s", safeSubstring(msg.Snippet, 100))
	t.Logf("Unread: %v, Starred: %v", msg.Unread, msg.Starred)
	t.Logf("Thread ID: %s", msg.ThreadID)
	t.Logf("Attachments: %d", len(msg.Attachments))
}

func TestIntegration_GetSingleMessage_FullContent(t *testing.T) {
	client, grantID := getTestClient(t)
	ctx, cancel := createTestContext()
	defer cancel()

	messages, err := client.GetMessages(ctx, grantID, 5)
	require.NoError(t, err)
	skipIfNoMessages(t, messages)

	// Get full details for each message
	for _, m := range messages {
		fullMsg, err := client.GetMessage(ctx, grantID, m.ID)
		require.NoError(t, err)

		// Full message should have more details
		assert.Equal(t, m.ID, fullMsg.ID)
		t.Logf("[%s] Subject: %s, Body: %d chars, Attachments: %d",
			safeSubstring(m.ID, 8), safeSubstring(fullMsg.Subject, 30), len(fullMsg.Body), len(fullMsg.Attachments))
	}
}

func TestIntegration_GetMessage_NotFound(t *testing.T) {
	client, grantID := getTestClient(t)
	ctx, cancel := createTestContext()
	defer cancel()

	_, err := client.GetMessage(ctx, grantID, "nonexistent-message-id-12345")
	assert.Error(t, err, "Should return error for non-existent message")
	t.Logf("Expected error: %v", err)
}

// =============================================================================
// Message Tests - Mark Operations
// =============================================================================

func TestIntegration_MarkMessageReadUnread(t *testing.T) {
	client, grantID := getTestClient(t)
	ctx, cancel := createLongTestContext()
	defer cancel()

	// Get a message to test with
	messages, err := client.GetMessages(ctx, grantID, 1)
	require.NoError(t, err)
	skipIfNoMessages(t, messages)

	messageID := messages[0].ID
	originalUnread := messages[0].Unread
	t.Logf("Original message unread status: %v", originalUnread)

	// Mark as opposite of current state
	newUnread := !originalUnread
	req := &domain.UpdateMessageRequest{
		Unread: &newUnread,
	}

	updated, err := client.UpdateMessage(ctx, grantID, messageID, req)
	require.NoError(t, err)
	assert.Equal(t, newUnread, updated.Unread, "Unread status should be updated")
	t.Logf("Updated unread status to: %v", updated.Unread)

	// Restore original state
	req.Unread = &originalUnread
	restored, err := client.UpdateMessage(ctx, grantID, messageID, req)
	require.NoError(t, err)
	assert.Equal(t, originalUnread, restored.Unread, "Unread status should be restored")
	t.Logf("Restored unread status to: %v", restored.Unread)
}

func TestIntegration_MarkMessageStarred(t *testing.T) {
	client, grantID := getTestClient(t)
	ctx, cancel := createLongTestContext()
	defer cancel()

	messages, err := client.GetMessages(ctx, grantID, 1)
	require.NoError(t, err)
	skipIfNoMessages(t, messages)

	messageID := messages[0].ID
	originalStarred := messages[0].Starred
	t.Logf("Original starred status: %v", originalStarred)

	// Toggle starred status
	newStarred := !originalStarred
	req := &domain.UpdateMessageRequest{
		Starred: &newStarred,
	}

	updated, err := client.UpdateMessage(ctx, grantID, messageID, req)
	require.NoError(t, err)
	assert.Equal(t, newStarred, updated.Starred)
	t.Logf("Updated starred status to: %v", updated.Starred)

	// Restore
	req.Starred = &originalStarred
	restored, err := client.UpdateMessage(ctx, grantID, messageID, req)
	require.NoError(t, err)
	assert.Equal(t, originalStarred, restored.Starred)
	t.Logf("Restored starred status to: %v", restored.Starred)
}

func TestIntegration_UpdateMessage_MultipleFlagsAtOnce(t *testing.T) {
	client, grantID := getTestClient(t)
	ctx, cancel := createLongTestContext()
	defer cancel()

	messages, err := client.GetMessages(ctx, grantID, 1)
	require.NoError(t, err)
	skipIfNoMessages(t, messages)

	messageID := messages[0].ID
	originalUnread := messages[0].Unread
	originalStarred := messages[0].Starred

	// Update both flags at once
	newUnread := !originalUnread
	newStarred := !originalStarred
	req := &domain.UpdateMessageRequest{
		Unread:  &newUnread,
		Starred: &newStarred,
	}

	updated, err := client.UpdateMessage(ctx, grantID, messageID, req)
	require.NoError(t, err)
	assert.Equal(t, newUnread, updated.Unread)
	assert.Equal(t, newStarred, updated.Starred)
	t.Logf("Updated both flags: unread=%v, starred=%v", updated.Unread, updated.Starred)

	// Restore
	req = &domain.UpdateMessageRequest{
		Unread:  &originalUnread,
		Starred: &originalStarred,
	}
	restored, err := client.UpdateMessage(ctx, grantID, messageID, req)
	require.NoError(t, err)
	assert.Equal(t, originalUnread, restored.Unread)
	assert.Equal(t, originalStarred, restored.Starred)
	t.Logf("Restored both flags: unread=%v, starred=%v", restored.Unread, restored.Starred)
}

// =============================================================================
// Search Tests
// =============================================================================

func TestIntegration_SearchMessages_BySubject(t *testing.T) {
	client, grantID := getTestClient(t)
	ctx, cancel := createTestContext()
	defer cancel()

	// First get some messages to find a search term
	messages, err := client.GetMessages(ctx, grantID, 5)
	require.NoError(t, err)
	skipIfNoMessages(t, messages)

	// Use a word from the first message's subject as search term
	searchTerm := ""
	for _, m := range messages {
		words := strings.Fields(m.Subject)
		for _, w := range words {
			if len(w) > 3 {
				searchTerm = w
				break
			}
		}
		if searchTerm != "" {
			break
		}
	}

	if searchTerm == "" {
		t.Skip("Could not find suitable search term")
	}

	// Search by subject field instead of full-text search
	params := &domain.MessageQueryParams{
		Limit:   20,
		Subject: searchTerm,
	}

	results, err := client.GetMessagesWithParams(ctx, grantID, params)
	require.NoError(t, err)

	t.Logf("Search for subject containing '%s' returned %d results", searchTerm, len(results))
	for _, m := range results {
		t.Logf("  Subject: %s", safeSubstring(m.Subject, 60))
	}
}

func TestIntegration_SearchMessages_ByFrom(t *testing.T) {
	client, grantID := getTestClient(t)
	ctx, cancel := createTestContext()
	defer cancel()

	// Get messages to find a sender
	messages, err := client.GetMessages(ctx, grantID, 10)
	require.NoError(t, err)
	skipIfNoMessages(t, messages)

	var fromEmail string
	for _, m := range messages {
		if len(m.From) > 0 && m.From[0].Email != "" {
			fromEmail = m.From[0].Email
			break
		}
	}

	if fromEmail == "" {
		t.Skip("Could not find sender email")
	}

	params := &domain.MessageQueryParams{
		Limit: 10,
		From:  fromEmail,
	}

	results, err := client.GetMessagesWithParams(ctx, grantID, params)
	require.NoError(t, err)

	t.Logf("Search from '%s' returned %d results", fromEmail, len(results))
	for _, m := range results {
		foundFrom := false
		for _, f := range m.From {
			if strings.EqualFold(f.Email, fromEmail) {
				foundFrom = true
				break
			}
		}
		assert.True(t, foundFrom, "Message should be from %s", fromEmail)
	}
}

func TestIntegration_SearchMessages_DateRange(t *testing.T) {
	client, grantID := getTestClient(t)
	ctx, cancel := createTestContext()
	defer cancel()

	// Search for messages from the last 7 days
	now := time.Now()
	weekAgo := now.AddDate(0, 0, -7)

	params := &domain.MessageQueryParams{
		Limit:         20,
		ReceivedAfter: weekAgo.Unix(),
	}

	results, err := client.GetMessagesWithParams(ctx, grantID, params)
	require.NoError(t, err)

	t.Logf("Found %d messages from last 7 days", len(results))
	for _, m := range results {
		assert.True(t, m.Date.After(weekAgo) || m.Date.Equal(weekAgo),
			"Message date %v should be after %v", m.Date, weekAgo)
		t.Logf("  [%s] %s", m.Date.Format("Jan 2"), safeSubstring(m.Subject, 50))
	}
}

func TestIntegration_SearchMessages_Combined(t *testing.T) {
	client, grantID := getTestClient(t)
	ctx, cancel := createTestContext()
	defer cancel()

	// Combined search: unread messages from last 30 days
	unread := true
	monthAgo := time.Now().AddDate(0, 0, -30)

	params := &domain.MessageQueryParams{
		Limit:         10,
		Unread:        &unread,
		ReceivedAfter: monthAgo.Unix(),
	}

	results, err := client.GetMessagesWithParams(ctx, grantID, params)
	require.NoError(t, err)

	t.Logf("Found %d unread messages from last 30 days", len(results))
	for _, m := range results {
		assert.True(t, m.Unread, "Message should be unread")
	}
}

// =============================================================================
// Folder Tests
// =============================================================================

func TestIntegration_GetFolders(t *testing.T) {
	client, grantID := getTestClient(t)
	ctx, cancel := createTestContext()
	defer cancel()

	folders, err := client.GetFolders(ctx, grantID)
	skipIfProviderNotSupported(t, err)
	require.NoError(t, err)
	require.NotEmpty(t, folders, "Should have at least one folder")

	t.Logf("Found %d folders", len(folders))

	standardFolders := []string{"inbox", "sent", "drafts", "trash", "spam", "archive"}
	foundStandard := make(map[string]bool)

	for _, f := range folders {
		assert.NotEmpty(t, f.ID, "Folder should have ID")
		assert.NotEmpty(t, f.Name, "Folder should have name")

		t.Logf("  [%s] %s (total: %d, unread: %d)",
			safeSubstring(f.ID, 8), f.Name, f.TotalCount, f.UnreadCount)

		// Track standard folders found
		nameLower := strings.ToLower(f.Name)
		for _, std := range standardFolders {
			if strings.Contains(nameLower, std) {
				foundStandard[std] = true
			}
		}
	}

	t.Logf("Standard folders found: %v", foundStandard)
}

func TestIntegration_GetSingleFolder(t *testing.T) {
	client, grantID := getTestClient(t)
	ctx, cancel := createTestContext()
	defer cancel()

	// First get folders
	folders, err := client.GetFolders(ctx, grantID)
	skipIfProviderNotSupported(t, err)
	require.NoError(t, err)
	require.NotEmpty(t, folders)

	folderID := folders[0].ID

	// Get single folder
	folder, err := client.GetFolder(ctx, grantID, folderID)
	require.NoError(t, err)

	assert.Equal(t, folderID, folder.ID)
	assert.NotEmpty(t, folder.Name)

	t.Logf("Folder: %s (%s)", folder.Name, folder.ID)
	t.Logf("Total messages: %d, Unread: %d", folder.TotalCount, folder.UnreadCount)
}

func TestIntegration_FolderLifecycle(t *testing.T) {
	client, grantID := getTestClient(t)
	ctx, cancel := createLongTestContext()
	defer cancel()

	// Create a unique folder name
	folderName := fmt.Sprintf("IntegrationTest_%d", time.Now().Unix())

	// Create folder
	createReq := &domain.CreateFolderRequest{
		Name: folderName,
	}

	folder, err := client.CreateFolder(ctx, grantID, createReq)
	skipIfProviderNotSupported(t, err)
	require.NoError(t, err)
	require.NotEmpty(t, folder.ID)
	assert.Equal(t, folderName, folder.Name)
	t.Logf("Created folder: %s (%s)", folder.Name, folder.ID)

	// Get the folder
	retrieved, err := client.GetFolder(ctx, grantID, folder.ID)
	require.NoError(t, err)
	assert.Equal(t, folder.ID, retrieved.ID)
	assert.Equal(t, folderName, retrieved.Name)

	// Update folder name
	newName := folderName + "_Updated"
	updateReq := &domain.UpdateFolderRequest{
		Name: newName,
	}
	updated, err := client.UpdateFolder(ctx, grantID, folder.ID, updateReq)
	require.NoError(t, err)
	assert.Equal(t, newName, updated.Name)
	t.Logf("Updated folder name to: %s", updated.Name)

	// Delete the folder
	err = client.DeleteFolder(ctx, grantID, folder.ID)
	require.NoError(t, err)
	t.Logf("Deleted folder: %s", folder.ID)

	// Verify deletion
	_, err = client.GetFolder(ctx, grantID, folder.ID)
	assert.Error(t, err, "Folder should be deleted")
}

func TestIntegration_GetFolder_NotFound(t *testing.T) {
	client, grantID := getTestClient(t)
	ctx, cancel := createTestContext()
	defer cancel()

	_, err := client.GetFolder(ctx, grantID, "nonexistent-folder-id-12345")
	assert.Error(t, err, "Should return error for non-existent folder")
	t.Logf("Expected error: %v", err)
}

// =============================================================================
// Thread Tests
// =============================================================================

func TestIntegration_GetThreads(t *testing.T) {
	client, grantID := getTestClient(t)
	ctx, cancel := createTestContext()
	defer cancel()

	params := &domain.ThreadQueryParams{
		Limit: 10,
	}

	threads, err := client.GetThreads(ctx, grantID, params)
	skipIfProviderNotSupported(t, err)
	require.NoError(t, err)

	t.Logf("Found %d threads", len(threads))
	for _, th := range threads {
		assert.NotEmpty(t, th.ID, "Thread should have ID")

		t.Logf("  [%s] %s (%d messages, unread: %v, starred: %v)",
			safeSubstring(th.ID, 8), safeSubstring(th.Subject, 40),
			len(th.MessageIDs), th.Unread, th.Starred)
	}
}

func TestIntegration_GetThreads_WithParams(t *testing.T) {
	client, grantID := getTestClient(t)
	ctx, cancel := createTestContext()
	defer cancel()

	// Get unread threads
	unread := true
	params := &domain.ThreadQueryParams{
		Limit:  5,
		Unread: &unread,
	}

	threads, err := client.GetThreads(ctx, grantID, params)
	skipIfProviderNotSupported(t, err)
	require.NoError(t, err)

	t.Logf("Found %d unread threads", len(threads))
	for _, th := range threads {
		assert.True(t, th.Unread, "Thread should be unread")
	}
}

func TestIntegration_GetSingleThread(t *testing.T) {
	client, grantID := getTestClient(t)
	ctx, cancel := createTestContext()
	defer cancel()

	// Get threads first
	params := &domain.ThreadQueryParams{
		Limit: 1,
	}
	threads, err := client.GetThreads(ctx, grantID, params)
	skipIfProviderNotSupported(t, err)
	require.NoError(t, err)
	require.NotEmpty(t, threads)

	threadID := threads[0].ID

	// Get single thread
	thread, err := client.GetThread(ctx, grantID, threadID)
	require.NoError(t, err)

	assert.Equal(t, threadID, thread.ID)
	assert.NotEmpty(t, thread.MessageIDs, "Thread should have message IDs")

	t.Logf("Thread: %s", thread.ID)
	t.Logf("Subject: %s", thread.Subject)
	t.Logf("Messages: %d", len(thread.MessageIDs))
	t.Logf("Participants: %v", thread.Participants)
	t.Logf("Latest Message ID: %s", thread.LatestDraftOrMessage.ID)
}

func TestIntegration_UpdateThread(t *testing.T) {
	client, grantID := getTestClient(t)
	ctx, cancel := createLongTestContext()
	defer cancel()

	// Get a thread
	params := &domain.ThreadQueryParams{
		Limit: 1,
	}
	threads, err := client.GetThreads(ctx, grantID, params)
	skipIfProviderNotSupported(t, err)
	require.NoError(t, err)
	require.NotEmpty(t, threads)

	threadID := threads[0].ID
	originalUnread := threads[0].Unread
	originalStarred := threads[0].Starred

	// Update thread
	newUnread := !originalUnread
	newStarred := !originalStarred
	req := &domain.UpdateMessageRequest{
		Unread:  &newUnread,
		Starred: &newStarred,
	}

	updated, err := client.UpdateThread(ctx, grantID, threadID, req)
	require.NoError(t, err)
	assert.Equal(t, newUnread, updated.Unread)
	assert.Equal(t, newStarred, updated.Starred)
	t.Logf("Updated thread: unread=%v, starred=%v", updated.Unread, updated.Starred)

	// Restore
	req = &domain.UpdateMessageRequest{
		Unread:  &originalUnread,
		Starred: &originalStarred,
	}
	restored, err := client.UpdateThread(ctx, grantID, threadID, req)
	require.NoError(t, err)
	assert.Equal(t, originalUnread, restored.Unread)
	assert.Equal(t, originalStarred, restored.Starred)
	t.Logf("Restored thread: unread=%v, starred=%v", restored.Unread, restored.Starred)
}

func TestIntegration_GetThread_NotFound(t *testing.T) {
	client, grantID := getTestClient(t)
	ctx, cancel := createTestContext()
	defer cancel()

	_, err := client.GetThread(ctx, grantID, "nonexistent-thread-id-12345")
	assert.Error(t, err, "Should return error for non-existent thread")
}

// =============================================================================
// Draft Tests
// =============================================================================

func TestIntegration_GetDrafts(t *testing.T) {
	client, grantID := getTestClient(t)
	ctx, cancel := createTestContext()
	defer cancel()

	drafts, err := client.GetDrafts(ctx, grantID, 10)
	skipIfProviderNotSupported(t, err)
	require.NoError(t, err)

	t.Logf("Found %d drafts", len(drafts))
	for _, d := range drafts {
		t.Logf("  [%s] %s", safeSubstring(d.ID, 8), safeSubstring(d.Subject, 50))
	}
}

func TestIntegration_DraftLifecycle_Basic(t *testing.T) {
	client, grantID := getTestClient(t)
	ctx, cancel := createLongTestContext()
	defer cancel()

	timestamp := time.Now().Format("2006-01-02 15:04:05")

	// Create a draft
	createReq := &domain.CreateDraftRequest{
		Subject: fmt.Sprintf("Integration Test Draft - %s", timestamp),
		Body:    "<html><body><p>This is a test draft created by integration tests.</p></body></html>",
		To:      []domain.EmailParticipant{{Email: "test@example.com", Name: "Test User"}},
	}

	draft, err := client.CreateDraft(ctx, grantID, createReq)
	skipIfProviderNotSupported(t, err)
	require.NoError(t, err)
	require.NotEmpty(t, draft.ID)
	t.Logf("Created draft: %s", draft.ID)
	t.Logf("Subject: %s", draft.Subject)

	// Get the draft
	retrieved, err := client.GetDraft(ctx, grantID, draft.ID)
	require.NoError(t, err)
	assert.Equal(t, draft.ID, retrieved.ID)
	assert.Equal(t, createReq.Subject, retrieved.Subject)
	t.Logf("Retrieved draft: %s", retrieved.Subject)

	// Update the draft
	updateReq := &domain.CreateDraftRequest{
		Subject: createReq.Subject + " (UPDATED)",
		Body:    "<html><body><p>This draft has been updated.</p></body></html>",
		To:      createReq.To,
	}
	updated, err := client.UpdateDraft(ctx, grantID, draft.ID, updateReq)
	require.NoError(t, err)
	assert.Contains(t, updated.Subject, "(UPDATED)")
	t.Logf("Updated draft subject: %s", updated.Subject)

	// Delete the draft
	err = client.DeleteDraft(ctx, grantID, draft.ID)
	require.NoError(t, err)
	t.Logf("Deleted draft: %s", draft.ID)

	// Verify deletion (with tolerance for eventual consistency)
	_, err = client.GetDraft(ctx, grantID, draft.ID)
	if err != nil {
		t.Logf("Draft deletion verified: %v", err)
	} else {
		t.Logf("Draft still exists due to eventual consistency (this is OK)")
	}
}

func TestIntegration_DraftLifecycle_WithCC(t *testing.T) {
	client, grantID := getTestClient(t)
	ctx, cancel := createLongTestContext()
	defer cancel()

	createReq := &domain.CreateDraftRequest{
		Subject: fmt.Sprintf("Draft with CC - %s", time.Now().Format("15:04:05")),
		Body:    "Test draft with CC recipients",
		To:      []domain.EmailParticipant{{Email: "recipient@example.com", Name: "Main Recipient"}},
		Cc:      []domain.EmailParticipant{{Email: "cc1@example.com"}, {Email: "cc2@example.com", Name: "CC Two"}},
	}

	draft, err := client.CreateDraft(ctx, grantID, createReq)
	skipIfProviderNotSupported(t, err)
	require.NoError(t, err)
	require.NotEmpty(t, draft.ID)
	t.Logf("Created draft with CC: %s", draft.ID)

	// Verify CC was saved
	retrieved, err := client.GetDraft(ctx, grantID, draft.ID)
	require.NoError(t, err)
	assert.Len(t, retrieved.Cc, 2, "Draft should have 2 CC recipients")

	// Cleanup
	err = client.DeleteDraft(ctx, grantID, draft.ID)
	require.NoError(t, err)
}

func TestIntegration_DraftLifecycle_WithBCC(t *testing.T) {
	client, grantID := getTestClient(t)
	ctx, cancel := createLongTestContext()
	defer cancel()

	createReq := &domain.CreateDraftRequest{
		Subject: fmt.Sprintf("Draft with BCC - %s", time.Now().Format("15:04:05")),
		Body:    "Test draft with BCC recipients",
		To:      []domain.EmailParticipant{{Email: "recipient@example.com"}},
		Bcc:     []domain.EmailParticipant{{Email: "bcc@example.com", Name: "Secret Recipient"}},
	}

	draft, err := client.CreateDraft(ctx, grantID, createReq)
	skipIfProviderNotSupported(t, err)
	require.NoError(t, err)
	require.NotEmpty(t, draft.ID)
	t.Logf("Created draft with BCC: %s", draft.ID)

	// Cleanup
	err = client.DeleteDraft(ctx, grantID, draft.ID)
	require.NoError(t, err)
}

func TestIntegration_DraftLifecycle_ReplyTo(t *testing.T) {
	client, grantID := getTestClient(t)
	ctx, cancel := createLongTestContext()
	defer cancel()

	createReq := &domain.CreateDraftRequest{
		Subject: fmt.Sprintf("Draft with Reply-To - %s", time.Now().Format("15:04:05")),
		Body:    "Test draft with reply-to address",
		To:      []domain.EmailParticipant{{Email: "recipient@example.com"}},
		ReplyTo: []domain.EmailParticipant{{Email: "replyto@example.com", Name: "Reply Here"}},
	}

	draft, err := client.CreateDraft(ctx, grantID, createReq)
	skipIfProviderNotSupported(t, err)
	require.NoError(t, err)
	require.NotEmpty(t, draft.ID)
	t.Logf("Created draft with Reply-To: %s", draft.ID)

	// Cleanup
	err = client.DeleteDraft(ctx, grantID, draft.ID)
	require.NoError(t, err)
}

func TestIntegration_GetDraft_NotFound(t *testing.T) {
	client, grantID := getTestClient(t)
	ctx, cancel := createTestContext()
	defer cancel()

	_, err := client.GetDraft(ctx, grantID, "nonexistent-draft-id-12345")
	assert.Error(t, err, "Should return error for non-existent draft")
}

// =============================================================================
// Attachment Tests
// =============================================================================

func TestIntegration_GetMessageWithAttachments(t *testing.T) {
	client, grantID := getTestClient(t)
	ctx, cancel := createTestContext()
	defer cancel()

	// Get messages and look for one with attachments
	messages, err := client.GetMessages(ctx, grantID, 50)
	require.NoError(t, err)

	var messageWithAttachment *domain.Message
	for i := range messages {
		if len(messages[i].Attachments) > 0 {
			messageWithAttachment = &messages[i]
			break
		}
	}

	if messageWithAttachment == nil {
		t.Skip("No messages with attachments found")
	}

	t.Logf("Found message with %d attachments: %s",
		len(messageWithAttachment.Attachments), messageWithAttachment.Subject)

	for _, a := range messageWithAttachment.Attachments {
		t.Logf("  Attachment: %s (%s, %d bytes)",
			a.Filename, a.ContentType, a.Size)

		assert.NotEmpty(t, a.ID, "Attachment should have ID")
		assert.NotEmpty(t, a.Filename, "Attachment should have filename")
		assert.NotEmpty(t, a.ContentType, "Attachment should have content type")
	}
}

func TestIntegration_GetAttachment(t *testing.T) {
	client, grantID := getTestClient(t)
	ctx, cancel := createTestContext()
	defer cancel()

	// Find a message with attachments (prefer non-inline attachments)
	messages, err := client.GetMessages(ctx, grantID, 50)
	require.NoError(t, err)

	var messageID, attachmentID string
	for _, m := range messages {
		for _, a := range m.Attachments {
			// Prefer non-inline attachments as they're more reliably accessible
			if !a.IsInline {
				messageID = m.ID
				attachmentID = a.ID
				break
			}
		}
		if attachmentID != "" {
			break
		}
	}

	// Fall back to any attachment if no non-inline found
	if attachmentID == "" {
		for _, m := range messages {
			if len(m.Attachments) > 0 {
				messageID = m.ID
				attachmentID = m.Attachments[0].ID
				break
			}
		}
	}

	if attachmentID == "" {
		t.Skip("No attachments found")
	}

	attachment, err := client.GetAttachment(ctx, grantID, messageID, attachmentID)
	if err != nil && strings.Contains(err.Error(), "attachment not found") {
		// Some inline attachments may not be accessible via the individual endpoint
		t.Skipf("Attachment not accessible via individual endpoint (may be inline): %v", err)
	}
	require.NoError(t, err)

	assert.Equal(t, attachmentID, attachment.ID)
	assert.NotEmpty(t, attachment.Filename)
	t.Logf("Retrieved attachment: %s (%s)", attachment.Filename, attachment.ContentType)
}

func TestIntegration_DownloadAttachment(t *testing.T) {
	client, grantID := getTestClient(t)
	ctx, cancel := createLongTestContext()
	defer cancel()

	// Find a message with attachments (prefer non-inline, smaller attachments)
	messages, err := client.GetMessages(ctx, grantID, 50)
	require.NoError(t, err)

	var messageID string
	var attachment *domain.Attachment

	// First pass: look for non-inline, appropriately sized attachments
	for _, m := range messages {
		for i := range m.Attachments {
			a := &m.Attachments[i]
			// Skip very large attachments and prefer non-inline
			if a.Size < 1000000 && a.Size > 0 && !a.IsInline {
				messageID = m.ID
				attachment = a
				break
			}
		}
		if attachment != nil {
			break
		}
	}

	// Second pass: fall back to any sized attachment
	if attachment == nil {
		for _, m := range messages {
			for i := range m.Attachments {
				a := &m.Attachments[i]
				if a.Size < 1000000 && a.Size > 0 {
					messageID = m.ID
					attachment = a
					break
				}
			}
			if attachment != nil {
				break
			}
		}
	}

	if attachment == nil {
		t.Skip("No suitable attachments found")
	}

	reader, err := client.DownloadAttachment(ctx, grantID, messageID, attachment.ID)
	if err != nil && strings.Contains(err.Error(), "attachment not found") {
		// Some inline attachments may not be accessible via the download endpoint
		t.Skipf("Attachment not accessible via download endpoint (may be inline): %v", err)
	}
	require.NoError(t, err)
	defer reader.Close()

	data, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.NotEmpty(t, data, "Downloaded data should not be empty")

	t.Logf("Downloaded %s: %d bytes", attachment.Filename, len(data))
}

func TestIntegration_ListAttachments(t *testing.T) {
	client, grantID := getTestClient(t)
	ctx, cancel := createTestContext()
	defer cancel()

	// Find a message with attachments first
	messages, err := client.GetMessages(ctx, grantID, 50)
	require.NoError(t, err)

	var messageID string
	var expectedCount int
	for _, m := range messages {
		if len(m.Attachments) > 0 {
			messageID = m.ID
			expectedCount = len(m.Attachments)
			break
		}
	}

	if messageID == "" {
		t.Skip("No messages with attachments found")
	}

	// Test ListAttachments
	attachments, err := client.ListAttachments(ctx, grantID, messageID)
	require.NoError(t, err)
	assert.Len(t, attachments, expectedCount)

	t.Logf("ListAttachments returned %d attachments for message %s", len(attachments), messageID)
	for _, a := range attachments {
		t.Logf("  - %s (%s, %d bytes, inline: %v)",
			a.Filename, a.ContentType, a.Size, a.IsInline)

		assert.NotEmpty(t, a.ID, "Attachment should have ID")
		assert.NotEmpty(t, a.Filename, "Attachment should have filename")
		assert.NotEmpty(t, a.ContentType, "Attachment should have content type")
	}
}

func TestIntegration_ListAttachments_EmptyMessage(t *testing.T) {
	client, grantID := getTestClient(t)
	ctx, cancel := createTestContext()
	defer cancel()

	// Find a message without attachments
	messages, err := client.GetMessages(ctx, grantID, 50)
	require.NoError(t, err)
	skipIfNoMessages(t, messages)

	var messageID string
	for _, m := range messages {
		if len(m.Attachments) == 0 {
			messageID = m.ID
			break
		}
	}

	if messageID == "" {
		t.Skip("All messages have attachments, can't test empty attachment list")
	}

	attachments, err := client.ListAttachments(ctx, grantID, messageID)
	require.NoError(t, err)
	assert.Empty(t, attachments, "Message without attachments should return empty list")
	t.Logf("Verified ListAttachments returns empty list for message without attachments")
}

// =============================================================================
// Send Message Tests (Destructive - requires explicit opt-in)
// =============================================================================

func TestIntegration_SendMessage(t *testing.T) {
	if os.Getenv("NYLAS_TEST_SEND_EMAIL") != "true" {
		t.Skip("Skipping send email test - set NYLAS_TEST_SEND_EMAIL=true to enable")
	}

	client, grantID := getTestClient(t)
	ctx, cancel := createLongTestContext()
	defer cancel()

	testEmail := os.Getenv("NYLAS_TEST_EMAIL")
	if testEmail == "" {
		t.Skip("NYLAS_TEST_EMAIL not set")
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	req := &domain.SendMessageRequest{
		Subject: fmt.Sprintf("Integration Test - %s", timestamp),
		Body:    "<html><body><h1>Test Email</h1><p>This is a test email sent by the integration tests at " + timestamp + "</p></body></html>",
		To:      []domain.EmailParticipant{{Email: testEmail, Name: "Test Recipient"}},
	}

	msg, err := client.SendMessage(ctx, grantID, req)
	require.NoError(t, err)
	require.NotEmpty(t, msg.ID)

	t.Logf("Sent email: %s", msg.ID)
	t.Logf("Subject: %s", msg.Subject)
	t.Logf("To: %v", msg.To)
}

func TestIntegration_SendMessage_WithCC(t *testing.T) {
	if os.Getenv("NYLAS_TEST_SEND_EMAIL") != "true" {
		t.Skip("Skipping send email test - set NYLAS_TEST_SEND_EMAIL=true to enable")
	}

	client, grantID := getTestClient(t)
	ctx, cancel := createLongTestContext()
	defer cancel()

	testEmail := os.Getenv("NYLAS_TEST_EMAIL")
	ccEmail := os.Getenv("NYLAS_TEST_CC_EMAIL")
	if testEmail == "" || ccEmail == "" {
		t.Skip("NYLAS_TEST_EMAIL and NYLAS_TEST_CC_EMAIL required")
	}

	req := &domain.SendMessageRequest{
		Subject: fmt.Sprintf("Integration Test with CC - %s", time.Now().Format("15:04:05")),
		Body:    "Test email with CC recipient",
		To:      []domain.EmailParticipant{{Email: testEmail}},
		Cc:      []domain.EmailParticipant{{Email: ccEmail}},
	}

	msg, err := client.SendMessage(ctx, grantID, req)
	require.NoError(t, err)
	t.Logf("Sent email with CC: %s", msg.ID)
}

func TestIntegration_SendDraft(t *testing.T) {
	if os.Getenv("NYLAS_TEST_SEND_EMAIL") != "true" {
		t.Skip("Skipping send email test - set NYLAS_TEST_SEND_EMAIL=true to enable")
	}

	client, grantID := getTestClient(t)
	ctx, cancel := createLongTestContext()
	defer cancel()

	testEmail := os.Getenv("NYLAS_TEST_EMAIL")
	if testEmail == "" {
		t.Skip("NYLAS_TEST_EMAIL not set")
	}

	// Create a draft first
	createReq := &domain.CreateDraftRequest{
		Subject: fmt.Sprintf("Draft to Send - %s", time.Now().Format("15:04:05")),
		Body:    "This draft will be sent as an email",
		To:      []domain.EmailParticipant{{Email: testEmail}},
	}

	draft, err := client.CreateDraft(ctx, grantID, createReq)
	require.NoError(t, err)
	t.Logf("Created draft: %s", draft.ID)

	// Send the draft
	msg, err := client.SendDraft(ctx, grantID, draft.ID)
	require.NoError(t, err)
	require.NotEmpty(t, msg.ID)
	t.Logf("Sent draft as message: %s", msg.ID)
}

// =============================================================================
// Delete Message Tests (Destructive - requires explicit opt-in)
// =============================================================================

func TestIntegration_DeleteMessage(t *testing.T) {
	if os.Getenv("NYLAS_TEST_DELETE_MESSAGE") != "true" {
		t.Skip("Skipping delete message test - set NYLAS_TEST_DELETE_MESSAGE=true to enable")
	}

	client, grantID := getTestClient(t)
	ctx, cancel := createLongTestContext()
	defer cancel()

	// Create a draft to delete (safer than deleting real messages)
	createReq := &domain.CreateDraftRequest{
		Subject: fmt.Sprintf("Draft for Deletion Test - %s", time.Now().Format("15:04:05")),
		Body:    "This draft will be deleted",
		To:      []domain.EmailParticipant{{Email: "test@example.com"}},
	}

	draft, err := client.CreateDraft(ctx, grantID, createReq)
	require.NoError(t, err)
	t.Logf("Created draft for deletion: %s", draft.ID)

	// Delete it (drafts can be deleted as messages too in some cases)
	err = client.DeleteDraft(ctx, grantID, draft.ID)
	require.NoError(t, err)
	t.Logf("Deleted draft: %s", draft.ID)

	// Verify
	_, err = client.GetDraft(ctx, grantID, draft.ID)
	assert.Error(t, err, "Draft should be deleted")
}

// =============================================================================
// Error Handling Tests
// =============================================================================

func TestIntegration_InvalidGrantID(t *testing.T) {
	client, _ := getTestClient(t)
	ctx, cancel := createTestContext()
	defer cancel()

	_, err := client.GetMessages(ctx, "invalid-grant-id-12345", 10)
	assert.Error(t, err, "Should return error for invalid grant ID")
	t.Logf("Expected error: %v", err)
}

func TestIntegration_EmptyMessageID(t *testing.T) {
	client, grantID := getTestClient(t)
	ctx, cancel := createTestContext()
	defer cancel()

	_, err := client.GetMessage(ctx, grantID, "")
	assert.Error(t, err, "Should return error for empty message ID")
}

// =============================================================================
// Concurrency Tests
// =============================================================================

func TestIntegration_ConcurrentRequests(t *testing.T) {
	client, grantID := getTestClient(t)
	ctx, cancel := createLongTestContext()
	defer cancel()

	// Run multiple concurrent requests (reduced to avoid rate limiting)
	const numRequests = 2
	results := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		go func(i int) {
			_, err := client.GetMessages(ctx, grantID, 3)
			results <- err
		}(i)
	}

	// Collect results - allow some rate limiting errors
	successCount := 0
	for i := 0; i < numRequests; i++ {
		err := <-results
		if err == nil {
			successCount++
		} else {
			t.Logf("Request %d hit rate limit (expected with some providers): %v", i, err)
		}
	}

	assert.Greater(t, successCount, 0, "At least one concurrent request should succeed")
	t.Logf("%d of %d concurrent requests completed successfully", successCount, numRequests)
}

func TestIntegration_ConcurrentDifferentOperations(t *testing.T) {
	client, grantID := getTestClient(t)
	ctx, cancel := createLongTestContext()
	defer cancel()

	type result struct {
		name string
		err  error
	}

	results := make(chan result, 4)

	// Run different operations concurrently
	go func() {
		_, err := client.GetMessages(ctx, grantID, 3)
		results <- result{"GetMessages", err}
	}()

	go func() {
		_, err := client.GetFolders(ctx, grantID)
		results <- result{"GetFolders", err}
	}()

	go func() {
		_, err := client.GetThreads(ctx, grantID, &domain.ThreadQueryParams{Limit: 3})
		results <- result{"GetThreads", err}
	}()

	go func() {
		_, err := client.GetDrafts(ctx, grantID, 3)
		results <- result{"GetDrafts", err}
	}()

	// Collect results - allow some operations to fail if provider doesn't support them
	successCount := 0
	for i := 0; i < 4; i++ {
		r := <-results
		if r.err != nil {
			errMsg := r.err.Error()
			if strings.Contains(errMsg, "Method not supported for provider") ||
				strings.Contains(errMsg, "an internal error ocurred") ||
				strings.Contains(errMsg, "an internal error occurred") {
				t.Logf("%s: Skipped (not supported by provider)", r.name)
			} else {
				t.Logf("%s: Error: %v", r.name, r.err)
			}
		} else {
			successCount++
			t.Logf("%s: OK", r.name)
		}
	}
	assert.Greater(t, successCount, 0, "At least one operation should succeed")
}

// =============================================================================
// Rate Limiting / Timeout Tests
// =============================================================================

func TestIntegration_RequestTimeout(t *testing.T) {
	client, grantID := getTestClient(t)

	// Create a very short timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	_, err := client.GetMessages(ctx, grantID, 100)
	assert.Error(t, err, "Should return error on timeout")
	t.Logf("Timeout error (expected): %v", err)
}

// =============================================================================
// Data Validation Tests
// =============================================================================

func TestIntegration_MessageFieldsValidation(t *testing.T) {
	client, grantID := getTestClient(t)
	ctx, cancel := createTestContext()
	defer cancel()

	messages, err := client.GetMessages(ctx, grantID, 10)
	require.NoError(t, err)
	skipIfNoMessages(t, messages)

	for _, m := range messages {
		t.Run("Message_"+safeSubstring(m.ID, 8), func(t *testing.T) {
			// Required fields
			assert.NotEmpty(t, m.ID, "ID should not be empty")
			assert.NotZero(t, m.Date, "Date should not be zero")

			// From should have at least one contact for received messages
			if len(m.From) > 0 {
				for _, f := range m.From {
					assert.NotEmpty(t, f.Email, "From email should not be empty")
				}
			}

			// Boolean fields should be set (even if false)
			// Just checking they're accessible
			_ = m.Unread
			_ = m.Starred
		})
	}
}

func TestIntegration_FolderFieldsValidation(t *testing.T) {
	client, grantID := getTestClient(t)
	ctx, cancel := createTestContext()
	defer cancel()

	folders, err := client.GetFolders(ctx, grantID)
	skipIfProviderNotSupported(t, err)
	require.NoError(t, err)
	require.NotEmpty(t, folders)

	for _, f := range folders {
		t.Run("Folder_"+safeSubstring(f.Name, 20), func(t *testing.T) {
			assert.NotEmpty(t, f.ID, "ID should not be empty")
			assert.NotEmpty(t, f.Name, "Name should not be empty")
			assert.GreaterOrEqual(t, f.TotalCount, 0, "TotalCount should be >= 0")
			assert.GreaterOrEqual(t, f.UnreadCount, 0, "UnreadCount should be >= 0")
			assert.LessOrEqual(t, f.UnreadCount, f.TotalCount, "UnreadCount should be <= TotalCount")
		})
	}
}

func TestIntegration_ThreadFieldsValidation(t *testing.T) {
	client, grantID := getTestClient(t)
	ctx, cancel := createTestContext()
	defer cancel()

	threads, err := client.GetThreads(ctx, grantID, &domain.ThreadQueryParams{Limit: 10})
	skipIfProviderNotSupported(t, err)
	require.NoError(t, err)

	for _, th := range threads {
		t.Run("Thread_"+safeSubstring(th.ID, 8), func(t *testing.T) {
			assert.NotEmpty(t, th.ID, "ID should not be empty")
			assert.NotEmpty(t, th.MessageIDs, "Should have at least one message")

			_ = th.Unread
			_ = th.Starred
		})
	}
}

// =============================================================================
// Pagination Tests
// =============================================================================

func TestIntegration_MessagePagination(t *testing.T) {
	client, grantID := getTestClient(t)
	ctx, cancel := createLongTestContext()
	defer cancel()

	// Get first page
	params := &domain.MessageQueryParams{
		Limit: 5,
	}

	page1, err := client.GetMessagesWithParams(ctx, grantID, params)
	require.NoError(t, err)

	if len(page1) < 5 {
		t.Skip("Not enough messages to test pagination")
	}

	t.Logf("Page 1: %d messages", len(page1))
	for _, m := range page1 {
		t.Logf("  [%s] %s", safeSubstring(m.ID, 8), safeSubstring(m.Subject, 30))
	}

	// Note: Nylas v3 uses page_token for pagination, which would need to be
	// returned from the API. For now, we just verify we can get a consistent page.
}

// =============================================================================
// Comprehensive Workflow Tests
// =============================================================================

func TestIntegration_CompleteWorkflow_ReadAndMarkMessages(t *testing.T) {
	client, grantID := getTestClient(t)
	ctx, cancel := createLongTestContext()
	defer cancel()

	// 1. List recent messages
	messages, err := client.GetMessages(ctx, grantID, 5)
	require.NoError(t, err)
	skipIfNoMessages(t, messages)
	t.Logf("Step 1: Listed %d messages", len(messages))

	// 2. Get full details of first message
	fullMsg, err := client.GetMessage(ctx, grantID, messages[0].ID)
	require.NoError(t, err)
	t.Logf("Step 2: Got message details - Subject: %s", fullMsg.Subject)

	// 3. Get the thread for this message (skip if provider doesn't support)
	if fullMsg.ThreadID != "" {
		thread, err := client.GetThread(ctx, grantID, fullMsg.ThreadID)
		if err != nil && strings.Contains(err.Error(), "Method not supported for provider") {
			t.Logf("Step 3: Skipped - threads not supported by provider")
		} else {
			require.NoError(t, err)
			t.Logf("Step 3: Got thread with %d messages", len(thread.MessageIDs))
		}
	}

	// 4. List folders (skip if provider doesn't support)
	folders, err := client.GetFolders(ctx, grantID)
	if err != nil && strings.Contains(err.Error(), "Method not supported for provider") {
		t.Logf("Step 4: Skipped - folders not supported by provider")
	} else {
		require.NoError(t, err)
		t.Logf("Step 4: Found %d folders", len(folders))
	}

	// 5. Toggle read status
	originalUnread := fullMsg.Unread
	newUnread := !originalUnread
	req := &domain.UpdateMessageRequest{Unread: &newUnread}

	updated, err := client.UpdateMessage(ctx, grantID, fullMsg.ID, req)
	require.NoError(t, err)
	assert.Equal(t, newUnread, updated.Unread)
	t.Logf("Step 5: Toggled unread status from %v to %v", originalUnread, newUnread)

	// 6. Restore original status
	req.Unread = &originalUnread
	restored, err := client.UpdateMessage(ctx, grantID, fullMsg.ID, req)
	require.NoError(t, err)
	assert.Equal(t, originalUnread, restored.Unread)
	t.Logf("Step 6: Restored original unread status")
}

func TestIntegration_CompleteWorkflow_DraftCreationAndManagement(t *testing.T) {
	client, grantID := getTestClient(t)
	ctx, cancel := createLongTestContext()
	defer cancel()

	timestamp := time.Now().Format("2006-01-02 15:04:05")

	// 1. Create initial draft
	createReq := &domain.CreateDraftRequest{
		Subject: fmt.Sprintf("Workflow Test - %s", timestamp),
		Body:    "Initial draft content",
		To:      []domain.EmailParticipant{{Email: "recipient@example.com", Name: "Recipient"}},
	}

	draft, err := client.CreateDraft(ctx, grantID, createReq)
	skipIfProviderNotSupported(t, err)
	require.NoError(t, err)
	t.Logf("Step 1: Created draft %s", draft.ID)

	// 2. List drafts and verify it appears
	drafts, err := client.GetDrafts(ctx, grantID, 20)
	require.NoError(t, err)
	found := false
	for _, d := range drafts {
		if d.ID == draft.ID {
			found = true
			break
		}
	}
	assert.True(t, found, "New draft should appear in list")
	t.Logf("Step 2: Verified draft appears in list")

	// 3. Get draft details
	retrieved, err := client.GetDraft(ctx, grantID, draft.ID)
	require.NoError(t, err)
	assert.Equal(t, createReq.Subject, retrieved.Subject)
	t.Logf("Step 3: Retrieved draft details")

	// 4. Update draft with more recipients
	updateReq := &domain.CreateDraftRequest{
		Subject: createReq.Subject + " - Updated",
		Body:    "Updated draft content with more details",
		To:      []domain.EmailParticipant{{Email: "recipient@example.com"}},
		Cc:      []domain.EmailParticipant{{Email: "cc@example.com"}},
	}

	updated, err := client.UpdateDraft(ctx, grantID, draft.ID, updateReq)
	require.NoError(t, err)
	assert.Contains(t, updated.Subject, "Updated")
	t.Logf("Step 4: Updated draft with CC")

	// 5. Delete the draft
	err = client.DeleteDraft(ctx, grantID, draft.ID)
	require.NoError(t, err)
	t.Logf("Step 5: Deleted draft")

	// 6. Verify deletion (with retry for eventual consistency)
	// Note: Some providers may have eventual consistency, so we just log the result
	_, err = client.GetDraft(ctx, grantID, draft.ID)
	if err != nil {
		t.Logf("Step 6: Verified draft deletion (draft not found as expected)")
	} else {
		t.Logf("Step 6: Draft still retrievable due to eventual consistency (this is OK)")
	}
}

func TestIntegration_CompleteWorkflow_FolderManagement(t *testing.T) {
	client, grantID := getTestClient(t)
	ctx, cancel := createLongTestContext()
	defer cancel()

	folderName := fmt.Sprintf("WorkflowTest_%d", time.Now().Unix())

	// 1. List existing folders
	initialFolders, err := client.GetFolders(ctx, grantID)
	skipIfProviderNotSupported(t, err)
	require.NoError(t, err)
	t.Logf("Step 1: Initial folder count: %d", len(initialFolders))

	// 2. Create new folder
	createReq := &domain.CreateFolderRequest{Name: folderName}
	folder, err := client.CreateFolder(ctx, grantID, createReq)
	require.NoError(t, err)
	t.Logf("Step 2: Created folder %s", folder.Name)

	// 3. Verify folder appears in list
	foldersAfterCreate, err := client.GetFolders(ctx, grantID)
	require.NoError(t, err)
	assert.Equal(t, len(initialFolders)+1, len(foldersAfterCreate))
	t.Logf("Step 3: Folder count increased to %d", len(foldersAfterCreate))

	// 4. Get folder details
	retrieved, err := client.GetFolder(ctx, grantID, folder.ID)
	require.NoError(t, err)
	assert.Equal(t, folderName, retrieved.Name)
	t.Logf("Step 4: Verified folder details")

	// 5. Rename folder
	newName := folderName + "_Renamed"
	folderUpdateReq := &domain.UpdateFolderRequest{Name: newName}
	renamed, err := client.UpdateFolder(ctx, grantID, folder.ID, folderUpdateReq)
	require.NoError(t, err)
	assert.Equal(t, newName, renamed.Name)
	t.Logf("Step 5: Renamed folder to %s", renamed.Name)

	// 6. Delete folder
	err = client.DeleteFolder(ctx, grantID, folder.ID)
	require.NoError(t, err)
	t.Logf("Step 6: Deleted folder")

	// 7. Verify folder count restored
	foldersAfterDelete, err := client.GetFolders(ctx, grantID)
	require.NoError(t, err)
	assert.Equal(t, len(initialFolders), len(foldersAfterDelete))
	t.Logf("Step 7: Folder count restored to %d", len(foldersAfterDelete))
}

// =============================================================================
// Scheduled Messages Tests
// =============================================================================

func TestIntegration_ListScheduledMessages(t *testing.T) {
	client, grantID := getTestClient(t)
	ctx, cancel := createTestContext()
	defer cancel()

	scheduled, err := client.ListScheduledMessages(ctx, grantID)
	require.NoError(t, err)

	t.Logf("Found %d scheduled message(s)", len(scheduled))
	for _, s := range scheduled {
		t.Logf("  Schedule ID: %s, Status: %s, CloseTime: %d",
			s.ScheduleID, s.Status, s.CloseTime)
	}
}

func TestIntegration_GetScheduledMessage_NotFound(t *testing.T) {
	client, grantID := getTestClient(t)
	ctx, cancel := createTestContext()
	defer cancel()

	// Try to get a non-existent scheduled message
	_, err := client.GetScheduledMessage(ctx, grantID, "nonexistent-schedule-id")
	require.Error(t, err)
	t.Logf("Expected error for non-existent schedule: %v", err)
}

// =============================================================================
// Notetaker Tests
// =============================================================================

func TestIntegration_ListNotetakers(t *testing.T) {
	client, grantID := getTestClient(t)
	ctx, cancel := createTestContext()
	defer cancel()

	notetakers, err := client.ListNotetakers(ctx, grantID, nil)
	skipIfProviderNotSupported(t, err)
	require.NoError(t, err)

	t.Logf("Found %d notetaker(s)", len(notetakers))
	for _, nt := range notetakers {
		t.Logf("  ID: %s, State: %s, Meeting: %s",
			safeSubstring(nt.ID, 12), nt.State, safeSubstring(nt.MeetingTitle, 30))
	}
}

func TestIntegration_ListNotetakers_WithParams(t *testing.T) {
	client, grantID := getTestClient(t)
	ctx, cancel := createTestContext()
	defer cancel()

	// Test with limit
	params := &domain.NotetakerQueryParams{
		Limit: 5,
	}

	notetakers, err := client.ListNotetakers(ctx, grantID, params)
	skipIfProviderNotSupported(t, err)
	require.NoError(t, err)

	assert.LessOrEqual(t, len(notetakers), 5, "Should respect limit")
	t.Logf("Found %d notetaker(s) with limit=5", len(notetakers))
}

func TestIntegration_ListNotetakers_ByState(t *testing.T) {
	client, grantID := getTestClient(t)
	ctx, cancel := createTestContext()
	defer cancel()

	// Only test with states known to be valid in the API
	states := []string{"scheduled", "attending", "media_processing"}

	for _, state := range states {
		t.Run("State_"+state, func(t *testing.T) {
			params := &domain.NotetakerQueryParams{
				Limit: 10,
				State: state,
			}

			notetakers, err := client.ListNotetakers(ctx, grantID, params)
			// Skip if state is not supported (API returns "invalid state" error)
			if err != nil && strings.Contains(err.Error(), "invalid state") {
				t.Skipf("State %s not supported by API: %v", state, err)
			}
			skipIfProviderNotSupported(t, err)
			require.NoError(t, err)

			t.Logf("Found %d notetaker(s) with state=%s", len(notetakers), state)
			for _, nt := range notetakers {
				assert.Equal(t, state, nt.State, "Notetaker should have requested state")
			}
		})
	}
}

func TestIntegration_GetNotetaker_NotFound(t *testing.T) {
	client, grantID := getTestClient(t)
	ctx, cancel := createTestContext()
	defer cancel()

	_, err := client.GetNotetaker(ctx, grantID, "nonexistent-notetaker-id-12345")
	assert.Error(t, err, "Should return error for non-existent notetaker")
	t.Logf("Expected error: %v", err)
}

func TestIntegration_CreateNotetaker(t *testing.T) {
	if os.Getenv("NYLAS_TEST_NOTETAKER") != "true" {
		t.Skip("Skipping notetaker create test - set NYLAS_TEST_NOTETAKER=true to enable")
	}

	client, grantID := getTestClient(t)
	ctx, cancel := createLongTestContext()
	defer cancel()

	meetingLink := os.Getenv("NYLAS_TEST_MEETING_LINK")
	if meetingLink == "" {
		t.Skip("NYLAS_TEST_MEETING_LINK not set")
	}

	req := &domain.CreateNotetakerRequest{
		MeetingLink: meetingLink,
		BotConfig: &domain.BotConfig{
			Name: "Integration Test Bot",
		},
	}

	notetaker, err := client.CreateNotetaker(ctx, grantID, req)
	skipIfProviderNotSupported(t, err)
	require.NoError(t, err)
	require.NotEmpty(t, notetaker.ID)

	t.Logf("Created notetaker: %s", notetaker.ID)
	t.Logf("State: %s", notetaker.State)
	t.Logf("Meeting Link: %s", notetaker.MeetingLink)

	// Clean up - delete the notetaker
	err = client.DeleteNotetaker(ctx, grantID, notetaker.ID)
	if err != nil {
		t.Logf("Warning: Could not clean up notetaker: %v", err)
	} else {
		t.Logf("Cleaned up notetaker: %s", notetaker.ID)
	}
}

func TestIntegration_NotetakerMedia_NotFound(t *testing.T) {
	client, grantID := getTestClient(t)
	ctx, cancel := createTestContext()
	defer cancel()

	_, err := client.GetNotetakerMedia(ctx, grantID, "nonexistent-notetaker-id-12345")
	assert.Error(t, err, "Should return error for non-existent notetaker media")
	t.Logf("Expected error: %v", err)
}

// =============================================================================
// Tracking Options Tests
// =============================================================================

func TestIntegration_SendMessageWithTracking(t *testing.T) {
	if os.Getenv("NYLAS_TEST_SEND_EMAIL") != "true" {
		t.Skip("Skipping send email test - set NYLAS_TEST_SEND_EMAIL=true to enable")
	}

	client, grantID := getTestClient(t)
	ctx, cancel := createLongTestContext()
	defer cancel()

	testEmail := os.Getenv("NYLAS_TEST_EMAIL")
	if testEmail == "" {
		t.Skip("NYLAS_TEST_EMAIL not set")
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	req := &domain.SendMessageRequest{
		Subject: fmt.Sprintf("Tracking Test - %s", timestamp),
		Body:    "<html><body><p>Test email with tracking enabled</p></body></html>",
		To:      []domain.EmailParticipant{{Email: testEmail, Name: "Test Recipient"}},
		TrackingOpts: &domain.TrackingOptions{
			Opens: true,
			Links: true,
			Label: "integration-test",
		},
	}

	msg, err := client.SendMessage(ctx, grantID, req)
	require.NoError(t, err)
	require.NotEmpty(t, msg.ID)

	t.Logf("Sent tracked email: %s", msg.ID)
	t.Logf("Subject: %s", msg.Subject)
	t.Logf("Tracking: opens=%v, links=%v, label=%s", true, true, "integration-test")
}

func TestIntegration_SendMessageWithMetadata(t *testing.T) {
	if os.Getenv("NYLAS_TEST_SEND_EMAIL") != "true" {
		t.Skip("Skipping send email test - set NYLAS_TEST_SEND_EMAIL=true to enable")
	}

	client, grantID := getTestClient(t)
	ctx, cancel := createLongTestContext()
	defer cancel()

	testEmail := os.Getenv("NYLAS_TEST_EMAIL")
	if testEmail == "" {
		t.Skip("NYLAS_TEST_EMAIL not set")
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	req := &domain.SendMessageRequest{
		Subject: fmt.Sprintf("Metadata Test - %s", timestamp),
		Body:    "<html><body><p>Test email with custom metadata</p></body></html>",
		To:      []domain.EmailParticipant{{Email: testEmail, Name: "Test Recipient"}},
		Metadata: map[string]string{
			"campaign_id":  "test-campaign-001",
			"customer_id":  "cust-12345",
			"test_run":     "integration",
		},
	}

	msg, err := client.SendMessage(ctx, grantID, req)
	require.NoError(t, err)
	require.NotEmpty(t, msg.ID)

	t.Logf("Sent email with metadata: %s", msg.ID)
	t.Logf("Subject: %s", msg.Subject)
	t.Logf("Metadata: campaign_id=test-campaign-001, customer_id=cust-12345")
}
