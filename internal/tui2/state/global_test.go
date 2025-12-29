package state

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
)

func TestNewGlobalState(t *testing.T) {
	tests := []struct {
		name     string
		grantID  string
		email    string
		provider string
	}{
		{
			name:     "creates state with valid values",
			grantID:  "grant-123",
			email:    "test@example.com",
			provider: "google",
		},
		{
			name:     "creates state with empty values",
			grantID:  "",
			email:    "",
			provider: "",
		},
		{
			name:     "creates state with special characters",
			grantID:  "grant-abc-123-xyz",
			email:    "user+tag@example.com",
			provider: "microsoft-365",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewGlobalState(nil, nil, tt.grantID, tt.email, tt.provider)

			if state == nil {
				t.Fatal("expected non-nil state")
			}

			// Verify all fields are set correctly
			if state.Client != nil {
				t.Error("client should be nil")
			}
			if state.GrantStore != nil {
				t.Error("grant store should be nil")
			}
			if state.GrantID != tt.grantID {
				t.Errorf("expected GrantID %q, got %q", tt.grantID, state.GrantID)
			}
			if state.Email != tt.email {
				t.Errorf("expected Email %q, got %q", tt.email, state.Email)
			}
			if state.Provider != tt.provider {
				t.Errorf("expected Provider %q, got %q", tt.provider, state.Provider)
			}

			// Verify default values
			if state.Theme != "k9s" {
				t.Errorf("expected Theme 'k9s', got %q", state.Theme)
			}
			if state.WindowSize.Width != 80 {
				t.Errorf("expected WindowSize.Width 80, got %d", state.WindowSize.Width)
			}
			if state.WindowSize.Height != 24 {
				t.Errorf("expected WindowSize.Height 24, got %d", state.WindowSize.Height)
			}
			if state.StatusMessage != "" {
				t.Errorf("expected empty StatusMessage, got %q", state.StatusMessage)
			}
			if state.StatusLevel != 0 {
				t.Errorf("expected StatusLevel 0, got %d", state.StatusLevel)
			}
			if state.OfflineMode {
				t.Error("expected OfflineMode false by default")
			}
			if state.RateLimiter == nil {
				t.Error("expected non-nil RateLimiter")
			}
		})
	}
}

func TestGlobalState_SetWindowSize(t *testing.T) {
	state := NewGlobalState(nil, nil, "grant", "email", "provider")

	tests := []struct {
		name   string
		width  int
		height int
	}{
		{
			name:   "set standard size",
			width:  120,
			height: 40,
		},
		{
			name:   "set small size",
			width:  20,
			height: 10,
		},
		{
			name:   "set large size",
			width:  200,
			height: 100,
		},
		{
			name:   "set zero size",
			width:  0,
			height: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state.SetWindowSize(tt.width, tt.height)

			if state.WindowSize.Width != tt.width {
				t.Errorf("expected Width %d, got %d", tt.width, state.WindowSize.Width)
			}
			if state.WindowSize.Height != tt.height {
				t.Errorf("expected Height %d, got %d", tt.height, state.WindowSize.Height)
			}
		})
	}
}

func TestGlobalState_SetStatus(t *testing.T) {
	state := NewGlobalState(nil, nil, "grant", "email", "provider")

	tests := []struct {
		name    string
		message string
		level   int
	}{
		{
			name:    "info message",
			message: "Operation completed successfully",
			level:   0,
		},
		{
			name:    "warning message",
			message: "Warning: Rate limit approaching",
			level:   1,
		},
		{
			name:    "error message",
			message: "Error: Failed to load messages",
			level:   2,
		},
		{
			name:    "empty message",
			message: "",
			level:   0,
		},
		{
			name:    "long message",
			message: "This is a very long status message that might need to be truncated in the UI but should be stored fully",
			level:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state.SetStatus(tt.message, tt.level)

			if state.StatusMessage != tt.message {
				t.Errorf("expected StatusMessage %q, got %q", tt.message, state.StatusMessage)
			}
			if state.StatusLevel != tt.level {
				t.Errorf("expected StatusLevel %d, got %d", tt.level, state.StatusLevel)
			}
		})
	}
}

func TestGlobalState_ClearStatus(t *testing.T) {
	state := NewGlobalState(nil, nil, "grant", "email", "provider")

	// Set a status first
	state.SetStatus("Test message", 2)

	// Verify it's set
	if state.StatusMessage != "Test message" {
		t.Error("status message not set before clear")
	}
	if state.StatusLevel != 2 {
		t.Error("status level not set before clear")
	}

	// Clear status
	state.ClearStatus()

	// Verify it's cleared
	if state.StatusMessage != "" {
		t.Errorf("expected empty StatusMessage after clear, got %q", state.StatusMessage)
	}
	if state.StatusLevel != 0 {
		t.Errorf("expected StatusLevel 0 after clear, got %d", state.StatusLevel)
	}
}

func TestGlobalState_StatusLifecycle(t *testing.T) {
	state := NewGlobalState(nil, nil, "grant", "email", "provider")

	// Should start empty
	if state.StatusMessage != "" || state.StatusLevel != 0 {
		t.Error("status should start empty")
	}

	// Set first status
	state.SetStatus("Loading...", 0)
	if state.StatusMessage != "Loading..." || state.StatusLevel != 0 {
		t.Error("first status not set correctly")
	}

	// Update status
	state.SetStatus("Error occurred", 2)
	if state.StatusMessage != "Error occurred" || state.StatusLevel != 2 {
		t.Error("status not updated correctly")
	}

	// Clear status
	state.ClearStatus()
	if state.StatusMessage != "" || state.StatusLevel != 0 {
		t.Error("status not cleared correctly")
	}

	// Set status again after clear
	state.SetStatus("Success!", 0)
	if state.StatusMessage != "Success!" || state.StatusLevel != 0 {
		t.Error("status not set correctly after clear")
	}
}

func TestGlobalState_RateLimiter(t *testing.T) {
	state := NewGlobalState(nil, nil, "grant", "email", "provider")

	if state.RateLimiter == nil {
		t.Fatal("expected non-nil RateLimiter")
	}

	// Test that rate limiter allows first call immediately
	allowed := state.RateLimiter.TryWait()
	if !allowed {
		t.Error("rate limiter should allow first call")
	}

	// Second call should be rate limited
	allowed = state.RateLimiter.TryWait()
	if allowed {
		t.Error("rate limiter should block second immediate call")
	}

	// Wait for rate limit to reset
	time.Sleep(600 * time.Millisecond)
	allowed = state.RateLimiter.TryWait()
	if !allowed {
		t.Error("rate limiter should allow call after waiting")
	}
}

func TestGlobalState_FieldMutability(t *testing.T) {
	state := NewGlobalState(nil, nil, "grant", "email", "provider")

	// Test that fields can be modified directly
	t.Run("modify theme", func(t *testing.T) {
		state.Theme = "custom"
		if state.Theme != "custom" {
			t.Error("theme not modified")
		}
	})

	t.Run("modify grant ID", func(t *testing.T) {
		state.GrantID = "new-grant-123"
		if state.GrantID != "new-grant-123" {
			t.Error("grant ID not modified")
		}
	})

	t.Run("modify email", func(t *testing.T) {
		state.Email = "new@example.com"
		if state.Email != "new@example.com" {
			t.Error("email not modified")
		}
	})

	t.Run("modify provider", func(t *testing.T) {
		state.Provider = "outlook"
		if state.Provider != "outlook" {
			t.Error("provider not modified")
		}
	})

	t.Run("modify offline mode", func(t *testing.T) {
		state.OfflineMode = true
		if !state.OfflineMode {
			t.Error("offline mode not modified")
		}
	})
}

func TestGlobalState_WindowSizeMsg(t *testing.T) {
	state := NewGlobalState(nil, nil, "grant", "email", "provider")

	// Test that WindowSize is a proper tea.WindowSizeMsg
	var _ tea.WindowSizeMsg = state.WindowSize

	// Test that SetWindowSize creates a proper WindowSizeMsg
	state.SetWindowSize(100, 50)
	if state.WindowSize.Width != 100 || state.WindowSize.Height != 50 {
		t.Error("WindowSizeMsg not created correctly")
	}
}

func TestGlobalState_NilClientHandling(t *testing.T) {
	// Test that state can be created with nil client (for testing scenarios)
	state := NewGlobalState(nil, nil, "grant", "email", "provider")

	if state == nil {
		t.Fatal("state should be created even with nil client")
	}
	if state.Client != nil {
		t.Error("client should be nil")
	}
	if state.GrantStore != nil {
		t.Error("grant store should be nil")
	}

	// Other fields should still be initialized
	if state.Theme != "k9s" {
		t.Error("theme should still be initialized")
	}
	if state.RateLimiter == nil {
		t.Error("rate limiter should still be initialized")
	}
}

func TestGlobalState_ConcurrentAccess(t *testing.T) {
	state := NewGlobalState(nil, nil, "grant", "email", "provider")

	// Test concurrent status updates (basic race detection)
	done := make(chan bool, 2)

	go func() {
		for i := 0; i < 100; i++ {
			state.SetStatus("Message A", 0)
			state.ClearStatus()
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			state.SetStatus("Message B", 1)
			state.ClearStatus()
		}
		done <- true
	}()

	<-done
	<-done

	// Should complete without panic
	// Final state should be cleared (but concurrent access isn't synchronized)
	// Just checking it doesn't panic - state values may vary due to race conditions
	_ = state.StatusMessage
	_ = state.StatusLevel
}
