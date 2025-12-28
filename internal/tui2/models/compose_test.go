package models

import (
	"strings"
	"testing"
	"time"

	"github.com/mqasimca/nylas/internal/adapters/keyring"
	"github.com/mqasimca/nylas/internal/adapters/nylas"
	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/tui2/state"
)

func TestNewCompose(t *testing.T) {
	client := nylas.NewMockClient()
	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "test@example.com", "google")

	data := ComposeData{
		Mode: ComposeModeNew,
	}

	c := NewCompose(global, data)

	if c == nil {
		t.Fatal("NewCompose returned nil")
	}

	if c.global != global {
		t.Error("global state not set correctly")
	}

	if c.mode != ComposeModeNew {
		t.Errorf("expected mode %v, got %v", ComposeModeNew, c.mode)
	}

	if !c.autosaveEnabled {
		t.Error("autosave should be enabled by default")
	}

	if c.autosaveInterval != 30*time.Second {
		t.Errorf("expected autosave interval 30s, got %v", c.autosaveInterval)
	}
}

func TestCompose_ParseRecipients(t *testing.T) {
	client := nylas.NewMockClient()
	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "test@example.com", "google")

	c := NewCompose(global, ComposeData{Mode: ComposeModeNew})

	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "single email",
			input:    "user@example.com",
			expected: 1,
		},
		{
			name:     "name and email",
			input:    "John Doe <john@example.com>",
			expected: 1,
		},
		{
			name:     "multiple emails",
			input:    "user1@example.com, user2@example.com",
			expected: 2,
		},
		{
			name:     "mixed format",
			input:    "John Doe <john@example.com>, jane@example.com",
			expected: 2,
		},
		{
			name:     "empty string",
			input:    "",
			expected: 0,
		},
		{
			name:     "invalid format",
			input:    "not an email",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recipients := c.parseRecipients(tt.input)
			if len(recipients) != tt.expected {
				t.Errorf("expected %d recipients, got %d", tt.expected, len(recipients))
			}
		})
	}
}

func TestCompose_ParseRecipients_Details(t *testing.T) {
	client := nylas.NewMockClient()
	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "test@example.com", "google")

	c := NewCompose(global, ComposeData{Mode: ComposeModeNew})

	// Test name extraction
	recipients := c.parseRecipients("John Doe <john@example.com>")
	if len(recipients) != 1 {
		t.Fatal("expected 1 recipient")
	}
	if recipients[0].Name != "John Doe" {
		t.Errorf("expected name 'John Doe', got '%s'", recipients[0].Name)
	}
	if recipients[0].Email != "john@example.com" {
		t.Errorf("expected email 'john@example.com', got '%s'", recipients[0].Email)
	}

	// Test email only
	recipients = c.parseRecipients("jane@example.com")
	if len(recipients) != 1 {
		t.Fatal("expected 1 recipient")
	}
	if recipients[0].Name != "" {
		t.Errorf("expected empty name, got '%s'", recipients[0].Name)
	}
	if recipients[0].Email != "jane@example.com" {
		t.Errorf("expected email 'jane@example.com', got '%s'", recipients[0].Email)
	}
}

func TestCompose_Validate(t *testing.T) {
	client := nylas.NewMockClient()
	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "test@example.com", "google")

	tests := []struct {
		name      string
		toValue   string
		wantValid bool
	}{
		{
			name:      "valid email",
			toValue:   "user@example.com",
			wantValid: true,
		},
		{
			name:      "valid name and email",
			toValue:   "John Doe <john@example.com>",
			wantValid: true,
		},
		{
			name:      "empty To field",
			toValue:   "",
			wantValid: false,
		},
		{
			name:      "invalid email format",
			toValue:   "not an email",
			wantValid: false,
		},
		{
			name:      "whitespace only",
			toValue:   "   ",
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCompose(global, ComposeData{Mode: ComposeModeNew})
			c.toInput.SetValue(tt.toValue)

			valid := c.validate()

			if valid != tt.wantValid {
				t.Errorf("expected valid=%v, got %v (errors: %v)", tt.wantValid, valid, c.validationErrors)
			}
		})
	}
}

func TestCompose_ComputeContentHash(t *testing.T) {
	client := nylas.NewMockClient()
	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "test@example.com", "google")

	c := NewCompose(global, ComposeData{Mode: ComposeModeNew})

	// Initial hash
	hash1 := c.computeContentHash()
	if hash1 == "" {
		t.Error("hash should not be empty")
	}

	// Same content should produce same hash
	hash2 := c.computeContentHash()
	if hash1 != hash2 {
		t.Error("same content should produce same hash")
	}

	// Different content should produce different hash
	c.toInput.SetValue("test@example.com")
	hash3 := c.computeContentHash()
	if hash1 == hash3 {
		t.Error("different content should produce different hash")
	}
}

func TestCompose_BuildQuotedBody(t *testing.T) {
	client := nylas.NewMockClient()
	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "test@example.com", "google")

	c := NewCompose(global, ComposeData{Mode: ComposeModeNew})

	msg := &domain.Message{
		From: []domain.EmailParticipant{
			{Name: "John Doe", Email: "john@example.com"},
		},
		Date:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		Body:    "This is the original message.\nSecond line.",
		Snippet: "Original snippet",
	}

	quoted := c.buildQuotedBody(msg)

	// Check for quote header
	if !strings.Contains(quoted, "On") {
		t.Error("quoted body should contain 'On'")
	}
	if !strings.Contains(quoted, "wrote:") {
		t.Error("quoted body should contain 'wrote:'")
	}

	// Check for quoted lines
	if !strings.Contains(quoted, "> This is the original message.") {
		t.Error("quoted body should contain quoted first line")
	}
	if !strings.Contains(quoted, "> Second line.") {
		t.Error("quoted body should contain quoted second line")
	}
}

func TestCompose_BuildForwardedBody(t *testing.T) {
	client := nylas.NewMockClient()
	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "test@example.com", "google")

	c := NewCompose(global, ComposeData{Mode: ComposeModeNew})

	msg := &domain.Message{
		From: []domain.EmailParticipant{
			{Name: "John Doe", Email: "john@example.com"},
		},
		To: []domain.EmailParticipant{
			{Name: "Jane Smith", Email: "jane@example.com"},
		},
		Date:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		Subject: "Test Subject",
		Body:    "Original message body",
	}

	forwarded := c.buildForwardedBody(msg)

	// Check for forward header
	if !strings.Contains(forwarded, "---------- Forwarded message ---------") {
		t.Error("forwarded body should contain forward header")
	}

	// Check for From field
	if !strings.Contains(forwarded, "From: John Doe <john@example.com>") {
		t.Error("forwarded body should contain From field")
	}

	// Check for To field
	if !strings.Contains(forwarded, "To: Jane Smith <jane@example.com>") {
		t.Error("forwarded body should contain To field")
	}

	// Check for Subject field
	if !strings.Contains(forwarded, "Subject: Test Subject") {
		t.Error("forwarded body should contain Subject field")
	}

	// Check for body
	if !strings.Contains(forwarded, "Original message body") {
		t.Error("forwarded body should contain original body")
	}
}

func TestCompose_PrefillReply(t *testing.T) {
	client := nylas.NewMockClient()
	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "test@example.com", "google")

	originalMsg := &domain.Message{
		ID:      "msg123",
		Subject: "Original Subject",
		From: []domain.EmailParticipant{
			{Name: "John Doe", Email: "john@example.com"},
		},
		Body: "Original message body",
		Date: time.Now(),
	}

	data := ComposeData{
		Mode:    ComposeModeReply,
		Message: originalMsg,
	}

	c := NewCompose(global, data)

	// Check subject has Re: prefix
	subject := c.subjectInput.Value()
	if !strings.HasPrefix(subject, "Re: ") {
		t.Errorf("reply subject should start with 'Re:', got '%s'", subject)
	}

	// Check To field is set to original sender
	toValue := c.toInput.Value()
	if !strings.Contains(toValue, "john@example.com") {
		t.Errorf("To field should contain original sender, got '%s'", toValue)
	}

	// Check body contains quoted content
	bodyValue := c.bodyInput.Value()
	if !strings.Contains(bodyValue, ">") {
		t.Error("reply body should contain quoted lines")
	}
}

func TestCompose_PrefillReplyAll(t *testing.T) {
	client := nylas.NewMockClient()
	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "test@example.com", "google")

	originalMsg := &domain.Message{
		ID:      "msg123",
		Subject: "Original Subject",
		From: []domain.EmailParticipant{
			{Name: "John Doe", Email: "john@example.com"},
		},
		To: []domain.EmailParticipant{
			{Email: "user1@example.com"},
		},
		Cc: []domain.EmailParticipant{
			{Email: "user2@example.com"},
		},
		Body: "Original message body",
		Date: time.Now(),
	}

	data := ComposeData{
		Mode:    ComposeModeReplyAll,
		Message: originalMsg,
	}

	c := NewCompose(global, data)

	// Check Cc field is shown and populated
	if !c.showCc {
		t.Error("Cc field should be shown for reply all")
	}

	ccValue := c.ccInput.Value()
	if !strings.Contains(ccValue, "user1@example.com") && !strings.Contains(ccValue, "user2@example.com") {
		t.Errorf("Cc field should contain recipients, got '%s'", ccValue)
	}
}

func TestCompose_PrefillForward(t *testing.T) {
	client := nylas.NewMockClient()
	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "test@example.com", "google")

	originalMsg := &domain.Message{
		ID:      "msg123",
		Subject: "Original Subject",
		From: []domain.EmailParticipant{
			{Name: "John Doe", Email: "john@example.com"},
		},
		Body: "Original message body",
		Date: time.Now(),
	}

	data := ComposeData{
		Mode:    ComposeModeForward,
		Message: originalMsg,
	}

	c := NewCompose(global, data)

	// Check subject has Fwd: prefix
	subject := c.subjectInput.Value()
	if !strings.HasPrefix(subject, "Fwd: ") {
		t.Errorf("forward subject should start with 'Fwd:', got '%s'", subject)
	}

	// Check To field is empty (user needs to specify)
	toValue := c.toInput.Value()
	if toValue != "" {
		t.Errorf("To field should be empty for forward, got '%s'", toValue)
	}

	// Check body contains forwarded content
	bodyValue := c.bodyInput.Value()
	if !strings.Contains(bodyValue, "---------- Forwarded message ---------") {
		t.Error("forward body should contain forward header")
	}
}

func TestCompose_FormatParticipant(t *testing.T) {
	client := nylas.NewMockClient()
	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "test@example.com", "google")

	c := NewCompose(global, ComposeData{Mode: ComposeModeNew})

	tests := []struct {
		name     string
		p        domain.EmailParticipant
		expected string
	}{
		{
			name:     "with name",
			p:        domain.EmailParticipant{Name: "John Doe", Email: "john@example.com"},
			expected: "John Doe <john@example.com>",
		},
		{
			name:     "without name",
			p:        domain.EmailParticipant{Email: "jane@example.com"},
			expected: "jane@example.com",
		},
		{
			name:     "empty name",
			p:        domain.EmailParticipant{Name: "", Email: "test@example.com"},
			expected: "test@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := c.formatParticipant(tt.p)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestCompose_StripExistingQuotes(t *testing.T) {
	client := nylas.NewMockClient()
	grantStore := keyring.NewGrantStore(keyring.NewMockSecretStore())
	global := state.NewGlobalState(client, grantStore, "grant123", "test@example.com", "google")

	c := NewCompose(global, ComposeData{Mode: ComposeModeNew})

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no quotes",
			input:    "This is a plain message",
			expected: "This is a plain message",
		},
		{
			name:     "simple quote",
			input:    "New text\n> Quoted line",
			expected: "New text",
		},
		{
			name:     "attribution line",
			input:    "New text\nOn Dec 27, 2025 at 3:05 PM, John wrote:\n> Old text",
			expected: "New text",
		},
		{
			name: "nested quotes",
			input: `Reply 3
On Dec 27, 2025 at 3:08 PM, User wrote:
> Reply 2
> On Dec 27, 2025 at 3:05 PM, User wrote:
> > Reply 1
> > On Dec 27, 2025 at 3:00 PM, User wrote:
> > > Original message`,
			expected: "Reply 3",
		},
		{
			name: "mixed content",
			input: `New reply text here
Some more text
On Dec 27, 2025 at 3:05 PM, John Doe wrote:
> This was the previous reply
> With multiple lines
> > And even older quotes`,
			expected: "New reply text here\nSome more text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := c.stripExistingQuotes(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
