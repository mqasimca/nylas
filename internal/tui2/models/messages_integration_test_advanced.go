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

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/mqasimca/nylas/internal/adapters/keyring"
	"github.com/mqasimca/nylas/internal/adapters/nylas"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/tui2/state"
)

// TestIntegration_MessageList_FetchAndDisplay tests the full workflow of fetching and displaying messages.
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
