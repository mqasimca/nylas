//go:build !integration

package ai

import (
	"testing"
	"time"

	"github.com/mqasimca/nylas/internal/domain"
	"github.com/stretchr/testify/assert"
)

// Note: Tests that require full mock interface implementations are in
// the integration test file. These unit tests focus on pure functions
// that can be tested without mocking external dependencies.

func TestNewEmailAnalyzer(t *testing.T) {
	t.Run("creates analyzer with nil dependencies", func(t *testing.T) {
		analyzer := NewEmailAnalyzer(nil, nil)

		assert.NotNil(t, analyzer)
		assert.Nil(t, analyzer.nylasClient)
		assert.Nil(t, analyzer.llmRouter)
	})
}

func TestEmailAnalyzer_BuildThreadContext(t *testing.T) {
	analyzer := &EmailAnalyzer{}

	tests := []struct {
		name     string
		thread   *domain.Thread
		messages []domain.Message
		wantIn   []string // Substrings expected in the output
	}{
		{
			name: "builds context with participants",
			thread: &domain.Thread{
				Subject: "Project Discussion",
				Participants: []domain.EmailParticipant{
					{Name: "Alice", Email: "alice@example.com"},
					{Email: "bob@example.com"},
				},
			},
			messages: []domain.Message{
				{
					From:    []domain.EmailParticipant{{Name: "Alice", Email: "alice@example.com"}},
					Body:    "Let's discuss the project.",
					Date:    time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
					Subject: "Project Discussion",
				},
			},
			wantIn: []string{
				"Email Thread: Project Discussion",
				"Participants: 2",
				"Messages: 1",
				"Alice <alice@example.com>",
				"bob@example.com",
				"Let's discuss the project",
			},
		},
		{
			name: "truncates long message bodies",
			thread: &domain.Thread{
				Subject:      "Long Email",
				Participants: []domain.EmailParticipant{},
			},
			messages: []domain.Message{
				{
					From: []domain.EmailParticipant{{Email: "sender@example.com"}},
					Body: string(make([]byte, 600)), // 600 chars
					Date: time.Now(),
				},
			},
			wantIn: []string{
				"...", // Should be truncated
			},
		},
		{
			name: "handles sender with no name",
			thread: &domain.Thread{
				Subject:      "Test",
				Participants: []domain.EmailParticipant{},
			},
			messages: []domain.Message{
				{
					From: []domain.EmailParticipant{{Email: "unknown@example.com"}},
					Body: "Message body",
					Date: time.Now(),
				},
			},
			wantIn: []string{
				"unknown@example.com",
			},
		},
		{
			name: "handles unknown sender",
			thread: &domain.Thread{
				Subject:      "No Sender",
				Participants: []domain.EmailParticipant{},
			},
			messages: []domain.Message{
				{
					From: []domain.EmailParticipant{},
					Body: "Anonymous message",
					Date: time.Now(),
				},
			},
			wantIn: []string{
				"Unknown",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.buildThreadContext(tt.thread, tt.messages)

			for _, want := range tt.wantIn {
				assert.Contains(t, result, want)
			}
		})
	}
}

func TestEmailAnalyzer_AnalyzeParticipants(t *testing.T) {
	analyzer := &EmailAnalyzer{}

	tests := []struct {
		name       string
		thread     *domain.Thread
		messages   []domain.Message
		wantCounts map[string]int // email -> expected message count
		wantReq    []string       // emails expected to be required
	}{
		{
			name: "counts messages per participant",
			thread: &domain.Thread{
				Participants: []domain.EmailParticipant{
					{Name: "Alice", Email: "alice@example.com"},
					{Name: "Bob", Email: "bob@example.com"},
				},
			},
			messages: []domain.Message{
				{From: []domain.EmailParticipant{{Email: "alice@example.com"}}, Body: "", Date: time.Now()},
				{From: []domain.EmailParticipant{{Email: "alice@example.com"}}, Body: "", Date: time.Now()},
				{From: []domain.EmailParticipant{{Email: "bob@example.com"}}, Body: "", Date: time.Now()},
			},
			wantCounts: map[string]int{
				"alice@example.com": 2,
				"bob@example.com":   1,
			},
			wantReq: []string{"alice@example.com"}, // 2 messages = required
		},
		{
			name: "detects mentions in body",
			thread: &domain.Thread{
				Participants: []domain.EmailParticipant{
					{Email: "mentioned@example.com"},
				},
			},
			messages: []domain.Message{
				{From: []domain.EmailParticipant{{Email: "other@example.com"}}, Body: "CC mentioned@example.com for this", Date: time.Now()},
				{From: []domain.EmailParticipant{{Email: "other@example.com"}}, Body: "Also mentioned@example.com", Date: time.Now()},
			},
			wantReq: []string{"mentioned@example.com"}, // Mentioned multiple times = required
		},
		{
			name: "high involvement for active senders",
			thread: &domain.Thread{
				Participants: []domain.EmailParticipant{
					{Email: "active@example.com"},
				},
			},
			messages: []domain.Message{
				{From: []domain.EmailParticipant{{Email: "active@example.com"}}, Body: "", Date: time.Now()},
				{From: []domain.EmailParticipant{{Email: "active@example.com"}}, Body: "", Date: time.Now()},
				{From: []domain.EmailParticipant{{Email: "active@example.com"}}, Body: "", Date: time.Now()},
			},
			wantCounts: map[string]int{"active@example.com": 3},
			wantReq:    []string{"active@example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.analyzeParticipants(tt.thread, tt.messages)

			// Check message counts
			for _, p := range result {
				if expectedCount, ok := tt.wantCounts[p.Email]; ok {
					assert.Equal(t, expectedCount, p.MessageCount, "message count for %s", p.Email)
				}
			}

			// Check required status
			for _, p := range result {
				for _, reqEmail := range tt.wantReq {
					if p.Email == reqEmail {
						assert.True(t, p.Required, "expected %s to be required", reqEmail)
					}
				}
			}
		})
	}
}

func TestEmailAnalyzer_DetectUrgencyIndicators(t *testing.T) {
	analyzer := &EmailAnalyzer{}

	tests := []struct {
		name     string
		messages []domain.Message
		wantIn   []string // Substrings expected in indicators
		wantLen  int      // Expected number of indicators (0 = don't check)
	}{
		{
			name: "detects urgent keywords in body",
			messages: []domain.Message{
				{Body: "This is URGENT, please respond ASAP", Subject: "Normal subject", Date: time.Now()},
			},
			wantIn: []string{"urgent", "asap"},
		},
		{
			name: "detects urgent keywords in subject",
			messages: []domain.Message{
				{Body: "Normal body", Subject: "CRITICAL: System down", Date: time.Now()},
			},
			wantIn: []string{"critical"},
		},
		{
			name: "detects deadline keywords",
			messages: []domain.Message{
				{Body: "Need this by today, deadline approaching", Subject: "Request", Date: time.Now()},
			},
			wantIn: []string{"today", "deadline"},
		},
		{
			name: "detects high message frequency",
			messages: func() []domain.Message {
				now := time.Now()
				msgs := make([]domain.Message, 6)
				for i := 0; i < 6; i++ {
					msgs[i] = domain.Message{
						From: []domain.EmailParticipant{{Email: "test@example.com"}},
						Date: now.Add(time.Duration(i) * time.Hour),
						Body: "Message",
					}
				}
				return msgs
			}(),
			wantIn: []string{"high activity"},
		},
		{
			name: "detects broad reach with many participants",
			messages: func() []domain.Message {
				msgs := make([]domain.Message, 6)
				for i := 0; i < 6; i++ {
					msgs[i] = domain.Message{
						From: []domain.EmailParticipant{{Email: "user" + string(rune('a'+i)) + "@example.com"}},
						Date: time.Now(),
						Body: "Message",
					}
				}
				return msgs
			}(),
			wantIn: []string{"participants", "broad reach"},
		},
		{
			name: "no indicators for calm thread",
			messages: []domain.Message{
				{Body: "Just checking in on the project", Subject: "Update", Date: time.Now()},
			},
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.detectUrgencyIndicators(tt.messages)

			if tt.wantLen > 0 {
				assert.Len(t, result, tt.wantLen)
			} else if len(tt.wantIn) > 0 {
				combined := ""
				for _, indicator := range result {
					combined += indicator + " "
				}
				for _, want := range tt.wantIn {
					assert.Contains(t, combined, want)
				}
			}
		})
	}
}

// Note: Integration tests for AnalyzeThread are in integration_email_analyzer_test.go
// These require full mock implementations of NylasClient and LLMRouter interfaces.
