package models

import (
	"context"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/mqasimca/nylas/internal/adapters/keyring"
	"github.com/mqasimca/nylas/internal/adapters/nylas"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/tui2/components"
	"github.com/mqasimca/nylas/internal/tui2/state"
)

// TestIntegration_MessageList_FetchAndDisplay tests the full workflow of fetching and displaying messages.
func TestIntegration_MessageList_FetchAndDisplay(t *testing.T) {
	// Create mock client with test data
	client := nylas.NewMockClient()
	client.GetMessagesFunc = func(ctx context.Context, grantID string, limit int) ([]domain.Message, error) {
		return []domain.Message{
			{
				ID:      "msg1",
				Subject: "Welcome to Nylas",
				From:    []domain.EmailParticipant{{Name: "Nylas Team", Email: "team@nylas.com"}},
				To:      []domain.EmailParticipant{{Email: "user@example.com"}},
				Date:    time.Now().Add(-1 * time.Hour),
				Snippet: "Thanks for using Nylas!",
			},
			{
				ID:      "msg2",
				Subject: "Important Update",
				From:    []domain.EmailParticipant{{Name: "Admin", Email: "admin@example.com"}},
				To:      []domain.EmailParticipant{{Email: "user@example.com"}},
				Date:    time.Now().Add(-24 * time.Hour),
				Snippet: "Please review the latest changes.",
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
	defer tm.Quit()

	// Wait for initial messages to load
	teatest.WaitFor(
		t, tm.Output(),
		func(bts []byte) bool {
			return len(model.messages) > 0
		},
		teatest.WithCheckInterval(time.Millisecond*10),
		teatest.WithDuration(time.Second*3),
	)

	// Verify messages were loaded
	if len(model.messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(model.messages))
	}

	if !client.GetMessagesCalled {
		t.Error("GetMessages should have been called")
	}
}

// TestIntegration_MessageList_NavigatePanes tests navigating between panes with Tab.
func TestIntegration_MessageList_NavigatePanes(t *testing.T) {
	client := nylas.NewMockClient()
	client.GetMessagesFunc = func(ctx context.Context, grantID string, limit int) ([]domain.Message, error) {
		return []domain.Message{
			{ID: "msg1", Subject: "Test 1", From: []domain.EmailParticipant{{Email: "test@example.com"}}},
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
	defer tm.Quit()

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

// TestIntegration_MessageList_SelectMessage tests selecting a message with Enter key.
func TestIntegration_MessageList_SelectMessage(t *testing.T) {
	client := nylas.NewMockClient()
	client.GetMessagesFunc = func(ctx context.Context, grantID string, limit int) ([]domain.Message, error) {
		return []domain.Message{
			{
				ID:      "msg-important",
				Subject: "Important: Read This",
				From:    []domain.EmailParticipant{{Name: "Boss", Email: "boss@example.com"}},
				Date:    time.Now(),
			},
		}, nil
	}

	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "user@example.com", "google")

	model := NewMessageList(global)
	model.global.SetWindowSize(120, 40)

	tm := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(120, 40))
	defer tm.Quit()

	// Wait for messages to load
	teatest.WaitFor(
		t, tm.Output(),
		func(bts []byte) bool {
			return len(model.messages) > 0
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
	if len(model.messages) == 0 {
		t.Error("Messages should still be loaded after selection")
	}

	if model.layout.GetFocused() != components.MessagePane {
		t.Error("Focus should remain on MessagePane after Enter")
	}
}

// TestIntegration_MessageList_FilterByFolder tests filtering messages by folder.
func TestIntegration_MessageList_FilterByFolder(t *testing.T) {
	client := nylas.NewMockClient()

	// Initial messages
	client.GetMessagesFunc = func(ctx context.Context, grantID string, limit int) ([]domain.Message, error) {
		return []domain.Message{
			{ID: "msg1", Subject: "Inbox Message"},
		}, nil
	}

	// Folder list
	client.GetFoldersFunc = func(ctx context.Context, grantID string) ([]domain.Folder, error) {
		return []domain.Folder{
			{ID: "inbox", Name: "Inbox", TotalCount: 10},
			{ID: "sent", Name: "Sent", TotalCount: 5},
		}, nil
	}

	// Filtered messages
	client.GetMessagesWithParamsFunc = func(ctx context.Context, grantID string, params *domain.MessageQueryParams) ([]domain.Message, error) {
		if len(params.In) > 0 && params.In[0] == "sent" {
			return []domain.Message{
				{ID: "msg-sent-1", Subject: "Sent Message 1"},
				{ID: "msg-sent-2", Subject: "Sent Message 2"},
			}, nil
		}
		return []domain.Message{}, nil
	}

	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "user@example.com", "google")

	model := NewMessageList(global)
	model.global.SetWindowSize(120, 40)

	tm := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(120, 40))
	defer tm.Quit()

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
			return model.selectedFolderID == "sent" && len(model.messages) > 1
		},
		teatest.WithCheckInterval(time.Millisecond*10),
		teatest.WithDuration(time.Second*2),
	)

	if !client.GetMessagesWithParamsCalled {
		t.Error("GetMessagesWithParams should have been called for folder filter")
	}

	if len(model.messages) != 2 {
		t.Errorf("Expected 2 sent messages, got %d", len(model.messages))
	}
}

// TestIntegration_MessageList_Refresh tests the refresh functionality.
func TestIntegration_MessageList_Refresh(t *testing.T) {
	callCount := 0
	client := nylas.NewMockClient()
	client.GetMessagesFunc = func(ctx context.Context, grantID string, limit int) ([]domain.Message, error) {
		callCount++
		return []domain.Message{
			{ID: "msg1", Subject: "Message 1"},
		}, nil
	}

	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "user@example.com", "google")

	model := NewMessageList(global)
	model.global.SetWindowSize(120, 40)

	tm := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(120, 40))
	defer tm.Quit()

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

	// Press 'r' to refresh
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})

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
		t.Errorf("Refresh should trigger another GetMessages call, initial=%d, current=%d", initialCalls, callCount)
	}
}

// TestIntegration_MessageList_BackNavigation tests going back with Esc.
func TestIntegration_MessageList_BackNavigation(t *testing.T) {
	client := nylas.NewMockClient()
	client.GetMessagesFunc = func(ctx context.Context, grantID string, limit int) ([]domain.Message, error) {
		return []domain.Message{}, nil
	}

	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "user@example.com", "google")

	model := NewMessageList(global)
	model.global.SetWindowSize(120, 40)

	tm := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(120, 40))
	defer tm.Quit()

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
