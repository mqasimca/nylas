//go:build teatestv1
// +build teatestv1

// This file is disabled because teatest package is still v1 and incompatible with Bubble Tea v2.
// Re-enable when teatest v2 is available by removing the build tag above.

package models

import (
	"context"
	"strings"
	"testing"
	"time"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/mqasimca/nylas/internal/adapters/keyring"
	"github.com/mqasimca/nylas/internal/adapters/nylas"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/tui2/components"
	"github.com/mqasimca/nylas/internal/tui2/state"
)

// TestIntegration_MessageList_FetchAndDisplay tests the full workflow of fetching and displaying messages.
func TestIntegration_MessageList_FetchAndDisplay(t *testing.T) {
	t.Skip("teatest package is still v1 and incompatible with Bubble Tea v2 - waiting for teatest v2 support")

	// Create mock client with test data
	client := nylas.NewMockClient()
	client.GetThreadsFunc = func(ctx context.Context, grantID string, params *domain.ThreadQueryParams) ([]domain.Thread, error) {
		return []domain.Thread{
			{
				ID:      "thread1",
				Subject: "Welcome Message",
				LatestDraftOrMessage: domain.Message{
					ID:      "msg1",
					Subject: "Welcome Message",
					From:    []domain.EmailParticipant{{Name: "Demo Team", Email: "team@example.com"}},
					To:      []domain.EmailParticipant{{Email: "user@example.com"}},
					Date:    time.Now().Add(-1 * time.Hour),
					Snippet: "Thanks for using our service!",
				},
				MessageIDs:            []string{"msg1"},
				LatestMessageRecvDate: time.Now().Add(-1 * time.Hour),
				Participants:          []domain.EmailParticipant{{Name: "Demo Team", Email: "team@example.com"}},
				Snippet:               "Thanks for using our service!",
			},
			{
				ID:      "thread2",
				Subject: "Important Update",
				LatestDraftOrMessage: domain.Message{
					ID:      "msg2",
					Subject: "Important Update",
					From:    []domain.EmailParticipant{{Name: "Admin", Email: "admin@example.com"}},
					To:      []domain.EmailParticipant{{Email: "user@example.com"}},
					Date:    time.Now().Add(-24 * time.Hour),
					Snippet: "Please review the latest changes.",
				},
				MessageIDs:            []string{"msg2"},
				LatestMessageRecvDate: time.Now().Add(-24 * time.Hour),
				Participants:          []domain.EmailParticipant{{Name: "Admin", Email: "admin@example.com"}},
				Snippet:               "Please review the latest changes.",
			},
		}, nil
	}

	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "user@example.com", "google")

	// Create model
	model := NewMessageList(global)
	model.global.SetWindowSize(120, 40)

	// Run test with teatest
	tm := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(120, 40))
	defer func() {
		_ = tm.Quit() // Cleanup test model
	}()

	// Wait for initial threads to load
	teatest.WaitFor(
		t, tm.Output(),
		func(bts []byte) bool {
			return len(model.threads) > 0
		},
		teatest.WithCheckInterval(time.Millisecond*10),
		teatest.WithDuration(time.Second*3),
	)

	// Verify threads were loaded
	if len(model.threads) != 2 {
		t.Errorf("Expected 2 threads, got %d", len(model.threads))
	}

	if !client.GetThreadsCalled {
		t.Error("GetThreads should have been called")
	}
}

// TestIntegration_MessageList_NavigatePanes tests navigating between panes with Tab.
func TestIntegration_MessageList_NavigatePanes(t *testing.T) {
	client := nylas.NewMockClient()
	client.GetThreadsFunc = func(ctx context.Context, grantID string, params *domain.ThreadQueryParams) ([]domain.Thread, error) {
		return []domain.Thread{
			{
				ID:                    "thread1",
				Subject:               "Test 1",
				LatestDraftOrMessage:  domain.Message{ID: "msg1", Subject: "Test 1", From: []domain.EmailParticipant{{Email: "test@example.com"}}},
				MessageIDs:            []string{"msg1"},
				Participants:          []domain.EmailParticipant{{Email: "test@example.com"}},
				LatestMessageRecvDate: time.Now(),
			},
		}, nil
	}
	client.GetFoldersFunc = func(ctx context.Context, grantID string) ([]domain.Folder, error) {
		return []domain.Folder{
			{ID: "inbox", Name: "Inbox", TotalCount: 10},
			{ID: "sent", Name: "Sent", TotalCount: 5},
		}, nil
	}

	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "user@example.com", "google")

	model := NewMessageList(global)
	model.global.SetWindowSize(120, 40)

	tm := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(120, 40))
	defer func() {
		_ = tm.Quit() // Cleanup test model
	}()

	// Wait for messages to load
	time.Sleep(100 * time.Millisecond)

	// Initial pane should be MessagePane
	if model.layout.GetFocused() != components.MessagePane {
		t.Errorf("Initial pane should be MessagePane, got %v", model.layout.GetFocused())
	}

	// Press Tab to go to next pane
	tm.Send(tea.KeyMsg{Type: tea.KeyTab})
	time.Sleep(50 * time.Millisecond)

	// Should now be on PreviewPane
	if model.layout.GetFocused() != components.PreviewPane {
		t.Errorf("After Tab, should be PreviewPane, got %v", model.layout.GetFocused())
	}

	// Press Tab again to go to FolderPane
	tm.Send(tea.KeyMsg{Type: tea.KeyTab})
	time.Sleep(100 * time.Millisecond) // Extra time for folder loading

	// Should now be on FolderPane and folders should be loading/loaded
	if model.layout.GetFocused() != components.FolderPane {
		t.Errorf("After second Tab, should be FolderPane, got %v", model.layout.GetFocused())
	}

	// Wait a bit for folders to load
	teatest.WaitFor(
		t, tm.Output(),
		func(bts []byte) bool {
			return model.foldersLoaded
		},
		teatest.WithCheckInterval(time.Millisecond*10),
		teatest.WithDuration(time.Second*2),
	)

	if !client.GetFoldersCalled {
		t.Error("GetFolders should have been called when focusing FolderPane")
	}
}

// TestIntegration_MessageList_SelectThread tests selecting a thread with Enter key.
func TestIntegration_MessageList_SelectMessage(t *testing.T) {
	client := nylas.NewMockClient()
	client.GetThreadsFunc = func(ctx context.Context, grantID string, params *domain.ThreadQueryParams) ([]domain.Thread, error) {
		return []domain.Thread{
			{
				ID:      "thread-important",
				Subject: "Important: Read This",
				LatestDraftOrMessage: domain.Message{
					ID:      "msg-important",
					Subject: "Important: Read This",
					From:    []domain.EmailParticipant{{Name: "Boss", Email: "boss@example.com"}},
					Date:    time.Now(),
				},
				MessageIDs:            []string{"msg-important"},
				Participants:          []domain.EmailParticipant{{Name: "Boss", Email: "boss@example.com"}},
				LatestMessageRecvDate: time.Now(),
			},
		}, nil
	}

	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "user@example.com", "google")

	model := NewMessageList(global)
	model.global.SetWindowSize(120, 40)

	tm := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(120, 40))
	defer func() {
		_ = tm.Quit() // Cleanup test model
	}()

	// Wait for messages to load
	teatest.WaitFor(
		t, tm.Output(),
		func(bts []byte) bool {
			return len(model.threads) > 0
		},
		teatest.WithCheckInterval(time.Millisecond*10),
		teatest.WithDuration(time.Second*2),
	)

	// Make sure we're on MessagePane
	model.layout.FocusPane(components.MessagePane)

	// Press Enter to select message - this should trigger navigation
	// We test that the model handles the key press without error
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	time.Sleep(50 * time.Millisecond)

	// Verify model is in a valid state after selection
	if len(model.threads) == 0 {
		t.Error("Messages should still be loaded after selection")
	}

	if model.layout.GetFocused() != components.MessagePane {
		t.Error("Focus should remain on MessagePane after Enter")
	}
}

// TestIntegration_MessageList_FilterByFolder tests filtering threads by folder.
func TestIntegration_MessageList_FilterByFolder(t *testing.T) {
	client := nylas.NewMockClient()

	// Initial and filtered threads
	client.GetThreadsFunc = func(ctx context.Context, grantID string, params *domain.ThreadQueryParams) ([]domain.Thread, error) {
		if params != nil && len(params.In) > 0 && params.In[0] == "sent" {
			return []domain.Thread{
				{
					ID:      "thread-sent-1",
					Subject: "Sent Message 1",
					LatestDraftOrMessage: domain.Message{
						ID:      "msg-sent-1",
						Subject: "Sent Message 1",
					},
					MessageIDs:            []string{"msg-sent-1"},
					LatestMessageRecvDate: time.Now(),
				},
				{
					ID:      "thread-sent-2",
					Subject: "Sent Message 2",
					LatestDraftOrMessage: domain.Message{
						ID:      "msg-sent-2",
						Subject: "Sent Message 2",
					},
					MessageIDs:            []string{"msg-sent-2"},
					LatestMessageRecvDate: time.Now(),
				},
			}, nil
		}
		return []domain.Thread{
			{
				ID:      "thread1",
				Subject: "Inbox Message",
				LatestDraftOrMessage: domain.Message{
					ID:      "msg1",
					Subject: "Inbox Message",
				},
				MessageIDs:            []string{"msg1"},
				LatestMessageRecvDate: time.Now(),
			},
		}, nil
	}

	// Folder list
	client.GetFoldersFunc = func(ctx context.Context, grantID string) ([]domain.Folder, error) {
		return []domain.Folder{
			{ID: "inbox", Name: "Inbox", TotalCount: 10},
			{ID: "sent", Name: "Sent", TotalCount: 5},
		}, nil
	}

	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "user@example.com", "google")

	model := NewMessageList(global)
	model.global.SetWindowSize(120, 40)

	tm := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(120, 40))
	defer func() {
		_ = tm.Quit() // Cleanup test model
	}()

	// Wait for initial messages
	time.Sleep(100 * time.Millisecond)

	// Navigate to FolderPane to load folders
	tm.Send(tea.KeyMsg{Type: tea.KeyTab}) // -> Preview
	time.Sleep(50 * time.Millisecond)
	tm.Send(tea.KeyMsg{Type: tea.KeyTab}) // -> Folders
	time.Sleep(100 * time.Millisecond)

	// Wait for folders to load
	teatest.WaitFor(
		t, tm.Output(),
		func(bts []byte) bool {
			return model.foldersLoaded
		},
		teatest.WithCheckInterval(time.Millisecond*10),
		teatest.WithDuration(time.Second*2),
	)

	// Set folders manually for testing (simulate loaded state)
	model.layout.SetFolders([]list.Item{
		components.FolderItem{Folder: domain.Folder{ID: "sent", Name: "Sent"}},
	})
	model.foldersLoaded = true

	// Press Enter to select "Sent" folder
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	// Wait for filtered messages to load
	teatest.WaitFor(
		t, tm.Output(),
		func(bts []byte) bool {
			return model.selectedFolderID == "sent" && len(model.threads) > 1
		},
		teatest.WithCheckInterval(time.Millisecond*10),
		teatest.WithDuration(time.Second*2),
	)

	if !client.GetThreadsCalled {
		t.Error("GetThreads should have been called for folder filter")
	}

	if len(model.threads) != 2 {
		t.Errorf("Expected 2 sent threads, got %d", len(model.threads))
	}
}

// TestIntegration_MessageList_Refresh tests the refresh functionality.
func TestIntegration_MessageList_Refresh(t *testing.T) {
	callCount := 0
	client := nylas.NewMockClient()
	client.GetThreadsFunc = func(ctx context.Context, grantID string, params *domain.ThreadQueryParams) ([]domain.Thread, error) {
		callCount++
		return []domain.Thread{
			{
				ID:      "thread1",
				Subject: "Message 1",
				LatestDraftOrMessage: domain.Message{
					ID:      "msg1",
					Subject: "Message 1",
					From:    []domain.EmailParticipant{{Email: "test@example.com"}},
				},
				MessageIDs:            []string{"msg1"},
				Participants:          []domain.EmailParticipant{{Email: "test@example.com"}},
				LatestMessageRecvDate: time.Now(),
			},
		}, nil
	}

	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "user@example.com", "google")

	model := NewMessageList(global)
	model.global.SetWindowSize(120, 40)

	tm := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(120, 40))
	defer func() {
		_ = tm.Quit() // Cleanup test model
	}()

	// Wait for initial load
	teatest.WaitFor(
		t, tm.Output(),
		func(bts []byte) bool {
			return callCount >= 1
		},
		teatest.WithCheckInterval(time.Millisecond*10),
		teatest.WithDuration(time.Second*2),
	)

	initialCalls := callCount

	// Press 'ctrl+r' to refresh
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlR})

	// Wait for refresh to complete
	teatest.WaitFor(
		t, tm.Output(),
		func(bts []byte) bool {
			return callCount > initialCalls
		},
		teatest.WithCheckInterval(time.Millisecond*10),
		teatest.WithDuration(time.Second*2),
	)

	if callCount <= initialCalls {
		t.Errorf("Refresh should trigger another GetThreads call, initial=%d, current=%d", initialCalls, callCount)
	}
}

// TestIntegration_MessageList_BackNavigation tests going back with Esc.
func TestIntegration_MessageList_BackNavigation(t *testing.T) {
	client := nylas.NewMockClient()
	client.GetThreadsFunc = func(ctx context.Context, grantID string, params *domain.ThreadQueryParams) ([]domain.Thread, error) {
		return []domain.Thread{}, nil
	}

	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "user@example.com", "google")

	model := NewMessageList(global)
	model.global.SetWindowSize(120, 40)

	tm := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(120, 40))
	defer func() {
		_ = tm.Quit() // Cleanup test model
	}()

	// Wait for initial load
	time.Sleep(100 * time.Millisecond)

	// Press Esc to go back - this should trigger BackMsg
	// We test that the model handles the key press without error
	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})
	time.Sleep(50 * time.Millisecond)

	// Verify model is in a valid state after Esc
	if model.err != nil {
		t.Errorf("Model should not have error after Esc: %v", model.err)
	}
}

// TestIntegration_ComposeNewMessage tests composing a new email workflow.
func TestIntegration_ComposeNewMessage(t *testing.T) {
	client := nylas.NewMockClient()
	client.SendMessageFunc = func(ctx context.Context, grantID string, req *domain.SendMessageRequest) (*domain.Message, error) {
		return &domain.Message{
			ID:      "msg-sent-123",
			Subject: req.Subject,
			To:      req.To,
			Body:    req.Body,
		}, nil
	}

	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "user@example.com", "google")

	// Create compose model with new message mode
	data := ComposeData{Mode: ComposeModeNew}
	model := NewCompose(global, data)
	model.global.SetWindowSize(120, 40)

	// Verify initial state
	if model.mode != ComposeModeNew {
		t.Errorf("Expected mode ComposeModeNew, got %v", model.mode)
	}

	// Verify form fields are initialized
	if model.toInput.Value() != "" {
		t.Error("To field should be empty for new message")
	}

	// Simulate form filling
	model.toInput.SetValue("test@example.com")
	model.subjectInput.SetValue("Test Email")
	model.bodyInput.SetValue("This is a test message.")

	// Verify validation passes
	if !model.validate() {
		t.Errorf("Validation should pass with valid data, errors: %v", model.validationErrors)
	}
}

// TestIntegration_ReplyToMessage tests the reply workflow.
func TestIntegration_ReplyToMessage(t *testing.T) {
	client := nylas.NewMockClient()
	client.GetThreadsFunc = func(ctx context.Context, grantID string, params *domain.ThreadQueryParams) ([]domain.Thread, error) {
		return []domain.Thread{
			{
				ID:      "thread1",
				Subject: "Original Subject",
				LatestDraftOrMessage: domain.Message{
					ID:      "original-msg",
					Subject: "Original Subject",
					From:    []domain.EmailParticipant{{Name: "John Doe", Email: "john@example.com"}},
					To:      []domain.EmailParticipant{{Email: "user@example.com"}},
					Body:    "Original message body",
					Date:    time.Now(),
				},
				MessageIDs:            []string{"original-msg"},
				Participants:          []domain.EmailParticipant{{Name: "John Doe", Email: "john@example.com"}},
				LatestMessageRecvDate: time.Now(),
			},
		}, nil
	}

	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "user@example.com", "google")

	// Start with message list
	listModel := NewMessageList(global)
	listModel.global.SetWindowSize(120, 40)

	tm := teatest.NewTestModel(t, listModel, teatest.WithInitialTermSize(120, 40))
	defer func() {
		_ = tm.Quit()
	}()

	// Wait for threads to load
	teatest.WaitFor(
		t, tm.Output(),
		func(bts []byte) bool {
			return len(listModel.threads) > 0
		},
		teatest.WithCheckInterval(time.Millisecond*10),
		teatest.WithDuration(time.Second*2),
	)

	// Press 'r' to reply - this should trigger NavigateMsg with compose data
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	time.Sleep(50 * time.Millisecond)

	// In a real app, the navigation would happen in app.go
	// For this test, we verify the model is in a valid state
	if len(listModel.threads) == 0 {
		t.Error("Threads should still be loaded after pressing reply")
	}

	// Test compose model with reply mode directly
	originalMsg := listModel.threads[0].LatestDraftOrMessage
	composeModel := NewCompose(global, ComposeData{
		Mode:    ComposeModeReply,
		Message: &originalMsg,
	})

	// Verify reply prefill
	if composeModel.toInput.Value() != "John Doe <john@example.com>" {
		t.Errorf("To field should be prefilled with original sender, got: %s", composeModel.toInput.Value())
	}

	subject := composeModel.subjectInput.Value()
	if subject != "Re: Original Subject" {
		t.Errorf("Subject should have Re: prefix, got: %s", subject)
	}

	body := composeModel.bodyInput.Value()
	if body == "" {
		t.Error("Body should contain quoted original message")
	}
	if !strings.Contains(body, ">") {
		t.Error("Body should contain quote markers")
	}
}

// TestIntegration_DraftAutosave tests the draft autosave functionality.
func TestIntegration_DraftAutosave(t *testing.T) {
	client := nylas.NewMockClient()
	client.CreateDraftFunc = func(ctx context.Context, grantID string, req *domain.CreateDraftRequest) (*domain.Draft, error) {
		return &domain.Draft{
			ID:      "draft-123",
			Subject: req.Subject,
			To:      req.To,
			Body:    req.Body,
		}, nil
	}
	client.UpdateDraftFunc = func(ctx context.Context, grantID, draftID string, req *domain.CreateDraftRequest) (*domain.Draft, error) {
		return &domain.Draft{
			ID:      draftID,
			Subject: req.Subject,
			To:      req.To,
			Body:    req.Body,
		}, nil
	}

	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "user@example.com", "google")

	// Create compose model
	data := ComposeData{Mode: ComposeModeNew}
	model := NewCompose(global, data)
	model.global.SetWindowSize(120, 40)

	// Verify autosave is enabled by default
	if !model.autosaveEnabled {
		t.Error("Autosave should be enabled by default")
	}

	if model.autosaveInterval != 30*time.Second {
		t.Errorf("Expected autosave interval 30s, got %v", model.autosaveInterval)
	}

	// Type some content to make draft dirty
	model.toInput.SetValue("test@example.com")
	model.subjectInput.SetValue("Test Draft")
	model.bodyInput.SetValue("Draft content")

	// Compute hash to test dirty detection
	hash1 := model.computeContentHash()
	if hash1 == "" {
		t.Error("Hash should not be empty")
	}

	// Modify content
	model.bodyInput.SetValue("Modified draft content")
	hash2 := model.computeContentHash()

	// Verify hash changed (dirty detection works)
	if hash1 == hash2 {
		t.Error("Hash should change when content changes")
	}
}

// TestIntegration_ToggleStar tests the star/unstar action workflow.
func TestIntegration_ToggleStar(t *testing.T) {
	client := nylas.NewMockClient()
	client.GetThreadFunc = func(ctx context.Context, grantID, threadID string) (*domain.Thread, error) {
		return &domain.Thread{
			ID:      threadID,
			Subject: "Test Message",
			LatestDraftOrMessage: domain.Message{
				ID:      "msg-123",
				Subject: "Test Message",
				From:    []domain.EmailParticipant{{Email: "test@example.com"}},
				Starred: false,
			},
			MessageIDs:            []string{"msg-123"},
			Participants:          []domain.EmailParticipant{{Email: "test@example.com"}},
			LatestMessageRecvDate: time.Now(),
		}, nil
	}
	client.GetMessageFunc = func(ctx context.Context, grantID, msgID string) (*domain.Message, error) {
		return &domain.Message{
			ID:      msgID,
			Subject: "Test Message",
			From:    []domain.EmailParticipant{{Email: "test@example.com"}},
			Starred: false,
		}, nil
	}
	client.UpdateMessageFunc = func(ctx context.Context, grantID, msgID string, req *domain.UpdateMessageRequest) (*domain.Message, error) {
		return &domain.Message{
			ID:      msgID,
			Subject: "Test Message",
			Starred: *req.Starred,
		}, nil
	}

	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "user@example.com", "google")

	// Create message detail model
	model := NewMessageDetail(global, "msg-123")
	model.global.SetWindowSize(120, 40)

	tm := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(120, 40))
	defer func() {
		_ = tm.Quit()
	}()

	// Wait for thread/messages to load
	teatest.WaitFor(
		t, tm.Output(),
		func(bts []byte) bool {
			return !model.loading && (model.thread != nil || model.message != nil)
		},
		teatest.WithCheckInterval(time.Millisecond*10),
		teatest.WithDuration(time.Second*2),
	)

	// Get the active message (from thread or single message)
	msg := model.getActiveMessage()
	if msg == nil {
		t.Fatal("No message loaded")
	}

	// Verify initial state
	if msg.Starred {
		t.Error("Message should not be starred initially")
	}

	// Verify help text includes star action
	help := model.buildHelpText()
	if !strings.Contains(help, "star") {
		t.Error("Help text should mention star action")
	}
}

// TestIntegration_DeleteMessage tests the delete message action with confirmation.
func TestIntegration_DeleteMessage(t *testing.T) {
	client := nylas.NewMockClient()
	client.GetThreadFunc = func(ctx context.Context, grantID, threadID string) (*domain.Thread, error) {
		return &domain.Thread{
			ID:      threadID,
			Subject: "Message to Delete",
			LatestDraftOrMessage: domain.Message{
				ID:      "msg-to-delete",
				Subject: "Message to Delete",
				From:    []domain.EmailParticipant{{Email: "test@example.com"}},
			},
			MessageIDs:            []string{"msg-to-delete"},
			Participants:          []domain.EmailParticipant{{Email: "test@example.com"}},
			LatestMessageRecvDate: time.Now(),
		}, nil
	}
	client.GetMessageFunc = func(ctx context.Context, grantID, msgID string) (*domain.Message, error) {
		return &domain.Message{
			ID:      msgID,
			Subject: "Message to Delete",
			From:    []domain.EmailParticipant{{Email: "test@example.com"}},
		}, nil
	}
	client.DeleteMessageFunc = func(ctx context.Context, grantID, msgID string) error {
		return nil
	}

	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "user@example.com", "google")

	// Create message detail model
	model := NewMessageDetail(global, "msg-to-delete")
	model.global.SetWindowSize(120, 40)

	tm := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(120, 40))
	defer func() {
		_ = tm.Quit()
	}()

	// Wait for thread/messages to load
	teatest.WaitFor(
		t, tm.Output(),
		func(bts []byte) bool {
			return !model.loading && (model.thread != nil || model.message != nil)
		},
		teatest.WithCheckInterval(time.Millisecond*10),
		teatest.WithDuration(time.Second*2),
	)

	// Get the active message (from thread or single message)
	msg := model.getActiveMessage()
	if msg == nil {
		t.Fatal("No message loaded")
	}

	if msg.Subject != "Message to Delete" {
		t.Errorf("Expected subject 'Message to Delete', got '%s'", msg.Subject)
	}

	// Verify help text includes delete action
	help := model.buildHelpText()
	if !strings.Contains(help, "delete") {
		t.Error("Help text should mention delete action")
	}

	// Verify pendingConfirmation is nil initially
	if model.pendingConfirmation != nil {
		t.Error("Confirmation should not be pending initially")
	}
}
